package application

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/postgres"
	redisinfra "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
)

// BenchmarkLargeScaleSeats は大規模座席数でのパフォーマンスを計測するベンチマークテスト
// 10万座席のイベントでの座席作成、検索、予約のパフォーマンスを実証します
func TestBenchmark_LargeScaleSeats(t *testing.T) {
	if testing.Short() {
		t.Skip("大規模ベンチマークテストはshortモードではスキップ")
	}

	cfg := config.Load()
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		t.Skipf("DB接続エラー: %v", err)
	}

	redisClient, err := redisinfra.NewClient(&redisinfra.Config{
		Host: cfg.Redis.Host, Port: cfg.Redis.Port,
	})
	if err != nil {
		t.Skipf("Redis接続エラー: %v", err)
	}
	lockManager := redisinfra.NewLockManager(redisClient)

	eventRepo := postgres.NewEventRepository(db)
	seatRepo := postgres.NewSeatRepository(db)
	reservationRepo := postgres.NewReservationRepository(db)

	eventService := NewEventService(eventRepo)
	seatService := NewSeatService(db, seatRepo, eventRepo, nil)
	reservationService := NewReservationService(db, reservationRepo, seatRepo, eventRepo, lockManager, nil)

	cleanup := func() {
		db.Exec("DELETE FROM reservation_seats")
		db.Exec("DELETE FROM reservations")
		db.Exec("DELETE FROM seats")
		db.Exec("DELETE FROM events")
		redisClient.Close()
		db.Close()
	}
	defer cleanup()

	ctx := context.Background()

	t.Run("10万座席ベンチマーク", func(t *testing.T) {
		const totalSeats = 100000

		// 1. イベント作成
		event, err := eventService.CreateEvent(ctx, CreateEventInput{
			Name:       "大規模コンサート - 10万人収容",
			Venue:      "新国立競技場",
			StartAt:    time.Now().Add(60 * 24 * time.Hour),
			EndAt:      time.Now().Add(60*24*time.Hour + 4*time.Hour),
			TotalSeats: totalSeats,
		})
		require.NoError(t, err)

		// 2. 10万座席を一括作成（バッチ処理で高速化）
		t.Log("=== 10万座席の一括作成開始 ===")
		startCreate := time.Now()

		const batchSize = 10000
		numBatches := totalSeats / batchSize

		for batch := 0; batch < numBatches; batch++ {
			prefix := fmt.Sprintf("SEC%02d", batch+1)
			_, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
				EventID: event.ID,
				Prefix:  prefix,
				Count:   batchSize,
				Price:   5000 + (batch * 1000), // セクションごとに価格を変える
			})
			require.NoError(t, err)

			if (batch+1)%5 == 0 {
				t.Logf("  %d/%d バッチ完了 (%d席)", batch+1, numBatches, (batch+1)*batchSize)
			}
		}

		createDuration := time.Since(startCreate)
		createRate := float64(totalSeats) / createDuration.Seconds()
		t.Logf("✅ 座席作成完了: %v (%.0f 席/秒)", createDuration, createRate)

		// 3. 空席数カウントのパフォーマンス
		t.Log("=== 空席数カウントのパフォーマンス計測 ===")
		startCount := time.Now()

		count, err := seatService.CountAvailableSeats(ctx, event.ID)
		require.NoError(t, err)
		require.Equal(t, totalSeats, count)

		countDuration := time.Since(startCount)
		t.Logf("✅ 空席数カウント: %v (COUNT: %d)", countDuration, count)

		// 4. 並行予約パフォーマンス（1000人が同時に異なる座席を予約）
		t.Log("=== 1000人同時予約のパフォーマンス計測 ===")
		const concurrentUsers = 1000
		var successCount int32
		var errorCount int32
		var wg sync.WaitGroup

		// 全座席を取得（予約対象）
		allSeats, err := seatService.GetSeatsByEvent(ctx, event.ID)
		require.NoError(t, err)
		require.Len(t, allSeats, totalSeats)

		startReserve := time.Now()

		for i := 0; i < concurrentUsers; i++ {
			wg.Add(1)
			go func(userNum int) {
				defer wg.Done()

				// 各ユーザーは1席ずつ予約（異なる座席）
				seatIdx := userNum * 10 // 衝突を避けるため10席間隔
				if seatIdx >= len(allSeats) {
					return
				}

				_, err := reservationService.CreateReservation(ctx, CreateReservationInput{
					EventID:        event.ID,
					UserID:         fmt.Sprintf("user-%05d", userNum),
					SeatIDs:        []string{allSeats[seatIdx].ID},
					IdempotencyKey: fmt.Sprintf("bench-reserve-%d-%d", time.Now().UnixNano(), userNum),
				})

				if err == nil {
					atomic.AddInt32(&successCount, 1)
				} else {
					atomic.AddInt32(&errorCount, 1)
				}
			}(i)
		}
		wg.Wait()

		reserveDuration := time.Since(startReserve)
		reserveRate := float64(successCount) / reserveDuration.Seconds()
		t.Logf("✅ 並行予約完了: %v", reserveDuration)
		t.Logf("   成功: %d, エラー: %d", successCount, errorCount)
		t.Logf("   予約処理速度: %.0f 予約/秒", reserveRate)

		// 5. 同一座席への競合予約（100人が同じ座席を予約）
		t.Log("=== 100人同時競合予約のパフォーマンス計測 ===")
		const competingUsers = 100
		targetSeatID := allSeats[50000].ID // 中央の座席を対象
		var competitionSuccess int32
		var competitionConflict int32

		startCompete := time.Now()

		var wg2 sync.WaitGroup
		for i := 0; i < competingUsers; i++ {
			wg2.Add(1)
			go func(userNum int) {
				defer wg2.Done()

				_, err := reservationService.CreateReservation(ctx, CreateReservationInput{
					EventID:        event.ID,
					UserID:         fmt.Sprintf("compete-user-%03d", userNum),
					SeatIDs:        []string{targetSeatID},
					IdempotencyKey: fmt.Sprintf("compete-%d-%d", time.Now().UnixNano(), userNum),
				})

				if err == nil {
					atomic.AddInt32(&competitionSuccess, 1)
				} else {
					atomic.AddInt32(&competitionConflict, 1)
				}
			}(i)
		}
		wg2.Wait()

		competeDuration := time.Since(startCompete)
		t.Logf("✅ 競合予約完了: %v", competeDuration)
		t.Logf("   成功: %d, 競合/エラー: %d", competitionSuccess, competitionConflict)

		// 検証
		require.Equal(t, int32(1), competitionSuccess, "競合予約では1人だけ成功するべき")
		require.Equal(t, int32(competingUsers-1), competitionConflict, "残りは全て失敗するべき")

		// 6. 最終結果サマリー
		t.Log("=================================================")
		t.Log("📊 ベンチマーク結果サマリー")
		t.Log("=================================================")
		t.Logf("総座席数: %d", totalSeats)
		t.Logf("座席作成: %v (%.0f 席/秒)", createDuration, createRate)
		t.Logf("空席カウント: %v", countDuration)
		t.Logf("並行予約 (%d人): %v (%.0f 予約/秒)", concurrentUsers, reserveDuration, reserveRate)
		t.Logf("競合予約 (%d人→1人成功): %v", competingUsers, competeDuration)
		t.Log("=================================================")
	})
}

// BenchmarkSeatQueries は座席クエリのベンチマークを計測
func BenchmarkSeatQueries(b *testing.B) {
	cfg := config.Load()
	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		b.Skipf("DB接続エラー: %v", err)
	}
	defer db.Close()

	eventRepo := postgres.NewEventRepository(db)
	seatRepo := postgres.NewSeatRepository(db)
	seatService := NewSeatService(db, seatRepo, eventRepo, nil)

	ctx := context.Background()

	// テストデータ準備
	event, _ := NewEventService(eventRepo).CreateEvent(ctx, CreateEventInput{
		Name:       "ベンチマーク用イベント",
		Venue:      "テスト会場",
		StartAt:    time.Now().Add(30 * 24 * time.Hour),
		EndAt:      time.Now().Add(30*24*time.Hour + 2*time.Hour),
		TotalSeats: 1000,
	})

	seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
		EventID: event.ID, Prefix: "BENCH", Count: 1000, Price: 5000,
	})

	b.Run("CountAvailableSeats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			seatService.CountAvailableSeats(ctx, event.ID)
		}
	})

	b.Run("GetSeatsByEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			seatService.GetSeatsByEvent(ctx, event.ID)
		}
	})
}

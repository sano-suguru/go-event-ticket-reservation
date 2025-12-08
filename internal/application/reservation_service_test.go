//go:build integration
// +build integration

package application

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/postgres"
	redisinfra "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
)

func setupTestEnv(t *testing.T) (*ReservationService, *SeatService, *EventService, func()) {
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
	txManager := postgres.NewTxManager(db)

	eventService := NewEventService(eventRepo)
	seatService := NewSeatService(seatRepo, eventRepo, nil)
	reservationService := NewReservationService(txManager, reservationRepo, seatRepo, eventRepo, lockManager, nil)

	cleanup := func() {
		db.Exec("DELETE FROM reservation_seats")
		db.Exec("DELETE FROM reservations")
		db.Exec("DELETE FROM seats")
		db.Exec("DELETE FROM events")
		redisClient.Close()
		db.Close()
	}

	return reservationService, seatService, eventService, cleanup
}

func TestConcurrentReservation(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	// イベント作成
	ev, err := eventService.CreateEvent(ctx, CreateEventInput{
		Name: "並行テストイベント", Venue: "テスト会場",
		StartAt: time.Now().Add(24 * time.Hour), EndAt: time.Now().Add(26 * time.Hour),
		TotalSeats: 10,
	})
	require.NoError(t, err)

	// 座席を1つだけ作成
	seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
		EventID: ev.ID, Prefix: "TEST", Count: 1, Price: 5000,
	})
	require.NoError(t, err)
	require.Len(t, seats, 1)
	seatID := seats[0].ID

	t.Run("10並行リクエストで1席のみ予約成功", func(t *testing.T) {
		const numGoroutines = 10
		var successCount int32
		var failCount int32
		var wg sync.WaitGroup

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(userNum int) {
				defer wg.Done()
				_, err := reservationService.CreateReservation(ctx, CreateReservationInput{
					EventID:        ev.ID,
					UserID:         "user-" + string(rune('A'+userNum)),
					SeatIDs:        []string{seatID},
					IdempotencyKey: "idem-concurrent-" + string(rune('A'+userNum)),
				})
				if err == nil {
					atomic.AddInt32(&successCount, 1)
				} else {
					atomic.AddInt32(&failCount, 1)
				}
			}(i)
		}
		wg.Wait()

		// 1つだけ成功するべき
		assert.Equal(t, int32(1), successCount, "成功は1つだけ")
		assert.Equal(t, int32(numGoroutines-1), failCount, "残りは全て失敗")
	})
}

func TestIdempotency(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	ev, err := eventService.CreateEvent(ctx, CreateEventInput{
		Name: "冪等性テストイベント", Venue: "テスト会場",
		StartAt: time.Now().Add(24 * time.Hour), EndAt: time.Now().Add(26 * time.Hour),
		TotalSeats: 10,
	})
	require.NoError(t, err)

	seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
		EventID: ev.ID, Prefix: "IDEM", Count: 2, Price: 5000,
	})
	require.NoError(t, err)

	t.Run("同じ冪等性キーで複数回リクエストしても同じ予約が返る", func(t *testing.T) {
		input := CreateReservationInput{
			EventID: ev.ID, UserID: "user-idem",
			SeatIDs:        []string{seats[0].ID},
			IdempotencyKey: "same-idem-key",
		}

		res1, err := reservationService.CreateReservation(ctx, input)
		require.NoError(t, err)

		res2, err := reservationService.CreateReservation(ctx, input)
		require.NoError(t, err)

		assert.Equal(t, res1.ID, res2.ID, "同じ予約IDが返るべき")
	})
}

func TestSeatAlreadyReserved(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	ev, err := eventService.CreateEvent(ctx, CreateEventInput{
		Name: "座席予約済みテスト", Venue: "テスト会場",
		StartAt: time.Now().Add(24 * time.Hour), EndAt: time.Now().Add(26 * time.Hour),
		TotalSeats: 10,
	})
	require.NoError(t, err)

	seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
		EventID: ev.ID, Prefix: "RES", Count: 1, Price: 5000,
	})
	require.NoError(t, err)

	t.Run("予約済み座席の再予約はエラー", func(t *testing.T) {
		// 最初の予約
		_, err := reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID: ev.ID, UserID: "user-first",
			SeatIDs:        []string{seats[0].ID},
			IdempotencyKey: "first-reservation",
		})
		require.NoError(t, err)

		// 2番目の予約（別のユーザー、別の冪等性キー）
		_, err = reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID: ev.ID, UserID: "user-second",
			SeatIDs:        []string{seats[0].ID},
			IdempotencyKey: "second-reservation",
		})
		assert.ErrorIs(t, err, seat.ErrSeatAlreadyReserved)
	})
}

func TestReservationConfirmAndCancel(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	ev, err := eventService.CreateEvent(ctx, CreateEventInput{
		Name: "確定キャンセルテスト", Venue: "テスト会場",
		StartAt: time.Now().Add(24 * time.Hour), EndAt: time.Now().Add(26 * time.Hour),
		TotalSeats: 10,
	})
	require.NoError(t, err)

	seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
		EventID: ev.ID, Prefix: "CC", Count: 2, Price: 5000,
	})
	require.NoError(t, err)

	t.Run("予約確定後は座席がconfirmed", func(t *testing.T) {
		res, err := reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID: ev.ID, UserID: "user-confirm",
			SeatIDs:        []string{seats[0].ID},
			IdempotencyKey: "confirm-test",
		})
		require.NoError(t, err)

		confirmed, err := reservationService.ConfirmReservation(ctx, res.ID)
		require.NoError(t, err)
		assert.Equal(t, "confirmed", string(confirmed.Status))
	})

	t.Run("キャンセル後は座席が再利用可能", func(t *testing.T) {
		res, err := reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID: ev.ID, UserID: "user-cancel",
			SeatIDs:        []string{seats[1].ID},
			IdempotencyKey: "cancel-test",
		})
		require.NoError(t, err)

		cancelled, err := reservationService.CancelReservation(ctx, res.ID)
		require.NoError(t, err)
		assert.Equal(t, "cancelled", string(cancelled.Status))

		// 同じ座席を再予約できる
		_, err = reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID: ev.ID, UserID: "user-reuse",
			SeatIDs:        []string{seats[1].ID},
			IdempotencyKey: "reuse-test",
		})
		require.NoError(t, err)
	})
}

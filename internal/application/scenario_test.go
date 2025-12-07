package application

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
)

// TestScenario_FullReservationFlow はチケット予約の完全なフローをテストします
// イベント作成 → 座席作成 → 予約 → 確定 → 座席状態確認
func TestScenario_FullReservationFlow(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("完全な予約フロー", func(t *testing.T) {
		// 1. イベント作成
		event, err := eventService.CreateEvent(ctx, CreateEventInput{
			Name:       "東京ドームコンサート 2025",
			Venue:      "東京ドーム",
			StartAt:    time.Now().Add(30 * 24 * time.Hour),
			EndAt:      time.Now().Add(30*24*time.Hour + 3*time.Hour),
			TotalSeats: 100,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)

		// 2. 座席を一括作成
		seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
			EventID: event.ID,
			Prefix:  "A",
			Count:   10,
			Price:   15000,
		})
		require.NoError(t, err)
		assert.Len(t, seats, 10)

		// 3. 空席数を確認
		availableCount, err := seatService.CountAvailableSeats(ctx, event.ID)
		require.NoError(t, err)
		assert.Equal(t, 10, availableCount)

		// 4. 予約を作成（2席）
		res, err := reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID:        event.ID,
			UserID:         "user-tanaka",
			SeatIDs:        []string{seats[0].ID, seats[1].ID},
			IdempotencyKey: "order-tanaka-001",
		})
		require.NoError(t, err)
		assert.Equal(t, reservation.StatusPending, res.Status)
		assert.Equal(t, 30000, res.TotalAmount) // 15000 * 2

		// 5. 予約を確定
		confirmed, err := reservationService.ConfirmReservation(ctx, res.ID)
		require.NoError(t, err)
		assert.Equal(t, reservation.StatusConfirmed, confirmed.Status)

		// 6. 空席数が減っていることを確認
		availableCount, err = seatService.CountAvailableSeats(ctx, event.ID)
		require.NoError(t, err)
		assert.Equal(t, 8, availableCount)
	})
}

// TestScenario_MultipleUsersCompeting は複数ユーザーが同じ座席を競合するシナリオ
func TestScenario_MultipleUsersCompeting(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("50人が同時に同じ座席を予約", func(t *testing.T) {
		// イベントと座席を準備
		event, err := eventService.CreateEvent(ctx, CreateEventInput{
			Name:       "人気アーティストライブ",
			Venue:      "武道館",
			StartAt:    time.Now().Add(14 * 24 * time.Hour),
			EndAt:      time.Now().Add(14*24*time.Hour + 2*time.Hour),
			TotalSeats: 1,
		})
		require.NoError(t, err)

		seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
			EventID: event.ID, Prefix: "VIP", Count: 1, Price: 50000,
		})
		require.NoError(t, err)
		targetSeatID := seats[0].ID

		// 50人が同時に予約を試みる
		const numUsers = 50
		var successCount int32
		var conflictCount int32
		var otherErrorCount int32
		var wg sync.WaitGroup

		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			go func(userNum int) {
				defer wg.Done()
				_, err := reservationService.CreateReservation(ctx, CreateReservationInput{
					EventID:        event.ID,
					UserID:         "user-" + string(rune('A'+userNum%26)) + string(rune('0'+userNum/26)),
					SeatIDs:        []string{targetSeatID},
					IdempotencyKey: "compete-" + time.Now().Format("20060102150405") + "-" + string(rune('A'+userNum%26)),
				})
				switch err {
				case nil:
					atomic.AddInt32(&successCount, 1)
				case seat.ErrSeatAlreadyReserved:
					atomic.AddInt32(&conflictCount, 1)
				default:
					atomic.AddInt32(&otherErrorCount, 1)
				}
			}(i)
		}
		wg.Wait()

		// 結果を検証
		assert.Equal(t, int32(1), successCount, "1人だけが予約成功")
		assert.Equal(t, int32(numUsers-1), conflictCount+otherErrorCount, "残りは全て失敗")
		t.Logf("成功: %d, 競合: %d, その他エラー: %d", successCount, conflictCount, otherErrorCount)
	})
}

// TestScenario_CancelAndRebook はキャンセル後の再予約シナリオ
func TestScenario_CancelAndRebook(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("キャンセルされた座席を別ユーザーが予約", func(t *testing.T) {
		// 準備
		event, err := eventService.CreateEvent(ctx, CreateEventInput{
			Name:       "キャンセル再予約テスト",
			Venue:      "テスト会場",
			StartAt:    time.Now().Add(7 * 24 * time.Hour),
			EndAt:      time.Now().Add(7*24*time.Hour + 2*time.Hour),
			TotalSeats: 1,
		})
		require.NoError(t, err)

		seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
			EventID: event.ID, Prefix: "S", Count: 1, Price: 10000,
		})
		require.NoError(t, err)
		seatID := seats[0].ID

		// ユーザーAが予約
		resA, err := reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID:        event.ID,
			UserID:         "user-A",
			SeatIDs:        []string{seatID},
			IdempotencyKey: "cancel-rebook-A",
		})
		require.NoError(t, err)

		// ユーザーBが同じ座席を予約しようとして失敗
		_, err = reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID:        event.ID,
			UserID:         "user-B",
			SeatIDs:        []string{seatID},
			IdempotencyKey: "cancel-rebook-B-1",
		})
		assert.ErrorIs(t, err, seat.ErrSeatAlreadyReserved)

		// ユーザーAがキャンセル
		cancelled, err := reservationService.CancelReservation(ctx, resA.ID)
		require.NoError(t, err)
		assert.Equal(t, reservation.StatusCancelled, cancelled.Status)

		// ユーザーBが再度予約して成功
		resB, err := reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID:        event.ID,
			UserID:         "user-B",
			SeatIDs:        []string{seatID},
			IdempotencyKey: "cancel-rebook-B-2",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resB.ID)
	})
}

// TestScenario_MultiSeatReservation は複数座席の一括予約シナリオ
func TestScenario_MultiSeatReservation(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("5席まとめて予約", func(t *testing.T) {
		event, err := eventService.CreateEvent(ctx, CreateEventInput{
			Name:       "グループ予約テスト",
			Venue:      "大ホール",
			StartAt:    time.Now().Add(10 * 24 * time.Hour),
			EndAt:      time.Now().Add(10*24*time.Hour + 2*time.Hour),
			TotalSeats: 20,
		})
		require.NoError(t, err)

		seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
			EventID: event.ID, Prefix: "G", Count: 10, Price: 8000,
		})
		require.NoError(t, err)

		// 5席を一括予約
		seatIDs := []string{seats[0].ID, seats[1].ID, seats[2].ID, seats[3].ID, seats[4].ID}
		res, err := reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID:        event.ID,
			UserID:         "group-leader",
			SeatIDs:        seatIDs,
			IdempotencyKey: "group-booking-001",
		})
		require.NoError(t, err)
		assert.Equal(t, 40000, res.TotalAmount) // 8000 * 5

		// 空席数が減っていることを確認
		availableCount, err := seatService.CountAvailableSeats(ctx, event.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, availableCount)
	})

	t.Run("一部の座席が予約済みの場合は全体が失敗", func(t *testing.T) {
		event, err := eventService.CreateEvent(ctx, CreateEventInput{
			Name:       "部分予約失敗テスト",
			Venue:      "テスト会場",
			StartAt:    time.Now().Add(11 * 24 * time.Hour),
			EndAt:      time.Now().Add(11*24*time.Hour + 2*time.Hour),
			TotalSeats: 10,
		})
		require.NoError(t, err)

		seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
			EventID: event.ID, Prefix: "P", Count: 5, Price: 5000,
		})
		require.NoError(t, err)

		// 座席1を先に予約
		_, err = reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID:        event.ID,
			UserID:         "first-user",
			SeatIDs:        []string{seats[0].ID},
			IdempotencyKey: "partial-first",
		})
		require.NoError(t, err)

		// 座席0-2を一括予約しようとして失敗（座席0が既に予約済み）
		_, err = reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID:        event.ID,
			UserID:         "second-user",
			SeatIDs:        []string{seats[0].ID, seats[1].ID, seats[2].ID},
			IdempotencyKey: "partial-second",
		})
		assert.Error(t, err)

		// 座席1と2はまだ空席のはず（トランザクションでロールバックされている）
		availableCount, err := seatService.CountAvailableSeats(ctx, event.ID)
		require.NoError(t, err)
		assert.Equal(t, 4, availableCount) // 5 - 1（最初の予約分のみ）
	})
}

// TestScenario_ConfirmedCannotBeModified は確定済み予約の変更不可シナリオ
func TestScenario_ConfirmedCannotBeModified(t *testing.T) {
	reservationService, seatService, eventService, cleanup := setupTestEnv(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("確定済み予約はキャンセルできない", func(t *testing.T) {
		event, err := eventService.CreateEvent(ctx, CreateEventInput{
			Name:       "確定後変更不可テスト",
			Venue:      "テスト会場",
			StartAt:    time.Now().Add(5 * 24 * time.Hour),
			EndAt:      time.Now().Add(5*24*time.Hour + 2*time.Hour),
			TotalSeats: 5,
		})
		require.NoError(t, err)

		seats, err := seatService.CreateBulkSeats(ctx, CreateBulkSeatsInput{
			EventID: event.ID, Prefix: "C", Count: 1, Price: 10000,
		})
		require.NoError(t, err)

		// 予約を作成して確定
		res, err := reservationService.CreateReservation(ctx, CreateReservationInput{
			EventID:        event.ID,
			UserID:         "confirmed-user",
			SeatIDs:        []string{seats[0].ID},
			IdempotencyKey: "confirm-no-cancel",
		})
		require.NoError(t, err)

		_, err = reservationService.ConfirmReservation(ctx, res.ID)
		require.NoError(t, err)

		// 確定済みの予約をキャンセルしようとしてエラー
		_, err = reservationService.CancelReservation(ctx, res.ID)
		assert.Error(t, err)
	})
}

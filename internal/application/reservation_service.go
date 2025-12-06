package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
	redislock "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
)

type ReservationService struct {
	db              *sqlx.DB
	reservationRepo reservation.Repository
	seatRepo        seat.Repository
	eventRepo       event.Repository
	lockManager     *redislock.LockManager
}

func NewReservationService(db *sqlx.DB, rr reservation.Repository, sr seat.Repository, er event.Repository, lm *redislock.LockManager) *ReservationService {
	return &ReservationService{db: db, reservationRepo: rr, seatRepo: sr, eventRepo: er, lockManager: lm}
}

type CreateReservationInput struct {
	EventID        string
	UserID         string
	SeatIDs        []string
	IdempotencyKey string
}

func (s *ReservationService) CreateReservation(ctx context.Context, input CreateReservationInput) (*reservation.Reservation, error) {
	// 冪等性チェック
	existing, err := s.reservationRepo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, reservation.ErrReservationNotFound) {
		return nil, fmt.Errorf("冪等性チェックに失敗: %w", err)
	}

	// 分散ロックを取得（座席IDをソートしてデッドロックを防止）
	lockKey := s.buildSeatLockKey(input.SeatIDs)
	var lock *redislock.DistributedLock
	if s.lockManager != nil {
		lock, err = s.lockManager.AcquireLockWithRetry(ctx, lockKey, 10*time.Second, 3, 100*time.Millisecond)
		if err != nil {
			if errors.Is(err, redislock.ErrLockNotAcquired) {
				return nil, fmt.Errorf("座席が他のユーザーによって処理中です")
			}
			return nil, fmt.Errorf("ロック取得に失敗: %w", err)
		}
		defer lock.Release(ctx)
	}

	// イベント確認
	ev, err := s.eventRepo.GetByID(ctx, input.EventID)
	if err != nil {
		return nil, fmt.Errorf("イベント取得に失敗: %w", err)
	}
	if !ev.IsBookingOpen() {
		return nil, event.ErrEventNotOpen
	}

	// 座席確認
	seats, err := s.seatRepo.GetByEventID(ctx, input.EventID)
	if err != nil {
		return nil, fmt.Errorf("座席取得に失敗: %w", err)
	}
	seatMap := make(map[string]*seat.Seat)
	for _, se := range seats {
		seatMap[se.ID] = se
	}
	var totalAmount int
	for _, id := range input.SeatIDs {
		se, ok := seatMap[id]
		if !ok {
			return nil, seat.ErrSeatNotFound
		}
		if !se.IsAvailable() {
			return nil, seat.ErrSeatAlreadyReserved
		}
		totalAmount += se.Price
	}

	// 予約作成
	res := reservation.NewReservation(input.EventID, input.UserID, input.IdempotencyKey, input.SeatIDs, totalAmount)
	if err := res.Validate(); err != nil {
		return nil, err
	}

	// トランザクション
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback()

	if err := s.reservationRepo.Create(ctx, tx, res); err != nil {
		return nil, err
	}
	if err := s.seatRepo.ReserveSeats(ctx, tx, input.SeatIDs, res.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("コミットに失敗: %w", err)
	}
	return res, nil
}

// buildSeatLockKey は座席IDからロックキーを生成（ソートしてデッドロック防止）
func (s *ReservationService) buildSeatLockKey(seatIDs []string) string {
	sorted := make([]string, len(seatIDs))
	copy(sorted, seatIDs)
	sort.Strings(sorted)
	return "seats:" + strings.Join(sorted, ",")
}

func (s *ReservationService) GetReservation(ctx context.Context, id string) (*reservation.Reservation, error) {
	return s.reservationRepo.GetByID(ctx, id)
}

func (s *ReservationService) GetUserReservations(ctx context.Context, userID string, limit, offset int) ([]*reservation.Reservation, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.reservationRepo.GetByUserID(ctx, userID, limit, offset)
}

func (s *ReservationService) ConfirmReservation(ctx context.Context, id string) (*reservation.Reservation, error) {
	res, err := s.reservationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := res.Confirm(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback()
	if err := s.seatRepo.ConfirmSeats(ctx, tx, res.SeatIDs); err != nil {
		return nil, err
	}
	if err := s.reservationRepo.Update(ctx, tx, res); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("コミットに失敗: %w", err)
	}
	return res, nil
}

func (s *ReservationService) CancelReservation(ctx context.Context, id string) (*reservation.Reservation, error) {
	res, err := s.reservationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := res.Cancel(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback()
	if err := s.seatRepo.ReleaseSeats(ctx, tx, res.SeatIDs); err != nil {
		return nil, err
	}
	if err := s.reservationRepo.Update(ctx, tx, res); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("コミットに失敗: %w", err)
	}
	return res, nil
}

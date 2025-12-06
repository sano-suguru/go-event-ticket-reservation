package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
)

type ReservationService struct {
	db              *sqlx.DB
	reservationRepo reservation.Repository
	seatRepo        seat.Repository
	eventRepo       event.Repository
}

func NewReservationService(db *sqlx.DB, rr reservation.Repository, sr seat.Repository, er event.Repository) *ReservationService {
	return &ReservationService{db: db, reservationRepo: rr, seatRepo: sr, eventRepo: er}
}

type CreateReservationInput struct {
	EventID        string
	UserID         string
	SeatIDs        []string
	IdempotencyKey string
}

func (s *ReservationService) CreateReservation(ctx context.Context, input CreateReservationInput) (*reservation.Reservation, error) {
	existing, err := s.reservationRepo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, reservation.ErrReservationNotFound) {
		return nil, fmt.Errorf("冪等性チェックに失敗: %w", err)
	}
	ev, err := s.eventRepo.GetByID(ctx, input.EventID)
	if err != nil {
		return nil, fmt.Errorf("イベント取得に失敗: %w", err)
	}
	if !ev.IsBookingOpen() {
		return nil, event.ErrEventNotOpen
	}
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
	res := reservation.NewReservation(input.EventID, input.UserID, input.IdempotencyKey, input.SeatIDs, totalAmount)
	if err := res.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback()
	// 先に予約を作成してIDを取得
	if err := s.reservationRepo.Create(ctx, tx, res); err != nil {
		return nil, err
	}
	// 座席を予約状態に更新
	if err := s.seatRepo.ReserveSeats(ctx, tx, input.SeatIDs, res.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("コミットに失敗: %w", err)
	}
	return res, nil
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

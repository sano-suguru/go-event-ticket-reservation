package application

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
)

type SeatService struct {
	db        *sqlx.DB
	seatRepo  seat.Repository
	eventRepo event.Repository
}

func NewSeatService(db *sqlx.DB, sr seat.Repository, er event.Repository) *SeatService {
	return &SeatService{db: db, seatRepo: sr, eventRepo: er}
}

type CreateSeatInput struct {
	EventID    string
	SeatNumber string
	Price      int
}

func (s *SeatService) CreateSeat(ctx context.Context, input CreateSeatInput) (*seat.Seat, error) {
	if _, err := s.eventRepo.GetByID(ctx, input.EventID); err != nil {
		return nil, fmt.Errorf("イベント取得に失敗: %w", err)
	}
	se := seat.NewSeat(input.EventID, input.SeatNumber, input.Price)
	if err := se.Validate(); err != nil {
		return nil, err
	}
	if err := s.seatRepo.Create(ctx, se); err != nil {
		return nil, err
	}
	return se, nil
}

type CreateBulkSeatsInput struct {
	EventID string
	Prefix  string
	Count   int
	Price   int
}

func (s *SeatService) CreateBulkSeats(ctx context.Context, input CreateBulkSeatsInput) ([]*seat.Seat, error) {
	if _, err := s.eventRepo.GetByID(ctx, input.EventID); err != nil {
		return nil, fmt.Errorf("イベント取得に失敗: %w", err)
	}
	seats := make([]*seat.Seat, 0, input.Count)
	for i := 1; i <= input.Count; i++ {
		seatNumber := fmt.Sprintf("%s-%d", input.Prefix, i)
		se := seat.NewSeat(input.EventID, seatNumber, input.Price)
		if err := se.Validate(); err != nil {
			return nil, err
		}
		seats = append(seats, se)
	}
	if err := s.seatRepo.CreateBulk(ctx, seats); err != nil {
		return nil, err
	}
	return seats, nil
}

func (s *SeatService) GetSeat(ctx context.Context, id string) (*seat.Seat, error) {
	return s.seatRepo.GetByID(ctx, id)
}

func (s *SeatService) GetSeatsByEvent(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	return s.seatRepo.GetByEventID(ctx, eventID)
}

func (s *SeatService) GetAvailableSeatsByEvent(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	return s.seatRepo.GetAvailableByEventID(ctx, eventID)
}

func (s *SeatService) CountAvailableSeats(ctx context.Context, eventID string) (int, error) {
	return s.seatRepo.CountAvailableByEventID(ctx, eventID)
}

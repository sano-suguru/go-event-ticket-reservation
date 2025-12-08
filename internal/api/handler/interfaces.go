package handler

import (
	"context"
	"time"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/application"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
)

// EventServiceInterface はイベントサービスのインターフェース
type EventServiceInterface interface {
	CreateEvent(ctx context.Context, input application.CreateEventInput) (*event.Event, error)
	GetEvent(ctx context.Context, id string) (*event.Event, error)
	ListEvents(ctx context.Context, limit, offset int) ([]*event.Event, error)
	UpdateEvent(ctx context.Context, input application.UpdateEventInput) (*event.Event, error)
	DeleteEvent(ctx context.Context, id string) error
}

// SeatServiceInterface は座席サービスのインターフェース
type SeatServiceInterface interface {
	CreateSeat(ctx context.Context, input application.CreateSeatInput) (*seat.Seat, error)
	CreateBulkSeats(ctx context.Context, input application.CreateBulkSeatsInput) ([]*seat.Seat, error)
	GetSeat(ctx context.Context, id string) (*seat.Seat, error)
	GetSeatsByEvent(ctx context.Context, eventID string) ([]*seat.Seat, error)
	GetAvailableSeatsByEvent(ctx context.Context, eventID string) ([]*seat.Seat, error)
	CountAvailableSeats(ctx context.Context, eventID string) (int, error)
}

// ReservationServiceInterface は予約サービスのインターフェース
type ReservationServiceInterface interface {
	CreateReservation(ctx context.Context, input application.CreateReservationInput) (*reservation.Reservation, error)
	GetReservation(ctx context.Context, id string) (*reservation.Reservation, error)
	GetUserReservations(ctx context.Context, userID string, limit, offset int) ([]*reservation.Reservation, error)
	ConfirmReservation(ctx context.Context, id string) (*reservation.Reservation, error)
	CancelReservation(ctx context.Context, id string) (*reservation.Reservation, error)
	CancelExpiredReservations(ctx context.Context, expireAfter time.Duration) (int, error)
}

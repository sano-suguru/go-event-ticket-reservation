package seat

import "time"

// Status は座席の状態を表す
type Status string

const (
	StatusAvailable Status = "available"
	StatusReserved  Status = "reserved"
	StatusConfirmed Status = "confirmed"
)

// Seat は座席エンティティを表す
type Seat struct {
	ID         string
	EventID    string
	SeatNumber string
	Status     Status
	Price      int
	ReservedBy *string // reservation_id
	ReservedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Version    int // 楽観的ロック用
}

// NewSeat は新しい座席を作成する
func NewSeat(eventID, seatNumber string, price int) *Seat {
	now := time.Now()
	return &Seat{
		EventID:    eventID,
		SeatNumber: seatNumber,
		Status:     StatusAvailable,
		Price:      price,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    0,
	}
}

// IsAvailable は座席が予約可能かを返す
func (s *Seat) IsAvailable() bool {
	return s.Status == StatusAvailable
}

// Reserve は座席を予約状態にする
func (s *Seat) Reserve(reservationID string) error {
	if s.Status != StatusAvailable {
		return ErrSeatNotAvailable
	}
	now := time.Now()
	s.Status = StatusReserved
	s.ReservedBy = &reservationID
	s.ReservedAt = &now
	s.UpdatedAt = now
	return nil
}

// Confirm は座席を確定状態にする
func (s *Seat) Confirm() error {
	if s.Status != StatusReserved {
		return ErrSeatNotReserved
	}
	s.Status = StatusConfirmed
	s.UpdatedAt = time.Now()
	return nil
}

// Release は座席を解放する
func (s *Seat) Release() {
	s.Status = StatusAvailable
	s.ReservedBy = nil
	s.ReservedAt = nil
	s.UpdatedAt = time.Now()
}

// Validate は座席の検証を行う
func (s *Seat) Validate() error {
	if s.EventID == "" {
		return ErrEventIDRequired
	}
	if s.SeatNumber == "" {
		return ErrSeatNumberRequired
	}
	if s.Price < 0 {
		return ErrInvalidPrice
	}
	return nil
}

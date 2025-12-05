package reservation

import "time"

// Status は予約の状態を表す
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusCancelled Status = "cancelled"
)

// Reservation は予約エンティティを表す
type Reservation struct {
	ID             string
	EventID        string
	UserID         string
	SeatIDs        []string
	Status         Status
	IdempotencyKey string
	ExpiresAt      time.Time
	ConfirmedAt    *time.Time
	TotalAmount    int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ReservationExpiration は予約の有効期限（デフォルト15分）
const ReservationExpiration = 15 * time.Minute

// NewReservation は新しい予約を作成する
func NewReservation(eventID, userID, idempotencyKey string, seatIDs []string, totalAmount int) *Reservation {
	now := time.Now()
	return &Reservation{
		EventID:        eventID,
		UserID:         userID,
		SeatIDs:        seatIDs,
		Status:         StatusPending,
		IdempotencyKey: idempotencyKey,
		ExpiresAt:      now.Add(ReservationExpiration),
		TotalAmount:    totalAmount,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// IsExpired は予約が期限切れかを返す
func (r *Reservation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// IsPending は予約が保留中かを返す
func (r *Reservation) IsPending() bool {
	return r.Status == StatusPending
}

// Confirm は予約を確定する
func (r *Reservation) Confirm() error {
	if r.Status != StatusPending {
		return ErrReservationNotPending
	}
	if r.IsExpired() {
		return ErrReservationExpired
	}
	now := time.Now()
	r.Status = StatusConfirmed
	r.ConfirmedAt = &now
	r.UpdatedAt = now
	return nil
}

// Cancel は予約をキャンセルする
func (r *Reservation) Cancel() error {
	if r.Status == StatusCancelled {
		return ErrReservationAlreadyCancelled
	}
	if r.Status == StatusConfirmed {
		return ErrReservationAlreadyConfirmed
	}
	r.Status = StatusCancelled
	r.UpdatedAt = time.Now()
	return nil
}

// Validate は予約の検証を行う
func (r *Reservation) Validate() error {
	if r.EventID == "" {
		return ErrEventIDRequired
	}
	if r.UserID == "" {
		return ErrUserIDRequired
	}
	if len(r.SeatIDs) == 0 {
		return ErrSeatIDsRequired
	}
	if r.IdempotencyKey == "" {
		return ErrIdempotencyKeyRequired
	}
	return nil
}

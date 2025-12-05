package event

import "time"

// Event はイベントエンティティを表す
type Event struct {
	ID          string
	Name        string
	Description string
	Venue       string
	StartAt     time.Time
	EndAt       time.Time
	TotalSeats  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int // 楽観的ロック用
}

// NewEvent は新しいイベントを作成する
func NewEvent(name, description, venue string, startAt, endAt time.Time, totalSeats int) *Event {
	now := time.Now()
	return &Event{
		Name:        name,
		Description: description,
		Venue:       venue,
		StartAt:     startAt,
		EndAt:       endAt,
		TotalSeats:  totalSeats,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     0,
	}
}

// Validate はイベントの検証を行う
func (e *Event) Validate() error {
	if e.Name == "" {
		return ErrEventNameRequired
	}
	if e.TotalSeats <= 0 {
		return ErrInvalidTotalSeats
	}
	if e.EndAt.Before(e.StartAt) {
		return ErrInvalidEventTime
	}
	return nil
}

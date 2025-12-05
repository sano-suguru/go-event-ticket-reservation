package application

import (
	"context"
	"fmt"
	"time"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
)

type EventService struct {
	eventRepo event.Repository
}

func NewEventService(eventRepo event.Repository) *EventService {
	return &EventService{eventRepo: eventRepo}
}

type CreateEventInput struct {
	Name        string
	Description string
	Venue       string
	StartAt     time.Time
	EndAt       time.Time
	TotalSeats  int
}

func (s *EventService) CreateEvent(ctx context.Context, input CreateEventInput) (*event.Event, error) {
	e := event.NewEvent(input.Name, input.Description, input.Venue, input.StartAt, input.EndAt, input.TotalSeats)
	if err := e.Validate(); err != nil {
		return nil, fmt.Errorf("バリデーションエラー: %w", err)
	}
	if err := s.eventRepo.Create(ctx, e); err != nil {
		return nil, fmt.Errorf("イベント作成に失敗しました: %w", err)
	}
	return e, nil
}

func (s *EventService) GetEvent(ctx context.Context, id string) (*event.Event, error) {
	return s.eventRepo.GetByID(ctx, id)
}

func (s *EventService) ListEvents(ctx context.Context, limit, offset int) ([]*event.Event, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.eventRepo.List(ctx, limit, offset)
}

type UpdateEventInput struct {
	ID          string
	Name        string
	Description string
	Venue       string
	StartAt     time.Time
	EndAt       time.Time
	TotalSeats  int
}

func (s *EventService) UpdateEvent(ctx context.Context, input UpdateEventInput) (*event.Event, error) {
	e, err := s.eventRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	e.Name = input.Name
	e.Description = input.Description
	e.Venue = input.Venue
	e.StartAt = input.StartAt
	e.EndAt = input.EndAt
	e.TotalSeats = input.TotalSeats
	if err := e.Validate(); err != nil {
		return nil, fmt.Errorf("バリデーションエラー: %w", err)
	}
	if err := s.eventRepo.Update(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, id string) error {
	return s.eventRepo.Delete(ctx, id)
}

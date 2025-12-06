package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
	redisinfra "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/logger"
)

const (
	seatCacheTTL = 30 * time.Second
)

type SeatService struct {
	db        *sqlx.DB
	seatRepo  seat.Repository
	eventRepo event.Repository
	cache     *redisinfra.SeatCache
}

func NewSeatService(db *sqlx.DB, sr seat.Repository, er event.Repository, cache *redisinfra.SeatCache) *SeatService {
	return &SeatService{db: db, seatRepo: sr, eventRepo: er, cache: cache}
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
	// キャッシュから取得を試みる
	if s.cache != nil {
		count, err := s.cache.GetAvailableCount(ctx, eventID)
		if err == nil {
			logger.Debug("キャッシュヒット", zap.String("event_id", eventID), zap.Int("count", count))
			return count, nil
		}
		if !errors.Is(err, redisinfra.ErrCacheMiss) {
			logger.Warn("キャッシュ取得エラー", zap.Error(err))
		}
	}

	// DBから取得
	count, err := s.seatRepo.CountAvailableByEventID(ctx, eventID)
	if err != nil {
		return 0, err
	}

	// キャッシュに保存
	if s.cache != nil {
		if cacheErr := s.cache.SetAvailableCount(ctx, eventID, count, seatCacheTTL); cacheErr != nil {
			logger.Warn("キャッシュ保存エラー", zap.Error(cacheErr))
		}
	}

	return count, nil
}

// InvalidateCache はイベントのキャッシュを無効化する
func (s *SeatService) InvalidateCache(ctx context.Context, eventID string) {
	if s.cache != nil {
		if err := s.cache.Invalidate(ctx, eventID); err != nil {
			logger.Warn("キャッシュ無効化エラー", zap.Error(err))
		}
	}
}

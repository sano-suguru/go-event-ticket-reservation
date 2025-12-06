package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
	redislock "github.com/sanosuguru/go-event-ticket-reservation/internal/infrastructure/redis"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/logger"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/metrics"
)

type ReservationService struct {
	db              *sqlx.DB
	reservationRepo reservation.Repository
	seatRepo        seat.Repository
	eventRepo       event.Repository
	lockManager     *redislock.LockManager
	seatCache       *redislock.SeatCache
}

func NewReservationService(db *sqlx.DB, rr reservation.Repository, sr seat.Repository, er event.Repository, lm *redislock.LockManager, cache *redislock.SeatCache) *ReservationService {
	return &ReservationService{db: db, reservationRepo: rr, seatRepo: sr, eventRepo: er, lockManager: lm, seatCache: cache}
}

type CreateReservationInput struct {
	EventID        string
	UserID         string
	SeatIDs        []string
	IdempotencyKey string
}

func (s *ReservationService) CreateReservation(ctx context.Context, input CreateReservationInput) (*reservation.Reservation, error) {
	log := logger.With(
		zap.String("event_id", input.EventID),
		zap.String("user_id", input.UserID),
		zap.String("idempotency_key", input.IdempotencyKey),
		zap.Strings("seat_ids", input.SeatIDs),
	)

	// 冪等性チェック
	existing, err := s.reservationRepo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
	if err == nil {
		log.Info("冪等性チェック: 既存予約を返却")
		return existing, nil
	}
	if !errors.Is(err, reservation.ErrReservationNotFound) {
		log.Error("冪等性チェックに失敗", zap.Error(err))
		return nil, fmt.Errorf("冪等性チェックに失敗: %w", err)
	}

	// 分散ロックを取得（座席IDをソートしてデッドロックを防止）
	lockKey := s.buildSeatLockKey(input.SeatIDs)
	var lock *redislock.DistributedLock
	if s.lockManager != nil {
		log.Debug("分散ロック取得中", zap.String("lock_key", lockKey))
		lockStart := time.Now()
		lock, err = s.lockManager.AcquireLockWithRetry(ctx, lockKey, 10*time.Second, 3, 100*time.Millisecond)
		lockDuration := time.Since(lockStart).Seconds()
		if err != nil {
			if m := metrics.Get(); m != nil {
				m.DistributedLockDuration.WithLabelValues("acquire", "failed").Observe(lockDuration)
				m.ReservationsTotal.WithLabelValues("lock_failed").Inc()
			}
			if errors.Is(err, redislock.ErrLockNotAcquired) {
				log.Warn("分散ロック取得失敗: 他のユーザーが処理中")
				return nil, fmt.Errorf("座席が他のユーザーによって処理中です")
			}
			log.Error("ロック取得に失敗", zap.Error(err))
			return nil, fmt.Errorf("ロック取得に失敗: %w", err)
		}
		if m := metrics.Get(); m != nil {
			m.DistributedLockDuration.WithLabelValues("acquire", "success").Observe(lockDuration)
		}
		defer func() {
			releaseStart := time.Now()
			_ = lock.Release(ctx)
			if m := metrics.Get(); m != nil {
				m.DistributedLockDuration.WithLabelValues("release", "success").Observe(time.Since(releaseStart).Seconds())
			}
		}()
		log.Debug("分散ロック取得成功")
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
		log.Error("座席取得に失敗", zap.Error(err))
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
			log.Warn("座席が見つからない", zap.String("seat_id", id))
			return nil, seat.ErrSeatNotFound
		}
		if !se.IsAvailable() {
			log.Warn("座席が既に予約済み", zap.String("seat_id", id), zap.String("status", string(se.Status)))
			if m := metrics.Get(); m != nil {
				m.ReservationsTotal.WithLabelValues("conflict").Inc()
			}
			return nil, seat.ErrSeatAlreadyReserved
		}
		totalAmount += se.Price
	}

	// 予約作成
	res := reservation.NewReservation(input.EventID, input.UserID, input.IdempotencyKey, input.SeatIDs, totalAmount)
	if validateErr := res.Validate(); validateErr != nil {
		log.Error("予約バリデーション失敗", zap.Error(validateErr))
		return nil, validateErr
	}

	// トランザクション
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error("トランザクション開始に失敗", zap.Error(err))
		return nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback()

	if err := s.reservationRepo.Create(ctx, tx, res); err != nil {
		log.Error("予約作成に失敗", zap.Error(err))
		return nil, err
	}
	if err := s.seatRepo.ReserveSeats(ctx, tx, input.SeatIDs, res.ID); err != nil {
		log.Error("座席予約に失敗", zap.Error(err))
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		log.Error("コミットに失敗", zap.Error(err))
		return nil, fmt.Errorf("コミットに失敗: %w", err)
	}

	// メトリクス記録: 予約成功
	if m := metrics.Get(); m != nil {
		m.ReservationsTotal.WithLabelValues("success").Inc()
		m.ActiveReservations.WithLabelValues("pending").Inc()
	}

	// キャッシュ無効化
	s.invalidateSeatCache(ctx, input.EventID)

	log.Info("予約作成成功", zap.String("reservation_id", res.ID), zap.Int("total_amount", totalAmount))
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
	if confirmErr := res.Confirm(); confirmErr != nil {
		return nil, confirmErr
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

	// メトリクス記録: 予約確定
	if m := metrics.Get(); m != nil {
		m.ActiveReservations.WithLabelValues("pending").Dec()
		m.ActiveReservations.WithLabelValues("confirmed").Inc()
	}

	return res, nil
}

func (s *ReservationService) CancelReservation(ctx context.Context, id string) (*reservation.Reservation, error) {
	res, err := s.reservationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cancelErr := res.Cancel(); cancelErr != nil {
		return nil, cancelErr
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

	// メトリクス記録: 予約キャンセル
	if m := metrics.Get(); m != nil {
		m.ActiveReservations.WithLabelValues("pending").Dec()
	}

	// キャッシュ無効化
	s.invalidateSeatCache(ctx, res.EventID)

	return res, nil
}

// invalidateSeatCache は座席キャッシュを無効化する
func (s *ReservationService) invalidateSeatCache(ctx context.Context, eventID string) {
	if s.seatCache != nil {
		if err := s.seatCache.Invalidate(ctx, eventID); err != nil {
			logger.Warn("キャッシュ無効化エラー", zap.String("event_id", eventID), zap.Error(err))
		}
	}
}

// CancelExpiredReservations は期限切れの予約をキャンセルして座席を解放する
func (s *ReservationService) CancelExpiredReservations(ctx context.Context, expireAfter time.Duration) (int, error) {
	expired, err := s.reservationRepo.GetExpiredPending(ctx, expireAfter)
	if err != nil {
		return 0, fmt.Errorf("期限切れ予約取得に失敗: %w", err)
	}

	canceledCount := 0
	for _, res := range expired {
		log := logger.With(
			zap.String("reservation_id", res.ID),
			zap.String("event_id", res.EventID),
			zap.String("user_id", res.UserID),
		)

		if err := res.Cancel(); err != nil {
			log.Warn("期限切れ予約のキャンセルに失敗（ステータス変更）", zap.Error(err))
			continue
		}

		tx, err := s.db.BeginTxx(ctx, nil)
		if err != nil {
			log.Error("トランザクション開始に失敗", zap.Error(err))
			continue
		}

		if err := s.seatRepo.ReleaseSeats(ctx, tx, res.SeatIDs); err != nil {
			log.Error("座席解放に失敗", zap.Error(err))
			_ = tx.Rollback()
			continue
		}

		if err := s.reservationRepo.Update(ctx, tx, res); err != nil {
			log.Error("予約更新に失敗", zap.Error(err))
			_ = tx.Rollback()
			continue
		}

		if err := tx.Commit(); err != nil {
			log.Error("コミットに失敗", zap.Error(err))
			continue
		}

		log.Info("期限切れ予約をキャンセル")
		canceledCount++
	}

	return canceledCount, nil
}

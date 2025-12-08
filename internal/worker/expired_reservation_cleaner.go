package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/pkg/logger"
)

// ReservationCleaner は期限切れ予約をキャンセルするインターフェース
type ReservationCleaner interface {
	CancelExpiredReservations(ctx context.Context, expireAfter time.Duration) (int, error)
}

// ExpiredReservationCleaner は期限切れ予約をクリーンアップするワーカー
type ExpiredReservationCleaner struct {
	reservationService ReservationCleaner
	interval           time.Duration
	expireAfter        time.Duration
	stopCh             chan struct{}
	doneCh             chan struct{}
}

// NewExpiredReservationCleaner は新しいクリーナーを作成
func NewExpiredReservationCleaner(
	rs ReservationCleaner,
	interval time.Duration,
	expireAfter time.Duration,
) *ExpiredReservationCleaner {
	return &ExpiredReservationCleaner{
		reservationService: rs,
		interval:           interval,
		expireAfter:        expireAfter,
		stopCh:             make(chan struct{}),
		doneCh:             make(chan struct{}),
	}
}

// Start はクリーナーを開始
func (c *ExpiredReservationCleaner) Start(ctx context.Context) {
	logger.Info("期限切れ予約クリーナー開始",
		zap.Duration("interval", c.interval),
		zap.Duration("expire_after", c.expireAfter),
	)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	defer close(c.doneCh)

	for {
		select {
		case <-ctx.Done():
			logger.Info("期限切れ予約クリーナー停止（コンテキストキャンセル）")
			return
		case <-c.stopCh:
			logger.Info("期限切れ予約クリーナー停止（シグナル受信）")
			return
		case <-ticker.C:
			c.cleanup(ctx)
		}
	}
}

// Stop はクリーナーを停止
func (c *ExpiredReservationCleaner) Stop() {
	close(c.stopCh)
	<-c.doneCh
}

// cleanup は期限切れ予約をキャンセル
func (c *ExpiredReservationCleaner) cleanup(ctx context.Context) {
	log := logger.Get()
	log.Debug("期限切れ予約のクリーンアップ開始")

	count, err := c.reservationService.CancelExpiredReservations(ctx, c.expireAfter)
	if err != nil {
		log.Error("期限切れ予約のクリーンアップ失敗", zap.Error(err))
		return
	}

	if count > 0 {
		log.Info("期限切れ予約をキャンセル", zap.Int("count", count))
	} else {
		log.Debug("期限切れ予約なし")
	}
}

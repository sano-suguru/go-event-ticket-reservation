package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReservationCleaner はReservationCleanerのモック
type MockReservationCleaner struct {
	mock.Mock
}

func (m *MockReservationCleaner) CancelExpiredReservations(ctx context.Context, expireAfter time.Duration) (int, error) {
	args := m.Called(ctx, expireAfter)
	return args.Int(0), args.Error(1)
}

func TestNewExpiredReservationCleaner(t *testing.T) {
	mockService := new(MockReservationCleaner)
	interval := 1 * time.Minute
	expireAfter := 15 * time.Minute

	cleaner := NewExpiredReservationCleaner(mockService, interval, expireAfter)

	assert.NotNil(t, cleaner)
	assert.Equal(t, interval, cleaner.interval)
	assert.Equal(t, expireAfter, cleaner.expireAfter)
	assert.NotNil(t, cleaner.stopCh)
	assert.NotNil(t, cleaner.doneCh)
}

func TestExpiredReservationCleaner_StopChannels(t *testing.T) {
	mockService := new(MockReservationCleaner)
	cleaner := NewExpiredReservationCleaner(
		mockService,
		1*time.Second,
		15*time.Minute,
	)

	// チャンネルが初期化されていることを確認
	assert.NotNil(t, cleaner.stopCh)
	assert.NotNil(t, cleaner.doneCh)

	// チャンネルがブロッキングされていないことを確認（送信可能）
	select {
	case <-cleaner.stopCh:
		t.Fatal("stopCh should not be closed initially")
	default:
		// 期待通り
	}
}

func TestExpiredReservationCleaner_Cleanup(t *testing.T) {
	t.Run("正常にクリーンアップが実行される", func(t *testing.T) {
		mockService := new(MockReservationCleaner)
		mockService.On("CancelExpiredReservations", mock.Anything, 15*time.Minute).Return(5, nil)

		cleaner := &ExpiredReservationCleaner{
			reservationService: mockService,
			interval:           1 * time.Minute,
			expireAfter:        15 * time.Minute,
			stopCh:             make(chan struct{}),
			doneCh:             make(chan struct{}),
		}

		cleaner.cleanup(context.Background())

		mockService.AssertExpectations(t)
	})

	t.Run("キャンセル対象がない場合も正常に動作する", func(t *testing.T) {
		mockService := new(MockReservationCleaner)
		mockService.On("CancelExpiredReservations", mock.Anything, 15*time.Minute).Return(0, nil)

		cleaner := &ExpiredReservationCleaner{
			reservationService: mockService,
			interval:           1 * time.Minute,
			expireAfter:        15 * time.Minute,
			stopCh:             make(chan struct{}),
			doneCh:             make(chan struct{}),
		}

		cleaner.cleanup(context.Background())

		mockService.AssertExpectations(t)
	})

	t.Run("エラーが発生しても継続する", func(t *testing.T) {
		mockService := new(MockReservationCleaner)
		mockService.On("CancelExpiredReservations", mock.Anything, 15*time.Minute).Return(0, assert.AnError)

		cleaner := &ExpiredReservationCleaner{
			reservationService: mockService,
			interval:           1 * time.Minute,
			expireAfter:        15 * time.Minute,
			stopCh:             make(chan struct{}),
			doneCh:             make(chan struct{}),
		}

		// パニックしないことを確認
		cleaner.cleanup(context.Background())

		mockService.AssertExpectations(t)
	})
}

func TestExpiredReservationCleaner_StartStop(t *testing.T) {
	t.Run("開始と停止が正常に動作する", func(t *testing.T) {
		mockService := new(MockReservationCleaner)
		// cleanup が呼ばれる可能性があるので、任意回数マッチさせる
		mockService.On("CancelExpiredReservations", mock.Anything, 100*time.Millisecond).Return(0, nil).Maybe()

		cleaner := NewExpiredReservationCleaner(mockService, 50*time.Millisecond, 100*time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// バックグラウンドで開始
		go cleaner.Start(ctx)

		// 少し待機
		time.Sleep(120 * time.Millisecond)

		// 停止
		cleaner.Stop()

		// Stop後はdoneChがcloseされている
		select {
		case <-cleaner.doneCh:
			// 正常に終了
		case <-time.After(1 * time.Second):
			t.Error("cleaner did not stop in time")
		}
	})

	t.Run("コンテキストキャンセルで停止する", func(t *testing.T) {
		mockService := new(MockReservationCleaner)
		mockService.On("CancelExpiredReservations", mock.Anything, 100*time.Millisecond).Return(0, nil).Maybe()

		cleaner := NewExpiredReservationCleaner(mockService, 50*time.Millisecond, 100*time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			cleaner.Start(ctx)
			close(done)
		}()

		// 少し待機してからコンテキストをキャンセル
		time.Sleep(80 * time.Millisecond)
		cancel()

		// 終了を待機
		select {
		case <-done:
			// 正常に終了
		case <-time.After(1 * time.Second):
			t.Error("cleaner did not stop after context cancel")
		}
	})
}

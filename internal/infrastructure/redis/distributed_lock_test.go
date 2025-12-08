package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockManager_AcquireLock(t *testing.T) {
	client, err := NewClient(&Config{Host: "localhost", Port: "6379"})
	if err != nil {
		t.Skip("Redis not available")
	}
	defer client.Close()

	ctx := context.Background()
	manager := NewLockManager(client)

	t.Run("ロックを取得できる", func(t *testing.T) {
		lock, err := manager.AcquireLock(ctx, "test-key-1", 5*time.Second)
		require.NoError(t, err)
		require.NotNil(t, lock)
		defer lock.Release(ctx)
	})

	t.Run("同じキーのロックは取得できない", func(t *testing.T) {
		lock1, err := manager.AcquireLock(ctx, "test-key-2", 5*time.Second)
		require.NoError(t, err)
		defer lock1.Release(ctx)

		lock2, err := manager.AcquireLock(ctx, "test-key-2", 5*time.Second)
		assert.ErrorIs(t, err, ErrLockNotAcquired)
		assert.Nil(t, lock2)
	})

	t.Run("解放後は再取得できる", func(t *testing.T) {
		lock1, err := manager.AcquireLock(ctx, "test-key-3", 5*time.Second)
		require.NoError(t, err)

		err = lock1.Release(ctx)
		require.NoError(t, err)

		lock2, err := manager.AcquireLock(ctx, "test-key-3", 5*time.Second)
		require.NoError(t, err)
		defer lock2.Release(ctx)
	})

	t.Run("リトライで取得できる", func(t *testing.T) {
		lock1, err := manager.AcquireLock(ctx, "test-key-4", 500*time.Millisecond)
		require.NoError(t, err)

		go func() {
			time.Sleep(300 * time.Millisecond)
			lock1.Release(ctx)
		}()

		lock2, err := manager.AcquireLockWithRetry(ctx, "test-key-4", 5*time.Second, 5, 100*time.Millisecond)
		require.NoError(t, err)
		defer lock2.Release(ctx)
	})

	t.Run("ロックを延長できる", func(t *testing.T) {
		lock, err := manager.AcquireLock(ctx, "test-key-extend", 1*time.Second)
		require.NoError(t, err)
		defer lock.Release(ctx)

		// ロックを延長
		err = lock.Extend(ctx, 5*time.Second)
		require.NoError(t, err)

		// まだロックを持っていることを確認
		lock2, err := manager.AcquireLock(ctx, "test-key-extend", 1*time.Second)
		assert.ErrorIs(t, err, ErrLockNotAcquired)
		assert.Nil(t, lock2)
	})

	t.Run("解放後は延長できない", func(t *testing.T) {
		lock, err := manager.AcquireLock(ctx, "test-key-extend-after-release", 1*time.Second)
		require.NoError(t, err)

		err = lock.Release(ctx)
		require.NoError(t, err)

		// 解放後に延長を試みる
		err = lock.Extend(ctx, 5*time.Second)
		assert.ErrorIs(t, err, ErrLockNotOwned)
	})
}

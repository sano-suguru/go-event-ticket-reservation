package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) *redis.Client {
	client, err := NewClient(&Config{Host: "localhost", Port: "6379"})
	if err != nil {
		t.Skip("Redis not available")
	}
	t.Cleanup(func() { client.Close() })
	return client
}

func TestSeatCache_GetAvailableCount(t *testing.T) {
	client := setupTestRedis(t)
	cache := NewSeatCache(client)
	ctx := context.Background()
	eventID := "test-event-123"

	t.Run("キャッシュミス時はErrCacheMissを返す", func(t *testing.T) {
		_, err := cache.GetAvailableCount(ctx, eventID)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("キャッシュにセットした値を取得できる", func(t *testing.T) {
		err := cache.SetAvailableCount(ctx, eventID, 100, 30*time.Second)
		require.NoError(t, err)

		count, err := cache.GetAvailableCount(ctx, eventID)
		require.NoError(t, err)
		assert.Equal(t, 100, count)
	})

	t.Run("キャッシュを無効化できる", func(t *testing.T) {
		err := cache.SetAvailableCount(ctx, eventID, 50, 30*time.Second)
		require.NoError(t, err)

		err = cache.Invalidate(ctx, eventID)
		require.NoError(t, err)

		_, err = cache.GetAvailableCount(ctx, eventID)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})
}

func TestSeatCache_TTL(t *testing.T) {
	client := setupTestRedis(t)
	cache := NewSeatCache(client)
	ctx := context.Background()
	eventID := "test-event-ttl"

	t.Run("TTL経過後はキャッシュミスになる", func(t *testing.T) {
		err := cache.SetAvailableCount(ctx, eventID, 100, 100*time.Millisecond)
		require.NoError(t, err)

		// TTL経過前
		count, err := cache.GetAvailableCount(ctx, eventID)
		require.NoError(t, err)
		assert.Equal(t, 100, count)

		// TTL経過後
		time.Sleep(150 * time.Millisecond)
		_, err = cache.GetAvailableCount(ctx, eventID)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})
}

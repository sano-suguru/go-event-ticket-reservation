package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCacheMiss = errors.New("キャッシュが見つかりません")
)

// SeatCache は座席情報のキャッシュを管理する
type SeatCache struct {
	client *redis.Client
}

// NewSeatCache は新しいSeatCacheインスタンスを作成する
func NewSeatCache(client *redis.Client) *SeatCache {
	return &SeatCache{client: client}
}

// GetAvailableCount はイベントの空席数をキャッシュから取得する
func (c *SeatCache) GetAvailableCount(ctx context.Context, eventID string) (int, error) {
	key := c.availableCountKey(eventID)
	val, err := c.client.Get(ctx, key).Int()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ErrCacheMiss
		}
		return 0, fmt.Errorf("キャッシュ取得に失敗: %w", err)
	}
	return val, nil
}

// SetAvailableCount はイベントの空席数をキャッシュに保存する
func (c *SeatCache) SetAvailableCount(ctx context.Context, eventID string, count int, ttl time.Duration) error {
	key := c.availableCountKey(eventID)
	err := c.client.Set(ctx, key, count, ttl).Err()
	if err != nil {
		return fmt.Errorf("キャッシュ保存に失敗: %w", err)
	}
	return nil
}

// Invalidate はイベントのキャッシュを無効化する
func (c *SeatCache) Invalidate(ctx context.Context, eventID string) error {
	key := c.availableCountKey(eventID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("キャッシュ無効化に失敗: %w", err)
	}
	return nil
}

func (c *SeatCache) availableCountKey(eventID string) string {
	return fmt.Sprintf("seats:available:%s", eventID)
}

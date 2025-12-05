package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
)

// NewClient はRedisクライアントを作成する
func NewClient(cfg *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// Ping はRedis接続を確認する
func Ping(ctx context.Context, client *redis.Client) error {
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis接続に失敗しました: %w", err)
	}
	return nil
}

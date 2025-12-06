package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrLockNotAcquired = errors.New("ロックを取得できませんでした")
	ErrLockNotOwned    = errors.New("ロックの所有者ではありません")
)

// DistributedLock は Redis を使用した分散ロック
type DistributedLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

// LockManager は分散ロックを管理する
type LockManager struct {
	client *redis.Client
}

func NewLockManager(client *redis.Client) *LockManager {
	return &LockManager{client: client}
}

// AcquireLock はロックを取得する
func (m *LockManager) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*DistributedLock, error) {
	lockKey := fmt.Sprintf("lock:%s", key)
	lockValue := uuid.New().String()

	// SetNX を使用してロックを取得（キーが存在しない場合のみ設定）
	ok, err := m.client.SetNX(ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("ロック取得に失敗: %w", err)
	}
	if !ok {
		return nil, ErrLockNotAcquired
	}

	return &DistributedLock{
		client: m.client,
		key:    lockKey,
		value:  lockValue,
		ttl:    ttl,
	}, nil
}

// AcquireLockWithRetry はリトライ付きでロックを取得する
func (m *LockManager) AcquireLockWithRetry(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryDelay time.Duration) (*DistributedLock, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		lock, err := m.AcquireLock(ctx, key, ttl)
		if err == nil {
			return lock, nil
		}
		lastErr = err
		if !errors.Is(err, ErrLockNotAcquired) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryDelay):
		}
	}
	return nil, lastErr
}

// Release はロックを解放する（Lua スクリプトで安全に解放）
func (l *DistributedLock) Release(ctx context.Context) error {
	// Lua スクリプトで所有者確認と削除をアトミックに実行
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Int()
	if err != nil {
		return fmt.Errorf("ロック解放に失敗: %w", err)
	}
	if result == 0 {
		return ErrLockNotOwned
	}
	return nil
}

// Extend はロックの有効期限を延長する
func (l *DistributedLock) Extend(ctx context.Context, ttl time.Duration) error {
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, ttl.Milliseconds()).Int()
	if err != nil {
		return fmt.Errorf("ロック延長に失敗: %w", err)
	}
	if result == 0 {
		return ErrLockNotOwned
	}
	l.ttl = ttl
	return nil
}

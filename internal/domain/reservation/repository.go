package reservation

import (
	"context"
	"time"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/transaction"
)

// Repository は予約リポジトリのインターフェース
type Repository interface {
	// Create は新しい予約を作成する（トランザクション必須）
	Create(ctx context.Context, tx transaction.Tx, reservation *Reservation) error

	// GetByID はIDから予約を取得する
	GetByID(ctx context.Context, id string) (*Reservation, error)

	// GetByIdempotencyKey は冪等性キーから予約を取得する
	GetByIdempotencyKey(ctx context.Context, key string) (*Reservation, error)

	// GetByUserID はユーザーIDから予約一覧を取得する
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Reservation, error)

	// Update は予約を更新する（トランザクション必須）
	Update(ctx context.Context, tx transaction.Tx, reservation *Reservation) error

	// GetExpiredPending は期限切れの保留中予約を取得する
	GetExpiredPending(ctx context.Context, expireAfter time.Duration) ([]*Reservation, error)
}

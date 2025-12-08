package seat

import (
	"context"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/transaction"
)

// Repository は座席リポジトリのインターフェース
type Repository interface {
	// Create は新しい座席を作成する
	Create(ctx context.Context, seat *Seat) error

	// CreateBulk は複数の座席を一括作成する
	CreateBulk(ctx context.Context, seats []*Seat) error

	// GetByID はIDから座席を取得する
	GetByID(ctx context.Context, id string) (*Seat, error)

	// GetByEventID はイベントIDから座席一覧を取得する
	GetByEventID(ctx context.Context, eventID string) ([]*Seat, error)

	// GetAvailableByEventID はイベントIDから利用可能な座席一覧を取得する
	GetAvailableByEventID(ctx context.Context, eventID string) ([]*Seat, error)

	// ReserveSeats は座席を予約状態に更新する（楽観的ロック、トランザクション必須）
	ReserveSeats(ctx context.Context, tx transaction.Tx, seatIDs []string, reservationID string) error

	// ConfirmSeats は座席を確定状態に更新する（トランザクション必須）
	ConfirmSeats(ctx context.Context, tx transaction.Tx, seatIDs []string) error

	// ReleaseSeats は座席を解放する（トランザクション必須）
	ReleaseSeats(ctx context.Context, tx transaction.Tx, seatIDs []string) error

	// CountAvailableByEventID はイベントの利用可能座席数を取得する
	CountAvailableByEventID(ctx context.Context, eventID string) (int, error)
}

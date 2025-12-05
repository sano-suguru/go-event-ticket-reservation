package event

import "context"

// Repository はイベントリポジトリのインターフェース
type Repository interface {
	// Create は新しいイベントを作成する
	Create(ctx context.Context, event *Event) error

	// GetByID はIDからイベントを取得する
	GetByID(ctx context.Context, id string) (*Event, error)

	// List はイベント一覧を取得する
	List(ctx context.Context, limit, offset int) ([]*Event, error)

	// Update はイベントを更新する（楽観的ロック）
	Update(ctx context.Context, event *Event) error

	// Delete はイベントを削除する
	Delete(ctx context.Context, id string) error
}

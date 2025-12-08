package transaction

import "context"

// Tx はトランザクションを表すインターフェース
// ドメイン層がインフラ層（sqlx等）に依存しないようにするための抽象化
type Tx interface {
	// Commit はトランザクションをコミットする
	Commit() error
	// Rollback はトランザクションをロールバックする
	Rollback() error
}

// Manager はトランザクションを管理するインターフェース
type Manager interface {
	// Begin は新しいトランザクションを開始する
	Begin(ctx context.Context) (Tx, error)
}

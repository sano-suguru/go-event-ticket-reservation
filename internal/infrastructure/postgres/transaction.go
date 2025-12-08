package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/transaction"
)

// TxWrapper は sqlx.Tx を transaction.Tx インターフェースでラップする
type TxWrapper struct {
	*sqlx.Tx
}

// Commit はトランザクションをコミットする
func (t *TxWrapper) Commit() error {
	return t.Tx.Commit()
}

// Rollback はトランザクションをロールバックする
func (t *TxWrapper) Rollback() error {
	return t.Tx.Rollback()
}

// TxManager は sqlx.DB を使用したトランザクションマネージャー
type TxManager struct {
	db *sqlx.DB
}

// NewTxManager は新しい TxManager を作成する
func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

// Begin は新しいトランザクションを開始する
func (m *TxManager) Begin(ctx context.Context) (transaction.Tx, error) {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &TxWrapper{Tx: tx}, nil
}

// UnwrapTx は transaction.Tx から sqlx.Tx を取り出す
// リポジトリ実装で使用する
func UnwrapTx(tx transaction.Tx) *sqlx.Tx {
	if wrapper, ok := tx.(*TxWrapper); ok {
		return wrapper.Tx
	}
	return nil
}

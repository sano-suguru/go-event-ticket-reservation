package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/config"
)

// NewConnection はPostgreSQLへの接続を作成する
func NewConnection(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("データベース接続に失敗しました: %w", err)
	}

	// 接続プール設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// Ping はデータベース接続を確認する
func Ping(ctx context.Context, db *sqlx.DB) error {
	return db.PingContext(ctx)
}

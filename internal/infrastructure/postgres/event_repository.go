package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/event"
)

// eventRow はDBの行を表す構造体
type eventRow struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Venue       *string   `db:"venue"`
	StartAt     time.Time `db:"start_at"`
	EndAt       time.Time `db:"end_at"`
	TotalSeats  int       `db:"total_seats"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Version     int       `db:"version"`
}

// toEntity はeventRowをEventエンティティに変換する
func (r *eventRow) toEntity() *event.Event {
	var desc, venue string
	if r.Description != nil {
		desc = *r.Description
	}
	if r.Venue != nil {
		venue = *r.Venue
	}
	return &event.Event{
		ID:          r.ID,
		Name:        r.Name,
		Description: desc,
		Venue:       venue,
		StartAt:     r.StartAt,
		EndAt:       r.EndAt,
		TotalSeats:  r.TotalSeats,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		Version:     r.Version,
	}
}

// EventRepository はイベントリポジトリのPostgreSQL実装
type EventRepository struct {
	db *sqlx.DB
}

// NewEventRepository はEventRepositoryを作成する
func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create は新しいイベントを作成する
func (r *EventRepository) Create(ctx context.Context, e *event.Event) error {
	query := `
		INSERT INTO events (name, description, venue, start_at, end_at, total_seats, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	var desc, venue *string
	if e.Description != "" {
		desc = &e.Description
	}
	if e.Venue != "" {
		venue = &e.Venue
	}

	err := r.db.QueryRowContext(ctx, query,
		e.Name, desc, venue, e.StartAt, e.EndAt, e.TotalSeats, e.CreatedAt, e.UpdatedAt, e.Version,
	).Scan(&e.ID)
	if err != nil {
		return fmt.Errorf("イベント作成に失敗しました: %w", err)
	}
	return nil
}

// GetByID はIDからイベントを取得する
func (r *EventRepository) GetByID(ctx context.Context, id string) (*event.Event, error) {
	query := `SELECT id, name, description, venue, start_at, end_at, total_seats, created_at, updated_at, version FROM events WHERE id = $1`

	var row eventRow
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, event.ErrEventNotFound
		}
		return nil, fmt.Errorf("イベント取得に失敗しました: %w", err)
	}
	return row.toEntity(), nil
}

// List はイベント一覧を取得する
func (r *EventRepository) List(ctx context.Context, limit, offset int) ([]*event.Event, error) {
	query := `
		SELECT id, name, description, venue, start_at, end_at, total_seats, created_at, updated_at, version 
		FROM events 
		ORDER BY start_at DESC 
		LIMIT $1 OFFSET $2
	`

	var rows []eventRow
	err := r.db.SelectContext(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("イベント一覧取得に失敗しました: %w", err)
	}

	events := make([]*event.Event, len(rows))
	for i, row := range rows {
		events[i] = row.toEntity()
	}
	return events, nil
}

// Update はイベントを更新する（楽観的ロック）
func (r *EventRepository) Update(ctx context.Context, e *event.Event) error {
	query := `
		UPDATE events
		SET name = $1, description = $2, venue = $3, start_at = $4, end_at = $5, 
		    total_seats = $6, updated_at = $7, version = version + 1
		WHERE id = $8 AND version = $9
	`

	var desc, venue *string
	if e.Description != "" {
		desc = &e.Description
	}
	if e.Venue != "" {
		venue = &e.Venue
	}

	result, err := r.db.ExecContext(ctx, query,
		e.Name, desc, venue, e.StartAt, e.EndAt, e.TotalSeats, time.Now(), e.ID, e.Version,
	)
	if err != nil {
		return fmt.Errorf("イベント更新に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("更新結果の確認に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return event.ErrEventNotFound
	}

	e.Version++
	return nil
}

// Delete はイベントを削除する
func (r *EventRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("イベント削除に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("削除結果の確認に失敗しました: %w", err)
	}
	if rowsAffected == 0 {
		return event.ErrEventNotFound
	}
	return nil
}

// インターフェースを満たしているか確認
var _ event.Repository = (*EventRepository)(nil)

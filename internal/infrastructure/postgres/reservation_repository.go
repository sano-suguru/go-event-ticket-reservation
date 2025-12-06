package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/reservation"
)

type reservationRow struct {
	ID             string     `db:"id"`
	EventID        string     `db:"event_id"`
	UserID         string     `db:"user_id"`
	Status         string     `db:"status"`
	IdempotencyKey string     `db:"idempotency_key"`
	TotalAmount    int        `db:"total_amount"`
	ExpiresAt      time.Time  `db:"expires_at"`
	ConfirmedAt    *time.Time `db:"confirmed_at"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

type ReservationRepository struct{ db *sqlx.DB }

func NewReservationRepository(db *sqlx.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(ctx context.Context, tx *sqlx.Tx, res *reservation.Reservation) error {
	query := `INSERT INTO reservations (event_id, user_id, status, idempotency_key, total_amount, expires_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	if err := tx.QueryRowContext(ctx, query, res.EventID, res.UserID, string(res.Status), res.IdempotencyKey, res.TotalAmount, res.ExpiresAt, res.CreatedAt, res.UpdatedAt).Scan(&res.ID); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return reservation.ErrIdempotencyKeyAlreadyExists
		}
		return fmt.Errorf("予約作成に失敗: %w", err)
	}
	if len(res.SeatIDs) > 0 {
		for _, seatID := range res.SeatIDs {
			if _, err := tx.ExecContext(ctx, `INSERT INTO reservation_seats (reservation_id, seat_id) VALUES ($1, $2)`, res.ID, seatID); err != nil {
				return fmt.Errorf("予約座席関連付けに失敗: %w", err)
			}
		}
	}
	return nil
}

func (r *ReservationRepository) GetByID(ctx context.Context, id string) (*reservation.Reservation, error) {
	var row reservationRow
	query := `SELECT id, event_id, user_id, status, idempotency_key, total_amount, expires_at, confirmed_at, created_at, updated_at FROM reservations WHERE id = $1`
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, reservation.ErrReservationNotFound
		}
		return nil, fmt.Errorf("予約取得に失敗: %w", err)
	}
	seatIDs, err := r.getSeatIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	return r.toEntity(&row, seatIDs), nil
}

func (r *ReservationRepository) GetByIdempotencyKey(ctx context.Context, key string) (*reservation.Reservation, error) {
	var row reservationRow
	if err := r.db.GetContext(ctx, &row, `SELECT id, event_id, user_id, status, idempotency_key, total_amount, expires_at, confirmed_at, created_at, updated_at FROM reservations WHERE idempotency_key = $1`, key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, reservation.ErrReservationNotFound
		}
		return nil, fmt.Errorf("予約取得に失敗: %w", err)
	}
	seatIDs, err := r.getSeatIDs(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return r.toEntity(&row, seatIDs), nil
}

func (r *ReservationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*reservation.Reservation, error) {
	var rows []reservationRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, event_id, user_id, status, idempotency_key, total_amount, expires_at, confirmed_at, created_at, updated_at FROM reservations WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("予約一覧取得に失敗: %w", err)
	}
	result := make([]*reservation.Reservation, len(rows))
	for i, row := range rows {
		seatIDs, err := r.getSeatIDs(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		result[i] = r.toEntity(&row, seatIDs)
	}
	return result, nil
}

func (r *ReservationRepository) Update(ctx context.Context, tx *sqlx.Tx, res *reservation.Reservation) error {
	query := `UPDATE reservations SET status = $1, confirmed_at = $2, updated_at = $3 WHERE id = $4`
	result, err := tx.ExecContext(ctx, query, string(res.Status), res.ConfirmedAt, res.UpdatedAt, res.ID)
	if err != nil {
		return fmt.Errorf("予約更新に失敗: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return reservation.ErrReservationNotFound
	}
	return nil
}

func (r *ReservationRepository) GetExpiredPending(ctx context.Context) ([]*reservation.Reservation, error) {
	var rows []reservationRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, event_id, user_id, status, idempotency_key, total_amount, expires_at, confirmed_at, created_at, updated_at FROM reservations WHERE status = 'pending' AND expires_at < NOW()`); err != nil {
		return nil, fmt.Errorf("期限切れ予約取得に失敗: %w", err)
	}
	result := make([]*reservation.Reservation, len(rows))
	for i, row := range rows {
		seatIDs, err := r.getSeatIDs(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		result[i] = r.toEntity(&row, seatIDs)
	}
	return result, nil
}

func (r *ReservationRepository) getSeatIDs(ctx context.Context, reservationID string) ([]string, error) {
	var seatIDs []string
	if err := r.db.SelectContext(ctx, &seatIDs, `SELECT seat_id FROM reservation_seats WHERE reservation_id = $1`, reservationID); err != nil {
		return nil, fmt.Errorf("座席ID取得に失敗: %w", err)
	}
	return seatIDs, nil
}

func (r *ReservationRepository) toEntity(row *reservationRow, seatIDs []string) *reservation.Reservation {
	return &reservation.Reservation{
		ID: row.ID, EventID: row.EventID, UserID: row.UserID,
		SeatIDs: seatIDs, Status: reservation.Status(row.Status),
		IdempotencyKey: row.IdempotencyKey, TotalAmount: row.TotalAmount,
		ExpiresAt: row.ExpiresAt, ConfirmedAt: row.ConfirmedAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

var _ reservation.Repository = (*ReservationRepository)(nil)

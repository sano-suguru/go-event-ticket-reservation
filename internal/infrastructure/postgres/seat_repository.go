package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/sanosuguru/go-event-ticket-reservation/internal/domain/seat"
)

type seatRow struct {
	ID         string     `db:"id"`
	EventID    string     `db:"event_id"`
	SeatNumber string     `db:"seat_number"`
	Status     string     `db:"status"`
	Price      int        `db:"price"`
	ReservedBy *string    `db:"reserved_by"`
	ReservedAt *time.Time `db:"reserved_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	Version    int        `db:"version"`
}

func (r *seatRow) toEntity() *seat.Seat {
	return &seat.Seat{
		ID: r.ID, EventID: r.EventID, SeatNumber: r.SeatNumber,
		Status: seat.Status(r.Status), Price: r.Price,
		ReservedBy: r.ReservedBy, ReservedAt: r.ReservedAt,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, Version: r.Version,
	}
}

type SeatRepository struct{ db *sqlx.DB }

func NewSeatRepository(db *sqlx.DB) *SeatRepository { return &SeatRepository{db: db} }

func (r *SeatRepository) Create(ctx context.Context, s *seat.Seat) error {
	query := `INSERT INTO seats (event_id, seat_number, status, price, created_at, updated_at, version) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	return r.db.QueryRowContext(ctx, query, s.EventID, s.SeatNumber, string(s.Status), s.Price, s.CreatedAt, s.UpdatedAt, s.Version).Scan(&s.ID)
}

func (r *SeatRepository) CreateBulk(ctx context.Context, seats []*seat.Seat) error {
	if len(seats) == 0 {
		return nil
	}

	// バッチサイズごとに分割してマルチバリューINSERTを実行
	const batchSize = 1000
	for i := 0; i < len(seats); i += batchSize {
		end := i + batchSize
		if end > len(seats) {
			end = len(seats)
		}
		batch := seats[i:end]

		if err := r.createBulkBatch(ctx, batch); err != nil {
			return err
		}
	}
	return nil
}

// createBulkBatch はバッチ単位でマルチバリューINSERTを実行
func (r *SeatRepository) createBulkBatch(ctx context.Context, seats []*seat.Seat) error {
	if len(seats) == 0 {
		return nil
	}

	// マルチバリューINSERTを構築
	query := `INSERT INTO seats (event_id, seat_number, status, price, created_at, updated_at, version) VALUES `
	args := make([]interface{}, 0, len(seats)*7)
	placeholders := make([]string, 0, len(seats))

	for i, s := range seats {
		base := i * 7
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7))
		args = append(args, s.EventID, s.SeatNumber, string(s.Status), s.Price, s.CreatedAt, s.UpdatedAt, s.Version)
	}

	query += strings.Join(placeholders, ", ")
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("座席一括作成に失敗: %w", err)
	}
	return nil
}

func (r *SeatRepository) GetByID(ctx context.Context, id string) (*seat.Seat, error) {
	query := `SELECT id, event_id, seat_number, status, price, reserved_by, reserved_at, created_at, updated_at, version FROM seats WHERE id = $1`
	var row seatRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, seat.ErrSeatNotFound
		}
		return nil, fmt.Errorf("座席取得に失敗: %w", err)
	}
	return row.toEntity(), nil
}

func (r *SeatRepository) GetByEventID(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	query := `SELECT id, event_id, seat_number, status, price, reserved_by, reserved_at, created_at, updated_at, version FROM seats WHERE event_id = $1 ORDER BY seat_number`
	var rows []seatRow
	if err := r.db.SelectContext(ctx, &rows, query, eventID); err != nil {
		return nil, err
	}
	seats := make([]*seat.Seat, len(rows))
	for i, row := range rows {
		seats[i] = row.toEntity()
	}
	return seats, nil
}

func (r *SeatRepository) GetAvailableByEventID(ctx context.Context, eventID string) ([]*seat.Seat, error) {
	query := `SELECT id, event_id, seat_number, status, price, reserved_by, reserved_at, created_at, updated_at, version FROM seats WHERE event_id = $1 AND status = 'available' ORDER BY seat_number`
	var rows []seatRow
	if err := r.db.SelectContext(ctx, &rows, query, eventID); err != nil {
		return nil, err
	}
	seats := make([]*seat.Seat, len(rows))
	for i, row := range rows {
		seats[i] = row.toEntity()
	}
	return seats, nil
}

func (r *SeatRepository) ReserveSeats(ctx context.Context, tx *sqlx.Tx, seatIDs []string, reservationID string) error {
	if len(seatIDs) == 0 {
		return nil
	}
	query := `UPDATE seats SET status = 'reserved', reserved_by = $1, reserved_at = NOW(), updated_at = NOW(), version = version + 1 WHERE id = ANY($2) AND status = 'available'`
	result, err := tx.ExecContext(ctx, query, reservationID, pq.Array(seatIDs))
	if err != nil {
		return fmt.Errorf("座席予約に失敗: %w", err)
	}
	rows, _ := result.RowsAffected()
	if int(rows) != len(seatIDs) {
		return seat.ErrSeatAlreadyReserved
	}
	return nil
}

func (r *SeatRepository) ConfirmSeats(ctx context.Context, tx *sqlx.Tx, seatIDs []string) error {
	if len(seatIDs) == 0 {
		return nil
	}
	query := `UPDATE seats SET status = 'confirmed', updated_at = NOW(), version = version + 1 WHERE id = ANY($1) AND status = 'reserved'`
	result, err := tx.ExecContext(ctx, query, pq.Array(seatIDs))
	if err != nil {
		return fmt.Errorf("座席確定に失敗: %w", err)
	}
	rows, _ := result.RowsAffected()
	if int(rows) != len(seatIDs) {
		return seat.ErrSeatNotReserved
	}
	return nil
}

func (r *SeatRepository) ReleaseSeats(ctx context.Context, tx *sqlx.Tx, seatIDs []string) error {
	if len(seatIDs) == 0 {
		return nil
	}
	query := `UPDATE seats SET status = 'available', reserved_by = NULL, reserved_at = NULL, updated_at = NOW(), version = version + 1 WHERE id = ANY($1)`
	_, err := tx.ExecContext(ctx, query, pq.Array(seatIDs))
	return err
}

func (r *SeatRepository) CountAvailableByEventID(ctx context.Context, eventID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM seats WHERE event_id = $1 AND status = 'available'`, eventID)
	return count, err
}

var _ seat.Repository = (*SeatRepository)(nil)

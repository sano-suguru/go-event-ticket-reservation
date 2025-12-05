-- 外部キー制約を削除
ALTER TABLE seats DROP CONSTRAINT IF EXISTS fk_seats_reserved_by;

-- インデックスを削除
DROP INDEX IF EXISTS idx_reservation_seats_seat;
DROP INDEX IF EXISTS idx_reservations_idempotency;
DROP INDEX IF EXISTS idx_reservations_expires;
DROP INDEX IF EXISTS idx_reservations_user;
DROP INDEX IF EXISTS idx_seats_event_status;

-- テーブルを削除（依存関係の順序に注意）
DROP TABLE IF EXISTS reservation_seats;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS seats;
DROP TABLE IF EXISTS events;

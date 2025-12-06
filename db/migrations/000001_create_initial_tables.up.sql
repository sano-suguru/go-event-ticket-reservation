-- events テーブル
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    venue VARCHAR(255),
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP NOT NULL,
    total_seats INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 0
);

-- seats テーブル
CREATE TABLE seats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    seat_number VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    price INTEGER NOT NULL,
    reserved_by UUID,
    reserved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 0,
    UNIQUE(event_id, seat_number)
);

-- reservations テーブル
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id),
    user_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    idempotency_key VARCHAR(255) UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    confirmed_at TIMESTAMP,
    total_amount INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- reservation_seats テーブル（中間テーブル）
CREATE TABLE reservation_seats (
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    seat_id UUID NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
    PRIMARY KEY (reservation_id, seat_id)
);

-- インデックス
CREATE INDEX idx_seats_event_status ON seats(event_id, status);
CREATE INDEX idx_reservations_user ON reservations(user_id);
CREATE INDEX idx_reservations_expires ON reservations(expires_at) WHERE status = 'pending';
CREATE INDEX idx_reservations_idempotency ON reservations(idempotency_key);
CREATE INDEX idx_reservation_seats_seat ON reservation_seats(seat_id);

-- seats.reserved_by の外部キー制約（reservations 作成後に追加）
ALTER TABLE seats 
    ADD CONSTRAINT fk_seats_reserved_by 
    FOREIGN KEY (reserved_by) REFERENCES reservations(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    slot_id UUID NOT NULL REFERENCES slots(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_unique_active_booking ON bookings(slot_id) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);

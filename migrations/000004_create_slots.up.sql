CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    is_booked BOOLEAN NOT NULL DEFAULT FALSE
);

ALTER TABLE slots ADD CONSTRAINT unique_room_slot UNIQUE (room_id, start_at);

CREATE INDEX IF NOT EXISTS idx_slots_available ON slots (room_id, start_at) WHERE is_booked = FALSE;

CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    days_of_week INTEGER[] NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,

    CONSTRAINT check_days_range CHECK (days_of_week <@ ARRAY[1,2,3,4,5,6,7]),
    CONSTRAINT check_time_range CHECK (start_time < end_time)
);

ALTER TABLE schedules ADD CONSTRAINT unique_room_id UNIQUE (room_id);

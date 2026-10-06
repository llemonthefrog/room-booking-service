BEGIN;

UPDATE bookings SET status = 'created' WHERE status = 'active';
ALTER TABLE bookings ALTER COLUMN status SET DEFAULT 'created';
ALTER TABLE bookings ADD CONSTRAINT check_booking_status
    CHECK (status IN ('created', 'cancelled'));

DROP INDEX idx_unique_active_booking;
CREATE UNIQUE INDEX idx_unique_active_booking ON bookings(slot_id) WHERE status = 'created';

COMMIT;

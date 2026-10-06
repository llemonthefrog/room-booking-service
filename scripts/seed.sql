INSERT INTO users (id, email, password_hash, role)
VALUES ('00000000-0000-0000-0000-000000000001', 'admin@avito.ru', 'hash', 'admin')
ON CONFLICT DO NOTHING;

INSERT INTO rooms (id, name)
VALUES
    ('00000000-0000-0000-0000-000000000002', 'Red Room'),
    ('00000000-0000-0000-0000-000000000003', 'Blue Room')
ON CONFLICT DO NOTHING;

INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
VALUES (
        '00000000-0000-0000-0000-000000000004',
        '00000000-0000-0000-0000-000000000002',
        ARRAY[1,2,3,4,5],
        '09:00:00',
        '18:00:00')
ON CONFLICT DO NOTHING;

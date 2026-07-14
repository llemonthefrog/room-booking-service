INSERT INTO users (id, email, password, role)
VALUES ('00000000-0000-0000-0000-000000000001', 'admin@avito.ru', 'hash', 'admin');

INSERT INTO rooms (id, name)
VALUES
    ('00000000-0000-0000-0000-000000000002', 'Red Room'),
    ('00000000-0000-0000-0000-000000000003', 'Blue Room');

INSERT INTO schedules (id, room_id, day_of_week, start_time, end_time)
VALUES (
        '00000000-0000-0000-0000-000000000004',
        '00000000-0000-0000-0000-000000000002',
        1,
        '09:00:00',
        '18:00:00');

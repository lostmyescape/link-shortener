INSERT INTO apps (id, name, secret)
VALUES (1, 'test', 'secret-test')
ON CONFLICT DO NOTHING;
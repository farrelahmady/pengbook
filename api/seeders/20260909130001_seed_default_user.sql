-- +goose Up
-- Seed: Default admin user
-- Password: admin123

INSERT INTO users (name, username, email, password_hash)
VALUES (
    'Admin',
    'admin',
    'admin@pengbook.id',
    '$2a$10$xO8/Pf4/dytDSx9DTM6qW.WLnTi2GlgzVS56S9aPTopBzauuQ7sTS'
)
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM users WHERE email = 'admin@pengbook.id';

-- Create admin user (idempotent)
INSERT INTO users (uuid, name, slug, email, password, created_at, updated_at)
VALUES (
    '018e6c43-1e5e-7a9b-8c6d-5f4e2d1c0b9a', -- Static administrative UUID
    'Luis Rodz',
    'luis-rodz',
    'frodrigue60@gmail.com',
    '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- standard hash for 'password'
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    password = EXCLUDED.password,
    updated_at = CURRENT_TIMESTAMP;

-- Assign owner role
INSERT INTO role_user (user_id, role_id, created_at, updated_at)
SELECT
    (SELECT id FROM users WHERE email = 'frodrigue60@gmail.com'),
    (SELECT id FROM roles WHERE slug = 'owner'),
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
WHERE EXISTS (SELECT 1 FROM roles WHERE slug = 'owner')
AND NOT EXISTS (
    SELECT 1 FROM role_user
    WHERE user_id = (SELECT id FROM users WHERE email = 'frodrigue60@gmail.com')
    AND role_id = (SELECT id FROM roles WHERE slug = 'owner')
);

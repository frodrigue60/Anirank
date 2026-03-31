INSERT INTO roles (name, slug, description, created_at, updated_at) VALUES
('Owner',         'owner',   'Owner of the website', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Administrator', 'admin',   'Full system access',   CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Editor',        'editor',  'Content management',   CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Creator',       'creator', 'Original content creator', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('User',          'user',    'Standard user',        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (slug) DO NOTHING;

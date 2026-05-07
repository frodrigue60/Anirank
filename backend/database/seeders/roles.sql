INSERT INTO roles (name, slug, description, weight, created_at, updated_at) VALUES
('Owner',         'owner',   'Owner of the website', 100, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Administrator', 'admin',   'Full system access',   80,  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Editor',        'editor',  'Content management',   60,  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('Creator',       'creator', 'Original content creator', 40,  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('User',          'user',    'Standard user',        10,  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (slug) DO UPDATE SET weight = EXCLUDED.weight;

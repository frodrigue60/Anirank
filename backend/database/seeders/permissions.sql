-- Cleanup legacy permissions
DELETE FROM permissions WHERE slug = 'taxonomy.manage';
DELETE FROM permissions WHERE slug IN ('taxonomy.years', 'taxonomy.seasons', 'taxonomy.formats', 'taxonomy.genres');

-- Insert Permissions
INSERT INTO permissions (name, slug, description, created_at, updated_at) VALUES 
('Create Anime', 'anime.create', 'Allows creating new anime entries', NOW(), NOW()),
('Edit Anime', 'anime.edit', 'Allows editing existing anime and metadata', NOW(), NOW()),
('Delete Anime', 'anime.delete', 'Allows removing anime from the system', NOW(), NOW()),
('Create Song', 'song.create', 'Allows adding new opening/ending themes', NOW(), NOW()),
('Edit Song', 'song.edit', 'Allows editing song data and variants/videos', NOW(), NOW()),
('Delete Song', 'song.delete', 'Allows removing songs from the system', NOW(), NOW()),
('Create Artist', 'artist.create', 'Allows adding new artists', NOW(), NOW()),
('Edit Artist', 'artist.edit', 'Allows editing artist profiles', NOW(), NOW()),
('Delete Artist', 'artist.delete', 'Allows removing artists', NOW(), NOW()),
('Manage Reports', 'reports.manage', 'Allows resolving moderation/content reports (songs, comments, users)', NOW(), NOW()),
('Manage Users', 'users.manage', 'Allows listing and editing user profiles', NOW(), NOW()),
('Manage Permissions', 'permissions.manage', 'Allows modifying role-permission mappings', NOW(), NOW()),
('Manage Tournaments', 'tournament.manage', 'Allows creating, seeding and managing tournaments', NOW(), NOW()),
('Manage Announcements', 'announcement.manage', 'Allows creating and managing platform announcements', NOW(), NOW()),
('Manage Badges', 'badge.manage', 'Allows creating and managing system badges', NOW(), NOW()),

-- Granular Taxonomies - Years
('Create Years', 'taxonomy.years.create', 'Allows adding new years', NOW(), NOW()),
('Edit Years', 'taxonomy.years.edit', 'Allows editing existing years', NOW(), NOW()),
('Delete Years', 'taxonomy.years.delete', 'Allows removing years', NOW(), NOW()),

-- Granular Taxonomies - Seasons
('Create Seasons', 'taxonomy.seasons.create', 'Allows adding new seasons', NOW(), NOW()),
('Edit Seasons', 'taxonomy.seasons.edit', 'Allows editing existing seasons', NOW(), NOW()),
('Delete Seasons', 'taxonomy.seasons.delete', 'Allows removing seasons', NOW(), NOW()),

-- Granular Taxonomies - Formats
('Create Formats', 'taxonomy.formats.create', 'Allows adding new anime formats', NOW(), NOW()),
('Edit Formats', 'taxonomy.formats.edit', 'Allows editing existing formats', NOW(), NOW()),
('Delete Formats', 'taxonomy.formats.delete', 'Allows removing formats', NOW(), NOW()),

-- Granular Taxonomies - Genres
('Create Genres', 'taxonomy.genres.create', 'Allows adding new genres', NOW(), NOW()),
('Edit Genres', 'taxonomy.genres.edit', 'Allows editing existing genres', NOW(), NOW()),
('Delete Genres', 'taxonomy.genres.delete', 'Allows removing genres', NOW(), NOW()),

-- Granular Taxonomies - Studios
('Create Studios', 'taxonomy.studios.create', 'Allows adding new studios', NOW(), NOW()),
('Edit Studios', 'taxonomy.studios.edit', 'Allows editing existing studios', NOW(), NOW()),
('Delete Studios', 'taxonomy.studios.delete', 'Allows removing studios', NOW(), NOW()),

-- Granular Taxonomies - Producers
('Create Producers', 'taxonomy.producers.create', 'Allows adding new producers', NOW(), NOW()),
('Edit Producers', 'taxonomy.producers.edit', 'Allows editing existing producers', NOW(), NOW()),
('Delete Producers', 'taxonomy.producers.delete', 'Allows removing producers', NOW(), NOW())

ON CONFLICT (slug) DO UPDATE SET 
    name = EXCLUDED.name, 
    description = EXCLUDED.description, 
    updated_at = NOW();

-- Helper function or direct inserts for Role-Permission mappings
-- We use subqueries to get IDs based on slugs

DO $$
DECLARE
    admin_id bigint;
    editor_id bigint;
    creator_id bigint;
BEGIN
    SELECT id INTO admin_id FROM roles WHERE slug = 'admin';
    SELECT id INTO editor_id FROM roles WHERE slug = 'editor';
    SELECT id INTO creator_id FROM roles WHERE slug = 'creator';

    -- Map Admin Permissions
    IF admin_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id)
        SELECT admin_id, id FROM permissions WHERE slug IN (
            'anime.create', 'anime.edit', 'anime.delete',
            'song.create', 'song.edit', 'song.delete',
            'artist.create', 'artist.edit', 'artist.delete',
            'reports.manage', 'users.manage', 'permissions.manage',
            'tournament.manage', 'announcement.manage',
            'badge.manage',
            'taxonomy.years.create', 'taxonomy.years.edit', 'taxonomy.years.delete',
            'taxonomy.seasons.create', 'taxonomy.seasons.edit', 'taxonomy.seasons.delete',
            'taxonomy.formats.create', 'taxonomy.formats.edit', 'taxonomy.formats.delete',
            'taxonomy.genres.create', 'taxonomy.genres.edit', 'taxonomy.genres.delete',
            'taxonomy.studios.create', 'taxonomy.studios.edit', 'taxonomy.studios.delete',
            'taxonomy.producers.create', 'taxonomy.producers.edit', 'taxonomy.producers.delete'
        ) ON CONFLICT DO NOTHING;
    END IF;

    -- Map Editor Permissions
    IF editor_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id)
        SELECT editor_id, id FROM permissions WHERE slug IN (
            'anime.create', 'anime.edit',
            'song.create', 'song.edit',
            'artist.create', 'artist.edit',
            'reports.manage',
            'tournament.manage', 'announcement.manage',
            'taxonomy.years.create', 'taxonomy.years.edit',
            'taxonomy.seasons.create', 'taxonomy.seasons.edit',
            'taxonomy.formats.create', 'taxonomy.formats.edit',
            'taxonomy.genres.create', 'taxonomy.genres.edit',
            'taxonomy.studios.create', 'taxonomy.studios.edit',
            'taxonomy.producers.create', 'taxonomy.producers.edit'
        ) ON CONFLICT DO NOTHING;
    END IF;

    -- Map Creator Permissions
    IF creator_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id)
        SELECT creator_id, id FROM permissions WHERE slug IN (
            'anime.create', 'song.create', 'artist.create',
            'taxonomy.years.create',
            'taxonomy.seasons.create',
            'taxonomy.formats.create',
            'taxonomy.genres.create',
            'taxonomy.studios.create',
            'taxonomy.producers.create'
        ) ON CONFLICT DO NOTHING;
    END IF;
END $$;

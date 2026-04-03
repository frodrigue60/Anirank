-- Automatic Badges Seeder
-- requirement_type: 'level', 'ratings', 'anilist', 'comments'

DO $$
DECLARE
    badge_record RECORD;
BEGIN
    -- Explorer I
    IF NOT EXISTS (SELECT 1 FROM badges WHERE name = 'Explorer I') THEN
        INSERT INTO badges (name, description, is_active, is_automatic, requirement_type, requirement_value, icon, created_at, updated_at)
        VALUES ('Explorer I', 'Reached level 1', true, true, 'level', 1, 'badges/level_5.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
    END IF;

    -- Explorer II
    IF NOT EXISTS (SELECT 1 FROM badges WHERE name = 'Explorer II') THEN
        INSERT INTO badges (name, description, is_active, is_automatic, requirement_type, requirement_value, icon, created_at, updated_at)
        VALUES ('Explorer II', 'Reached level 2', true, true, 'level', 2, 'badges/level_10.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
    END IF;

    -- Explorer III
    IF NOT EXISTS (SELECT 1 FROM badges WHERE name = 'Explorer III') THEN
        INSERT INTO badges (name, description, is_active, is_automatic, requirement_type, requirement_value, icon, created_at, updated_at)
        VALUES ('Explorer III', 'Reached level 3', true, true, 'level', 3, 'badges/level_20.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
    END IF;

    -- Music Critic
    IF NOT EXISTS (SELECT 1 FROM badges WHERE name = 'Music Critic') THEN
        INSERT INTO badges (name, description, is_active, is_automatic, requirement_type, requirement_value, icon, created_at, updated_at)
        VALUES ('Music Critic', 'Rated 5 songs', true, true, 'ratings', 5, 'badges/critic_10.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
    END IF;

    -- Vocal Member
    IF NOT EXISTS (SELECT 1 FROM badges WHERE name = 'Vocal Member') THEN
        INSERT INTO badges (name, description, is_active, is_automatic, requirement_type, requirement_value, icon, created_at, updated_at)
        VALUES ('Vocal Member', 'Posted 5 comments', true, true, 'comments', 5, 'badges/comments_5.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
    END IF;

    -- Anifan
    IF NOT EXISTS (SELECT 1 FROM badges WHERE name = 'Anifan') THEN
        INSERT INTO badges (name, description, is_active, is_automatic, requirement_type, requirement_value, icon, created_at, updated_at)
        VALUES ('Anifan', 'Connected your AniList account', true, true, 'anilist', 1, 'badges/anilist_connected.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
    END IF;
END;
$$;

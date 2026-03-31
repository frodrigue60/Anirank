INSERT INTO xp_activities (key, xp_amount, description, cooldown_seconds, created_at, updated_at) VALUES
('rate_song',        500,  'Rating a song (OP / ED)',            0,     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('add_favorite',     1000, 'Adding a song or artist to favorites', 0,     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('comment',          500,  'Leaving a comment',                  300,   CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('daily_login',      200,  'First login of the day',             86400, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('create_playlist',  5000, 'Creating a new playlist',            300,   CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('add_to_playlist',  1000, 'Adding a song to a playlist',         0,     CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (key) DO UPDATE SET
    xp_amount = EXCLUDED.xp_amount,
    description = EXCLUDED.description,
    cooldown_seconds = EXCLUDED.cooldown_seconds,
    updated_at = CURRENT_TIMESTAMP;

-- Insert levels 1 to 100 with min_xp calculation: 500 * (i - 1) * i
DO $$
BEGIN
    FOR i IN 1..100 LOOP
        INSERT INTO levels (level, min_xp, name, created_at, updated_at)
        VALUES (i, 500 * (i - 1) * i, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        ON CONFLICT (level) DO UPDATE SET
            min_xp = EXCLUDED.min_xp,
            updated_at = CURRENT_TIMESTAMP;
    END LOOP;
END;
$$;

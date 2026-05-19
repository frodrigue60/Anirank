-- Migration: Deduplicate imported data and establish unique constraints/indexes to prevent future duplicates.
-- Targets: songs, song_variants, and artist_song pivot relationships.

-- 1. Create a temp table to map duplicate songs to their kept target song
CREATE TEMP TABLE temp_song_mapping AS
WITH song_duplicates AS (
    SELECT 
        id,
        COALESCE(anime_themes_id::text, anime_id || '_' || type_id || '_' || theme_num) AS dup_group,
        ROW_NUMBER() OVER (
            PARTITION BY COALESCE(anime_themes_id::text, anime_id || '_' || type_id || '_' || theme_num)
            ORDER BY 
                (anime_themes_id IS NOT NULL) DESC,
                views DESC, 
                id ASC
        ) as rn
    FROM songs
),
song_mapping AS (
    SELECT 
        d.id AS old_song_id,
        k.id AS new_song_id
    FROM song_duplicates d
    JOIN song_duplicates k ON d.dup_group = k.dup_group AND k.rn = 1
    WHERE d.rn > 1
)
SELECT old_song_id, new_song_id FROM song_mapping;

-- 2. Deduplicate unique-constrained referencing tables BEFORE updating keys to avoid constraint violations

-- song_ratings
DELETE FROM song_ratings sr
WHERE EXISTS (
    SELECT 1 FROM temp_song_mapping m
    WHERE sr.song_id = m.old_song_id
    AND EXISTS (
        SELECT 1 FROM song_ratings sr2
        WHERE sr2.user_id = sr.user_id
        AND sr2.song_id = m.new_song_id
    )
);

UPDATE song_ratings sr
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE sr.song_id = m.old_song_id;

-- song_reactions
DELETE FROM song_reactions sr
WHERE EXISTS (
    SELECT 1 FROM temp_song_mapping m
    WHERE sr.song_id = m.old_song_id
    AND EXISTS (
        SELECT 1 FROM song_reactions sr2
        WHERE sr2.user_id = sr.user_id
        AND sr2.song_id = m.new_song_id
    )
);

UPDATE song_reactions sr
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE sr.song_id = m.old_song_id;

-- song_user (favorites)
DELETE FROM song_user su
WHERE EXISTS (
    SELECT 1 FROM temp_song_mapping m
    WHERE su.song_id = m.old_song_id
    AND EXISTS (
        SELECT 1 FROM song_user su2
        WHERE su2.user_id = su.user_id
        AND su2.song_id = m.new_song_id
    )
);

UPDATE song_user su
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE su.song_id = m.old_song_id;

-- playlist_song
DELETE FROM playlist_song ps
WHERE EXISTS (
    SELECT 1 FROM temp_song_mapping m
    WHERE ps.song_id = m.old_song_id
    AND EXISTS (
        SELECT 1 FROM playlist_song ps2
        WHERE ps2.playlist_id = ps.playlist_id
        AND ps2.song_id = m.new_song_id
    )
);

UPDATE playlist_song ps
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE ps.song_id = m.old_song_id;

-- daily_metrics
DELETE FROM daily_metrics dm
WHERE EXISTS (
    SELECT 1 FROM temp_song_mapping m
    WHERE dm.song_id = m.old_song_id
    AND EXISTS (
        SELECT 1 FROM daily_metrics dm2
        WHERE dm2.date = dm.date
        AND dm2.song_id = m.new_song_id
    )
);

UPDATE daily_metrics dm
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE dm.song_id = m.old_song_id;

-- ranking_histories
DELETE FROM ranking_histories rh
WHERE EXISTS (
    SELECT 1 FROM temp_song_mapping m
    WHERE rh.song_id = m.old_song_id
    AND EXISTS (
        SELECT 1 FROM ranking_histories rh2
        WHERE rh2.date = rh.date
        AND rh2.song_id = m.new_song_id
    )
);

UPDATE ranking_histories rh
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE rh.song_id = m.old_song_id;

-- music_genre_song
DELETE FROM music_genre_song mgs
WHERE EXISTS (
    SELECT 1 FROM temp_song_mapping m
    WHERE mgs.song_id = m.old_song_id
    AND EXISTS (
        SELECT 1 FROM music_genre_song mgs2
        WHERE mgs2.music_genre_id = mgs.music_genre_id
        AND mgs2.song_id = m.new_song_id
    )
);

UPDATE music_genre_song mgs
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE mgs.song_id = m.old_song_id;

-- song_reports
DELETE FROM song_reports sr
WHERE EXISTS (
    SELECT 1 FROM temp_song_mapping m
    WHERE sr.song_id = m.old_song_id
    AND EXISTS (
        SELECT 1 FROM song_reports sr2
        WHERE sr2.user_id = sr.user_id
        AND sr2.song_id = m.new_song_id
        AND sr2.is_accepted = false
    )
) AND sr.is_accepted = false;

UPDATE song_reports sr
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE sr.song_id = m.old_song_id;

-- comments (no unique constraints on song_id)
UPDATE comments c
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE c.song_id = m.old_song_id;

-- tournament_matchups (no unique constraints on song IDs)
UPDATE tournament_matchups tm SET song1_id = m.new_song_id FROM temp_song_mapping m WHERE tm.song1_id = m.old_song_id;
UPDATE tournament_matchups tm SET song2_id = m.new_song_id FROM temp_song_mapping m WHERE tm.song2_id = m.old_song_id;
UPDATE tournament_matchups tm SET winner_song_id = m.new_song_id FROM temp_song_mapping m WHERE tm.winner_song_id = m.old_song_id;

-- tournaments
UPDATE tournaments t SET winner_song_id = m.new_song_id FROM temp_song_mapping m WHERE t.winner_song_id = m.old_song_id;

-- tournament_votes
UPDATE tournament_votes tv SET song_id = m.new_song_id FROM temp_song_mapping m WHERE tv.song_id = m.old_song_id;

-- artist_song
UPDATE artist_song asong
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE asong.song_id = m.old_song_id;

-- song_variants (update reference to parent song before deduplicating variants)
UPDATE song_variants sv
SET song_id = m.new_song_id
FROM temp_song_mapping m
WHERE sv.song_id = m.old_song_id;


-- 3. Create a temp table to map duplicate song_variants to their kept target variant
CREATE TEMP TABLE temp_variant_mapping AS
WITH variant_duplicates AS (
    SELECT 
        id,
        song_id,
        slug,
        COALESCE(anime_themes_id::text, song_id || '_' || slug) AS dup_group,
        ROW_NUMBER() OVER (
            PARTITION BY COALESCE(anime_themes_id::text, song_id || '_' || slug)
            ORDER BY 
                (anime_themes_id IS NOT NULL) DESC,
                id ASC
        ) as rn
    FROM song_variants
),
variant_mapping AS (
    SELECT 
        d.id AS old_variant_id,
        k.id AS new_variant_id
    FROM variant_duplicates d
    JOIN variant_duplicates k ON d.dup_group = k.dup_group AND k.rn = 1
    WHERE d.rn > 1
)
SELECT old_variant_id, new_variant_id FROM variant_mapping;

-- Delete duplicate videos to prevent unique constraint violation on idx_videos_variant_src
DELETE FROM videos v
WHERE EXISTS (
    SELECT 1 FROM temp_variant_mapping m
    WHERE v.song_variant_id = m.old_variant_id
    AND EXISTS (
        SELECT 1 FROM videos v2
        WHERE v2.video_src = v.video_src
        AND v2.song_variant_id = m.new_variant_id
    )
);

-- Update videos to point to kept variant
UPDATE videos v
SET song_variant_id = m.new_variant_id
FROM temp_variant_mapping m
WHERE v.song_variant_id = m.old_variant_id;

-- Delete redundant song_variants
DELETE FROM song_variants
WHERE id IN (SELECT old_variant_id FROM temp_variant_mapping);


-- 4. Delete redundant songs
DELETE FROM songs
WHERE id IN (SELECT old_song_id FROM temp_song_mapping);


-- 5. Deduplicate artist_song pivot table
WITH artist_song_duplicates AS (
    SELECT 
        id,
        ROW_NUMBER() OVER (
            PARTITION BY artist_id, song_id
            ORDER BY id ASC
        ) as rn
    FROM artist_song
)
DELETE FROM artist_song
WHERE id IN (
    SELECT id FROM artist_song_duplicates WHERE rn > 1
);


-- 6. Cleanup temp tables
DROP TABLE IF EXISTS temp_song_mapping;
DROP TABLE IF EXISTS temp_variant_mapping;


-- 7. Create Constraints & Indexes safely

-- Ensure unique index on artist_song(artist_id, song_id) to prevent duplicate links
CREATE UNIQUE INDEX IF NOT EXISTS idx_artist_song_uniq ON artist_song(artist_id, song_id);

-- Ensure songs_anime_themes_id_unique UNIQUE constraint on songs
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'songs_anime_themes_id_unique'
    ) THEN
        ALTER TABLE songs ADD CONSTRAINT songs_anime_themes_id_unique UNIQUE (anime_themes_id);
    END IF;
END $$;

-- Ensure song_variants_anime_themes_id_unique UNIQUE constraint on song_variants
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'song_variants_anime_themes_id_unique'
    ) THEN
        ALTER TABLE song_variants ADD CONSTRAINT song_variants_anime_themes_id_unique UNIQUE (anime_themes_id);
    END IF;
END $$;

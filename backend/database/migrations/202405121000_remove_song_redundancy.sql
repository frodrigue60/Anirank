-- Migration to remove redundancy in the songs table
-- This script ensures type_id is populated and then drops type and slug columns

-- 1. Ensure type_id is correctly populated for all songs (Safety check)
-- This assumes song_types table has slugs 'OP', 'ED', etc.
UPDATE songs s
SET type_id = st.id
FROM song_types st
WHERE (s.type = st.slug OR s.type = st.name)
  AND (s.type_id IS NULL OR s.type_id = 0);

-- 2. Drop the redundant columns
ALTER TABLE songs DROP COLUMN IF EXISTS type;
ALTER TABLE songs DROP COLUMN IF EXISTS slug;

-- 3. Add an index to keep slug searching fast via (type_id, theme_num)
CREATE INDEX IF NOT EXISTS idx_songs_type_theme ON songs (type_id, theme_num);

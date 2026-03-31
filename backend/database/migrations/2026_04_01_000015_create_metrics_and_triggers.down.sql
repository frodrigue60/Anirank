-- Drop Triggers and Functions
DROP TRIGGER IF EXISTS trig_variant_views_update ON song_variants;
DROP TRIGGER IF EXISTS trig_song_views_update ON songs;
DROP FUNCTION IF EXISTS fn_update_daily_variant_views();
DROP FUNCTION IF EXISTS fn_update_daily_song_views();

DROP TRIGGER IF EXISTS trg_artist_song_counters_status ON songs;
DROP TRIGGER IF EXISTS trg_artist_song_counters_pivot ON artist_song;
DROP TRIGGER IF EXISTS trg_artist_song_counters_deletion ON songs;
DROP FUNCTION IF EXISTS fn_update_artist_song_counters_status();
DROP FUNCTION IF EXISTS fn_update_artist_song_counters_pivot();
DROP FUNCTION IF EXISTS fn_update_artist_song_counters_deletion();

DROP TRIGGER IF EXISTS trg_update_producer_anime_count ON anime_producer;
DROP FUNCTION IF EXISTS fn_update_producer_anime_count();

DROP TRIGGER IF EXISTS trg_update_studio_anime_count ON anime_studio;
DROP FUNCTION IF EXISTS fn_update_studio_anime_count();

DROP TRIGGER IF EXISTS trg_update_anime_songs_count ON songs;
DROP FUNCTION IF EXISTS update_anime_songs_count();

DROP TRIGGER IF EXISTS trg_update_song_average_score ON song_ratings;
DROP FUNCTION IF EXISTS update_song_average_score();

DROP TRIGGER IF EXISTS trg_update_artist_favorites_count ON artist_user;
DROP FUNCTION IF EXISTS update_artist_favorites_count();

DROP TRIGGER IF EXISTS trg_update_song_favorites_count ON song_user;
DROP FUNCTION IF EXISTS update_song_favorites_count();

-- Drop Trigram Indexes
DROP INDEX IF EXISTS artists_name_trgm_idx;
DROP INDEX IF EXISTS songs_jp_trgm_idx;
DROP INDEX IF EXISTS songs_en_trgm_idx;
DROP INDEX IF EXISTS songs_romaji_trgm_idx;
DROP INDEX IF EXISTS animes_title_trgm_idx;

-- Drop Extensions
-- DROP EXTENSION IF EXISTS pg_trgm; -- Usually safer not to drop extensions if others might use it

-- Drop Metrics Tables
DROP TABLE IF EXISTS ranking_histories;
DROP TABLE IF EXISTS daily_metrics;

-- Drop Infrastructure Tables
DROP TABLE IF EXISTS failed_jobs;
DROP TABLE IF EXISTS cache;

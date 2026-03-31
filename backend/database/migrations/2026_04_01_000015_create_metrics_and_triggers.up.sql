-- ─── Infrastructure Tables ──────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS cache (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    expiration INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS failed_jobs (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    connection TEXT NOT NULL,
    queue TEXT NOT NULL,
    payload TEXT NOT NULL,
    exception TEXT NOT NULL,
    failed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ─── Metrics Tables ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS daily_metrics (
    id BIGSERIAL PRIMARY KEY,
    song_id BIGINT NULL REFERENCES songs(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    views_count INTEGER DEFAULT 0,
    new_users_count INTEGER DEFAULT 0,
    new_ratings_count INTEGER DEFAULT 0,
    new_songs_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_song_date UNIQUE (song_id, date)
);

-- Site-wide metrics index
CREATE UNIQUE INDEX IF NOT EXISTS daily_metrics_site_wide_unique ON daily_metrics (date) WHERE song_id IS NULL;

CREATE TABLE IF NOT EXISTS ranking_histories (
    id BIGSERIAL PRIMARY KEY,
    song_id BIGINT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    rank INTEGER NOT NULL,
    seasonal_rank INTEGER,
    score DECIMAL(8, 2),
    date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_song_rank_date UNIQUE (song_id, date)
);

-- ─── PostgreSQL Functions & Triggers ─────────────────────────────────────────

-- 1. Song Favorites Counter
CREATE OR REPLACE FUNCTION update_song_favorites_count() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE songs SET favorites_count = favorites_count + 1 WHERE id = NEW.song_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE songs SET favorites_count = GREATEST(0, favorites_count - 1) WHERE id = OLD.song_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_song_favorites_count ON song_user;
CREATE TRIGGER trg_update_song_favorites_count
AFTER INSERT OR DELETE ON song_user
FOR EACH ROW EXECUTE FUNCTION update_song_favorites_count();

-- 2. Artist Favorites Counter
CREATE OR REPLACE FUNCTION update_artist_favorites_count() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE artists SET favorites_count = favorites_count + 1 WHERE id = NEW.artist_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE artists SET favorites_count = GREATEST(0, favorites_count - 1) WHERE id = OLD.artist_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_artist_favorites_count ON artist_user;
CREATE TRIGGER trg_update_artist_favorites_count
AFTER INSERT OR DELETE ON artist_user
FOR EACH ROW EXECUTE FUNCTION update_artist_favorites_count();

-- 3. Song Average Score
CREATE OR REPLACE FUNCTION update_song_average_score() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT' OR TG_OP = 'UPDATE') THEN
        UPDATE songs SET average_score = COALESCE((SELECT AVG(rating) FROM song_ratings WHERE song_id = NEW.song_id), 0) WHERE id = NEW.song_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE songs SET average_score = COALESCE((SELECT AVG(rating) FROM song_ratings WHERE song_id = OLD.song_id), 0) WHERE id = OLD.song_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_song_average_score ON song_ratings;
CREATE TRIGGER trg_update_song_average_score
AFTER INSERT OR UPDATE OR DELETE ON song_ratings
FOR EACH ROW EXECUTE FUNCTION update_song_average_score();

-- 4. Anime Songs Count
CREATE OR REPLACE FUNCTION update_anime_songs_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE animes SET songs_count = songs_count + 1 WHERE id = NEW.anime_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE animes SET songs_count = GREATEST(0, songs_count - 1) WHERE id = OLD.anime_id;
    ELSIF (TG_OP = 'UPDATE' AND (OLD.anime_id IS DISTINCT FROM NEW.anime_id)) THEN
        IF (OLD.anime_id IS NOT NULL) THEN
            UPDATE animes SET songs_count = GREATEST(0, songs_count - 1) WHERE id = OLD.anime_id;
        END IF;
        IF (NEW.anime_id IS NOT NULL) THEN
            UPDATE animes SET songs_count = songs_count + 1 WHERE id = NEW.anime_id;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_anime_songs_count ON songs;
CREATE TRIGGER trg_update_anime_songs_count
AFTER INSERT OR DELETE OR UPDATE ON songs
FOR EACH ROW EXECUTE FUNCTION update_anime_songs_count();

-- 5. Studio Anime Count
CREATE OR REPLACE FUNCTION fn_update_studio_anime_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE studios SET anime_count = anime_count + 1 WHERE id = NEW.studio_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE studios SET anime_count = GREATEST(0, anime_count - 1) WHERE id = OLD.studio_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_studio_anime_count ON anime_studio;
CREATE TRIGGER trg_update_studio_anime_count
AFTER INSERT OR DELETE ON anime_studio
FOR EACH ROW EXECUTE FUNCTION fn_update_studio_anime_count();

-- 6. Producer Anime Count
CREATE OR REPLACE FUNCTION fn_update_producer_anime_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE producers SET anime_count = anime_count + 1 WHERE id = NEW.producer_id;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE producers SET anime_count = GREATEST(0, anime_count - 1) WHERE id = OLD.producer_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_update_producer_anime_count ON anime_producer;
CREATE TRIGGER trg_update_producer_anime_count
AFTER INSERT OR DELETE ON anime_producer
FOR EACH ROW EXECUTE FUNCTION fn_update_producer_anime_count();

-- 7. Artist Song Counters
CREATE OR REPLACE FUNCTION fn_update_artist_song_counters_deletion()
RETURNS TRIGGER AS $$
BEGIN
    IF (OLD.status = TRUE) THEN
        UPDATE artists SET enabled_songs = GREATEST(0, enabled_songs - 1)
        WHERE id IN (SELECT artist_id FROM artist_song WHERE song_id = OLD.id);
    ELSE
        UPDATE artists SET disabled_songs = GREATEST(0, disabled_songs - 1)
        WHERE id IN (SELECT artist_id FROM artist_song WHERE song_id = OLD.id);
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_artist_song_counters_deletion ON songs;
CREATE TRIGGER trg_artist_song_counters_deletion
BEFORE DELETE ON songs
FOR EACH ROW EXECUTE FUNCTION fn_update_artist_song_counters_deletion();

CREATE OR REPLACE FUNCTION fn_update_artist_song_counters_pivot()
RETURNS TRIGGER AS $$
DECLARE
    song_status BOOLEAN;
BEGIN
    IF (TG_OP = 'INSERT') THEN
        SELECT status INTO song_status FROM songs WHERE id = NEW.song_id;
        IF (song_status = TRUE) THEN
            UPDATE artists SET enabled_songs = enabled_songs + 1 WHERE id = NEW.artist_id;
        ELSE
            UPDATE artists SET disabled_songs = disabled_songs + 1 WHERE id = NEW.artist_id;
        END IF;
    ELSIF (TG_OP = 'DELETE') THEN
        SELECT status INTO song_status FROM songs WHERE id = OLD.song_id;
        IF (song_status IS NOT NULL) THEN
            IF (song_status = TRUE) THEN
                UPDATE artists SET enabled_songs = GREATEST(0, enabled_songs - 1) WHERE id = OLD.artist_id;
            ELSE
                UPDATE artists SET disabled_songs = GREATEST(0, disabled_songs - 1) WHERE id = OLD.artist_id;
            END IF;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_artist_song_counters_pivot ON artist_song;
CREATE TRIGGER trg_artist_song_counters_pivot
AFTER INSERT OR DELETE ON artist_song
FOR EACH ROW EXECUTE FUNCTION fn_update_artist_song_counters_pivot();

CREATE OR REPLACE FUNCTION fn_update_artist_song_counters_status()
RETURNS TRIGGER AS $$
BEGIN
    IF (OLD.status = FALSE AND NEW.status = TRUE) THEN
        UPDATE artists SET
            enabled_songs = enabled_songs + 1,
            disabled_songs = GREATEST(0, disabled_songs - 1)
        WHERE id IN (SELECT artist_id FROM artist_song WHERE song_id = NEW.id);
    ELSIF (OLD.status = TRUE AND NEW.status = FALSE) THEN
        UPDATE artists SET
            enabled_songs = GREATEST(0, enabled_songs - 1),
            disabled_songs = disabled_songs + 1
        WHERE id IN (SELECT artist_id FROM artist_song WHERE song_id = NEW.id);
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_artist_song_counters_status ON songs;
CREATE TRIGGER trg_artist_song_counters_status
AFTER UPDATE OF status ON songs
FOR EACH ROW EXECUTE FUNCTION fn_update_artist_song_counters_status();

-- 8. Daily Metrics: Song Views
CREATE OR REPLACE FUNCTION fn_update_daily_song_views()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO daily_metrics (song_id, date, views_count, created_at, updated_at)
    VALUES (NEW.id, CURRENT_DATE, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    ON CONFLICT (song_id, date)
    DO UPDATE SET
        views_count = daily_metrics.views_count + 1,
        updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trig_song_views_update ON songs;
CREATE TRIGGER trig_song_views_update
AFTER UPDATE OF views ON songs
FOR EACH ROW
WHEN (NEW.views > OLD.views)
EXECUTE FUNCTION fn_update_daily_song_views();

-- 9. Daily Metrics: Song Variant Views
CREATE OR REPLACE FUNCTION fn_update_daily_variant_views()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO daily_metrics (song_id, date, views_count, created_at, updated_at)
    VALUES (NEW.song_id, CURRENT_DATE, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    ON CONFLICT (song_id, date)
    DO UPDATE SET
        views_count = daily_metrics.views_count + 1,
        updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trig_variant_views_update ON song_variants;
CREATE TRIGGER trig_variant_views_update
AFTER UPDATE OF views ON song_variants
FOR EACH ROW
WHEN (NEW.views > OLD.views)
EXECUTE FUNCTION fn_update_daily_variant_views();

-- 10. Trigram Search Indexes
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS animes_title_trgm_idx ON animes USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS songs_romaji_trgm_idx ON songs USING gin (song_romaji gin_trgm_ops);
CREATE INDEX IF NOT EXISTS songs_en_trgm_idx ON songs USING gin (song_en gin_trgm_ops);
CREATE INDEX IF NOT EXISTS songs_jp_trgm_idx ON songs USING gin (song_jp gin_trgm_ops);
CREATE INDEX IF NOT EXISTS artists_name_trgm_idx ON artists USING gin (name gin_trgm_ops);

-- Migration: Add anime title variants and synonyms
-- Date: 2026-06-02

-- 1. Add columns to animes table
ALTER TABLE animes ADD COLUMN IF NOT EXISTS title_english TEXT;
ALTER TABLE animes ADD COLUMN IF NOT EXISTS title_native TEXT;
ALTER TABLE animes ADD COLUMN IF NOT EXISTS synonyms TEXT[] NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_animes_synonyms ON animes USING GIN (synonyms);

-- 2. Modify search_index to support extra_terms
ALTER TABLE search_index DROP COLUMN IF EXISTS search_vector;
ALTER TABLE search_index ADD COLUMN IF NOT EXISTS extra_terms TEXT;
ALTER TABLE search_index ADD COLUMN search_vector tsvector GENERATED ALWAYS AS (
    setweight(to_tsvector('simple', COALESCE(title, '')), 'A') ||
    setweight(to_tsvector('simple', COALESCE(subtitle, '')), 'B') ||
    setweight(to_tsvector('simple', COALESCE(extra_terms, '')), 'C')
) STORED;

CREATE INDEX IF NOT EXISTS idx_search_index_vector ON search_index USING gin (search_vector);

-- 3. Update sync_search_index trigger function
CREATE OR REPLACE FUNCTION public.sync_search_index() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            DECLARE
                v_title       TEXT;
                v_subtitle    TEXT;
                v_slug        TEXT;
                v_image       TEXT;
                v_type        VARCHAR(50);
                v_st_slug     TEXT;
                v_extra_terms TEXT;
            BEGIN
                v_type := TG_ARGV[0];

                -- Handle DELETE: remove the entry from the index
                IF (TG_OP = 'DELETE') THEN
                    DELETE FROM search_index
                     WHERE item_type = v_type
                       AND item_id   = OLD.uuid;
                    RETURN OLD;
                END IF;

                -- Map each table to its searchable fields
                CASE v_type
                    WHEN 'anime' THEN
                        v_title       := NEW.title;
                        v_slug        := NEW.slug;
                        v_image       := NEW.cover;
                        v_subtitle    := 'Anime';
                        v_extra_terms := COALESCE(NEW.title_english, '') || ' ' || COALESCE(NEW.title_native, '') || ' ' || array_to_string(NEW.synonyms, ' ');

                    WHEN 'song' THEN
                        v_title    := coalesce(NEW.song_romaji, NEW.song_en, NEW.song_jp);
                        
                        -- Fetch the anime slug and song type slug to recreate context
                        SELECT slug INTO v_slug FROM animes WHERE id = NEW.anime_id;
                        SELECT slug INTO v_st_slug FROM song_types WHERE id = NEW.type_id;
                        
                        v_slug     := v_slug || '/' || (coalesce(v_st_slug, '') || coalesce(NEW.theme_num, ''));
                        v_image    := NULL;
                        v_subtitle := 'Song • ' || coalesce(v_st_slug, 'Theme');

                    WHEN 'artist' THEN
                        v_title    := NEW.name;
                        v_slug     := NEW.slug;
                        v_image    := NEW.avatar;
                        v_subtitle := 'Artist';

                    WHEN 'user' THEN
                        v_title    := NEW.name;
                        v_slug     := NEW.slug;
                        v_image    := NEW.avatar;
                        v_subtitle := 'User';

                    WHEN 'studio' THEN
                        v_title    := NEW.name;
                        v_slug     := NEW.slug;
                        v_image    := NEW.logo;
                        v_subtitle := 'Studio';

                    WHEN 'producer' THEN
                        v_title    := NEW.name;
                        v_slug     := NEW.slug;
                        v_image    := NEW.logo;
                        v_subtitle := 'Producer';

                    ELSE
                        RETURN NEW;
                END CASE;

                -- Guard: skip if there is no meaningful title to index
                IF v_title IS NULL OR trim(v_title) = '' THEN
                    RETURN NEW;
                END IF;

                -- UPSERT — insert or update on (item_type, item_id) conflict
                INSERT INTO search_index (item_type, item_id, title, subtitle, slug, image_url, extra_terms, updated_at)
                VALUES (v_type, NEW.uuid, v_title, v_subtitle, v_slug, v_image, v_extra_terms, NOW())
                ON CONFLICT (item_type, item_id) DO UPDATE SET
                    title       = EXCLUDED.title,
                    subtitle    = EXCLUDED.subtitle,
                    slug        = EXCLUDED.slug,
                    image_url   = EXCLUDED.image_url,
                    extra_terms = EXCLUDED.extra_terms,
                    updated_at  = NOW();

                RETURN NEW;
            END;
            $$;

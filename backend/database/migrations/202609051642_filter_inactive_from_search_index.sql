-- Exclude inactive (status=false) anime/song/artist rows from public search_index,
-- and purge stale inactive entries already present.

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
                v_anime_ok    BOOLEAN;
            BEGIN
                v_type := TG_ARGV[0];

                -- Handle DELETE: remove the entry from the index
                IF (TG_OP = 'DELETE') THEN
                    DELETE FROM search_index
                     WHERE item_type = v_type
                       AND item_id   = OLD.uuid;

                    -- Deleting an anime also drops its song search rows.
                    IF v_type = 'anime' THEN
                        DELETE FROM search_index si
                        USING songs s
                        WHERE si.item_type = 'song'
                          AND si.item_id = s.uuid
                          AND s.anime_id = OLD.id;
                    END IF;

                    RETURN OLD;
                END IF;

                -- Inactive catalog rows must not appear in public search.
                IF v_type IN ('anime', 'song', 'artist') AND NOT COALESCE(NEW.status, false) THEN
                    DELETE FROM search_index
                     WHERE item_type = v_type
                       AND item_id   = NEW.uuid;

                    IF v_type = 'anime' THEN
                        DELETE FROM search_index si
                        USING songs s
                        WHERE si.item_type = 'song'
                          AND si.item_id = s.uuid
                          AND s.anime_id = NEW.id;
                    END IF;

                    RETURN NEW;
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
                        -- Parent anime must also be published.
                        SELECT a.status INTO v_anime_ok
                          FROM animes a
                         WHERE a.id = NEW.anime_id;
                        IF NOT COALESCE(v_anime_ok, false) THEN
                            DELETE FROM search_index
                             WHERE item_type = 'song'
                               AND item_id   = NEW.uuid;
                            RETURN NEW;
                        END IF;

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

-- Purge stale inactive rows already sitting in the index.
DELETE FROM search_index si
WHERE si.item_type = 'anime'
  AND NOT EXISTS (
      SELECT 1 FROM animes a
      WHERE a.uuid = si.item_id AND a.status = true
  );

DELETE FROM search_index si
WHERE si.item_type = 'song'
  AND NOT EXISTS (
      SELECT 1
      FROM songs s
      JOIN animes a ON a.id = s.anime_id
      WHERE s.uuid = si.item_id
        AND s.status = true
        AND a.status = true
  );

DELETE FROM search_index si
WHERE si.item_type = 'artist'
  AND NOT EXISTS (
      SELECT 1 FROM artists ar
      WHERE ar.uuid = si.item_id AND ar.status = true
  );

-- Migration to fix sync_search_index function after songs table schema change
-- The original function used NEW.type and NEW.slug which were dropped from songs table.

CREATE OR REPLACE FUNCTION public.sync_search_index() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
            DECLARE
                v_title    TEXT;
                v_subtitle TEXT;
                v_slug     TEXT;
                v_image    TEXT;
                v_type     VARCHAR(50);
                v_st_slug  TEXT;
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
                        v_title    := NEW.title;
                        v_slug     := NEW.slug;
                        v_image    := NEW.cover;
                        v_subtitle := 'Anime';

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
                INSERT INTO search_index (item_type, item_id, title, subtitle, slug, image_url, updated_at)
                VALUES (v_type, NEW.uuid, v_title, v_subtitle, v_slug, v_image, NOW())
                ON CONFLICT (item_type, item_id) DO UPDATE SET
                    title      = EXCLUDED.title,
                    subtitle   = EXCLUDED.subtitle,
                    slug       = EXCLUDED.slug,
                    image_url  = EXCLUDED.image_url,
                    updated_at = NOW();

                RETURN NEW;
            END;
            $$;

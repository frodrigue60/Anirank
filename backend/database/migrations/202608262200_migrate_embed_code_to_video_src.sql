-- Migration: Promote embed_code into video_src, then drop embed_code
-- Date: 2026-08-26
-- Policy: UPDATE existing rows only — never DELETE video rows.
-- video_src becomes the single source of truth (S3/R2 paths or preserved legacy URLs).

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'videos'
          AND column_name = 'embed_code'
    ) THEN
        -- 1) Copy embed payload into video_src when video_src is empty.
        --    Prefer iframe src="...", else raw embed_code, else a stable legacy key.
        --    If the candidate would collide with another row of the same variant,
        --    append '#e{id}' so the unique (song_variant_id, video_src) index stays valid.
        UPDATE videos AS v
        SET
            video_src = sub.new_src,
            updated_at = CURRENT_TIMESTAMP
        FROM (
            SELECT
                v0.id,
                CASE
                    WHEN EXISTS (
                        SELECT 1
                        FROM videos AS x
                        WHERE x.song_variant_id = v0.song_variant_id
                          AND x.id <> v0.id
                          AND x.video_src IS NOT NULL
                          AND x.video_src <> ''
                          AND x.video_src = COALESCE(
                              NULLIF(substring(v0.embed_code FROM 'src="([^"]+)"'), ''),
                              NULLIF(btrim(v0.embed_code), '')
                          )
                    ) THEN
                        COALESCE(
                            NULLIF(substring(v0.embed_code FROM 'src="([^"]+)"'), ''),
                            NULLIF(btrim(v0.embed_code), ''),
                            'legacy-embed'
                        ) || '#e' || v0.id::text
                    ELSE
                        COALESCE(
                            NULLIF(substring(v0.embed_code FROM 'src="([^"]+)"'), ''),
                            NULLIF(btrim(v0.embed_code), ''),
                            'legacy-embed/' || v0.id::text
                        )
                END AS new_src
            FROM videos AS v0
            WHERE (v0.video_src IS NULL OR btrim(v0.video_src) = '')
              AND v0.embed_code IS NOT NULL
              AND btrim(v0.embed_code) <> ''
        ) AS sub
        WHERE v.id = sub.id
          AND (v.video_src IS NULL OR btrim(v.video_src) = '');

        -- 2) Drop embed_code — video_src already holds the preserved value for former embed-only rows.
        ALTER TABLE videos DROP COLUMN embed_code;
    END IF;
END $$;

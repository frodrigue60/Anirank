ALTER TABLE public.song_variants ADD COLUMN IF NOT EXISTS episodes text;
ALTER TABLE public.song_variants ADD COLUMN IF NOT EXISTS nsfw boolean DEFAULT false NOT NULL;

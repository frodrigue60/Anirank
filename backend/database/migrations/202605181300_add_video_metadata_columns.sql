-- Migration Name: 202605181300_add_video_metadata_columns.sql
-- Description: Adds metadata columns (is_nc, is_bd, resolution) to public.videos

DO $$ 
BEGIN
    -- Add is_nc column (No Credits)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'is_nc'
    ) THEN
        ALTER TABLE public.videos ADD COLUMN is_nc boolean DEFAULT false NOT NULL;
    END IF;

    -- Add is_bd column (Blu-ray Disc source)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'is_bd'
    ) THEN
        ALTER TABLE public.videos ADD COLUMN is_bd boolean DEFAULT false NOT NULL;
    END IF;

    -- Add resolution column (e.g. 1080, 720)
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'resolution'
    ) THEN
        ALTER TABLE public.videos ADD COLUMN resolution integer;
    END IF;
END $$;

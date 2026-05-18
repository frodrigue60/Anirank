-- Migration: Expand video metadata columns and multi-video support
-- Date: 2026-05-18
-- Author: Antigravity

-- 1. Idempotently add the new metadata columns to the public.videos table
ALTER TABLE videos ADD COLUMN IF NOT EXISTS is_uncensored BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS is_subbed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS is_lyrics BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE videos ADD COLUMN IF NOT EXISTS source VARCHAR(50) NOT NULL DEFAULT 'TV';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS overlap VARCHAR(50) NOT NULL DEFAULT 'None';

-- 2. Migrate existing records: if they were flagged as 'is_bd = true', set their source to 'BD'
UPDATE videos SET source = 'BD' WHERE is_bd = true;

-- 3. Add indices to speed up queries filtering by source, uncensored status, subbed, or lyrics
CREATE INDEX IF NOT EXISTS idx_videos_source ON videos(source);
CREATE INDEX IF NOT EXISTS idx_videos_is_uncensored ON videos(is_uncensored);
CREATE INDEX IF NOT EXISTS idx_videos_is_subbed ON videos(is_subbed);
CREATE INDEX IF NOT EXISTS idx_videos_is_lyrics ON videos(is_lyrics);

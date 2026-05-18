-- Migration: Add unique constraint on videos for song_variant_id and video_src
-- Date: 2026-05-18
-- Author: Antigravity

-- Create unique index to allow ON CONFLICT (song_variant_id, video_src) DO NOTHING/UPDATE
CREATE UNIQUE INDEX IF NOT EXISTS idx_videos_variant_src ON videos(song_variant_id, video_src);

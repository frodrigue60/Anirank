-- Add Softban and Automod Indices
-- Author: Luis Rodz
-- Date: 2024-05-11

-- Add is_softbanned to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_softbanned BOOLEAN DEFAULT FALSE;

-- Add indices for rate limiting efficiency
CREATE INDEX IF NOT EXISTS idx_comments_user_created ON comments(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_song_ratings_user_created ON song_ratings(user_id, created_at DESC);

-- Add index for pending reports count
CREATE INDEX IF NOT EXISTS idx_user_reports_status_reported ON user_reports(status, reported_user_id);
CREATE INDEX IF NOT EXISTS idx_comment_reports_status_comment ON comment_reports(status, comment_id);

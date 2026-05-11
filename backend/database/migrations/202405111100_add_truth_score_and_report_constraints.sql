-- Add truth_score to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS truth_score INT DEFAULT 100;

-- Ensure unique reports per entity per user while status is pending (false)
-- Song Reports
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_song_report_pending 
ON song_reports (user_id, song_id) 
WHERE status = false;

-- Comment Reports
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_comment_report_pending 
ON comment_reports (user_id, comment_id) 
WHERE status = false;

-- User Reports
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_report_pending 
ON user_reports (reporter_user_id, reported_user_id) 
WHERE status = false;

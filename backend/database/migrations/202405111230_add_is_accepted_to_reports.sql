-- Add is_accepted to report tables
ALTER TABLE song_reports ADD COLUMN IF NOT EXISTS is_accepted BOOLEAN DEFAULT false;
ALTER TABLE comment_reports ADD COLUMN IF NOT EXISTS is_accepted BOOLEAN DEFAULT false;
ALTER TABLE user_reports ADD COLUMN IF NOT EXISTS is_accepted BOOLEAN DEFAULT false;

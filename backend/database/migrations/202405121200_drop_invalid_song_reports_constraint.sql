-- Drop the invalid status check constraint on song_reports
-- The constraint incorrectly expected 'fixed' or 'pending' strings while the column is BOOLEAN
ALTER TABLE song_reports DROP CONSTRAINT IF EXISTS reports_status_check;

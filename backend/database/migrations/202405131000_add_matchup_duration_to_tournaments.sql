-- Add matchup_duration_hours to tournaments table
-- Default to 48 hours for backward compatibility
ALTER TABLE tournaments ADD COLUMN matchup_duration_hours INT DEFAULT 48;

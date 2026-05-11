-- Add snapshot columns for immutable evidence
ALTER TABLE song_reports ADD COLUMN snapshot TEXT;
ALTER TABLE comment_reports ADD COLUMN snapshot TEXT;
ALTER TABLE user_reports ADD COLUMN snapshot JSONB;

-- Comment for documentation
COMMENT ON COLUMN comment_reports.snapshot IS 'Immutable copy of the comment content at the time of report';
COMMENT ON COLUMN user_reports.snapshot IS 'Immutable snapshot of the reported user profile (name, bio, avatar) at the time of report';

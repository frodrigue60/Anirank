-- Alter the default settings JSON for new users to include all notification keys
ALTER TABLE user_notification_settings 
    ALTER COLUMN settings SET DEFAULT '{"social_follow":true,"comment_reply":true,"user_request_feedback":true,"artist_new_song":true,"level_up":true,"badge_award":true}'::json;

-- Update existing rows to merge the new defaults seamlessly without overwriting existing preferences
UPDATE user_notification_settings
SET settings = (
    settings::jsonb || 
    '{"artist_new_song":true,"level_up":true,"badge_award":true,"user_request_feedback":true}'::jsonb
)::json
WHERE settings IS NOT NULL;

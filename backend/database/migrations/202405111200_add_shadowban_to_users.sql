-- Add is_shadowbanned to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_shadowbanned BOOLEAN DEFAULT false;

-- Index for shadowbanned users
CREATE INDEX IF NOT EXISTS idx_users_shadowbanned ON users (is_shadowbanned) WHERE is_shadowbanned = true;

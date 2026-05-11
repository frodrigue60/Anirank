-- Add is_shadowbanned to comments and song_ratings
ALTER TABLE comments ADD COLUMN IF NOT EXISTS is_shadowbanned BOOLEAN DEFAULT false;
ALTER TABLE song_ratings ADD COLUMN IF NOT EXISTS is_shadowbanned BOOLEAN DEFAULT false;
ALTER TABLE song_reactions ADD COLUMN IF NOT EXISTS is_shadowbanned BOOLEAN DEFAULT false;
ALTER TABLE comment_reactions ADD COLUMN IF NOT EXISTS is_shadowbanned BOOLEAN DEFAULT false;

-- Index for filtering shadowbanned content
CREATE INDEX IF NOT EXISTS idx_comments_shadowbanned ON comments (is_shadowbanned) WHERE is_shadowbanned = true;
CREATE INDEX IF NOT EXISTS idx_song_ratings_shadowbanned ON song_ratings (is_shadowbanned) WHERE is_shadowbanned = true;
CREATE INDEX IF NOT EXISTS idx_song_reactions_shadowbanned ON song_reactions (is_shadowbanned) WHERE is_shadowbanned = true;
CREATE INDEX IF NOT EXISTS idx_comment_reactions_shadowbanned ON comment_reactions (is_shadowbanned) WHERE is_shadowbanned = true;

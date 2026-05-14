-- Ensure all interaction tables have is_shadowbanned column (Idempotent)
-- This fixes issues where previous migrations might have been partially applied or skipped

DO $$ 
BEGIN
    -- comments
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='comments' AND column_name='is_shadowbanned') THEN
        ALTER TABLE comments ADD COLUMN is_shadowbanned BOOLEAN DEFAULT false;
    END IF;

    -- song_ratings
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='song_ratings' AND column_name='is_shadowbanned') THEN
        ALTER TABLE song_ratings ADD COLUMN is_shadowbanned BOOLEAN DEFAULT false;
    END IF;

    -- song_reactions
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='song_reactions' AND column_name='is_shadowbanned') THEN
        ALTER TABLE song_reactions ADD COLUMN is_shadowbanned BOOLEAN DEFAULT false;
    END IF;

    -- comment_reactions
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='comment_reactions' AND column_name='is_shadowbanned') THEN
        ALTER TABLE comment_reactions ADD COLUMN is_shadowbanned BOOLEAN DEFAULT false;
    END IF;
END $$;

-- Ensure indices exist
CREATE INDEX IF NOT EXISTS idx_comments_shadowbanned ON comments (is_shadowbanned) WHERE is_shadowbanned = true;
CREATE INDEX IF NOT EXISTS idx_song_ratings_shadowbanned ON song_ratings (is_shadowbanned) WHERE is_shadowbanned = true;
CREATE INDEX IF NOT EXISTS idx_song_reactions_shadowbanned ON song_reactions (is_shadowbanned) WHERE is_shadowbanned = true;
CREATE INDEX IF NOT EXISTS idx_comment_reactions_shadowbanned ON comment_reactions (is_shadowbanned) WHERE is_shadowbanned = true;

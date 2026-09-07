-- Read-only generated playlist query support.
CREATE INDEX IF NOT EXISTS idx_song_ratings_user_updated_song
    ON song_ratings (user_id, updated_at DESC, song_id);

CREATE INDEX IF NOT EXISTS idx_song_reactions_likes_user_updated_song
    ON song_reactions (user_id, updated_at DESC, song_id)
    WHERE type = 1;

CREATE INDEX IF NOT EXISTS idx_songs_year_status_rating_popularity
    ON songs (year_id, status, average_score DESC, likes_count DESC, views DESC);

CREATE INDEX IF NOT EXISTS idx_songs_year_season_status_rating_popularity
    ON songs (year_id, season_id, status, average_score DESC, likes_count DESC, views DESC);

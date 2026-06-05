-- Idempotent migration to add indices on foreign keys used by AMQ song query
CREATE INDEX IF NOT EXISTS idx_songs_anime_id ON songs(anime_id);
CREATE INDEX IF NOT EXISTS idx_songs_type_id ON songs(type_id);
CREATE INDEX IF NOT EXISTS idx_song_variants_song_id ON song_variants(song_id);
CREATE INDEX IF NOT EXISTS idx_videos_song_variant_id ON videos(song_variant_id);

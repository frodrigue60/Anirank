CREATE TABLE IF NOT EXISTS songs (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    song_romaji VARCHAR(255),
    song_jp VARCHAR(255),
    song_en VARCHAR(255),
    theme_num VARCHAR(10) DEFAULT '1',
    type VARCHAR(10) DEFAULT 'OP' CHECK (type IN ('OP', 'ED', 'INS', 'OTH')),
    slug VARCHAR(255) NOT NULL,
    status BOOLEAN DEFAULT FALSE,
    views BIGINT DEFAULT 0,
    likes_count BIGINT DEFAULT 0,
    dislikes_count BIGINT DEFAULT 0,
    favorites_count INTEGER DEFAULT 0,
    average_score DECIMAL(5, 2) DEFAULT 0.00,
    prev_main_rank INTEGER,
    prev_seasonal_rank INTEGER,
    animethemes_id VARCHAR(255) UNIQUE,
    anime_id BIGINT NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    season_id BIGINT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    year_id BIGINT NOT NULL REFERENCES years(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_anime_theme UNIQUE (anime_id, type, theme_num)
);

CREATE INDEX IF NOT EXISTS songs_type_idx ON songs (type);
CREATE INDEX IF NOT EXISTS songs_views_idx ON songs (views);
CREATE INDEX IF NOT EXISTS songs_slug_idx ON songs (slug);
CREATE INDEX IF NOT EXISTS songs_status_idx ON songs (status);
CREATE INDEX IF NOT EXISTS idx_songs_ranks ON songs (prev_main_rank, prev_seasonal_rank);

CREATE TABLE IF NOT EXISTS song_variants (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    version_number BIGINT DEFAULT 1,
    song_id BIGINT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    views BIGINT DEFAULT 0,
    slug VARCHAR(255) NOT NULL,
    spoiler BOOLEAN DEFAULT FALSE,
    status BOOLEAN DEFAULT FALSE,
    season_id BIGINT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    year_id BIGINT NOT NULL REFERENCES years(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_song_version UNIQUE (song_id, version_number)
);

CREATE TABLE IF NOT EXISTS videos (
    id BIGSERIAL PRIMARY KEY,
    video_src VARCHAR(255),
    embed_code TEXT,
    type VARCHAR(50) DEFAULT 'file',
    status BOOLEAN DEFAULT FALSE,
    song_variant_id BIGINT NOT NULL REFERENCES song_variants(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS artist_song (
    artist_id BIGINT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    song_id BIGINT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    PRIMARY KEY (artist_id, song_id)
);

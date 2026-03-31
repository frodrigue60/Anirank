CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified_at TIMESTAMP WITH TIME ZONE NULL,
    password VARCHAR(255) NULL,
    last_login_at TIMESTAMP WITH TIME ZONE NULL,
    score_format_id BIGINT NULL REFERENCES score_formats(id) ON DELETE RESTRICT,
    about VARCHAR(500) NULL,
    profile_color VARCHAR(20) NULL,
    banner VARCHAR(255) NULL,
    xp BIGINT DEFAULT 0,
    level INTEGER DEFAULT 1,
    anilist_id BIGINT UNIQUE NULL,
    anilist_username VARCHAR(191) NULL,
    anilist_access_token TEXT NULL,
    anilist_refresh_token TEXT NULL,
    anilist_token_expires_at TIMESTAMP WITH TIME ZONE NULL,
    google_id VARCHAR(255) UNIQUE NULL,
    google_email VARCHAR(255) UNIQUE NULL,
    google_access_token TEXT NULL,
    google_refresh_token TEXT NULL,
    google_token_expires_at TIMESTAMP WITH TIME ZONE NULL,
    remember_token VARCHAR(100) NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_google_id ON users (google_id);
CREATE INDEX IF NOT EXISTS idx_users_google_email ON users (google_email);

CREATE TABLE IF NOT EXISTS password_resets (
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX IF NOT EXISTS password_resets_email_index ON password_resets (email);

CREATE TABLE IF NOT EXISTS personal_access_tokens (
    id BIGSERIAL PRIMARY KEY,
    tokenable_type VARCHAR(255) NOT NULL,
    tokenable_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    token VARCHAR(64) UNIQUE NOT NULL,
    abilities TEXT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE NULL,
    expires_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS personal_access_tokens_tokenable_idx ON personal_access_tokens (tokenable_type, tokenable_id);

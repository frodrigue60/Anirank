CREATE TABLE IF NOT EXISTS anime_studio (
    anime_id BIGINT NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    studio_id BIGINT NOT NULL REFERENCES studios(id) ON DELETE CASCADE,
    PRIMARY KEY (anime_id, studio_id)
);

CREATE TABLE IF NOT EXISTS anime_producer (
    anime_id BIGINT NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    producer_id BIGINT NOT NULL REFERENCES producers(id) ON DELETE CASCADE,
    PRIMARY KEY (anime_id, producer_id)
);

CREATE TABLE IF NOT EXISTS anime_external_link (
    anime_id BIGINT NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    external_link_id BIGINT NOT NULL REFERENCES external_links(id) ON DELETE CASCADE,
    PRIMARY KEY (anime_id, external_link_id)
);

CREATE TABLE IF NOT EXISTS anime_genre (
    anime_id BIGINT NOT NULL REFERENCES animes(id) ON DELETE CASCADE,
    genre_id BIGINT NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (anime_id, genre_id)
);

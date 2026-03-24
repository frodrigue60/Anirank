package dto

type AnimeMinimalDTO struct {
	ID           uint64  `json:"id"`
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	ThumbnailUrl *string `json:"thumbnail_url,omitempty"`
	BannerUrl    *string `json:"banner_url,omitempty"`
	SongsCount   int     `json:"songs_count"`
	Season       *string `json:"season,omitempty"`
	Year         *string `json:"year,omitempty"`
	Format       *string `json:"format,omitempty"`
}

type SongAnimeDTO struct {
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	ThumbnailUrl string  `json:"thumbnail_url"`
	BannerUrl    *string `json:"banner_url,omitempty"`
}

type AnimeDTO struct {
	AnimeMinimalDTO
	Description   *string            `json:"description,omitempty"`
	Studios       []StudioDTO       `json:"studios,omitempty"`
	Producers     []ProducerDTO     `json:"producers,omitempty"`
	Genres        []GenreDTO        `json:"genres,omitempty"`
	ExternalLinks []ExternalLinkDTO `json:"external_links,omitempty"`
}

type StudioDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProducerDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type GenreDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ExternalLinkDTO struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
	Icon *string `json:"icon,omitempty"`
}

package dto

type AnimeMinimalDTO struct {
	ID         uint64     `json:"id"`
	Title      string     `json:"title"`
	Slug       string     `json:"slug"`
	CoverUrl   *string    `json:"cover_url,omitempty"`
	BannerUrl  *string    `json:"banner_url,omitempty"`
	SongsCount int        `json:"songs_count"`
	Season     *SeasonDTO `json:"season,omitempty"`
	Year       *YearDTO   `json:"year,omitempty"`
	Format     *FormatDTO `json:"format,omitempty"`
}

type SeasonDTO struct {
	Name string `json:"name"`
}

type YearDTO struct {
	Name string `json:"name"`
}

type FormatDTO struct {
	Name string `json:"name"`
}

type SongAnimeDTO struct {
	Title     string  `json:"title"`
	Slug      string  `json:"slug"`
	CoverUrl  string  `json:"cover_url"`
	BannerUrl *string `json:"banner_url,omitempty"`
}

type AnimeDTO struct {
	AnimeMinimalDTO
	Description   *string            `json:"description,omitempty"`
	Studios       []StudioDTO       `json:"studios,omitempty"`
	Producers     []ProducerDTO     `json:"producers,omitempty"`
	Genres        []GenreDTO        `json:"genres,omitempty"`
	ExternalLinks []ExternalLinkDTO `json:"external_links,omitempty"`
}

type ExternalLinkDTO struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
	Icon *string `json:"icon,omitempty"`
}

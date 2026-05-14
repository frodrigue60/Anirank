package dto

import (
	"anirank/api/internal/domain"
)

type AnimeMinimalDTO struct {
	AnilistID  *int64     `json:"anilist_id"`
	Title      string     `json:"title"`
	Slug       string     `json:"slug"`
	CoverUrl   *string    `json:"cover_url"`
	CoverSources []domain.ImageSource `json:"cover_sources,omitempty"`
	BannerUrl  *string    `json:"banner_url"`
	BannerSources []domain.ImageSource `json:"banner_sources,omitempty"`
	SongsCount    int        `json:"songs_count"`
	EnabledSongs  int        `json:"enabled_songs"`
	DisabledSongs int        `json:"disabled_songs"`
	Season        *SeasonDTO `json:"season,omitempty"`
	Year          *YearDTO   `json:"year,omitempty"`
	Format        *FormatDTO `json:"format,omitempty"`
}

type SeasonDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type YearDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type FormatDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SongAnimeDTO struct {
	Title     string  `json:"title"`
	Slug      string  `json:"slug"`
	CoverUrl  string  `json:"cover_url"`
	CoverSources []domain.ImageSource `json:"cover_sources,omitempty"`
	BannerUrl *string `json:"banner_url"`
	BannerSources []domain.ImageSource `json:"banner_sources,omitempty"`
	AnilistID *int64  `json:"anilist_id"`
	Season    *SeasonDTO `json:"season,omitempty"`
	Year      *YearDTO   `json:"year,omitempty"`
}


type AnimeDTO struct {
	AnimeMinimalDTO
	Description   *string            `json:"description,omitempty"`
	Studios       []StudioDTO       `json:"studios,omitempty"`
	Producers     []ProducerDTO     `json:"producers,omitempty"`
	Genres        []GenreDTO        `json:"genres,omitempty"`
	Songs         []SongMinimalDTO  `json:"songs,omitempty"`
	ExternalLinks []ExternalLinkDTO `json:"external_links,omitempty"`
}

type ExternalLinkDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
	Icon *string `json:"icon,omitempty"`
}

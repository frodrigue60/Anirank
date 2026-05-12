package dto

import "anirank/api/internal/domain"

type StudioDTO struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	LogoUrl    *string `json:"logo_url,omitempty"`
	BannerUrl  *string `json:"banner_url,omitempty"`
	LogoSources   []domain.ImageSource `json:"logo_sources,omitempty"`
	BannerSources []domain.ImageSource `json:"banner_sources,omitempty"`
	AnimeCount int     `json:"anime_count"`
}

type ProducerDTO struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	LogoUrl    *string `json:"logo_url,omitempty"`
	BannerUrl  *string `json:"banner_url,omitempty"`
	LogoSources   []domain.ImageSource `json:"logo_sources,omitempty"`
	BannerSources []domain.ImageSource `json:"banner_sources,omitempty"`
	AnimeCount int     `json:"anime_count"`
}

type GenreDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type InitDataDTO struct {
	Years     []YearDTO     `json:"years"`
	Seasons   []SeasonDTO   `json:"seasons"`
	Formats   []FormatDTO   `json:"formats"`
	Genres    []GenreDTO    `json:"genres"`
	SongTypes []SongTypeDTO `json:"song_types"`
}

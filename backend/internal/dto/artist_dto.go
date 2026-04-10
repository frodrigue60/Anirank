package dto

import (
	"anirank/api/internal/domain"
)

type ArtistMinimalDTO struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	NameJP        *string              `json:"name_jp,omitempty"`
	Slug          string               `json:"slug"`
	AvatarUrl     *string              `json:"avatar_url"`
	AvatarSources []domain.ImageSource `json:"avatar_sources,omitempty"`
	BannerUrl     *string              `json:"banner_url,omitempty"`
	BannerSources []domain.ImageSource `json:"banner_sources,omitempty"`
	EnabledSongs  int                  `json:"enabled_songs"`
	DisabledSongs int                  `json:"disabled_songs"`
}

type ArtistDTO struct {
	ArtistMinimalDTO
	FavoritesCount uint64 `json:"favorites_count"`
	IsFavorited    bool   `json:"is_favorited"`
}

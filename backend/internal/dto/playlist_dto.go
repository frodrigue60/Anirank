package dto

import (
	"anirank/api/internal/domain"
	"time"
)

type PlaylistMinimalDTO struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	BannerUrl     *string `json:"banner_url,omitempty"`
	BannerSources []domain.ImageSource `json:"banner_sources,omitempty"`
	SongCount    int     `json:"song_count"`
	IsPublic     bool    `json:"is_public"`
	ContainsSong bool    `json:"contains_song,omitempty"`
}

type PlaylistDTO struct {
	PlaylistMinimalDTO
	Description *string        `json:"description,omitempty"`
	User        UserMinimalDTO `json:"user"`
	Songs       []SongMinimalDTO `json:"songs,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

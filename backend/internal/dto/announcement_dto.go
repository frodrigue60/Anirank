package dto

import (
	"anirank/api/internal/domain"
	"time"
)

type AnnouncementDTO struct {
	ID        any        `json:"id"` // Numeric for Admin, UUID for Public
	UUID      string     `json:"uuid,omitempty"`
	Title     string     `json:"title"`
	Content   *string    `json:"content,omitempty"`
	Type      string     `json:"type"`
	Icon      *string    `json:"icon,omitempty"`
	URL       *string    `json:"url,omitempty"`
	ImageUrl  *string             `json:"image_url,omitempty"`
	ImageSources []domain.ImageSource `json:"image_sources,omitempty"`
	Priority  int                `json:"priority"`
	IsActive  bool       `json:"is_active"`
	StartsAt  *time.Time `json:"starts_at,omitempty"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
}

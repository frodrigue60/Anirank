package domain

import (
	"context"
	"time"
)

type Announcement struct {
	ID        uint64     `db:"id" json:"admin_id"`
	UUID      string     `db:"uuid" json:"id"`
	Title     string     `db:"title" json:"title" form:"title"`
	Content   *string    `db:"content" json:"content" form:"content"`
	Type      string     `db:"type" json:"type" form:"type"`
	Icon      *string    `db:"icon" json:"icon" form:"icon"`
	URL       *string    `db:"url" json:"url" form:"url"`
	Image     *string    `db:"image" json:"image"` // Base path in S3
	ImageUrl  *string    `db:"-" json:"image_url,omitempty"` // Computed URL
	ImageSources []ImageSource `db:"-" json:"image_sources,omitempty"` // Computed resolutions
	Priority  int        `db:"priority" json:"priority" form:"priority"`
	IsActive  bool       `db:"is_active" json:"is_active" form:"is_active"`
	StartsAt  *time.Time `db:"starts_at" json:"starts_at"`
	EndsAt    *time.Time `db:"ends_at" json:"ends_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

type AnnouncementFilters struct {
	ActiveOnly bool
	Search     string
}

type AnnouncementRepository interface {
	GetByID(ctx context.Context, id uint64) (*Announcement, error)
	GetByUUID(ctx context.Context, uuid string) (*Announcement, error)
	GetAll(ctx context.Context, filters AnnouncementFilters, limit, offset int) ([]Announcement, error)
	Count(ctx context.Context, filters AnnouncementFilters) (int, error)
	GetActive(ctx context.Context) ([]Announcement, error)
	Create(ctx context.Context, a *Announcement) error
	Update(ctx context.Context, a *Announcement) error
	Delete(ctx context.Context, id uint64) error
	ToggleActive(ctx context.Context, id uint64) error
}

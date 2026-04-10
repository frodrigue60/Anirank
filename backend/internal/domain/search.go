package domain

import "context"

type SearchIndexItem struct {
	ID        uint64  `db:"id" json:"-"`
	ItemType  string  `db:"item_type" json:"item_type"`
	ItemUUID  string  `db:"item_id" json:"item_uuid"`
	Title     string  `db:"title" json:"title"`
	Subtitle  *string `db:"subtitle" json:"subtitle"`
	Slug      string  `db:"slug" json:"slug"`
	ImageURL  *string `db:"image_url" json:"image_url"`
	AnimeSlug *string `db:"-" json:"anime_slug,omitempty"` // For song navigation
	Rank      float64 `db:"rank" json:"rank"`
}

type SearchRepository interface {
	GlobalSearch(ctx context.Context, term string, limit int) ([]SearchIndexItem, error)
}

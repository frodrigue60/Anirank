package domain

import (
	"context"
	"time"
)

// Playlist represents a user collection of songs
type Playlist struct {
	ID          uint64    `db:"id" json:"id"`
	UUID        string    `db:"uuid" json:"uuid"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description"`
	UserID      uint64    `db:"user_id" json:"user_id"`
	IsPublic    bool      `db:"is_public" json:"is_public"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	// Relationships
	User  *User          `db:"-" json:"user,omitempty"`
	Songs []PlaylistSong `db:"-" json:"songs,omitempty"`

	// Enriched
	SongCount    int     `db:"song_count" json:"song_count"`
	ContainsSong bool    `db:"contains_song" json:"contains_song"`
	LatestBanner *string `db:"latest_banner" json:"-"`
	BannerUrl    *string `db:"-" json:"banner_url"`
}

type PlaylistFilters struct {
	Search   string
	IsPublic *bool
	UserID   *uint64
	Sort     string
}

// PlaylistSong represents the pivot row enriched with Song data
type PlaylistSong struct {
	PlaylistID uint64 `db:"playlist_id" json:"playlist_id"`
	SongID     uint64 `db:"song_id" json:"song_id"`
	Position   int    `db:"position" json:"position"`

	// Enriched
	Song *Song `db:"-" json:"song,omitempty"`
}

type PlaylistRepository interface {
	// CRUD
	GetByID(ctx context.Context, id uint64) (*Playlist, error)
	GetByUUID(ctx context.Context, uuid string) (*Playlist, error)
	GetByUserID(ctx context.Context, userID uint64, includePrivate bool, limit, offset int) ([]Playlist, error)
	GetByUserIDWithSongCheck(ctx context.Context, userID, songID uint64, includePrivate bool, limit, offset int) ([]Playlist, error)
	CountByUserID(ctx context.Context, userID uint64, includePrivate bool) (int, error)
	Create(ctx context.Context, playlist *Playlist) error
	Update(ctx context.Context, playlist *Playlist) error
	Delete(ctx context.Context, id, userID uint64) error

	// Songs Management
	AddSong(ctx context.Context, playlistID, songID uint64, position int) error
	RemoveSong(ctx context.Context, playlistID, songID uint64) error
	UpdateSongPositions(ctx context.Context, playlistID uint64, items []PlaylistSong) error

	// Retrieval
	GetSongs(ctx context.Context, playlistID uint64) ([]PlaylistSong, error)

	// Public discovery
	GetPaginatedPublicPlaylists(ctx context.Context, limit, offset int, filters PlaylistFilters) ([]Playlist, error)
	CountPublicPlaylists(ctx context.Context, filters PlaylistFilters) (int, error)
}

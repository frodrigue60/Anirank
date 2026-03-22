package domain

import (
	"context"
	"time"
)

type Song struct {
	ID            uint64    `db:"id" json:"id"`
	SongRomaji    *string   `db:"song_romaji" json:"song_romaji"`
	SongJP        *string   `db:"song_jp" json:"song_jp"`
	SongEN        *string   `db:"song_en" json:"song_en"`
	ThemeNum      string    `db:"theme_num" json:"theme_num"`
	Type          string    `db:"type" json:"type"`
	Slug          string    `db:"slug" json:"slug"`
	AnimeID       uint64    `db:"anime_id" json:"anime_id"`
	SeasonID      uint64    `db:"season_id" json:"season_id"`
	YearID        uint64    `db:"year_id" json:"year_id"`
	Views         uint64    `db:"views" json:"views"`
	LikesCount    uint64    `db:"likes_count" json:"likes_count"`
	DislikesCount uint64    `db:"dislikes_count" json:"dislikes_count"`
	FavoritesCount uint64    `db:"favorites_count" json:"favorites_count"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
	Status        bool      `db:"status" json:"status"`

	// Computed for frontend
	Name          string  `db:"-" json:"name"`
	TypeName      string  `db:"-" json:"type_name"`
	AverageRating float64  `db:"average_score" json:"average_rating"`
	UserRating    *float64 `db:"-" json:"user_rating,omitempty"`
	IsFavorited   bool     `db:"-" json:"is_favorited"`
	IsLiked       bool     `db:"-" json:"is_liked"`
	IsDisliked    bool     `db:"-" json:"is_disliked"`
	IsReported    bool     `db:"-" json:"is_reported"`
	PrevRank         *int    `db:"prev_rank" json:"prev_rank,omitempty"`
	PrevMainRank     *int    `db:"prev_main_rank" json:"-"`
	PrevSeasonalRank *int    `db:"prev_seasonal_rank" json:"-"`
	PartialArtistInactive bool `db:"partial_artist_inactive" json:"partial_artist_inactive"`

	// Relations
	Anime    *Anime        `db:"anime" json:"anime,omitempty"`
	Season   *Season       `db:"-" json:"season,omitempty"`
	Year     *Year         `db:"-" json:"year,omitempty"`
	Variants []SongVariant `db:"-" json:"song_variants,omitempty"`
	Artists  []Artist      `db:"-" json:"artists,omitempty"`
}

type SongVariant struct {
	ID            uint64    `db:"id" json:"id"`
	VersionNumber uint64    `db:"version_number" json:"version_number"`
	SongID        uint64    `db:"song_id" json:"song_id"`
	Slug          string    `db:"slug" json:"slug"`
	Views         uint64    `db:"views" json:"views"`
	SeasonID      uint64    `db:"season_id" json:"season_id"`
	YearID        uint64    `db:"year_id" json:"year_id"`
	Spoiler       bool      `db:"spoiler" json:"spoiler"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
	Status        bool      `db:"status" json:"status"`

	// Computed for frontend
	Video *SongVariantVideo `db:"-" json:"video,omitempty"`

	// Relations
	Song   *Song   `db:"-" json:"song,omitempty"`
	Season *Season `db:"-" json:"season,omitempty"`
	Year   *Year   `db:"-" json:"year,omitempty"`
}

type SongVariantVideo struct {
	Type     string  `json:"type"`
	EmbedUrl *string `json:"embed_url,omitempty"`
	LocalUrl *string `json:"local_url,omitempty"`
}

type Artist struct {
	ID        uint64    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	NameJP    *string   `db:"name_jp" json:"name_jp"`
	Slug      string    `db:"slug" json:"slug"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Status    bool      `db:"status" json:"status"`

	// Image fields
	Avatar    *string `db:"avatar" json:"-"`
	AvatarUrl *string `db:"-" json:"avatar_url,omitempty"`

	// Computed
	SongsCount     int    `db:"songs_count" json:"songs_count"`
	FavoritesCount uint64 `db:"favorites_count" json:"favorites_count"`
	IsFavorited    bool   `db:"-" json:"is_favorited"`
}

type ArtistFilters struct {
	Search  string
	Sort    string
	IsAdmin bool
}

type SongFilters struct {
	Search   string
	YearID   uint64
	SeasonID uint64
	GenreID  uint64
	AnimeID  uint64
	Status   *bool
	Type     string
	Sort     string
	IsAdmin  bool
}

// Repositories
type SongRepository interface {
	GetByID(ctx context.Context, id uint64) (*Song, error)
	GetBySlug(ctx context.Context, slug string) (*Song, error)
	GetByAnimeIDAndSlug(ctx context.Context, animeID uint64, slug string) (*Song, error)
	GetPaginated(ctx context.Context, limit, offset int, filters SongFilters) ([]Song, error)
	GetByAnimeID(ctx context.Context, animeID uint64) ([]Song, error)
	GetByArtistID(ctx context.Context, artistID uint64, limit, offset int, filters SongFilters) ([]Song, error)
	GetRanking(ctx context.Context, rankingType, songType string, limit, offset int) ([]Song, error)
	Count(ctx context.Context, filters SongFilters) (int, error)
	CountByArtistID(ctx context.Context, artistID uint64, filters SongFilters) (int, error)
	CountFavoritesByUserID(ctx context.Context, userID uint64) (int, error)
	GetFavoritesByUserID(ctx context.Context, userID uint64, limit, offset int) ([]Song, error)
	CountRanking(ctx context.Context, rankingType, songType string) (int, error)
	IncrementViews(ctx context.Context, id uint64) error
	GetMany(ctx context.Context, ids []uint64) ([]Song, error)

	// Admin CRUD
	Create(ctx context.Context, song *Song) error
	Update(ctx context.Context, song *Song) error
	Delete(ctx context.Context, id uint64) error

	// Relations
	GetVariantsBySongID(ctx context.Context, songID uint64) ([]SongVariant, error)
	GetArtistsBySongID(ctx context.Context, songID uint64, isAdmin bool) ([]Artist, error)
	SyncArtists(ctx context.Context, songID uint64, artistIDs []uint64) error
	ToggleStatus(ctx context.Context, id uint64) error

	// Sitemap
	GetPublicSlugs(ctx context.Context) ([]SitemapItem, error)
}

type SongVariantRepository interface {
	GetByID(ctx context.Context, id uint64) (*SongVariant, error)
	GetPaginated(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]SongVariant, error)
	Count(ctx context.Context, filters map[string]interface{}) (int, error)
	IncrementViews(ctx context.Context, id uint64) error

	// Admin CRUD
	Create(ctx context.Context, variant *SongVariant) error
	Update(ctx context.Context, variant *SongVariant) error
	Delete(ctx context.Context, id uint64) error
	ToggleStatus(ctx context.Context, id uint64) error
}

type ArtistRepository interface {
	GetByID(ctx context.Context, id uint64) (*Artist, error)
	GetBySlug(ctx context.Context, slug string) (*Artist, error)
	GetPaginated(ctx context.Context, limit, offset int, filters ArtistFilters) ([]Artist, error)
	Count(ctx context.Context, filters ArtistFilters) (int, error)

	// Catalog queries
	CountFavoritesByUserID(ctx context.Context, userID uint64) (int, error)
	GetFavoritesByUserID(ctx context.Context, userID uint64, limit, offset int) ([]Artist, error)

	// Admin CRUD
	Create(ctx context.Context, artist *Artist) error
	Update(ctx context.Context, artist *Artist) error
	UpdateAvatar(ctx context.Context, id uint64, avatar string) error
	Delete(ctx context.Context, id uint64) error
	ToggleStatus(ctx context.Context, id uint64) error

	// Sitemap
	GetPublicSlugs(ctx context.Context) ([]SitemapItem, error)
}

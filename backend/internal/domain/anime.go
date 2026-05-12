package domain

import (
	"context"
	"time"
)

// Anime
type Anime struct {
	ID          uint64  `db:"id" json:"id"`
	UUID        string  `db:"uuid" json:"uuid"`
	Title       string  `db:"title" json:"title" form:"title"`
	Slug        string  `db:"slug" json:"slug" form:"slug"`
	Description *string `db:"description" json:"description" form:"description"`
	AnilistID   *int64  `db:"anilist_id" json:"anilist_id" form:"anilist_id"`
	Status      bool    `db:"status" json:"status" form:"status"`
	Cover       *string `db:"cover" json:"cover"`
	Banner      *string `db:"banner" json:"banner"`
	CoverUrl    *string `db:"-" json:"cover_url,omitempty"`
	BannerUrl   *string `db:"-" json:"banner_url,omitempty"`
	CoverSources []ImageSource `db:"-" json:"cover_sources,omitempty"`
	BannerSources []ImageSource `db:"-" json:"banner_sources,omitempty"`
	YearID      uint64  `json:"year_id" db:"year_id" form:"year_id"`
	SeasonID    uint64  `json:"season_id" db:"season_id" form:"season_id"`
	FormatID    uint64  `json:"format_id" db:"format_id" form:"format_id"`
	AnimeThemesID *uint64 `db:"anime_themes_id" json:"animethemes_id,omitempty"`

	// Input-only fields for relations (manual creation)
	StudiosString   string `json:"-" form:"studios"`
	ProducersString string `json:"-" form:"producers"`
	GenresString    string `json:"-" form:"genres"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relations
	Year          *Year          `json:"year,omitempty"`
	Season        *Season        `json:"season,omitempty"`
	Format        *Format        `json:"format,omitempty"`
	Songs         []Song         `db:"-" json:"songs,omitempty"`
	Studios       []Studio       `db:"-" json:"studios,omitempty"`
	Producers     []Producer     `db:"-" json:"producers,omitempty"`
	Genres        []Genre        `db:"-" json:"genres,omitempty"`
	ExternalLinks []ExternalLink `db:"-" json:"external_links,omitempty"`
	SongsCount    int            `db:"songs_count" json:"songs_count"`
	EnabledSongs  int            `db:"enabled_songs" json:"enabled_songs"`
	DisabledSongs int            `db:"disabled_songs" json:"disabled_songs"`
}

// Taxonomies
type Year struct {
	ID        uint64    `db:"id" json:"id"`
	UUID      string    `db:"uuid" json:"uuid"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	Current   bool      `db:"current" json:"current"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Season struct {
	ID        uint64    `db:"id" json:"id"`
	UUID      string    `db:"uuid" json:"uuid"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	Current   bool      `db:"current" json:"current"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Format struct {
	ID        uint64    `db:"id" json:"id"`
	UUID      string    `db:"uuid" json:"uuid"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Studio struct {
	ID        uint64    `db:"id" json:"id"`
	UUID      string    `db:"uuid" json:"uuid"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Computed
	Logo         *string `db:"logo" json:"logo,omitempty"`
	LogoUrl      *string `db:"-" json:"logo_url,omitempty"`
	AnimeCount   int     `db:"anime_count" json:"anime_count"`
	LatestBanner *string `db:"latest_banner" json:"latest_banner,omitempty"`
	BannerUrl    *string `db:"-" json:"banner_url,omitempty"`
	LogoSources   []ImageSource `db:"-" json:"logo_sources,omitempty"`
	BannerSources []ImageSource `db:"-" json:"banner_sources,omitempty"`
}

type Producer struct {
	ID        uint64    `db:"id" json:"id"`
	UUID      string    `db:"uuid" json:"uuid"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Computed
	Logo         *string `db:"logo" json:"logo,omitempty"`
	LogoUrl      *string `db:"-" json:"logo_url,omitempty"`
	AnimeCount   int     `db:"anime_count" json:"anime_count"`
	LatestBanner *string `db:"latest_banner" json:"latest_banner,omitempty"`
	BannerUrl    *string `db:"-" json:"banner_url,omitempty"`
	LogoSources   []ImageSource `db:"-" json:"logo_sources,omitempty"`
	BannerSources []ImageSource `db:"-" json:"banner_sources,omitempty"`
}

type Genre struct {
	ID        uint64    `db:"id" json:"id"`
	UUID      string    `db:"uuid" json:"uuid"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ExternalLink struct {
	ID        uint64    `db:"id" json:"id"`
	UUID      string    `db:"uuid" json:"uuid"`
	Icon      *string   `db:"icon" json:"icon"`
	Type      string    `db:"type" json:"type"`
	Name      string    `db:"name" json:"name"`
	URL       string    `db:"url" json:"url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type AnimeFilters struct {
	Year    string
	Season  string
	Format  string
	Genre   string
	Status  *bool
	Search  string
	Sort    string
	IsAdmin bool
}

type StudioFilters struct {
	Search string
	Sort   string
}

type ProducerFilters struct {
	Search string
	Sort   string
}

// Repositories
type AnimeRepository interface {
	GetByID(ctx context.Context, id uint64) (*Anime, error)
	GetByUUID(ctx context.Context, uuid string) (*Anime, error)
	GetMany(ctx context.Context, ids []uint64) ([]Anime, error)
	GetBySlug(ctx context.Context, slug string) (*Anime, error)
	GetByAnilistID(ctx context.Context, anilistID int64) (*Anime, error)
	GetByAnilistIDs(ctx context.Context, anilistIDs []int) ([]Anime, error)
	GetPaginated(ctx context.Context, limit, offset int, filters AnimeFilters) ([]Anime, error)
	Count(ctx context.Context, filters AnimeFilters) (int, error)
	Create(ctx context.Context, anime *Anime) error
	Update(ctx context.Context, anime *Anime) error
	Delete(ctx context.Context, id uint64) error

	// Relations Loaders
	LoadRelations(ctx context.Context, anime *Anime, isAdmin bool) error
	LoadManyRelations(ctx context.Context, animes []Anime, isAdmin bool) error

	// Search implementation
	Search(ctx context.Context, term string, limit int) ([]Anime, error)

	// Admin
	ToggleStatus(ctx context.Context, id uint64) error

	// Relational Updates
	UpdateStudios(ctx context.Context, animeID uint64, studioIDs []uint64) error
	UpdateGenres(ctx context.Context, animeID uint64, genreIDs []uint64) error
	UpdateProducers(ctx context.Context, animeID uint64, producerIDs []uint64) error
	UpdateExternalLinks(ctx context.Context, animeID uint64, links []ExternalLink) error
	RecountAnimeStats(ctx context.Context, ids []uint64) error
	BatchDelete(ctx context.Context, ids []uint64) error

	// Sitemap
	GetPublicSlugs(ctx context.Context) ([]SitemapItem, error)
}

type TaxonomyRepository interface {
	GetAllYears(ctx context.Context) ([]Year, error)
	GetAllSeasons(ctx context.Context) ([]Season, error)
	GetAllFormats(ctx context.Context) ([]Format, error)
	GetAllGenres(ctx context.Context) ([]Genre, error)
	GetAllStudios(ctx context.Context) ([]Studio, error)
	GetAllProducers(ctx context.Context) ([]Producer, error)

	SearchStudios(ctx context.Context, term string, limit int) ([]Studio, error)
	SearchProducers(ctx context.Context, term string, limit int) ([]Producer, error)
	SearchGenres(ctx context.Context, term string, limit int) ([]Genre, error)

	// Paginated listing for catalog pages
	GetPaginatedStudios(ctx context.Context, limit, offset int, filters StudioFilters) ([]Studio, error)
	GetStudioBySlug(ctx context.Context, slug string) (*Studio, error)
	GetAnimesByStudioID(ctx context.Context, studioID uint64, limit, offset int) ([]Anime, error)
	CountAnimesByStudioID(ctx context.Context, studioID uint64) (int, error)

	GetPaginatedProducers(ctx context.Context, limit, offset int, filters ProducerFilters) ([]Producer, error)
	GetProducerBySlug(ctx context.Context, slug string) (*Producer, error)
	GetAnimesByProducerID(ctx context.Context, producerID uint64, limit, offset int) ([]Anime, error)
	CountAnimesByProducerID(ctx context.Context, producerID uint64) (int, error)

	// Counts for pagination
	CountStudios(ctx context.Context, filters StudioFilters) (int, error)
	CountProducers(ctx context.Context, filters ProducerFilters) (int, error)

	// Utility
	GetCurrentYear(ctx context.Context) (*Year, error)
	GetCurrentSeason(ctx context.Context) (*Season, error)

	// Integration Utility
	GetOrCreateYear(ctx context.Context, name string) (*Year, error)
	GetOrCreateSeason(ctx context.Context, name string) (*Season, error)
	GetOrCreateFormat(ctx context.Context, name string) (*Format, error)
	GetOrCreateGenre(ctx context.Context, name string) (*Genre, error)
	GetOrCreateStudio(ctx context.Context, name string) (*Studio, error)
	GetOrCreateProducer(ctx context.Context, name string) (*Producer, error)

	GetByYear(ctx context.Context, name int) (*Year, error)
	GetBySeason(ctx context.Context, name string) (*Season, error)
	GetByFormat(ctx context.Context, name string) (*Format, error)
	GetByGenre(ctx context.Context, name string) (*Genre, error)
	GetByStudio(ctx context.Context, name string) (*Studio, error)
	GetByProducer(ctx context.Context, name string) (*Producer, error)

	// Admin CRUD
	GetYearByID(ctx context.Context, id uint64) (*Year, error)
	CreateYear(ctx context.Context, year *Year) error
	UpdateYear(ctx context.Context, year *Year) error
	DeleteYear(ctx context.Context, id uint64) error
	ToggleYearCurrent(ctx context.Context, id uint64) error

	GetSeasonByID(ctx context.Context, id uint64) (*Season, error)
	CreateSeason(ctx context.Context, season *Season) error
	UpdateSeason(ctx context.Context, season *Season) error
	DeleteSeason(ctx context.Context, id uint64) error
	ToggleSeasonCurrent(ctx context.Context, id uint64) error

	GetFormatByID(ctx context.Context, id uint64) (*Format, error)
	CreateFormat(ctx context.Context, format *Format) error
	UpdateFormat(ctx context.Context, format *Format) error
	DeleteFormat(ctx context.Context, id uint64) error

	GetGenreByID(ctx context.Context, id uint64) (*Genre, error)
	CreateGenre(ctx context.Context, genre *Genre) error
	UpdateGenre(ctx context.Context, genre *Genre) error
	DeleteGenre(ctx context.Context, id uint64) error
}

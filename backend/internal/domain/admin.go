package domain

import (
	"context"
	"time"
)

type DashboardStats struct {
	TotalAnimes     int `json:"total_animes"`
	TotalSongs      int `json:"total_songs"`
	PendingReports  int `json:"pending_reports"`
	PendingRequests int `json:"pending_requests"`
	ActiveUsersDay   int `json:"active_users_day"`
	TotalUsers       int `json:"total_users"`
	TotalArtists     int `json:"total_artists"`
	TotalPlaylists   int `json:"total_playlists"`
	TotalTournaments int `json:"total_tournaments"`

	// Pending Approval Detailed
	PendingAnimes   int `json:"pending_animes"`
	PendingSongs    int `json:"pending_songs"`
	PendingVariants int `json:"pending_variants"`
	PendingVideos   int `json:"pending_videos"`
	PendingArtists  int `json:"pending_artists"`

	// Reports Detailed
	SongReports    int `json:"song_reports"`
	CommentReports int `json:"comment_reports"`

	// Active Metrics
	ActiveTournaments int `json:"active_tournaments"`

	RecentReports  []SongReport  `json:"recent_reports"`
	RecentRequests []UserRequest `json:"recent_requests"`
}

type DailyMetric struct {
	Date            time.Time `json:"date"`
	ViewsCount      int       `json:"views_count" db:"views_count"`
	NewUsersCount   int       `json:"new_users_count" db:"new_users_count"`
	NewRatingsCount int       `json:"new_ratings_count" db:"new_ratings_count"`
	NewSongsCount   int       `json:"new_songs_count" db:"new_songs_count"`
}

type AdminRepository interface {
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	GetDailyMetrics(ctx context.Context, days int) ([]DailyMetric, error)

	// XP Activities
	GetAllXPActivities(ctx context.Context) ([]XPActivity, error)
	UpdateXPActivity(ctx context.Context, activity *XPActivity) error
}

// AnilistBatchImportItemError is one failed ID from a batch AniList import.
type AnilistBatchImportItemError struct {
	AnilistID int    `json:"anilist_id"`
	Message   string `json:"message"`
}

// AnilistBatchImportResult is returned by POST /admin/animes/batch-from-anilist.
type AnilistBatchImportResult struct {
	Requested   int                           `json:"requested"`
	Imported    int                           `json:"imported"`
	Failed      int                           `json:"failed"`
	ImportedIDs []int                         `json:"imported_ids"`
	Errors      []AnilistBatchImportItemError `json:"errors"`
}

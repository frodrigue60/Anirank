package domain

import "context"

type StatPoint struct {
	Date  string `json:"date" db:"date"`
	Count int    `json:"count" db:"count"`
}

type StaffDashboardStats struct {
	TotalUsers    int `json:"total_users" db:"total_users"`
	TotalAnimes   int `json:"total_animes" db:"total_animes"`
	TotalSongs    int `json:"total_songs" db:"total_songs"`
	TotalRatings  int `json:"total_ratings" db:"total_ratings"`
	TotalComments int `json:"total_comments" db:"total_comments"`
}

type UserDistribution struct {
	Label string `json:"label" db:"label"`
	Value int    `json:"value" db:"value"`
}

type SiteStats struct {
	Overviews         StaffDashboardStats `json:"overviews"`
	UserGrowth        []StatPoint         `json:"user_growth"`
	RatingGrowth      []StatPoint         `json:"rating_growth"`
	SongGrowth        []StatPoint         `json:"song_growth"`
	LevelDistribution []UserDistribution  `json:"level_distribution"`
	ScoreDistribution []UserDistribution  `json:"score_distribution"`
}

type StatsUsecase interface {
	GetSiteStats(ctx context.Context) (*SiteStats, error)
}

type StatsRepository interface {
	GetTotals(ctx context.Context) (*StaffDashboardStats, error)
	GetUserGrowth(ctx context.Context, days int) ([]StatPoint, error)
	GetRatingGrowth(ctx context.Context, days int) ([]StatPoint, error)
	GetSongGrowth(ctx context.Context, days int) ([]StatPoint, error)
	GetLevelDistribution(ctx context.Context) ([]UserDistribution, error)
	GetScoreDistribution(ctx context.Context) ([]UserDistribution, error)
}

package dto

import "time"

type UserScoreBucketDTO struct {
	Label      string  `json:"label"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type UserRecentRatingDTO struct {
	Rating  float64     `json:"rating"`
	RatedAt time.Time   `json:"rated_at"`
	Song    SongSlimDTO `json:"song"`
}

type UserRatingInsightsDTO struct {
	AverageScore  *float64              `json:"average_score"`
	TotalRatings  int                   `json:"total_ratings"`
	Distribution  []UserScoreBucketDTO  `json:"score_distribution"`
	RecentRatings []UserRecentRatingDTO `json:"recent_ratings"`
}

package public

import (
	"context"
	"anirank/api/internal/domain"
	"time"
)

type statsUsecase struct {
	statsRepo domain.StatsRepository
	cache     domain.Cache
}

func NewStatsUsecase(statsRepo domain.StatsRepository, appCache domain.Cache) domain.StatsUsecase {
	return &statsUsecase{
		statsRepo: statsRepo,
		cache:     appCache,
	}
}

func (u *statsUsecase) GetSiteStats(ctx context.Context) (*domain.SiteStats, error) {
	cacheKey := "stats:site"
	var stats domain.SiteStats

	// Try Cache
	if err := u.cache.Get(ctx, cacheKey, &stats); err == nil {
		return &stats, nil
	}

	totals, err := u.statsRepo.GetTotals(ctx)
	if err != nil {
		return nil, err
	}

	userGrowth, err := u.statsRepo.GetUserGrowth(ctx, 30)
	if err != nil {
		return nil, err
	}

	ratingGrowth, err := u.statsRepo.GetRatingGrowth(ctx, 30)
	if err != nil {
		return nil, err
	}

	songGrowth, err := u.statsRepo.GetSongGrowth(ctx, 30)
	if err != nil {
		return nil, err
	}

	levelDist, err := u.statsRepo.GetLevelDistribution(ctx)
	if err != nil {
		return nil, err
	}

	scoreDist, err := u.statsRepo.GetScoreDistribution(ctx)
	if err != nil {
		return nil, err
	}

	// Fill missing days with 0 to ensure frontend gets a continuous line
	userGrowth = u.fillMissingDates(userGrowth, 30)
	ratingGrowth = u.fillMissingDates(ratingGrowth, 30)
	songGrowth = u.fillMissingDates(songGrowth, 30)

	// Ensure all 10 buckets for levels and scores are present
	levelDist = u.fillMissingBuckets(levelDist)
	scoreDist = u.fillMissingBuckets(scoreDist)

	res := &domain.SiteStats{
		Overviews:         *totals,
		UserGrowth:        userGrowth,
		RatingGrowth:      ratingGrowth,
		SongGrowth:        songGrowth,
		LevelDistribution: levelDist,
		ScoreDistribution: scoreDist,
	}

	// Store in cache for 1 hour
	_ = u.cache.Set(ctx, cacheKey, res, 1*time.Hour)

	return res, nil
}

func (u *statsUsecase) fillMissingDates(points []domain.StatPoint, days int) []domain.StatPoint {
	m := make(map[string]int)
	for _, p := range points {
		// handle date format if needed, assuming YYYY-MM-DD
		m[p.Date] = p.Count
	}

	result := make([]domain.StatPoint, 0, days)
	now := time.Now()
	for i := days - 1; i >= 0; i-- {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		count := 0
		if v, ok := m[d]; ok {
			count = v
		}
		result = append(result, domain.StatPoint{
			Date:  d,
			Count: count,
		})
	}
	return result
}

func (u *statsUsecase) fillMissingBuckets(points []domain.UserDistribution) []domain.UserDistribution {
	m := make(map[string]int)
	for _, p := range points {
		m[p.Label] = p.Value
	}

	buckets := []string{"10", "20", "30", "40", "50", "60", "70", "80", "90", "100"}
	result := make([]domain.UserDistribution, 0, len(buckets))
	for _, b := range buckets {
		val := 0
		if v, ok := m[b]; ok {
			val = v
		}
		result = append(result, domain.UserDistribution{
			Label: b,
			Value: val,
		})
	}
	return result
}

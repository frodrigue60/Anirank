package postgres

import (
	"context"
	"anirank/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type statsRepository struct {
	db *sqlx.DB
}

func NewStatsRepository(db *sqlx.DB) domain.StatsRepository {
	return &statsRepository{db: db}
}

func (r *statsRepository) GetTotals(ctx context.Context) (*domain.StaffDashboardStats, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM users) as total_users,
			(SELECT COUNT(*) FROM animes WHERE status = true) as total_animes,
			(SELECT COUNT(*) FROM songs WHERE status = true) as total_songs,
			(SELECT COUNT(*) FROM artists WHERE status = true) as total_artists,
			(SELECT COUNT(*) FROM song_ratings) as total_ratings,
			(SELECT COUNT(*) FROM comments) as total_comments
	`
	var stats domain.StaffDashboardStats
	err := r.db.GetContext(ctx, &stats, query)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *statsRepository) GetUserGrowth(ctx context.Context, days int) ([]domain.StatPoint, error) {
	query := `
		SELECT TO_CHAR(date, 'YYYY-MM-DD') as date, new_users_count as count 
		FROM daily_metrics 
		WHERE song_id IS NULL AND date >= CURRENT_DATE - (INTERVAL '1 day' * $1)
		ORDER BY date ASC
	`
	
	var points []domain.StatPoint
	err := r.db.SelectContext(ctx, &points, query, days)
	return points, err
}

func (r *statsRepository) GetRatingGrowth(ctx context.Context, days int) ([]domain.StatPoint, error) {
	query := `
		SELECT TO_CHAR(date, 'YYYY-MM-DD') as date, new_ratings_count as count 
		FROM daily_metrics 
		WHERE song_id IS NULL AND date >= CURRENT_DATE - (INTERVAL '1 day' * $1)
		ORDER BY date ASC
	`
	
	var points []domain.StatPoint
	err := r.db.SelectContext(ctx, &points, query, days)
	return points, err
}

func (r *statsRepository) GetSongGrowth(ctx context.Context, days int) ([]domain.StatPoint, error) {
	query := `
		SELECT TO_CHAR(date, 'YYYY-MM-DD') as date, new_songs_count as count 
		FROM daily_metrics 
		WHERE song_id IS NULL AND date >= CURRENT_DATE - (INTERVAL '1 day' * $1)
		ORDER BY date ASC
	`
	
	var points []domain.StatPoint
	err := r.db.SelectContext(ctx, &points, query, days)
	return points, err
}

func (r *statsRepository) GetLevelDistribution(ctx context.Context) ([]domain.UserDistribution, error) {
	query := `
		SELECT 
			CASE 
				WHEN level <= 10 THEN '10'
				WHEN level <= 20 THEN '20'
				WHEN level <= 30 THEN '30'
				WHEN level <= 40 THEN '40'
				WHEN level <= 50 THEN '50'
				WHEN level <= 60 THEN '60'
				WHEN level <= 70 THEN '70'
				WHEN level <= 80 THEN '80'
				WHEN level <= 90 THEN '90'
				ELSE '100'
			END as label,
			COUNT(*) as value
		FROM users
		GROUP BY 1
		ORDER BY 1::integer ASC
	`
	var dist []domain.UserDistribution
	err := r.db.SelectContext(ctx, &dist, query)
	return dist, err
}

func (r *statsRepository) GetScoreDistribution(ctx context.Context) ([]domain.UserDistribution, error) {
	query := `
		SELECT 
			CASE 
				WHEN rating <= 10 THEN '10'
				WHEN rating <= 20 THEN '20'
				WHEN rating <= 30 THEN '30'
				WHEN rating <= 40 THEN '40'
				WHEN rating <= 50 THEN '50'
				WHEN rating <= 60 THEN '60'
				WHEN rating <= 70 THEN '70'
				WHEN rating <= 80 THEN '80'
				WHEN rating <= 90 THEN '90'
				ELSE '100'
			END as label,
			COUNT(*) as value
		FROM song_ratings
		GROUP BY 1
		ORDER BY 1::integer ASC
	`
	var dist []domain.UserDistribution
	err := r.db.SelectContext(ctx, &dist, query)
	return dist, err
}

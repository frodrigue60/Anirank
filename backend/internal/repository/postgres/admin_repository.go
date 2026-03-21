package postgres

import (
	"context"
	"time"

	"anirank/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type adminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetDashboardStats(ctx context.Context) (*domain.DashboardStats, error) {
	stats := &domain.DashboardStats{}

	// Total Counts
	_ = r.db.GetContext(ctx, &stats.TotalAnimes, "SELECT COUNT(*) FROM animes")
	_ = r.db.GetContext(ctx, &stats.TotalSongs, "SELECT COUNT(*) FROM songs")
	_ = r.db.GetContext(ctx, &stats.TotalUsers, "SELECT COUNT(*) FROM users")
	_ = r.db.GetContext(ctx, &stats.TotalArtists, "SELECT COUNT(*) FROM artists")
	_ = r.db.GetContext(ctx, &stats.TotalPlaylists, "SELECT COUNT(*) FROM playlists")
	_ = r.db.GetContext(ctx, &stats.TotalTournaments, "SELECT COUNT(*) FROM tournaments")

	// Pending Approval Detailed
	_ = r.db.GetContext(ctx, &stats.PendingAnimes, "SELECT COUNT(*) FROM animes WHERE status = false")
	_ = r.db.GetContext(ctx, &stats.PendingSongs, "SELECT COUNT(*) FROM songs WHERE status = false")
	_ = r.db.GetContext(ctx, &stats.PendingVariants, "SELECT COUNT(*) FROM song_variants WHERE status = false")
	_ = r.db.GetContext(ctx, &stats.PendingVideos, "SELECT COUNT(*) FROM videos WHERE status = false")
	_ = r.db.GetContext(ctx, &stats.PendingArtists, "SELECT COUNT(*) FROM artists WHERE status = false")

	// Reports Detailed
	_ = r.db.GetContext(ctx, &stats.SongReports, "SELECT COUNT(*) FROM song_reports WHERE status = false")
	_ = r.db.GetContext(ctx, &stats.CommentReports, "SELECT COUNT(*) FROM comment_reports WHERE status = false")
	stats.PendingReports = stats.SongReports + stats.CommentReports

	// Pending Requests
	_ = r.db.GetContext(ctx, &stats.PendingRequests, "SELECT COUNT(*) FROM user_requests WHERE status = false")

	// Active Metrics
	yesterday := time.Now().Add(-24 * time.Hour)
	_ = r.db.GetContext(ctx, &stats.ActiveUsersDay, "SELECT COUNT(*) FROM users WHERE last_login_at >= $1", yesterday)
	_ = r.db.GetContext(ctx, &stats.ActiveTournaments, "SELECT COUNT(*) FROM tournaments WHERE status = 'active'")

	return stats, nil
}

func (r *adminRepository) GetDailyMetrics(ctx context.Context, days int) ([]domain.DailyMetric, error) {
	// PostgreSQL syntax for date interval subtraction
	query := `
		SELECT date, SUM(views_count) as views_count 
		FROM daily_metrics 
		WHERE date >= CURRENT_DATE - (INTERVAL '1 day' * $1)
		GROUP BY date 
		ORDER BY date ASC
	`
	rows, err := r.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []domain.DailyMetric
	for rows.Next() {
		var m domain.DailyMetric
		var dateVal any 
		if err := rows.Scan(&dateVal, &m.ViewsCount); err != nil {
			return nil, err
		}
		
		switch v := dateVal.(type) {
		case time.Time:
			m.Date = v
		case []byte:
			parsedDate, _ := time.Parse("2006-01-02", string(v))
			m.Date = parsedDate
		case string:
			parsedDate, _ := time.Parse("2006-01-02", v)
			m.Date = parsedDate
		}
		
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (r *adminRepository) GetAllXPActivities(ctx context.Context) ([]domain.XPActivity, error) {
	var activities []domain.XPActivity
	query := "SELECT * FROM xp_activities ORDER BY id ASC"
	err := r.db.SelectContext(ctx, &activities, query)
	return activities, err
}

func (r *adminRepository) UpdateXPActivity(ctx context.Context, activity *domain.XPActivity) error {
	query := `
		UPDATE xp_activities 
		SET xp_amount = $1, description = $2, cooldown_seconds = $3, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, activity.XPAmount, activity.Description, activity.CooldownSeconds, activity.ID)
	return err
}

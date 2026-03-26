package postgres

import (
	"anirank/api/internal/domain"
	"context"

	"github.com/jmoiron/sqlx"
)

type activityRepository struct {
	db *sqlx.DB
}

func NewActivityRepository(db *sqlx.DB) domain.ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) GetPaginated(ctx context.Context, limit, offset int) ([]domain.Activity, error) {
	var activities []domain.Activity
	query := `SELECT * FROM activities ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &activities, query, limit, offset)
	if err != nil {
		return nil, err
	}
	if activities == nil {
		activities = []domain.Activity{}
	}
	return activities, nil
}

func (r *activityRepository) Create(ctx context.Context, a *domain.Activity) error {
	query := `INSERT INTO activities (user_id, action_type, target_id, target_type, action_value, created_at, updated_at) 
			  VALUES (:user_id, :action_type, :target_id, :target_type, :action_value, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			  RETURNING id`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, a).Scan(&a.ID)
	return err
}

func (r *activityRepository) DeleteByTarget(ctx context.Context, userID uint64, actionType string, targetID uint64, targetType string) error {
	query := `DELETE FROM activities WHERE user_id = $1 AND action_type = $2 AND target_id = $3 AND target_type = $4`
	_, err := r.db.ExecContext(ctx, query, userID, actionType, targetID, targetType)
	return err
}

func (r *activityRepository) Exists(ctx context.Context, userID uint64, actionType string, targetID uint64, targetType string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM activities WHERE user_id = $1 AND action_type = $2 AND target_id = $3 AND target_type = $4)`
	err := r.db.GetContext(ctx, &exists, query, userID, actionType, targetID, targetType)
	return exists, err
}
func (r *activityRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM activities`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type xpRepository struct {
	db *sqlx.DB
}

func NewXPRepository(db *sqlx.DB) domain.XPRepository {
	return &xpRepository{db: db}
}

func (r *xpRepository) GetActivityByKey(ctx context.Context, key string) (*domain.XPActivity, error) {
	var activity domain.XPActivity
	// Use double quotes for reserved word "key" in PostgreSQL
	query := `SELECT * FROM xp_activities WHERE "key" = $1`
	err := r.db.GetContext(ctx, &activity, query, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &activity, err
}

func (r *xpRepository) GetLastLogByActivity(ctx context.Context, userID, activityID uint64) (*domain.XPLog, error) {
	var log domain.XPLog
	query := `SELECT * FROM xp_logs WHERE user_id = $1 AND xp_activity_id = $2 ORDER BY created_at DESC LIMIT 1`
	err := r.db.GetContext(ctx, &log, query, userID, activityID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &log, err
}

func (r *xpRepository) GetLogByActivityAndMetadata(ctx context.Context, userID, activityID uint64, metadataKey string, metadataValue interface{}) (*domain.XPLog, error) {
	var log domain.XPLog
	// PostgreSQL JSONB extraction: metadata->>'key'
	query := fmt.Sprintf(`SELECT * FROM xp_logs WHERE user_id = $1 AND xp_activity_id = $2 AND metadata->>'%s' = $3 LIMIT 1`, metadataKey)
	
	// Convert value to string for JSONB extraction (which returns text)
	val := fmt.Sprintf("%v", metadataValue)
	if s, ok := metadataValue.(string); ok {
		val = s
	} else {
		// If it's complex, marshal it. But usually it's an ID or simple string.
		b, _ := json.Marshal(metadataValue)
		val = string(b)
		// Strip quotes if marshaled as string
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
	}
	
	err := r.db.GetContext(ctx, &log, query, userID, activityID, val)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &log, err
}

func (r *xpRepository) CreateLog(ctx context.Context, log *domain.XPLog) error {
	query := `INSERT INTO xp_logs (user_id, xp_activity_id, xp_amount, metadata, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			  RETURNING id`
	err := r.db.QueryRowContext(ctx, query, log.UserID, log.XPActivityID, log.XPAmount, log.Metadata).Scan(&log.ID)
	return err
}

func (r *xpRepository) GetCurrentLevel(ctx context.Context, xp uint64) (uint32, error) {
	var level uint32
	query := `SELECT level FROM levels WHERE $1 >= min_xp ORDER BY min_xp DESC LIMIT 1`
	err := r.db.GetContext(ctx, &level, query, xp)
	if errors.Is(err, sql.ErrNoRows) {
		return 1, nil // Default to level 1
	}
	return level, err
}

func (r *xpRepository) UpdateUserXPAndLevel(ctx context.Context, userID uint64, xpAmount int, newLevel uint32) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE users SET xp = xp + $1, level = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err = tx.ExecContext(ctx, query, xpAmount, newLevel, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type notificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}

	query := `
		INSERT INTO notifications (id, user_id, type, subject_id, subject_type, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := r.db.ExecContext(ctx, query, n.ID, n.UserID, n.Type, n.SubjectID, n.SubjectType, string(n.Data))
	return err
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uint64, notificationType string, limit, offset int) ([]domain.Notification, int, error) {
	var notifications []domain.Notification
	var total int

	countQuery := "SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND deleted_at IS NULL"
	countArgs := []interface{}{userID}
	i := 2

	if notificationType != "" {
		countQuery += fmt.Sprintf(" AND type = $%d", i)
		countArgs = append(countArgs, notificationType)
		i++
	}

	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, type, subject_id, subject_type, data, read_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{userID}
	j := 2

	if notificationType != "" {
		query += fmt.Sprintf(" AND type = $%d", j)
		args = append(args, notificationType)
		j++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", j, j+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &notifications, query, args...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, 0, err
	}

	if notifications == nil {
		notifications = []domain.Notification{}
	}

	return notifications, total, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id string, userID uint64) error {
	query := "UPDATE notifications SET read_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2 AND read_at IS NULL"
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uint64) error {
	query := "UPDATE notifications SET read_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND read_at IS NULL"
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *notificationRepository) Delete(ctx context.Context, id string, userID uint64) error {
	query := "UPDATE notifications SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2"
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uint64) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL AND deleted_at IS NULL"
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

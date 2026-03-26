package domain

import (
	"context"
	"encoding/json"
	"time"
)

type Notification struct {
	ID          string          `db:"id" json:"id"`
	UserID      uint64          `db:"user_id" json:"user_id"`
	Type        string          `db:"type" json:"type"`
	SubjectID   *uint64         `db:"subject_id" json:"subject_id,omitempty"`
	SubjectUUID *string         `db:"subject_uuid" json:"subject_uuid,omitempty"`
	SubjectType *string         `db:"subject_type" json:"subject_type,omitempty"`
	Data        json.RawMessage `db:"data" json:"data"`
	ReadAt      *time.Time      `db:"read_at" json:"read_at,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *Notification) error
	GetByUserID(ctx context.Context, userID uint64, notificationType string, limit, offset int) ([]Notification, int, error)
	MarkAsRead(ctx context.Context, id string, userID uint64) error
	MarkAllAsRead(ctx context.Context, userID uint64) error
	Delete(ctx context.Context, id string, userID uint64) error
	GetUnreadCount(ctx context.Context, userID uint64) (int, error)
}

type NotificationUsecase interface {
	GetNotifications(ctx context.Context, userID uint64, notificationType string, limit, offset int) ([]Notification, int, int, error)
	MarkAsRead(ctx context.Context, id string, userID uint64) error
	MarkAllAsRead(ctx context.Context, userID uint64) error
	Delete(ctx context.Context, id string, userID uint64) error
	GetUnreadCount(ctx context.Context, userID uint64) (int, error)
}

package domain

import (
	"context"
	"time"
)

type XPActivity struct {
	ID              uint64    `db:"id" json:"id"`
	Key             string    `db:"key" json:"key"`
	XPAmount        int       `db:"xp_amount" json:"xp_amount"`
	Description     *string   `db:"description" json:"description"`
	CooldownSeconds int       `db:"cooldown_seconds" json:"cooldown_seconds"`
	CreatedAt       *time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt       *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type XPLog struct {
	ID           uint64    `db:"id" json:"id"`
	UserID       uint64    `db:"user_id" json:"user_id"`
	XPActivityID uint64    `db:"xp_activity_id" json:"xp_activity_id"`
	XPAmount     int       `db:"xp_amount" json:"xp_amount"`
	Metadata     []byte    `db:"metadata" json:"metadata,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Level struct {
	Level     uint32    `db:"level" json:"level"`
	MinXP     uint64    `db:"min_xp" json:"min_xp"`
	Name      *string   `db:"name" json:"name"`
	BadgeID   *uint64   `db:"badge_id" json:"badge_id,omitempty"`
	CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type XPRepository interface {
	GetActivityByKey(ctx context.Context, key string) (*XPActivity, error)
	GetLastLogByActivity(ctx context.Context, userID, activityID uint64) (*XPLog, error)
	GetLogByActivityAndMetadata(ctx context.Context, userID, activityID uint64, metadataKey string, metadataValue interface{}) (*XPLog, error)
	CreateLog(ctx context.Context, log *XPLog) error
	GetCurrentLevel(ctx context.Context, xp uint64) (uint32, error)
	UpdateUserXPAndLevel(ctx context.Context, userID uint64, xpAmount int, newLevel uint32) error
}

type XPUsecase interface {
	AwardXP(ctx context.Context, userID uint64, activityKey string, metadata map[string]interface{}) error
	CheckDailyLogin(ctx context.Context, userID uint64) error
}

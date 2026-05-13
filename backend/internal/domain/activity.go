package domain

import (
	"context"
	"time"
)

type Activity struct {
	ID          uint64    `db:"id" json:"id"`
	UserID      uint64    `db:"user_id" json:"user_id"`
	ActionType  string    `db:"action_type" json:"action_type"`
	TargetID    uint64    `db:"target_id" json:"target_id"`
	TargetType  string    `db:"target_type" json:"target_type"`
	ActionValue *string   `db:"action_value" json:"action_value"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	// Compatibility aliases for frontend
	Action string `db:"-" json:"action"`

	// Relations
	User   *User       `db:"-" json:"user,omitempty"`
	Target interface{} `db:"-" json:"target,omitempty"`

	// Virtual fields for frontend (populated if type matches)
	Song       *Song   `db:"-" json:"song,omitempty"`
	Artist     *Artist `db:"-" json:"artist,omitempty"`
	UserTarget *User   `db:"-" json:"user_target,omitempty"`
	Badge      *Badge  `db:"-" json:"badge,omitempty"`
}

type ActivityRepository interface {
	GetPaginated(ctx context.Context, limit, offset int) ([]Activity, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, activity *Activity) error
	DeleteByTarget(ctx context.Context, userID uint64, actionType string, targetID uint64, targetType string) error
	Exists(ctx context.Context, userID uint64, actionType string, targetID uint64, targetType string) (bool, error)
}

type ActivityUsecase interface {
	GetFeed(ctx context.Context, limit, offset int) ([]Activity, error)
	GetCount(ctx context.Context) (int, error)
	LogActivity(ctx context.Context, userID uint64, actionType string, targetID uint64, targetType string, value *string) error
}

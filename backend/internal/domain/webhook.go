package domain

import (
	"context"
	"time"
)

type Webhook struct {
	ID           uint64    `db:"id" json:"-"`
	UUID         string    `db:"uuid" json:"uuid"`
	Name         string    `db:"name" json:"name"`
	URL          string    `db:"url" json:"url"`
	Provider     string    `db:"provider" json:"provider"` // e.g., "discord"
	IsActive     bool      `db:"is_active" json:"is_active"`
	ContentTypes []string  `db:"content_types" json:"content_types"` // e.g., ["anime", "song"]
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type WebhookRepository interface {
	GetByID(ctx context.Context, id uint64) (*Webhook, error)
	GetByUUID(ctx context.Context, uuid string) (*Webhook, error)
	GetAll(ctx context.Context) ([]Webhook, error)
	GetByContentType(ctx context.Context, contentType string) ([]Webhook, error)
	Create(ctx context.Context, webhook *Webhook) error
	Update(ctx context.Context, webhook *Webhook) error
	Delete(ctx context.Context, id uint64) error
}

type WebhookUsecase interface {
	GetByUUID(ctx context.Context, uuid string) (*Webhook, error)
	GetAll(ctx context.Context) ([]Webhook, error)
	GetForContent(ctx context.Context, contentType string) ([]Webhook, error)
	Create(ctx context.Context, webhook *Webhook) error
	Update(ctx context.Context, webhook *Webhook) error
	Delete(ctx context.Context, uuid string) error
	
	// Triggering
	TriggerForAnime(ctx context.Context, webhookUUID string, animeID uint64) error
	TriggerForSong(ctx context.Context, webhookUUID string, songID uint64) error
	TestWebhook(ctx context.Context, uuid string) error

	// Notify all relevant webhooks
	NotifyNewAnime(ctx context.Context, animeID uint64) error
	NotifyNewSong(ctx context.Context, songID uint64) error
}

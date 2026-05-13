package postgres

import (
	"context"
	"database/sql"
	"time"

	"anirank/api/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type webhookRepository struct {
	db *sqlx.DB
}

func NewWebhookRepository(db *sqlx.DB) domain.WebhookRepository {
	return &webhookRepository{db: db}
}

type webhookModel struct {
	ID           uint64         `db:"id"`
	UUID         string         `db:"uuid"`
	Name         string         `db:"name"`
	URL          string         `db:"url"`
	Provider     string         `db:"provider"`
	IsActive     bool           `db:"is_active"`
	ContentTypes pq.StringArray `db:"content_types"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

func (m *webhookModel) toDomain() *domain.Webhook {
	return &domain.Webhook{
		ID:           m.ID,
		UUID:         m.UUID,
		Name:         m.Name,
		URL:          m.URL,
		Provider:     m.Provider,
		IsActive:     m.IsActive,
		ContentTypes: []string(m.ContentTypes),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (r *webhookRepository) GetByID(ctx context.Context, id uint64) (*domain.Webhook, error) {
	var model webhookModel
	err := r.db.GetContext(ctx, &model, "SELECT * FROM webhooks WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *webhookRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Webhook, error) {
	var model webhookModel
	err := r.db.GetContext(ctx, &model, "SELECT * FROM webhooks WHERE uuid = $1", uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *webhookRepository) GetAll(ctx context.Context) ([]domain.Webhook, error) {
	var models []webhookModel
	err := r.db.SelectContext(ctx, &models, "SELECT * FROM webhooks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	
	webhooks := make([]domain.Webhook, len(models))
	for i, m := range models {
		webhooks[i] = *m.toDomain()
	}
	return webhooks, nil
}

func (r *webhookRepository) GetByContentType(ctx context.Context, contentType string) ([]domain.Webhook, error) {
	var models []webhookModel
	err := r.db.SelectContext(ctx, &models, "SELECT * FROM webhooks WHERE is_active = true AND $1 = ANY(content_types)", contentType)
	if err != nil {
		return nil, err
	}
	
	webhooks := make([]domain.Webhook, len(models))
	for i, m := range models {
		webhooks[i] = *m.toDomain()
	}
	return webhooks, nil
}

func (r *webhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	query := `
		INSERT INTO webhooks (name, url, provider, is_active, content_types)
		VALUES (:name, :url, :provider, :is_active, :content_types)
		RETURNING id, uuid, created_at, updated_at
	`
	
	arg := map[string]interface{}{
		"name":          webhook.Name,
		"url":           webhook.URL,
		"provider":      webhook.Provider,
		"is_active":     webhook.IsActive,
		"content_types": pq.Array(webhook.ContentTypes),
	}

	rows, err := r.db.NamedQueryContext(ctx, query, arg)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&webhook.ID, &webhook.UUID, &webhook.CreatedAt, &webhook.UpdatedAt); err != nil {
			return err
		}
	}

	return nil
}

func (r *webhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	query := `
		UPDATE webhooks
		SET name = :name, url = :url, provider = :provider, is_active = :is_active, content_types = :content_types, updated_at = CURRENT_TIMESTAMP
		WHERE uuid = :uuid
	`
	
	arg := map[string]interface{}{
		"uuid":          webhook.UUID,
		"name":          webhook.Name,
		"url":           webhook.URL,
		"provider":      webhook.Provider,
		"is_active":     webhook.IsActive,
		"content_types": pq.Array(webhook.ContentTypes),
	}

	result, err := r.db.NamedExecContext(ctx, query, arg)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *webhookRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM webhooks WHERE id = $1", id)
	return err
}

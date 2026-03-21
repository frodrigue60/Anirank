package postgres

import (
	"context"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type badgeRepository struct {
	db *sqlx.DB
}

func NewBadgeRepository(db *sqlx.DB) domain.BadgeRepository {
	return &badgeRepository{db: db}
}

func (r *badgeRepository) GetByID(ctx context.Context, id uint64) (*domain.Badge, error) {
	var badge domain.Badge
	query := "SELECT * FROM badges WHERE id = $1"
	err := r.db.GetContext(ctx, &badge, query, id)
	if err != nil {
		return nil, err
	}
	return &badge, nil
}

func (r *badgeRepository) GetAll(ctx context.Context) ([]domain.Badge, error) {
	var badges []domain.Badge
	query := "SELECT * FROM badges ORDER BY name ASC"
	err := r.db.SelectContext(ctx, &badges, query)
	if err != nil {
		return nil, err
	}
	if badges == nil {
		badges = []domain.Badge{}
	}
	return badges, nil
}

func (r *badgeRepository) Create(ctx context.Context, badge *domain.Badge) error {
	query := `
		INSERT INTO badges (name, description, icon, is_active, created_at, updated_at)
		VALUES (:name, :description, :icon, :is_active, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, badge).Scan(&badge.ID)
	return err
}

func (r *badgeRepository) Update(ctx context.Context, badge *domain.Badge) error {
	query := `
		UPDATE badges 
		SET name = :name, 
		    description = :description, 
		    is_active = :is_active,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, badge)
	return err
}

func (r *badgeRepository) UpdateIcon(ctx context.Context, id uint64, iconPath string) error {
	query := "UPDATE badges SET icon = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, iconPath, id)
	return err
}

func (r *badgeRepository) Delete(ctx context.Context, id uint64) error {
	query := "DELETE FROM badges WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

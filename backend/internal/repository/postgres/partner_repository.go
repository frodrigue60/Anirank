package postgres

import (
	"context"
	"database/sql"
	"errors"

	"anirank/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type partnerRepository struct {
	db *sqlx.DB
}

func NewPartnerRepository(db *sqlx.DB) domain.PartnerRepository {
	return &partnerRepository{db: db}
}

func (r *partnerRepository) GetByID(ctx context.Context, id uint64) (*domain.Partner, error) {
	var p domain.Partner
	query := `SELECT * FROM partners WHERE id = $1`
	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *partnerRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Partner, error) {
	var p domain.Partner
	query := `SELECT * FROM partners WHERE uuid = $1`
	err := r.db.GetContext(ctx, &p, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *partnerRepository) GetAll(ctx context.Context, onlyActive bool) ([]domain.Partner, error) {
	var partners []domain.Partner
	query := `SELECT * FROM partners`
	if onlyActive {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY sort_order ASC, created_at DESC`

	err := r.db.SelectContext(ctx, &partners, query)
	return partners, err
}

func (r *partnerRepository) Create(ctx context.Context, p *domain.Partner) error {
	query := `
		INSERT INTO partners (name, url, banner, description, type, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, uuid, created_at, updated_at`
	
	return r.db.QueryRowContext(ctx, query,
		p.Name, p.URL, p.Banner, p.Description, p.Type, p.SortOrder, p.IsActive,
	).Scan(&p.ID, &p.UUID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *partnerRepository) Update(ctx context.Context, p *domain.Partner) error {
	query := `
		UPDATE partners
		SET name = $1, url = $2, banner = $3, description = $4, type = $5, sort_order = $6, is_active = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8`
	
	_, err := r.db.ExecContext(ctx, query,
		p.Name, p.URL, p.Banner, p.Description, p.Type, p.SortOrder, p.IsActive, p.ID,
	)
	return err
}

func (r *partnerRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM partners WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

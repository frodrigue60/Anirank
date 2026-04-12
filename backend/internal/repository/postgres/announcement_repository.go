package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"anirank/api/internal/domain"
	"github.com/jmoiron/sqlx"
)

type postgresAnnouncementRepository struct {
	db *sqlx.DB
}

func NewAnnouncementRepository(db *sqlx.DB) domain.AnnouncementRepository {
	return &postgresAnnouncementRepository{db: db}
}

func (r *postgresAnnouncementRepository) GetByID(ctx context.Context, id uint64) (*domain.Announcement, error) {
	var a domain.Announcement
	query := `SELECT * FROM announcements WHERE id = $1`
	if err := r.db.GetContext(ctx, &a, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(404, "Announcement not found", err)
		}
		return nil, err
	}
	return &a, nil
}

func (r *postgresAnnouncementRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Announcement, error) {
	var a domain.Announcement
	query := `SELECT * FROM announcements WHERE uuid = $1`
	if err := r.db.GetContext(ctx, &a, query, uuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewAppError(404, "Announcement not found", err)
		}
		return nil, err
	}
	return &a, nil
}

func (r *postgresAnnouncementRepository) GetAll(ctx context.Context, filters domain.AnnouncementFilters, limit, offset int) ([]domain.Announcement, error) {
	var announcements []domain.Announcement
	query := `SELECT * FROM announcements WHERE 1=1`
	args := []interface{}{}
	i := 1

	if filters.ActiveOnly {
		now := time.Now()
		query += fmt.Sprintf(` AND is_active = true AND (starts_at IS NULL OR starts_at <= $%d) AND (ends_at IS NULL OR ends_at >= $%d)`, i, i+1)
		args = append(args, now, now)
		i += 2
	}

	if filters.Search != "" {
		query += fmt.Sprintf(` AND (title ILIKE $%d OR content ILIKE $%d)`, i, i+1)
		args = append(args, "%"+filters.Search+"%", "%"+filters.Search+"%")
		i += 2
	}

	query += ` ORDER BY priority DESC, created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, i, i+1)
		args = append(args, limit, offset)
		i += 2
	}

	if err := r.db.SelectContext(ctx, &announcements, query, args...); err != nil {
		return nil, err
	}
	if announcements == nil {
		announcements = []domain.Announcement{}
	}
	return announcements, nil
}

func (r *postgresAnnouncementRepository) Count(ctx context.Context, filters domain.AnnouncementFilters) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM announcements WHERE 1=1`
	args := []interface{}{}
	i := 1

	if filters.ActiveOnly {
		now := time.Now()
		query += fmt.Sprintf(` AND is_active = true AND (starts_at IS NULL OR starts_at <= $%d) AND (ends_at IS NULL OR ends_at >= $%d)`, i, i+1)
		args = append(args, now, now)
		i += 2
	}

	if filters.Search != "" {
		query += fmt.Sprintf(` AND (title ILIKE $%d OR content ILIKE $%d)`, i, i+1)
		args = append(args, "%"+filters.Search+"%", "%"+filters.Search+"%")
		i += 2
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

func (r *postgresAnnouncementRepository) GetActive(ctx context.Context) ([]domain.Announcement, error) {
	return r.GetAll(ctx, domain.AnnouncementFilters{ActiveOnly: true}, 0, 0)
}

func (r *postgresAnnouncementRepository) Create(ctx context.Context, a *domain.Announcement) error {
	query := `INSERT INTO announcements (
		uuid, title, content, type, icon, url, image, priority, is_active, starts_at, ends_at, created_at, updated_at
	) VALUES (
		:uuid, :title, :content, :type, :icon, :url, :image, :priority, :is_active, :starts_at, :ends_at, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
	) RETURNING id`
	
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, a).Scan(&a.ID)
	return err
}

func (r *postgresAnnouncementRepository) Update(ctx context.Context, a *domain.Announcement) error {
	query := `UPDATE announcements SET 
		uuid = :uuid,
		title = :title, 
		content = :content, 
		type = :type, 
		icon = :icon, 
		url = :url, 
		image = :image, 
		priority = :priority, 
		is_active = :is_active, 
		starts_at = :starts_at, 
		ends_at = :ends_at, 
		updated_at = CURRENT_TIMESTAMP
	WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, a)
	return err
}

func (r *postgresAnnouncementRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM announcements WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *postgresAnnouncementRepository) ToggleActive(ctx context.Context, id uint64) error {
	query := `UPDATE announcements SET is_active = NOT is_active, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

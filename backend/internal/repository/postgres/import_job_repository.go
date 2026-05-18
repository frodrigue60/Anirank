package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

// importJobRepository is the PostgreSQL implementation of domain.ImportJobRepository.
type importJobRepository struct {
	db *sqlx.DB
}

// NewImportJobRepository creates a new ImportJobRepository.
func NewImportJobRepository(db *sqlx.DB) domain.ImportJobRepository {
	return &importJobRepository{db: db}
}

// Create inserts a new ImportJob record.
func (r *importJobRepository) Create(ctx context.Context, job *domain.ImportJob) error {
	errorsJSON, err := marshalErrors(job.Errors)
	if err != nil {
		return fmt.Errorf("import_job: marshal errors: %w", err)
	}

	now := time.Now().UTC()
	job.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO import_jobs (id, source, status, current_page, total_pages, processed, created, skipped, errors, started_at, finished_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		job.ID, job.Source, job.Status,
		job.CurrentPage, job.TotalPages,
		job.Processed, job.Created, job.Skipped,
		errorsJSON,
		job.StartedAt, job.FinishedAt, now,
	)
	return err
}

// GetByID fetches an ImportJob by its UUID string.
func (r *importJobRepository) GetByID(ctx context.Context, id string) (*domain.ImportJob, error) {
	var job domain.ImportJob
	err := r.db.GetContext(ctx, &job, `
		SELECT id, source, status, current_page, total_pages, processed, created, skipped,
		       errors, started_at, finished_at, updated_at
		FROM import_jobs
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	job.Errors, _ = unmarshalErrors(job.ErrorsJSON)
	return &job, nil
}

// GetLatest fetches the most recently updated ImportJob for a given source.
func (r *importJobRepository) GetLatest(ctx context.Context, source string) (*domain.ImportJob, error) {
	var job domain.ImportJob
	err := r.db.GetContext(ctx, &job, `
		SELECT id, source, status, current_page, total_pages, processed, created, skipped,
		       errors, started_at, finished_at, updated_at
		FROM import_jobs
		WHERE source = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, source)
	if err != nil {
		return nil, err
	}

	job.Errors, _ = unmarshalErrors(job.ErrorsJSON)
	return &job, nil
}

// UpdateProgress persists current progress counters and status of a running job.
func (r *importJobRepository) UpdateProgress(ctx context.Context, job *domain.ImportJob) error {
	errorsJSON, err := marshalErrors(job.Errors)
	if err != nil {
		return fmt.Errorf("import_job: marshal errors on update: %w", err)
	}

	now := time.Now().UTC()
	job.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, `
		UPDATE import_jobs SET
			status       = $2,
			current_page = $3,
			total_pages  = $4,
			processed    = $5,
			created      = $6,
			skipped      = $7,
			errors       = $8,
			started_at   = $9,
			finished_at  = $10,
			updated_at   = $11
		WHERE id = $1
	`,
		job.ID, job.Status,
		job.CurrentPage, job.TotalPages,
		job.Processed, job.Created, job.Skipped,
		errorsJSON,
		job.StartedAt, job.FinishedAt, now,
	)
	return err
}

// Cancel marks a pending or running job as canceled.
func (r *importJobRepository) Cancel(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_jobs
		SET status = $2, updated_at = $3
		WHERE id = $1 AND status IN ('pending', 'running')
	`, id, domain.ImportJobCanceled, time.Now().UTC())
	return err
}

// CleanStaleJobs marks any 'running' or 'pending' jobs as 'failed' because the server restarted.
func (r *importJobRepository) CleanStaleJobs(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE import_jobs
		SET status = 'failed', errors = '["Server restarted while job was active"]', finished_at = $1, updated_at = $1
		WHERE status IN ('pending', 'running')
	`, time.Now().UTC())
	return err
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func marshalErrors(errs []string) (string, error) {
	if errs == nil {
		errs = []string{}
	}
	b, err := json.Marshal(errs)
	return string(b), err
}

func unmarshalErrors(raw string) ([]string, error) {
	var errs []string
	if raw == "" || raw == "null" {
		return []string{}, nil
	}
	err := json.Unmarshal([]byte(raw), &errs)
	return errs, err
}

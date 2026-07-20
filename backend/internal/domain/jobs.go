package domain

import (
	"context"
	"time"
)

type JobsRepository interface {
	// CalculateDailyRankings aggregates views, ratings, and creates snapshots of ranks in ranking_histories.
	CalculateDailyRankings(ctx context.Context) error

	// SnapshotRankingPositions calculates current positions and persists them as prev_rank in songs table.
	SnapshotRankingPositions(ctx context.Context) error

	// SynchronizeDailySiteMetrics aggregates growth metrics like new users and ratings for the current/previous day.
	SynchronizeDailySiteMetrics(ctx context.Context) error
}

// ─── Import Jobs ─────────────────────────────────────────────────────────────

// ImportJobStatus represents the lifecycle state of a bulk import job.
type ImportJobStatus string

const (
	ImportJobPending  ImportJobStatus = "pending"
	ImportJobRunning  ImportJobStatus = "running"
	ImportJobDone     ImportJobStatus = "done"
	ImportJobFailed   ImportJobStatus = "failed"
	ImportJobCanceled ImportJobStatus = "canceled"
)

// ImportJob tracks the progress of a bulk data import from an external source.
type ImportJob struct {
	ID          string          `db:"id" json:"id"`
	Source      string          `db:"source" json:"source"`       // "animethemes" | "anilist" | "backfill_titles" | "at_song_incremental"
	Status      ImportJobStatus `db:"status" json:"status"`
	CurrentPage int             `db:"current_page" json:"current_page"`
	TotalPages  int             `db:"total_pages" json:"total_pages"`
	Processed   int             `db:"processed" json:"processed"`
	Created     int             `db:"created" json:"created"`
	Skipped     int             `db:"skipped" json:"skipped"`
	ErrorsJSON  string          `db:"errors" json:"errors_json"`       // JSONB serialized []string
	StartedAt   *time.Time      `db:"started_at" json:"started_at"`
	FinishedAt  *time.Time      `db:"finished_at" json:"finished_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`

	// Transient — not persisted, populated from ErrorsJSON after fetch
	Errors []string `db:"-" json:"errors"`
	// Transient — 1 = AnimeThemes import, 2 = AniList enrichment (bulk import only)
	Phase int `db:"-" json:"phase"`
}

// ImportJobRepository handles persistence of ImportJob entities.
type ImportJobRepository interface {
	Create(ctx context.Context, job *ImportJob) error
	GetByID(ctx context.Context, id string) (*ImportJob, error)
	GetLatest(ctx context.Context, source string) (*ImportJob, error)
	UpdateProgress(ctx context.Context, job *ImportJob) error
	Cancel(ctx context.Context, id string) error
	CleanStaleJobs(ctx context.Context) error
}


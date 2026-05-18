-- Migration: 202505181200_create_import_jobs_table.sql
-- Creates the import_jobs table to track bulk data import progress.
-- Idempotent: uses IF NOT EXISTS throughout.

CREATE TABLE IF NOT EXISTS import_jobs (
    id           TEXT        PRIMARY KEY,
    source       TEXT        NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending',
    current_page INT         NOT NULL DEFAULT 0,
    total_pages  INT         NOT NULL DEFAULT 0,
    processed    INT         NOT NULL DEFAULT 0,
    created      INT         NOT NULL DEFAULT 0,
    skipped      INT         NOT NULL DEFAULT 0,
    errors       JSONB       NOT NULL DEFAULT '[]'::jsonb,
    started_at   TIMESTAMPTZ,
    finished_at  TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_import_jobs_source ON import_jobs (source);
CREATE INDEX IF NOT EXISTS idx_import_jobs_status ON import_jobs (status);

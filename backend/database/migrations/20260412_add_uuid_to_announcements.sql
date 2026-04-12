-- Migration: Add UUID to announcements
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS uuid UUID;

-- Enable pgcrypto for gen_random_uuid() if not yet enabled (Postgres < 13)
-- CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Populate existing rows with UUIDs
UPDATE announcements SET uuid = gen_random_uuid() WHERE uuid IS NULL;

-- Make it NOT NULL and UNIQUE
ALTER TABLE announcements ALTER COLUMN uuid SET NOT NULL;
ALTER TABLE announcements ADD CONSTRAINT announcements_uuid_unique UNIQUE (uuid);

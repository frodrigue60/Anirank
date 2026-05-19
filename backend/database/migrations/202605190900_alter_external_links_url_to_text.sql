-- Migration: Alter external_links.url to TEXT
-- Description: Solves SQLSTATE 22001 (value too long) when AniList provides urls longer than 191 characters.

ALTER TABLE external_links ALTER COLUMN url TYPE TEXT;

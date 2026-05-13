-- Migration to add UUID to tournament_matchups
ALTER TABLE tournament_matchups ADD COLUMN uuid UUID DEFAULT gen_random_uuid();

-- Add UNIQUE constraint and NOT NULL
ALTER TABLE tournament_matchups ALTER COLUMN uuid SET NOT NULL;
ALTER TABLE tournament_matchups ADD CONSTRAINT tournament_matchups_uuid_unique UNIQUE (uuid);

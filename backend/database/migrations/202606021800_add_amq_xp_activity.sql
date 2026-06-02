-- Idempotent migration to add AMQ completion XP activity
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM xp_activities WHERE key = 'amq_completion') THEN
        INSERT INTO xp_activities (key, xp_amount, description, cooldown_seconds)
        VALUES ('amq_completion', 15, 'Awarded upon completing an Anime Music Quiz match', 0);
    END IF;
END $$;

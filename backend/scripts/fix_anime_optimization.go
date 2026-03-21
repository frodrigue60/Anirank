package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	fmt.Println("Applying Anime Optimization (songs_count)...")

	sql := `
	-- 1. Add column if not exists
	DO $$ 
	BEGIN 
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='animes' AND column_name='songs_count') THEN
			ALTER TABLE animes ADD COLUMN songs_count INTEGER DEFAULT 0;
		END IF;
	END $$;

	-- 2. Initial sync
	UPDATE animes a SET songs_count = (SELECT COUNT(*) FROM songs s WHERE s.anime_id = a.id);

	-- 3. Trigger function
	CREATE OR REPLACE FUNCTION update_anime_songs_count()
	RETURNS TRIGGER AS $$
	BEGIN
		IF (TG_OP = 'INSERT') THEN
			UPDATE animes SET songs_count = songs_count + 1 WHERE id = NEW.anime_id;
		ELSIF (TG_OP = 'DELETE') THEN
			UPDATE animes SET songs_count = GREATEST(0, songs_count - 1) WHERE id = OLD.anime_id;
		ELSIF (TG_OP = 'UPDATE' AND OLD.anime_id <> NEW.anime_id) THEN
			UPDATE animes SET songs_count = GREATEST(0, songs_count - 1) WHERE id = OLD.anime_id;
			UPDATE animes SET songs_count = songs_count + 1 WHERE id = NEW.anime_id;
		END IF;
		RETURN NULL;
	END;
	$$ LANGUAGE plpgsql;

	-- 4. Trigger
	DROP TRIGGER IF EXISTS trg_update_anime_songs_count ON songs;
	CREATE TRIGGER trg_update_anime_songs_count
	AFTER INSERT OR DELETE OR UPDATE ON songs
	FOR EACH ROW EXECUTE FUNCTION update_anime_songs_count();
	`

	_, err = db.Exec(sql)
	if err != nil {
		log.Fatalf("Failed to execute SQL: %v", err)
	}

	fmt.Println("✅ Anime Optimization Applied Successfully!")
}

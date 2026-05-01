package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	_ = godotenv.Load("../.env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	}

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	queries := []struct {
		name  string
		query string
	}{
		{
			"Global Ranking (Score)",
			"EXPLAIN ANALYZE SELECT song_romaji, average_score FROM songs WHERE status = true ORDER BY average_score DESC LIMIT 50",
		},
		{
			"Seasonal Ranking (Views)",
			"EXPLAIN ANALYZE SELECT song_romaji FROM songs WHERE season_id = (SELECT id FROM seasons LIMIT 1) AND year_id = (SELECT id FROM years LIMIT 1) AND status = true ORDER BY views DESC LIMIT 20",
		},
		{
			"User Leaderboard (XP)",
			"EXPLAIN ANALYZE SELECT name, xp FROM users ORDER BY xp DESC LIMIT 10",
		},
		{
			"User Lookup by UUID",
			"EXPLAIN ANALYZE SELECT id FROM users WHERE uuid = '550e8400-e29b-41d4-a716-446655440000' LIMIT 1",
		},
		{
			"Recent Activity Feed",
			"EXPLAIN ANALYZE SELECT a.*, u.name FROM activities a JOIN users u ON a.user_id = u.id ORDER BY a.created_at DESC LIMIT 30",
		},
	}

	for _, q := range queries {
		fmt.Printf("\n=== AUDIT: %s ===\n", q.name)
		rows, err := db.Query(q.query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		for rows.Next() {
			var line string
			rows.Scan(&line)
			fmt.Println(line)
		}
		rows.Close()
	}
}

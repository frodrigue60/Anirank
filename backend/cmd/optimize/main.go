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

	statements := []string{
		"CREATE INDEX IF NOT EXISTS idx_songs_status_score_desc ON songs (average_score DESC) WHERE status = true",
		"CREATE INDEX IF NOT EXISTS idx_songs_seasonal_perf ON songs (season_id, year_id, status, views DESC)",
		"CREATE INDEX IF NOT EXISTS idx_activities_user_lookup ON activities (user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_users_xp_ranking ON users (xp DESC)",
	}

	for _, stmt := range statements {
		fmt.Printf("Executing: %s\n", stmt)
		_, err := db.Exec(stmt)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("Success!")
		}
	}

	fmt.Println("\nIndexes applied successfully. Running ANALYZE to update statistics...")
	db.Exec("ANALYZE songs")
	db.Exec("ANALYZE users")
	db.Exec("ANALYZE activities")
	fmt.Println("Database stats updated.")
}

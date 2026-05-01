package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	_ = godotenv.Load("../.env")
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found in cwd, using OS env vars")
	}

	dbURL := os.Getenv("DATABASE_URL")
	var db *sqlx.DB
	var err error

	if dbURL == "" {
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPass, dbHost, dbPort, dbName)
	}

	db, err = sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Starting Large Scale Seeder...")
	start := time.Now()

	// 1. Get taxonomy IDs (Years, Seasons, Formats)
	var yearIDs []int64
	var seasonIDs []int64
	var formatIDs []int64

	_ = db.Select(&yearIDs, "SELECT id FROM years")
	_ = db.Select(&seasonIDs, "SELECT id FROM seasons")
	_ = db.Select(&formatIDs, "SELECT id FROM formats")

	fmt.Printf("Found %d years, %d seasons, %d formats\n", len(yearIDs), len(seasonIDs), len(formatIDs))

	if len(yearIDs) == 0 {
		fmt.Println("Seeding basic years...")
		db.MustExec("INSERT INTO years (name, slug, current) VALUES ('2024', '2024', true), ('2023', '2023', false)")
		db.Select(&yearIDs, "SELECT id FROM years")
	}
	if len(seasonIDs) == 0 {
		fmt.Println("Seeding basic seasons...")
		db.MustExec("INSERT INTO seasons (name, slug, current) VALUES ('Winter', 'winter', true), ('Spring', 'spring', false)")
		db.Select(&seasonIDs, "SELECT id FROM seasons")
	}
	if len(formatIDs) == 0 {
		fmt.Println("Seeding basic formats...")
		db.MustExec("INSERT INTO formats (name, slug) VALUES ('TV', 'tv'), ('Movie', 'movie')")
		db.Select(&formatIDs, "SELECT id FROM formats")
	}

	if len(yearIDs) == 0 || len(seasonIDs) == 0 || len(formatIDs) == 0 {
		log.Fatalf("Taxonomy tables are STILL empty. Check database permissions.")
	}

	// 2. Seed Animes
	animeCount := 1000
	fmt.Printf("Seeding %d Animes...\n", animeCount)
	animeIDs := seedAnimes(db, animeCount, yearIDs, seasonIDs, formatIDs)

	// 3. Seed Songs
	songCount := 50000
	fmt.Printf("Seeding %d Songs...\n", songCount)
	songIDs := seedSongs(db, songCount, animeIDs, yearIDs, seasonIDs)

	// 4. Seed Users
	userCount := 10000
	fmt.Printf("Seeding %d Users...\n", userCount)
	userIDs := seedUsers(db, userCount)

	// 5. Seed Ratings
	ratingCount := 100000
	fmt.Printf("Seeding %d Ratings...\n", ratingCount)
	seedRatings(db, ratingCount, userIDs, songIDs)

	fmt.Printf("Done! Total time: %v\n", time.Since(start))
}

func seedAnimes(db *sqlx.DB, count int, years, seasons, formats []int64) []int64 {
	ids := []int64{}
	batchSize := 500
	for i := 0; i < count; i += batchSize {
		size := batchSize
		if i+batchSize > count {
			size = count - i
		}

		query := "INSERT INTO animes (title, slug, status, year_id, season_id, format_id, uuid, created_at, updated_at) VALUES "
		vals := []interface{}{}
		for j := 0; j < size; j++ {
			title := gofakeit.Sentence(3)
			slug := strings.ToLower(strings.ReplaceAll(title, " ", "-")) + "-" + uuid.New().String()[:8]
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, NOW(), NOW()),", j*7+1, j*7+2, j*7+3, j*7+4, j*7+5, j*7+6, j*7+7)
			vals = append(vals, title, slug, true, randChoice(years), randChoice(seasons), randChoice(formats), uuid.New())
		}
		query = query[:len(query)-1] + " RETURNING id"
		
		var batchIDs []int64
		err := db.Select(&batchIDs, query, vals...)
		if err != nil {
			log.Fatalf("Error seeding animes: %v", err)
		}
		ids = append(ids, batchIDs...)
	}
	return ids
}

func seedSongs(db *sqlx.DB, count int, animeIDs, years, seasons []int64) []int64 {
	ids := []int64{}
	batchSize := 500
	types := []string{"OP", "ED", "INS"}
	
	for i := 0; i < count; i += batchSize {
		size := batchSize
		if i+batchSize > count {
			size = count - i
		}

		query := "INSERT INTO songs (song_romaji, type, slug, anime_id, year_id, season_id, status, uuid, created_at, updated_at) VALUES "
		vals := []interface{}{}
		for j := 0; j < size; j++ {
			title := gofakeit.BookTitle()
			slug := strings.ToLower(strings.ReplaceAll(title, " ", "-")) + "-" + uuid.New().String()[:8]
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, NOW(), NOW()),", j*8+1, j*8+2, j*8+3, j*8+4, j*8+5, j*8+6, j*8+7, j*8+8)
			vals = append(vals, title, randString(types), slug, randChoice(animeIDs), randChoice(years), randChoice(seasons), true, uuid.New())
		}
		query = query[:len(query)-1] + " RETURNING id"
		
		var batchIDs []int64
		err := db.Select(&batchIDs, query, vals...)
		if err != nil {
			log.Fatalf("Error seeding songs: %v", err)
		}
		ids = append(ids, batchIDs...)
	}
	return ids
}

func seedUsers(db *sqlx.DB, count int) []int64 {
	ids := []int64{}
	batchSize := 500
	for i := 0; i < count; i += batchSize {
		size := batchSize
		if i+batchSize > count {
			size = count - i
		}

		query := "INSERT INTO users (uuid, name, slug, email, password, created_at, updated_at) VALUES "
		vals := []interface{}{}
		for j := 0; j < size; j++ {
			name := gofakeit.Name()
			email := gofakeit.Email() + "-" + uuid.New().String()[:4]
			slug := strings.ToLower(strings.ReplaceAll(name, " ", "-")) + "-" + uuid.New().String()[:4]
			query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, NOW(), NOW()),", j*5+1, j*5+2, j*5+3, j*5+4, j*5+5)
			vals = append(vals, uuid.New(), name, slug, email, "$2a$10$Un6jXm5Z6W5Z6W5Z6W5Z6u") // hashed 'password'
		}
		query = query[:len(query)-1] + " RETURNING id"
		
		var batchIDs []int64
		err := db.Select(&batchIDs, query, vals...)
		if err != nil {
			log.Fatalf("Error seeding users: %v", err)
		}
		ids = append(ids, batchIDs...)
	}
	return ids
}

func seedRatings(db *sqlx.DB, count int, userIDs, songIDs []int64) {
	batchSize := 1000
	for i := 0; i < count; i += batchSize {
		size := batchSize
		if i+batchSize > count {
			size = count - i
		}

		query := "INSERT INTO song_ratings (rating, user_id, song_id, created_at, updated_at) VALUES "
		vals := []interface{}{}
		for j := 0; j < size; j++ {
			query += fmt.Sprintf("($%d, $%d, $%d, NOW(), NOW()),", j*3+1, j*3+2, j*3+3)
			vals = append(vals, rand.Intn(10)+1, randChoice(userIDs), randChoice(songIDs))
		}
		query = query[:len(query)-1]
		
		_, err := db.Exec(query, vals...)
		if err != nil {
			log.Fatalf("Error seeding ratings: %v", err)
		}
	}
}

func randChoice(choices []int64) int64 {
	return choices[rand.Intn(len(choices))]
}

func randString(choices []string) string {
	return choices[rand.Intn(len(choices))]
}

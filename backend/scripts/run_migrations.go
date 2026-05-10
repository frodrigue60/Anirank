package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	_ = godotenv.Load()        // Try local backend/.env
	_ = godotenv.Load("../.env") // Try root .env

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback to individual variables
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		if dbUser != "" && dbHost != "" {
			dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
		} else {
			log.Fatal("DATABASE_URL or DB_USER/DB_HOST not found in .env")
		}
	}

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("Connected to database. Checking state...")

	// 1. Check if DB is empty
	var tableCount int
	err = db.Get(&tableCount, "SELECT count(*) FROM pg_catalog.pg_tables WHERE schemaname = 'public'")
	if err != nil {
		log.Fatalf("Failed to count tables: %v", err)
	}

	migrationDir := "./database/migrations"

	if tableCount <= 1 { // Only 'migrations' or empty
		fmt.Println("Database appears empty. Running schema_dump.sql first...")
		dumpPath := filepath.Join(migrationDir, "schema_dump.sql")
		content, err := os.ReadFile(dumpPath)
		if err != nil {
			log.Fatalf("Failed to read schema_dump.sql: %v", err)
		}

		// Clean up the content for generic SQL execution if needed
		// (e.g., removing \ commands if they cause issues)
		lines := strings.Split(string(content), "\n")
		var cleanLines []string
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "\\") {
				continue
			}
			cleanLines = append(cleanLines, line)
		}
		cleanSQL := strings.Join(cleanLines, "\n")

		_, err = db.Exec(cleanSQL)
		if err != nil {
			log.Printf("Warning: schema_dump.sql had some errors (likely owner/permission related), but continuing... Error: %v", err)
		} else {
			fmt.Println("schema_dump.sql executed successfully.")
		}
		
		// Reset search path because schema_dump often clears it
		_, _ = db.Exec("SET search_path TO public")
	}

	// 2. Ensure migrations table exists (in case schema_dump didn't have it)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			migration VARCHAR(191) NOT NULL,
			batch INTEGER NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// 3. Get list of already run migrations
	var runMigrations []string
	err = db.Select(&runMigrations, "SELECT migration FROM migrations")
	if err != nil {
		// If table was just created by schema_dump but has different structure, this might fail.
		// But usually it's fine.
		log.Printf("Note: Could not fetch from migrations table (might be empty or missing columns): %v", err)
	}

	runMap := make(map[string]bool)
	for _, m := range runMigrations {
		runMap[m] = true
	}

	// 4. Find all migration files
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	var pendingFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") && f.Name() != "schema_dump.sql" {
			if !runMap[f.Name()] {
				pendingFiles = append(pendingFiles, f.Name())
			}
		}
	}

	sort.Strings(pendingFiles)

	if len(pendingFiles) == 0 {
		fmt.Println("No pending migrations found.")
		return
	}

	// 5. Run pending migrations
	fmt.Printf("Found %d pending migrations. Executing...\n", len(pendingFiles))

	var lastBatch int
	_ = db.Get(&lastBatch, "SELECT COALESCE(MAX(batch), 0) FROM migrations")
	newBatch := lastBatch + 1

	for _, filename := range pendingFiles {
		fmt.Printf("Running migration: %s... ", filename)
		content, err := os.ReadFile(filepath.Join(migrationDir, filename))
		if err != nil {
			log.Fatalf("\nFailed to read migration file %s: %v", filename, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			log.Fatalf("\nFailed to start transaction: %v", err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			log.Fatalf("\nFailed to execute migration %s: %v", filename, err)
		}

		_, err = tx.Exec("INSERT INTO migrations (migration, batch) VALUES ($1, $2)", filename, newBatch)
		if err != nil {
			tx.Rollback()
			log.Fatalf("\nFailed to record migration %s: %v", filename, err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatalf("\nFailed to commit migration %s: %v", filename, err)
		}
		fmt.Println("SUCCESS")
	}

	fmt.Println("All migrations completed successfully!")
}

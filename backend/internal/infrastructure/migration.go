package infrastructure

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

// RunMigrations executes all pending SQL migrations in the specified directory.
// It tracks already executed migrations in a 'migrations' table.
func RunMigrations(db *sqlx.DB, migrationDir string) error {
	log.Printf("🔍 Checking for pending migrations in: %s", migrationDir)

	// 1. Ensure migrations table exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			migration VARCHAR(191) NOT NULL,
			batch INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 2. Check if DB is empty (only migrations table)
	var tableCount int
	err = db.Get(&tableCount, "SELECT count(*) FROM pg_catalog.pg_tables WHERE schemaname = 'public'")
	if err != nil {
		return fmt.Errorf("failed to count tables: %w", err)
	}

	// If only 'migrations' table exists or empty, we might want to run schema_dump.sql
	// but in production we should be careful. The original script does it.
	if tableCount <= 1 {
		dumpPath := filepath.Join(migrationDir, "schema_dump.sql")
		if _, err := os.Stat(dumpPath); err == nil {
			log.Println("📂 Database appears empty. Executing schema_dump.sql...")
			content, err := os.ReadFile(dumpPath)
			if err != nil {
				return fmt.Errorf("failed to read schema_dump.sql: %w", err)
			}

			// Clean SQL for generic execution
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
				log.Printf("⚠️  Warning: schema_dump.sql had some errors (likely owner/permission related): %v", err)
			} else {
				log.Println("✅ schema_dump.sql executed successfully")
			}
			
			// Reset search path
			_, _ = db.Exec("SET search_path TO public")
		}
	}

	// 3. Get list of already run migrations
	var runMigrations []string
	err = db.Select(&runMigrations, "SELECT migration FROM migrations")
	if err != nil {
		log.Printf("ℹ️  Note: Could not fetch from migrations table (might be empty): %v", err)
	}

	runMap := make(map[string]bool)
	for _, m := range runMigrations {
		runMap[m] = true
	}

	// 4. Find all migration files
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
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
		log.Println("✅ No pending migrations found.")
		return nil
	}

	// 5. Run pending migrations
	log.Printf("🚀 Found %d pending migrations. Executing...", len(pendingFiles))

	var lastBatch int
	_ = db.Get(&lastBatch, "SELECT COALESCE(MAX(batch), 0) FROM migrations")
	newBatch := lastBatch + 1

	for _, filename := range pendingFiles {
		log.Printf("🔨 Running migration: %s", filename)
		content, err := os.ReadFile(filepath.Join(migrationDir, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("failed to start transaction for %s: %w", filename, err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		_, err = tx.Exec("INSERT INTO migrations (migration, batch) VALUES ($1, $2)", filename, newBatch)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}
		log.Printf("✅ Success: %s", filename)
	}

	log.Println("✨ All migrations completed successfully!")
	return nil
}

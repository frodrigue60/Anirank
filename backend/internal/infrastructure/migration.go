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
	absPath, err := filepath.Abs(migrationDir)
	if err != nil {
		log.Printf("⚠️  Warning: Could not resolve absolute path for migrations: %v", err)
		absPath = migrationDir
	}

	log.Printf("🔍 Checking for pending migrations in: %s", absPath)

	// 1. Ensure migrations table exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			migration VARCHAR(191) NOT NULL,
			batch INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 2. Check current DB state
	var tableCount int
	err = db.Get(&tableCount, "SELECT count(*) FROM pg_catalog.pg_tables WHERE schemaname = 'public'")
	if err != nil {
		return fmt.Errorf("failed to count tables: %w", err)
	}

	// If only 'migrations' table exists or empty, we run schema_dump.sql
	if tableCount <= 1 {
		dumpPath := filepath.Join(absPath, "schema_dump.sql")
		if _, err := os.Stat(dumpPath); err == nil {
			log.Println("📂 Database appears empty. Executing schema_dump.sql...")
			content, err := os.ReadFile(dumpPath)
			if err != nil {
				return fmt.Errorf("failed to read schema_dump.sql: %w", err)
			}

			// Clean SQL for generic execution (remove \ commands)
			lines := strings.Split(string(content), "\n")
			var cleanLines []string
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "\\") || strings.HasPrefix(trimmed, "--") && (strings.Contains(trimmed, "PostgreSQL") || strings.Contains(trimmed, "Dumped")) {
					continue
				}
				cleanLines = append(cleanLines, line)
			}
			cleanSQL := strings.Join(cleanLines, "\n")

			_, err = db.Exec(cleanSQL)
			if err != nil {
				log.Printf("⚠️  Warning: schema_dump.sql had some errors (continuing...): %v", err)
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
		log.Printf("ℹ️  Note: Could not fetch from migrations table: %v", err)
	}

	runMap := make(map[string]bool)
	for _, m := range runMigrations {
		runMap[m] = true
	}

	// 4. Find all migration files
	files, err := os.ReadDir(absPath)
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
		content, err := os.ReadFile(filepath.Join(absPath, filename))
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
			// Check if error is "already exists" - if so, maybe we should record it and continue?
			// But it's safer to use IF NOT EXISTS in the SQL.
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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found.")
	}
	ctx := context.Background()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	fmt.Println("Starting Taxonomy Deduplication & Slug Sanitization...")

	// 1. Dedupe Studios
	if err := dedupeEntity(ctx, pool, "studios", "anime_studio", "studio_id"); err != nil {
		log.Printf("Error deduping studios: %v\n", err)
	}

	// 2. Dedupe Producers
	if err := dedupeEntity(ctx, pool, "producers", "anime_producer", "producer_id"); err != nil {
		log.Printf("Error deduping producers: %v\n", err)
	}

	// 3. Dedupe Genres
	if err := dedupeEntity(ctx, pool, "genres", "anime_genre", "genre_id"); err != nil {
		log.Printf("Error deduping genres: %v\n", err)
	}

	fmt.Println("Deduplication Complete!")
}

type Entity struct {
	ID   uint64
	Name string
	Slug string
}

func dedupeEntity(ctx context.Context, pool *pgxpool.Pool, table, bridgeTable, bridgeFK string) error {
	fmt.Printf("\n--- Processing Table: %s ---\n", table)

	// Fetch all records
	rows, err := pool.Query(ctx, fmt.Sprintf("SELECT id, name, slug FROM %s ORDER BY id ASC", table))
	if err != nil {
		return err
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug); err != nil {
			continue
		}
		entities = append(entities, e)
	}

	// Group by sanitized slug
	groups := make(map[string][]Entity)
	for _, e := range entities {
		newSlug := slugify(e.Name)
		groups[newSlug] = append(groups[newSlug], e)
	}

	for slug, group := range groups {
		master := group[0]
		duplicates := group[1:]

		// 1. Sanitize Master Slug if different
		if master.Slug != slug {
			fmt.Printf("[%s] Updating master slug: %s -> %s\n", master.Name, master.Slug, slug)
			_, err = pool.Exec(ctx, fmt.Sprintf("UPDATE %s SET slug = $1, updated_at = NOW() WHERE id = $2", table), slug, master.ID)
			if err != nil {
				log.Printf("Error updating master slug: %v\n", err)
			}
		}

		if len(duplicates) == 0 {
			continue
		}

		fmt.Printf("[%s] Found %d duplicates. Merging into ID %d...\n", master.Name, len(duplicates), master.ID)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}

		for _, dup := range duplicates {
			// A. Move Relationships
			// We handle potential unique constraint conflicts by deleting the duplicate link if master link already exists
			
			// Identify conflicting links
			var conflictingAnimeIDs []uint64
			queryConflict := fmt.Sprintf(`
				SELECT anime_id FROM %s WHERE %s = $1 
				AND anime_id IN (SELECT anime_id FROM %s WHERE %s = $2)`, 
				bridgeTable, bridgeFK, bridgeTable, bridgeFK)
			
			confRows, err := tx.Query(ctx, queryConflict, dup.ID, master.ID)
			if err == nil {
				for confRows.Next() {
					var aID uint64
					if err := confRows.Scan(&aID); err == nil {
						conflictingAnimeIDs = append(conflictingAnimeIDs, aID)
					}
				}
				confRows.Close()
			}

			// Delete conflicting links from duplicate
			if len(conflictingAnimeIDs) > 0 {
				_, err = tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE %s = $1 AND anime_id = ANY($2)", bridgeTable, bridgeFK), dup.ID, conflictingAnimeIDs)
				if err != nil {
					tx.Rollback(ctx)
					return fmt.Errorf("failed to delete conflicting links: %v", err)
				}
			}

			// Update remaining links to master
			_, err = tx.Exec(ctx, fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2", bridgeTable, bridgeFK, bridgeFK), master.ID, dup.ID)
			if err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("failed to migrate links: %v", err)
			}

			// B. Delete Duplicate
			_, err = tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", table), dup.ID)
			if err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("failed to delete duplicate %s: %v", table, err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}
		fmt.Printf("[%s] Successfully merged.\n", master.Name)
	}

	return nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	// Replace underscores and spaces with hyphens
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric (except hyphen)
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s = result.String()

	// Remove duplicate hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	return strings.Trim(s, "-")
}

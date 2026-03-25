package main

import (
	"log"
	"os"

	"anirank/api/internal/infrastructure"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	// Load env from possible locations
	_ = godotenv.Load("../../.env")      // backend/.env
	_ = godotenv.Load("../../../.env")   // root/.env
	_ = godotenv.Load()                  // current dir

	dbURL := os.Getenv("DATABASE_URL")
	var db *sqlx.DB
	var err error

	if dbURL != "" {
		db, err = infrastructure.NewDatabaseConnectionFromURL(dbURL)
	} else {
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		db, err = infrastructure.NewDatabaseConnection(dbUser, dbPass, dbHost, dbPort, dbName)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// 1. Find duplicate artist names
	var duplicates []struct {
		Name  string `db:"name"`
		Count int    `db:"count"`
	}
	err = db.Select(&duplicates, "SELECT name, COUNT(*) FROM artists GROUP BY name HAVING COUNT(*) > 1")
	if err != nil {
		log.Fatalf("Failed to query duplicates: %v", err)
	}

	log.Printf("Found %d duplicate artist names", len(duplicates))

	for _, d := range duplicates {
		log.Printf("Processing artist: %s", d.Name)

		// Get all IDs for this name, ordered by ID (smallest first as master)
		var ids []uint64
		err = db.Select(&ids, "SELECT id FROM artists WHERE name = $1 ORDER BY id ASC", d.Name)
		if err != nil {
			log.Printf("  Error fetching IDs for %s: %v", d.Name, err)
			continue
		}

		if len(ids) < 2 {
			continue
		}

		masterID := ids[0]
		otherIDs := ids[1:]

		for _, otherID := range otherIDs {
			log.Printf("  Merging ID %d into %d", otherID, masterID)

			// Start Tx
			tx, err := db.Beginx()
			if err != nil {
				log.Printf("    Failed to start transaction: %v", err)
				continue
			}

			// Merge artist_song
			// Delete rows from other that already exist for master
			_, err = tx.Exec("DELETE FROM artist_song WHERE artist_id = $1 AND song_id IN (SELECT song_id FROM artist_song WHERE artist_id = $2)", otherID, masterID)
			if err != nil {
				log.Printf("    Error cleaning artist_song for %d: %v", otherID, err)
				tx.Rollback()
				continue
			}
			// Update remaining rows
			_, err = tx.Exec("UPDATE artist_song SET artist_id = $1 WHERE artist_id = $2", masterID, otherID)
			if err != nil {
				log.Printf("    Error updating artist_song for %d: %v", otherID, err)
				tx.Rollback()
				continue
			}

			// Merge artist_user
			_, err = tx.Exec("DELETE FROM artist_user WHERE artist_id = $1 AND user_id IN (SELECT user_id FROM artist_user WHERE artist_id = $2)", otherID, masterID)
			if err != nil {
				log.Printf("    Error cleaning artist_user for %d: %v", otherID, err)
				tx.Rollback()
				continue
			}
			_, err = tx.Exec("UPDATE artist_user SET artist_id = $1 WHERE artist_id = $2", masterID, otherID)
			if err != nil {
				log.Printf("    Error updating artist_user for %d: %v", otherID, err)
				tx.Rollback()
				continue
			}

			// Delete the duplicate artist
			_, err = tx.Exec("DELETE FROM artists WHERE id = $1", otherID)
			if err != nil {
				log.Printf("    Error deleting artist %d: %v", otherID, err)
				tx.Rollback()
				continue
			}

			if err := tx.Commit(); err != nil {
				log.Printf("    Failed to commit transaction: %v", err)
			} else {
				log.Printf("    Successfully merged ID %d", otherID)
			}
		}
	}

	log.Println("Artist merge complete!")
}

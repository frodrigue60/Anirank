package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"anirank/api/internal/infrastructure"
	"anirank/api/internal/pkg/avatar"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load env: repo-root .env then local .env as overrides.
	_ = godotenv.Load("../.env")
	_ = godotenv.Overload(".env")
	ctx := context.Background()

	// 1. DB Connection
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

	// 2. Storage Initialization (Dynamic R2/S3)
	storage, err := infrastructure.InitStorageFromEnv(ctx)
	if err != nil {
		log.Fatalf("Unable to initialize storage: %v\n", err)
	}

	log.Println("Starting Artist Avatar Generation & R2 Migration...")

	// 3. Clear existing avatars folder
	log.Println("Clearing existing avatars in artists/avatars/...")
	existingFiles, err := storage.ListFiles(ctx, "artists/avatars/")
	if err != nil {
		log.Printf("Warning: failed to list existing files for cleanup: %v\n", err)
	} else {
		for _, f := range existingFiles {
			_ = storage.DeleteFile(ctx, f)
		}
		log.Printf("Cleaned %d files.\n", len(existingFiles))
	}

	// 4. Fetch all artists
	rows, err := pool.Query(ctx, "SELECT id, name FROM artists")
	if err != nil {
		log.Fatalf("Query error: %v\n", err)
	}
	defer rows.Close()

	type artistRec struct {
		ID   uint64
		Name string
	}
	var artists []artistRec
	for rows.Next() {
		var a artistRec
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			continue
		}
		artists = append(artists, a)
	}

	log.Printf("Found %d total artists to process.\n", len(artists))

	// 5. Process each artist
	for i, a := range artists {
		log.Printf("[%d/%d] Processing %s...\n", i+1, len(artists), a.Name)

		// Generate Local Avatar (256px, AVIF)
		res, err := avatar.Generate(ctx, a.Name, 256)
		if err != nil {
			log.Printf("  Error generating for %s: %v\n", a.Name, err)
			continue
		}
		data := res.Data

		// Upload to R2
		contentType := "image/avif"
		newPath := fmt.Sprintf("artists/avatars/%d_%s.avif", a.ID, uuid.New().String()[:8])

		_, err = storage.UploadFile(ctx, newPath, bytes.NewReader(data), res.Size, contentType)
		if err != nil {
			log.Printf("  Error uploading to R2 for %s: %v\n", a.Name, err)
			continue
		}

		// Update DB
		_, err = pool.Exec(ctx, "UPDATE artists SET avatar = $1, updated_at = NOW() WHERE id = $2", newPath, a.ID)
		if err != nil {
			log.Printf("  Error updating DB for %s: %v\n", a.Name, err)
		} else {
			log.Printf("  Successfully generated avatar: %s\n", newPath)
		}

		// Small delay to be safe
		time.Sleep(50 * time.Millisecond)
	}

	log.Println("Artist Avatar Generation Complete!")
}

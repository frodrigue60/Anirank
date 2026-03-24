package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"anirank/api/internal/infrastructure"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found.")
	}
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

	// 2. Storage Initialization (R2/S3)
	storage, err := infrastructure.NewS3Storage(
		ctx,
		os.Getenv("S3_ACCESS_KEY"),
		os.Getenv("S3_SECRET_KEY"),
		os.Getenv("S3_REGION"),
		os.Getenv("S3_BUCKET"),
		os.Getenv("S3_ENDPOINT"),
		os.Getenv("S3_PUBLIC_URL"),
	)
	if err != nil {
		log.Fatalf("Unable to initialize storage: %v\n", err)
	}

	log.Println("Starting Artist Avatar Generation & R2 Migration...")

	// 3. Fetch artists without avatar
	rows, err := pool.Query(ctx, "SELECT id, name FROM artists WHERE avatar IS NULL OR avatar = ''")
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

	log.Printf("Found %d artists needing avatars.\n", len(artists))

	// 4. Process each artist
	for i, a := range artists {
		log.Printf("[%d/%d] Processing %s...\n", i+1, len(artists), a.Name)

		// Generate UI Avatar URL
		// Example: https://ui-avatars.com/api/?name=John+Doe&size=512&background=random&color=fff
		avatarSourceURL := fmt.Sprintf("https://ui-avatars.com/api/?name=%s&size=512&background=random&color=fff", url.QueryEscape(a.Name))

		// Download
		resp, err := http.Get(avatarSourceURL)
		if err != nil {
			log.Printf("  Error downloading for %s: %v\n", a.Name, err)
			continue
		}
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("  Error reading data for %s: %v\n", a.Name, err)
			continue
		}

		// Upload to R2
		contentType := "image/png"
		newPath := fmt.Sprintf("artists/avatars/%d_%s.png", a.ID, uuid.New().String()[:8])
		
		_, err = storage.UploadFile(ctx, newPath, bytes.NewReader(data), int64(len(data)), contentType)
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

		// Rate limit protection
		time.Sleep(200 * time.Millisecond)
	}

	log.Println("Artist Avatar Generation Complete!")
}

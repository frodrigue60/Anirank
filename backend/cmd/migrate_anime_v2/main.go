package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/disintegration/gift"
	"image"
	"image/jpeg"

	"anirank/api/internal/pkg/imageutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"net/http"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found.")
	}
	ctx := context.Background()

	// 1. PostgreSQL connection
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

	// 2. S3 / MinIO Client
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKeyID := os.Getenv("S3_ACCESS_KEY")
	secretAccessKey := os.Getenv("S3_SECRET_KEY")
	region := os.Getenv("S3_REGION")
	bucketName := os.Getenv("S3_BUCKET")

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		log.Fatal(err)
	}

	clientOpts := func(o *s3.Options) {
		o.UsePathStyle = true
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	}
	s3Client := s3.NewFromConfig(cfg, clientOpts)

	log.Println("Migration V2 (Anime Mirror & Path Fix) Started...")

	// 3. Process Animes
	query := "SELECT id, anilist_id, cover, banner FROM animes"
	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatal(err)
	}

	type animeRec struct {
		ID        uint64
		AnilistID *int64
		Cover     *string
		Banner    *string
	}
	var records []animeRec
	for rows.Next() {
		var r animeRec
		if err := rows.Scan(&r.ID, &r.AnilistID, &r.Cover, &r.Banner); err != nil {
			continue
		}
		records = append(records, r)
	}
	rows.Close()

	totalProcessed := 0
	totalErrors := 0

	for _, r := range records {
		updates := make(map[string]string)

		// Process Cover
		if r.Cover != nil && *r.Cover != "" {
			newPath, err := processMirrorAndClean(ctx, s3Client, bucketName, *r.Cover, "animes/covers", r.AnilistID)
			if err == nil && newPath != *r.Cover {
				updates["cover"] = newPath
			} else if err != nil {
				log.Printf("[Anime %d] Cover Error: %v\n", r.ID, err)
				totalErrors++
			}
		}

		// Process Banner
		if r.Banner != nil && *r.Banner != "" {
			newPath, err := processMirrorAndClean(ctx, s3Client, bucketName, *r.Banner, "animes/banners", r.AnilistID)
			if err == nil && newPath != *r.Banner {
				updates["banner"] = newPath
			} else if err != nil {
				log.Printf("[Anime %d] Banner Error: %v\n", r.ID, err)
				totalErrors++
			}
		}

		if len(updates) > 0 {
			setClauses := []string{}
			args := []interface{}{}
			argID := 1
			for field, val := range updates {
				setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argID))
				args = append(args, val)
				argID++
			}
			args = append(args, r.ID)
			updateQuery := fmt.Sprintf("UPDATE animes SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argID)
			_, err := pool.Exec(ctx, updateQuery, args...)
			if err != nil {
				log.Printf("[Anime %d] DB Update Error: %v\n", r.ID, err)
				totalErrors++
			} else {
				totalProcessed++
			}
		}
	}

	log.Printf("Migration V2 Complete! Updated: %d, Errors: %d\n", totalProcessed, totalErrors)
}

func processMirrorAndClean(ctx context.Context, s3Client *s3.Client, bucket, oldPath, targetPrefix string, anilistID *int64) (string, error) {
	// Skip if already webp AND in the right place
	if strings.HasSuffix(strings.ToLower(oldPath), ".webp") && strings.HasPrefix(oldPath, targetPrefix) {
		return oldPath, nil
	}

	var buffer []byte
	var err error

	if strings.HasPrefix(oldPath, "http") {
		// Mirror external image
		resp, err := http.Get(oldPath)
		if err != nil {
			return oldPath, fmt.Errorf("HTTP Get %s failed: %w", oldPath, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return oldPath, fmt.Errorf("HTTP status %d for %s", resp.StatusCode, oldPath)
		}
		buffer, err = io.ReadAll(resp.Body)
		if err != nil {
			return oldPath, err
		}
	} else {
		// Download from S3
		obj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(oldPath),
		})
		if err != nil {
			// If not found, skip
			return oldPath, fmt.Errorf("S3 GetObject %s failed: %w", oldPath, err)
		}
		defer obj.Body.Close()
		buffer, err = io.ReadAll(obj.Body)
		if err != nil {
			return oldPath, err
		}
	}

	// Optimize to JPEG (Supports AVIF/modern formats via fallback)
	img, _, err := imageutil.Decode(bytes.NewReader(buffer))
	if err != nil {
		return oldPath, fmt.Errorf("image decode failed for %s: %w", oldPath, err)
	}

	gi := gift.New()
	dst := image.NewRGBA(gi.Bounds(img.Bounds()))
	gi.Draw(dst, img)

	var b bytes.Buffer
	err = jpeg.Encode(&b, dst, &jpeg.Options{Quality: 80})
	if err != nil {
		return oldPath, fmt.Errorf("jpeg encode failed for %s: %w", oldPath, err)
	}
	newImage := b.Bytes()

	// Generate Clean Path
	idStr := "0"
	if anilistID != nil {
		idStr = fmt.Sprintf("%d", *anilistID)
	} else {
		// Try extracting from old path if possible (e.g. bx180746)
		re := regexp.MustCompile(`bx(\d+)`)
		match := re.FindStringSubmatch(oldPath)
		if len(match) > 1 {
			idStr = match[1]
		}
	}
	newPath := fmt.Sprintf("%s/%s_%s.jpg", targetPrefix, idStr, uuid.New().String())

	// Upload
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(newPath),
		Body:          bytes.NewReader(newImage),
		ContentType:   aws.String("image/jpeg"),
		ContentLength: aws.Int64(int64(len(newImage))),
	})
	if err != nil {
		return oldPath, fmt.Errorf("PutObject %s failed: %w", newPath, err)
	}

	log.Printf("Migrated: %s -> %s\n", oldPath, newPath)

	// Delete Old if it was internal
	if !strings.HasPrefix(oldPath, "http") {
		_, _ = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(oldPath),
		})
	}

	return newPath, nil
}

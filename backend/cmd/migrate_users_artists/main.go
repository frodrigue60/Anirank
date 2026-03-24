package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
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

	log.Println("Migration Step 2 (Users & Artists) Started...")

	// 3. Process Users
	processTable(ctx, pool, s3Client, bucketName, "users", []string{"avatar", "banner"}, "users")

	// 4. Process Artists
	processTable(ctx, pool, s3Client, bucketName, "artists", []string{"avatar"}, "artists/avatars")

	log.Println("Migration Step 2 Complete!")
}

func processTable(ctx context.Context, pool *pgxpool.Pool, s3Client *s3.Client, bucket, table string, fields []string, prefixBase string) {
	query := fmt.Sprintf("SELECT id, %s FROM %s", strings.Join(fields, ", "), table)
	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Printf("Error querying table %s: %v\n", table, err)
		return
	}
	defer rows.Close()

	type record struct {
		ID     uint64
		Values [](*string)
	}
	var records []record
	for rows.Next() {
		r := record{Values: make([]*string, len(fields))}
		scanArgs := make([]interface{}, len(fields)+1)
		scanArgs[0] = &r.ID
		for i := range fields {
			scanArgs[i+1] = &r.Values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			continue
		}
		records = append(records, r)
	}

	for _, r := range records {
		updates := make(map[string]string)
		for i, field := range fields {
			if r.Values[i] != nil && *r.Values[i] != "" {
				targetPrefix := prefixBase
				if table == "users" {
					targetPrefix = "users/" + field + "s" // users/avatars, users/banners
				}
				newPath, err := processMigration(ctx, s3Client, bucket, *r.Values[i], targetPrefix, r.ID)
				if err == nil && newPath != *r.Values[i] {
					updates[field] = newPath
				} else if err != nil {
					log.Printf("[%s %d] %s Error: %v\n", table, r.ID, field, err)
				}
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
			updateQuery := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", table, strings.Join(setClauses, ", "), argID)
			_, err := pool.Exec(ctx, updateQuery, args...)
			if err != nil {
				log.Printf("[%s %d] DB Update Error: %v\n", table, r.ID, err)
			}
		}
	}
}

func processMigration(ctx context.Context, s3Client *s3.Client, bucket, oldPath, targetPrefix string, id uint64) (string, error) {
	// Skip if already webp AND in the right place
	if strings.HasSuffix(strings.ToLower(oldPath), ".webp") && strings.HasPrefix(oldPath, targetPrefix) {
		return oldPath, nil
	}

	var buffer []byte
	var err error

	if strings.HasPrefix(oldPath, "http") {
		resp, err := http.Get(oldPath)
		if err != nil {
			return oldPath, fmt.Errorf("HTTP Get %s failed: %w", oldPath, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return oldPath, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		buffer, err = io.ReadAll(resp.Body)
		if err != nil {
			return oldPath, err
		}
	} else {
		obj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(oldPath),
		})
		if err != nil {
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
		return oldPath, fmt.Errorf("image decode failed: %w", err)
	}

	gi := gift.New()
	dst := image.NewRGBA(gi.Bounds(img.Bounds()))
	gi.Draw(dst, img)

	var b bytes.Buffer
	err = jpeg.Encode(&b, dst, &jpeg.Options{Quality: 80})
	if err != nil {
		return oldPath, fmt.Errorf("jpeg encode failed: %w", err)
	}
	newImage := b.Bytes()

	// Generate Clean Path
	newPath := fmt.Sprintf("%s/%d_%s.jpg", targetPrefix, id, uuid.New().String())

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

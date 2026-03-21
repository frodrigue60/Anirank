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
	"github.com/h2non/bimg"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found in cwd or failed to load.")
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
		log.Fatalln("Failed to load AWS config:", err)
	}

	clientOpts := func(o *s3.Options) {
		o.UsePathStyle = true
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	}
	s3Client := s3.NewFromConfig(cfg, clientOpts)

	log.Println("Migration Script Started. Processing Images...")

	// 3. Define tables
	type target struct {
		Table  string
		Fields []string
	}
	targets := []target{
		{"users", []string{"avatar", "banner"}},
		{"animes", []string{"cover", "banner"}},
		{"artists", []string{"avatar"}},
		{"announcements", []string{"image"}},
	}

	totalProcessed := 0
	totalErrors := 0
	totalSkipped := 0

	for _, t := range targets {
		log.Printf("Processing table: %s...\n", t.Table)
		fieldsStr := strings.Join(t.Fields, ", ")
		query := fmt.Sprintf("SELECT id, %s FROM %s", fieldsStr, t.Table)

		rows, err := pool.Query(ctx, query)
		if err != nil {
			log.Printf("Skipping table %s: %v\n", t.Table, err)
			continue
		}

		// Read into slices to avoid keeping the connection row iterator open for too long
		type record struct {
			ID     uint64
			Fields []string
		}
		var records []record

		for rows.Next() {
			values := make([]interface{}, len(t.Fields)+1)
			var id uint64
			values[0] = &id

			stringPointers := make([]*string, len(t.Fields))
			for i := range stringPointers {
				values[i+1] = &stringPointers[i]
			}

			if err := rows.Scan(values...); err != nil {
				continue
			}

			rec := record{ID: id, Fields: make([]string, len(t.Fields))}
			for i, valPtr := range stringPointers {
				if valPtr != nil {
					rec.Fields[i] = *valPtr
				}
			}
			records = append(records, rec)
		}
		rows.Close()

		// Process records
		for _, rec := range records {
			updates := make(map[string]string)

			for i, fieldName := range t.Fields {
				oldPath := rec.Fields[i]
				if oldPath == "" || strings.HasSuffix(strings.ToLower(oldPath), ".webp") || strings.HasPrefix(oldPath, "http") {
					totalSkipped++
					continue
				}

				// Download Image
				obj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(oldPath),
				})
				if err != nil {
					log.Printf("[%s %d] Failed to get object %s: %v\n", t.Table, rec.ID, oldPath, err)
					totalErrors++
					continue
				}

				buffer, err := io.ReadAll(obj.Body)
				obj.Body.Close()
				if err != nil {
					log.Printf("[%s %d] Failed to read object %s: %v\n", t.Table, rec.ID, oldPath, err)
					totalErrors++
					continue
				}

				// Process Image with bimg
				options := bimg.Options{
					Type:    bimg.WEBP,
					Quality: 80,
				}

				newImage, err := bimg.NewImage(buffer).Process(options)
				if err != nil {
					log.Printf("[%s %d] Failed to convert image %s to webp: %v\n", t.Table, rec.ID, oldPath, err)
					totalErrors++
					continue
				}

				// Upload New Image
				lastIndex := strings.LastIndex(oldPath, ".")
				var newPath string
				if lastIndex == -1 {
					newPath = oldPath + ".webp"
				} else {
					newPath = oldPath[:lastIndex] + ".webp"
				}

				_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
					Bucket:      aws.String(bucketName),
					Key:         aws.String(newPath),
					Body:        bytes.NewReader(newImage),
					ContentType: aws.String("image/webp"),
					ContentLength: aws.Int64(int64(len(newImage))),
				})
				if err != nil {
					log.Printf("[%s %d] Failed to upload new image %s: %v\n", t.Table, rec.ID, newPath, err)
					totalErrors++
					continue
				}

				updates[fieldName] = newPath
				log.Printf("[%s %d] Converted %s -> %s\n", t.Table, rec.ID, oldPath, newPath)
				totalProcessed++

				// Delete Old Image
				_, err = s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(oldPath),
				})
				if err != nil {
					log.Printf("[%s %d] Warning: failed to delete old image %s: %v\n", t.Table, rec.ID, oldPath, err)
				}
			}

			// Update DB
			if len(updates) > 0 {
				setClauses := []string{}
				args := []interface{}{}
				argID := 1
				for fieldName, newPath := range updates {
					setClauses = append(setClauses, fmt.Sprintf("%s = $%d", fieldName, argID))
					args = append(args, newPath)
					argID++
				}
				args = append(args, rec.ID)

				updateQuery := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", t.Table, strings.Join(setClauses, ", "), argID)
				_, err := pool.Exec(ctx, updateQuery, args...)
				if err != nil {
					log.Printf("[%s %d] Failed to update database: %v\n", t.Table, rec.ID, err)
					totalErrors++
				}
			}
		}
	}

	log.Printf("Migration Complete! Processed: %d, Skipped: %d, Errors: %d\n", totalProcessed, totalSkipped, totalErrors)
}

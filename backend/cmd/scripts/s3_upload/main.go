package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"anirank/api/internal/pkg/avatar"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type Artist struct {
	ID   uint64 `db:"id"`
	Name string `db:"name"`
	Slug string `db:"slug"`
}

type Anime struct {
	ID        uint64 `db:"id"`
	AnilistID uint64 `db:"anilist_id"`
	Slug      string `db:"slug"`
}

type AnilistResponse struct {
	Data struct {
		Media struct {
			CoverImage struct {
				ExtraLarge string `json:"extraLarge"`
			} `json:"coverImage"`
			BannerImage string `json:"bannerImage"`
		} `json:"Media"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	} `json:"errors"`
}

func main() {
	// Load .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env not found, using system env variables")
	}

	// Connect DB
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := sqlx.Connect("mysql", dbURL)
	if err != nil {
		log.Fatalf("Connecting to DB failed: %v", err)
	}
	defer db.Close()

	// Connect S3
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("S3_REGION")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			"",
		)),
	)
	if err != nil {
		log.Fatalf("Loading AWS config failed: %v", err)
	}
	s3client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if os.Getenv("S3_ENDPOINT") != "" {
			o.BaseEndpoint = aws.String("http://" + os.Getenv("S3_ENDPOINT"))
			o.UsePathStyle = true
		}
	})
	bucket := os.Getenv("S3_BUCKET")

	log.Println("--- Starting Migration ---")

	// 1. Process Artists
	log.Println("Processing Artists...")
	var artists []Artist
	if err := db.Select(&artists, "SELECT id, name, slug FROM artists"); err != nil {
		log.Fatalf("Fetching artists failed: %v", err)
	}

	for _, artist := range artists {
		log.Printf("Generating avatar for artist: %s", artist.Name)

		// Generate Local Avatar (256px, AVIF)
		res, err := avatar.Generate(context.Background(), artist.Name, 256)
		if err != nil {
			log.Printf("Failed to generate avatar for %s: %v", artist.Name, err)
			continue
		}
		bodyBytes := res.Data

		s3Key := fmt.Sprintf("artists/%s-avatar-%d.avif", artist.Slug, artist.ID)

		// Upload to S3
		_, err = s3client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(s3Key),
			Body:        bytes.NewReader(bodyBytes),
			ContentType: aws.String("image/png"),
		})
	}

	// 2. Process Animes from AniList
	log.Println("Processing Animes from AniList...")
	var animes []Anime
	if err := db.Select(&animes, "SELECT id, anilist_id, slug FROM animes WHERE anilist_id IS NOT NULL"); err != nil {
		log.Fatalf("Fetching animes failed: %v", err)
	}

	for _, anime := range animes {
		log.Printf("Fetching AniList data for anime Slug: %s (AniList %d)", anime.Slug, anime.AnilistID)

		query := `query ($id: Int) { Media (id: $id, type: ANIME) { coverImage { extraLarge } bannerImage } }`
		variables := map[string]interface{}{"id": anime.AnilistID}
		reqBody, _ := json.Marshal(map[string]interface{}{
			"query":     query,
			"variables": variables,
		})

		req, _ := http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Failed to fetch from AniList for %s: %v", anime.Slug, err)
			continue
		}

		var alResp AnilistResponse
		if err := json.NewDecoder(resp.Body).Decode(&alResp); err != nil {
			log.Printf("Failed to decode AniList response for %s: %v", anime.Slug, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()
		if len(alResp.Errors) > 0 {
			log.Printf("AniList returned errors for %s: %v", anime.Slug, alResp.Errors)
			if alResp.Errors[0].Status == 429 {
				log.Println("Rate limited! Sleeping for 60 seconds...")
				time.Sleep(60 * time.Second)
			}
			continue
		}

		media := alResp.Data.Media

		if media.CoverImage.ExtraLarge != "" {
			uploadUrlToS3(s3client, db, bucket, anime.ID, anime.Slug, "thumbnail", media.CoverImage.ExtraLarge)
		}

		if media.BannerImage != "" {
			uploadUrlToS3(s3client, db, bucket, anime.ID, anime.Slug, "banner", media.BannerImage)
		}

		time.Sleep(3 * time.Second) // Be more conservative with rate limiting
	}

	log.Println("--- Migration Completed Successfully ---")
}

func uploadUrlToS3(s3client *s3.Client, db *sqlx.DB, bucket string, animeID uint64, slug, imageType, urlStr string) {
	resp, err := http.Get(urlStr)
	if err != nil {
		log.Printf("Failed to download %s from %s: %v", imageType, urlStr, err)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Printf("Failed to read %s body: %v", imageType, err)
		return
	}

	ext := "jpg"
	if strings.Contains(urlStr, ".png") {
		ext = "png"
	} else if strings.Contains(urlStr, ".webp") {
		ext = "webp"
	}
	s3Key := fmt.Sprintf("animes/%s-%s-%d.%s", slug, imageType, animeID, ext)
	contentType := "image/jpeg"
	if ext == "png" {
		contentType = "image/png"
	} else if ext == "webp" {
		contentType = "image/webp"
	}

	_, err = s3client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(s3Key),
		Body:        bytes.NewReader(bodyBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Printf("Failed to upload %s to S3: %v", imageType, err)
		return
	}

	var existingImageId uint64
	err = db.Get(&existingImageId, "SELECT id FROM images WHERE imageable_id = ? AND imageable_type = 'App\\\\Models\\\\Anime' AND type = ?", animeID, imageType)

	if err != nil {
		_, dbErr := db.Exec("INSERT INTO images (path, type, imageable_id, imageable_type, created_at, updated_at) VALUES (?, ?, ?, 'App\\\\Models\\\\Anime', NOW(), NOW())", s3Key, imageType, animeID)
		if dbErr != nil {
			log.Printf("DB Insert failed: %v", dbErr)
		}
	} else {
		_, dbErr := db.Exec("UPDATE images SET path = ?, updated_at = NOW() WHERE id = ?", s3Key, existingImageId)
		if dbErr != nil {
			log.Printf("DB Update failed: %v", dbErr)
		}
	}
}

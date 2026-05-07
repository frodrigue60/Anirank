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
	"path/filepath"
	"strings"
	"time"

	"anirank/api/internal/infrastructure"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Anilist GraphQL Structs
type AnilistResponse struct {
	Data struct {
		Page struct {
			Media []struct {
				ID         int    `json:"id"`
				CoverImage struct {
					ExtraLarge string `json:"extraLarge"`
				} `json:"coverImage"`
				BannerImage string   `json:"bannerImage"`
				Genres      []string `json:"genres"`
				Studios     struct {
					Edges []struct {
						IsMain bool `json:"isMain"`
						Node   struct {
							ID   int    `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"studios"`
			} `json:"media"`
		} `json:"page"`
	} `json:"data"`
}

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

	// 2. Storage Initialization (Dynamic R2/S3)
	storage, err := infrastructure.InitStorageFromEnv(ctx)
	if err != nil {
		log.Fatalf("Unable to initialize storage: %v\n", err)
	}

	log.Println("Starting AniList Image Sync & R2 Migration...")

	// 3. Fetch animes with anilist_id
	rows, err := pool.Query(ctx, "SELECT id, anilist_id, title FROM animes WHERE anilist_id IS NOT NULL")
	if err != nil {
		log.Fatalf("Query error: %v\n", err)
	}
	defer rows.Close()

	type animeRec struct {
		ID        uint64
		AnilistID int64
		Title     string
	}
	var animes []animeRec
	for rows.Next() {
		var a animeRec
		if err := rows.Scan(&a.ID, &a.AnilistID, &a.Title); err != nil {
			continue
		}
		animes = append(animes, a)
	}

	log.Printf("Found %d animes with AniList IDs.\n", len(animes))

	// 4. Batch processing (AniList allows up to 50 per page)
	batchSize := 25
	for i := 0; i < len(animes); i += batchSize {
		end := i + batchSize
		if end > len(animes) {
			end = len(animes)
		}
		batch := animes[i:end]
		
		var ids []int64
		for _, a := range batch {
			ids = append(ids, a.AnilistID)
		}

		log.Printf("Processing batch %d-%d...\n", i+1, end)
		
		// 5. Fetch from AniList
		mediaData, err := fetchAnilistMedia(ids)
		if err != nil {
			log.Printf("Error fetching from AniList for batch: %v\n", err)
			continue
		}

		// 6. Mirror each anime
		for _, a := range batch {
			media, ok := mediaData[a.AnilistID]
			if !ok {
				continue
			}

			updates := make(map[string]string)

			// Cover
			if media.CoverImage.ExtraLarge != "" {
				newCover, err := mirrorImage(ctx, storage, media.CoverImage.ExtraLarge, "animes/covers", a.AnilistID)
				if err == nil {
					updates["cover"] = newCover
				} else {
					log.Printf("[%s] Cover Error: %v\n", a.Title, err)
				}
			}

			// Banner
			if media.BannerImage != "" {
				newBanner, err := mirrorImage(ctx, storage, media.BannerImage, "animes/banners", a.AnilistID)
				if err == nil {
					updates["banner"] = newBanner
				} else {
					log.Printf("[%s] Banner Error: %v\n", a.Title, err)
				}
			}

			// Update DB
			if len(updates) > 0 {
				setClauses := []string{}
				args := []interface{}{}
				for field, val := range updates {
					setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, len(args)+1))
					args = append(args, val)
				}
				args = append(args, a.ID)
				query := fmt.Sprintf("UPDATE animes SET %s, updated_at = NOW() WHERE id = $%d", strings.Join(setClauses, ", "), len(args))
				_, err = pool.Exec(ctx, query, args...)
				if err != nil {
					log.Printf("[%s] DB Update Error: %v\n", a.Title, err)
				} else {
					log.Printf("[%s] Successfully synced images.\n", a.Title)
				}
			}

			// 7. Process Studios and Producers
			tx, err := pool.Begin(ctx)
			if err != nil {
				continue
			}

			for _, edge := range media.Studios.Edges {
				name := edge.Node.Name
				slug := slugify(name)
				
				if edge.IsMain {
					// Studio
					var studioID uint64
					err = tx.QueryRow(ctx, "SELECT id FROM studios WHERE slug = $1", slug).Scan(&studioID)
					if err != nil {
						err = tx.QueryRow(ctx, "INSERT INTO studios (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", name, slug).Scan(&studioID)
					}
					if err == nil {
						_, _ = tx.Exec(ctx, "INSERT INTO anime_studio (anime_id, studio_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", a.ID, studioID)
					}
				} else {
					// Producer
					var producerID uint64
					err = tx.QueryRow(ctx, "SELECT id FROM producers WHERE slug = $1", slug).Scan(&producerID)
					if err != nil {
						err = tx.QueryRow(ctx, "INSERT INTO producers (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", name, slug).Scan(&producerID)
					}
					if err == nil {
						_, _ = tx.Exec(ctx, "INSERT INTO anime_producer (anime_id, producer_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", a.ID, producerID)
					}
				}
			}

			// 8. Process Genres
			for _, gName := range media.Genres {
				gSlug := slugify(gName)
				var genreID uint64
				err = tx.QueryRow(ctx, "SELECT id FROM genres WHERE slug = $1", gSlug).Scan(&genreID)
				if err != nil {
					err = tx.QueryRow(ctx, "INSERT INTO genres (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", gName, gSlug).Scan(&genreID)
				}
				if err == nil {
					_, _ = tx.Exec(ctx, "INSERT INTO anime_genre (anime_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", a.ID, genreID)
				}
			}

			tx.Commit(ctx)
		}

		// Respect rate limits
		if i+batchSize < len(animes) {
			log.Println("Waiting for rate limit safety...")
			time.Sleep(2 * time.Second)
		}
	}

	log.Println("AniList Sync Complete!")
}

func fetchAnilistMedia(ids []int64) (map[int64]struct {
	CoverImage struct {
		ExtraLarge string `json:"extraLarge"`
	} `json:"coverImage"`
	BannerImage string   `json:"bannerImage"`
	Genres      []string `json:"genres"`
	Studios     struct {
		Edges []struct {
			IsMain bool `json:"isMain"`
			Node   struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"studios"`
}, error) {
	query := `
	query ($ids: [Int]) {
		Page(page: 1, perPage: 50) {
			media(id_in: $ids, type: ANIME) {
				id
				coverImage {
					extraLarge
				}
				bannerImage
				genres
				studios {
					edges {
						isMain
						node {
							id
							name
						}
					}
				}
			}
		}
	}`

	body, _ := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": map[string]interface{}{"ids": ids},
	})

	resp, err := http.Post("https://graphql.anilist.co", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var alResp AnilistResponse
	if err := json.NewDecoder(resp.Body).Decode(&alResp); err != nil {
		return nil, err
	}

	results := make(map[int64]struct {
		CoverImage struct {
			ExtraLarge string `json:"extraLarge"`
		} `json:"coverImage"`
		BannerImage string   `json:"bannerImage"`
		Genres      []string `json:"genres"`
		Studios     struct {
			Edges []struct {
				IsMain bool `json:"isMain"`
				Node   struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"studios"`
	})

	for _, m := range alResp.Data.Page.Media {
		results[int64(m.ID)] = struct {
			CoverImage struct {
				ExtraLarge string `json:"extraLarge"`
			} `json:"coverImage"`
			BannerImage string   `json:"bannerImage"`
			Genres      []string `json:"genres"`
			Studios     struct {
				Edges []struct {
					IsMain bool `json:"isMain"`
					Node   struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"studios"`
		}{
			CoverImage:  m.CoverImage,
			BannerImage: m.BannerImage,
			Genres:      m.Genres,
			Studios:     m.Studios,
		}
	}

	return results, nil
}

func mirrorImage(ctx context.Context, storage *infrastructure.S3Storage, sourceURL, prefix string, anilistID int64) (string, error) {
	// Download
	resp, err := http.Get(sourceURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Detect Extension
	ext := filepath.Ext(sourceURL)
	if ext == "" {
		ext = ".jpg"
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	newPath := fmt.Sprintf("%s/%d_%s%s", prefix, anilistID, uuid.New().String()[:8], ext)

	// Upload
	_, err = storage.UploadFile(ctx, newPath, bytes.NewReader(data), int64(len(data)), contentType)
	if err != nil {
		return "", err
	}

	return newPath, nil
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

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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// API Structs
type ATResponse struct {
	Anime []ATAnime `json:"anime"`
}

type ATAnime struct {
	Name        string       `json:"name"`
	Slug        string       `json:"slug"`
	Year        int          `json:"year"`
	Season      string       `json:"season"`
	Images      []ATImage    `json:"images"`
	AnimeThemes []ATTheme    `json:"animethemes"`
	Studios     []ATTaxonomy `json:"studios"`
	Synopsis    string       `json:"synopsis"`
	MediaFormat string       `json:"media_format"`
	Resources   []ATResource `json:"resources"`
}

type ATResource struct {
	Site string `json:"site"`
	Link string `json:"link"`
}

type ATTaxonomy struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ATImage struct {
	Link  string `json:"link"`
	Facet string `json:"facet"`
}

type ATTheme struct {
	Type              string    `json:"type"`
	Sequence          int       `json:"sequence"`
	Song              *ATSong   `json:"song"`
	AnimeThemeEntries []ATEntry `json:"animethemeentries"`
}

type ATEntry struct {
	Version int    `json:"version"`
	Notes   string `json:"notes"`
}

type ATSong struct {
	// ... (rest of structs)
	Title   string     `json:"title"`
	Artists []ATArtist `json:"artists"`
}

type ATArtist struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// AniList Structs
type AniListResponse struct {
	Data struct {
		Media struct {
			ID          int    `json:"id"`
			Title       struct {
				Romaji  string `json:"romaji"`
				English string `json:"english"`
				Native  string `json:"native"`
			} `json:"title"`
			Description string `json:"description"`
			CoverImage  struct {
				ExtraLarge string `json:"extraLarge"`
			} `json:"coverImage"`
			BannerImage string   `json:"bannerImage"`
			Genres      []string `json:"genres"`
			Studios     struct {
				Edges []struct {
					IsMain bool `json:"isMain"`
					Node   struct {
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"studios"`
			ExternalLinks []struct {
				Site string `json:"site"`
				URL  string `json:"url"`
			} `json:"externalLinks"`
		} `json:"Media"`
	} `json:"data"`
}

const anilistMediaQuery = `
query ($id: Int) {
  Media(id: $id, type: ANIME) {
    id
    title { romaji english native }
    description(asHtml: false)
    coverImage { extraLarge }
    bannerImage
    genres
    studios {
      edges {
        isMain
        node { name }
      }
    }
    externalLinks { site url }
  }
}
`

func fetchFromAniList(ctx context.Context, client *http.Client, id int64) (*AniListResponse, error) {
	payload := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     anilistMediaQuery,
		Variables: map[string]interface{}{"id": id},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://graphql.anilist.co", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AniList API returns %d", resp.StatusCode)
	}

	var alResp AniListResponse
	if err := json.NewDecoder(resp.Body).Decode(&alResp); err != nil {
		return nil, err
	}
	return &alResp, nil
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found.")
	}
	ctx := context.Background()

	// DB Setup
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

	year := 2025
	seasonName := "Winter"

	fmt.Printf("Starting Hydration for %s %d...\n", seasonName, year)

	// 1. Fetch from API
	url := fmt.Sprintf("https://api.animethemes.moe/anime?include=animethemes.song.artists,images,animethemes.animethemeentries,studios,resources&filter[year]=%d&filter[season]=%s&page[size]=100", year, strings.ToLower(seasonName))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch from API: %v\n", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v\n", err)
	}

	var atResp ATResponse
	if err := json.Unmarshal(body, &atResp); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v\n", err)
	}

	fmt.Printf("API returned %d animes\n", len(atResp.Anime))

	// 2. Process
	for _, a := range atResp.Anime {
		fmt.Printf("Processing Anime: %s\n", a.Name)

		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Printf("Failed to start transaction for %s: %v\n", a.Name, err)
			continue
		}

		// Upsert Year
		var yearID uint64
		err = tx.QueryRow(ctx, "SELECT id FROM years WHERE name = $1", fmt.Sprintf("%d", a.Year)).Scan(&yearID)
		if err != nil {
			err = tx.QueryRow(ctx, "INSERT INTO years (name, current, created_at, updated_at) VALUES ($1, false, NOW(), NOW()) RETURNING id", fmt.Sprintf("%d", a.Year)).Scan(&yearID)
			if err != nil {
				tx.Rollback(ctx)
				log.Printf("Year Error: %v\n", err)
				continue
			}
		}

		// Upsert Season
		var seasonID uint64
		// Normalize season name to Title Case (e.g., "winter" -> "Winter")
		normalizedSeason := strings.Title(strings.ToLower(a.Season))
		err = tx.QueryRow(ctx, "SELECT id FROM seasons WHERE name = $1", normalizedSeason).Scan(&seasonID)
		if err != nil {
			err = tx.QueryRow(ctx, "INSERT INTO seasons (name, current, created_at, updated_at) VALUES ($1, false, NOW(), NOW()) RETURNING id", normalizedSeason).Scan(&seasonID)
			if err != nil {
				tx.Rollback(ctx)
				log.Printf("Season Error: %v\n", err)
				continue
			}
		}

		// Upsert Format
		var formatID uint64
		if a.MediaFormat == "" {
			a.MediaFormat = "TV" // Default fallback
		}
		formatSlug := strings.ToLower(a.MediaFormat)
		err = tx.QueryRow(ctx, "SELECT id FROM formats WHERE slug = $1", formatSlug).Scan(&formatID)
		if err != nil {
			err = tx.QueryRow(ctx, "INSERT INTO formats (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", a.MediaFormat, formatSlug).Scan(&formatID)
			if err != nil {
				// Non-critical error, just log it
				log.Printf("Format Error: %v\n", err)
			}
		}
		// Find Image
		coverUrl := ""
		for _, img := range a.Images {
			if img.Facet == "Large Cover" {
				coverUrl = img.Link
				break
			}
		}

		// Find Anilist ID
		var anilistID int64
		for _, res := range a.Resources {
			if res.Site == "AniList" {
				parts := strings.Split(strings.TrimRight(res.Link, "/"), "/")
				if len(parts) > 0 {
					fmt.Sscanf(parts[len(parts)-1], "%d", &anilistID)
				}
				break
			}
		}

		// Enforce hyphenated slug for anime
		animeSlug := slugify(a.Slug)

		// 1.5 Fetch from AniList if available
		var alData *AniListResponse
		if anilistID > 0 {
			fmt.Printf("  Fetching AniList data for ID %d...\n", anilistID)
			alData, err = fetchFromAniList(ctx, client, anilistID)
			if err != nil {
				log.Printf("  AniList Error for %d: %v\n", anilistID, err)
			} else {
				// Override AnimeThemes data with AniList data
				if alData.Data.Media.Title.Romaji != "" {
					a.Name = alData.Data.Media.Title.Romaji
				}
				if alData.Data.Media.Description != "" {
					a.Synopsis = alData.Data.Media.Description
				}
				if alData.Data.Media.CoverImage.ExtraLarge != "" {
					coverUrl = alData.Data.Media.CoverImage.ExtraLarge
				}
			}
			// Rate Limit: 10 r/s (100ms)
			time.Sleep(100 * time.Millisecond)
		}

		// Find existing Anime by AniList ID first, then by slug
		var animeID uint64
		found := false
		var errLookup error

		if anilistID > 0 {
			errLookup = tx.QueryRow(ctx, "SELECT id FROM animes WHERE anilist_id = $1", anilistID).Scan(&animeID)
			if errLookup == nil {
				found = true
			}
		}

		if !found {
			errLookup = tx.QueryRow(ctx, "SELECT id FROM animes WHERE slug = $1", animeSlug).Scan(&animeID)
			if errLookup == nil {
				found = true
			}
		}

		// Flush existing association data for this anime if found
		if found {
			// We keep songs and their variants to update them instead of deleting
			// but we clean up pivot tables for taxonomies to ensure a clean refresh of associations
			_, _ = tx.Exec(ctx, "DELETE FROM anime_studio WHERE anime_id = $1", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM anime_producer WHERE anime_id = $1", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM anime_genre WHERE anime_id = $1", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM anime_external_link WHERE anime_id = $1", animeID)

			// Update the anime itself (enforcing new slug and metadata)
			bannerUrl := ""
			if alData != nil {
				bannerUrl = alData.Data.Media.BannerImage
			}

			_, err = tx.Exec(ctx, `
				UPDATE animes 
				SET title = $1, slug = $2, cover = $3, banner = $4, description = $5, format_id = $6, anilist_id = $7, season_id = $8, year_id = $9, updated_at = NOW() 
				WHERE id = $10`, a.Name, animeSlug, coverUrl, bannerUrl, a.Synopsis, formatID, anilistID, seasonID, yearID, animeID)
		} else {
			bannerUrl := ""
			if alData != nil {
				bannerUrl = alData.Data.Media.BannerImage
			}
			// Not found at all, insert new
			animeUUID := uuid.New().String()
			err = tx.QueryRow(ctx, `
				INSERT INTO animes (uuid, title, slug, cover, banner, description, format_id, anilist_id, season_id, year_id, status, created_at, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true, NOW(), NOW()) 
				RETURNING id`, animeUUID, a.Name, animeSlug, coverUrl, bannerUrl, a.Synopsis, formatID, anilistID, seasonID, yearID).Scan(&animeID)
		}
		if err != nil {
			tx.Rollback(ctx)
			log.Printf("Anime Error: %v\n", err)
			continue
		}

		// Process Studios & Producers (from AniList if available, otherwise AT)
		if alData != nil {
			for _, edge := range alData.Data.Media.Studios.Edges {
				var targetID uint64
				sSlug := slugify(edge.Node.Name)
				if edge.IsMain {
					err = tx.QueryRow(ctx, "SELECT id FROM studios WHERE slug = $1", sSlug).Scan(&targetID)
					if err != nil {
						err = tx.QueryRow(ctx, "INSERT INTO studios (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", edge.Node.Name, sSlug).Scan(&targetID)
					}
					if err == nil {
						_, _ = tx.Exec(ctx, "INSERT INTO anime_studio (anime_id, studio_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", animeID, targetID)
					}
				} else {
					err = tx.QueryRow(ctx, "SELECT id FROM producers WHERE slug = $1", sSlug).Scan(&targetID)
					if err != nil {
						err = tx.QueryRow(ctx, "INSERT INTO producers (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", edge.Node.Name, sSlug).Scan(&targetID)
					}
					if err == nil {
						_, _ = tx.Exec(ctx, "INSERT INTO anime_producer (anime_id, producer_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", animeID, targetID)
					}
				}
			}

			// Process Genres (AniList only)
			for _, gName := range alData.Data.Media.Genres {
				var genreID uint64
				gSlug := slugify(gName)
				err = tx.QueryRow(ctx, "SELECT id FROM genres WHERE slug = $1", gSlug).Scan(&genreID)
				if err != nil {
					err = tx.QueryRow(ctx, "INSERT INTO genres (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", gName, gSlug).Scan(&genreID)
				}
				if err == nil {
					_, _ = tx.Exec(ctx, "INSERT INTO anime_genre (anime_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", animeID, genreID)
				}
			}

			// Process External Links (AniList only)
			for _, l := range alData.Data.Media.ExternalLinks {
				var linkID uint64
				// Check if link with same URL already exists
				err = tx.QueryRow(ctx, "SELECT id FROM external_links WHERE url = $1 LIMIT 1", l.URL).Scan(&linkID)
				if err != nil {
					// Not found, create new
					err = tx.QueryRow(ctx, "INSERT INTO external_links (name, url, type, created_at, updated_at) VALUES ($1, $2, 'info', NOW(), NOW()) RETURNING id", l.Site, l.URL).Scan(&linkID)
				} else {
					// Update name just in case it changed
					_, _ = tx.Exec(ctx, "UPDATE external_links SET name = $1, updated_at = NOW() WHERE id = $2", l.Site, linkID)
				}

				if err == nil {
					_, _ = tx.Exec(ctx, "INSERT INTO anime_external_link (anime_id, external_link_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) ON CONFLICT DO NOTHING", animeID, linkID)
				}
			}
		} else {
			// Fallback to AT Studios
			for _, s := range a.Studios {
				var studioID uint64
				sSlug := slugify(s.Name)
				err = tx.QueryRow(ctx, "SELECT id FROM studios WHERE slug = $1", sSlug).Scan(&studioID)
				if err != nil {
					err = tx.QueryRow(ctx, "INSERT INTO studios (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", s.Name, sSlug).Scan(&studioID)
				}
				if err == nil {
					_, _ = tx.Exec(ctx, "INSERT INTO anime_studio (anime_id, studio_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", animeID, studioID)
				}
			}
		}

		// Process Themes
		for _, t := range a.AnimeThemes {
			if t.Song == nil {
				continue
			}

			themeNum := t.Sequence
			if themeNum == 0 {
				themeNum = 1
			}

			songSlug := fmt.Sprintf("%s%d", t.Type, themeNum)

			var songID uint64
			err = tx.QueryRow(ctx, "SELECT id FROM songs WHERE slug = $1 AND anime_id = $2", songSlug, animeID).Scan(&songID)
			if err == nil {
				_, err = tx.Exec(ctx, "UPDATE songs SET song_romaji = $1, season_id = $2, year_id = $3, updated_at = NOW() WHERE id = $4", t.Song.Title, seasonID, yearID, songID)
			} else {
				err = tx.QueryRow(ctx, `
					INSERT INTO songs (song_romaji, song_jp, song_en, slug, type, theme_num, anime_id, season_id, year_id, status, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, true, NOW(), NOW())
					RETURNING id`, t.Song.Title, t.Song.Title, t.Song.Title, songSlug, t.Type, fmt.Sprintf("%d", themeNum), animeID, seasonID, yearID).Scan(&songID)
			}

			if err != nil {
				log.Printf("Song Error (%s): %v\n", songSlug, err)
				continue
			}

			// Process Variants (Entries)
			for _, entry := range t.AnimeThemeEntries {
				version := entry.Version
				if version == 0 {
					version = 1
				}
				variantSlug := fmt.Sprintf("V%d", version)

				var variantID uint64
				err = tx.QueryRow(ctx, "SELECT id FROM song_variants WHERE slug = $1 AND song_id = $2", variantSlug, songID).Scan(&variantID)
				if err != nil {
					err = tx.QueryRow(ctx, `
						INSERT INTO song_variants (version_number, song_id, slug, views, season_id, year_id, status, created_at, updated_at)
						VALUES ($1, $2, $3, 0, $4, $5, true, NOW(), NOW())
						RETURNING id`, version, songID, variantSlug, seasonID, yearID).Scan(&variantID)
					if err != nil {
						log.Printf("Variant Error (%s): %v\n", variantSlug, err)
					}
				}
			}

			// Process Artists
			for _, art := range t.Song.Artists {
				var artistID uint64
				cleanName := strings.TrimSpace(art.Name)
				aSlug := slugify(cleanName)

				// Find existing Artist by name (case-insensitive) OR slug
				err = tx.QueryRow(ctx, "SELECT id FROM artists WHERE LOWER(name) = LOWER($1) OR slug = $2 LIMIT 1", cleanName, aSlug).Scan(&artistID)
				if err == nil {
					// Update existing to ensure name matches Exactly (case-sensitive update) and slug matches
					_, err = tx.Exec(ctx, "UPDATE artists SET name = $1, slug = $2, updated_at = NOW() WHERE id = $3", cleanName, aSlug, artistID)
				} else {
					// Not found, insert new
					err = tx.QueryRow(ctx, `
						INSERT INTO artists (name, slug, status, created_at, updated_at)
						VALUES ($1, $2, true, NOW(), NOW())
						RETURNING id`, cleanName, aSlug).Scan(&artistID)
				}

				if err != nil {
					log.Printf("Artist Error (%s): %v\n", cleanName, err)
					continue
				}

				// Link Artist-Song
				_, err = tx.Exec(ctx, "INSERT INTO artist_song (artist_id, song_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", artistID, songID)
				if err != nil {
					// If the link table also doesn't have a unique constraint, we check first
					var dummy uint64
					errCheck := tx.QueryRow(ctx, "SELECT artist_id FROM artist_song WHERE artist_id = $1 AND song_id = $2", artistID, songID).Scan(&dummy)
					if errCheck != nil {
						_, _ = tx.Exec(ctx, "INSERT INTO artist_song (artist_id, song_id) VALUES ($1, $2)", artistID, songID)
					}
				}
			}
		}

		if err := tx.Commit(ctx); err != nil {
			log.Printf("Commit Error for %s: %v\n", a.Name, err)
		} else {
			fmt.Printf("Successfully hydrated: %s\n", a.Name)
		}
	}

	// Post-hydration: Recalculate all artist song counts from ground truth
	fmt.Println("Recalculating artist song counts...")
	_, err = pool.Exec(ctx, `
		UPDATE artists SET enabled_songs = sub.cnt
		FROM (
			SELECT a.id, COALESCE(COUNT(s.id), 0) AS cnt
			FROM artists a
			LEFT JOIN artist_song asng ON asng.artist_id = a.id
			LEFT JOIN songs s ON s.id = asng.song_id AND s.status = true
			GROUP BY a.id
		) sub
		WHERE artists.id = sub.id`)
	if err != nil {
		log.Printf("Error recalculating enabled_songs: %v\n", err)
	}

	_, err = pool.Exec(ctx, `
		UPDATE artists SET disabled_songs = sub.cnt
		FROM (
			SELECT a.id, COALESCE(COUNT(s.id), 0) AS cnt
			FROM artists a
			LEFT JOIN artist_song asng ON asng.artist_id = a.id
			LEFT JOIN songs s ON s.id = asng.song_id AND s.status = false
			GROUP BY a.id
		) sub
		WHERE artists.id = sub.id`)
	if err != nil {
		log.Printf("Error recalculating disabled_songs: %v\n", err)
	}

	fmt.Println("Hydration Complete!")
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

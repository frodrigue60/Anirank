package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

	year := 2026
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

		// Flush existing data for this anime if found
		if found {
			// We delete songs and their variants to ensure a clean re-hydration
			_, _ = tx.Exec(ctx, "DELETE FROM song_variants WHERE song_id IN (SELECT id FROM songs WHERE anime_id = $1)", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM artist_song WHERE song_id IN (SELECT id FROM songs WHERE anime_id = $1)", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM songs WHERE anime_id = $1", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM anime_studio WHERE anime_id = $1", animeID)

			// Update the anime itself (enforcing new slug and metadata)
			_, err = tx.Exec(ctx, `
				UPDATE animes 
				SET title = $1, slug = $2, cover = $3, description = $4, format_id = $5, anilist_id = $6, season_id = $7, year_id = $8, updated_at = NOW() 
				WHERE id = $9`, a.Name, animeSlug, coverUrl, a.Synopsis, formatID, anilistID, seasonID, yearID, animeID)
		} else {
			// Not found at all, insert new
			err = tx.QueryRow(ctx, `
				INSERT INTO animes (title, slug, cover, description, format_id, anilist_id, season_id, year_id, status, created_at, updated_at) 
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, NOW(), NOW()) 
				RETURNING id`, a.Name, animeSlug, coverUrl, a.Synopsis, formatID, anilistID, seasonID, yearID).Scan(&animeID)
		}
		if err != nil {
			tx.Rollback(ctx)
			log.Printf("Anime Error: %v\n", err)
			continue
		}

		// Process Studios
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
				aSlug := slugify(art.Name)
				err = tx.QueryRow(ctx, "SELECT id FROM artists WHERE slug = $1", aSlug).Scan(&artistID)
				if err == nil {
					_, err = tx.Exec(ctx, "UPDATE artists SET name = $1, updated_at = NOW() WHERE id = $2", art.Name, artistID)
				} else {
					err = tx.QueryRow(ctx, `
						INSERT INTO artists (name, slug, status, created_at, updated_at)
						VALUES ($1, $2, true, NOW(), NOW())
						RETURNING id`, art.Name, aSlug).Scan(&artistID)
				}

				if err != nil {
					log.Printf("Artist Error (%s): %v\n", art.Slug, err)
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

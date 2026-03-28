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
	Title   string     `json:"title"`
	Artists []ATArtist `json:"artists"`
}

type ATArtist struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ATResource struct {
	Site string `json:"site"`
	Link string `json:"link"`
}

type ATArtistData struct {
	ID        uint64       `json:"id"`
	Resources []ATResource `json:"resources"`
}

type ATArtistResponse struct {
	Artists []ATArtistData `json:"artists"`
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("API returned error status %d: %s\n", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v\n", err)
	}

	var atResp ATResponse
	if err := json.Unmarshal(body, &atResp); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v\n", err)
	}

	fmt.Printf("API returned %d animes\n", len(atResp.Anime))

	// 1.5 Batch Fetch Artist Resources
	uniqueArtistIDs := make(map[uint64]bool)
	for _, a := range atResp.Anime {
		for _, theme := range a.AnimeThemes {
			if theme.Song != nil {
				for _, art := range theme.Song.Artists {
					uniqueArtistIDs[art.ID] = true
				}
			}
		}
	}

	artistResourcesMap := make(map[uint64][]ATResource)
	if len(uniqueArtistIDs) > 0 {
		artistIDs := make([]uint64, 0, len(uniqueArtistIDs))
		for id := range uniqueArtistIDs {
			artistIDs = append(artistIDs, id)
		}

		fmt.Printf("Fetching resources for %d artists in batches of 100...\n", len(artistIDs))
		for i := 0; i < len(artistIDs); i += 100 {
			end := i + 100
			if end > len(artistIDs) {
				end = len(artistIDs)
			}
			chunk := artistIDs[i:end]
			idStrings := make([]string, len(chunk))
			for j, id := range chunk {
				idStrings[j] = fmt.Sprintf("%d", id)
			}

			artistUrl := fmt.Sprintf("https://api.animethemes.moe/artist?include=resources&filter[id]=%s&page[size]=100", strings.Join(idStrings, ","))
			aResp, err := client.Get(artistUrl)
			if err != nil {
				log.Printf("Warning: Failed to fetch artists batch: %v\n", err)
				continue
			}
			var artResp ATArtistResponse
			if err := json.NewDecoder(aResp.Body).Decode(&artResp); err == nil {
				for _, artData := range artResp.Artists {
					artistResourcesMap[artData.ID] = artData.Resources
				}
			}
			aResp.Body.Close()
			time.Sleep(500 * time.Millisecond)
		}
	}

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
			a.MediaFormat = "TV"
		}
		formatSlug := strings.ToLower(a.MediaFormat)
		err = tx.QueryRow(ctx, "SELECT id FROM formats WHERE slug = $1", formatSlug).Scan(&formatID)
		if err != nil {
			err = tx.QueryRow(ctx, "INSERT INTO formats (name, slug, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id", a.MediaFormat, formatSlug).Scan(&formatID)
			if err != nil {
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

		animeSlug := slugify(a.Slug)

		// 1.5 Fetch from AniList if available
		var alData *AniListResponse
		if anilistID > 0 {
			fmt.Printf("  Fetching AniList data for ID %d...\n", anilistID)
			alData, err = fetchFromAniList(ctx, client, anilistID)
			if err != nil {
				log.Printf("  AniList Error for %d: %v\n", anilistID, err)
			} else {
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
			time.Sleep(100 * time.Millisecond)
		}

		// Find existing Anime
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

		if found {
			_, _ = tx.Exec(ctx, "DELETE FROM anime_studio WHERE anime_id = $1", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM anime_producer WHERE anime_id = $1", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM anime_genre WHERE anime_id = $1", animeID)
			_, _ = tx.Exec(ctx, "DELETE FROM anime_external_link WHERE anime_id = $1", animeID)

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

		// Studios & Producers
		if alData != nil {
			for _, edge := range alData.Data.Media.Studios.Edges {
				var targetID uint64
				sSlug := slugify(edge.Node.Name)
				if edge.IsMain {
					err = tx.QueryRow(ctx, "SELECT id FROM studios WHERE slug = $1", sSlug).Scan(&targetID)
					if err != nil {
						err = tx.QueryRow(ctx, "INSERT INTO studios (uuid, name, slug, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id", uuid.New().String(), edge.Node.Name, sSlug).Scan(&targetID)
					}
					if err == nil {
						_, _ = tx.Exec(ctx, "INSERT INTO anime_studio (anime_id, studio_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", animeID, targetID)
					}
				} else {
					err = tx.QueryRow(ctx, "SELECT id FROM producers WHERE slug = $1", sSlug).Scan(&targetID)
					if err != nil {
						err = tx.QueryRow(ctx, "INSERT INTO producers (uuid, name, slug, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id", uuid.New().String(), edge.Node.Name, sSlug).Scan(&targetID)
					}
					if err == nil {
						_, _ = tx.Exec(ctx, "INSERT INTO anime_producer (anime_id, producer_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", animeID, targetID)
					}
				}
			}

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

			for _, l := range alData.Data.Media.ExternalLinks {
				var linkID uint64
				err = tx.QueryRow(ctx, "SELECT id FROM external_links WHERE url = $1 LIMIT 1", l.URL).Scan(&linkID)
				if err != nil {
					err = tx.QueryRow(ctx, "INSERT INTO external_links (name, url, type, created_at, updated_at) VALUES ($1, $2, 'info', NOW(), NOW()) RETURNING id", l.Site, l.URL).Scan(&linkID)
				} else {
					_, _ = tx.Exec(ctx, "UPDATE external_links SET name = $1, updated_at = NOW() WHERE id = $2", l.Site, linkID)
				}
				if err == nil {
					_, _ = tx.Exec(ctx, "INSERT INTO anime_external_link (anime_id, external_link_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) ON CONFLICT DO NOTHING", animeID, linkID)
				}
			}
		} else {
			for _, s := range a.Studios {
				var studioID uint64
				sSlug := slugify(s.Name)
				err = tx.QueryRow(ctx, "SELECT id FROM studios WHERE slug = $1", sSlug).Scan(&studioID)
				if err != nil {
					err = tx.QueryRow(ctx, "INSERT INTO studios (uuid, name, slug, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id", uuid.New().String(), s.Name, sSlug).Scan(&studioID)
				}
				if err == nil {
					_, _ = tx.Exec(ctx, "INSERT INTO anime_studio (anime_id, studio_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", animeID, studioID)
				}
			}
		}

		// Themes
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
					INSERT INTO songs (uuid, song_romaji, song_jp, song_en, slug, type, theme_num, anime_id, season_id, year_id, status, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true, NOW(), NOW())
					RETURNING id`, uuid.New().String(), t.Song.Title, t.Song.Title, t.Song.Title, songSlug, t.Type, fmt.Sprintf("%d", themeNum), animeID, seasonID, yearID).Scan(&songID)
			}
			if err != nil {
				log.Printf("Song Error: %v\n", err)
				continue
			}

			// Variants
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
						INSERT INTO song_variants (uuid, version_number, song_id, slug, views, season_id, year_id, status, created_at, updated_at)
						VALUES ($1, $2, $3, $4, 0, $5, $6, true, NOW(), NOW())
						RETURNING id`, uuid.New().String(), version, songID, variantSlug, seasonID, yearID).Scan(&variantID)
				}
			}

			// Artists
			for _, art := range t.Song.Artists {
				var artistID uint64
				cleanName := strings.TrimSpace(art.Name)
				aSlug := slugify(cleanName)

				var alID *uint64
				resources := artistResourcesMap[art.ID]
				for _, res := range resources {
					if res.Site == "AniList" && strings.Contains(res.Link, "anilist.co/staff/") {
						parts := strings.Split(strings.TrimRight(res.Link, "/"), "/")
						if len(parts) > 0 {
							idStr := parts[len(parts)-1]
							var idVal uint64
							if _, err := fmt.Sscanf(idStr, "%d", &idVal); err == nil {
								alID = &idVal
							}
						}
						break
					}
				}

				err = tx.QueryRow(ctx, "SELECT id FROM artists WHERE LOWER(name) = LOWER($1) OR slug = $2 LIMIT 1", cleanName, aSlug).Scan(&artistID)
				if err == nil {
					_, err = tx.Exec(ctx, "UPDATE artists SET name = $1, slug = $2, anilist_id = $3, updated_at = NOW() WHERE id = $4", cleanName, aSlug, alID, artistID)
				} else {
					err = tx.QueryRow(ctx, `
						INSERT INTO artists (uuid, name, slug, anilist_id, status, created_at, updated_at)
						VALUES ($1, $2, $3, $4, true, NOW(), NOW())
						RETURNING id`, uuid.New().String(), cleanName, aSlug, alID).Scan(&artistID)
				}
				if err == nil {
					_, _ = tx.Exec(ctx, "INSERT INTO artist_song (artist_id, song_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", artistID, songID)
				}
			}
		}

		if err := tx.Commit(ctx); err != nil {
			log.Printf("Commit Error for %s: %v\n", a.Name, err)
		} else {
			fmt.Printf("Successfully hydrated: %s\n", a.Name)
		}
	}

	fmt.Println("Recalculating artist song counts...")
	_, _ = pool.Exec(ctx, `
		UPDATE artists SET enabled_songs = sub.cnt
		FROM (
			SELECT a.id, COALESCE(COUNT(s.id), 0) AS cnt
			FROM artists a
			LEFT JOIN artist_song asng ON asng.artist_id = a.id
			LEFT JOIN songs s ON s.id = asng.song_id AND s.status = true
			GROUP BY a.id
		) sub
		WHERE artists.id = sub.id`)

	_, _ = pool.Exec(ctx, `
		UPDATE artists SET disabled_songs = sub.cnt
		FROM (
			SELECT a.id, COALESCE(COUNT(s.id), 0) AS cnt
			FROM artists a
			LEFT JOIN artist_song asng ON asng.artist_id = a.id
			LEFT JOIN songs s ON s.id = asng.song_id AND s.status = false
			GROUP BY a.id
		) sub
		WHERE artists.id = sub.id`)

	fmt.Println("Hydration Complete!")
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s = result.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

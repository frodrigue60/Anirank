package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Staff and response structures
type Staff struct {
	ID   int `json:"id"`
	Name struct {
		Full        string   `json:"full"`
		Native      string   `json:"native"`
		Alternative []string `json:"alternative"`
	} `json:"name"`
	Image struct {
		Large string `json:"large"`
	} `json:"image"`
}

type AniListResponse struct {
	Data struct {
		Page struct {
			Staff []Staff `json:"staff"`
		} `json:"Page"`
	} `json:"data"`
}

type ATArtistResponse struct {
	Artists []struct {
		ID        uint64 `json:"id"`
		Name      string `json:"name"`
		Resources []struct {
			Site string `json:"site"`
			Link string `json:"link"`
		} `json:"resources"`
	} `json:"artists"`
}

func main() {
	// Try to find .env in current or parent dirs
	curr, _ := os.Getwd()
	envPath := filepath.Join(curr, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envPath = filepath.Join(curr, "..", ".env")
	}
	_ = godotenv.Load(envPath)

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Fallback to localhost if we are running manually outside docker
	if os.Getenv("DB_HOST") == "db" || os.Getenv("DB_HOST") == "" {
		dbURL = strings.Replace(dbURL, "@db:", "@localhost:", 1)
		if os.Getenv("DB_HOST") == "" {
			dbURL = strings.Replace(dbURL, "@:", "@localhost:", 1)
		}
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// 1. Fetch artists missing either ID
	rows, err := pool.Query(ctx, "SELECT id, name, animethemes_id, anilist_id FROM artists WHERE animethemes_id IS NULL OR anilist_id IS NULL")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	type Artist struct {
		ID    uint64
		Name  string
		AT_ID *uint64
		AL_ID *uint64
	}

	var targets []Artist
	for rows.Next() {
		var a Artist
		if err := rows.Scan(&a.ID, &a.Name, &a.AT_ID, &a.AL_ID); err != nil {
			log.Fatal(err)
		}
		if a.AT_ID != nil && a.AL_ID != nil {
			continue
		}
		targets = append(targets, a)
	}
	rows.Close()

	fmt.Printf("Starting backfill for %d artists...\n", len(targets))

	for i, a := range targets {
		fmt.Printf("[%d/%d] Processing: %s (ID: %d)...\n", i+1, len(targets), a.Name, a.ID)
		
		updated := false

		// Phase 1: Try to get AniList ID if missing
		if a.AL_ID == nil {
			var staff *Staff
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				staff, err = searchAniList(a.Name)
				if err == nil {
					break
				}
				if strings.Contains(err.Error(), "rate limited") {
					fmt.Printf("  ⚠ Rate limited! Waiting 30s (Attempt %d/3)...\n", attempt+1)
					time.Sleep(30 * time.Second)
					continue
				}
				break
			}

			if err == nil && staff != nil {
				if isArtistMatchConfident(a.Name, *staff) {
					idVal := uint64(staff.ID)
					a.AL_ID = &idVal
					fmt.Printf("  ✓ Found AniList ID: %d\n", staff.ID)
					updated = true
				} else {
					fmt.Printf("  ⚠ Low confidence AniList match: %s\n", staff.Name.Full)
				}
			} else if err != nil {
				fmt.Printf("  ✗ AniList Error: %v\n", err)
			}
			time.Sleep(3 * time.Second) // Respect rate limit
		}

		// Phase 2: Try to get AnimeThemes ID if missing
		if a.AT_ID == nil {
			atID, alIDFromAT, err := searchAnimeThemes(a.Name)
			if err == nil && atID != 0 {
				a.AT_ID = &atID
				fmt.Printf("  ✓ Found AnimeThemes ID: %d\n", atID)
				if a.AL_ID == nil && alIDFromAT != 0 {
					a.AL_ID = &alIDFromAT
					fmt.Printf("  ✓ Found AniList ID from AnimeThemes: %d\n", alIDFromAT)
				}
				updated = true
			} else if err != nil {
				fmt.Printf("  ✗ AnimeThemes Error: %v\n", err)
			}
			time.Sleep(500 * time.Millisecond)
		}

		if updated {
			_, err = pool.Exec(ctx, "UPDATE artists SET animethemes_id = $1, anilist_id = $2, updated_at = NOW() WHERE id = $3", a.AT_ID, a.AL_ID, a.ID)
			if err != nil {
				fmt.Printf("  ✗ Update failed: %v\n", err)
			}
		}
	}

	fmt.Println("Backfill complete!")
}

func searchAniList(name string) (*Staff, error) {
	query := `query ($search: String) { Page(page: 1, perPage: 3) { staff(search: $search) { id name { full native alternative } image { large } } } }`
	vars := map[string]interface{}{"search": name}
	payload := map[string]interface{}{"query": query, "variables": vars}
	body, _ := json.Marshal(payload)

	resp, err := http.Post("https://graphql.anilist.co", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited")
	}

	var alResp AniListResponse
	if err := json.NewDecoder(resp.Body).Decode(&alResp); err != nil {
		return nil, err
	}

	if len(alResp.Data.Page.Staff) > 0 {
		return &alResp.Data.Page.Staff[0], nil
	}
	return nil, nil
}

func searchAnimeThemes(name string) (uint64, uint64, error) {
	url := fmt.Sprintf("https://api.animethemes.moe/artist?include=resources&filter[name]=%s", strings.ReplaceAll(name, " ", "%20"))
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("http %d", resp.StatusCode)
	}

	var atResp ATArtistResponse
	if err := json.NewDecoder(resp.Body).Decode(&atResp); err != nil {
		return 0, 0, err
	}

	if len(atResp.Artists) > 0 {
		art := atResp.Artists[0]
		var alID uint64
		for _, res := range art.Resources {
			if res.Site == "AniList" && strings.Contains(res.Link, "anilist.co/staff/") {
				parts := strings.Split(strings.TrimRight(res.Link, "/"), "/")
				if len(parts) > 0 {
					fmt.Sscanf(parts[len(parts)-1], "%d", &alID)
				}
			}
		}
		return art.ID, alID, nil
	}
	return 0, 0, nil
}

func isArtistMatchConfident(originalName string, staff Staff) bool {
	normOriginal := normalizeName(originalName)
	if normOriginal == "" { return false }
	if calculateSimilarity(normOriginal, normalizeName(staff.Name.Full)) >= 0.85 { return true }
	if staff.Name.Native != "" && calculateSimilarity(normOriginal, normalizeName(staff.Name.Native)) >= 0.85 { return true }
	for _, alt := range staff.Name.Alternative {
		if calculateSimilarity(normOriginal, normalizeName(alt)) >= 0.85 { return true }
	}
	return false
}

func normalizeName(s string) string {
	s = strings.ToLower(s)
	reParen := regexp.MustCompile(`\s*\([^)]*\)`)
	s = reParen.ReplaceAllString(s, "")
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func calculateSimilarity(s1, s2 string) float64 {
	if s1 == "" || s2 == "" { return 0 }
	if s1 == s2 { return 1.0 }
	dist := levenshteinDistance(s1, s2)
	maxLen := len(s1); if len(s2) > maxLen { maxLen = len(s2) }
	return 1.0 - (float64(dist) / float64(maxLen))
}

func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)
	column := make([]int, len1+1)
	for y := 1; y <= len1; y++ { column[y] = y }
	for x := 1; x <= len2; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= len1; y++ {
			oldkey := column[y]
			incr := 0; if r1[y-1] != r2[x-1] { incr = 1 }
			column[y] = minInt(column[y]+1, minInt(column[y-1]+1, lastkey+incr))
			lastkey = oldkey
		}
	}
	return column[len1]
}

func minInt(a, b int) int {
	if a < b { return a }
	return b
}

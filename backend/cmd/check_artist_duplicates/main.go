package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	
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

	fmt.Println("=== Checking for AniList ID Duplicates ===")
	queryAL := `SELECT anilist_id, COUNT(*) as count FROM artists WHERE anilist_id IS NOT NULL GROUP BY anilist_id HAVING COUNT(*) > 1`
	rows, _ := pool.Query(ctx, queryAL)
	for rows.Next() {
		var alID uint64; var count int; rows.Scan(&alID, &count)
		fmt.Printf("- AniList ID %d: %d artists\n", alID, count)
		drows, _ := pool.Query(ctx, "SELECT id, name, slug FROM artists WHERE anilist_id = $1", alID)
		for drows.Next() {
			var id uint64; var name, slug string; drows.Scan(&id, &name, &slug)
			fmt.Printf("    * ID: %d, Name: %s, Slug: %s\n", id, name, slug)
		}
		drows.Close()
	}
	rows.Close()

	fmt.Println("\n=== Checking for Slug Duplicates ===")
	querySlug := `SELECT slug, COUNT(*) as count FROM artists GROUP BY slug HAVING COUNT(*) > 1`
	rows, _ = pool.Query(ctx, querySlug)
	for rows.Next() {
		var slug string; var count int; rows.Scan(&slug, &count)
		fmt.Printf("- Slug '%s': %d records\n", slug, count)
		drows, _ := pool.Query(ctx, "SELECT id, name, anilist_id FROM artists WHERE slug = $1", slug)
		for drows.Next() {
			var id uint64; var name string; var alID *uint64; drows.Scan(&id, &name, &alID)
			alStr := "nil"; if alID != nil { alStr = fmt.Sprintf("%d", *alID) }
			fmt.Printf("    * ID: %d, Name: %s, AL_ID: %s\n", id, name, alStr)
		}
		drows.Close()
	}
	rows.Close()

	fmt.Println("\n=== Checking for Potential Duplicates (Partial Name) ===")
	rows, _ = pool.Query(ctx, "SELECT id, name FROM artists WHERE anilist_id IS NOT NULL")
	type Node struct { ID uint64; Name string }; var withID []Node
	for rows.Next() { var n Node; rows.Scan(&n.ID, &n.Name); withID = append(withID, n) }
	rows.Close()
	rows, _ = pool.Query(ctx, "SELECT id, name FROM artists WHERE anilist_id IS NULL")
	var withoutID []Node
	for rows.Next() { var n Node; rows.Scan(&n.ID, &n.Name); withoutID = append(withoutID, n) }
	rows.Close()

	for _, n1 := range withoutID {
		for _, n2 := range withID {
			if calculateSimilarity(normalizeName(n1.Name), normalizeName(n2.Name)) >= 0.90 {
				fmt.Printf("- Potential Match: '%s' (ID %d) vs '%s' (ID %d)\n", n1.Name, n1.ID, n2.Name, n2.ID)
				break
			}
		}
	}
	fmt.Println("Done.")
}

func normalizeName(s string) string {
	s = strings.ToLower(s); reParen := regexp.MustCompile(`\s*\([^)]*\)`); s = reParen.ReplaceAllString(s, "")
	var result strings.Builder; for _, r := range s { if unicode.IsLetter(r) || unicode.IsDigit(r) { result.WriteRune(r) } }
	return result.String()
}

func calculateSimilarity(s1, s2 string) float64 {
	if s1 == "" || s2 == "" { return 0 }; if s1 == s2 { return 1.0 }
	dist := levenshteinDistance(s1, s2); maxLen := len(s1); if len(s2) > maxLen { maxLen = len(s2) }
	return 1.0 - (float64(dist) / float64(maxLen))
}

func levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2); len1, len2 := len(r1), len(r2)
	column := make([]int, len1+1); for y := 1; y <= len1; y++ { column[y] = y }
	for x := 1; x <= len2; x++ {
		column[0] = x; lastkey := x - 1
		for y := 1; y <= len1; y++ {
			oldkey := column[y]; incr := 0; if r1[y-1] != r2[x-1] { incr = 1 }
			column[y] = minInt(column[y]+1, minInt(column[y-1]+1, lastkey+incr))
			lastkey = oldkey
		}
	}
	return column[len1]
}

func minInt(a, b int) int { if a < b { return a }; return b }

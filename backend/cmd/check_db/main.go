package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found.")
	}
	ctx := context.Background()

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

	// 1. Check Columns
	log.Println("--- Checking Columns ---")
	queryCols := `
		SELECT table_name, column_name 
		FROM information_schema.columns 
		WHERE table_name IN ('studios', 'producers') AND column_name = 'anime_count';
	`
	rows, err := pool.Query(ctx, queryCols)
	if err != nil {
		log.Fatalf("Column Query Error: %v\n", err)
	}
	for rows.Next() {
		var table, col string
		if err := rows.Scan(&table, &col); err == nil {
			log.Printf("Found Column: %s.%s\n", table, col)
		}
	}
	rows.Close()

	// 2. Check Triggers
	log.Println("\n--- Checking Triggers ---")
	queryTriggers := `
		SELECT trigger_name, event_object_table 
		FROM information_schema.triggers 
		WHERE event_object_table IN ('anime_studio', 'anime_producer');
	`
	rows, err = pool.Query(ctx, queryTriggers)
	if err != nil {
		log.Fatalf("Trigger Query Error: %v\n", err)
	}
	for rows.Next() {
		var name, table string
		if err := rows.Scan(&name, &table); err == nil {
			log.Printf("Found Trigger: %s on table %s\n", name, table)
		}
	}
	rows.Close()

	// 3. Check Current Counts
	log.Println("\n--- Checking Current Counts (Top 5) ---")
	queryCounts := `
		SELECT name, anime_count FROM studios WHERE anime_count > 0 ORDER BY anime_count DESC LIMIT 5;
	`
	rows, err = pool.Query(ctx, queryCounts)
	if err != nil {
		log.Println("Studio Count Query Error (maybe no data yet or column missing):", err)
	} else {
		for rows.Next() {
			var name string
			var count int
			if err := rows.Scan(&name, &count); err == nil {
				log.Printf("Studio [%s]: %d animes\n", name, count)
			}
		}
		rows.Close()
	}

	log.Println("\nDatabase Check Complete!")
}

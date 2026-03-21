package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not loaded, using environment variables")
	}

	// Use DB_HOST, DB_USER, etc from .env
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	fmt.Println("Connecting to database...")
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	fmt.Println("--- Database Optimization Audit ---")

	// 1. Check Columns
	fmt.Println("\n--- Denormalized Columns Check ---")
	checkColumn(db, "songs", "average_rating")
	checkColumn(db, "songs", "average_score")
	checkColumn(db, "songs", "favorites_count")
	checkColumn(db, "animes", "songs_count")

	// 2. Check Indexes
	fmt.Println("\n--- Key Indexes Check ---")
	// Checking common names or types
	checkIndexByType(db, "songs", "gin")
	checkIndex(db, "songs_anime_id_idx")

	// 4. List Columns of animes
	fmt.Println("\n--- Columns in animes table ---")
	listColumns(db, "animes")
	
	fmt.Println("\n--- Columns in songs table ---")
	listColumns(db, "songs")
}

func listColumns(db *sqlx.DB, table string) {
	var columns []string
	query := "SELECT column_name FROM information_schema.columns WHERE table_name = $1"
	err := db.Select(&columns, query, table)
	if err != nil {
		fmt.Printf("❌ Error listing columns for %s: %v\n", table, err)
		return
	}
	for _, c := range columns {
		fmt.Printf("- %s\n", c)
	}
}

func checkColumn(db *sqlx.DB, table, column string) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.columns 
			WHERE table_name = $1 AND column_name = $2
		)
	`
	err := db.Get(&exists, query, table, column)
	if err != nil {
		fmt.Printf("❌ Error checking column %s.%s: %v\n", table, column, err)
		return
	}
	if exists {
		fmt.Printf("✅ Column found: %s.%s\n", table, column)
	} else {
		fmt.Printf("⚠️ Column MISSING: %s.%s\n", table, column)
	}
}

func checkIndex(db *sqlx.DB, indexName string) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = $1)"
	err := db.Get(&exists, query, indexName)
	if err != nil {
		fmt.Printf("❌ Error checking index %s: %v\n", indexName, err)
		return
	}
	if exists {
		fmt.Printf("✅ Index found: %s\n", indexName)
	} else {
		fmt.Printf("⚠️ Index MISSING or differently named: %s\n", indexName)
	}
}

func checkIndexByType(db *sqlx.DB, tableName, indexType string) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM pg_indexes 
		WHERE tablename = $1 AND indexdef ILIKE '%' || $2 || '%'
	`
	err := db.Get(&count, query, tableName, indexType)
	if err != nil {
		fmt.Printf("❌ Error checking %s indexes on %s: %v\n", indexType, tableName, err)
		return
	}
	if count > 0 {
		fmt.Printf("✅ %d %s index(es) found on %s\n", count, indexType, tableName)
	} else {
		fmt.Printf("⚠️ No %s indexes found on %s\n", indexType, tableName)
	}
}

func listTriggers(db *sqlx.DB) {
	var triggers []struct {
		TriggerName string `db:"tgname"`
	}
	query := `
		SELECT DISTINCT tgname 
		FROM pg_trigger 
		WHERE tgisinternal = false
	`
	err := db.Select(&triggers, query)
	if err != nil {
		fmt.Printf("❌ Error listing triggers: %v\n", err)
		return
	}
	if len(triggers) > 0 {
		for _, t := range triggers {
			fmt.Printf("✅ Trigger identified: %s\n", t.TriggerName)
		}
	} else {
		fmt.Printf("⚠️ No custom triggers found.\n")
	}
}

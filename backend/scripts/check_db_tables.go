package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../.env")
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	var tables []string
	err = db.Select(&tables, "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public'")
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}

	fmt.Printf("Tables in public schema (%d):\n", len(tables))
	for _, t := range tables {
		fmt.Println("-", t)
	}
}

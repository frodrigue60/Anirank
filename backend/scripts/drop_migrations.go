package main

import (
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

	_, err = db.Exec("DROP TABLE IF EXISTS migrations CASCADE")
	if err != nil {
		log.Fatalf("Failed to drop migrations: %v", err)
	}
}

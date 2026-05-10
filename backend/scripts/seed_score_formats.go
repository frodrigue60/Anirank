package main

import (
	"context"
	"log"
	"os"
	"anirank/api/internal/repository/postgres"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../.env")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not found")
	}

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	seeder := postgres.NewScoreFormatSeeder(db)
	if err := seeder.Seed(context.Background()); err != nil {
		log.Fatalf("Failed to seed: %v", err)
	}

	log.Println("Done!")
}

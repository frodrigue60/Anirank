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

	fmt.Println("Seeding base roles...")

	roles := []struct {
		Name string
		Slug string
	}{
		{"Owner", "owner"},
		{"Admin", "admin"},
		{"Editor", "editor"},
		{"Creator", "creator"},
		{"User", "user"},
	}

	for _, r := range roles {
		_, err := db.Exec(`
			INSERT INTO roles (name, slug, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (slug) DO NOTHING
		`, r.Name, r.Slug)
		if err != nil {
			fmt.Printf("Error seeding role %s: %v\n", r.Slug, err)
		}
	}

	fmt.Println("Base roles seeded successfully.")
}

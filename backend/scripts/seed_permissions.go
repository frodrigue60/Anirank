package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	_ = godotenv.Load(".env")

	dbURL := os.Getenv("DATABASE_URL")
	var db *sqlx.DB
	var err error

	if dbURL != "" {
		// Simple URL connection
		db, err = sqlx.Connect("pgx", dbURL)
	} else {
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		name := os.Getenv("DB_NAME")

		addr := host
		if port != "" && !strings.Contains(host, ":") {
			addr = fmt.Sprintf("%s:%s", host, port)
		}

		dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, pass, addr, name)
		db, err = sqlx.Connect("pgx", dsn)
	}
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	fmt.Println("Connected to database. Seeding permissions...")
	
	// Perform cleanup of redundant/legacy permissions
	_, _ = db.Exec("DELETE FROM permissions WHERE slug = 'taxonomy.manage'")
	_, _ = db.Exec("DELETE FROM permissions WHERE slug IN ('taxonomy.years', 'taxonomy.seasons', 'taxonomy.formats', 'taxonomy.genres')")

	// 1. Define Permissions
	permissions := []struct {
		Name string
		Slug string
		Desc string
	}{
		{"Create Anime", "anime.create", "Allows creating new anime entries"},
		{"Edit Anime", "anime.edit", "Allows editing existing anime and metadata"},
		{"Delete Anime", "anime.delete", "Allows removing anime from the system"},
		{"Create Song", "song.create", "Allows adding new opening/ending themes"},
		{"Edit Song", "song.edit", "Allows editing song data and variants/videos"},
		{"Delete Song", "song.delete", "Allows removing songs from the system"},
		{"Create Artist", "artist.create", "Allows adding new artists"},
		{"Edit Artist", "artist.edit", "Allows editing artist profiles"},
		{"Delete Artist", "artist.delete", "Allows removing artists"},
		{"Manage Reports", "reports.manage", "Allows resolving moderation/content reports (songs, comments, users)"},
		{"Manage Users", "users.manage", "Allows listing and editing user profiles"},
		{"Manage Permissions", "permissions.manage", "Allows modifying role-permission mappings"},
		{"Manage Tournaments", "tournament.manage", "Allows creating, seeding and managing tournaments"},
		{"Manage Announcements", "announcement.manage", "Allows creating and managing platform announcements"},
		{"Manage Badges", "badge.manage", "Allows creating and managing system badges"},

		// Granular Taxonomies - Years
		{"Create Years", "taxonomy.years.create", "Allows adding new years"},
		{"Edit Years", "taxonomy.years.edit", "Allows editing existing years"},
		{"Delete Years", "taxonomy.years.delete", "Allows removing years"},

		// Granular Taxonomies - Seasons
		{"Create Seasons", "taxonomy.seasons.create", "Allows adding new seasons"},
		{"Edit Seasons", "taxonomy.seasons.edit", "Allows editing existing seasons"},
		{"Delete Seasons", "taxonomy.seasons.delete", "Allows removing seasons"},

		// Granular Taxonomies - Formats
		{"Create Formats", "taxonomy.formats.create", "Allows adding new anime formats"},
		{"Edit Formats", "taxonomy.formats.edit", "Allows editing existing formats"},
		{"Delete Formats", "taxonomy.formats.delete", "Allows removing formats"},

		// Granular Taxonomies - Genres
		{"Create Genres", "taxonomy.genres.create", "Allows adding new genres"},
		{"Edit Genres", "taxonomy.genres.edit", "Allows editing existing genres"},
		{"Delete Genres", "taxonomy.genres.delete", "Allows removing genres"},

		// Granular Taxonomies - Studios
		{"Create Studios", "taxonomy.studios.create", "Allows adding new studios"},
		{"Edit Studios", "taxonomy.studios.edit", "Allows editing existing studios"},
		{"Delete Studios", "taxonomy.studios.delete", "Allows removing studios"},

		// Granular Taxonomies - Producers
		{"Create Producers", "taxonomy.producers.create", "Allows adding new producers"},
		{"Edit Producers", "taxonomy.producers.edit", "Allows editing existing producers"},
		{"Delete Producers", "taxonomy.producers.delete", "Allows removing producers"},
	}

	for _, p := range permissions {
		_, err := db.Exec(`
			INSERT INTO permissions (name, slug, description, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			ON CONFLICT (slug) DO UPDATE SET name = $1, description = $3, updated_at = NOW()
		`, p.Name, p.Slug, p.Desc)
		if err != nil {
			fmt.Printf("Error seeding permission %s: %v\n", p.Slug, err)
		}
	}

	roleMappings := map[string][]string{
		"admin": {
			"anime.create", "anime.edit", "anime.delete",
			"song.create", "song.edit", "song.delete",
			"artist.create", "artist.edit", "artist.delete",
			"reports.manage", "users.manage", "permissions.manage",
			"tournament.manage", "announcement.manage",
			"badge.manage",
			"taxonomy.years.create", "taxonomy.years.edit", "taxonomy.years.delete",
			"taxonomy.seasons.create", "taxonomy.seasons.edit", "taxonomy.seasons.delete",
			"taxonomy.formats.create", "taxonomy.formats.edit", "taxonomy.formats.delete",
			"taxonomy.genres.create", "taxonomy.genres.edit", "taxonomy.genres.delete",
			"taxonomy.studios.create", "taxonomy.studios.edit", "taxonomy.studios.delete",
			"taxonomy.producers.create", "taxonomy.producers.edit", "taxonomy.producers.delete",
		},
		"editor": {
			"anime.create", "anime.edit",
			"song.create", "song.edit",
			"artist.create", "artist.edit",
			"reports.manage",
			"tournament.manage", "announcement.manage",
			"taxonomy.years.create", "taxonomy.years.edit",
			"taxonomy.seasons.create", "taxonomy.seasons.edit",
			"taxonomy.formats.create", "taxonomy.formats.edit",
			"taxonomy.genres.create", "taxonomy.genres.edit",
			"taxonomy.studios.create", "taxonomy.studios.edit",
			"taxonomy.producers.create", "taxonomy.producers.edit",
		},
		"creator": {
			"anime.create", "song.create", "artist.create",
			"taxonomy.years.create",
			"taxonomy.seasons.create",
			"taxonomy.formats.create",
			"taxonomy.genres.create",
			"taxonomy.studios.create",
			"taxonomy.producers.create",
		},
	}

	for roleSlug, perms := range roleMappings {
		var roleID uint64
		err := db.Get(&roleID, "SELECT id FROM roles WHERE slug = $1", roleSlug)
		if err != nil {
			fmt.Printf("Role %s not found in DB, skipping...\n", roleSlug)
			continue
		}

		// Clear existing (optional, but cleaner for a seeder)
		// _, _ = db.Exec("DELETE FROM role_permissions WHERE role_id = $1", roleID)

		for _, permSlug := range perms {
			var permID uint64
			err := db.Get(&permID, "SELECT id FROM permissions WHERE slug = $1", permSlug)
			if err != nil {
				continue
			}

			_, err = db.Exec(`
				INSERT INTO role_permissions (role_id, permission_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, roleID, permID)
			if err != nil {
				fmt.Printf("Error mapping role %s to permission %s: %v\n", roleSlug, permSlug, err)
			}
		}
	}

	fmt.Println("Seeding completed successfully!")
}

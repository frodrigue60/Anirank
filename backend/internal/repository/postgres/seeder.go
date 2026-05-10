package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type ScoreFormatSeeder struct {
	db *sqlx.DB
}

func NewScoreFormatSeeder(db *sqlx.DB) *ScoreFormatSeeder {
	return &ScoreFormatSeeder{db: db}
}

func (s *ScoreFormatSeeder) Seed(ctx context.Context) error {
	// Only run if data doesn't exist
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM score_formats")
	if err == nil && count > 0 {
		return nil
	}

	log.Println("🌱 Seeding score formats (initial data missing)...")

	formats := []struct {
		Name        string
		Slug        string
		Description string
	}{
		{
			Name:        "100 Point",
			Slug:        "POINT_100",
			Description: "Standard 100-point scale (e.g., 85, 92)",
		},
		{
			Name:        "10 Point Decimal",
			Slug:        "POINT_10_DECIMAL",
			Description: "10-point scale with one decimal point (e.g., 7.5, 8.2)",
		},
		{
			Name:        "10 Point Integer",
			Slug:        "POINT_10",
			Description: "Simple 10-point scale without decimals (e.g., 7, 8)",
		},
		{
			Name:        "5 Point",
			Slug:        "POINT_5",
			Description: "5-point rating system",
		},
	}

	// Prepare valid slugs for exclusion
	validSlugs := []string{}
	for _, f := range formats {
		validSlugs = append(validSlugs, f.Slug)
	}

	// Delete formats not in our valid list
	query, args, _ := sqlx.In("DELETE FROM score_formats WHERE slug NOT IN (?)", validSlugs)
	query = s.db.Rebind(query)
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("Warning: failed to cleanup old score formats: %v", err)
	}

	for _, f := range formats {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO score_formats (name, slug, description, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			ON CONFLICT (slug) DO UPDATE SET 
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				updated_at = NOW()
		`, f.Name, f.Slug, f.Description)
		if err != nil {
			return fmt.Errorf("failed to seed score format %s: %v", f.Slug, err)
		}
	}

	log.Println("✅ Score formats seeded successfully")
	return nil
}

type SongTypeSeeder struct {
	db *sqlx.DB
}

func NewSongTypeSeeder(db *sqlx.DB) *SongTypeSeeder {
	return &SongTypeSeeder{db: db}
}

func (s *SongTypeSeeder) Seed(ctx context.Context) error {
	// Only run if data doesn't exist
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM song_types")
	if err == nil && count > 0 {
		return nil
	}

	log.Println("🌱 Seeding song types (initial data missing)...")

	types := []struct {
		Name        string
		Slug        string
		Description string
	}{
		{
			Name:        "Opening",
			Slug:        "OP",
			Description: "Anime Opening Theme",
		},
		{
			Name:        "Ending",
			Slug:        "ED",
			Description: "Anime Ending Theme",
		},
		{
			Name:        "Insert Song",
			Slug:        "INS",
			Description: "Songs played during anime episodes",
		},
		{
			Name:        "Soundtrack",
			Slug:        "OST",
			Description: "Original Soundtrack pieces",
		},
	}

	// Prepare valid slugs for exclusion
	validSlugs := []string{}
	for _, t := range types {
		validSlugs = append(validSlugs, t.Slug)
	}

	// Delete types not in our valid list
	query, args, _ := sqlx.In("DELETE FROM song_types WHERE slug NOT IN (?)", validSlugs)
	query = s.db.Rebind(query)
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("Warning: failed to cleanup old song types: %v", err)
	}

	for _, t := range types {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO song_types (uuid, name, slug, description, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, NOW())
			ON CONFLICT (slug) DO UPDATE SET 
				name = EXCLUDED.name,
				description = EXCLUDED.description
		`, t.Name, t.Slug, t.Description)
		if err != nil {
			return fmt.Errorf("failed to seed song type %s: %v", t.Slug, err)
		}
	}

	log.Println("✅ Song types seeded successfully")
	return nil
}

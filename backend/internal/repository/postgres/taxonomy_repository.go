package postgres

import (
	"context"
	"fmt"
	"log"
	"strings"

	"anirank/api/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type taxonomyRepository struct {
	db *sqlx.DB
}

func NewTaxonomyRepository(db *sqlx.DB) domain.TaxonomyRepository {
	return &taxonomyRepository{db: db.Unsafe()}
}

func (r *taxonomyRepository) GetAllYears(ctx context.Context) ([]domain.Year, error) {
	var years []domain.Year
	query := "SELECT id, name, slug, current, created_at, updated_at FROM years ORDER BY name DESC"
	err := r.db.SelectContext(ctx, &years, query)
	return years, err
}

func (r *taxonomyRepository) GetAllSeasons(ctx context.Context) ([]domain.Season, error) {
	var seasons []domain.Season
	query := "SELECT id, name, slug, current, created_at, updated_at FROM seasons ORDER BY id ASC"
	err := r.db.SelectContext(ctx, &seasons, query)
	return seasons, err
}

func (r *taxonomyRepository) GetAllFormats(ctx context.Context) ([]domain.Format, error) {
	var formats []domain.Format
	query := "SELECT id, name, slug, created_at, updated_at FROM formats ORDER BY name ASC"
	err := r.db.SelectContext(ctx, &formats, query)
	return formats, err
}

func (r *taxonomyRepository) GetAllGenres(ctx context.Context) ([]domain.Genre, error) {
	var genres []domain.Genre
	query := "SELECT id, name, slug, created_at, updated_at FROM genres ORDER BY name ASC"
	err := r.db.SelectContext(ctx, &genres, query)
	return genres, err
}

func (r *taxonomyRepository) GetAllStudios(ctx context.Context) ([]domain.Studio, error) {
	var studios []domain.Studio
	query := "SELECT id, uuid, name, slug, created_at, updated_at FROM studios ORDER BY name ASC" // Needs limits/search for large dataset
	err := r.db.SelectContext(ctx, &studios, query)
	return studios, err
}

func (r *taxonomyRepository) GetAllProducers(ctx context.Context) ([]domain.Producer, error) {
	var producers []domain.Producer
	query := "SELECT id, uuid, name, slug, created_at, updated_at FROM producers ORDER BY name ASC"
	err := r.db.SelectContext(ctx, &producers, query)
	return producers, err
}

func (r *taxonomyRepository) SearchStudios(ctx context.Context, term string, limit int) ([]domain.Studio, error) {
	var studios []domain.Studio
	query := "SELECT id, uuid, name, slug, created_at, updated_at FROM studios WHERE name ILIKE $1 ORDER BY name ASC LIMIT $2"
	err := r.db.SelectContext(ctx, &studios, query, "%"+term+"%", limit)
	return studios, err
}

func (r *taxonomyRepository) SearchProducers(ctx context.Context, term string, limit int) ([]domain.Producer, error) {
	var producers []domain.Producer
	query := "SELECT id, uuid, name, slug, created_at, updated_at FROM producers WHERE name ILIKE $1 ORDER BY name ASC LIMIT $2"
	err := r.db.SelectContext(ctx, &producers, query, "%"+term+"%", limit)
	return producers, err
}

func (r *taxonomyRepository) SearchGenres(ctx context.Context, term string, limit int) ([]domain.Genre, error) {
	var genres []domain.Genre
	query := "SELECT id, name, slug, created_at, updated_at FROM genres WHERE name ILIKE $1 ORDER BY name ASC LIMIT $2"
	err := r.db.SelectContext(ctx, &genres, query, "%"+term+"%", limit)
	return genres, err
}

// ---- Catalog queries ----

func (r *taxonomyRepository) GetPaginatedStudios(ctx context.Context, limit, offset int, filters domain.StudioFilters) ([]domain.Studio, error) {
	var studios []domain.Studio

	// Subquery for latest banner
	bannerSubquery := "(SELECT a.banner FROM animes a JOIN anime_studio asu ON a.id = asu.anime_id WHERE asu.studio_id = s.id AND a.status = true ORDER BY a.year_id DESC, a.season_id DESC LIMIT 1)"

	query := "SELECT id, uuid, name, slug, logo, anime_count, " + bannerSubquery + " as latest_banner FROM studios s WHERE anime_count > 0"
	var args []interface{}

	if filters.Search != "" {
		query += " AND name ILIKE $1"
		args = append(args, "%"+filters.Search+"%")
	}

	// Sorting
	switch filters.Sort {
	case "name_desc":
		query += " ORDER BY name DESC"
	case "most_series":
		query += " ORDER BY anime_count DESC, name ASC"
	case "least_series":
		query += " ORDER BY anime_count ASC, name ASC"
	default:
		query += " ORDER BY name ASC"
	}

	argCount := len(args)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &studios, query, args...)
	if err != nil {
		fmt.Printf("GetPaginatedStudios Error: %v\nQuery: %s\nArgs: %v\n", err, query, args)
		return nil, err
	}

	if len(studios) > 0 {
		log.Printf("[DEBUG-Repo] First Studio: %s, AnimeCount: %d\n", studios[0].Name, studios[0].AnimeCount)
	}

	if studios == nil {
		studios = []domain.Studio{}
	}

	return studios, nil
}

func (r *taxonomyRepository) GetStudioBySlug(ctx context.Context, slug string) (*domain.Studio, error) {
	var s domain.Studio
	query := "SELECT id, uuid, name, slug, logo, anime_count, created_at, updated_at FROM studios WHERE slug = $1"
	err := r.db.GetContext(ctx, &s, query, slug)
	if err != nil {
		fmt.Printf("[CRITICAL-DEBUG] GetStudioBySlug failed for slug %s: %v\n", slug, err)
		return nil, err
	}
	return &s, nil
}

func (r *taxonomyRepository) GetAnimesByStudioID(ctx context.Context, studioID uint64, limit, offset int) ([]domain.Anime, error) {
	var animes []domain.Anime
	query := `
		SELECT 
			a.id, a.uuid, a.title, a.slug, a.description, a.anilist_id, a.status, a.cover, a.banner, a.year_id, a.season_id, a.format_id, a.created_at, a.updated_at,
			f.id as "format.id", f.name as "format.name",
			y.id as "year.id", y.name as "year.name",
			s.id as "season.id", s.name as "season.name"
		FROM animes a
		LEFT JOIN formats f ON a.format_id = f.id
		LEFT JOIN years y ON a.year_id = y.id
		LEFT JOIN seasons s ON a.season_id = s.id
		JOIN anime_studio asu ON a.id = asu.anime_id
		WHERE asu.studio_id = $1 AND a.status = true
		ORDER BY a.year_id DESC, a.season_id DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &animes, query, studioID, limit, offset)
	if animes == nil {
		animes = []domain.Anime{}
	}
	return animes, err
}

func (r *taxonomyRepository) CountAnimesByStudioID(ctx context.Context, studioID uint64) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM animes a
		JOIN anime_studio asu ON a.id = asu.anime_id
		WHERE asu.studio_id = $1 AND a.status = true
	`
	err := r.db.GetContext(ctx, &count, query, studioID)
	return count, err
}

func (r *taxonomyRepository) GetPaginatedProducers(ctx context.Context, limit, offset int, filters domain.ProducerFilters) ([]domain.Producer, error) {
	var producers []domain.Producer

	// Subquery for latest banner (matching Studios pattern)
	bannerSubquery := "(SELECT a.banner FROM animes a JOIN anime_producer apu ON a.id = apu.anime_id WHERE apu.producer_id = p.id AND a.status = true ORDER BY a.year_id DESC, a.season_id DESC LIMIT 1)"

	query := "SELECT id, uuid, name, slug, logo, anime_count, " + bannerSubquery + " as latest_banner FROM producers p WHERE anime_count > 0"
	var args []interface{}

	if filters.Search != "" {
		query += " AND name ILIKE $1"
		args = append(args, "%"+filters.Search+"%")
	}

	// Sorting
	switch filters.Sort {
	case "name_desc":
		query += " ORDER BY name DESC"
	case "most_series":
		query += " ORDER BY anime_count DESC, name ASC"
	case "least_series":
		query += " ORDER BY anime_count ASC, name ASC"
	default:
		query += " ORDER BY name ASC"
	}

	argCount := len(args)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &producers, query, args...)
	if producers == nil {
		producers = []domain.Producer{}
	}
	return producers, err
}

func (r *taxonomyRepository) GetProducerBySlug(ctx context.Context, slug string) (*domain.Producer, error) {
	var p domain.Producer
	query := "SELECT id, uuid, name, slug, logo, anime_count, created_at, updated_at FROM producers WHERE slug = $1"
	err := r.db.GetContext(ctx, &p, query, slug)
	if err != nil {
		fmt.Printf("[CRITICAL-DEBUG] GetProducerBySlug failed for slug %s: %v\n", slug, err)
		return nil, err
	}
	return &p, nil
}

func (r *taxonomyRepository) GetAnimesByProducerID(ctx context.Context, producerID uint64, limit, offset int) ([]domain.Anime, error) {
	var animes []domain.Anime
	query := `
		SELECT 
			a.id, a.uuid, a.title, a.slug, a.description, a.anilist_id, a.status, a.cover, a.banner, a.year_id, a.season_id, a.format_id, a.created_at, a.updated_at,
			f.id as "format.id", f.name as "format.name",
			y.id as "year.id", y.name as "year.name",
			s.id as "season.id", s.name as "season.name"
		FROM animes a
		LEFT JOIN formats f ON a.format_id = f.id
		LEFT JOIN years y ON a.year_id = y.id
		LEFT JOIN seasons s ON a.season_id = s.id
		JOIN anime_producer apu ON a.id = apu.anime_id
		WHERE apu.producer_id = $1 AND a.status = true
		ORDER BY a.year_id DESC, a.season_id DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &animes, query, producerID, limit, offset)
	if animes == nil {
		animes = []domain.Anime{}
	}
	return animes, err
}

func (r *taxonomyRepository) CountAnimesByProducerID(ctx context.Context, producerID uint64) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM animes a
		JOIN anime_producer ap ON a.id = ap.anime_id
		WHERE ap.producer_id = $1 AND a.status = true
	`
	err := r.db.GetContext(ctx, &count, query, producerID)
	return count, err
}

func (r *taxonomyRepository) CountStudios(ctx context.Context, filters domain.StudioFilters) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM studios s WHERE anime_count > 0"
	var args []interface{}

	if filters.Search != "" {
		query += " AND name ILIKE $1"
		args = append(args, "%"+filters.Search+"%")
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

func (r *taxonomyRepository) CountProducers(ctx context.Context, filters domain.ProducerFilters) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM producers p WHERE anime_count > 0"
	var args []interface{}

	if filters.Search != "" {
		query += " AND name ILIKE $1"
		args = append(args, "%"+filters.Search+"%")
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

func (r *taxonomyRepository) GetCurrentYear(ctx context.Context) (*domain.Year, error) {
	var y domain.Year
	query := "SELECT id, name, slug, current, created_at, updated_at FROM years WHERE current = true LIMIT 1"
	err := r.db.GetContext(ctx, &y, query)
	if err != nil {
		return nil, err
	}
	return &y, nil
}

func (r *taxonomyRepository) GetCurrentSeason(ctx context.Context) (*domain.Season, error) {
	var s domain.Season
	query := "SELECT id, name, slug, current, created_at, updated_at FROM seasons WHERE current = true LIMIT 1"
	err := r.db.GetContext(ctx, &s, query)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func makeSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func (r *taxonomyRepository) GetOrCreateYear(ctx context.Context, name string) (*domain.Year, error) {
	var y domain.Year
	err := r.db.GetContext(ctx, &y, "SELECT id, name, slug, current, created_at, updated_at FROM years WHERE name = $1 LIMIT 1", name)
	if err == nil {
		return &y, nil
	}

	slug := makeSlug(name)
	query := "INSERT INTO years (name, slug, current, created_at, updated_at) VALUES ($1, $2, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err = r.db.QueryRowContext(ctx, query, name, slug).Scan(&y.ID)
	if err != nil {
		return nil, err
	}
	y.Name = name
	y.Slug = slug
	return &y, nil
}

func (r *taxonomyRepository) GetOrCreateSeason(ctx context.Context, name string) (*domain.Season, error) {
	name = strings.Title(strings.ToLower(name))
	var s domain.Season
	err := r.db.GetContext(ctx, &s, "SELECT id, name, slug, current, created_at, updated_at FROM seasons WHERE name = $1 LIMIT 1", name)
	if err == nil {
		return &s, nil
	}

	slug := makeSlug(name)
	query := "INSERT INTO seasons (name, slug, current, created_at, updated_at) VALUES ($1, $2, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err = r.db.QueryRowContext(ctx, query, name, slug).Scan(&s.ID)
	if err != nil {
		return nil, err
	}
	s.Name = name
	s.Slug = slug
	return &s, nil
}

func (r *taxonomyRepository) GetOrCreateFormat(ctx context.Context, name string) (*domain.Format, error) {
	var f domain.Format
	err := r.db.GetContext(ctx, &f, "SELECT id, name, slug, created_at, updated_at FROM formats WHERE name = $1 LIMIT 1", name)
	if err == nil {
		return &f, nil
	}

	slug := makeSlug(name)
	query := "INSERT INTO formats (name, slug, created_at, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err = r.db.QueryRowContext(ctx, query, name, slug).Scan(&f.ID)
	if err != nil {
		return nil, err
	}
	f.Name = name
	f.Slug = slug
	return &f, nil
}

func (r *taxonomyRepository) GetOrCreateGenre(ctx context.Context, name string) (*domain.Genre, error) {
	var g domain.Genre
	err := r.db.GetContext(ctx, &g, "SELECT id, name, slug, created_at, updated_at FROM genres WHERE name = $1 LIMIT 1", name)
	if err == nil {
		return &g, nil
	}

	slug := makeSlug(name)
	query := "INSERT INTO genres (name, slug, created_at, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err = r.db.QueryRowContext(ctx, query, name, slug).Scan(&g.ID)
	if err != nil {
		return nil, err
	}
	g.Name = name
	g.Slug = slug
	return &g, nil
}

func (r *taxonomyRepository) GetOrCreateStudio(ctx context.Context, name string) (*domain.Studio, error) {
	var s domain.Studio
	err := r.db.GetContext(ctx, &s, "SELECT id, uuid, name, slug, created_at, updated_at FROM studios WHERE name = $1 LIMIT 1", name)
	if err == nil {
		return &s, nil
	}

	slug := makeSlug(name)
	uid := uuid.New().String()
	query := "INSERT INTO studios (uuid, name, slug, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err = r.db.QueryRowContext(ctx, query, uid, name, slug).Scan(&s.ID)
	if err != nil {
		return nil, err
	}
	s.UUID = uid
	s.Name = name
	s.Slug = slug
	return &s, nil
}

func (r *taxonomyRepository) GetOrCreateProducer(ctx context.Context, name string) (*domain.Producer, error) {
	var p domain.Producer
	err := r.db.GetContext(ctx, &p, "SELECT id, uuid, name, slug, created_at, updated_at FROM producers WHERE name = $1 LIMIT 1", name)
	if err == nil {
		return &p, nil
	}

	slug := makeSlug(name)
	uid := uuid.New().String()
	query := "INSERT INTO producers (uuid, name, slug, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err = r.db.QueryRowContext(ctx, query, uid, name, slug).Scan(&p.ID)
	if err != nil {
		return nil, err
	}
	p.UUID = uid
	p.Name = name
	p.Slug = slug
	return &p, nil
}

func (r *taxonomyRepository) GetYearByID(ctx context.Context, id uint64) (*domain.Year, error) {
	var y domain.Year
	query := "SELECT id, name, slug, current, created_at, updated_at FROM years WHERE id = $1"
	err := r.db.GetContext(ctx, &y, query, id)
	if err != nil {
		return nil, err
	}
	return &y, nil
}

// Admin CRUD
func (r *taxonomyRepository) CreateYear(ctx context.Context, year *domain.Year) error {
	if year.Current {
		_, _ = r.db.ExecContext(ctx, "UPDATE years SET current = false")
	}
	if year.Slug == "" {
		year.Slug = makeSlug(year.Name)
	}
	query := "INSERT INTO years (name, slug, current, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, year.Name, year.Slug, year.Current).Scan(&year.ID)
	return err
}

func (r *taxonomyRepository) UpdateYear(ctx context.Context, year *domain.Year) error {
	if year.Current {
		_, _ = r.db.ExecContext(ctx, "UPDATE years SET current = false WHERE id != $1", year.ID)
	}
	if year.Slug == "" {
		year.Slug = makeSlug(year.Name)
	}
	_, err := r.db.ExecContext(ctx, "UPDATE years SET name = $1, slug = $2, current = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4", year.Name, year.Slug, year.Current, year.ID)
	return err
}

func (r *taxonomyRepository) ToggleYearCurrent(ctx context.Context, id uint64) error {
	var current bool
	err := r.db.GetContext(ctx, &current, "SELECT current FROM years WHERE id = $1", id)
	if err != nil {
		return err
	}

	newCurrent := !current

	if newCurrent {
		_, _ = r.db.ExecContext(ctx, "UPDATE years SET current = false WHERE id != $1", id)
	}
	_, err = r.db.ExecContext(ctx, "UPDATE years SET current = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", newCurrent, id)
	return err
}

func (r *taxonomyRepository) DeleteYear(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM years WHERE id = $1", id)
	return err
}

func (r *taxonomyRepository) GetSeasonByID(ctx context.Context, id uint64) (*domain.Season, error) {
	var s domain.Season
	query := "SELECT id, name, slug, current, created_at, updated_at FROM seasons WHERE id = $1"
	err := r.db.GetContext(ctx, &s, query, id)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *taxonomyRepository) CreateSeason(ctx context.Context, season *domain.Season) error {
	if season.Current {
		_, _ = r.db.ExecContext(ctx, "UPDATE seasons SET current = false")
	}
	if season.Slug == "" {
		season.Slug = makeSlug(season.Name)
	}
	query := "INSERT INTO seasons (name, slug, current, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, season.Name, season.Slug, season.Current).Scan(&season.ID)
	return err
}

func (r *taxonomyRepository) UpdateSeason(ctx context.Context, season *domain.Season) error {
	if season.Current {
		_, _ = r.db.ExecContext(ctx, "UPDATE seasons SET current = false WHERE id != $1", season.ID)
	}
	if season.Slug == "" {
		season.Slug = makeSlug(season.Name)
	}
	_, err := r.db.ExecContext(ctx, "UPDATE seasons SET name = $1, slug = $2, current = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4", season.Name, season.Slug, season.Current, season.ID)
	return err
}

func (r *taxonomyRepository) ToggleSeasonCurrent(ctx context.Context, id uint64) error {
	var current bool
	err := r.db.GetContext(ctx, &current, "SELECT current FROM seasons WHERE id = $1", id)
	if err != nil {
		return err
	}

	newCurrent := !current

	if newCurrent {
		_, _ = r.db.ExecContext(ctx, "UPDATE seasons SET current = false WHERE id != $1", id)
	}
	_, err = r.db.ExecContext(ctx, "UPDATE seasons SET current = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", newCurrent, id)
	return err
}

func (r *taxonomyRepository) DeleteSeason(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM seasons WHERE id = $1", id)
	return err
}

func (r *taxonomyRepository) GetFormatByID(ctx context.Context, id uint64) (*domain.Format, error) {
	var f domain.Format
	query := "SELECT id, name, slug, created_at, updated_at FROM formats WHERE id = $1"
	err := r.db.GetContext(ctx, &f, query, id)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *taxonomyRepository) CreateFormat(ctx context.Context, format *domain.Format) error {
	if format.Slug == "" {
		format.Slug = makeSlug(format.Name)
	}
	query := "INSERT INTO formats (name, slug, created_at, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, format.Name, format.Slug).Scan(&format.ID)
	return err
}

func (r *taxonomyRepository) UpdateFormat(ctx context.Context, format *domain.Format) error {
	if format.Slug == "" {
		format.Slug = makeSlug(format.Name)
	}
	_, err := r.db.ExecContext(ctx, "UPDATE formats SET name = $1, slug = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3", format.Name, format.Slug, format.ID)
	return err
}

func (r *taxonomyRepository) DeleteFormat(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM formats WHERE id = $1", id)
	return err
}

func (r *taxonomyRepository) GetGenreByID(ctx context.Context, id uint64) (*domain.Genre, error) {
	var g domain.Genre
	query := "SELECT id, name, slug, created_at, updated_at FROM genres WHERE id = $1"
	err := r.db.GetContext(ctx, &g, query, id)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *taxonomyRepository) CreateGenre(ctx context.Context, genre *domain.Genre) error {
	if genre.Slug == "" {
		genre.Slug = makeSlug(genre.Name)
	}
	query := "INSERT INTO genres (name, slug, created_at, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id"
	err := r.db.QueryRowContext(ctx, query, genre.Name, genre.Slug).Scan(&genre.ID)
	return err
}

func (r *taxonomyRepository) UpdateGenre(ctx context.Context, genre *domain.Genre) error {
	if genre.Slug == "" {
		genre.Slug = makeSlug(genre.Name)
	}
	_, err := r.db.ExecContext(ctx, "UPDATE genres SET name = $1, slug = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3", genre.Name, genre.Slug, genre.ID)
	return err
}

func (r *taxonomyRepository) DeleteGenre(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM genres WHERE id = $1", id)
	return err
}

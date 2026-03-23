package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type artistRepository struct {
	db *sqlx.DB
}

func NewArtistRepository(db *sqlx.DB) domain.ArtistRepository {
	return &artistRepository{db: db}
}

func (r *artistRepository) GetByID(ctx context.Context, id uint64) (*domain.Artist, error) {
	var a domain.Artist
	query := `
		SELECT 
			id, name, name_jp, slug, created_at, updated_at, avatar, status, favorites_count,
			(SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = artists.id) as songs_count,
			(SELECT ani.banner 
			 FROM animes ani
			 JOIN songs s ON s.anime_id = ani.id
			 JOIN artist_song asong ON asong.song_id = s.id
			 WHERE asong.artist_id = artists.id
			 ORDER BY s.created_at DESC
			 LIMIT 1) as banner
		FROM artists
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &a, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *artistRepository) GetBySlug(ctx context.Context, slug string) (*domain.Artist, error) {
	var a domain.Artist
	query := `
		SELECT 
			id, name, name_jp, slug, created_at, updated_at, avatar, status, favorites_count,
			(SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = artists.id) as songs_count,
			(SELECT ani.banner 
			 FROM animes ani
			 JOIN songs s ON s.anime_id = ani.id
			 JOIN artist_song asong ON asong.song_id = s.id
			 WHERE asong.artist_id = artists.id
			 ORDER BY s.created_at DESC
			 LIMIT 1) as banner
		FROM artists
		WHERE slug = $1
	`
	err := r.db.GetContext(ctx, &a, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

// ---- Admin CRUD ----

func (r *artistRepository) Create(ctx context.Context, artist *domain.Artist) error {
	query := `
		INSERT INTO artists (name, name_jp, slug, avatar, status, created_at, updated_at) 
		VALUES (:name, :name_jp, :slug, :avatar, :status, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, artist)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&artist.ID)
	}
	return nil
}

func (r *artistRepository) Update(ctx context.Context, artist *domain.Artist) error {
	query := `
		UPDATE artists 
		SET name = :name, name_jp = :name_jp, slug = :slug, avatar = :avatar, status = :status, updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	res, err := r.db.NamedExecContext(ctx, query, artist)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("artist not found or no changes made")
	}
	return err
}

func (r *artistRepository) UpdateAvatar(ctx context.Context, id uint64, avatar string) error {
	query := "UPDATE artists SET avatar = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	res, err := r.db.ExecContext(ctx, query, avatar, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("artist not found")
	}
	return err
}

func (r *artistRepository) Delete(ctx context.Context, id uint64) error {
	query := "DELETE FROM artists WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("artist not found")
	}
	return err
}

func (r *artistRepository) ToggleStatus(ctx context.Context, id uint64) error {
	query := `UPDATE artists SET status = NOT status, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("artist not found")
	}
	return err
}

// ---- Catalog queries ----

func (r *artistRepository) GetPaginated(ctx context.Context, limit, offset int, filters domain.ArtistFilters) ([]domain.Artist, error) {
	var artists []domain.Artist
	baseQuery := `
		SELECT 
			a.*, 
			(SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = a.id) as songs_count
		FROM artists a
	`

	query := baseQuery
	var args []interface{}

	i := 1
	if filters.Search != "" {
		query += fmt.Sprintf(" WHERE a.name ILIKE $%d", i)
		args = append(args, "%"+filters.Search+"%")
		i++
		if !filters.IsAdmin {
			query += " AND a.status = true"
		}
	} else if !filters.IsAdmin {
		query += " WHERE a.status = true"
	}

	// Sorting
	switch filters.Sort {
	case "most_themes":
		query += " ORDER BY songs_count DESC, a.name ASC"
	case "least_themes":
		query += " ORDER BY songs_count ASC, a.name ASC"
	case "name_desc":
		query += " ORDER BY a.name DESC"
	case "name_asc":
		query += " ORDER BY a.name ASC"
	default:
		query += " ORDER BY a.created_at DESC, a.id DESC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &artists, query, args...)
	if artists == nil {
		artists = []domain.Artist{}
	}
	return artists, err
}

func (r *artistRepository) Count(ctx context.Context, filters domain.ArtistFilters) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM artists"
	var args []interface{}

	if filters.Search != "" {
		query += " WHERE name ILIKE $1"
		args = append(args, "%"+filters.Search+"%")
		if !filters.IsAdmin {
			query += " AND status = true"
		}
	} else if !filters.IsAdmin {
		query += " WHERE status = true"
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}
func (r *artistRepository) CountFavoritesByUserID(ctx context.Context, userID uint64) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM artist_user f
		JOIN artists a ON a.id = f.artist_id
		WHERE f.user_id = $1 AND a.status = true
	`
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *artistRepository) GetFavoritesByUserID(ctx context.Context, userID uint64, limit, offset int) ([]domain.Artist, error) {
	var artists []domain.Artist
	query := `
		SELECT a.*, 
		       (SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = a.id) as songs_count
		FROM artists a
		JOIN artist_user f ON a.id = f.artist_id
		WHERE f.user_id = $1 AND a.status = true
		GROUP BY a.id, f.created_at
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &artists, query, userID, limit, offset)
	if artists == nil {
		artists = []domain.Artist{}
	}
	return artists, err
}

func (r *artistRepository) GetFeatured(ctx context.Context, limit int) ([]domain.Artist, error) {
	var artists []domain.Artist
	query := `
		SELECT 
			id, name, name_jp, slug, created_at, updated_at, avatar, status,
			(SELECT COUNT(*) FROM artist_user f WHERE f.artist_id = artists.id) as favorites_count,
			(SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = artists.id) as songs_count,
			(SELECT ani.banner 
			 FROM animes ani
			 JOIN songs s ON s.anime_id = ani.id
			 JOIN artist_song asong ON asong.song_id = s.id
			 WHERE asong.artist_id = artists.id
			 ORDER BY s.created_at DESC
			 LIMIT 1) as banner
		FROM artists
		WHERE status = true
		ORDER BY favorites_count DESC, songs_count DESC, name ASC
		LIMIT $1
	`

	err := r.db.SelectContext(ctx, &artists, query, limit)
	if artists == nil {
		artists = []domain.Artist{}
	}
	return artists, err
}

func (r *artistRepository) Search(ctx context.Context, term string, limit int) ([]domain.Artist, error) {
	var artists []domain.Artist
	query := `
		SELECT id, name, slug, avatar 
		FROM artists 
		WHERE name ILIKE $1 AND status = true
		ORDER BY name ASC 
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &artists, query, term, limit)
	if artists == nil {
		artists = []domain.Artist{}
	}
	return artists, err
}

func (r *artistRepository) GetPublicSlugs(ctx context.Context) ([]domain.SitemapItem, error) {
	var items []domain.SitemapItem
	query := `SELECT slug as loc, updated_at as lastmod FROM artists WHERE status = true`
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}

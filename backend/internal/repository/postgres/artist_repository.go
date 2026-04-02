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
			id, uuid, name, name_jp, slug, created_at, updated_at, avatar, status, favorites_count, anime_themes_id, anilist_id,
			enabled_songs, disabled_songs,
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

func (r *artistRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Artist, error) {
	var a domain.Artist
	query := `
		SELECT 
			id, uuid, name, name_jp, slug, created_at, updated_at, avatar, status, favorites_count, anime_themes_id, anilist_id,
			enabled_songs, disabled_songs,
			(SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = artists.id) as songs_count,
			(SELECT ani.banner 
			 FROM animes ani
			 JOIN songs s ON s.anime_id = ani.id
			 JOIN artist_song asong ON asong.song_id = s.id
			 WHERE asong.artist_id = artists.id
			 ORDER BY s.created_at DESC
			 LIMIT 1) as banner
		FROM artists
		WHERE uuid = $1
	`
	err := r.db.GetContext(ctx, &a, query, uuid)
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
			id, uuid, name, name_jp, slug, created_at, updated_at, avatar, status, favorites_count, anime_themes_id, anilist_id,
			enabled_songs, disabled_songs,
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

func (r *artistRepository) GetByAnilistID(ctx context.Context, id uint64) (*domain.Artist, error) {
	var a domain.Artist
	query := `
		SELECT 
			id, uuid, name, name_jp, slug, created_at, updated_at, avatar, status, favorites_count, anime_themes_id, anilist_id,
			enabled_songs, disabled_songs,
			(SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = artists.id) as songs_count,
			(SELECT ani.banner 
			 FROM animes ani
			 JOIN songs s ON s.anime_id = ani.id
			 JOIN artist_song asong ON asong.song_id = s.id
			 WHERE asong.artist_id = artists.id
			 ORDER BY s.created_at DESC
			 LIMIT 1) as banner
		FROM artists
		WHERE anilist_id = $1
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

func (r *artistRepository) GetByAnimeThemesID(ctx context.Context, id uint64) (*domain.Artist, error) {
	var a domain.Artist
	query := `
		SELECT 
			id, uuid, name, name_jp, slug, created_at, updated_at, avatar, status, favorites_count, anime_themes_id, anilist_id,
			enabled_songs, disabled_songs,
			(SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = artists.id) as songs_count,
			(SELECT ani.banner 
			 FROM animes ani
			 JOIN songs s ON s.anime_id = ani.id
			 JOIN artist_song asong ON asong.song_id = s.id
			 WHERE asong.artist_id = artists.id
			 ORDER BY s.created_at DESC
			 LIMIT 1) as banner
		FROM artists
		WHERE anime_themes_id = $1
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

// ---- Admin CRUD ----

func (r *artistRepository) Create(ctx context.Context, artist *domain.Artist) error {
	query := `
		INSERT INTO artists (uuid, name, name_jp, slug, avatar, status, anime_themes_id, anilist_id, created_at, updated_at) 
		VALUES (:uuid, :name, :name_jp, :slug, :avatar, :status, :anime_themes_id, :anilist_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
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
		SET name = :name, name_jp = :name_jp, slug = :slug, avatar = :avatar, status = :status, 
		    anime_themes_id = :anime_themes_id, anilist_id = :anilist_id, updated_at = CURRENT_TIMESTAMP
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
			id, uuid, name, name_jp, slug, created_at, updated_at, avatar, status, anilist_id,
			enabled_songs, disabled_songs,
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

func (r *artistRepository) GetMany(ctx context.Context, ids []uint64) ([]domain.Artist, error) {
	var artists []domain.Artist
	if len(ids) == 0 {
		return []domain.Artist{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT 
			id, uuid, name, name_jp, slug, created_at, updated_at, avatar, status, favorites_count, anime_themes_id, anilist_id,
			enabled_songs, disabled_songs,
			(SELECT COUNT(*) FROM artist_song asong WHERE asong.artist_id = artists.id) as songs_count,
			(SELECT ani.banner 
			 FROM animes ani
			 JOIN songs s ON s.anime_id = ani.id
			 JOIN artist_song asong ON asong.song_id = s.id
			 WHERE asong.artist_id = artists.id
			 ORDER BY s.created_at DESC
			 LIMIT 1) as banner
		FROM artists
		WHERE id IN (?)
	`, ids)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &artists, query, args...)
	if artists == nil {
		artists = []domain.Artist{}
	}
	return artists, err
}

func (r *artistRepository) Search(ctx context.Context, term string, limit int) ([]domain.Artist, error) {
	var artists []domain.Artist
	query := `
		SELECT id, uuid, name, slug, avatar 
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

func (r *artistRepository) MergeDuplicateArtists(ctx context.Context, progress chan<- string) error {
	sendProgress := func(msg string) {
		if progress != nil {
			progress <- msg
		}
	}

	// 1. Find duplicate artist names
	var duplicates []struct {
		Name  string `db:"name"`
		Count int    `db:"count"`
	}
	err := r.db.SelectContext(ctx, &duplicates, "SELECT name, COUNT(*) as count FROM artists GROUP BY name HAVING COUNT(*) > 1")
	if err != nil {
		return fmt.Errorf("failed to query duplicates: %w", err)
	}

	sendProgress(fmt.Sprintf("Found %d duplicate artist names. Starting merge...", len(duplicates)))

	for _, d := range duplicates {
		sendProgress(fmt.Sprintf("Processing artist: %s", d.Name))

		// Get all IDs for this name, ordered by ID (smallest first as master)
		var ids []uint64
		err = r.db.SelectContext(ctx, &ids, "SELECT id FROM artists WHERE name = $1 ORDER BY id ASC", d.Name)
		if err != nil {
			sendProgress(fmt.Sprintf("  Error fetching IDs for %s: %v", d.Name, err))
			continue
		}

		if len(ids) < 2 {
			continue
		}

		masterID := ids[0]
		otherIDs := ids[1:]

		for _, otherID := range otherIDs {
			sendProgress(fmt.Sprintf("  Merging ID %d into %d...", otherID, masterID))

			// Start Tx
			tx, err := r.db.BeginTxx(ctx, nil)
			if err != nil {
				sendProgress(fmt.Sprintf("    Failed to start transaction: %v", err))
				continue
			}

			// Merge artist_song
			// Delete rows from other that already exist for master to avoid unique constraint violations
			_, err = tx.ExecContext(ctx, "DELETE FROM artist_song WHERE artist_id = $1 AND song_id IN (SELECT song_id FROM artist_song WHERE artist_id = $2)", otherID, masterID)
			if err != nil {
				sendProgress(fmt.Sprintf("    Error cleaning artist_song for %d: %v", otherID, err))
				tx.Rollback()
				continue
			}
			// Update remaining rows
			_, err = tx.ExecContext(ctx, "UPDATE artist_song SET artist_id = $1 WHERE artist_id = $2", masterID, otherID)
			if err != nil {
				sendProgress(fmt.Sprintf("    Error updating artist_song for %d: %v", otherID, err))
				tx.Rollback()
				continue
			}

			// Merge artist_user (favorites)
			_, err = tx.ExecContext(ctx, "DELETE FROM artist_user WHERE artist_id = $1 AND user_id IN (SELECT user_id FROM artist_user WHERE artist_id = $2)", otherID, masterID)
			if err != nil {
				sendProgress(fmt.Sprintf("    Error cleaning artist_user for %d: %v", otherID, err))
				tx.Rollback()
				continue
			}
			_, err = tx.ExecContext(ctx, "UPDATE artist_user SET artist_id = $1 WHERE artist_id = $2", masterID, otherID)
			if err != nil {
				sendProgress(fmt.Sprintf("    Error updating artist_user for %d: %v", otherID, err))
				tx.Rollback()
				continue
			}

			// Delete the duplicate artist
			_, err = tx.ExecContext(ctx, "DELETE FROM artists WHERE id = $1", otherID)
			if err != nil {
				sendProgress(fmt.Sprintf("    Error deleting artist %d: %v", otherID, err))
				tx.Rollback()
				continue
			}

			if err := tx.Commit(); err != nil {
				sendProgress(fmt.Sprintf("    Failed to commit transaction: %v", err))
			} else {
				sendProgress(fmt.Sprintf("    Successfully merged ID %d", otherID))
			}
		}
	}

	sendProgress("Recalculating statistics for all artists...")
	if err := r.RecountArtistStats(ctx, nil); err != nil {
		sendProgress(fmt.Sprintf("Warning: Failed to recount statistics: %v", err))
	}

	sendProgress("Artist merge complete!")
	return nil
}

func (r *artistRepository) RecountArtistStats(ctx context.Context, artistID *uint64) error {
	// 1. Recount enabled songs
	enabledQuery := `
		UPDATE artists SET enabled_songs = sub.cnt
		FROM (
			SELECT a.id, COALESCE(COUNT(s.id), 0) AS cnt
			FROM artists a
			LEFT JOIN artist_song asng ON asng.artist_id = a.id
			LEFT JOIN songs s ON s.id = asng.song_id AND s.status = true
			WHERE ($1::bigint IS NULL OR a.id = $1)
			GROUP BY a.id
		) sub
		WHERE artists.id = sub.id AND ($1::bigint IS NULL OR artists.id = $1)
	`
	_, err := r.db.ExecContext(ctx, enabledQuery, artistID)
	if err != nil {
		return fmt.Errorf("failed to recount enabled songs: %w", err)
	}

	// 2. Recount disabled songs
	disabledQuery := `
		UPDATE artists SET disabled_songs = sub.cnt
		FROM (
			SELECT a.id, COALESCE(COUNT(s.id), 0) AS cnt
			FROM artists a
			LEFT JOIN artist_song asng ON asng.artist_id = a.id
			LEFT JOIN songs s ON s.id = asng.song_id AND s.status = false
			WHERE ($1::bigint IS NULL OR a.id = $1)
			GROUP BY a.id
		) sub
		WHERE artists.id = sub.id AND ($1::bigint IS NULL OR artists.id = $1)
	`
	_, err = r.db.ExecContext(ctx, disabledQuery, artistID)
	if err != nil {
		return fmt.Errorf("failed to recount disabled songs: %w", err)
	}

	// 3. Recount favorites_count
	favQuery := `
		UPDATE artists SET favorites_count = (SELECT COUNT(*) FROM artist_user WHERE artist_id = artists.id)
		WHERE ($1::bigint IS NULL OR id = $1)
	`
	_, err = r.db.ExecContext(ctx, favQuery, artistID)
	if err != nil {
		return fmt.Errorf("failed to recount favorites: %w", err)
	}

	return nil
}

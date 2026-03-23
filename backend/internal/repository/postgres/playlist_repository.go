package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type playlistRepository struct {
	db *sqlx.DB
}

func NewPlaylistRepository(db *sqlx.DB) domain.PlaylistRepository {
	return &playlistRepository{db: db}
}

// ---- CRUD Playlists ----

func (r *playlistRepository) GetByID(ctx context.Context, id uint64) (*domain.Playlist, error) {
	query := `
		SELECT p.*,
		       u.id as "user.id",
		       u.name as "user.name",
		       u.slug as "user.slug",
		       (SELECT COUNT(*) FROM playlist_song WHERE playlist_id = p.id) as song_count,
		       (
				SELECT a.banner 
				FROM playlist_song ps
				JOIN songs s ON ps.song_id = s.id
				JOIN animes a ON s.anime_id = a.id
				WHERE ps.playlist_id = p.id
				ORDER BY ps.id DESC
				LIMIT 1
			) as latest_banner
		FROM playlists p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`
	// Simple mapping for nested struct
	type resStruct struct {
		domain.Playlist
		UID   uint64  `db:"user.id"`
		UName string  `db:"user.name"`
		USlug *string `db:"user.slug"`
	}
	var res resStruct
	err := r.db.GetContext(ctx, &res, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	playlist := res.Playlist
	playlist.User = &domain.User{
		ID:   res.UID,
		Name: res.UName,
		Slug: res.USlug,
	}

	return &playlist, nil
}

func (r *playlistRepository) GetByUserID(ctx context.Context, userID uint64, includePrivate bool, limit, offset int) ([]domain.Playlist, error) {
	return r.GetByUserIDWithSongCheck(ctx, userID, 0, includePrivate, limit, offset)
}

func (r *playlistRepository) GetByUserIDWithSongCheck(ctx context.Context, userID, songID uint64, includePrivate bool, limit, offset int) ([]domain.Playlist, error) {
	var playlists []domain.Playlist

	query := `
		SELECT 
			p.*,
			(SELECT COUNT(*) FROM playlist_song WHERE playlist_id = p.id) as song_count,
			(SELECT EXISTS(SELECT 1 FROM playlist_song WHERE playlist_id = p.id AND song_id = $1)) as contains_song,
			(
				SELECT a.banner 
				FROM playlist_song ps
				JOIN songs s ON ps.song_id = s.id
				JOIN animes a ON s.anime_id = a.id
				WHERE ps.playlist_id = p.id
				ORDER BY ps.id DESC
				LIMIT 1
			) as latest_banner
		FROM playlists p
		WHERE user_id = $2
	`
	args := []interface{}{songID, userID}
	i := 3

	if !includePrivate {
		query += " AND is_public = true"
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &playlists, query, args...)
	if err != nil {
		return nil, err
	}
	if playlists == nil {
		playlists = []domain.Playlist{}
	}
	return playlists, nil
}

func (r *playlistRepository) Create(ctx context.Context, playlist *domain.Playlist) error {
	query := `
		INSERT INTO playlists (name, description, user_id, is_public, created_at, updated_at) 
		VALUES (:name, :description, :user_id, :is_public, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, playlist).Scan(&playlist.ID)
	return err
}

func (r *playlistRepository) Update(ctx context.Context, playlist *domain.Playlist) error {
	query := `
		UPDATE playlists 
		SET name = :name, description = :description, is_public = :is_public, updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND user_id = :user_id
	`
	res, err := r.db.NamedExecContext(ctx, query, playlist)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("playlist not found or no changes made")
	}
	return err
}

func (r *playlistRepository) Delete(ctx context.Context, id, userID uint64) error {
	query := "DELETE FROM playlists WHERE id = $1 AND user_id = $2"
	res, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("unauthorized or playlist not found")
	}
	return err
}

// ---- Pivot Management (playlist_song) ----

func (r *playlistRepository) AddSong(ctx context.Context, playlistID, songID uint64, position int) error {
	query := `
		INSERT INTO playlist_song (playlist_id, song_id, position) 
		VALUES ($1, $2, $3)
		ON CONFLICT (playlist_id, song_id) DO UPDATE SET position = EXCLUDED.position
	`
	_, err := r.db.ExecContext(ctx, query, playlistID, songID, position)
	return err
}

func (r *playlistRepository) RemoveSong(ctx context.Context, playlistID, songID uint64) error {
	query := "DELETE FROM playlist_song WHERE playlist_id = $1 AND song_id = $2"
	_, err := r.db.ExecContext(ctx, query, playlistID, songID)
	return err
}

func (r *playlistRepository) UpdateSongPositions(ctx context.Context, playlistID uint64, items []domain.PlaylistSong) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "UPDATE playlist_song SET position = $1 WHERE playlist_id = $2 AND song_id = $3"
	for _, item := range items {
		_, err = tx.ExecContext(ctx, query, item.Position, playlistID, item.SongID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *playlistRepository) GetSongs(ctx context.Context, playlistID uint64) ([]domain.PlaylistSong, error) {
	var items []domain.PlaylistSong

	query := `
		SELECT 
			ps.playlist_id, ps.song_id, ps.position,
			s.id as "song.id",
			s.song_romaji as "song.song_romaji",
			s.song_jp as "song.song_jp",
			s.song_en as "song.song_en",
			s.theme_num as "song.theme_num",
			s.type as "song.type",
			s.slug as "song.slug",
			s.anime_id as "song.anime_id",
			s.views as "song.views"
		FROM playlist_song ps
		JOIN songs s ON ps.song_id = s.id
		WHERE ps.playlist_id = $1
		ORDER BY ps.position ASC, s.created_at DESC
	`

	type resStruct struct {
		PlaylistID uint64 `db:"playlist_id"`
		SongID     uint64 `db:"song_id"`
		Position   int    `db:"position"`

		S_ID      uint64  `db:"song.id"`
		S_Romaji  *string `db:"song.song_romaji"`
		S_Jp      *string `db:"song.song_jp"`
		S_En      *string `db:"song.song_en"`
		S_Num     string  `db:"song.theme_num"`
		S_Type    string  `db:"song.type"`
		S_Slug    string  `db:"song.slug"`
		S_AnimeID uint64  `db:"song.anime_id"`
		S_Views   uint64  `db:"song.views"`
	}

	var rows []resStruct
	err := r.db.SelectContext(ctx, &rows, query, playlistID)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		item := domain.PlaylistSong{
			PlaylistID: row.PlaylistID,
			SongID:     row.SongID,
			Position:   row.Position,
			Song: &domain.Song{
				ID:         row.S_ID,
				SongRomaji: row.S_Romaji,
				SongJP:     row.S_Jp,
				SongEN:     row.S_En,
				ThemeNum:   row.S_Num,
				Type:       row.S_Type,
				Slug:       row.S_Slug,
				AnimeID:    row.S_AnimeID,
				Views:      row.S_Views,
			},
		}
		items = append(items, item)
	}

	if items == nil {
		items = []domain.PlaylistSong{}
	}

	return items, nil
}

func (r *playlistRepository) GetPaginatedPublicPlaylists(ctx context.Context, limit, offset int, filters domain.PlaylistFilters) ([]domain.Playlist, error) {
	var playlists []domain.Playlist

	query := `
		SELECT 
			p.*,
			u.id as "user.id",
			u.name as "user.name",
			u.slug as "user.slug",
			(SELECT COUNT(*) FROM playlist_song WHERE playlist_id = p.id) as song_count,
			(
				SELECT a.banner 
				FROM playlist_song ps
				JOIN songs s ON ps.song_id = s.id
				JOIN animes a ON s.anime_id = a.id
				WHERE ps.playlist_id = p.id
				ORDER BY ps.id DESC
				LIMIT 1
			) as latest_banner
		FROM playlists p
		JOIN users u ON p.user_id = u.id
		WHERE p.is_public = true
	`

	args := []interface{}{}
	i := 1
	if filters.Search != "" {
		query += fmt.Sprintf(" AND p.name ILIKE $%d", i)
		args = append(args, "%"+filters.Search+"%")
		i++
	}

	query += fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	// resStruct for nested mapping
	type resStruct struct {
		domain.Playlist
		UID   uint64  `db:"user.id"`
		UName string  `db:"user.name"`
		USlug *string `db:"user.slug"`
	}

	var rows []resStruct
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		p := row.Playlist
		p.User = &domain.User{
			ID:   row.UID,
			Name: row.UName,
			Slug: row.USlug,
		}
		playlists = append(playlists, p)
	}

	if playlists == nil {
		playlists = []domain.Playlist{}
	}

	return playlists, nil
}

func (r *playlistRepository) CountPublicPlaylists(ctx context.Context, filters domain.PlaylistFilters) (int, error) {
	query := "SELECT COUNT(*) FROM playlists p WHERE is_public = true"
	args := []interface{}{}

	if filters.Search != "" {
		query += " AND p.name ILIKE $1"
		args = append(args, "%"+filters.Search+"%")
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

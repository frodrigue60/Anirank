package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type songRepository struct {
	db *sqlx.DB
}

func NewSongRepository(db *sqlx.DB) domain.SongRepository {
	return &songRepository{db: db.Unsafe()}
}

func (r *songRepository) GetByID(ctx context.Context, id uint64) (*domain.Song, error) {
	var s domain.Song
	query := `
		SELECT s.*, 
		       a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug", a.cover AS "anime.cover",
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s 
		LEFT JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.id = $1
	`
	err := r.db.GetContext(ctx, &s, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *songRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Song, error) {
	var s domain.Song
	query := `
		SELECT s.*, 
		       a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug", a.cover AS "anime.cover",
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s 
		LEFT JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.uuid = $1
	`
	err := r.db.GetContext(ctx, &s, query, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *songRepository) GetBySlug(ctx context.Context, slug string) (*domain.Song, error) {
	var s domain.Song
	query := `
		SELECT s.*,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s 
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.slug = $1
	`
	err := r.db.GetContext(ctx, &s, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *songRepository) GetByAnimeIDAndSlug(ctx context.Context, animeID uint64, slug string) (*domain.Song, error) {
	var s domain.Song
	query := `
		SELECT s.*,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s 
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.anime_id = $1 AND s.slug = $2
	`
	err := r.db.GetContext(ctx, &s, query, animeID, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *songRepository) GetPaginated(ctx context.Context, limit, offset int, filters domain.SongFilters) ([]domain.Song, error) {
	var songs []domain.Song
	query := `
		SELECT s.*, 
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description",
		       EXISTS (SELECT 1 FROM artist_song asong JOIN artists art ON asong.artist_id = art.id WHERE asong.song_id = s.id AND art.status = false) as partial_artist_inactive
		FROM songs s
		JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
	`
	var args []interface{}
	var whereClauses []string

	// Enforce active anime status for public routes
	if !filters.IsAdmin {
		whereClauses = append(whereClauses, "a.status = true")
	}

	i := 1
	
	if filters.IsAdmin {
		if filters.YearID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.year_id = $%d", i))
			args = append(args, filters.YearID)
			i++
		}
		if filters.SeasonID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.season_id = $%d", i))
			args = append(args, filters.SeasonID)
			i++
		}
		if filters.GenreID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id IN (SELECT anime_id FROM anime_genre WHERE genre_id = $%d)", i))
			args = append(args, filters.GenreID)
			i++
		}
	} else {
		if filters.Year != "" && filters.Year != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.year_id IN (SELECT id FROM years WHERE slug = $%d)", i))
			args = append(args, filters.Year)
			i++
		}
		if filters.Season != "" && filters.Season != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.season_id IN (SELECT id FROM seasons WHERE slug = $%d)", i))
			args = append(args, filters.Season)
			i++
		}
		if filters.Genre != "" && filters.Genre != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id IN (SELECT ag.anime_id FROM anime_genre ag JOIN genres g ON ag.genre_id = g.id WHERE g.slug = $%d)", i))
			args = append(args, filters.Genre)
			i++
		}
	}

	if filters.TypeID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("s.type_id = $%d", i))
		args = append(args, filters.TypeID)
		i++
	} else if filters.Type != "" && filters.Type != "any" {
		whereClauses = append(whereClauses, fmt.Sprintf("st.slug = $%d", i))
		args = append(args, filters.Type)
		i++
	}
	if filters.Format != "" && filters.Format != "any" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.format_id IN (SELECT id FROM formats WHERE slug = $%d)", i))
		args = append(args, filters.Format)
		i++
	}
	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(s.song_romaji ILIKE $%d OR s.song_jp ILIKE $%d OR s.song_en ILIKE $%d OR a.title ILIKE $%d OR s.id IN (SELECT song_id FROM artist_song asong JOIN artists art ON asong.artist_id = art.id WHERE asong.song_id = s.id AND art.name ILIKE $%d))", i, i+1, i+2, i+3, i+4))
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		i += 5
	}
	if filters.AnimeID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id = $%d", i))
		args = append(args, filters.AnimeID)
		i++
	}
	if filters.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("s.status = $%d", i))
		args = append(args, *filters.Status)
		i++
	}

	// Enforce active status for public routes
	if !filters.IsAdmin {
		whereClauses = append(whereClauses, "a.status = true", "s.status = true")
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + joinWS(whereClauses, " AND ")
	}

	// Sorting
	switch filters.Sort {
	case "rating":
		query += " ORDER BY s.average_score DESC, s.created_at DESC"
	case "rating_asc":
		query += " ORDER BY s.average_score ASC, s.created_at DESC"
	case "favorites":
		query += " ORDER BY s.favorites_count DESC, s.created_at DESC"
	case "views":
		query += " ORDER BY s.views DESC, s.created_at DESC"
	case "recently_added":
		query += " ORDER BY s.created_at DESC, s.id DESC"
	case "random":
		query += " ORDER BY RANDOM()"
	default:
		query += " ORDER BY s.created_at DESC, s.id DESC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &songs, query, args...)
	if err != nil {
		return nil, err
	}

	if songs == nil {
		songs = []domain.Song{}
	}

	return songs, nil
}

func joinWS(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	res := strs[0]
	for i := 1; i < len(strs); i++ {
		res += sep + strs[i]
	}
	return res
}

func (r *songRepository) Count(ctx context.Context, filters domain.SongFilters) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM songs s
		JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
	`
	var args []interface{}
	var whereClauses []string

	if !filters.IsAdmin {
		whereClauses = append(whereClauses, "a.status = true", "s.status = true")
	}

	i := 1

	if filters.IsAdmin {
		if filters.YearID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.year_id = $%d", i))
			args = append(args, filters.YearID)
			i++
		}
		if filters.SeasonID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.season_id = $%d", i))
			args = append(args, filters.SeasonID)
			i++
		}
		if filters.GenreID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id IN (SELECT anime_id FROM anime_genre WHERE genre_id = $%d)", i))
			args = append(args, filters.GenreID)
			i++
		}
	} else {
		if filters.Year != "" && filters.Year != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.year_id IN (SELECT id FROM years WHERE slug = $%d)", i))
			args = append(args, filters.Year)
			i++
		}
		if filters.Season != "" && filters.Season != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.season_id IN (SELECT id FROM seasons WHERE slug = $%d)", i))
			args = append(args, filters.Season)
			i++
		}
		if filters.Genre != "" && filters.Genre != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id IN (SELECT ag.anime_id FROM anime_genre ag JOIN genres g ON ag.genre_id = g.id WHERE g.slug = $%d)", i))
			args = append(args, filters.Genre)
			i++
		}
	}

	if filters.TypeID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("s.type_id = $%d", i))
		args = append(args, filters.TypeID)
		i++
	} else if filters.Type != "" && filters.Type != "any" {
		whereClauses = append(whereClauses, fmt.Sprintf("st.slug = $%d", i))
		args = append(args, filters.Type)
		i++
	}
	if filters.Format != "" && filters.Format != "any" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.format_id IN (SELECT id FROM formats WHERE slug = $%d)", i))
		args = append(args, filters.Format)
		i++
	}
	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(s.song_romaji ILIKE $%d OR s.song_jp ILIKE $%d OR s.song_en ILIKE $%d OR a.title ILIKE $%d OR s.id IN (SELECT song_id FROM artist_song asong JOIN artists art ON asong.artist_id = art.id WHERE asong.song_id = s.id AND art.name ILIKE $%d))", i, i+1, i+2, i+3, i+4))
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		i += 5
	}
	if filters.AnimeID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id = $%d", i))
		args = append(args, filters.AnimeID)
		i++
	}
	if filters.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("s.status = $%d", i))
		args = append(args, *filters.Status)
		i++
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + joinWS(whereClauses, " AND ")
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

func (r *songRepository) Create(ctx context.Context, song *domain.Song) error {
	// Legacy fallback: if type_id is missing but type is present, try to map it
	if (song.TypeID == nil || *song.TypeID == 0) && song.Type != "" {
		_ = r.db.GetContext(ctx, &song.TypeID, "SELECT id FROM song_types WHERE slug = $1 OR name = $1 LIMIT 1", song.Type)
	}

	query := `
		INSERT INTO songs (song_romaji, song_jp, song_en, uuid, theme_num, type, type_id, slug, anime_id, season_id, year_id, views, status, anime_themes_id, created_at, updated_at) 
		VALUES (:song_romaji, :song_jp, :song_en, :uuid, :theme_num, :type, :type_id, :slug, :anime_id, :season_id, :year_id, :views, :status, :anime_themes_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, song).Scan(&song.ID)
	return err
}

func (r *songRepository) Update(ctx context.Context, song *domain.Song) error {
	// Legacy fallback: if type_id is missing but type is present, try to map it
	if (song.TypeID == nil || *song.TypeID == 0) && song.Type != "" {
		_ = r.db.GetContext(ctx, &song.TypeID, "SELECT id FROM song_types WHERE slug = $1 OR name = $1 LIMIT 1", song.Type)
	}

	query := `
		UPDATE songs 
		SET song_romaji = :song_romaji, song_jp = :song_jp, song_en = :song_en, 
		    theme_num = :theme_num, type = :type, type_id = :type_id, slug = :slug, anime_id = :anime_id, 
		    season_id = :season_id, year_id = :year_id, views = :views, status = :status, 
		    anime_themes_id = :anime_themes_id, updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	res, err := r.db.NamedExecContext(ctx, query, song)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("song not found or no changes made")
	}
	return err
}

func (r *songRepository) Delete(ctx context.Context, id uint64) error {
	query := "DELETE FROM songs WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("song not found")
	}
	return err
}

func (r *songRepository) GetVariantsBySongID(ctx context.Context, songID uint64) ([]domain.SongVariant, error) {
	query := `
		SELECT 
			sv.id, sv.uuid, sv.version_number, sv.song_id, sv.slug, sv.views, sv.season_id, sv.year_id, sv.spoiler, sv.status, sv.created_at, sv.updated_at,
			v.video_src, v.embed_code
		FROM song_variants sv
		LEFT JOIN videos v ON sv.id = v.song_variant_id
		WHERE sv.song_id = $1
		ORDER BY sv.version_number ASC
	`

	type VariantWithVideoStruct struct {
		domain.SongVariant
		VideoSrc  *string `db:"video_src"`
		EmbedCode *string `db:"embed_code"`
	}

	var rows []VariantWithVideoStruct
	err := r.db.SelectContext(ctx, &rows, query, songID)
	if err != nil {
		return nil, err
	}

	var variants []domain.SongVariant
	for _, row := range rows {
		v := row.SongVariant
		if row.VideoSrc != nil || row.EmbedCode != nil {
			v.Video = &domain.SongVariantVideo{
				EmbedUrl: extractSrcFromIframe(row.EmbedCode),
				LocalUrl: row.VideoSrc,
			}
			
			if row.VideoSrc != nil && *row.VideoSrc != "" {
				v.Video.Type = "file"
			} else if row.EmbedCode != nil && *row.EmbedCode != "" {
				v.Video.Type = "embed"
			}
		}
		variants = append(variants, v)
	}

	if variants == nil {
		variants = []domain.SongVariant{}
	}
	return variants, nil
}

func (r *songRepository) GetArtistsBySongID(ctx context.Context, songID uint64, isAdmin bool) ([]domain.Artist, error) {
	var artists []domain.Artist
	query := `
		SELECT a.* 
		FROM artists a
		JOIN artist_song asong ON a.id = asong.artist_id
		WHERE asong.song_id = $1
	`
	if !isAdmin {
		query += " AND a.status = true"
	}

	err := r.db.SelectContext(ctx, &artists, query, songID)
	if err != nil {
		return nil, err
	}

	if artists == nil {
		artists = []domain.Artist{}
	}
	return artists, nil
}

func (r *songRepository) SyncArtists(ctx context.Context, songID uint64, artistIDs []uint64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old
	_, err = tx.ExecContext(ctx, "DELETE FROM artist_song WHERE song_id = $1", songID)
	if err != nil {
		return err
	}

	// Insert new
	if len(artistIDs) > 0 {
		insertQuery := "INSERT INTO artist_song (artist_id, song_id) VALUES ($1, $2)"
		for _, aid := range artistIDs {
			_, err = tx.ExecContext(ctx, insertQuery, aid, songID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *songRepository) GetByAnimeID(ctx context.Context, animeID uint64) ([]domain.Song, error) {
	var songs []domain.Song
	query := `
		SELECT s.*, 
		       s.average_score, 
		       s.favorites_count,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description",
		       EXISTS (SELECT 1 FROM artist_song asong JOIN artists art ON asong.artist_id = art.id WHERE asong.song_id = s.id AND art.status = false) as partial_artist_inactive
		FROM songs s
		JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.anime_id = $1 AND a.status = true AND s.status = true
		ORDER BY s.type_id ASC, s.theme_num ASC
	`
	err := r.db.SelectContext(ctx, &songs, query, animeID)
	if songs == nil {
		songs = []domain.Song{}
	}
	return songs, err
}

func (r *songRepository) GetMany(ctx context.Context, ids []uint64) ([]domain.Song, error) {
	if len(ids) == 0 {
		return []domain.Song{}, nil
	}
	query, args, err := sqlx.In(`
		SELECT s.*, s.average_score, s.favorites_count,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s 
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.id IN (?)
	`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var songs []domain.Song
	err = r.db.SelectContext(ctx, &songs, query, args...)
	if songs == nil {
		songs = []domain.Song{}
	}
	return songs, err
}

func (r *songRepository) GetByArtistID(ctx context.Context, artistID uint64, limit, offset int, filters domain.SongFilters) ([]domain.Song, error) {
	var songs []domain.Song
	query := `
		SELECT s.*, s.average_score, s.favorites_count,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s
		JOIN artist_song asng ON s.id = asng.song_id
		JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
	`
	var args []interface{}
	var whereClauses []string

	i := 1
	whereClauses = append(whereClauses, fmt.Sprintf("asng.artist_id = $%d", i), "a.status = true", "s.status = true")
	args = append(args, artistID)
	i++

	if filters.IsAdmin {
		if filters.YearID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.year_id = $%d", i))
			args = append(args, filters.YearID)
			i++
		}
		if filters.SeasonID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.season_id = $%d", i))
			args = append(args, filters.SeasonID)
			i++
		}
		if filters.GenreID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id IN (SELECT anime_id FROM anime_genre WHERE genre_id = $%d)", i))
			args = append(args, filters.GenreID)
			i++
		}
	} else {
		if filters.Year != "" && filters.Year != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.year_id IN (SELECT id FROM years WHERE slug = $%d)", i))
			args = append(args, filters.Year)
			i++
		}
		if filters.Season != "" && filters.Season != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.season_id IN (SELECT id FROM seasons WHERE slug = $%d)", i))
			args = append(args, filters.Season)
			i++
		}
		if filters.Genre != "" && filters.Genre != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id IN (SELECT ag.anime_id FROM anime_genre ag JOIN genres g ON ag.genre_id = g.id WHERE g.slug = $%d)", i))
			args = append(args, filters.Genre)
			i++
		}
	}

	if filters.Type != "" && filters.Type != "any" {
		whereClauses = append(whereClauses, fmt.Sprintf("s.type = $%d", i))
		args = append(args, filters.Type)
		i++
	}

	if filters.Format != "" && filters.Format != "any" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.format_id IN (SELECT id FROM formats WHERE slug = $%d)", i))
		args = append(args, filters.Format)
		i++
	}

	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(s.song_romaji ILIKE $%d OR s.song_jp ILIKE $%d OR s.song_en ILIKE $%d)", i, i+1, i+2))
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		i += 3
	}

	query += " WHERE " + joinWS(whereClauses, " AND ")

	// Sorting
	switch filters.Sort {
	case "rating":
		query += " ORDER BY s.average_score DESC, s.created_at DESC"
	case "rating_asc":
		query += " ORDER BY s.average_score ASC, s.created_at DESC"
	case "favorites":
		query += " ORDER BY s.favorites_count DESC, s.created_at DESC"
	case "recently_added":
		query += " ORDER BY s.created_at DESC, s.id DESC"
	default:
		query += " ORDER BY s.created_at DESC, s.id DESC"
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &songs, query, args...)
	if songs == nil {
		songs = []domain.Song{}
	}
	return songs, err
}

func (r *songRepository) CountByArtistID(ctx context.Context, artistID uint64, filters domain.SongFilters) (int, error) {
	var count int
	query := `
		SELECT COUNT(DISTINCT s.id) 
		FROM songs s
		JOIN artist_song asng ON s.id = asng.song_id
		JOIN animes a ON s.anime_id = a.id
	`
	var args []interface{}
	var whereClauses []string

	i := 1
	whereClauses = append(whereClauses, fmt.Sprintf("asng.artist_id = $%d", i), "a.status = true", "s.status = true")
	args = append(args, artistID)
	i++

	if filters.IsAdmin {
		if filters.YearID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.year_id = $%d", i))
			args = append(args, filters.YearID)
			i++
		}
		if filters.SeasonID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.season_id = $%d", i))
			args = append(args, filters.SeasonID)
			i++
		}
		if filters.GenreID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id IN (SELECT anime_id FROM anime_genre WHERE genre_id = $%d)", i))
			args = append(args, filters.GenreID)
			i++
		}
	} else {
		if filters.Year != "" && filters.Year != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.year_id IN (SELECT id FROM years WHERE slug = $%d)", i))
			args = append(args, filters.Year)
			i++
		}
		if filters.Season != "" && filters.Season != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.season_id IN (SELECT id FROM seasons WHERE slug = $%d)", i))
			args = append(args, filters.Season)
			i++
		}
		if filters.Genre != "" && filters.Genre != "any" {
			whereClauses = append(whereClauses, fmt.Sprintf("s.anime_id IN (SELECT ag.anime_id FROM anime_genre ag JOIN genres g ON ag.genre_id = g.id WHERE g.slug = $%d)", i))
			args = append(args, filters.Genre)
			i++
		}
	}

	if filters.Type != "" && filters.Type != "any" {
		whereClauses = append(whereClauses, fmt.Sprintf("s.type = $%d", i))
		args = append(args, filters.Type)
		i++
	}

	if filters.Format != "" && filters.Format != "any" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.format_id IN (SELECT id FROM formats WHERE slug = $%d)", i))
		args = append(args, filters.Format)
		i++
	}
	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(s.song_romaji ILIKE $%d OR s.song_jp ILIKE $%d OR s.song_en ILIKE $%d)", i, i+1, i+2))
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	query += " WHERE " + joinWS(whereClauses, " AND ")

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

func (r *songRepository) CountFavoritesByUserID(ctx context.Context, userID uint64) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM song_user f
		WHERE f.user_id = $1
	`
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *songRepository) GetFavoritesByUserID(ctx context.Context, userID uint64, limit, offset int) ([]domain.Song, error) {
	var songs []domain.Song
	// One row per favorite; do not join song_ratings here — it multiplies rows and breaks
	// GROUP BY ... ORDER BY f.created_at under PostgreSQL rules.
	query := `
		SELECT s.*, 
		       s.average_score, s.favorites_count,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s
		JOIN song_user f ON s.id = f.song_id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &songs, query, userID, limit, offset)
	if songs == nil {
		songs = []domain.Song{}
	}
	return songs, err
}

func (r *songRepository) GetRanking(ctx context.Context, rankingType, songType string, limit, offset int) ([]domain.Song, error) {
	var songs []domain.Song
	var query string
	var args []interface{}

	if rankingType == "seasonal" {
		query = `
			SELECT s.*, 
				s.average_score,
				s.favorites_count,
				s.prev_seasonal_rank as prev_rank,
				st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
			FROM songs s
			JOIN animes a ON s.anime_id = a.id
			JOIN seasons se ON s.season_id = se.id
			JOIN years y ON s.year_id = y.id
			LEFT JOIN song_types st ON s.type_id = st.id
			WHERE se.current = true AND y.current = true AND a.status = true AND s.status = true
		`
	} else {
		query = `
			SELECT s.*, 
				s.average_score,
				s.favorites_count,
				s.prev_main_rank as prev_rank,
				st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
			FROM songs s
			JOIN animes a ON s.anime_id = a.id
			LEFT JOIN song_types st ON s.type_id = st.id
			WHERE a.status = true AND s.status = true
		`
	}

	if songType != "" && songType != "all" {
		query += " AND s.type = $1"
		args = append(args, songType)
	}

	i := len(args)
	query += fmt.Sprintf(` 
		ORDER BY s.average_score DESC, 
		s.views DESC 
		LIMIT $%d OFFSET $%d
	`, i+1, i+2)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &songs, query, args...)
	if songs == nil {
		songs = []domain.Song{}
	}
	return songs, err
}

func (r *songRepository) CountRanking(ctx context.Context, rankingType, songType string) (int, error) {
	var count int
	var query string
	var args []interface{}

	switch rankingType {
	case "seasonal":
		query = `
			SELECT COUNT(*) 
			FROM songs s
			JOIN animes a ON s.anime_id = a.id
			JOIN seasons se ON s.season_id = se.id
			JOIN years y ON s.year_id = y.id
			LEFT JOIN song_types st ON s.type_id = st.id
			WHERE se.current = true AND y.current = true AND a.status = true AND s.status = true
		`
	default: // global
		query = `
			SELECT COUNT(*) 
			FROM songs s
			JOIN animes a ON s.anime_id = a.id
			LEFT JOIN song_types st ON s.type_id = st.id
			WHERE a.status = true AND s.status = true
		`
	}

	if songType != "" && songType != "all" {
		query += " AND s.type = $1"
		args = append(args, songType)
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

func (r *songRepository) IncrementViews(ctx context.Context, id uint64) error {
	query := "UPDATE songs SET views = views + 1 WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *songRepository) Search(ctx context.Context, term string, limit int) ([]domain.Song, error) {
	var songs []domain.Song
	query := `
		SELECT 
			s.id, s.uuid, s.song_romaji, s.song_jp, s.song_en, s.slug, s.theme_num,
			a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug",
			st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s
		JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE (s.song_romaji ILIKE $1 OR s.song_jp ILIKE $2 OR s.song_en ILIKE $3) AND a.status = true AND s.status = true
		ORDER BY s.created_at DESC
		LIMIT $4
	`
	err := r.db.SelectContext(ctx, &songs, query, term, term, term, limit)
	if songs == nil {
		songs = []domain.Song{}
	}
	return songs, err
}

func extractSrcFromIframe(iframe *string) *string {
	if iframe == nil || *iframe == "" {
		return nil
	}

	re := regexp.MustCompile(`src="([^"]+)"`)
	matches := re.FindStringSubmatch(*iframe)
	if len(matches) > 1 {
		src := matches[1]
		return &src
	}

	return iframe
}

func (r *songRepository) ToggleStatus(ctx context.Context, id uint64) error {
	query := "UPDATE songs SET status = NOT status WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *songRepository) GetPublicSlugs(ctx context.Context) ([]domain.SitemapItem, error) {
	var items []domain.SitemapItem
	// Filter by both song status AND anime status
	query := `
		SELECT a.slug || '/' || s.slug as loc, s.updated_at as lastmod 
		FROM songs s
		JOIN animes a ON s.anime_id = a.id
		WHERE s.status = true AND a.status = true
	`
	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}

func (r *songRepository) GetSongTypes(ctx context.Context) ([]domain.SongType, error) {
	var types []domain.SongType
	query := `SELECT id, uuid, name, slug, description FROM song_types ORDER BY name ASC`
	err := r.db.SelectContext(ctx, &types, query)
	if types == nil {
		types = []domain.SongType{}
	}
	return types, err
}

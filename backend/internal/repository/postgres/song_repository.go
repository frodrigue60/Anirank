package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"anirank/api/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type songRepository struct {
	db *sqlx.DB
}

func NewSongRepository(db *sqlx.DB) domain.SongRepository {
	return &songRepository{db: db.Unsafe()}
}

const songColumns = `s.id, s.uuid, s.song_romaji, s.song_jp, s.song_en, s.theme_num, st.slug AS type, s.type_id, (st.slug || s.theme_num) AS slug, s.anime_id, s.season_id, s.year_id, s.views, s.likes_count, s.dislikes_count, s.favorites_count, s.average_score, s.status, s.anime_themes_id, s.prev_main_rank, s.prev_seasonal_rank, s.created_at, s.updated_at`

func (r *songRepository) GetByID(ctx context.Context, id uint64) (*domain.Song, error) {
	var s domain.Song
	query := fmt.Sprintf(`
		SELECT %s, 
		       a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug", a.cover AS "anime.cover", a.banner AS "anime.banner", a.year_id AS "anime.year_id", a.season_id AS "anime.season_id",
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"

		FROM songs s 
		LEFT JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.id = $1
	`, songColumns)
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
	query := fmt.Sprintf(`
		SELECT %s, 
		       a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug", a.cover AS "anime.cover", a.banner AS "anime.banner", a.year_id AS "anime.year_id", a.season_id AS "anime.season_id",
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"

		FROM songs s 
		LEFT JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.uuid = $1
	`, songColumns)
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
	query := fmt.Sprintf(`
		SELECT %s,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s 
		JOIN song_types st ON s.type_id = st.id
		WHERE (st.slug || s.theme_num) = $1
	`, songColumns)
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
	query := fmt.Sprintf(`
		SELECT %s,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s 
		JOIN song_types st ON s.type_id = st.id
		WHERE s.anime_id = $1 AND (st.slug || s.theme_num) = $2
	`, songColumns)
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
	query := fmt.Sprintf(`
		SELECT %s, 
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description",
		       a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug", a.cover AS "anime.cover", a.banner AS "anime.banner",
		       EXISTS (SELECT 1 FROM artist_song asong JOIN artists art ON asong.artist_id = art.id WHERE asong.song_id = s.id AND art.status = false) as partial_artist_inactive
		FROM songs s
		JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
	`, songColumns)
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
	} else if filters.Type != "" && filters.Type != "any" && filters.Type != "all" {
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

	if filters.Cursor != "" {
		// Use ID as a simple cursor for now
		var cursorID uint64
		fmt.Sscanf(filters.Cursor, "%d", &cursorID)
		if cursorID > 0 {
			whereClauses = append(whereClauses, fmt.Sprintf("s.id < $%d", i))
			args = append(args, cursorID)
			i++
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + joinWS(whereClauses, " AND ")
	}

	// Sorting
	switch filters.Sort {
	case "rating":
		query += " ORDER BY s.average_score DESC, s.created_at DESC, s.id DESC"
	case "rating_asc":
		query += " ORDER BY s.average_score ASC, s.created_at DESC, s.id DESC"
	case "favorites":
		query += " ORDER BY s.favorites_count DESC, s.created_at DESC, s.id DESC"
	case "views":
		query += " ORDER BY s.views DESC, s.created_at DESC, s.id DESC"
	case "recently_added":
		query += " ORDER BY s.created_at DESC, s.id DESC"
	case "random":
		query += " ORDER BY RANDOM()"
	default:
		query += " ORDER BY s.created_at DESC, s.id DESC"
	}

	// Only apply OFFSET if cursor is not present (for compatibility)
	if filters.Cursor != "" {
		query += fmt.Sprintf(" LIMIT $%d", i)
		args = append(args, limit)
	} else {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
		args = append(args, limit, offset)
	}

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
	} else if filters.Type != "" && filters.Type != "any" && filters.Type != "all" {
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
		INSERT INTO songs (song_romaji, song_jp, song_en, uuid, theme_num, type_id, anime_id, season_id, year_id, views, status, anime_themes_id, created_at, updated_at) 
		VALUES (:song_romaji, :song_jp, :song_en, :uuid, :theme_num, :type_id, :anime_id, :season_id, :year_id, :views, :status, :anime_themes_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
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
		    theme_num = :theme_num, type_id = :type_id, anime_id = :anime_id, 
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

func (r *songRepository) GetVariantsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64][]domain.SongVariant, error) {
	if len(songIDs) == 0 {
		return map[uint64][]domain.SongVariant{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT 
			sv.id, sv.uuid, sv.version_number, sv.song_id, sv.slug, sv.views, sv.season_id, sv.year_id, sv.episodes, sv.spoiler, sv.nsfw, sv.status, sv.created_at, sv.updated_at,
			v.video_src, v.embed_code,
			COALESCE(v.is_nc, false) AS is_nc,
			COALESCE(v.is_bd, false) AS is_bd,
			COALESCE(v.resolution, 0) AS resolution,
			COALESCE(v.is_uncensored, false) AS is_uncensored,
			COALESCE(v.is_subbed, false) AS is_subbed,
			COALESCE(v.is_lyrics, false) AS is_lyrics,
			COALESCE(v.source, 'TV') AS source,
			COALESCE(v.overlap, 'None') AS overlap
		FROM song_variants sv
		LEFT JOIN videos v ON sv.id = v.song_variant_id
		WHERE sv.song_id IN (?)
		ORDER BY sv.song_id, sv.version_number ASC, v.resolution DESC, v.is_nc DESC
	`, songIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	type VariantWithVideoStruct struct {
		domain.SongVariant
		VideoSrc     *string `db:"video_src"`
		EmbedCode    *string `db:"embed_code"`
		IsNC         bool    `db:"is_nc"`
		IsBD         bool    `db:"is_bd"`
		Resolution   int     `db:"resolution"`
		IsUncensored bool    `db:"is_uncensored"`
		IsSubbed     bool    `db:"is_subbed"`
		IsLyrics     bool    `db:"is_lyrics"`
		Source       string  `db:"source"`
		Overlap      string  `db:"overlap"`
	}

	var rows []VariantWithVideoStruct
	err = r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	// Group rows by variant ID to avoid duplicates
	variantMap := make(map[uint64]*domain.SongVariant)
	// Track order of variants per song to maintain sorting
	songVariantsOrder := make(map[uint64][]uint64)

	for _, row := range rows {
		svID := row.SongVariant.ID
		songID := row.SongVariant.SongID

		if _, exists := variantMap[svID]; !exists {
			v := row.SongVariant
			v.Videos = []domain.SongVariantVideo{}
			variantMap[svID] = &v
			songVariantsOrder[songID] = append(songVariantsOrder[songID], svID)
		}

		if row.VideoSrc != nil || row.EmbedCode != nil {
			vid := domain.SongVariantVideo{
				EmbedCode:    row.EmbedCode,
				VideoSrc:     row.VideoSrc,
				EmbedUrl:     extractSrcFromIframe(row.EmbedCode),
				LocalUrl:     row.VideoSrc,
				IsNC:         row.IsNC,
				IsBD:         row.IsBD,
				Resolution:   row.Resolution,
				IsUncensored: row.IsUncensored,
				IsSubbed:     row.IsSubbed,
				IsLyrics:     row.IsLyrics,
				Source:       row.Source,
				Overlap:      row.Overlap,
			}

			if row.VideoSrc != nil && *row.VideoSrc != "" {
				vid.Type = "file"
			} else if row.EmbedCode != nil && *row.EmbedCode != "" {
				vid.Type = "embed"
			}

			variantMap[svID].Videos = append(variantMap[svID].Videos, vid)
		}
	}

	// Populate single Video fallback and build final result
	result := make(map[uint64][]domain.SongVariant)
	for songID, svIDs := range songVariantsOrder {
		result[songID] = make([]domain.SongVariant, 0, len(svIDs))
		for _, svID := range svIDs {
			v := variantMap[svID]
			if len(v.Videos) > 0 {
				v.Video = &v.Videos[0]
			}
			result[songID] = append(result[songID], *v)
		}
	}

	return result, nil
}

func (r *songRepository) GetVariantsBySongID(ctx context.Context, songID uint64) ([]domain.SongVariant, error) {
	query := `
		SELECT 
			sv.id, sv.uuid, sv.version_number, sv.song_id, sv.slug, sv.views, sv.season_id, sv.year_id, sv.episodes, sv.spoiler, sv.nsfw, sv.status, sv.created_at, sv.updated_at,
			v.video_src, v.embed_code,
			COALESCE(v.is_nc, false) AS is_nc,
			COALESCE(v.is_bd, false) AS is_bd,
			COALESCE(v.resolution, 0) AS resolution,
			COALESCE(v.is_uncensored, false) AS is_uncensored,
			COALESCE(v.is_subbed, false) AS is_subbed,
			COALESCE(v.is_lyrics, false) AS is_lyrics,
			COALESCE(v.source, 'TV') AS source,
			COALESCE(v.overlap, 'None') AS overlap
		FROM song_variants sv
		LEFT JOIN videos v ON sv.id = v.song_variant_id
		WHERE sv.song_id = $1
		ORDER BY sv.version_number ASC, v.resolution DESC, v.is_nc DESC
	`

	type VariantWithVideoStruct struct {
		domain.SongVariant
		VideoSrc     *string `db:"video_src"`
		EmbedCode    *string `db:"embed_code"`
		IsNC         bool    `db:"is_nc"`
		IsBD         bool    `db:"is_bd"`
		Resolution   int     `db:"resolution"`
		IsUncensored bool    `db:"is_uncensored"`
		IsSubbed     bool    `db:"is_subbed"`
		IsLyrics     bool    `db:"is_lyrics"`
		Source       string  `db:"source"`
		Overlap      string  `db:"overlap"`
	}

	var rows []VariantWithVideoStruct
	err := r.db.SelectContext(ctx, &rows, query, songID)
	if err != nil {
		return nil, err
	}

	variantMap := make(map[uint64]*domain.SongVariant)
	var orderedIDs []uint64

	for _, row := range rows {
		svID := row.SongVariant.ID
		if _, exists := variantMap[svID]; !exists {
			v := row.SongVariant
			v.Videos = []domain.SongVariantVideo{}
			variantMap[svID] = &v
			orderedIDs = append(orderedIDs, svID)
		}

		if row.VideoSrc != nil || row.EmbedCode != nil {
			vid := domain.SongVariantVideo{
				EmbedCode:    row.EmbedCode,
				VideoSrc:     row.VideoSrc,
				EmbedUrl:     extractSrcFromIframe(row.EmbedCode),
				LocalUrl:     row.VideoSrc,
				IsNC:         row.IsNC,
				IsBD:         row.IsBD,
				Resolution:   row.Resolution,
				IsUncensored: row.IsUncensored,
				IsSubbed:     row.IsSubbed,
				IsLyrics:     row.IsLyrics,
				Source:       row.Source,
				Overlap:      row.Overlap,
			}

			if row.VideoSrc != nil && *row.VideoSrc != "" {
				vid.Type = "file"
			} else if row.EmbedCode != nil && *row.EmbedCode != "" {
				vid.Type = "embed"
			}

			variantMap[svID].Videos = append(variantMap[svID].Videos, vid)
		}
	}

	variants := make([]domain.SongVariant, 0, len(orderedIDs))
	for _, svID := range orderedIDs {
		v := variantMap[svID]
		if len(v.Videos) > 0 {
			v.Video = &v.Videos[0]
		}
		variants = append(variants, *v)
	}

	if len(variants) == 0 {
		variants = []domain.SongVariant{}
	}
	return variants, nil
}

func (r *songRepository) GetArtistsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64][]domain.Artist, error) {
	if len(songIDs) == 0 {
		return map[uint64][]domain.Artist{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT a.*, asong.song_id as linked_song_id
		FROM artists a
		JOIN artist_song asong ON a.id = asong.artist_id
		WHERE asong.song_id IN (?) AND a.status = true
	`, songIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	type ArtistWithSongID struct {
		domain.Artist
		LinkedSongID uint64 `db:"linked_song_id"`
	}

	var rows []ArtistWithSongID
	err = r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	result := make(map[uint64][]domain.Artist)
	for _, row := range rows {
		result[row.LinkedSongID] = append(result[row.LinkedSongID], row.Artist)
	}

	return result, nil
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

func (r *songRepository) GetByAnimeID(ctx context.Context, animeID uint64, isAdmin bool) ([]domain.Song, error) {
	var songs []domain.Song
	query := fmt.Sprintf(`
		SELECT %s, 
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description",
		       EXISTS (SELECT 1 FROM artist_song asong JOIN artists art ON asong.artist_id = art.id WHERE asong.song_id = s.id AND art.status = false) as partial_artist_inactive
		FROM songs s
		JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE s.anime_id = $1
	`, songColumns)
	if !isAdmin {
		query += " AND a.status = true AND s.status = true"
	}
	query += " ORDER BY s.type_id ASC, LENGTH(s.theme_num) ASC, s.theme_num ASC"
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
	query, args, err := sqlx.In(fmt.Sprintf(`
		SELECT %s,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description",
		       a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.cover AS "anime.cover", a.banner AS "anime.banner", a.slug AS "anime.slug"
		FROM songs s 
		LEFT JOIN song_types st ON s.type_id = st.id
		LEFT JOIN animes a ON s.anime_id = a.id
		WHERE s.id IN (?)
	`, songColumns), ids)
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
	query := fmt.Sprintf(`
		SELECT %s,
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s
		JOIN artist_song asng ON s.id = asng.song_id
		JOIN animes a ON s.anime_id = a.id
		LEFT JOIN song_types st ON s.type_id = st.id
	`, songColumns)
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
		whereClauses = append(whereClauses, fmt.Sprintf("(s.song_romaji ILIKE $%d OR s.song_jp ILIKE $%d OR s.song_en ILIKE $%d)", i, i+1, i+2))
		searchTerm := "%" + filters.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		i += 3
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + joinWS(whereClauses, " AND ")
	}

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
	query := fmt.Sprintf(`
		SELECT %s, 
		       st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description"
		FROM songs s
		JOIN song_user f ON s.id = f.song_id
		LEFT JOIN song_types st ON s.type_id = st.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`, songColumns)

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
		query = fmt.Sprintf(`
			SELECT %s, 
				s.prev_seasonal_rank as prev_rank,
				st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description",
				a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug", a.cover AS "anime.cover", a.banner AS "anime.banner"
			FROM songs s
			JOIN animes a ON s.anime_id = a.id
			JOIN seasons se ON s.season_id = se.id
			JOIN years y ON s.year_id = y.id
			LEFT JOIN song_types st ON s.type_id = st.id
			WHERE se.current = true AND y.current = true AND a.status = true AND s.status = true
		`, songColumns)
	} else {
		query = fmt.Sprintf(`
			SELECT %s, 
				s.prev_main_rank as prev_rank,
				st.id AS "song_type.id", st.uuid AS "song_type.uuid", st.name AS "song_type.name", st.slug AS "song_type.slug", st.description AS "song_type.description",
				a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug", a.cover AS "anime.cover", a.banner AS "anime.banner"
			FROM songs s
			JOIN animes a ON s.anime_id = a.id
			LEFT JOIN song_types st ON s.type_id = st.id
			WHERE a.status = true AND s.status = true
		`, songColumns)
	}

	if songType != "" && songType != "all" {
		query += " AND st.slug = $1"
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
		query += " AND st.slug = $1"
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
			s.id, s.uuid, s.song_romaji, s.song_jp, s.song_en, (st.slug || s.theme_num) AS slug, s.theme_num,
			a.id AS "anime.id", a.uuid AS "anime.uuid", a.title AS "anime.title", a.slug AS "anime.slug", a.cover AS "anime.cover", a.banner AS "anime.banner",
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
		SELECT a.slug || '/' || (st.slug || s.theme_num) as loc, s.updated_at as lastmod 
		FROM songs s
		JOIN animes a ON s.anime_id = a.id
		JOIN song_types st ON s.type_id = st.id
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

// UpsertSongFromAnimeThemes inserts a song by its anime_themes_id.
// Returns (true, nil) if created, (false, nil) if already existed.
func (r *songRepository) UpsertSongFromAnimeThemes(ctx context.Context, song *domain.Song) (bool, error) {
	song.UUID = uuid.New().String()
	now := time.Now().UTC()

	// Resolve type_id from slug (OP/ED/IN/INS/OTH)
	var typeID uint64
	if err := r.db.GetContext(ctx, &typeID, `SELECT id FROM song_types WHERE slug = $1 LIMIT 1`, song.Type); err != nil {
		// Create the type if it doesn't exist
		newTypeUUID := uuid.New().String()
		errInsert := r.db.QueryRowContext(ctx, `
			INSERT INTO song_types (uuid, name, slug) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
			RETURNING id
		`, newTypeUUID, song.Type, song.Type).Scan(&typeID)
		
		if errInsert != nil {
			// Fallback to first type if completely failed
			_ = r.db.GetContext(ctx, &typeID, `SELECT id FROM song_types ORDER BY id LIMIT 1`)
		}
	}
	song.TypeID = &typeID

	var returnedID uint64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO songs (uuid, song_romaji, song_jp, theme_num, type_id, anime_id, season_id, year_id, views, status, anime_themes_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, false, $9, $10, $10)
		ON CONFLICT (anime_themes_id) DO NOTHING
		RETURNING id
	`,
		song.UUID, song.SongRomaji, song.SongJP,
		song.ThemeNum, typeID, song.AnimeID,
		song.SeasonID, song.YearID,
		song.AnimeThemesID, now,
	).Scan(&returnedID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err2 := r.db.QueryRowContext(ctx,
				`SELECT id FROM songs WHERE anime_themes_id = $1`, song.AnimeThemesID,
			).Scan(&song.ID)
			return false, err2
		}
		return false, err
	}
	song.ID = returnedID
	return true, nil
}

// UpsertVariantFromAnimeThemes inserts a song_variant and its video row.
// Returns (true, nil) if created, (false, nil) if already existed.
func (r *songRepository) UpsertVariantFromAnimeThemes(ctx context.Context, v *domain.SongVariant, videos []domain.SongVariantVideo) (bool, error) {
	v.UUID = uuid.New().String()
	now := time.Now().UTC()

	var returnedID uint64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO song_variants (uuid, version_number, song_id, slug, views, season_id, year_id, spoiler, nsfw, status, anime_themes_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 0, $5, $6, $7, $8, false, $9, $10, $10)
		ON CONFLICT (anime_themes_id) DO NOTHING
		RETURNING id
	`,
		v.UUID, v.VersionNumber, v.SongID, v.Slug,
		v.SeasonID, v.YearID,
		v.Spoiler, v.NSFW,
		v.AnimeThemesID, now,
	).Scan(&returnedID)

	var isCreated bool = true
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			isCreated = false
			err2 := r.db.QueryRowContext(ctx,
				`SELECT id FROM song_variants WHERE anime_themes_id = $1`, v.AnimeThemesID,
			).Scan(&v.ID)
			if err2 != nil {
				return false, err2
			}
		} else {
			return false, err
		}
	} else {
		v.ID = returnedID
	}

	// Insert all video rows
	for _, video := range videos {
		if video.VideoSrc != nil && *video.VideoSrc != "" {
			_, errVideo := r.db.ExecContext(ctx, `
				INSERT INTO videos (song_variant_id, video_src, is_nc, is_bd, resolution, is_uncensored, is_subbed, is_lyrics, source, overlap, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, false, $11, $11)
				ON CONFLICT (song_variant_id, video_src) DO UPDATE SET
					is_nc = EXCLUDED.is_nc,
					is_bd = EXCLUDED.is_bd,
					resolution = EXCLUDED.resolution,
					is_uncensored = EXCLUDED.is_uncensored,
					is_subbed = EXCLUDED.is_subbed,
					is_lyrics = EXCLUDED.is_lyrics,
					source = EXCLUDED.source,
					overlap = EXCLUDED.overlap,
					updated_at = EXCLUDED.updated_at
			`, v.ID, video.VideoSrc, video.IsNC, video.IsBD, video.Resolution, video.IsUncensored, video.IsSubbed, video.IsLyrics, video.Source, video.Overlap, now)
			if errVideo != nil {
				fmt.Printf("[UpsertVariantFromAnimeThemes] Error inserting video %s for variant %d: %v\n", *video.VideoSrc, v.ID, errVideo)
			}
		}
	}

	return isCreated, nil
}

// LinkArtistToSong creates an artist_song pivot row idempotently.
func (r *songRepository) LinkArtistToSong(ctx context.Context, songID, artistID uint64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO artist_song (artist_id, song_id, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (artist_id, song_id) DO NOTHING
	`, artistID, songID)
	return err
}

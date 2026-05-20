package postgres

import (
	"context"
	"errors"
	"strconv"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type songVariantRepository struct {
	db *sqlx.DB
}

func NewSongVariantRepository(db *sqlx.DB) domain.SongVariantRepository {
	return &songVariantRepository{db: db}
}

func (r *songVariantRepository) GetByID(ctx context.Context, id uint64) (*domain.SongVariant, error) {
	query := `
		SELECT 
			sv.id, sv.uuid, sv.version_number, sv.song_id, sv.slug, sv.views, sv.season_id, sv.year_id, sv.episodes, sv.spoiler, sv.nsfw, sv.status, sv.created_at, sv.updated_at, sv.anime_themes_id,
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
		WHERE sv.id = $1
		ORDER BY v.resolution DESC, v.is_nc DESC
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
	err := r.db.SelectContext(ctx, &rows, query, id)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, domain.ErrNotFound
	}

	// Group rows
	v := rows[0].SongVariant
	v.Videos = []domain.SongVariantVideo{}

	for _, row := range rows {
		if row.VideoSrc != nil || row.EmbedCode != nil {
			vid := domain.SongVariantVideo{
				EmbedCode:    row.EmbedCode,
				VideoSrc:     row.VideoSrc,
				EmbedUrl:     row.EmbedCode, // Initial assignment
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

			v.Videos = append(v.Videos, vid)
		}
	}

	if len(v.Videos) > 0 {
		v.Video = &v.Videos[0]
	}

	return &v, nil
}

func (r *songVariantRepository) GetPaginated(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]domain.SongVariant, error) {
	var variants []domain.SongVariant
	query := `
		SELECT sv.* FROM song_variants sv
		JOIN songs s ON sv.song_id = s.id
		JOIN animes a ON s.anime_id = a.id
		WHERE 1=1
	`
	isAdmin, _ := filters["is_admin"].(bool)
	if !isAdmin {
		query += " AND a.status = true AND s.status = true AND sv.status = true"
	}

	search, _ := filters["search"].(string)
	animeID, _ := filters["anime_id"].(uint64)
	status, hasStatus := filters["status"].(bool)

	args := []interface{}{limit, offset}
	argCount := 3

	if search != "" {
		query += " AND (sv.slug ILIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+search+"%")
		argCount++
	}

	if animeID != 0 {
		query += " AND a.id = $" + strconv.Itoa(argCount)
		args = append(args, animeID)
		argCount++
	}

	if hasStatus {
		query += " AND sv.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	query += " ORDER BY sv.created_at DESC, sv.id DESC"
	query += " LIMIT $1 OFFSET $2"
	err := r.db.SelectContext(ctx, &variants, query, args...)
	return variants, err
}

func (r *songVariantRepository) Count(ctx context.Context, filters map[string]interface{}) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM song_variants sv
		JOIN songs s ON sv.song_id = s.id
		JOIN animes a ON s.anime_id = a.id
		WHERE 1=1
	`
	isAdmin, _ := filters["is_admin"].(bool)
	if !isAdmin {
		query += " AND a.status = true AND s.status = true AND sv.status = true"
	}

	search, _ := filters["search"].(string)
	animeID, _ := filters["anime_id"].(uint64)
	status, hasStatus := filters["status"].(bool)

	var args []interface{}
	argCount := 1

	if search != "" {
		query += " AND (sv.slug ILIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+search+"%")
		argCount++
	}

	if animeID != 0 {
		query += " AND a.id = $" + strconv.Itoa(argCount)
		args = append(args, animeID)
		argCount++
	}

	if hasStatus {
		query += " AND sv.status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

func (r *songVariantRepository) IncrementViews(ctx context.Context, id uint64) error {
	query := "UPDATE song_variants SET views = views + 1 WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *songVariantRepository) Create(ctx context.Context, variant *domain.SongVariant) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO song_variants (uuid, version_number, song_id, slug, views, season_id, year_id, episodes, spoiler, nsfw, status, anime_themes_id, created_at, updated_at) 
		VALUES (:uuid, :version_number, :song_id, :slug, :views, :season_id, :year_id, :episodes, :spoiler, :nsfw, :status, :anime_themes_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, variant).Scan(&variant.ID)
	if err != nil {
		return err
	}

	// In legacy structure, we also insert into the videos table
	if variant.Video != nil && (variant.Video.LocalUrl != nil || variant.Video.EmbedUrl != nil) {
		videoQuery := `
			INSERT INTO videos (song_variant_id, video_src, embed_code, is_nc, is_bd, resolution, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err = tx.ExecContext(ctx, videoQuery, variant.ID, variant.Video.VideoSrc, variant.Video.EmbedCode, variant.Video.IsNC, variant.Video.IsBD, variant.Video.Resolution, variant.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *songVariantRepository) Update(ctx context.Context, variant *domain.SongVariant) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE song_variants 
		SET version_number = :version_number, song_id = :song_id, slug = :slug, views = :views, 
		    season_id = :season_id, year_id = :year_id, episodes = :episodes, spoiler = :spoiler, nsfw = :nsfw, status = :status, 
		    anime_themes_id = :anime_themes_id, updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	_, err = tx.NamedExecContext(ctx, query, variant)
	if err != nil {
		return err
	}

	// Upsert the Video. Delete existing then insert to resolve complexity.
	_, err = tx.ExecContext(ctx, "DELETE FROM videos WHERE song_variant_id = $1", variant.ID)
	if err != nil {
		return err
	}

	if variant.Video != nil && (variant.Video.LocalUrl != nil || variant.Video.EmbedUrl != nil) {
		videoQuery := `
			INSERT INTO videos (song_variant_id, video_src, embed_code, is_nc, is_bd, resolution, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err = tx.ExecContext(ctx, videoQuery, variant.ID, variant.Video.VideoSrc, variant.Video.EmbedCode, variant.Video.IsNC, variant.Video.IsBD, variant.Video.Resolution, variant.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *songVariantRepository) Delete(ctx context.Context, id uint64) error {
	// The video table has ON DELETE CASCADE for song_variant_id
	query := "DELETE FROM song_variants WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("song variant not found")
	}
	return err
}

func (r *songVariantRepository) ToggleStatus(ctx context.Context, id uint64) error {
	query := "UPDATE song_variants SET status = NOT status WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *songVariantRepository) ToggleSpoiler(ctx context.Context, id uint64) error {
	query := "UPDATE song_variants SET spoiler = NOT spoiler WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *songVariantRepository) ToggleNSFW(ctx context.Context, id uint64) error {
	query := "UPDATE song_variants SET nsfw = NOT nsfw WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *songVariantRepository) DeleteVideo(ctx context.Context, variantID uint64, videoSrc *string, embedCode *string) error {
	var query string
	var args []interface{}

	if videoSrc != nil && *videoSrc != "" {
		query = "DELETE FROM videos WHERE song_variant_id = $1 AND video_src = $2"
		args = []interface{}{variantID, *videoSrc}
	} else if embedCode != nil && *embedCode != "" {
		query = "DELETE FROM videos WHERE song_variant_id = $1 AND embed_code = $2"
		args = []interface{}{variantID, *embedCode}
	} else {
		query = "DELETE FROM videos WHERE song_variant_id = $1"
		args = []interface{}{variantID}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

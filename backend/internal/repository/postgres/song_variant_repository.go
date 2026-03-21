package postgres

import (
	"context"
	"database/sql"
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
			sv.id, sv.version_number, sv.song_id, sv.slug, sv.views, sv.season_id, sv.year_id, sv.spoiler, sv.status, sv.created_at, sv.updated_at,
			v.video_src, v.embed_code
		FROM song_variants sv
		LEFT JOIN videos v ON sv.id = v.song_variant_id
		WHERE sv.id = $1
	`

	type VariantWithVideoStruct struct {
		domain.SongVariant
		VideoSrc  *string `db:"video_src"`
		EmbedCode *string `db:"embed_code"`
	}

	var row VariantWithVideoStruct
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	v := row.SongVariant
	if row.VideoSrc != nil || row.EmbedCode != nil {
		v.Video = &domain.SongVariantVideo{
			EmbedUrl: row.EmbedCode,
			LocalUrl: row.VideoSrc,
		}

		// Infer type
		if row.VideoSrc != nil && *row.VideoSrc != "" {
			v.Video.Type = "file"
		} else if row.EmbedCode != nil && *row.EmbedCode != "" {
			v.Video.Type = "embed"
		}
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

	query += " ORDER BY sv.id DESC"
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
		INSERT INTO song_variants (version_number, song_id, slug, views, season_id, year_id, spoiler, status, created_at, updated_at) 
		VALUES (:version_number, :song_id, :slug, :views, :season_id, :year_id, :spoiler, :status, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
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
			INSERT INTO videos (song_variant_id, video_src, embed_code, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err = tx.ExecContext(ctx, videoQuery, variant.ID, variant.Video.LocalUrl, variant.Video.EmbedUrl, variant.Status)
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
		    season_id = :season_id, year_id = :year_id, spoiler = :spoiler, status = :status, updated_at = CURRENT_TIMESTAMP
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
			INSERT INTO videos (song_variant_id, video_src, embed_code, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`
		_, err = tx.ExecContext(ctx, videoQuery, variant.ID, variant.Video.LocalUrl, variant.Video.EmbedUrl, variant.Status)
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

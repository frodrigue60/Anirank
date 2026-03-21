package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type moderationRepository struct {
	db *sqlx.DB
}

func NewModerationRepository(db *sqlx.DB) domain.ModerationRepository {
	return &moderationRepository{db: db}
}

// ---- User Facing (Create) ----

func (r *moderationRepository) CreateSongReport(ctx context.Context, report *domain.SongReport) error {
	query := `
		INSERT INTO song_reports (song_id, user_id, source, title, content, status, created_at, updated_at) 
		VALUES (:song_id, :user_id, :source, :title, :content, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, report).Scan(&report.ID)
	return err
}

func (r *moderationRepository) IsSongReportedByUser(ctx context.Context, userID, songID uint64) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM song_reports WHERE user_id = $1 AND song_id = $2 AND status = false"
	err := r.db.GetContext(ctx, &count, query, userID, songID)
	return count > 0, err
}

func (r *moderationRepository) CreateCommentReport(ctx context.Context, report *domain.CommentReport) error {
	query := `
		INSERT INTO comment_reports (comment_id, user_id, source, title, content, status, created_at, updated_at) 
		VALUES (:comment_id, :user_id, :source, :title, :content, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, report).Scan(&report.ID)
	return err
}

func (r *moderationRepository) CreateUserRequest(ctx context.Context, request *domain.UserRequest) error {
	query := `
		INSERT INTO user_requests (title, content, user_id, status, created_at, updated_at) 
		VALUES (:title, :content, :user_id, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, request).Scan(&request.ID)
	return err
}

// ---- Admin Facing (Read & Update) ----

func (r *moderationRepository) GetSongReports(ctx context.Context, status *bool, limit, offset int) ([]domain.SongReport, error) {
	query := `
		SELECT r.*, 
		       u.id as "user.id", u.name as "user.name", u.slug as "user.slug",
		       s.id as "song.id", s.slug as "song.slug", s.song_romaji as "song.song_romaji",
		       s.song_en as "song.song_en", s.song_jp as "song.song_jp",
		       a.id as "anime.id", a.slug as "anime.slug", a.title as "anime.title"
		FROM song_reports r
		JOIN users u ON r.user_id = u.id
		JOIN songs s ON r.song_id = s.id
		LEFT JOIN animes a ON s.anime_id = a.id
		WHERE 1=1
	`
	var args []interface{}
	i := 1
	if status != nil {
		query += fmt.Sprintf(" AND r.status = $%d", i)
		args = append(args, *status)
		i++
	}
	query += fmt.Sprintf(`
		ORDER BY r.created_at DESC
		LIMIT $%d OFFSET $%d
	`, i, i+1)
	args = append(args, limit, offset)

	type ReportRow struct {
		domain.SongReport
		UserID     uint64  `db:"user.id"`
		UserName   string  `db:"user.name"`
		UserSlug   *string `db:"user.slug"`
		SongID     uint64  `db:"song.id"`
		SongSlug   string  `db:"song.slug"`
		SongTitle  *string `db:"song.song_romaji"`
		SongEN     *string `db:"song.song_en"`
		SongJP     *string `db:"song.song_jp"`
		AnimeID    *uint64 `db:"anime.id"`
		AnimeSlug  *string `db:"anime.slug"`
		AnimeTitle *string `db:"anime.title"`
	}

	var rows []ReportRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	var reports []domain.SongReport
	for _, row := range rows {
		rep := row.SongReport
		rep.User = &domain.User{
			ID:   row.UserID,
			Name: row.UserName,
			Slug: row.UserSlug,
		}
		rep.Song = &domain.Song{
			ID:         row.SongID,
			Slug:       row.SongSlug,
			SongRomaji: row.SongTitle,
			SongEN:     row.SongEN,
			SongJP:     row.SongJP,
		}
		if row.AnimeID != nil {
			rep.Song.Anime = &domain.Anime{
				ID:    *row.AnimeID,
				Slug:  *row.AnimeSlug,
				Title: *row.AnimeTitle,
			}
		}
		reports = append(reports, rep)
	}

	if reports == nil {
		reports = []domain.SongReport{}
	}

	return reports, nil
}

func (r *moderationRepository) GetSongReport(ctx context.Context, reportID uint64) (*domain.SongReport, error) {
	query := `
		SELECT r.*, 
		       u.id as "user.id", u.name as "user.name", u.slug as "user.slug",
		       s.id as "song.id", s.slug as "song.slug", s.song_romaji as "song.song_romaji",
		       s.song_en as "song.song_en", s.song_jp as "song.song_jp",
		       a.id as "anime.id", a.slug as "anime.slug", a.title as "anime.title"
		FROM song_reports r
		JOIN users u ON r.user_id = u.id
		JOIN songs s ON r.song_id = s.id
		LEFT JOIN animes a ON s.anime_id = a.id
		WHERE r.id = $1
	`

	type ReportRow struct {
		domain.SongReport
		UserID     uint64  `db:"user.id"`
		UserName   string  `db:"user.name"`
		UserSlug   *string `db:"user.slug"`
		SongID     uint64  `db:"song.id"`
		SongSlug   string  `db:"song.slug"`
		SongTitle  *string `db:"song.song_romaji"`
		SongEN     *string `db:"song.song_en"`
		SongJP     *string `db:"song.song_jp"`
		AnimeID    *uint64 `db:"anime.id"`
		AnimeSlug  *string `db:"anime.slug"`
		AnimeTitle *string `db:"anime.title"`
	}

	var row ReportRow
	err := r.db.GetContext(ctx, &row, query, reportID)
	if err != nil {
		return nil, err
	}

	rep := row.SongReport
	rep.User = &domain.User{
		ID:   row.UserID,
		Name: row.UserName,
		Slug: row.UserSlug,
	}
	rep.Song = &domain.Song{
		ID:         row.SongID,
		Slug:       row.SongSlug,
		SongRomaji: row.SongTitle,
		SongEN:     row.SongEN,
		SongJP:     row.SongJP,
	}
	if row.AnimeID != nil {
		rep.Song.Anime = &domain.Anime{
			ID:    *row.AnimeID,
			Slug:  *row.AnimeSlug,
			Title: *row.AnimeTitle,
		}
	}

	return &rep, nil
}

func (r *moderationRepository) ResolveSongReport(ctx context.Context, reportID uint64) error {
	query := "UPDATE song_reports SET status = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, reportID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("report not found or already resolved")
	}

	return err
}

func (r *moderationRepository) GetCommentReports(ctx context.Context, status *bool, limit, offset int) ([]domain.CommentReport, error) {
	query := `
		SELECT r.*, 
		       u.id as "user.id", u.name as "user.name", u.slug as "user.slug",
		       c.id as "comment.id", c.content as "comment.content"
		FROM comment_reports r
		JOIN users u ON r.user_id = u.id
		JOIN comments c ON r.comment_id = c.id
		WHERE 1=1
	`
	var args []interface{}
	i := 1
	if status != nil {
		query += fmt.Sprintf(" AND r.status = $%d", i)
		args = append(args, *status)
		i++
	}
	query += fmt.Sprintf(`
		ORDER BY r.created_at DESC
		LIMIT $%d OFFSET $%d
	`, i, i+1)
	args = append(args, limit, offset)

	type ReportRow struct {
		domain.CommentReport
		UserID         uint64  `db:"user.id"`
		UserName       string  `db:"user.name"`
		UserSlug       *string `db:"user.slug"`
		CommentID      uint64  `db:"comment.id"`
		CommentContent string  `db:"comment.content"`
	}

	var rows []ReportRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	var reports []domain.CommentReport
	for _, row := range rows {
		rep := row.CommentReport
		rep.User = &domain.User{
			ID:   row.UserID,
			Name: row.UserName,
			Slug: row.UserSlug,
		}
		rep.Comment = &domain.Comment{
			ID:      row.CommentID,
			Content: row.CommentContent,
		}
		reports = append(reports, rep)
	}

	if reports == nil {
		reports = []domain.CommentReport{}
	}

	return reports, nil
}

func (r *moderationRepository) GetCommentReport(ctx context.Context, reportID uint64) (*domain.CommentReport, error) {
	query := `
		SELECT r.*, 
		       u.id as "user.id", u.name as "user.name", u.slug as "user.slug",
		       c.id as "comment.id", c.content as "comment.content"
		FROM comment_reports r
		JOIN users u ON r.user_id = u.id
		JOIN comments c ON r.comment_id = c.id
		WHERE r.id = $1
	`

	type ReportRow struct {
		domain.CommentReport
		UserID         uint64  `db:"user.id"`
		UserName       string  `db:"user.name"`
		UserSlug       *string `db:"user.slug"`
		CommentID      uint64  `db:"comment.id"`
		CommentContent string  `db:"comment.content"`
	}

	var row ReportRow
	err := r.db.GetContext(ctx, &row, query, reportID)
	if err != nil {
		return nil, err
	}

	rep := row.CommentReport
	rep.User = &domain.User{
		ID:   row.UserID,
		Name: row.UserName,
		Slug: row.UserSlug,
	}
	rep.Comment = &domain.Comment{
		ID:      row.CommentID,
		Content: row.CommentContent,
	}

	return &rep, nil
}

func (r *moderationRepository) ResolveCommentReport(ctx context.Context, reportID uint64) error {
	query := "UPDATE comment_reports SET status = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, reportID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("report not found or already resolved")
	}

	return err
}

func (r *moderationRepository) GetUserRequests(ctx context.Context, status *bool, limit, offset int) ([]domain.UserRequest, error) {
	query := `
		SELECT ur.*, 
		       u.id as "user.id", u.name as "user.name", u.slug as "user.slug"
		FROM user_requests ur
		JOIN users u ON ur.user_id = u.id
		WHERE 1=1
	`
	var args []interface{}
	i := 1
	if status != nil {
		query += fmt.Sprintf(" AND ur.status = $%d", i)
		args = append(args, *status)
		i++
	}
	query += fmt.Sprintf(`
		ORDER BY ur.created_at ASC
		LIMIT $%d OFFSET $%d
	`, i, i+1)
	args = append(args, limit, offset)

	type RequestRow struct {
		domain.UserRequest
		UserID   uint64  `db:"user.id"`
		UserName string  `db:"user.name"`
		UserSlug *string `db:"user.slug"`
	}

	var rows []RequestRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	var reqs []domain.UserRequest
	for _, row := range rows {
		req := row.UserRequest
		req.User = &domain.User{
			ID:   row.UserID,
			Name: row.UserName,
			Slug: row.UserSlug,
		}
		reqs = append(reqs, req)
	}

	if reqs == nil {
		reqs = []domain.UserRequest{}
	}

	return reqs, nil
}

func (r *moderationRepository) GetUserRequest(ctx context.Context, requestID uint64) (*domain.UserRequest, error) {
	query := `
		SELECT ur.*, 
		       u.id as "user.id", u.name as "user.name", u.slug as "user.slug"
		FROM user_requests ur
		JOIN users u ON ur.user_id = u.id
		WHERE ur.id = $1
	`

	type RequestRow struct {
		domain.UserRequest
		UserID   uint64  `db:"user.id"`
		UserName string  `db:"user.name"`
		UserSlug *string `db:"user.slug"`
	}

	var row RequestRow
	err := r.db.GetContext(ctx, &row, query, requestID)
	if err != nil {
		return nil, err
	}

	req := row.UserRequest
	req.User = &domain.User{
		ID:   row.UserID,
		Name: row.UserName,
		Slug: row.UserSlug,
	}

	return &req, nil
}

func (r *moderationRepository) UpdateUserRequestStatus(ctx context.Context, requestID uint64, status bool, adminID uint64) error {
	var query string
	var res sql.Result
	var err error

	if status {
		query = "UPDATE user_requests SET status = true, attended_by = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
		res, err = r.db.ExecContext(ctx, query, adminID, requestID)
	} else {
		query = "UPDATE user_requests SET status = false, attended_by = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1"
		res, err = r.db.ExecContext(ctx, query, requestID)
	}

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("request not found")
	}

	return err
}

func (r *moderationRepository) DeleteSongReport(ctx context.Context, reportID uint64) error {
	query := "DELETE FROM song_reports WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, reportID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("report not found")
	}

	return err
}

func (r *moderationRepository) DeleteCommentReport(ctx context.Context, reportID uint64) error {
	query := "DELETE FROM comment_reports WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, reportID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("report not found")
	}

	return err
}

func (r *moderationRepository) DeleteUserRequest(ctx context.Context, requestID uint64) error {
	query := "DELETE FROM user_requests WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, requestID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("request not found")
	}

	return err
}

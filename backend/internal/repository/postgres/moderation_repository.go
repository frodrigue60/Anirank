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

func (r *moderationRepository) GetSongReportsByUserAndSongIDs(ctx context.Context, userID uint64, songIDs []uint64) (map[uint64]bool, error) {
	if len(songIDs) == 0 {
		return map[uint64]bool{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT song_id 
		FROM song_reports 
		WHERE user_id = ? AND song_id IN (?) AND status = false
	`, userID, songIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var reportedIDs []uint64
	err = r.db.SelectContext(ctx, &reportedIDs, query, args...)
	if err != nil {
		return nil, err
	}

	res := make(map[uint64]bool)
	for _, id := range reportedIDs {
		res[id] = true
	}
	return res, nil
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
		       u.truth_score as "user.truth_score", u.is_shadowbanned as "user.is_shadowbanned",
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
		IsAccepted bool `db:"is_accepted"`
		domain.SongReport
		UserID         uint64  `db:"user.id"`
		UserName       string  `db:"user.name"`
		UserSlug       *string `db:"user.slug"`
		UserTruthScore int     `db:"user.truth_score"`
		UserShadow     bool    `db:"user.is_shadowbanned"`
		SongID         uint64  `db:"song.id"`
		SongSlug       string  `db:"song.slug"`
		SongTitle      *string `db:"song.song_romaji"`
		SongEN         *string `db:"song.song_en"`
		SongJP         *string `db:"song.song_jp"`
		AnimeID        *uint64 `db:"anime.id"`
		AnimeSlug      *string `db:"anime.slug"`
		AnimeTitle     *string `db:"anime.title"`
	}

	var rows []ReportRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	var reports []domain.SongReport
	for _, row := range rows {
		rep := row.SongReport
		rep.IsAccepted = row.IsAccepted
		rep.User = &domain.User{
			ID:             row.UserID,
			Name:           row.UserName,
			Slug:           row.UserSlug,
			TruthScore:     row.UserTruthScore,
			IsShadowbanned: row.UserShadow,
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
		       u.truth_score as "user.truth_score", u.is_shadowbanned as "user.is_shadowbanned",
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
		IsAccepted bool `db:"is_accepted"`
		domain.SongReport
		UserID         uint64  `db:"user.id"`
		UserName       string  `db:"user.name"`
		UserSlug       *string `db:"user.slug"`
		UserTruthScore int     `db:"user.truth_score"`
		UserShadow     bool    `db:"user.is_shadowbanned"`
		SongID         uint64  `db:"song.id"`
		SongSlug       string  `db:"song.slug"`
		SongTitle      *string `db:"song.song_romaji"`
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
	rep.IsAccepted = row.IsAccepted
	rep.User = &domain.User{
		ID:             row.UserID,
		Name:           row.UserName,
		Slug:           row.UserSlug,
		TruthScore:     row.UserTruthScore,
		IsShadowbanned: row.UserShadow,
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

func (r *moderationRepository) ResolveSongReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	query := "UPDATE song_reports SET status = true, is_accepted = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	res, err := r.db.ExecContext(ctx, query, isAccepted, reportID)
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
		       u.truth_score as "user.truth_score", u.is_shadowbanned as "user.is_shadowbanned",
		       c.id as "comment.id", c.content as "comment.content", c.user_id as "comment.user_id"
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
		IsAccepted bool `db:"is_accepted"`
		domain.CommentReport
		UserID         uint64  `db:"user.id"`
		UserName       string  `db:"user.name"`
		UserSlug       *string `db:"user.slug"`
		UserTruthScore int     `db:"user.truth_score"`
		UserShadow     bool    `db:"user.is_shadowbanned"`
		CommentID      uint64  `db:"comment.id"`
		CommentContent string  `db:"comment.content"`
		CommentUserID  uint64  `db:"comment.user_id"`
	}

	var rows []ReportRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	var reports []domain.CommentReport
	for _, row := range rows {
		rep := row.CommentReport
		rep.IsAccepted = row.IsAccepted
		rep.User = &domain.User{
			ID:             row.UserID,
			Name:           row.UserName,
			Slug:           row.UserSlug,
			TruthScore:     row.UserTruthScore,
			IsShadowbanned: row.UserShadow,
		}
		rep.Comment = &domain.Comment{
			ID:      row.CommentID,
			Content: row.CommentContent,
			UserID:  row.CommentUserID,
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
		       u.truth_score as "user.truth_score", u.is_shadowbanned as "user.is_shadowbanned",
		       c.id as "comment.id", c.content as "comment.content", c.user_id as "comment.user_id"
		FROM comment_reports r
		JOIN users u ON r.user_id = u.id
		JOIN comments c ON r.comment_id = c.id
		WHERE r.id = $1
	`

	type ReportRow struct {
		IsAccepted bool `db:"is_accepted"`
		domain.CommentReport
		UserID         uint64  `db:"user.id"`
		UserName       string  `db:"user.name"`
		UserSlug       *string `db:"user.slug"`
		UserTruthScore int     `db:"user.truth_score"`
		UserShadow     bool    `db:"user.is_shadowbanned"`
		CommentID      uint64  `db:"comment.id"`
		CommentContent string  `db:"comment.content"`
		CommentUserID  uint64  `db:"comment.user_id"`
	}

	var row ReportRow
	err := r.db.GetContext(ctx, &row, query, reportID)
	if err != nil {
		return nil, err
	}

	rep := row.CommentReport
	rep.IsAccepted = row.IsAccepted
	rep.User = &domain.User{
		ID:             row.UserID,
		Name:           row.UserName,
		Slug:           row.UserSlug,
		TruthScore:     row.UserTruthScore,
		IsShadowbanned: row.UserShadow,
	}
	rep.Comment = &domain.Comment{
		ID:      row.CommentID,
		Content: row.CommentContent,
			UserID:  row.CommentUserID,
	}

	return &rep, nil
}

func (r *moderationRepository) ResolveCommentReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	query := "UPDATE comment_reports SET status = true, is_accepted = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	res, err := r.db.ExecContext(ctx, query, isAccepted, reportID)
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

// ---- User Reports (CRUD) ----

func (r *moderationRepository) CreateUserReport(ctx context.Context, report *domain.UserReport) error {
	query := `
		INSERT INTO user_reports (reported_user_id, reporter_user_id, source, reason, content, status, created_at, updated_at) 
		VALUES (:reported_user_id, :reporter_user_id, :source, :reason, :content, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, report).Scan(&report.ID)
	return err
}

func (r *moderationRepository) IsUserReportedByReporter(ctx context.Context, reporterID, reportedID uint64) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM user_reports WHERE reporter_user_id = $1 AND reported_user_id = $2 AND status = false"
	err := r.db.GetContext(ctx, &count, query, reporterID, reportedID)
	return count > 0, err
}

func (r *moderationRepository) GetUserReports(ctx context.Context, status *bool, limit, offset int) ([]domain.UserReport, error) {
	query := `
		SELECT r.*, 
		       u1.id as "reported.id", u1.name as "reported.name", u1.slug as "reported.slug", u1.avatar as "reported.avatar",
		       u1.truth_score as "reported.truth_score", u1.is_shadowbanned as "reported.is_shadowbanned",
		       u2.id as "reporter.id", u2.name as "reporter.name", u2.slug as "reporter.slug", u2.avatar as "reporter.avatar",
		       u2.truth_score as "reporter.truth_score", u2.is_shadowbanned as "reporter.is_shadowbanned"
		FROM user_reports r
		JOIN users u1 ON r.reported_user_id = u1.id
		JOIN users u2 ON r.reporter_user_id = u2.id
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
		IsAccepted bool `db:"is_accepted"`
		domain.UserReport
		ReportedID     uint64  `db:"reported.id"`
		ReportedName   string  `db:"reported.name"`
		ReportedSlug   *string `db:"reported.slug"`
		ReportedAvatar *string `db:"reported.avatar"`
		ReportedScore  int     `db:"reported.truth_score"`
		ReportedShadow bool    `db:"reported.is_shadowbanned"`
		ReporterID     uint64  `db:"reporter.id"`
		ReporterName   string  `db:"reporter.name"`
		ReporterSlug   *string `db:"reporter.slug"`
		ReporterAvatar *string `db:"reporter.avatar"`
		ReporterScore  int     `db:"reporter.truth_score"`
		ReporterShadow bool    `db:"reporter.is_shadowbanned"`
	}

	var rows []ReportRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	var reports []domain.UserReport
	for _, row := range rows {
		rep := row.UserReport
		rep.IsAccepted = row.IsAccepted
		rep.ReportedUser = &domain.User{
			ID:             row.ReportedID,
			Name:           row.ReportedName,
			Slug:           row.ReportedSlug,
			Avatar:         row.ReportedAvatar,
			TruthScore:     row.ReportedScore,
			IsShadowbanned: row.ReportedShadow,
		}
		rep.ReporterUser = &domain.User{
			ID:             row.ReporterID,
			Name:           row.ReporterName,
			Slug:           row.ReporterSlug,
			Avatar:         row.ReporterAvatar,
			TruthScore:     row.ReporterScore,
			IsShadowbanned: row.ReporterShadow,
		}
		reports = append(reports, rep)
	}

	if reports == nil {
		reports = []domain.UserReport{}
	}

	return reports, nil
}

func (r *moderationRepository) GetUserReport(ctx context.Context, reportID uint64) (*domain.UserReport, error) {
	query := `
		SELECT r.*, 
		       u1.id as "reported.id", u1.name as "reported.name", u1.slug as "reported.slug", u1.avatar as "reported.avatar",
		       u1.truth_score as "reported.truth_score", u1.is_shadowbanned as "reported.is_shadowbanned",
		       u2.id as "reporter.id", u2.name as "reporter.name", u2.slug as "reporter.slug", u2.avatar as "reporter.avatar",
		       u2.truth_score as "reporter.truth_score", u2.is_shadowbanned as "reporter.is_shadowbanned"
		FROM user_reports r
		JOIN users u1 ON r.reported_user_id = u1.id
		JOIN users u2 ON r.reporter_user_id = u2.id
		WHERE r.id = $1
	`

	type ReportRow struct {
		IsAccepted bool `db:"is_accepted"`
		domain.UserReport
		ReportedID     uint64  `db:"reported.id"`
		ReportedName   string  `db:"reported.name"`
		ReportedSlug   *string `db:"reported.slug"`
		ReportedAvatar *string `db:"reported.avatar"`
		ReportedScore  int     `db:"reported.truth_score"`
		ReportedShadow bool    `db:"reported.is_shadowbanned"`
		ReporterID     uint64  `db:"reporter.id"`
		ReporterName   string  `db:"reporter.name"`
		ReporterSlug   *string `db:"reporter.slug"`
		ReporterAvatar *string `db:"reporter.avatar"`
		ReporterScore  int     `db:"reporter.truth_score"`
		ReporterShadow bool    `db:"reporter.is_shadowbanned"`
	}

	var row ReportRow
	err := r.db.GetContext(ctx, &row, query, reportID)
	if err != nil {
		return nil, err
	}

	rep := row.UserReport
	rep.IsAccepted = row.IsAccepted
	rep.ReportedUser = &domain.User{
		ID:             row.ReportedID,
		Name:           row.ReportedName,
		Slug:           row.ReportedSlug,
		Avatar:         row.ReportedAvatar,
		TruthScore:     row.ReportedScore,
		IsShadowbanned: row.ReportedShadow,
	}
	rep.ReporterUser = &domain.User{
		ID:             row.ReporterID,
		Name:           row.ReporterName,
		Slug:           row.ReporterSlug,
		Avatar:         row.ReporterAvatar,
		TruthScore:     row.ReporterScore,
		IsShadowbanned: row.ReporterShadow,
	}

	return &rep, nil
}

func (r *moderationRepository) ResolveUserReport(ctx context.Context, reportID uint64, isAccepted bool) error {
	query := "UPDATE user_reports SET status = true, is_accepted = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2"
	res, err := r.db.ExecContext(ctx, query, isAccepted, reportID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("report not found or already resolved")
	}

	return err
}

func (r *moderationRepository) DeleteUserReport(ctx context.Context, reportID uint64) error {
	query := "DELETE FROM user_reports WHERE id = $1"
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

// Shadowban & Truth Score Management

func (r *moderationRepository) ShadowbanUser(ctx context.Context, userID uint64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update user
	_, err = tx.ExecContext(ctx, "UPDATE users SET is_shadowbanned = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1", userID)
	if err != nil {
		return err
	}

	// 2. Shadowban all existing interactions
	_, err = tx.ExecContext(ctx, "UPDATE comments SET is_shadowbanned = true WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE song_ratings SET is_shadowbanned = true WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE song_reactions SET is_shadowbanned = true WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE comment_reactions SET is_shadowbanned = true WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *moderationRepository) UnshadowbanUser(ctx context.Context, userID uint64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update user
	_, err = tx.ExecContext(ctx, "UPDATE users SET is_shadowbanned = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1", userID)
	if err != nil {
		return err
	}

	// 2. Unshadowban all interactions
	_, err = tx.ExecContext(ctx, "UPDATE comments SET is_shadowbanned = false WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE song_ratings SET is_shadowbanned = false WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE song_reactions SET is_shadowbanned = false WHERE user_id = $1", userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE comment_reactions SET is_shadowbanned = false WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *moderationRepository) SetCommentShadowban(ctx context.Context, commentID uint64, isShadowbanned bool) error {
	_, err := r.db.ExecContext(ctx, "UPDATE comments SET is_shadowbanned = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", isShadowbanned, commentID)
	return err
}

func (r *moderationRepository) SetRatingShadowban(ctx context.Context, ratingID uint64, isShadowbanned bool) error {
	_, err := r.db.ExecContext(ctx, "UPDATE song_ratings SET is_shadowbanned = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", isShadowbanned, ratingID)
	return err
}

func (r *moderationRepository) UpdateUserTruthScore(ctx context.Context, userID uint64, delta int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET truth_score = truth_score + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", delta, userID)
	return err
}

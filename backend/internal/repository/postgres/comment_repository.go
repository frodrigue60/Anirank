package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type commentRepository struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) domain.CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) GetByEntity(ctx context.Context, userID *uint64, entityID uint64, entityType string, limit, offset int) ([]domain.Comment, error) {
	var comments []domain.Comment

	// Load parent comments, sorting by latest. We also eager load user details simply using a JOIN.
	// Since we are mapping joined columns to a nested struct "User", sqlx can struggle without specific tags
	// or specific structs if the User struct isn't perfectly mapped. We'll use a local struct for scanning
	// and map to domain manual for 100% safety.
	type commentWithUser struct {
		ID         uint64    `db:"id"`
		ParentID   *uint64   `db:"parent_id"`
		SongID     *uint64   `db:"song_id"`
		UserID     uint64    `db:"user_id"`
		Content    string    `db:"content"`
		Created_At time.Time `db:"created_at"`
		Updated_At time.Time `db:"updated_at"`
		// User fields
		UID          uint64  `db:"uid"`
		UName        string  `db:"uname"`
		USlug        *string `db:"uslug"`
		UScore       *string `db:"uscore"`
		UAvatar      *string `db:"uavatar"`
		RCount       int     `db:"rcount"`
		Likes        int     `db:"likes_count"`
		Dislikes     int     `db:"dislikes_count"`
		UserReaction int8    `db:"user_reaction"`
	}

	var querySafe string
	var args []interface{}

	if userID != nil {
		querySafe = `
			SELECT 
				c.id, c.parent_id, c.song_id, c.user_id, c.content, c.likes_count, c.dislikes_count, c.created_at, c.updated_at,
				u.id as uid, u.name as uname, u.slug as uslug, u.avatar as uavatar, sf.slug as uscore,
				(SELECT COUNT(*) FROM comments WHERE parent_id = c.id) as rcount,
				COALESCE(cr.type, 0) as user_reaction
			FROM comments c
			JOIN users u ON c.user_id = u.id
			LEFT JOIN score_formats sf ON u.score_format_id = sf.id
			LEFT JOIN comment_reactions cr ON c.id = cr.comment_id AND cr.user_id = $1
			WHERE c.song_id = $2 AND c.parent_id IS NULL
			ORDER BY c.created_at DESC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{*userID, entityID, limit, offset}
	} else {
		querySafe = `
			SELECT 
				c.id, c.parent_id, c.song_id, c.user_id, c.content, c.likes_count, c.dislikes_count, c.created_at, c.updated_at,
				u.id as uid, u.name as uname, u.slug as uslug, u.avatar as uavatar, sf.slug as uscore,
				(SELECT COUNT(*) FROM comments WHERE parent_id = c.id) as rcount,
				0 as user_reaction
			FROM comments c
			JOIN users u ON c.user_id = u.id
			LEFT JOIN score_formats sf ON u.score_format_id = sf.id
			WHERE c.song_id = $1 AND c.parent_id IS NULL
			ORDER BY c.created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{entityID, limit, offset}
	}

	var rows []commentWithUser
	err := r.db.SelectContext(ctx, &rows, querySafe, args...)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		c := domain.Comment{
			ID:            row.ID,
			ParentID:      row.ParentID,
			SongID:        row.SongID,
			UserID:        row.UserID,
			Content:       row.Content,
			LikesCount:    row.Likes,
			DislikesCount: row.Dislikes,
			Created_At:    row.Created_At,
			Updated_At:    row.Updated_At,
			RepliesCount:  row.RCount,
			IsLiked:       row.UserReaction == 1,
			IsDisliked:    row.UserReaction == -1,
		}

		c.User = &domain.User{
			ID:          row.UID,
			Name:        row.UName,
			Slug:        row.USlug,
			Avatar:      row.UAvatar,
			ScoreFormat: row.UScore,
		}

		comments = append(comments, c)
	}

	return comments, nil
}

func (r *commentRepository) GetReplies(ctx context.Context, userID *uint64, parentID uint64, limit, offset int) ([]domain.Comment, error) {
	var comments []domain.Comment

	type commentWithUser struct {
		domain.Comment
		LikesCount    int     `db:"likes_count"`
		DislikesCount int     `db:"dislikes_count"`
		UserID        uint64  `db:"uid"`
		UserName      string  `db:"uname"`
		UserSlug      *string `db:"uslug"`
		UserAvatar    *string `db:"uavatar"`
		UserScoreFmt  *string `db:"uscore"`
		UserReaction  int8    `db:"user_reaction"`
	}

	var querySafe string
	var args []interface{}

	if userID != nil {
		querySafe = `
			SELECT 
				c.*,
				c.likes_count,
				c.dislikes_count,
				u.id as uid, u.name as uname, u.slug as uslug, u.avatar as uavatar, sf.slug as uscore,
				COALESCE(cr.type, 0) as user_reaction
			FROM comments c
			JOIN users u ON c.user_id = u.id
			LEFT JOIN score_formats sf ON u.score_format_id = sf.id
			LEFT JOIN comment_reactions cr ON c.id = cr.comment_id AND cr.user_id = $1
			WHERE c.parent_id = $2
			ORDER BY c.created_at ASC
			LIMIT $3 OFFSET $4
		`
		args = []interface{}{*userID, parentID, limit, offset}
	} else {
		querySafe = `
			SELECT 
				c.*,
				c.likes_count,
				c.dislikes_count,
				u.id as uid, u.name as uname, u.slug as uslug, u.avatar as uavatar, sf.slug as uscore,
				0 as user_reaction
			FROM comments c
			JOIN users u ON c.user_id = u.id
			LEFT JOIN score_formats sf ON u.score_format_id = sf.id
			WHERE c.parent_id = $1
			ORDER BY c.created_at ASC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{parentID, limit, offset}
	}

	var rows []commentWithUser
	err := r.db.SelectContext(ctx, &rows, querySafe, args...)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		c := row.Comment

		c.User = &domain.User{
			ID:          row.UserID,
			Name:        row.UserName,
			Slug:        row.UserSlug,
			Avatar:      row.UserAvatar,
			ScoreFormat: row.UserScoreFmt,
		}

		// Map timestamps and counters explicitly
		c.Created_At = row.Comment.Created_At
		c.Updated_At = row.Comment.Updated_At
		c.LikesCount = row.LikesCount
		c.DislikesCount = row.DislikesCount
		c.IsLiked = row.UserReaction == 1
		c.IsDisliked = row.UserReaction == -1

		comments = append(comments, c)
	}

	return comments, nil
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	query := `
		INSERT INTO comments (parent_id, song_id, user_id, content, created_at, updated_at)
		VALUES (:parent_id, :song_id, :user_id, :content, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	err = stmt.QueryRowContext(ctx, comment).Scan(&comment.ID, &comment.Created_At, &comment.Updated_At)
	return err
}

func (r *commentRepository) Delete(ctx context.Context, id, userID uint64) error {
	query := "DELETE FROM comments WHERE id = $1 AND user_id = $2"
	res, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return errors.New("unauthorized or comment not found")
	}

	return err
}

func (r *commentRepository) GetByID(ctx context.Context, id uint64) (*domain.Comment, error) {
	var c domain.Comment
	query := "SELECT * FROM comments WHERE id = $1"
	err := r.db.GetContext(ctx, &c, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *commentRepository) GetCount(ctx context.Context, songID uint64) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM comments WHERE song_id = $1 AND parent_id IS NULL"
	err := r.db.GetContext(ctx, &count, query, songID)
	return count, err
}

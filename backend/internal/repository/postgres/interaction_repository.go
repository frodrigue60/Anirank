package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/jmoiron/sqlx"
)

type interactionRepository struct {
	db *sqlx.DB
}

func NewInteractionRepository(db *sqlx.DB) domain.InteractionRepository {
	return &interactionRepository{db: db.Unsafe()}
}

// Ratings
func (r *interactionRepository) UpsertRating(ctx context.Context, rating *domain.Rating) error {
	query := `
		INSERT INTO song_ratings (rating, song_id, user_id, created_at, updated_at) 
		VALUES (:rating, :song_id, :user_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (song_id, user_id) DO UPDATE SET 
			rating = EXCLUDED.rating, 
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.NamedExecContext(ctx, query, rating)
	return err
}

func (r *interactionRepository) GetRatingByUser(ctx context.Context, userID, songID uint64) (*domain.Rating, error) {
	var rating domain.Rating
	query := "SELECT * FROM song_ratings WHERE user_id = $1 AND song_id = $2"
	err := r.db.GetContext(ctx, &rating, query, userID, songID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &rating, err
}

func (r *interactionRepository) GetAverageRating(ctx context.Context, songID uint64) (float64, error) {
	var avg sql.NullFloat64
	query := "SELECT AVG(rating) FROM song_ratings WHERE song_id = $1"
	err := r.db.GetContext(ctx, &avg, query, songID)
	if err != nil {
		return 0, err
	}
	if !avg.Valid {
		return 0, nil
	}
	return avg.Float64, nil
}

// Favorites
func (r *interactionRepository) ToggleFavorite(ctx context.Context, favorite *domain.Favorite) (bool, error) {
	var table string
	var entityCol string

	switch favorite.FavoritableType {
	case domain.TypeArtist:
		table = "artist_user"
		entityCol = "artist_id"
	case domain.TypeSong:
		table = "song_user"
		entityCol = "song_id"
	default:
		return false, fmt.Errorf("unsupported favoritable type: %s", favorite.FavoritableType)
	}

	// Check if exists first
	var id uint64
	checkQuery := "SELECT id FROM " + table + " WHERE user_id = $1 AND " + entityCol + " = $2"
	err := r.db.GetContext(ctx, &id, checkQuery, favorite.UserID, favorite.FavoritableID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	if err == nil {
		// Existing favorite found, so delete it (toggle off)
		deleteQuery := "DELETE FROM " + table + " WHERE id = $1"
		_, err = r.db.ExecContext(ctx, deleteQuery, id)
		return false, err
	}

	// Not found, create it
	insertQuery := fmt.Sprintf(`
		INSERT INTO %s (user_id, %s, created_at, updated_at)
		VALUES (:user_id, :favoritable_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, table, entityCol)

	_, err = r.db.NamedExecContext(ctx, insertQuery, favorite)
	return true, err
}

func (r *interactionRepository) IsFavoritedByUser(ctx context.Context, userID, entityID uint64, entityType string) (bool, error) {
	var table string
	var entityCol string

	switch entityType {
	case domain.TypeArtist:
		table = "artist_user"
		entityCol = "artist_id"
	case domain.TypeSong:
		table = "song_user"
		entityCol = "song_id"
	default:
		return false, fmt.Errorf("unsupported favoritable type: %s", entityType)
	}

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM " + table + " WHERE user_id = $1 AND " + entityCol + " = $2)"
	err := r.db.GetContext(ctx, &exists, query, userID, entityID)
	return exists, err
}

func (r *interactionRepository) UpsertSongReaction(ctx context.Context, userID, songID uint64, reactionType int8) (int, int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	// 1. Upsert reaction and get its previously existing type
	var existing int8 = 0
	var reactionID uint64

	// We use a single query to UPSERT or at least determine what happened.
	// Actually, toggling is hard with ON CONFLICT because we might want to DELETE.
	// A better way is to SELECT FOR UPDATE to lock the row.
	err = tx.QueryRowContext(ctx, "SELECT id, type FROM song_reactions WHERE user_id = $1 AND song_id = $2 FOR UPDATE", userID, songID).Scan(&reactionID, &existing)
	
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, err
	}

	// 2. Logic to update reactions
	if err == nil && existing == reactionType {
		// Toggle OFF
		_, err = tx.ExecContext(ctx, "DELETE FROM song_reactions WHERE id = $1", reactionID)
		if err != nil {
			return 0, 0, err
		}
	} else if err == nil {
		// Change type
		_, err = tx.ExecContext(ctx, "UPDATE song_reactions SET type = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", reactionType, reactionID)
		if err != nil {
			return 0, 0, err
		}
	} else {
		// New reaction
		_, err = tx.ExecContext(ctx, "INSERT INTO song_reactions (user_id, song_id, type, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", userID, songID, reactionType)
		if err != nil {
			return 0, 0, err
		}
	}

	// 3. Update song counts using exact counts from the reactions table
	var newLikes, newDislikes int
	updateQuery := `
		UPDATE songs 
		SET likes_count = (SELECT count(*) FROM song_reactions WHERE song_id = $1 AND type = 1),
		    dislikes_count = (SELECT count(*) FROM song_reactions WHERE song_id = $1 AND type = -1),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING likes_count, dislikes_count
	`
	err = tx.QueryRowContext(ctx, updateQuery, songID).Scan(&newLikes, &newDislikes)
	if err != nil {
		return 0, 0, err
	}


	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return newLikes, newDislikes, nil
}

func (r *interactionRepository) UpsertCommentReaction(ctx context.Context, userID, commentID uint64, reactionType int8) (int, int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	// 1. Check if user already reacted (with lock)
	var existing int8
	var reactionID uint64
	err = tx.QueryRowContext(ctx, "SELECT id, type FROM comment_reactions WHERE user_id = $1 AND comment_id = $2 FOR UPDATE", userID, commentID).Scan(&reactionID, &existing)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, err
	}

	// 2. Logic
	if err == nil && reactionType == existing {
		// Toggle OFF (remove reaction)
		_, err = tx.ExecContext(ctx, "DELETE FROM comment_reactions WHERE id = $1", reactionID)
		if err != nil {
			return 0, 0, err
		}
	} else if err == nil {
		// Change type
		_, err = tx.ExecContext(ctx, "UPDATE comment_reactions SET type = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", reactionType, reactionID)
		if err != nil {
			return 0, 0, err
		}
	} else {
		// New reaction
		_, err = tx.ExecContext(ctx, "INSERT INTO comment_reactions (user_id, comment_id, type, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", userID, commentID, reactionType)
		if err != nil {
			return 0, 0, err
		}
	}

	// 3. Update comment counts using exact counts
	var newLikes, newDislikes int
	updateQuery := `
		UPDATE comments 
		SET likes_count = (SELECT count(*) FROM comment_reactions WHERE comment_id = $1 AND type = 1),
		    dislikes_count = (SELECT count(*) FROM comment_reactions WHERE comment_id = $1 AND type = -1),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING likes_count, dislikes_count
	`
	err = tx.QueryRowContext(ctx, updateQuery, commentID).Scan(&newLikes, &newDislikes)
	if err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return newLikes, newDislikes, nil
}

// Reactions (Requires TX logic for atomicity modifying parent counters along with the pivot table)
func (r *interactionRepository) ToggleReaction(ctx context.Context, reaction *domain.Reaction) error {
	// Rest of old implementation or just leave it for now if not used for songs
	return nil
}
func (r *interactionRepository) GetReactionByUser(ctx context.Context, userID, entityID uint64, entityType string) (*domain.Reaction, error) {
	var reaction domain.Reaction
	query := "SELECT id, user_id, song_id AS reactable_id, type, created_at, updated_at FROM song_reactions WHERE user_id = $1 AND song_id = $2"
	err := r.db.GetContext(ctx, &reaction, query, userID, entityID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &reaction, err
}

func (r *interactionRepository) GetCounters(ctx context.Context, entityID uint64, entityType string) (*domain.ReactionCounter, error) {
	var counter domain.ReactionCounter
	var table string

	if entityType == domain.TypeSong {
		table = "songs"
	} else if entityType == domain.TypeComment {
		table = "comments"
	} else {
		// If entityType is unknown, return zero counters
		return &domain.ReactionCounter{
			ReactableID:   entityID,
			ReactableType: entityType,
			LikesCount:    0,
			DislikesCount: 0,
		}, nil
	}

	query := "SELECT id as reactable_id, $1 as reactable_type, likes_count, dislikes_count FROM " + table + " WHERE id = $2"
	err := r.db.GetContext(ctx, &counter, query, entityType, entityID)

	// If it doesn't exist, we just return empty counters rather than ErrNotFound
	if errors.Is(err, sql.ErrNoRows) {
		return &domain.ReactionCounter{
			ReactableID:   entityID,
			ReactableType: entityType,
			LikesCount:    0,
			DislikesCount: 0,
		}, nil
	}
	return &counter, err
}

type activityRow struct {
	ActivityType string         `db:"activity_type"`
	ActivityID   uint64         `db:"activity_id"`
	UserID       uint64         `db:"user_id"`
	UserName     string         `db:"user_name"`
	TargetID     uint64         `db:"target_id"`
	TargetType   string         `db:"target_type"`
	Value        sql.NullString `db:"value"` // Might be a string containing rating num, or comment text
	CreatedAt    string         `db:"created_at"`
}

// Activity Feed (UNION ALL query to combine ratings, favorites and comments)
func (r *interactionRepository) GetRecentActivities(ctx context.Context, limit int) ([]domain.ActivityItem, error) {
	query := `
		SELECT * FROM (
			SELECT 'rating' as activity_type, r.id as activity_id, r.user_id, u.name as user_name, r.song_id as target_id, 'song' as target_type, r.rating::text as value, r.created_at
			FROM song_ratings r JOIN users u ON r.user_id = u.id
			UNION ALL
			SELECT 'favorite' as activity_type, f.id as activity_id, f.user_id, u.name as user_name, f.song_id as target_id, 'song' as target_type, NULL as value, f.created_at
			FROM song_user f JOIN users u ON f.user_id = u.id
			UNION ALL
			SELECT 'comment' as activity_type, c.id as activity_id, c.user_id, u.name as user_name, c.song_id as target_id, 'song' as target_type, c.content as value, c.created_at
			FROM comments c JOIN users u ON c.user_id = u.id
		) as activities
		ORDER BY created_at DESC
		LIMIT $1
	`

	var rows []activityRow
	err := r.db.SelectContext(ctx, &rows, query, limit)
	if err != nil {
		return nil, err
	}

	// Map to Domain Output
	var feed []domain.ActivityItem
	for _, row := range rows {
		item := domain.ActivityItem{
			Type:     row.ActivityType,
			UserID:   row.UserID,
			TargetID: row.TargetID,
			User: domain.User{
				ID:   row.UserID,
				Name: row.UserName,
			},
		}

		if row.Value.Valid {
			item.Value = row.Value.String
		}

		feed = append(feed, item)
	}

	return feed, nil
}

package domain

import (
	"context"
	"time"
)

// Supported polymorphic types
const (
	TypeSong    = "song"
	TypeArtist  = "artist"
	TypeComment = "comment"
	TypeAnime   = "anime"
)

// Interactions & Polymorphic Emulations

type Rating struct {
	ID            uint64    `db:"id" json:"id"`
	Rating        float64   `db:"rating" json:"rating"`
	SongID        uint64    `db:"song_id" json:"song_id"`
	UserID        uint64    `db:"user_id" json:"user_id"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type Reaction struct {
	ID            uint64    `db:"id" json:"id"`
	UserID        uint64    `db:"user_id" json:"user_id"`
	ReactableID   uint64    `db:"song_id" json:"reactable_id"`
	ReactableType string    `db:"reactable_type" json:"reactable_type"`
	Type          int8      `db:"type" json:"type"` // 1 = like, -1 = dislike
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type SongReaction struct {
	ID        uint64    `db:"id" json:"id"`
	UserID    uint64    `db:"user_id" json:"user_id"`
	SongID    uint64    `db:"song_id" json:"song_id"`
	Type      int8      `db:"type" json:"type"` // 1 = like, -1 = dislike
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ReactionCounter struct {
	ID            uint64    `db:"id" json:"id"`
	ReactableID   uint64    `db:"reactable_id" json:"reactable_id"`
	ReactableType string    `db:"reactable_type" json:"reactable_type"`
	LikesCount    uint64    `db:"likes_count" json:"likes_count"`
	DislikesCount uint64    `db:"dislikes_count" json:"dislikes_count"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type Favorite struct {
	ID              uint64    `db:"id" json:"id"`
	UserID          uint64    `db:"user_id" json:"user_id"`
	FavoritableID   uint64    `db:"favoritable_id" json:"favoritable_id"`
	FavoritableType string    `db:"favoritable_type" json:"favoritable_type"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type UserSongInteraction struct {
	SongID      uint64
	IsFavorited bool
	Reaction    int8     // 1 for like, -1 for dislike, 0 for none
	Rating      *float64
}

type ActivityItem struct {
	Type       string      `json:"type"` // "rating", "favorite", "comment"
	UserID     uint64      `json:"user_id"`
	User       User        `json:"user"`
	TargetID   uint64      `json:"target_id"`
	TargetType string      `json:"target_type"`
	Target     interface{} `json:"target"`          // The Anime, Song or Comment
	Value      interface{} `json:"value,omitempty"` // Example: Rating score
	CreatedAt  string      `json:"created_at"`      // Returning raw string from MySQL
}

type Comment struct {
	ID         uint64    `db:"id" json:"id"`
	UUID       string    `db:"uuid" json:"uuid"`
	ParentID   *uint64   `db:"parent_id" json:"parent_id"`
	SongID     *uint64   `db:"song_id" json:"song_id,omitempty"`
	UserID     uint64    `db:"user_id" json:"user_id"`
	Content    string    `db:"content" json:"content"`
	Created_At time.Time `db:"created_at" json:"created_at"`
	Updated_At time.Time `db:"updated_at" json:"updated_at"`

	// Relational Output Data
	User          *User     `db:"-" json:"user,omitempty"`
	RepliesCount  int       `db:"-" json:"replies_count,omitempty"`
	Replies       []Comment `db:"-" json:"replies,omitempty"`
	LikesCount    int       `db:"likes_count" json:"likes_count"`
	DislikesCount int       `db:"dislikes_count" json:"dislikes_count"`
	IsLiked       bool      `db:"-" json:"is_liked,omitempty"`
	IsDisliked    bool      `db:"-" json:"is_disliked,omitempty"`
}

// Repositories
type InteractionRepository interface {
	// Ratings
	UpsertRating(ctx context.Context, rating *Rating) error
	GetRatingByUser(ctx context.Context, userID, songID uint64) (*Rating, error)
	GetAverageRating(ctx context.Context, songID uint64) (float64, error)
	GetAverageRatingsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64]float64, error)
	GetUserInteractionsBySongIDs(ctx context.Context, userID uint64, songIDs []uint64) (map[uint64]UserSongInteraction, error)
	CountRatingsByUser(ctx context.Context, userID uint64) (int, error) // For automatic badges

	// Reactions (Likes & Dislikes)
	ToggleReaction(ctx context.Context, reaction *Reaction) error // Executes TX to increment/decrement Counter too.
	UpsertSongReaction(ctx context.Context, userID, songID uint64, reactionType int8) (likesCount, dislikesCount int, err error)
	UpsertCommentReaction(ctx context.Context, userID, commentID uint64, reactionType int8) (likesCount, dislikesCount int, err error)
	GetReactionByUser(ctx context.Context, userID, entityID uint64, entityType string) (*Reaction, error)
	GetCounters(ctx context.Context, entityID uint64, entityType string) (*ReactionCounter, error)

	// Favorites
	ToggleFavorite(ctx context.Context, favorite *Favorite) (bool, error) // Returns true if added, false if removed
	IsFavoritedByUser(ctx context.Context, userID, entityID uint64, entityType string) (bool, error)

	// Activity Feed
	GetRecentActivities(ctx context.Context, limit int) ([]ActivityItem, error)

	// Notifications Helpers
	GetUsersWhoFavorited(ctx context.Context, entityID uint64, entityType string) ([]uint64, error)
}

type CommentRepository interface {
	GetByEntity(ctx context.Context, userID *uint64, entityID uint64, entityType string, limit, offset int) ([]Comment, error)
	GetReplies(ctx context.Context, userID *uint64, parentID uint64, limit, offset int) ([]Comment, error)
	GetByID(ctx context.Context, id uint64) (*Comment, error)
	GetByUUID(ctx context.Context, uuid string) (*Comment, error)
	GetCountByEntity(ctx context.Context, entityID uint64, entityType string) (int, error)
	GetCountByUser(ctx context.Context, userID uint64) (int, error) // For automatic badges
	GetRepliesCount(ctx context.Context, parentID uint64) (int, error)
	GetCount(ctx context.Context, songID uint64) (int, error)
	Create(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id, userID uint64) error // Authorization baked in via userID verification
}

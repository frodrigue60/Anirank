package domain

import (
	"context"
	"time"
)

// SongReport represents a user report against a specific song.
type SongReport struct {
	ID        uint64    `db:"id" json:"id"`
	SongID    uint64    `db:"song_id" json:"song_id"`
	UserID    uint64    `db:"user_id" json:"user_id"`
	Source    string    `db:"source" json:"source"` // E.g., 'web', 'app', 'ext'
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	Status    bool      `db:"status" json:"status"` // false: pending, true: fixed
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relational
	Song *Song `db:"-" json:"song,omitempty"`
	User *User `db:"-" json:"user,omitempty"`
}

// CommentReport represents a user report against a specific comment.
type CommentReport struct {
	ID        uint64    `db:"id" json:"id"`
	CommentID uint64    `db:"comment_id" json:"comment_id"`
	UserID    uint64    `db:"user_id" json:"user_id"`
	Source    string    `db:"source" json:"source"` // E.g., 'web', 'app', 'ext'
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	Status    bool      `db:"status" json:"status"` // false: pending, true: fixed
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relational
	Comment *Comment `db:"-" json:"comment,omitempty"`
	User    *User    `db:"-" json:"user,omitempty"`
}

// UserRequest represents a generic request or ticket created by a user.
type UserRequest struct {
	ID         uint64    `db:"id" json:"id"`
	Title      string    `db:"title" json:"title"`
	Content    string    `db:"content" json:"content"`
	UserID     uint64    `db:"user_id" json:"user_id"`
	AttendedBy *uint64   `db:"attended_by" json:"attended_by"` // nullable
	Status     bool      `db:"status" json:"status"`           // false: pending, true: attended
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`

	// Relational
	User          *User `db:"-" json:"user,omitempty"`
	AttendedAdmin *User `db:"-" json:"attended_admin,omitempty"`
}

type ModerationRepository interface {
	// User Facing
	CreateSongReport(ctx context.Context, report *SongReport) error
	IsSongReportedByUser(ctx context.Context, userID, songID uint64) (bool, error)
	CreateCommentReport(ctx context.Context, report *CommentReport) error
	CreateUserRequest(ctx context.Context, request *UserRequest) error

	// Admin Facing
	GetSongReports(ctx context.Context, status *bool, limit, offset int) ([]SongReport, error)
	GetSongReport(ctx context.Context, reportID uint64) (*SongReport, error)
	ResolveSongReport(ctx context.Context, reportID uint64) error
	DeleteSongReport(ctx context.Context, reportID uint64) error

	GetCommentReports(ctx context.Context, status *bool, limit, offset int) ([]CommentReport, error)
	GetCommentReport(ctx context.Context, reportID uint64) (*CommentReport, error)
	ResolveCommentReport(ctx context.Context, reportID uint64) error
	DeleteCommentReport(ctx context.Context, reportID uint64) error

	GetUserRequests(ctx context.Context, status *bool, limit, offset int) ([]UserRequest, error)
	GetUserRequest(ctx context.Context, requestID uint64) (*UserRequest, error)
	UpdateUserRequestStatus(ctx context.Context, requestID uint64, status bool, adminID uint64) error
	DeleteUserRequest(ctx context.Context, requestID uint64) error
}

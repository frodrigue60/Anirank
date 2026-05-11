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
	Status     bool      `db:"status" json:"status"` // false: pending, true: fixed
	IsAccepted bool      `db:"is_accepted" json:"is_accepted"`
	Snapshot   *string   `db:"snapshot" json:"snapshot,omitempty"`
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
	Status     bool      `db:"status" json:"status"` // false: pending, true: fixed
	IsAccepted bool      `db:"is_accepted" json:"is_accepted"`
	Snapshot   *string   `db:"snapshot" json:"snapshot,omitempty"`
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

// UserReport represents a user report against another user.
type UserReport struct {
	ID             uint64    `db:"id" json:"id"`
	ReportedUserID uint64    `db:"reported_user_id" json:"reported_user_id"`
	ReporterUserID uint64    `db:"reporter_user_id" json:"reporter_user_id"`
	Source         string    `db:"source" json:"source"` // E.g., 'web', 'app', 'ext'
	Reason         string    `db:"reason" json:"reason"`
	Content        string    `db:"content" json:"content"`
	Status         bool      `db:"status" json:"status"` // false: pending, true: fixed
	IsAccepted     bool      `db:"is_accepted" json:"is_accepted"`
	Snapshot       *string   `db:"snapshot" json:"snapshot,omitempty"` // JSON string
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`

	// Relational
	ReportedUser *User `db:"-" json:"reported_user,omitempty"`
	ReporterUser *User `db:"-" json:"reporter_user,omitempty"`
}

type ModerationRepository interface {
	// User Facing
	CreateSongReport(ctx context.Context, report *SongReport) error
	IsSongReportedByUser(ctx context.Context, userID, songID uint64) (bool, error)
	GetSongReportsByUserAndSongIDs(ctx context.Context, userID uint64, songIDs []uint64) (map[uint64]bool, error)
	CreateCommentReport(ctx context.Context, report *CommentReport) error
	CreateUserRequest(ctx context.Context, request *UserRequest) error

	// Admin Facing
	GetSongReports(ctx context.Context, status *bool, limit, offset int) ([]SongReport, error)
	GetSongReport(ctx context.Context, reportID uint64) (*SongReport, error)
	ResolveSongReport(ctx context.Context, reportID uint64, isAccepted bool) error
	DeleteSongReport(ctx context.Context, reportID uint64) error

	GetCommentReports(ctx context.Context, status *bool, limit, offset int) ([]CommentReport, error)
	GetCommentReport(ctx context.Context, reportID uint64) (*CommentReport, error)
	ResolveCommentReport(ctx context.Context, reportID uint64, isAccepted bool) error
	DeleteCommentReport(ctx context.Context, reportID uint64) error

	GetUserRequests(ctx context.Context, status *bool, limit, offset int) ([]UserRequest, error)
	GetUserRequest(ctx context.Context, requestID uint64) (*UserRequest, error)
	UpdateUserRequestStatus(ctx context.Context, requestID uint64, status bool, adminID uint64) error
	DeleteUserRequest(ctx context.Context, requestID uint64) error

	// User Reports
	CreateUserReport(ctx context.Context, report *UserReport) error
	IsUserReportedByReporter(ctx context.Context, reporterID, reportedID uint64) (bool, error)
	GetUserReports(ctx context.Context, status *bool, limit, offset int) ([]UserReport, error)
	GetUserReport(ctx context.Context, reportID uint64) (*UserReport, error)
	ResolveUserReport(ctx context.Context, reportID uint64, isAccepted bool) error
	DeleteUserReport(ctx context.Context, reportID uint64) error

	// Shadowban & Truth Score Management
	ShadowbanUser(ctx context.Context, userID uint64) error
	UnshadowbanUser(ctx context.Context, userID uint64) error
	SetCommentShadowban(ctx context.Context, commentID uint64, isShadowbanned bool) error
	SetRatingShadowban(ctx context.Context, ratingID uint64, isShadowbanned bool) error
	UpdateUserTruthScore(ctx context.Context, userID uint64, delta int) error
	GetPendingReportsCount(ctx context.Context, userID uint64) (int, error)
	GetCommentReportsCountByTrustedUsers(ctx context.Context, commentID uint64, minScore int) (int, error)
	GetUserReportsCountByTrustedUsers(ctx context.Context, reportedUserID uint64, minScore int) (int, error)
}

type ModerationUsecase interface {
	CreateSongReport(ctx context.Context, userID uint64, req *SongReport) error
	CreateUserRequest(ctx context.Context, userID uint64, req *UserRequest) error
	CreateCommentReport(ctx context.Context, userID uint64, req *CommentReport) error
	GetSongReports(ctx context.Context, status *bool, limit, offset int) ([]SongReport, error)
	GetSongReport(ctx context.Context, reportID uint64) (*SongReport, error)
	ResolveSongReport(ctx context.Context, reportID uint64, isAccepted bool) error
	DeleteSongReport(ctx context.Context, reportID uint64) error
	GetCommentReports(ctx context.Context, status *bool, limit, offset int) ([]CommentReport, error)
	GetCommentReport(ctx context.Context, reportID uint64) (*CommentReport, error)
	ResolveCommentReport(ctx context.Context, reportID uint64, isAccepted bool) error
	DeleteCommentReport(ctx context.Context, reportID uint64) error
	GetUserRequests(ctx context.Context, status *bool, limit, offset int) ([]UserRequest, error)
	GetUserRequest(ctx context.Context, requestID uint64) (*UserRequest, error)
	UpdateUserRequestStatus(ctx context.Context, requestID uint64, status bool, adminID uint64) error
	DeleteUserRequest(ctx context.Context, requestID uint64) error
	CreateUserReport(ctx context.Context, userID uint64, req *UserReport) error
	GetUserReports(ctx context.Context, status *bool, limit, offset int) ([]UserReport, error)
	GetUserReport(ctx context.Context, reportID uint64) (*UserReport, error)
	ResolveUserReport(ctx context.Context, reportID uint64, isAccepted bool) error
	DeleteUserReport(ctx context.Context, reportID uint64) error
	ValidateInteraction(ctx context.Context, userID uint64, content string) (bool, error)
}

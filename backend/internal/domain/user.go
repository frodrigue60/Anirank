package domain

import (
	"context"
	"time"
)

type User struct {
	ID              uint64     `db:"id" json:"id"`
	UUID            string     `db:"uuid" json:"uuid"`
	Name            string     `db:"name" json:"name"`
	Slug            *string    `db:"slug" json:"slug"`
	Email           string     `db:"email" json:"email"`
	EmailVerifiedAt *time.Time `db:"email_verified_at" json:"email_verified_at,omitempty"`
	Password        string     `db:"password" json:"-"` // Hidden from JSON
	LastLoginAt     *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	ScoreFormatID   *uint64    `db:"score_format_id" json:"score_format_id,omitempty"`
	ScoreFormat     *string    `db:"score_format" json:"score_format"`
	Avatar          *string    `db:"avatar" json:"-"`
	AvatarUrl       *string    `db:"-" json:"avatar_url,omitempty"`
	Banner          *string    `db:"banner" json:"-"`
	BannerUrl       *string    `db:"-" json:"banner_url,omitempty"`
	RememberToken   *string    `db:"remember_token" json:"-"`
	XP              uint64     `db:"xp" json:"xp"`
	Level           uint32     `db:"level" json:"level"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`

	// RBAC Relations
	Roles  []Role  `db:"-" json:"roles,omitempty"`
	Badges []Badge `db:"-" json:"badges,omitempty"`

	// Stats
	FollowersCount int  `db:"followers_count" json:"followers_count"`
	FollowingCount int  `db:"following_count" json:"following_count"`
	RatingsCount   int  `db:"ratings_count" json:"ratings_count"`
	IsFollowing    bool `db:"is_following" json:"is_following"`

	// AniList Link
	AnilistID             *uint64    `db:"anilist_id" json:"anilist_id,omitempty"`
	AnilistUsername       *string    `db:"anilist_username" json:"anilist_username,omitempty"`
	AnilistAccessToken    *string    `db:"anilist_access_token" json:"-"`
	AnilistRefreshToken   *string    `db:"anilist_refresh_token" json:"-"`
	AnilistTokenExpiresAt *time.Time `db:"anilist_token_expires_at" json:"-"`

	// Google Link
	GoogleID             *string    `db:"google_id" json:"google_id,omitempty"`
	GoogleEmail          *string    `db:"google_email" json:"google_email,omitempty"`
	GoogleAccessToken    *string    `db:"google_access_token" json:"-"`
	GoogleRefreshToken   *string    `db:"google_refresh_token" json:"-"`
	GoogleTokenExpiresAt *time.Time `db:"google_token_expires_at" json:"-"`
	
	// Profile customization
	About        *string `db:"about" json:"about,omitempty"`
	ProfileColor *string `db:"profile_color" json:"profile_color,omitempty"`
}

type RankingUser struct {
	User
	RatingsCount  int `db:"ratings_count" json:"ratings_count"`
	CommentsCount int `db:"comments_count" json:"comments_count"`
}

type Follow struct {
	FollowerID uint64    `db:"follower_id" json:"follower_id"`
	FollowedID uint64    `db:"followed_id" json:"followed_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type Role struct {
	ID          uint64       `db:"id" json:"id"`
	Name        string       `db:"name" json:"name"`
	Slug        string       `db:"slug" json:"slug"`
	Weight      int          `db:"weight" json:"weight"`
	Description *string      `db:"description" json:"description"`
	Permissions []Permission `db:"-" json:"permissions,omitempty"`
	CreatedAt   *time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time    `db:"updated_at" json:"updated_at"`
}

type Permission struct {
	ID          uint64    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Slug        string    `db:"slug" json:"slug"`
	Description *string   `db:"description" json:"description"`
	CreatedAt   *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at"`
}

type Badge struct {
	ID               uint64     `db:"id" json:"id"`
	UUID             string     `db:"uuid" json:"uuid"`
	Name             string     `db:"name" json:"name"`
	Description      *string    `db:"description" json:"description"`
	Icon             *string    `db:"icon" json:"-"`
	IconUrl          *string    `db:"-" json:"icon_url,omitempty"`
	IsActive         bool       `db:"is_active" json:"is_active"`
	IsAutomatic      bool       `db:"is_automatic" json:"is_automatic"`
	RequirementType  *string    `db:"requirement_type" json:"requirement_type,omitempty"`
	RequirementValue *int       `db:"requirement_value" json:"requirement_value,omitempty"`
	CreatedAt        *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        *time.Time `db:"updated_at" json:"updated_at"`
}

// Repositories
type UserRepository interface {
	GetByID(ctx context.Context, id uint64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*User, error)
	GetByAnilistID(ctx context.Context, anilistID uint64) (*User, error)
	GetBySlug(ctx context.Context, slug string) (*User, error)
	GetByUUID(ctx context.Context, uuid string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint64) error

	SetImage(ctx context.Context, userID uint64, imageType, imagePath string) error
	UpdateScoreFormat(ctx context.Context, userID uint64, format string) error
	UpdatePassword(ctx context.Context, userID uint64, hashedPassword string) error

	// Roles Loaders
	GetRolesByUserID(ctx context.Context, userID uint64) ([]Role, error)
	GetRoles(ctx context.Context) ([]Role, error)
	UpdateRoles(ctx context.Context, userID uint64, roleIDs []uint64) error

	// Badges Loaders
	GetBadgesByUserID(ctx context.Context, userID uint64) ([]Badge, error)
	UpdateBadges(ctx context.Context, userID uint64, badgeIDs []uint64) error

	// Permissions
	GetPermissionsByRoleID(ctx context.Context, roleID uint64) ([]Permission, error)
	GetPermissionsByUserID(ctx context.Context, userID uint64) ([]Permission, error)
	GetAllPermissions(ctx context.Context) ([]Permission, error)
	UpdateRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error

	// Admin
	GetUsers(ctx context.Context, page, limit int, search string) ([]User, int, error)
	GetRanking(ctx context.Context, sortBy string, limit, offset int) ([]RankingUser, int, error)

	// Follows
	Follow(ctx context.Context, followerID, followedID uint64) error
	Unfollow(ctx context.Context, followerID, followedID uint64) error
	IsFollowing(ctx context.Context, followerID, followedID uint64) (bool, error)
	GetFollowersCount(ctx context.Context, userID uint64) (int, error)
	GetFollowingCount(ctx context.Context, userID uint64) (int, error)
	GetFollowers(ctx context.Context, userID uint64, limit, offset int) ([]User, error)
	GetFollowing(ctx context.Context, userID uint64, limit, offset int) ([]User, error)
	GetMany(ctx context.Context, ids []uint64) ([]User, error)
}

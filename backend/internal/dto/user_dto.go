package dto

import "time"

type UserMinimalDTO struct {
	ID        string  `json:"id"`
	UUID      string  `json:"uuid"`
	Name      string  `json:"name"`
	Slug      *string `json:"slug"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
	XP        uint64  `json:"xp"`
	Level     uint32  `json:"level"`
	BannerUrl *string `json:"banner_url,omitempty"`
}

type BadgeDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IconUrl     *string `json:"icon_url,omitempty"`
}

type UserDTO struct {
	UserMinimalDTO
	About        *string  `json:"about,omitempty"`
	ProfileColor   *string    `json:"profile_color,omitempty"`
	FollowersCount int        `json:"followers_count"`
	FollowingCount int        `json:"following_count"`
	RatingsCount   int        `json:"ratings_count"`
	IsFollowing    bool       `json:"is_following"`
	Roles          []string   `json:"roles,omitempty"`
	Badges         []BadgeDTO `json:"badges,omitempty"`
	AnilistID      *uint64    `json:"anilist_id,omitempty"`
	AnilistUsername *string    `json:"anilist_username,omitempty"`
	GoogleID       *string    `json:"google_id,omitempty"`
	GoogleEmail    *string    `json:"google_email,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type AuthUserDTO struct {
	UserDTO
	Email         string  `json:"email"`
	ScoreFormatID *uint64 `json:"score_format_id,omitempty"`
	ScoreFormat   *string `json:"score_format,omitempty"`
}

type AuthResponseDTO struct {
	Token string      `json:"token"`
	User  AuthUserDTO `json:"user"`
}

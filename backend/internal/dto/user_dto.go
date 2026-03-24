package dto

import "time"

type UserMinimalDTO struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	Slug      *string `json:"slug"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
	XP        uint64  `json:"xp"`
	Level     uint32  `json:"level"`
}

type BadgeDTO struct {
	ID          uint64  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IconUrl     *string `json:"image_url,omitempty"`
}

type UserDTO struct {
	UserMinimalDTO
	BannerUrl      *string    `json:"banner_url,omitempty"`
	About          *string    `json:"about,omitempty"`
	ProfileColor   *string    `json:"profile_color,omitempty"`
	FollowersCount int        `json:"followers_count"`
	FollowingCount int        `json:"following_count"`
	RatingsCount   int        `json:"ratings_count"`
	IsFollowing    bool       `json:"is_following"`
	Badges         []BadgeDTO `json:"badges,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type AuthUserDTO struct {
	UserDTO
	Email         string   `json:"email"`
	ScoreFormatID *uint64  `json:"score_format_id,omitempty"`
	ScoreFormat   *string  `json:"score_format,omitempty"`
	Roles         []string `json:"roles,omitempty"`
}

type AuthResponseDTO struct {
	Token string      `json:"token"`
	User  AuthUserDTO `json:"user"`
}

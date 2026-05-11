package dto

import (
	"anirank/api/internal/domain"
	"time"
)

type UserMinimalDTO struct {
	ID            string               `json:"id"`
	UUID          string               `json:"uuid"`
	Name          string               `json:"name"`
	Slug          *string              `json:"slug"`
	AvatarUrl     *string              `json:"avatar_url,omitempty"`
	AvatarSources []domain.ImageSource `json:"avatar_sources,omitempty"`
	XP            uint64               `json:"xp"`
	Level         uint32               `json:"level"`
	RatingsCount  int                  `json:"ratings_count"`
	CommentsCount int                  `json:"comments_count"`
	BannerUrl     *string              `json:"banner_url,omitempty"`
	BannerSources []domain.ImageSource `json:"banner_sources,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
}

type BadgeDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IconUrl     *string `json:"icon_url,omitempty"`
}

type UserSocialIdentityDTO struct {
	Provider         string `json:"provider"`
	ProviderUsername string `json:"provider_username"`
}

type UserDTO struct {
	UserMinimalDTO
	About            *string                 `json:"about,omitempty"`
	ProfileColor     *string                 `json:"profile_color,omitempty"`
	FollowersCount   int                     `json:"followers_count"`
	FollowingCount   int                     `json:"following_count"`
	IsFollowing      bool                    `json:"is_following"`
	TruthScore       int                     `json:"truth_score"`
	IsShadowbanned   bool                    `json:"is_shadowbanned"`
	Roles            []string                `json:"roles,omitempty"`
	Badges           []BadgeDTO              `json:"badges,omitempty"`
	SocialIdentities []UserSocialIdentityDTO `json:"social_identities,omitempty"`
}

type AuthUserDTO struct {
	UserDTO
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	ScoreFormat     *string    `json:"score_format,omitempty"`
}

type AuthResponseDTO struct {
	Token string      `json:"token"`
	User  AuthUserDTO `json:"user"`
}

package domain

import (
	"time"
)

const (
	ProviderAnilist = "anilist"
	ProviderGoogle  = "google"
	ProviderDiscord = "discord"
)

type UserSocialIdentity struct {
	ID               string     `db:"id" json:"id"`
	UserID           uint64     `db:"user_id" json:"user_id"`
	Provider         string     `db:"provider" json:"provider"`
	ProviderID       string     `db:"provider_id" json:"provider_id"`
	ProviderUsername *string    `db:"provider_username" json:"provider_username"`
	AccessToken      *string    `db:"access_token" json:"-"`
	RefreshToken     *string    `db:"refresh_token" json:"-"`
	ExpiresAt        *time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
}

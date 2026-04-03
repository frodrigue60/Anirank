package domain

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

type BadgeRepository interface {
	GetByID(ctx context.Context, id uint64) (*Badge, error)
	GetAll(ctx context.Context) ([]Badge, error)
	GetAutomatic(ctx context.Context) ([]Badge, error)
	GetUserBadgeIDs(ctx context.Context, userID uint64) ([]uint64, error)
	Award(ctx context.Context, userID, badgeID uint64) error
	Create(ctx context.Context, badge *Badge) error
	Update(ctx context.Context, badge *Badge) error
	Delete(ctx context.Context, id uint64) error
	UpdateIcon(ctx context.Context, id uint64, iconPath string) error
}

type BadgeUsecase interface {
	GetByID(ctx context.Context, id uint64) (*Badge, error)
	GetAll(ctx context.Context) ([]Badge, error)
	Create(ctx context.Context, badge *Badge, meta AuditMetadata) error
	Update(ctx context.Context, badge *Badge, meta AuditMetadata) error
	Delete(ctx context.Context, id uint64, meta AuditMetadata) error
	HandleBadgeIcon(c *fiber.Ctx, badge *Badge) error
	ResolveBadgeURL(badge *Badge)
	ResolveBadgesURLs(badges []Badge)

	// Automation
	CheckAndAwardBadges(ctx context.Context, userID uint64, triggerType string) error
}

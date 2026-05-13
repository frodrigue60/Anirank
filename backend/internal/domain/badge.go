package domain

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

type BadgeRepository interface {
	GetByID(ctx context.Context, id uint64) (*Badge, error)
	GetMany(ctx context.Context, ids []uint64) ([]Badge, error)
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

	// Batch Processing
	ProcessBadgeIcons(ctx context.Context, progress chan<- string) error
}

// FilterHighestBadges groups badges by requirement_type and keeps only the one with the highest requirement_value.
// Badges without a requirement_type are always included.
func FilterHighestBadges(badges []Badge) []Badge {
	highest := make(map[string]Badge)
	var specials []Badge

	for _, b := range badges {
		if b.RequirementType == nil || *b.RequirementType == "" {
			specials = append(specials, b)
			continue
		}

		typeName := *b.RequirementType
		currentBest, exists := highest[typeName]

		val := 0
		if b.RequirementValue != nil {
			val = *b.RequirementValue
		}

		bestVal := -1
		if exists && currentBest.RequirementValue != nil {
			bestVal = *currentBest.RequirementValue
		}

		if !exists || val > bestVal {
			highest[typeName] = b
		}
	}

	result := make([]Badge, 0, len(highest)+len(specials))
	result = append(result, specials...)
	for _, b := range highest {
		result = append(result, b)
	}
	return result
}

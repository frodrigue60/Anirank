package usecase

import (
	"context"
	"encoding/json"
	"time"

	"anirank/api/internal/domain"
)

type xpUsecase struct {
	xpRepo       domain.XPRepository
	userRepo     domain.UserRepository
	badgeUsecase domain.BadgeUsecase
}

func NewXPUsecase(xpRepo domain.XPRepository, userRepo domain.UserRepository, badgeUsecase domain.BadgeUsecase) domain.XPUsecase {
	return &xpUsecase{
		xpRepo:       xpRepo,
		userRepo:     userRepo,
		badgeUsecase: badgeUsecase,
	}
}

func (u *xpUsecase) AwardXP(ctx context.Context, userID uint64, activityKey string, metadata map[string]interface{}) error {
	// 1. Activity Identification
	activity, err := u.xpRepo.GetActivityByKey(ctx, activityKey)
	if err != nil {
		return err
	}

	// 2. Uniqueness/Target-Specific Validation
	// These activities only award XP ONCE per specific target (e.g., once per song)
	targetActivities := map[string]string{
		"rate_song":       "song_id",
		"add_favorite":    "song_id",
		"add_to_playlist": "song_id",
		"create_playlist": "playlist_id",
		"comment":         "comment_id",
		"reply":           "comment_id",
	}

	if targetIDKey, ok := targetActivities[activityKey]; ok {
		if targetID, exists := metadata[targetIDKey]; exists {
			existingLog, err := u.xpRepo.GetLogByActivityAndMetadata(ctx, userID, activity.ID, targetIDKey, targetID)
			if err == nil && existingLog != nil {
				return nil // Already awarded XP for this specific target
			}
		}
	}

	// 3. Global Cooldown Validation
	// Prevents spamming the same type of activity too frequently
	if activity.CooldownSeconds > 0 {
		lastLog, err := u.xpRepo.GetLastLogByActivity(ctx, userID, activity.ID)
		if err == nil && lastLog != nil {
			// Special handling for daily_login (calendar day based)
			if activityKey == "daily_login" {
				now := time.Now().UTC()
				if lastLog.CreatedAt.Year() == now.Year() && lastLog.CreatedAt.YearDay() == now.YearDay() {
					return nil // Already logged in today
				}
			} else {
				// Standard duration-based cooldown
				if time.Since(lastLog.CreatedAt).Seconds() < float64(activity.CooldownSeconds) {
					return nil // In cooldown
				}
			}
		}
	}

	// 4. Transactional Awarding
	metadataJSON, _ := json.Marshal(metadata)
	log := &domain.XPLog{
		UserID:       userID,
		XPActivityID: activity.ID,
		XPAmount:     activity.XPAmount,
		Metadata:     metadataJSON,
	}

	if err := u.xpRepo.CreateLog(ctx, log); err != nil {
		return err
	}

	// Fetch user to get current XP
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	newXP := user.XP + uint64(activity.XPAmount)
	newLevel, err := u.xpRepo.GetCurrentLevel(ctx, newXP)
	if err != nil {
		return err
	}

	if err := u.xpRepo.UpdateUserXPAndLevel(ctx, userID, activity.XPAmount, newLevel); err != nil {
		return err
	}

	// 5. Automatic Badge Check
	_ = u.badgeUsecase.CheckAndAwardBadges(ctx, userID, "level")

	return nil
}

func (u *xpUsecase) CheckDailyLogin(ctx context.Context, userID uint64) error {
	return u.AwardXP(ctx, userID, "daily_login", map[string]interface{}{})
}

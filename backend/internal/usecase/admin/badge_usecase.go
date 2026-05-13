package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type BadgeUsecase struct {
	repo            domain.BadgeRepository
	userRepo        domain.UserRepository
	interactionRepo domain.InteractionRepository
	commentRepo     domain.CommentRepository
	storage         infrastructure.StorageService
	mediaService    infrastructure.MediaService
	auditUsecase    domain.AuditLogUsecase
	activityUsecase domain.ActivityUsecase
	notificationUsecase domain.NotificationUsecase
	evaluators      map[string]BadgeEvaluator // Strategy pattern: triggerType → evaluator
}

func NewBadgeUsecase(
	repo domain.BadgeRepository,
	userRepo domain.UserRepository,
	interactionRepo domain.InteractionRepository,
	commentRepo domain.CommentRepository,
	storage infrastructure.StorageService,
	mediaService infrastructure.MediaService,
	auditUsecase domain.AuditLogUsecase,
	activityUsecase domain.ActivityUsecase,
	notificationUsecase domain.NotificationUsecase,
) domain.BadgeUsecase {
	return &BadgeUsecase{
		repo:                repo,
		userRepo:            userRepo,
		interactionRepo:     interactionRepo,
		commentRepo:         commentRepo,
		storage:             storage,
		mediaService:        mediaService,
		auditUsecase:        auditUsecase,
		activityUsecase:     activityUsecase,
		notificationUsecase: notificationUsecase,
		evaluators:          buildEvaluators(userRepo, interactionRepo, commentRepo),
	}
}

func (u *BadgeUsecase) GetByID(ctx context.Context, id uint64) (*domain.Badge, error) {
	badge, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewAppError(404, "Badge not found", err)
	}
	u.ResolveBadgeURL(badge)
	return badge, nil
}

func (u *BadgeUsecase) GetAll(ctx context.Context) ([]domain.Badge, error) {
	badges, err := u.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	u.ResolveBadgesURLs(badges)
	return badges, nil
}

func (u *BadgeUsecase) Create(ctx context.Context, badge *domain.Badge, meta domain.AuditMetadata) error {
	if badge.Name == "" {
		return domain.NewAppError(400, "Badge name is required", nil)
	}
	if badge.UUID == "" {
		badge.UUID = uuid.New().String()
	}
	if err := u.repo.Create(ctx, badge); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "created", badge.ID, "badge", nil, badge, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *BadgeUsecase) Update(ctx context.Context, badge *domain.Badge, meta domain.AuditMetadata) error {
	if badge.Name == "" {
		return domain.NewAppError(400, "Badge name is required", nil)
	}

	existing, err := u.repo.GetByID(ctx, badge.ID)
	if err != nil {
		return domain.NewAppError(404, "Badge not found", err)
	}

	badge.Icon = existing.Icon // Preserve existing icon path unless specifically uploaded

	if err := u.repo.Update(ctx, badge); err != nil {
		return err
	}
	_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "updated", badge.ID, "badge", existing, badge, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *BadgeUsecase) Delete(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.repo.GetByID(ctx, id)
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	if existing != nil {
		_ = u.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "badge", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	}
	return nil
}

// HandleBadgeIcon manages the multipart file upload parsing, processing, and passing to storage
func (u *BadgeUsecase) HandleBadgeIcon(c *fiber.Ctx, badge *domain.Badge) error {
	file, err := c.FormFile("icon")
	if err == nil && file != nil {
		f, err := file.Open()
		if err != nil {
			return err
		}
		defer f.Close()

		// Use MediaService with Square preset (256x256 base, with sm 128x128 variant)
		path, _, err := u.mediaService.UploadWithResolutions(c.Context(), "badges", badge.ID, f, infrastructure.PresetSquare)
		if err != nil {
			return fmt.Errorf("failed to upload badge icon: %w", err)
		}

		// Update icon in DB
		if err := u.repo.UpdateIcon(c.Context(), badge.ID, path); err != nil {
			return fmt.Errorf("failed to update badge icon in DB: %w", err)
		}
		badge.Icon = &path
	}
	return nil
}

func (u *BadgeUsecase) ResolveBadgeURL(badge *domain.Badge) {
	if badge.Icon != nil && *badge.Icon != "" {
		url := u.mediaService.GetURL(*badge.Icon)
		badge.IconUrl = &url
		badge.IconSources = u.mediaService.GetImageSources(*badge.Icon)
	}
}

func (u *BadgeUsecase) ResolveBadgesURLs(badges []domain.Badge) {
	for i := range badges {
		u.ResolveBadgeURL(&badges[i])
	}
}

func (u *BadgeUsecase) sendBadgeNotification(ctx context.Context, userID uint64, badge *domain.Badge) {
	u.ResolveBadgeURL(badge)

	data := map[string]interface{}{
		"badge_name":         badge.Name,
		"badge_icon":         badge.IconUrl,
		"badge_icon_sources": badge.IconSources,
	}
	dataJSON, _ := json.Marshal(data)

	notif := &domain.Notification{
		UserID:      userID,
		Type:        domain.NotifBadgeAwarded,
		SubjectID:   &badge.ID,
		SubjectUUID: &badge.UUID,
		SubjectType: utils.Ptr("badge"),
		Data:        dataJSON,
	}

	_ = u.notificationUsecase.Create(ctx, notif)
}

func (u *BadgeUsecase) CheckAndAwardBadges(ctx context.Context, userID uint64, triggerType string) error {
	// 1. Find the evaluator for this trigger type — if none registered, skip silently
	evaluator, ok := u.evaluators[triggerType]
	if !ok {
		return nil
	}

	// 2. Get all automatic badges that match this trigger type
	autoBadges, err := u.repo.GetAutomatic(ctx)
	if err != nil {
		return err
	}

	if len(autoBadges) == 0 {
		return nil
	}

	// 3. Get user's already-earned badges to avoid duplicates
	userBadgeIDs, err := u.repo.GetUserBadgeIDs(ctx, userID)
	if err != nil {
		return err
	}
	earnedMap := make(map[uint64]bool, len(userBadgeIDs))
	for _, id := range userBadgeIDs {
		earnedMap[id] = true
	}

	// 4. Evaluate each eligible badge using the registered strategy
	for _, badge := range autoBadges {
		if earnedMap[badge.ID] {
			continue // Already earned
		}
		if badge.RequirementType == nil || *badge.RequirementType != triggerType {
			continue // Doesn't match this trigger
		}

		requiredValue := 0
		if badge.RequirementValue != nil {
			requiredValue = *badge.RequirementValue
		}

		shouldAward, err := evaluator.CanAward(ctx, userID, requiredValue)
		if err != nil {
			// Log and continue — one failed evaluation shouldn't block others
			fmt.Printf("[BadgeEvaluator] Error evaluating badge %d for user %d: %v\n", badge.ID, userID, err)
			continue
		}

		if shouldAward {
			if err := u.repo.Award(ctx, userID, badge.ID); err != nil {
				fmt.Printf("[BadgeUsecase] Error awarding badge %d to user %d: %v\n", badge.ID, userID, err)
			} else {
				// Log badge award activity
				_ = u.activityUsecase.LogActivity(ctx, userID, "badge_award", badge.ID, "badge", nil)

				// Send notification to user
				u.sendBadgeNotification(ctx, userID, &badge)
			}
		}
	}

	return nil
}

func (u *BadgeUsecase) ProcessBadgeIcons(ctx context.Context, progress chan<- string) error {
	sendProgress := func(msg string) {
		if progress != nil {
			select {
			case progress <- msg:
			default:
			}
		}
	}

	badges, err := u.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	sendProgress(fmt.Sprintf("Processing %d badges for icons...", len(badges)))

	for i, badge := range badges {
		if badge.Icon == nil || *badge.Icon == "" {
			continue
		}

		sendProgress(fmt.Sprintf("[%d/%d] Processing icon for: %s", i+1, len(badges), badge.Name))

		file, err := u.mediaService.GetFile(ctx, *badge.Icon)
		if err != nil {
			sendProgress(fmt.Sprintf("✗ Failed to get file for %s: %v", badge.Name, err))
			continue
		}

		newPath, _, err := u.mediaService.UploadWithResolutions(ctx, "badges", badge.ID, file, infrastructure.PresetSquare)
		file.Close()
		if err != nil {
			sendProgress(fmt.Sprintf("✗ Failed to process %s: %v", badge.Name, err))
			continue
		}

		oldPath := *badge.Icon
		if err := u.repo.UpdateIcon(ctx, badge.ID, newPath); err != nil {
			sendProgress(fmt.Sprintf("✗ Failed to update DB for %s: %v", badge.Name, err))
			continue
		}

		if oldPath != newPath && !strings.HasPrefix(oldPath, "http") {
			_ = u.storage.DeleteFile(ctx, oldPath)
		}

		sendProgress(fmt.Sprintf("✓ Success: %s", badge.Name))
	}

	sendProgress("Completed badge icon processing!")
	return nil
}

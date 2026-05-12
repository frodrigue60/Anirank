package admin

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/pkg/imageutil"

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
) domain.BadgeUsecase {
	return &BadgeUsecase{
		repo:            repo,
		userRepo:        userRepo,
		interactionRepo: interactionRepo,
		commentRepo:     commentRepo,
		storage:         storage,
		mediaService:    mediaService,
		auditUsecase:    auditUsecase,
		evaluators:      buildEvaluators(userRepo, interactionRepo, commentRepo),
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

		// 1. Decode image
		img, _, err := imageutil.Decode(f)
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}

		// 2. Resize to standard badge size (128x128)
		processedImg := imageutil.Fill(img, 128, 128)

		// 3. Encode to AVIF
		avifData, err := imageutil.EncodeAVIF(processedImg, 80)
		if err != nil {
			return fmt.Errorf("failed to encode to avif: %w", err)
		}

		// 4. Prepare path and storage
		path := fmt.Sprintf("badges/%s-%d.avif", time.Now().Format("20060102150405"), badge.ID)
		
		// Delete old if exists
		if badge.Icon != nil && *badge.Icon != "" {
			_ = u.storage.DeleteFile(c.Context(), *badge.Icon)
		}

		// Upload processed data
		reader := bytes.NewReader(avifData)
		if _, err := u.storage.UploadFile(c.Context(), path, reader, int64(len(avifData)), "image/avif"); err == nil {
			// Update icon in DB
			_ = u.repo.UpdateIcon(c.Context(), badge.ID, path)
			badge.Icon = &path
		}
	}
	return nil
}

func (u *BadgeUsecase) ResolveBadgeURL(badge *domain.Badge) {
	if badge.Icon != nil && *badge.Icon != "" {
		url := u.storage.GetURL(*badge.Icon)
		badge.IconUrl = &url
	}
}

func (u *BadgeUsecase) ResolveBadgesURLs(badges []domain.Badge) {
	for i := range badges {
		u.ResolveBadgeURL(&badges[i])
	}
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

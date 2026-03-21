package admin

import (
	"context"
	"fmt"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"

	"github.com/gofiber/fiber/v2"
)

type BadgeUsecase struct {
	repo         domain.BadgeRepository
	storage      infrastructure.StorageService
	auditUsecase domain.AuditLogUsecase
}

func NewBadgeUsecase(repo domain.BadgeRepository, storage infrastructure.StorageService, auditUsecase domain.AuditLogUsecase) domain.BadgeUsecase {
	return &BadgeUsecase{repo: repo, storage: storage, auditUsecase: auditUsecase}
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

// HandleBadgeIcon manages the multipart file upload parsing and passing to storage
func (u *BadgeUsecase) HandleBadgeIcon(c *fiber.Ctx, badge *domain.Badge) error {
	file, err := c.FormFile("icon")
	if err == nil && file != nil {
		// Found an uploaded image
		ext := "png" // Default, you could parse content-type
		if len(file.Filename) > 4 && file.Filename[len(file.Filename)-4:] == ".jpg" {
			ext = "jpg"
		} else if len(file.Filename) > 4 && file.Filename[len(file.Filename)-4:] == "jpeg" {
			ext = "jpeg"
		} else if len(file.Filename) > 4 && file.Filename[len(file.Filename)-4:] == ".gif" {
			ext = "gif"
		} else if len(file.Filename) > 4 && file.Filename[len(file.Filename)-4:] == "webp" {
			ext = "webp"
		}

		f, err := file.Open()
		if err == nil {
			defer f.Close()

			path := fmt.Sprintf("badges/%s-%d.%s", time.Now().Format("20060102150405"), badge.ID, ext)
			contentType := "image/" + ext
			if ext == "jpg" {
				contentType = "image/jpeg"
			}

			// Delete old if exists (optonal, depends on requirements)
			if badge.Icon != nil && *badge.Icon != "" {
				_ = u.storage.DeleteFile(c.Context(), *badge.Icon)
			}

			if _, err := u.storage.UploadFile(c.Context(), path, f, file.Size, contentType); err == nil {
				// Update icon in DB
				_ = u.repo.UpdateIcon(c.Context(), badge.ID, path)
				badge.Icon = &path
			}
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

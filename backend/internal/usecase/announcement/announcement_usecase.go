package announcement

import (
	"context"
	"mime/multipart"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/security"
	"github.com/google/uuid"
)

type AnnouncementUsecase struct {
	repo    domain.AnnouncementRepository
	media   infrastructure.MediaService
}

func NewAnnouncementUsecase(repo domain.AnnouncementRepository, media infrastructure.MediaService) *AnnouncementUsecase {
	return &AnnouncementUsecase{repo: repo, media: media}
}

func (u *AnnouncementUsecase) GetByID(ctx context.Context, id uint64) (*domain.Announcement, error) {
	a, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.enrichAnnouncement(a)
	return a, nil
}

func (u *AnnouncementUsecase) GetPublicAnnouncements(ctx context.Context) ([]domain.Announcement, error) {
	announcements, err := u.repo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	for i := range announcements {
		u.enrichAnnouncement(&announcements[i])
	}
	return announcements, nil
}

func (u *AnnouncementUsecase) GetAllAnnouncements(ctx context.Context, filters domain.AnnouncementFilters, limit, offset int) ([]domain.Announcement, error) {
	announcements, err := u.repo.GetAll(ctx, filters, limit, offset)
	if err != nil {
		return nil, err
	}
	for i := range announcements {
		u.enrichAnnouncement(&announcements[i])
	}
	return announcements, nil
}

func (u *AnnouncementUsecase) GetCount(ctx context.Context, filters domain.AnnouncementFilters) (int, error) {
	return u.repo.Count(ctx, filters)
}

func (u *AnnouncementUsecase) Create(ctx context.Context, a *domain.Announcement) error {
	a.Content = security.SanitizeHTMLPtr(a.Content)
	if a.UUID == "" {
		a.UUID = uuid.New().String()
	}
	return u.repo.Create(ctx, a)
}

func (u *AnnouncementUsecase) Update(ctx context.Context, a *domain.Announcement) error {
	a.Content = security.SanitizeHTMLPtr(a.Content)
	return u.repo.Update(ctx, a)
}

func (u *AnnouncementUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}

func (u *AnnouncementUsecase) ToggleActive(ctx context.Context, id uint64) error {
	return u.repo.ToggleActive(ctx, id)
}

func (u *AnnouncementUsecase) UploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Upload using MediaService with Announcement preset (multi-resolution)
	path, _, err := u.media.UploadWithResolutions(ctx, "announcements", 0, src, infrastructure.PresetAnnouncement)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (u *AnnouncementUsecase) enrichAnnouncement(a *domain.Announcement) {
	if a.Image != nil {
		url := u.media.GetURL(*a.Image)
		a.ImageUrl = &url
		a.ImageSources = u.media.GetImageSources(*a.Image)
	}
}

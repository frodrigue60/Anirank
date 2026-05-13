package admin

import (
	"context"
	"io"
	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
)

type partnerUsecase struct {
	repo         domain.PartnerRepository
	mediaService infrastructure.MediaService
}

func NewPartnerUsecase(repo domain.PartnerRepository, mediaService infrastructure.MediaService) domain.PartnerUsecase {
	return &partnerUsecase{
		repo:         repo,
		mediaService: mediaService,
	}
}

func (u *partnerUsecase) GetActivePartners(ctx context.Context) ([]domain.Partner, error) {
	partners, err := u.repo.GetAll(ctx, true)
	if err != nil {
		return nil, err
	}

	if partners == nil {
		partners = []domain.Partner{}
	}

	for i := range partners {
		u.enrichPartner(&partners[i])
	}

	return partners, nil
}

func (u *partnerUsecase) AdminGetAllPartners(ctx context.Context) ([]domain.Partner, error) {
	partners, err := u.repo.GetAll(ctx, false)
	if err != nil {
		return nil, err
	}

	if partners == nil {
		partners = []domain.Partner{}
	}

	for i := range partners {
		u.enrichPartner(&partners[i])
	}

	return partners, nil
}

func (u *partnerUsecase) AdminGetPartnerByUUID(ctx context.Context, uuid string) (*domain.Partner, error) {
	p, err := u.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	u.enrichPartner(p)
	return p, nil
}

func (u *partnerUsecase) AdminCreatePartner(ctx context.Context, p *domain.Partner) error {
	if err := u.repo.Create(ctx, p); err != nil {
		return err
	}
	u.enrichPartner(p)
	return nil
}

func (u *partnerUsecase) AdminUpdatePartner(ctx context.Context, uuid string, p *domain.Partner) error {
	existing, err := u.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	existing.Name = p.Name
	existing.URL = p.URL
	
	// Only update banner if a new one was uploaded
	if p.Banner != nil && *p.Banner != "" {
		existing.Banner = p.Banner
	}
	
	existing.Description = p.Description
	existing.Type = p.Type
	existing.SortOrder = p.SortOrder
	existing.IsActive = p.IsActive

	if err := u.repo.Update(ctx, existing); err != nil {
		return err
	}
	
	u.enrichPartner(existing)
	// Update the input pointer so the handler returns enriched data
	*p = *existing
	
	return nil
}

func (u *partnerUsecase) AdminDeletePartner(ctx context.Context, uuid string) error {
	existing, err := u.repo.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	return u.repo.Delete(ctx, existing.ID)
}

func (u *partnerUsecase) UploadBanner(ctx context.Context, p *domain.Partner, file io.Reader) error {
	path, _, err := u.mediaService.UploadWithResolutions(ctx, "partners/banners", p.ID, file, infrastructure.PresetLandscape)
	if err != nil {
		return err
	}
	p.Banner = &path
	return nil
}

func (u *partnerUsecase) enrichPartner(p *domain.Partner) {
	if p.Banner != nil {
		url := u.mediaService.GetURL(*p.Banner)
		p.BannerURL = &url
		p.BannerSources = u.mediaService.GetImageSources(*p.Banner)
	}
}


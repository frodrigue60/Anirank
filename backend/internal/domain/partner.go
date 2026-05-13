package domain

import (
	"context"
	"io"
	"time"
)

type Partner struct {
	ID          uint64    `db:"id" json:"-"`
	UUID        string    `db:"uuid" json:"uuid" form:"uuid"`
	Name        string    `db:"name" json:"name" form:"name"`
	URL         string    `db:"url" json:"url" form:"url"`
	Banner      *string   `db:"banner" json:"-"`
	BannerURL   *string   `db:"-" json:"banner_url,omitempty"`
	BannerSources []ImageSource `db:"-" json:"banner_sources,omitempty"`
	Description *string   `db:"description" json:"description" form:"description"`
	Type        string    `db:"type" json:"type" form:"type"`
	SortOrder   int       `db:"sort_order" json:"sort_order,string" form:"sort_order"`
	IsActive    bool      `db:"is_active" json:"is_active,string" form:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type PartnerRepository interface {
	GetByID(ctx context.Context, id uint64) (*Partner, error)
	GetByUUID(ctx context.Context, uuid string) (*Partner, error)
	GetAll(ctx context.Context, onlyActive bool) ([]Partner, error)
	Create(ctx context.Context, partner *Partner) error
	Update(ctx context.Context, partner *Partner) error
	Delete(ctx context.Context, id uint64) error
}

type PartnerUsecase interface {
	GetActivePartners(ctx context.Context) ([]Partner, error)
	AdminGetAllPartners(ctx context.Context) ([]Partner, error)
	AdminGetPartnerByUUID(ctx context.Context, uuid string) (*Partner, error)
	AdminCreatePartner(ctx context.Context, partner *Partner) error
	AdminUpdatePartner(ctx context.Context, uuid string, partner *Partner) error
	AdminDeletePartner(ctx context.Context, uuid string) error
	UploadBanner(ctx context.Context, partner *Partner, file io.Reader) error
}
func Ptr[T any](v T) *T {
	return &v
}

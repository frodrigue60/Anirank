package public

import (
	"context"

	"anirank/api/internal/domain"

	"golang.org/x/sync/errgroup"
)

type InitData struct {
	Years     []domain.Year     `json:"years"`
	Seasons   []domain.Season   `json:"seasons"`
	Formats   []domain.Format   `json:"formats"`
	Genres    []domain.Genre    `json:"genres"`
	SongTypes []domain.SongType `json:"song_types"`
}

type DiscoveryUsecase struct {
	taxonomyRepo domain.TaxonomyRepository
	songRepo     domain.SongRepository
}

func NewDiscoveryUsecase(tr domain.TaxonomyRepository, sr domain.SongRepository) *DiscoveryUsecase {
	return &DiscoveryUsecase{
		taxonomyRepo: tr,
		songRepo:     sr,
	}
}

// GetInitData fetches basic metadata needed by the SPA asynchronously
func (u *DiscoveryUsecase) GetInitData(ctx context.Context) (*InitData, error) {
	var data InitData
	g, gCtx := errgroup.WithContext(ctx)

	// Fetch concurrently
	g.Go(func() error {
		years, err := u.taxonomyRepo.GetAllYears(gCtx)
		data.Years = years
		return err
	})

	g.Go(func() error {
		seasons, err := u.taxonomyRepo.GetAllSeasons(gCtx)
		data.Seasons = seasons
		return err
	})

	g.Go(func() error {
		formats, err := u.taxonomyRepo.GetAllFormats(gCtx)
		data.Formats = formats
		return err
	})

	g.Go(func() error {
		genres, err := u.taxonomyRepo.GetAllGenres(gCtx)
		data.Genres = genres
		return err
	})

	g.Go(func() error {
		types, err := u.songRepo.GetSongTypes(gCtx)
		data.SongTypes = types
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, domain.NewAppError(500, "failed configuring initial data", err)
	}

	return &data, nil
}

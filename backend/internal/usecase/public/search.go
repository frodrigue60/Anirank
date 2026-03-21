package public

import (
	"context"
	"strings"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"

	"golang.org/x/sync/errgroup"
)

type SearchResult struct {
	Animes  []domain.Anime  `json:"animes"`
	Songs   []domain.Song   `json:"songs"`
	Artists []domain.Artist `json:"artists"`
	Studios []domain.Studio `json:"studios"`
	Users   []domain.User   `json:"users"`
}

type SearchUsecase struct {
	animeRepo    domain.AnimeRepository
	songRepo     domain.SongRepository
	artistRepo   domain.ArtistRepository
	taxonomyRepo domain.TaxonomyRepository
	userRepo     domain.UserRepository
	storage      infrastructure.StorageService
}

// Interface extensions needed for searching
type SearchableAnimeRepo interface {
	Search(ctx context.Context, term string, limit int) ([]domain.Anime, error)
}
type SearchableSongRepo interface {
	Search(ctx context.Context, term string, limit int) ([]domain.Song, error)
}
type SearchableArtistRepo interface {
	Search(ctx context.Context, term string, limit int) ([]domain.Artist, error)
}
type SearchableTaxonomyRepo interface {
	SearchStudios(ctx context.Context, term string, limit int) ([]domain.Studio, error)
}
type SearchableUserRepo interface {
	Search(ctx context.Context, term string, limit int) ([]domain.User, error)
}

func NewSearchUsecase(ar domain.AnimeRepository, sr domain.SongRepository, artistR domain.ArtistRepository, tr domain.TaxonomyRepository, ur domain.UserRepository, storage infrastructure.StorageService) *SearchUsecase {
	return &SearchUsecase{
		animeRepo:    ar,
		songRepo:     sr,
		artistRepo:   artistR,
		taxonomyRepo: tr,
		userRepo:     ur,
		storage:      storage,
	}
}

func (u *SearchUsecase) GlobalSearch(ctx context.Context, term string) (*SearchResult, error) {
	if len(strings.TrimSpace(term)) < 3 {
		return nil, domain.NewAppError(400, "Search term must be at least 3 characters", nil)
	}

	termWrapped := "%" + term + "%" // Prepare for LIKE query

	var result SearchResult
	g, gCtx := errgroup.WithContext(ctx)
	limit := 10 // Max results per category

	g.Go(func() error {
		if repo, ok := u.animeRepo.(SearchableAnimeRepo); ok {
			res, err := repo.Search(gCtx, termWrapped, limit)
			result.Animes = res
			return err
		}
		return nil
	})

	g.Go(func() error {
		if repo, ok := u.songRepo.(SearchableSongRepo); ok {
			res, err := repo.Search(gCtx, termWrapped, limit)
			result.Songs = res
			return err
		}
		return nil
	})

	g.Go(func() error {
		if repo, ok := u.artistRepo.(SearchableArtistRepo); ok {
			res, err := repo.Search(gCtx, termWrapped, limit)
			result.Artists = res
			return err
		}
		return nil
	})

	g.Go(func() error {
		if repo, ok := u.taxonomyRepo.(SearchableTaxonomyRepo); ok {
			res, err := repo.SearchStudios(gCtx, termWrapped, limit)
			result.Studios = res
			return err
		}
		return nil
	})

	g.Go(func() error {
		if repo, ok := u.userRepo.(SearchableUserRepo); ok {
			res, err := repo.Search(gCtx, termWrapped, limit)
			result.Users = res
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, domain.NewAppError(500, "Global search failed", err)
	}

	// Resolve URLs
	for i := range result.Animes {
		if result.Animes[i].Cover != nil {
			url := u.storage.GetURL(*result.Animes[i].Cover)
			result.Animes[i].CoverUrl = &url
		}
		if result.Animes[i].Banner != nil {
			url := u.storage.GetURL(*result.Animes[i].Banner)
			result.Animes[i].BannerUrl = &url
		}
	}

	for i := range result.Artists {
		if result.Artists[i].Avatar != nil {
			url := u.storage.GetURL(*result.Artists[i].Avatar)
			result.Artists[i].AvatarUrl = &url
		}
	}

	for i := range result.Users {
		if result.Users[i].Avatar != nil {
			url := u.storage.GetURL(*result.Users[i].Avatar)
			result.Users[i].AvatarUrl = &url
		}
	}

	for i := range result.Studios {
		if result.Studios[i].Logo != nil {
			url := u.storage.GetURL(*result.Studios[i].Logo)
			result.Studios[i].LogoUrl = &url
		}
	}

	return &result, nil
}

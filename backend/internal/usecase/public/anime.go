package public

import (
	"context"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
)

type AnimeUsecase struct {
	animeRepo    domain.AnimeRepository
	songRepo     domain.SongRepository
	mediaService infrastructure.MediaService
}

func NewAnimeUsecase(ar domain.AnimeRepository, sr domain.SongRepository, media infrastructure.MediaService) *AnimeUsecase {
	return &AnimeUsecase{
		animeRepo:    ar,
		songRepo:     sr,
		mediaService: media,
	}
}

func (u *AnimeUsecase) GetAnimeBySlug(ctx context.Context, slug string) (*domain.Anime, error) {
	anime, err := u.animeRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, domain.NewAppError(404, "Anime not found", err)
	}

	// Load core relations like Studios, Producers, Genres, Songs, Images
	if err := u.animeRepo.LoadRelations(ctx, anime, false); err != nil {
		return nil, domain.NewAppError(500, "Failed to load anime relations", err)
	}

	u.enrichAnime(ctx, anime)

	// Enrich songs with computed Name, Artists, and active variants
	if len(anime.Songs) > 0 {
		u.enrichAnimeSongs(ctx, anime.Songs)
	}

	return anime, nil
}

func (u *AnimeUsecase) enrichAnimeSongs(ctx context.Context, songs []domain.Song) {
	songIDs := make([]uint64, len(songs))
	for i := range songs {
		s := &songs[i]
		if s.SongRomaji != nil {
			s.Name = *s.SongRomaji
		} else if s.SongEN != nil {
			s.Name = *s.SongEN
		} else if s.SongJP != nil {
			s.Name = *s.SongJP
		}
		artists, _ := u.songRepo.GetArtistsBySongID(ctx, s.ID, false)
		s.Artists = artists
		songIDs[i] = s.ID
	}

	variantsMap, err := u.songRepo.GetVariantsBySongIDs(ctx, songIDs)
	if err != nil {
		return
	}

	for i := range songs {
		songs[i].Variants = activeVariantsForSong(variantsMap[songs[i].ID], u.mediaService)
	}
}

func (u *AnimeUsecase) GetPaginatedAnimes(ctx context.Context, limit, offset int, filters domain.AnimeFilters) ([]domain.Anime, int, error) {
	animes, err := u.animeRepo.GetPaginated(ctx, limit, offset, filters)
	if err != nil {
		return nil, 0, domain.NewAppError(500, "Failed to load animes list", err)
	}

	total, _ := u.animeRepo.Count(ctx, filters)

	// Lightweight pagination eager load (only essential info for lists)
	if err := u.animeRepo.LoadManyRelations(ctx, animes, false); err != nil {
		return nil, 0, domain.NewAppError(500, "Failed to load animes relations", err)
	}

	for i := range animes {
		u.enrichAnime(ctx, &animes[i])
	}

	return animes, total, nil
}

func (u *AnimeUsecase) enrichAnime(ctx context.Context, anime *domain.Anime) {
	anime.CoverUrl = u.mediaService.Resolve(anime.Cover)
	if anime.Cover != nil && *anime.Cover != "" {
		anime.CoverSources = u.mediaService.GetImageSources(*anime.Cover)
	}
	anime.BannerUrl = u.mediaService.Resolve(anime.Banner)
	if anime.Banner != nil && *anime.Banner != "" {
		anime.BannerSources = u.mediaService.GetImageSources(*anime.Banner)
	}
}

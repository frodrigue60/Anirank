package public

import (
	"context"
	"strings"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
)

type SearchResult struct {
	Animes  []domain.Anime  `json:"animes"`
	Songs   []domain.Song   `json:"songs"`
	Artists []domain.Artist `json:"artists"`
	Studios []domain.Studio `json:"studios"`
	Users   []domain.User   `json:"users"`
}

type SearchUsecase struct {
	searchRepo domain.SearchRepository
	storage    infrastructure.StorageService
}

func NewSearchUsecase(searchRepo domain.SearchRepository, storage infrastructure.StorageService) *SearchUsecase {
	return &SearchUsecase{
		searchRepo: searchRepo,
		storage:    storage,
	}
}

func (u *SearchUsecase) GlobalSearch(ctx context.Context, term string) (*SearchResult, error) {
	if len(strings.TrimSpace(term)) < 3 {
		return nil, domain.NewAppError(400, "Search term must be at least 3 characters", nil)
	}

	// 1. Performing a single unified query (limit 50 to have enough variety)
	items, err := u.searchRepo.GlobalSearch(ctx, term, 50)
	if err != nil {
		return nil, domain.NewAppError(500, "Global search failed", err)
	}

	var result SearchResult
	// Initialize slices to avoid nulls in JSON
	result.Animes = []domain.Anime{}
	result.Songs = []domain.Song{}
	result.Artists = []domain.Artist{}
	result.Studios = []domain.Studio{}
	result.Users = []domain.User{}

	// 2. Mapping flat list to categories
	for _, item := range items {
		// Resolve image URL if present
		var imgUrl *string
		if item.ImageURL != nil {
			url := u.storage.GetURL(*item.ImageURL)
			imgUrl = &url
		}

		switch item.ItemType {
		case "anime":
			result.Animes = append(result.Animes, domain.Anime{
				UUID:     item.ItemUUID,
				Title:    item.Title,
				Slug:     item.Slug,
				CoverUrl: imgUrl,
			})
		case "song":
			// We need to reconstruct enough of the Song struct for the frontend
			// Note: subtitle usually contains "Song Type - Anime Title"
			// slug usually contains "anime-slug/song-slug" if we want direct links
			
			// For now, we split the slug if it contains a slash
			songSlug := item.Slug
			animeSlug := ""
			if strings.Contains(item.Slug, "/") {
				parts := strings.Split(item.Slug, "/")
				animeSlug = parts[0]
				songSlug = parts[1]
			}

			songType := ""
			if item.Subtitle != nil {
				songType = strings.Replace(*item.Subtitle, "Song • ", "", 1)
			}

			// Copy title to local variable to take pointer
			songTitle := item.Title

			result.Songs = append(result.Songs, domain.Song{
				UUID:        item.ItemUUID,
				SongRomaji:  &songTitle,
				Slug:        songSlug,
				Type:        songType,
				Anime:       &domain.Anime{Slug: animeSlug},
			})
		case "artist":
			result.Artists = append(result.Artists, domain.Artist{
				UUID:      item.ItemUUID,
				Name:      item.Title,
				Slug:      item.Slug,
				AvatarUrl: imgUrl,
			})
		case "studio":
			result.Studios = append(result.Studios, domain.Studio{
				UUID:    item.ItemUUID,
				Name:    item.Title,
				Slug:    item.Slug,
				LogoUrl: imgUrl,
			})
		case "user":
			userSlug := item.Slug
			result.Users = append(result.Users, domain.User{
				UUID:      item.ItemUUID,
				Name:      item.Title,
				Slug:      &userSlug,
				AvatarUrl: imgUrl,
			})
		}
	}

	return &result, nil
}

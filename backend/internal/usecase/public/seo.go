package public

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"anirank/api/internal/domain"
)

type seoUsecase struct {
	animeRepo  domain.AnimeRepository
	songRepo   domain.SongRepository
	artistRepo domain.ArtistRepository
}

func NewSEOUsecase(ar domain.AnimeRepository, sr domain.SongRepository, artistR domain.ArtistRepository) domain.SEOUsecase {
	return &seoUsecase{
		animeRepo:  ar,
		songRepo:   sr,
		artistRepo: artistR,
	}
}

var (
	songRegex   = regexp.MustCompile(`^/songs/([^/]+)/([^/]+)$`)
	artistRegex = regexp.MustCompile(`^/artists/([^/]+)$`)
	animeRegex  = regexp.MustCompile(`^/animes/([^/]+)$`)
)

func (u *seoUsecase) getAPIURL() string {
	url := os.Getenv("PUBLIC_API_URL")
	if url == "" {
		return "https://anirank.work/api"
	}
	return strings.TrimSuffix(url, "/")
}

func (u *seoUsecase) GetMetadata(path string) (*domain.SEOData, error) {
	ctx := context.Background()
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "https://anirank.com"
	}
	apiURL := u.getAPIURL()

	siteName := "AniRank"
	defaultTitle := "AniRank - The Ultimate Anime Music Ranking Platform"
	defaultDesc := "Discover, rank, and listen to the best anime openings and endings. Create playlists, share with friends, and find your next favorite anime song."
	defaultImage := fmt.Sprintf("%s/og/home", apiURL)

	// 1. Home
	if path == "/" || path == "" {
		return &domain.SEOData{
			Title:       defaultTitle,
			Description: defaultDesc,
			Image:       fmt.Sprintf("%s/og/home", apiURL),
			URL:         appURL,
			Type:        "website",
			SiteName:    siteName,
		}, nil
	}

	// 2. Songs
	if matches := songRegex.FindStringSubmatch(path); len(matches) > 2 {
		animeSlug := matches[1]
		songSlug := matches[2]

		anime, err := u.animeRepo.GetBySlug(ctx, animeSlug)
		if err == nil {
			song, err := u.songRepo.GetByAnimeIDAndSlug(ctx, anime.ID, songSlug)
			if err == nil {
				// Compute song name
				songName := ""
				if song.SongRomaji != nil && *song.SongRomaji != "" {
					songName = *song.SongRomaji
				} else if song.SongEN != nil && *song.SongEN != "" {
					songName = *song.SongEN
				} else if song.SongJP != nil && *song.SongJP != "" {
					songName = *song.SongJP
				}

				// Get artists for the song
				artists, _ := u.songRepo.GetArtistsBySongID(ctx, song.ID, false)
				artistNames := []string{}
				for _, a := range artists {
					artistNames = append(artistNames, a.Name)
				}
				artistsStr := strings.Join(artistNames, ", ")

				title := fmt.Sprintf("%s - %s (%s) | %s", songName, artistsStr, anime.Title, siteName)
				desc := fmt.Sprintf("Listen to %s by %s, from the anime %s. Rank it and join the AniRank community.", songName, artistsStr, anime.Title)
				
				// Use dynamic OG image generator
				image := fmt.Sprintf("%s/og/song/%s/%s", apiURL, animeSlug, songSlug)

				return &domain.SEOData{
					Title:       title,
					Description: desc,
					Image:       image,
					URL:         appURL + path,
					Type:        "music.song",
					SiteName:    siteName,
				}, nil
			}
		}
	}

	// 3. Artists
	if matches := artistRegex.FindStringSubmatch(path); len(matches) > 1 {
		slug := matches[1]
		artist, err := u.artistRepo.GetBySlug(ctx, slug)
		if err == nil {
			title := fmt.Sprintf("%s - Anime Songs | %s", artist.Name, siteName)
			desc := fmt.Sprintf("Explore all anime openings and endings by %s. See their rankings and discography on AniRank.", artist.Name)
			
			// Use dynamic OG image generator
			image := fmt.Sprintf("%s/og/artist/%s", apiURL, slug)

			return &domain.SEOData{
				Title:       title,
				Description: desc,
				Image:       image,
				URL:         appURL + path,
				Type:        "profile",
				SiteName:    siteName,
			}, nil
		}
	}

	// 4. Anime
	if matches := animeRegex.FindStringSubmatch(path); len(matches) > 1 {
		slug := matches[1]
		anime, err := u.animeRepo.GetBySlug(ctx, slug)
		if err == nil {
			title := fmt.Sprintf("%s - Openings & Endings | %s", anime.Title, siteName)
			desc := fmt.Sprintf("Discover and rank all the themes from %s. Find the best songs and artists for this anime on AniRank.", anime.Title)
			
			// Use dynamic OG image generator
			image := fmt.Sprintf("%s/og/anime/%s", apiURL, slug)

			return &domain.SEOData{
				Title:       title,
				Description: desc,
				Image:       image,
				URL:         appURL + path,
				Type:        "video.tv_show",
				SiteName:    siteName,
			}, nil
		}
	}

	// Default fallback
	return &domain.SEOData{
		Title:       defaultTitle,
		Description: defaultDesc,
		Image:       defaultImage,
		URL:         appURL + path,
		Type:        "website",
		SiteName:    siteName,
	}, nil
}



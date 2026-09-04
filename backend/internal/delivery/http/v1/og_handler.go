package v1

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure/og"
	"anirank/api/internal/usecase/playlist"
	"anirank/api/internal/usecase/public"

	"github.com/gofiber/fiber/v2"
)

type OGHandler struct {
	generator       *og.Generator
	animeUsecase    *public.AnimeUsecase
	catalogUsecase  *public.CatalogUsecase
	playlistUsecase *playlist.PlaylistUsecase
	statsUsecase    domain.StatsUsecase
}

func NewOGHandler(g *og.Generator, a *public.AnimeUsecase, c *public.CatalogUsecase, p *playlist.PlaylistUsecase, s domain.StatsUsecase) *OGHandler {
	return &OGHandler{
		generator:       g,
		animeUsecase:    a,
		catalogUsecase:  c,
		playlistUsecase: p,
		statsUsecase:    s,
	}
}

// respondGenerated runs PNG generation under the concurrency semaphore, then caches and sends.
func (h *OGHandler) respondGenerated(c *fiber.Ctx, cacheKey string, nocache bool, maxAge int, alwaysCache bool, generate func() (image.Image, error)) error {
	var buffer bytes.Buffer
	acquired, err := h.generator.TryGenerate(func() error {
		img, genErr := generate()
		if genErr != nil {
			return genErr
		}
		return png.Encode(&buffer, img)
	})
	if !acquired {
		c.Set("Retry-After", "5")
		return c.Status(http.StatusServiceUnavailable).SendString("OG generation busy, retry shortly")
	}
	if err != nil {
		log.Printf("[OG Handler] generate failed: %v", err)
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	if alwaysCache || !nocache {
		_ = h.generator.SaveCache(cacheKey, buffer.Bytes())
	}
	if !nocache {
		c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	} else {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}
	c.Set("Content-Type", "image/png")
	return c.Send(buffer.Bytes())
}

func (h *OGHandler) serveCached(c *fiber.Ctx, cacheKey string, nocache bool, maxAge int) (bool, error) {
	if nocache {
		return false, nil
	}
	path := h.generator.GetCachePath(cacheKey)
	if _, err := os.Stat(path); err != nil {
		return false, nil
	}
	c.Set("Content-Type", "image/png")
	c.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	return true, c.SendFile(path)
}

func (h *OGHandler) SongOG(c *fiber.Ctx) error {
	animeSlug := c.Params("anime_slug")
	songSlug := c.Params("song_slug")
	cacheKey := fmt.Sprintf("song_v2_%s_%s_v%d", animeSlug, songSlug, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	if ok, err := h.serveCached(c, cacheKey, nocache, 86400); ok {
		return err
	}

	song, _, err := h.catalogUsecase.GetSongByAnimeSongSlug(c.Context(), nil, animeSlug, songSlug)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("Song not found")
	}

	title := song.Name
	artists := ""
	for i, a := range song.Artists {
		if i > 0 {
			artists += ", "
		}
		artists += a.Name
	}
	animeTitle := ""
	bgUrl := ""
	if song.Anime != nil {
		animeTitle = song.Anime.Title
		if song.Anime.BannerUrl != nil {
			bgUrl = *song.Anime.BannerUrl
		} else if song.Anime.CoverUrl != nil {
			bgUrl = *song.Anime.CoverUrl
		}
	}

	return h.respondGenerated(c, cacheKey, nocache, 86400, false, func() (image.Image, error) {
		return h.generator.GenerateSongOG(title, artists, animeTitle, song.Type, song.AverageRating, bgUrl)
	})
}

func (h *OGHandler) AnimeOG(c *fiber.Ctx) error {
	slug := c.Params("slug")
	cacheKey := fmt.Sprintf("anime_v1_%s_v%d", slug, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	if ok, err := h.serveCached(c, cacheKey, nocache, 86400); ok {
		return err
	}

	anime, err := h.animeUsecase.GetAnimeBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("Anime not found")
	}

	studios := ""
	for i, s := range anime.Studios {
		if i > 0 {
			studios += ", "
		}
		studios += s.Name
	}

	bannerUrl := ""
	if anime.BannerUrl != nil {
		bannerUrl = *anime.BannerUrl
	} else if anime.CoverUrl != nil {
		bannerUrl = *anime.CoverUrl
	}

	return h.respondGenerated(c, cacheKey, nocache, 86400, false, func() (image.Image, error) {
		return h.generator.GenerateAnimeOG(anime.Title, studios, anime.SongsCount, 0, bannerUrl)
	})
}

func (h *OGHandler) ArtistOG(c *fiber.Ctx) error {
	slug := c.Params("slug")
	cacheKey := fmt.Sprintf("artist_v2_%s_v%d", slug, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	if ok, err := h.serveCached(c, cacheKey, nocache, 86400); ok {
		return err
	}

	artist, _, totalSongs, err := h.catalogUsecase.GetSongsByArtistSlug(c.Context(), nil, slug, 1, 0, domain.SongFilters{})
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("Artist not found")
	}

	avatarUrl := ""
	if artist.AvatarUrl != nil {
		avatarUrl = *artist.AvatarUrl
	}

	bannerUrl := ""
	if artist.LatestBannerUrl != nil {
		bannerUrl = *artist.LatestBannerUrl
	}

	return h.respondGenerated(c, cacheKey, nocache, 86400, false, func() (image.Image, error) {
		return h.generator.GenerateArtistOG(artist.Name, totalSongs, int(artist.FavoritesCount), avatarUrl, bannerUrl)
	})
}

func (h *OGHandler) UserOG(c *fiber.Ctx) error {
	slug := c.Params("slug")
	cacheKey := fmt.Sprintf("user_v2_%s_v%d", slug, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	if ok, err := h.serveCached(c, cacheKey, nocache, 86400); ok {
		return err
	}

	user, err := h.catalogUsecase.GetUserBySlug(c.Context(), nil, slug)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("User not found")
	}

	avatarUrl := ""
	if user.AvatarUrl != nil {
		avatarUrl = *user.AvatarUrl
	}
	bannerUrl := ""
	if user.BannerUrl != nil {
		bannerUrl = *user.BannerUrl
	}

	return h.respondGenerated(c, cacheKey, nocache, 86400, false, func() (image.Image, error) {
		return h.generator.GenerateUserOG(user.Name, int(user.Level), int(user.XP), user.FollowersCount, user.RatingsCount, avatarUrl, bannerUrl)
	})
}

func (h *OGHandler) HomeOG(c *fiber.Ctx) error {
	cacheKey := fmt.Sprintf("home_v7_v%d", h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	if ok, err := h.serveCached(c, cacheKey, nocache, 3600); ok {
		return err
	}

	stats, err := h.statsUsecase.GetSiteStats(c.Context())
	if err != nil {
		stats = &domain.SiteStats{}
	}

	return h.respondGenerated(c, cacheKey, nocache, 3600, false, func() (image.Image, error) {
		return h.generator.GenerateHomeOG(
			stats.Overviews.TotalSongs,
			stats.Overviews.TotalUsers,
			stats.Overviews.TotalAnimes,
			stats.Overviews.TotalArtists,
		)
	})
}

func (h *OGHandler) PlaylistOG(c *fiber.Ctx) error {
	idStr := c.Params("pid")
	var id uint64
	fmt.Sscanf(idStr, "%d", &id)

	cacheKey := fmt.Sprintf("playlist_v1_%s_v%d", idStr, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	if ok, err := h.serveCached(c, cacheKey, nocache, 86400); ok {
		return err
	}

	playlist, err := h.playlistUsecase.GetPlaylist(c.Context(), id, nil)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("Playlist not found")
	}

	userName := ""
	if playlist.User != nil {
		userName = playlist.User.Name
	}

	bannerUrl := ""
	if playlist.BannerUrl != nil {
		bannerUrl = *playlist.BannerUrl
	}

	// Preserve previous behavior: always write disk cache for playlists.
	return h.respondGenerated(c, cacheKey, nocache, 86400, true, func() (image.Image, error) {
		return h.generator.GeneratePlaylistOG(playlist.Name, userName, playlist.SongCount, bannerUrl)
	})
}

func (h *OGHandler) FlushOGCache(c *fiber.Ctx) error {
	if err := h.generator.FlushCache(); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to flush OG cache: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "OG cache flushed successfully",
	})
}

func (h *OGHandler) RankingOG(c *fiber.Ctx) error {
	rankingType := c.Params("type") // "seasonal" or "global"
	if rankingType != "seasonal" && rankingType != "global" {
		return c.Status(http.StatusBadRequest).SendString("Invalid ranking type")
	}

	cacheKey := fmt.Sprintf("ranking_%s_v%d", rankingType, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	if ok, err := h.serveCached(c, cacheKey, nocache, 3600); ok {
		return err
	}

	ranking, err := h.catalogUsecase.GetSongRanking(c.Context(), nil, rankingType, "all", 3, 0)
	if err != nil {
		log.Printf("[OG Handler] Error getting ranking for OG: %v", err)
	}

	var songs []domain.Song
	if ranking != nil {
		songs = ranking.Songs
	}

	return h.respondGenerated(c, cacheKey, nocache, 3600, false, func() (image.Image, error) {
		return h.generator.GenerateRankingOG(rankingType, songs)
	})
}

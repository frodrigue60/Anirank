package v1

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"

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

func (h *OGHandler) SongOG(c *fiber.Ctx) error {
	animeSlug := c.Params("anime_slug")
	songSlug := c.Params("song_slug")
	cacheKey := fmt.Sprintf("song_v2_%s_%s_v%d", animeSlug, songSlug, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	// Check cache
	if !nocache {
		if data, ok := h.generator.GetCache(cacheKey); ok {
			c.Set("Content-Type", "image/png")
			c.Set("Cache-Control", "public, max-age=86400")
			return c.Send(data)
		}
	}

	song, _, err := h.catalogUsecase.GetSongByAnimeSongSlug(c.Context(), nil, animeSlug, songSlug)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("Song not found")
	}

	// Prepare data for generator
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

	img, err := h.generator.GenerateSongOG(title, artists, animeTitle, song.Type, song.AverageRating, bgUrl)
	if err != nil {
		fmt.Printf("[OG Handler] Error generating Song OG for %s/%s: %v\n", animeSlug, songSlug, err)
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Save to cache
	if !nocache {
		h.generator.SaveCache(cacheKey, buffer.Bytes())
		c.Set("Cache-Control", "public, max-age=86400")
	} else {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	c.Set("Content-Type", "image/png")
	return c.Send(buffer.Bytes())
}

func (h *OGHandler) AnimeOG(c *fiber.Ctx) error {
	slug := c.Params("slug")
	cacheKey := fmt.Sprintf("anime_v1_%s_v%d", slug, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	// Check cache
	if !nocache {
		if data, ok := h.generator.GetCache(cacheKey); ok {
			c.Set("Content-Type", "image/png")
			c.Set("Cache-Control", "public, max-age=86400")
			return c.Send(data)
		}
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

	img, err := h.generator.GenerateAnimeOG(anime.Title, studios, anime.SongsCount, 0, bannerUrl)
	if err != nil {
		fmt.Printf("[OG Handler] Error generating Anime OG for %s: %v\n", slug, err)
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Save to cache
	if !nocache {
		h.generator.SaveCache(cacheKey, buffer.Bytes())
		c.Set("Cache-Control", "public, max-age=86400")
	} else {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	c.Set("Content-Type", "image/png")
	return c.Send(buffer.Bytes())
}

func (h *OGHandler) ArtistOG(c *fiber.Ctx) error {
	slug := c.Params("slug")
	cacheKey := fmt.Sprintf("artist_v2_%s_v%d", slug, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	// Check cache
	if !nocache {
		if data, ok := h.generator.GetCache(cacheKey); ok {
			c.Set("Content-Type", "image/png")
			c.Set("Cache-Control", "public, max-age=86400")
			return c.Send(data)
		}
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

	img, err := h.generator.GenerateArtistOG(artist.Name, totalSongs, int(artist.FavoritesCount), avatarUrl, bannerUrl)
	if err != nil {
		fmt.Printf("[OG Handler] Error generating Artist OG for %s: %v\n", slug, err)
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Save to cache
	if !nocache {
		h.generator.SaveCache(cacheKey, buffer.Bytes())
		c.Set("Cache-Control", "public, max-age=86400")
	} else {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	c.Set("Content-Type", "image/png")
	return c.Send(buffer.Bytes())
}

func (h *OGHandler) UserOG(c *fiber.Ctx) error {
	slug := c.Params("slug")
	cacheKey := fmt.Sprintf("user_v2_%s_v%d", slug, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	// Check cache
	if !nocache {
		if data, ok := h.generator.GetCache(cacheKey); ok {
			c.Set("Content-Type", "image/png")
			c.Set("Cache-Control", "public, max-age=86400")
			return c.Send(data)
		}
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

	fmt.Printf("[OG User] slug=%s avatar_raw=%v banner_raw=%v avatarUrl=%s bannerUrl=%s\n", slug, user.Avatar, user.Banner, avatarUrl, bannerUrl)

	img, err := h.generator.GenerateUserOG(user.Name, int(user.Level), int(user.XP), user.FollowersCount, user.RatingsCount, avatarUrl, bannerUrl)
	if err != nil {
		fmt.Printf("[OG Handler] Error generating User OG for %s: %v\n", slug, err)
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Save to cache
	if !nocache {
		h.generator.SaveCache(cacheKey, buffer.Bytes())
		c.Set("Cache-Control", "public, max-age=86400")
	} else {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	c.Set("Content-Type", "image/png")
	return c.Send(buffer.Bytes())
}

func (h *OGHandler) HomeOG(c *fiber.Ctx) error {
	cacheKey := fmt.Sprintf("home_v6_v%d", h.generator.GetVersion())

	nocache := c.Query("nocache") == "true"

	// Check cache
	if !nocache {
		if data, ok := h.generator.GetCache(cacheKey); ok {
			c.Set("Content-Type", "image/png")
			c.Set("Cache-Control", "public, max-age=3600") // 1h for home stats
			return c.Send(data)
		}
	}

	stats, err := h.statsUsecase.GetSiteStats(c.Context())
	if err != nil {
		// Fallback to default if stats fail
		stats = &domain.SiteStats{}
	}

	img, err := h.generator.GenerateHomeOG(
		stats.Overviews.TotalSongs,
		stats.Overviews.TotalUsers,
		stats.Overviews.TotalAnimes,
		stats.Overviews.TotalArtists,
	)
	if err != nil {
		fmt.Printf("[OG Handler] Error generating Home OG: %v\n", err)
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Save to cache
	if !nocache {
		h.generator.SaveCache(cacheKey, buffer.Bytes())
		c.Set("Cache-Control", "public, max-age=3600")
	} else {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	c.Set("Content-Type", "image/png")
	return c.Send(buffer.Bytes())
}

func (h *OGHandler) PlaylistOG(c *fiber.Ctx) error {
	idStr := c.Params("pid")
	var id uint64
	fmt.Sscanf(idStr, "%d", &id)

	cacheKey := fmt.Sprintf("playlist_v1_%s_v%d", idStr, h.generator.GetVersion())
	nocache := c.Query("nocache") == "true"

	// Check cache
	if !nocache {
		if data, ok := h.generator.GetCache(cacheKey); ok {
			c.Set("Content-Type", "image/png")
			c.Set("Cache-Control", "public, max-age=86400")
			return c.Send(data)
		}
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

	img, err := h.generator.GeneratePlaylistOG(playlist.Name, userName, playlist.SongCount, bannerUrl)
	if err != nil {
		fmt.Printf("[OG Handler] Error generating Playlist OG for %s: %v\n", idStr, err)
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, img); err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}

	// Save to cache
	h.generator.SaveCache(cacheKey, buffer.Bytes())

	if !nocache {
		c.Set("Cache-Control", "public, max-age=86400")
	} else {
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	c.Set("Content-Type", "image/png")
	return c.Send(buffer.Bytes())
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

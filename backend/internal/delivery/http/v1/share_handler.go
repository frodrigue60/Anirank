package v1

import (
	"fmt"
	"os"
	"strings"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/playlist"
	"anirank/api/internal/usecase/public"

	"github.com/gofiber/fiber/v2"
)

type ShareHandler struct {
	animeUsecase    *public.AnimeUsecase
	catalogUsecase  *public.CatalogUsecase
	playlistUsecase *playlist.PlaylistUsecase
}

func NewShareHandler(a *public.AnimeUsecase, c *public.CatalogUsecase, p *playlist.PlaylistUsecase) *ShareHandler {
	return &ShareHandler{
		animeUsecase:    a,
		catalogUsecase:  c,
		playlistUsecase: p,
	}
}

func (h *ShareHandler) isBot(userAgent string) bool {
	userAgent = strings.ToLower(userAgent)
	bots := []string{
		"twitterbot",
		"facebookexternalhit",
		"discordbot",
		"slackbot",
		"telegrambot",
		"whatsapp",
		"googlebot",
		"bingbot",
		"linkedinbot",
		"embedly",
		"quora link preview",
		"outbrain",
		"pinterest",
		"vkshare",
	}
	for _, bot := range bots {
		if strings.Contains(userAgent, bot) {
			return true
		}
	}
	return false
}

func (h *ShareHandler) getFrontendURL() string {
	url := os.Getenv("FRONTEND_URL")
	if url == "" {
		return "https://anirank.work"
	}
	return strings.TrimSuffix(url, "/")
}

func (h *ShareHandler) getAPIURL() string {
	url := os.Getenv("PUBLIC_API_URL")
	if url == "" {
		return "http://localhost:8080/api"
	}
	return strings.TrimSuffix(url, "/")
}

func (h *ShareHandler) renderMeta(c *fiber.Ctx, title, description, image, urlType, slug string) error {
	frontendURL := h.getFrontendURL()
	fullURL := fmt.Sprintf("%s/%s/%s", frontendURL, urlType, slug)
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <meta name="description" content="%s">
    
    <!-- Open Graph -->
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:image" content="%s">
    <meta property="og:url" content="%s">
    <meta property="og:type" content="website">
    
    <!-- Twitter -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="%s">
    <meta name="twitter:description" content="%s">
    <meta name="twitter:image" content="%s">
    
    <meta http-equiv="refresh" content="0;url=%s">
</head>
<body>
    <p>Redirecting to <a href="%s">AniRank</a>...</p>
</body>
</html>`, title, description, title, description, image, fullURL, title, description, image, fullURL, fullURL)

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}

func (h *ShareHandler) AnimeShare(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if !h.isBot(c.Get("User-Agent")) {
		return c.Redirect(h.getFrontendURL() + "/animes/" + slug)
	}

	anime, err := h.animeUsecase.GetAnimeBySlug(c.Context(), slug)
	if err != nil {
		return c.Redirect(h.getFrontendURL() + "/animes/" + slug)
	}

	title := fmt.Sprintf("%s - AniRank", anime.Title)
	description := ""
	if anime.Description != nil {
		description = *anime.Description
	}
	if len(description) > 160 {
		description = description[:157] + "..."
	}
	image := fmt.Sprintf("%s/og/anime/%s", h.getAPIURL(), slug)

	return h.renderMeta(c, title, description, image, "animes", slug)
}

func (h *ShareHandler) SongShare(c *fiber.Ctx) error {
	animeSlug := c.Params("anime_slug")
	songSlug := c.Params("song_slug")
	if !h.isBot(c.Get("User-Agent")) {
		return c.Redirect(h.getFrontendURL() + "/songs/" + animeSlug + "/" + songSlug)
	}

	song, _, err := h.catalogUsecase.GetSongByAnimeSongSlug(c.Context(), nil, animeSlug, songSlug)
	if err != nil {
		return c.Redirect(h.getFrontendURL() + "/songs/" + animeSlug + "/" + songSlug)
	}

	title := fmt.Sprintf("%s - %s - AniRank", song.Name, song.Anime.Title)
	artists := ""
	for i, a := range song.Artists {
		if i > 0 {
			artists += ", "
		}
		artists += a.Name
	}
	description := fmt.Sprintf("Listen to %s by %s from %s.", song.Name, artists, song.Anime.Title)
	image := fmt.Sprintf("%s/og/song/%s/%s", h.getAPIURL(), animeSlug, songSlug)

	return h.renderMeta(c, title, description, image, "songs/"+animeSlug, songSlug)
}

func (h *ShareHandler) ArtistShare(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if !h.isBot(c.Get("User-Agent")) {
		return c.Redirect(h.getFrontendURL() + "/artists/" + slug)
	}

	artist, _, _, err := h.catalogUsecase.GetSongsByArtistSlug(c.Context(), nil, slug, 1, 0, domain.SongFilters{})
	if err != nil {
		return c.Redirect(h.getFrontendURL() + "/artists/" + slug)
	}

	title := fmt.Sprintf("%s - Artist - AniRank", artist.Name)
	description := fmt.Sprintf("Discover all anime theme songs by %s on AniRank.", artist.Name)
	image := fmt.Sprintf("%s/og/artist/%s", h.getAPIURL(), slug)

	return h.renderMeta(c, title, description, image, "artists", slug)
}

func (h *ShareHandler) UserShare(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if !h.isBot(c.Get("User-Agent")) {
		return c.Redirect(h.getFrontendURL() + "/users/" + slug)
	}

	user, err := h.catalogUsecase.GetUserBySlug(c.Context(), nil, slug)
	if err != nil {
		return c.Redirect(h.getFrontendURL() + "/users/" + slug)
	}

	userSlug := ""
	if user.Slug != nil {
		userSlug = *user.Slug
	}

	title := fmt.Sprintf("%s's Profile - AniRank", user.Name)
	description := fmt.Sprintf("Check out %s's anime theme song favorites and stats on AniRank.", user.Name)
	image := fmt.Sprintf("%s/og/user/%s", h.getAPIURL(), userSlug)

	return h.renderMeta(c, title, description, image, "users", userSlug)
}

func (h *ShareHandler) PlaylistShare(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var id uint64
	fmt.Sscanf(idStr, "%d", &id)

	if !h.isBot(c.Get("User-Agent")) {
		return c.Redirect(h.getFrontendURL() + "/playlists/" + idStr)
	}

	playlist, err := h.playlistUsecase.GetPlaylist(c.Context(), id, nil)
	if err != nil {
		return c.Redirect(h.getFrontendURL() + "/playlists/" + idStr)
	}

	userName := ""
	if playlist.User != nil {
		userName = playlist.User.Name
	}

	title := fmt.Sprintf("%s - Playlist by %s - AniRank", playlist.Name, userName)
	description := fmt.Sprintf("Check out the \"%s\" playlist curated by %s on AniRank. Featuring %d songs.", playlist.Name, userName, playlist.SongCount)
	image := fmt.Sprintf("%s/og/playlist/%d", h.getAPIURL(), id)

	return h.renderMeta(c, title, description, image, "playlists", idStr)
}

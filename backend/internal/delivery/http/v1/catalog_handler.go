package v1

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	"anirank/api/internal/usecase/public"

	"github.com/gofiber/fiber/v2"
)

// CatalogHandler serves all public catalog browsing endpoints that
// were missing from the initial API build.
type CatalogHandler struct {
	usecase *public.CatalogUsecase
}

func NewCatalogHandler(u *public.CatalogUsecase) *CatalogHandler {
	return &CatalogHandler{usecase: u}
}

func (h *CatalogHandler) getUserID(c *fiber.Ctx) *uint64 {
	val := c.Locals("user_id")
	if id, ok := val.(uint64); ok {
		return &id
	}
	if f, ok := val.(float64); ok {
		id := uint64(f)
		return &id
	}
	return nil
}

// ─── Songs ───

// SongIndex handles GET /api/songs
func (h *CatalogHandler) SongIndex(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	yearID, _ := strconv.ParseUint(c.Query("year_id", "0"), 10, 64)
	seasonID, _ := strconv.ParseUint(c.Query("season_id", "0"), 10, 64)

	filters := domain.SongFilters{
		Search:   c.Query("name", ""),
		YearID:   yearID,
		SeasonID: seasonID,
		Year:     c.Query("year", ""),
		Season:   c.Query("season", ""),
		Genre:    c.Query("genre", ""),
		Type:     c.Query("type", ""),
		Sort:     c.Query("sort", ""),
	}

	userID := h.getUserID(c)
	songs, total, err := h.usecase.GetPaginatedSongs(c.Context(), userID, limit, offset, filters)
	if err != nil {
		return err
	}

	songDTOs := make([]dto.SongMinimalDTO, len(songs))
	for i, s := range songs {
		songDTOs[i] = dto.ToSongMinimalDTO(&s)
	}

	return c.JSON(paginatedResponse(c, songDTOs, total, page, limit))
}

// SongShow handles GET /api/songs/:anime_slug/:song_slug
func (h *CatalogHandler) SongShow(c *fiber.Ctx) error {
	userID := h.getUserID(c)

	ctx := context.WithValue(c.Context(), "client_ip", c.IP())
	song, related, err := h.usecase.GetSongByAnimeSongSlug(ctx, userID, c.Params("anime_slug"), c.Params("song_slug"))
	if err != nil {
		return err
	}

	relatedDTOs := make([]dto.SongMinimalDTO, len(related))
	for i, r := range related {
		relatedDTOs[i] = dto.ToSongMinimalDTO(&r)
	}

	return c.JSON(fiber.Map{
		"data":    dto.ToSongDTO(song),
		"related": relatedDTOs,
	})
}

// SongRanking handles GET /api/songs/ranking/:type
func (h *CatalogHandler) SongRanking(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	rankingType := c.Params("type", "global")
	songType := c.Query("type", "all")

	userID := h.getUserID(c)
	ranking, err := h.usecase.GetSongRanking(c.Context(), userID, rankingType, songType, limit, offset)
	if err != nil {
		return err
	}

	songDTOs := make([]dto.SongMinimalDTO, len(ranking.Songs))
	for i, s := range ranking.Songs {
		songDTOs[i] = dto.ToSongMinimalDTO(&s)
	}

	return c.JSON(paginatedResponse(c, songDTOs, ranking.Total, page, limit))
}

// ─── Artists ───

// ArtistIndex handles GET /api/artists
func (h *CatalogHandler) ArtistIndex(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	filters := domain.ArtistFilters{
		Search: c.Query("name", ""),
		Sort:   c.Query("sort", ""),
	}

	artists, total, err := h.usecase.GetPaginatedArtists(c.Context(), limit, offset, filters)
	if err != nil {
		return err
	}

	artistDTOs := make([]dto.ArtistDTO, len(artists))
	for i, a := range artists {
		artistDTOs[i] = dto.ToArtistDTO(&a)
	}

	return c.JSON(paginatedResponse(c, artistDTOs, total, page, limit))
}

func (h *CatalogHandler) ArtistShow(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	filters := domain.SongFilters{
		Search: c.Query("name", ""),
		Year:   c.Query("year", ""),
		Season: c.Query("season", ""),
		Genre:  c.Query("genre", ""),
		Type:   c.Query("type", ""),
		Sort:   c.Query("sort", "recent"),
	}

	userID := h.getUserID(c)
	artist, songs, total, err := h.usecase.GetSongsByArtistSlug(c.Context(), userID, c.Params("slug"), limit, offset, filters)
	if err != nil {
		return err
	}

	songDTOs := make([]dto.SongMinimalDTO, len(songs))
	for i, s := range songs {
		songDTOs[i] = dto.ToSongMinimalDTO(&s)
	}

	response := paginatedResponse(c, songDTOs, total, page, limit)
	response["artist"] = dto.ToArtistDTO(artist)

	return c.JSON(response)
}

// ─── Studios ───

// StudioIndex handles GET /api/studios
func (h *CatalogHandler) StudioIndex(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	filters := domain.StudioFilters{
		Search: c.Query("name", ""),
		Sort:   c.Query("sort", "name_asc"),
	}

	studios, total, err := h.usecase.GetPaginatedStudios(c.Context(), limit, offset, filters)
	if err != nil {
		return err
	}

	studioDTOs := make([]dto.StudioDTO, len(studios))
	for i, s := range studios {
		studioDTOs[i] = dto.ToStudioDTO(&s)
	}

	return c.JSON(paginatedResponse(c, studioDTOs, total, page, limit))
}

// StudioShow handles GET /api/studios/:slug
func (h *CatalogHandler) StudioShow(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	studio, animes, total, err := h.usecase.GetAnimesByStudioSlug(c.Context(), c.Params("slug"), limit, offset)
	if err != nil {
		return err
	}

	animeDTOs := make([]dto.AnimeMinimalDTO, len(animes))
	for i, a := range animes {
		animeDTOs[i] = dto.ToAnimeMinimalDTO(&a)
	}

	return c.JSON(fiber.Map{
		"studio": dto.ToStudioDTO(studio),
		"data":   paginatedResponse(c, animeDTOs, total, page, limit),
	})
}

// ─── Producers ───

// ProducerIndex handles GET /api/producers
func (h *CatalogHandler) ProducerIndex(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	filters := domain.ProducerFilters{
		Search: c.Query("name", ""),
		Sort:   c.Query("sort", "name_asc"),
	}

	producers, total, err := h.usecase.GetPaginatedProducers(c.Context(), limit, offset, filters)
	if err != nil {
		return err
	}

	producerDTOs := make([]dto.ProducerDTO, len(producers))
	for i, p := range producers {
		producerDTOs[i] = dto.ToProducerDTO(&p)
	}

	return c.JSON(paginatedResponse(c, producerDTOs, total, page, limit))
}

// ProducerShow handles GET /api/producers/:slug
func (h *CatalogHandler) ProducerShow(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	producer, animes, total, err := h.usecase.GetAnimesByProducerSlug(c.Context(), c.Params("slug"), limit, offset)
	if err != nil {
		return err
	}

	animeDTOs := make([]dto.AnimeMinimalDTO, len(animes))
	for i, a := range animes {
		animeDTOs[i] = dto.ToAnimeMinimalDTO(&a)
	}

	return c.JSON(fiber.Map{
		"producer": dto.ToProducerDTO(producer),
		"data":     paginatedResponse(c, animeDTOs, total, page, limit),
	})
}

// ─── Playlists ───

// PlaylistIndex handles GET /api/playlists
func (h *CatalogHandler) PlaylistIndex(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	filters := domain.PlaylistFilters{
		Search: c.Query("name", ""),
	}

	playlists, total, err := h.usecase.GetPaginatedPlaylists(c.Context(), limit, offset, filters)
	if err != nil {
		return err
	}

	dtoPlaylists := make([]dto.PlaylistMinimalDTO, len(playlists))
	for i, p := range playlists {
		dtoPlaylists[i] = dto.ToPlaylistMinimalDTO(&p)
	}

	return c.JSON(paginatedResponse(c, dtoPlaylists, total, page, limit))
}

// ─── Users ───

// UserProfile handles GET /api/users/:slug
func (h *CatalogHandler) UserProfile(c *fiber.Ctx) error {
	requestingUserID := h.getUserID(c)
	user, err := h.usecase.GetUserBySlug(c.Context(), requestingUserID, c.Params("slug"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": dto.ToUserDTO(user)})
}

// UserPlaylists handles GET /api/users/:slug/playlists
func (h *CatalogHandler) UserPlaylists(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	requestingUserID := h.getUserID(c)
	playlists, total, err := h.usecase.GetUserPlaylists(c.Context(), requestingUserID, c.Params("slug"), limit, offset)
	if err != nil {
		return err
	}

	dtoPlaylists := make([]dto.PlaylistMinimalDTO, len(playlists))
	for i, p := range playlists {
		dtoPlaylists[i] = dto.ToPlaylistMinimalDTO(&p)
	}

	return c.JSON(paginatedResponse(c, dtoPlaylists, total, page, limit))
}

type userFavoritesReq struct {
	UserID   string `json:"user_id"`
	UserUUID string `json:"user_uuid"`
	Page     int    `json:"page"`
}

// UserFavorites handles POST /api/users/favorites
func (h *CatalogHandler) UserFavorites(c *fiber.Ctx) error {
	var req userFavoritesReq
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request payload", err)
	}

	limit, offset := parsePagination(c, 24)
	page := req.Page
	if page < 1 {
		page = 1
	}
	offset = (page - 1) * limit

	userID := req.UserID
	if userID == "" {
		userID = req.UserUUID
	}

	songs, total, err := h.usecase.GetUserFavorites(c.Context(), userID, limit, offset)
	if err != nil {
		return err
	}

	songDTOs := make([]dto.SongMinimalDTO, len(songs))
	for i, s := range songs {
		songDTOs[i] = dto.ToSongMinimalDTO(&s)
	}

	return c.JSON(paginatedResponse(c, songDTOs, total, page, limit))
}

// UserArtistFavorites handles POST /api/users/artists/favorites
func (h *CatalogHandler) UserArtistFavorites(c *fiber.Ctx) error {
	var req userFavoritesReq
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request payload", err)
	}

	limit, offset := parsePagination(c, 24)
	page := req.Page
	if page < 1 {
		page = 1
	}
	offset = (page - 1) * limit

	userID := req.UserID
	if userID == "" {
		userID = req.UserUUID
	}

	artists, total, err := h.usecase.GetUserFavoriteArtists(c.Context(), userID, limit, offset)
	if err != nil {
		return err
	}

	artistDTOs := make([]dto.ArtistDTO, len(artists))
	for i, a := range artists {
		artistDTOs[i] = dto.ToArtistDTO(&a)
	}

	return c.JSON(paginatedResponse(c, artistDTOs, total, page, limit))
}

// UserFollowers handles GET /api/users/:slug/followers
func (h *CatalogHandler) UserFollowers(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	users, total, err := h.usecase.GetUserFollowers(c.Context(), c.Params("slug"), limit, offset)
	if err != nil {
		return err
	}

	userDTOs := make([]dto.UserMinimalDTO, len(users))
	for i, u := range users {
		userDTOs[i] = dto.ToUserMinimalDTO(&u)
	}

	return c.JSON(paginatedResponse(c, userDTOs, total, page, limit))
}

// UserFollowing handles GET /api/users/:slug/following
func (h *CatalogHandler) UserFollowing(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	users, total, err := h.usecase.GetUserFollowing(c.Context(), c.Params("slug"), limit, offset)
	if err != nil {
		return err
	}

	userDTOs := make([]dto.UserMinimalDTO, len(users))
	for i, u := range users {
		userDTOs[i] = dto.ToUserMinimalDTO(&u)
	}

	return c.JSON(paginatedResponse(c, userDTOs, total, page, limit))
}

// UserAnilistList handles GET /api/users/:slug/anilist-list
func (h *CatalogHandler) UserAnilistList(c *fiber.Ctx) error {
	status := c.Query("status", "ALL") // AniList status: CURRENT, PLANNING, COMPLETED, DROPPED, PAUSED
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	items, total, err := h.usecase.GetUserAnilistList(c.Context(), c.Params("slug"), status, page, limit)
	if err != nil {
		return err
	}

	return c.JSON(paginatedResponse(c, items, total, page, limit))
}

// ─── Home ───

// Home handles GET /api/home
func (h *CatalogHandler) Home(c *fiber.Ctx) error {
	userID := h.getUserID(c)
	data, err := h.usecase.GetHomeData(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": dto.ToHomeDTO(data)})
}

// UserRanking handles GET /api/users/ranking
func (h *CatalogHandler) UserRanking(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	sortBy := c.Query("sort", "xp")
	users, total, err := h.usecase.GetUserRanking(c.Context(), sortBy, limit, offset)
	if err != nil {
		return err
	}

	// RankingUser embeds User
	userDTOs := make([]dto.UserMinimalDTO, len(users))
	for i, u := range users {
		userDTOs[i] = dto.ToUserMinimalDTO(&u.User)
	}

	return c.JSON(paginatedResponse(c, userDTOs, total, page, limit))
}

// GetSitemap handles GET /api/v1/catalog/sitemap
func (h *CatalogHandler) GetSitemap(c *fiber.Ctx) error {
	data, err := h.usecase.GetSitemapData(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": data})
}

// GetSitemapXML handles GET /api/v1/catalog/sitemap.xml
func (h *CatalogHandler) GetSitemapXML(c *fiber.Ctx) error {
	data, err := h.usecase.GetSitemapData(c.Context())
	if err != nil {
		return err
	}

	siteURL := os.Getenv("APP_URL")
	if siteURL == "" {
		siteURL = "https://anirank.work"
	}

	// Manual XML construction to match the frontend template
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>` + siteURL + `/</loc>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>` + siteURL + `/animes</loc>
    <changefreq>weekly</changefreq>
    <priority>0.9</priority>
  </url>
  <url>
    <loc>` + siteURL + `/songs</loc>
    <changefreq>weekly</changefreq>
    <priority>0.9</priority>
  </url>
  <url>
    <loc>` + siteURL + `/artists</loc>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>`

	for _, item := range data {
		xml += `
  <url>
    <loc>` + siteURL + item.Loc + `</loc>
    <lastmod>` + item.LastMod.Format(time.RFC3339) + `</lastmod>
    <changefreq>` + item.ChangeFreq + `</changefreq>
    <priority>` + fmt.Sprintf("%.1f", item.Priority) + `</priority>
  </url>`
	}

	xml += `
</urlset>`

	c.Set("Content-Type", "application/xml")
	return c.SendString(xml)
}

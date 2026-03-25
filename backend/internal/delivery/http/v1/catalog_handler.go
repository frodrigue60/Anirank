package v1

import (
	"context"
	"math"
	"strconv"
	"strings"

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

func parsePagination(c *fiber.Ctx, defaultLimit int) (int, int) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := defaultLimit
	offset := (page - 1) * limit
	return limit, offset
}

// paginatedResponse builds a Laravel-compatible paginated JSON envelope
// that the SvelteKit frontend expects: { data, current_page, last_page, total, per_page, next_page_url, prev_page_url }
func paginatedResponse(c *fiber.Ctx, items interface{}, total int, page, perPage int) fiber.Map {
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	path := c.Path()
	queryParams := c.Queries()

	// Strip /api prefix if present to avoid doubling with frontend baseURL
	cleanPath := strings.TrimPrefix(path, "/api")

	buildURL := func(p int) string {
		q := make(map[string]string)
		for k, v := range queryParams {
			q[k] = v
		}
		q["page"] = strconv.Itoa(p)
		
		// Return full URL to ensure axios doesn't prepend baseURL redundantly,
		// but keep the path relative to API root if that's preferred.
		// Actually, using the absolute domain path without /api is best for the current frontend api client.
		u := cleanPath + "?"
		for k, v := range q {
			u += k + "=" + v + "&"
		}
		return u[:len(u)-1]
	}

	response := fiber.Map{
		"data": items,
		"pagination": fiber.Map{
			"total":        total,
			"per_page":     perPage,
			"current_page": page,
			"last_page":    lastPage,
			"has_more":     page < lastPage,
		},
		"links": fiber.Map{
			"self": buildURL(page),
		},
	}

	if page < lastPage {
		response["links"].(fiber.Map)["next"] = buildURL(page + 1)
	} else {
		response["links"].(fiber.Map)["next"] = nil
	}

	if page > 1 {
		response["links"].(fiber.Map)["prev"] = buildURL(page - 1)
	} else {
		response["links"].(fiber.Map)["prev"] = nil
	}

	return response
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
		"success": true,
		"song":    dto.ToSongDTO(song),
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

	yearID, _ := strconv.ParseUint(c.Query("year_id", "0"), 10, 64)
	seasonID, _ := strconv.ParseUint(c.Query("season_id", "0"), 10, 64)

	filters := domain.SongFilters{
		Search:   c.Query("name", ""),
		YearID:   yearID,
		SeasonID: seasonID,
		Type:     c.Query("type", ""),
		Sort:     c.Query("sort", ""),
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

	return c.JSON(fiber.Map{
		"success": true,
		"artist":  dto.ToArtistDTO(artist),
		"data":    paginatedResponse(c, songDTOs, total, page, limit),
	})
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
		"success": true,
		"studio":  dto.ToStudioDTO(studio),
		"data":    paginatedResponse(c, animeDTOs, total, page, limit),
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
		"success":  true,
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
	// For now we return playlists as is since we didn't define a PlaylistDTO yet
	return c.JSON(paginatedResponse(c, playlists, total, page, limit))
}

// ─── Users ───

// UserProfile handles GET /api/users/:slug
func (h *CatalogHandler) UserProfile(c *fiber.Ctx) error {
	requestingUserID := h.getUserID(c)
	user, err := h.usecase.GetUserBySlug(c.Context(), requestingUserID, c.Params("slug"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "user": dto.ToUserDTO(user)})
}

// UserPlaylists handles GET /api/users/:slug/playlists
func (h *CatalogHandler) UserPlaylists(c *fiber.Ctx) error {
	requestingUserID := h.getUserID(c)
	playlists, err := h.usecase.GetUserPlaylists(c.Context(), requestingUserID, c.Params("slug"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "playlists": playlists})
}

type userFavoritesReq struct {
	UserID uint64 `json:"user_id"`
	Page   int    `json:"page"`
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

	songs, total, err := h.usecase.GetUserFavorites(c.Context(), req.UserID, limit, offset)
	if err != nil {
		return err
	}

	songDTOs := make([]dto.SongMinimalDTO, len(songs))
	for i, s := range songs {
		songDTOs[i] = dto.ToSongMinimalDTO(&s)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    paginatedResponse(c, songDTOs, total, page, limit),
	})
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

	artists, total, err := h.usecase.GetUserFavoriteArtists(c.Context(), req.UserID, limit, offset)
	if err != nil {
		return err
	}

	artistDTOs := make([]dto.ArtistDTO, len(artists))
	for i, a := range artists {
		artistDTOs[i] = dto.ToArtistDTO(&a)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    paginatedResponse(c, artistDTOs, total, page, limit),
	})
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

// ─── Home ───

// Home handles GET /api/home
func (h *CatalogHandler) Home(c *fiber.Ctx) error {
	userID := h.getUserID(c)
	data, err := h.usecase.GetHomeData(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": dto.ToHomeDTO(data)})
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
	return c.JSON(fiber.Map{"success": true, "data": data})
}

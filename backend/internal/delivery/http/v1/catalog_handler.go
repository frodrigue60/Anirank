package v1

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/public"

	"github.com/gofiber/fiber/v2"
)

// CatalogHandler serves all public catalog browsing endpoints that
// were missing from the initial API build.
type CatalogHandler struct {
	usecase *public.CatalogUsecase
}

func NewCatalogHandler(u *public.CatalogUsecase) *CatalogHandler {
	os.Stderr.WriteString("[DEBUG] CatalogHandler Initialized\n")
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
// that the SvelteKit frontend expects: { data, current_page, last_page, total, per_page }
func paginatedResponse(items interface{}, total int, page, perPage int) fiber.Map {
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}
	return fiber.Map{
		"data":         items,
		"current_page": page,
		"last_page":    lastPage,
		"per_page":     perPage,
		"total":        total,
	}
}

// ─── Songs ───

// SongIndex handles GET /api/songs
// @Summary List Songs
// @Tags Songs
// @Produce json
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object
// @Router /songs [get]
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
	return c.JSON(paginatedResponse(songs, total, page, limit))
}

// SongShow handles GET /api/songs/:anime_slug/:song_slug
// @Summary Get Song details
// @Tags Songs
// @Produce json
// @Param anime_slug path string true "Anime Slug"
// @Param song_slug path string true "Song Slug"
// @Success 200 {object} object{song=domain.Song,related=[]domain.Song}
// @Router /songs/{anime_slug}/{song_slug} [get]
func (h *CatalogHandler) SongShow(c *fiber.Ctx) error {
	userID := h.getUserID(c)

	ctx := context.WithValue(c.Context(), "client_ip", c.IP())
	song, related, err := h.usecase.GetSongByAnimeSongSlug(ctx, userID, c.Params("anime_slug"), c.Params("song_slug"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "song": song, "related": related})
}

// SongRanking handles GET /api/songs/ranking/:type
// @Summary Song Rankings
// @Tags Songs
// @Produce json
// @Param type path string true "Ranking type (global or seasonal)"
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object{data=[]domain.Song}
// @Router /songs/ranking/{type} [get]
func (h *CatalogHandler) SongRanking(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 50)
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

	return c.JSON(fiber.Map{
		"songs":          paginatedResponse(ranking.Songs, ranking.Total, page, limit),
		"current_season": ranking.CurrentSeason,
		"current_year":   ranking.CurrentYear,
	})
}

// ─── Artists ───

// ArtistIndex handles GET /api/artists
// @Summary List Artists
// @Tags Artists
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param name query string false "Search by name"
// @Success 200 {object} object
// @Router /artists [get]
func (h *CatalogHandler) ArtistIndex(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 48)
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
	return c.JSON(fiber.Map{"artists": paginatedResponse(artists, total, page, limit)})
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

	var userID *uint64
	val := c.Locals("user_id")
	if id, ok := val.(uint64); ok {
		userID = &id
	} else if f, ok := val.(float64); ok {
		id := uint64(f)
		userID = &id
	}

	artist, songs, total, err := h.usecase.GetSongsByArtistSlug(c.Context(), userID, c.Params("slug"), limit, offset, filters)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"artist":  artist,
		"songs":   paginatedResponse(songs, total, page, limit),
	})
}

// ─── Studios ───

// StudioIndex handles GET /api/studios
// @Summary List Studios
// @Tags Studios
// @Produce json
// @Param page query int false "Page number"
// @Param search query string false "Search by name"
// @Success 200 {object} object
// @Router /studios [get]
func (h *CatalogHandler) StudioIndex(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 48)
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
	return c.JSON(fiber.Map{"studios": paginatedResponse(studios, total, page, limit)})
}

// StudioShow handles GET /api/studios/:slug
// @Summary Get Studio details and animes
// @Tags Studios
// @Produce json
// @Param slug path string true "Studio Slug"
// @Success 200 {object} object{studio=domain.Studio,animes=[]domain.Anime}
// @Router /studios/{slug} [get]
func (h *CatalogHandler) StudioShow(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 24)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	os.Stderr.WriteString(fmt.Sprintf("[DEBUG] StudioShow Request - Slug: %s\n", c.Params("slug")))

	studio, animes, total, err := h.usecase.GetAnimesByStudioSlug(c.Context(), c.Params("slug"), limit, offset)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("[DEBUG] StudioShow Error: %v\n", err))
		return err
	}
	os.Stderr.WriteString(fmt.Sprintf("[DEBUG] StudioShow Success - Studio: %s\n", studio.Name))
	return c.JSON(fiber.Map{
		"success": true,
		"studio":  studio,
		"animes":  paginatedResponse(animes, total, page, limit),
	})
}

// ─── Producers ───

// ProducerIndex handles GET /api/producers
// @Summary List Producers
// @Tags Producers
// @Produce json
// @Param page query int false "Page number"
// @Param search query string false "Search by name"
// @Success 200 {object} object
// @Router /producers [get]
func (h *CatalogHandler) ProducerIndex(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 48)
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
	return c.JSON(fiber.Map{"producers": paginatedResponse(producers, total, page, limit)})
}

// ProducerShow handles GET /api/producers/:slug
// @Summary Get Producer details and animes
// @Tags Producers
// @Produce json
// @Param slug path string true "Producer Slug"
// @Success 200 {object} object{producer=domain.Producer,animes=[]domain.Anime}
// @Router /producers/{slug} [get]
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
	return c.JSON(fiber.Map{
		"success":  true,
		"producer": producer,
		"animes":   paginatedResponse(animes, total, page, limit),
	})
}

// ─── Playlists ───

// PlaylistIndex handles GET /api/playlists
// @Summary List Public Playlists
// @Tags Playlists
// @Produce json
// @Param page query int false "Page number"
// @Param name query string false "Search by name"
// @Success 200 {object} object
// @Router /playlists [get]
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
	return c.JSON(fiber.Map{"playlists": paginatedResponse(playlists, total, page, limit)})
}

// ─── Users ───

// UserProfile handles GET /api/users/:slug
// @Summary Get User Profile
// @Tags Users
// @Produce json
// @Param slug path string true "User Slug"
// @Success 200 {object} object{user=domain.User}
// @Router /users/{slug} [get]
func (h *CatalogHandler) UserProfile(c *fiber.Ctx) error {
	var requestingUserID *uint64
	val := c.Locals("user_id")
	if id, ok := val.(uint64); ok {
		requestingUserID = &id
	} else if f, ok := val.(float64); ok {
		id := uint64(f)
		requestingUserID = &id
	}

	user, err := h.usecase.GetUserBySlug(c.Context(), requestingUserID, c.Params("slug"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "user": user})
}

// UserPlaylists handles GET /api/users/:slug/playlists
// @Summary Get User Playlists
// @Tags Users
// @Produce json
// @Param slug path string true "User Slug"
// @Success 200 {object} object{playlists=[]domain.Playlist}
// @Router /users/{slug}/playlists [get]
func (h *CatalogHandler) UserPlaylists(c *fiber.Ctx) error {
	var requestingUserID *uint64
	val := c.Locals("user_id")
	if id, ok := val.(uint64); ok {
		requestingUserID = &id
	} else if f, ok := val.(float64); ok {
		id := uint64(f)
		requestingUserID = &id
	}

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
// @Summary Get User Favorite Songs
// @Tags Users
// @Produce json
// @Accept json
// @Param request body object{user_id=int,page=int} true "Favorites query"
// @Success 200 {object} object{songs=object}
// @Router /users/favorites [post]
func (h *CatalogHandler) UserFavorites(c *fiber.Ctx) error {
	var req userFavoritesReq
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request payload", err)
	}

	limit := 24
	page := req.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	songs, total, err := h.usecase.GetUserFavorites(c.Context(), req.UserID, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"songs":   paginatedResponse(songs, total, page, limit),
	})
}

// UserArtistFavorites handles POST /api/users/artists/favorites
// @Summary Get User Favorite Artists
// @Tags Users
// @Produce json
// @Accept json
// @Param request body object{user_id=int,page=int} true "Favorites query"
// @Success 200 {object} object{artists=object}
// @Router /users/artists/favorites [post]
func (h *CatalogHandler) UserArtistFavorites(c *fiber.Ctx) error {
	var req userFavoritesReq
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request payload", err)
	}

	limit := 24
	page := req.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	artists, total, err := h.usecase.GetUserFavoriteArtists(c.Context(), req.UserID, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"artists": paginatedResponse(artists, total, page, limit),
	})
}

// UserFollowers handles GET /api/users/:slug/followers
// @Summary Get User Followers
// @Tags Users
// @Produce json
// @Param slug path string true "User Slug"
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object
// @Router /users/{slug}/followers [get]
func (h *CatalogHandler) UserFollowers(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 48)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	users, total, err := h.usecase.GetUserFollowers(c.Context(), c.Params("slug"), limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(paginatedResponse(users, total, page, limit))
}

// UserFollowing handles GET /api/users/:slug/following
// @Summary Get User Following
// @Tags Users
// @Produce json
// @Param slug path string true "User Slug"
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object
// @Router /users/{slug}/following [get]
func (h *CatalogHandler) UserFollowing(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 48)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	users, total, err := h.usecase.GetUserFollowing(c.Context(), c.Params("slug"), limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(paginatedResponse(users, total, page, limit))
}

// ─── Home ───

// Home handles GET /api/home
// @Summary Homepage Data
// @Tags Home
// @Produce json
// @Success 200 {object} public.HomeData
// @Router /home [get]
func (h *CatalogHandler) Home(c *fiber.Ctx) error {
	userID := h.getUserID(c)
	data, err := h.usecase.GetHomeData(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

// UserRanking handles GET /api/users/ranking
// @Summary User Ranking
// @Tags Users
// @Produce json
// @Param sort query string false "Sort by (xp, ratings, comments)" default(xp)
// @Param page query int false "Page number" default(1)
// @Success 200 {object} object
// @Router /users/ranking [get]
func (h *CatalogHandler) UserRanking(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 50)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}

	sortBy := c.Query("sort", "xp")
	users, total, err := h.usecase.GetUserRanking(c.Context(), sortBy, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(paginatedResponse(users, total, page, limit))
}

// GetSitemap handles GET /api/v1/catalog/sitemap
// @Summary Get Sitemap Data
// @Tags Catalog
// @Produce json
// @Success 200 {object} object{data=[]domain.SitemapItem}
// @Router /catalog/sitemap [get]
func (h *CatalogHandler) GetSitemap(c *fiber.Ctx) error {
	data, err := h.usecase.GetSitemapData(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

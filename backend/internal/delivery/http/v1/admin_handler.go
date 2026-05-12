package v1

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/admin"
	"bufio"
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"reflect"
)

type AdminHandler struct {
	usecase      *admin.AdminUsecase
	songRepo     domain.SongRepository
	userRepo     domain.UserRepository
	animeRepo    domain.AnimeRepository
	artistRepo   domain.ArtistRepository
	playlistRepo domain.PlaylistRepository
}

func NewAdminHandler(
	usecase *admin.AdminUsecase,
	songRepo domain.SongRepository,
	userRepo domain.UserRepository,
	animeRepo domain.AnimeRepository,
	artistRepo domain.ArtistRepository,
	playlistRepo domain.PlaylistRepository,
) *AdminHandler {
	return &AdminHandler{
		usecase:      usecase,
		songRepo:     songRepo,
		userRepo:     userRepo,
		animeRepo:    animeRepo,
		artistRepo:   artistRepo,
		playlistRepo: playlistRepo,
	}
}

func (h *AdminHandler) paginatedResponse(c *fiber.Ctx, items interface{}, total int, page, perPage int) fiber.Map {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
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
			"last_page":    totalPages,
			"has_more":     page < totalPages,
		},
		"links": fiber.Map{
			"self": buildURL(page),
		},
	}

	if page < totalPages {
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

func (h *AdminHandler) getAuditMetadata(c *fiber.Ctx) domain.AuditMetadata {
	actorID, _ := c.Locals("user_id").(uint64)
	roles, _ := c.Locals("user_roles").([]string)

	primaryRole := ""
	if len(roles) > 0 {
		primaryRole = roles[0]
		// Prioritize most powerful roles for permission checks
		for _, r := range roles {
			if r == "owner" || r == "admin" {
				primaryRole = r
				break
			}
		}
	}

	return domain.AuditMetadata{
		ActorID:   actorID,
		Role:      primaryRole,
		URL:       c.OriginalURL(),
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}
}

func (h *AdminHandler) resolveID(c *fiber.Ctx, entityType string) (uint64, error) {
	paramID := c.Params("id")
	if paramID == "" {
		return 0, domain.NewAppError(400, "Missing ID parameter", nil)
	}

	id, err := strconv.ParseUint(paramID, 10, 64)
	if err == nil {
		return id, nil
	}

	// Try UUID resolution
	switch entityType {
	case "user":
		user, err := h.userRepo.GetByUUID(c.Context(), paramID)
		if err == nil {
			return user.ID, nil
		}
	case "song":
		song, err := h.songRepo.GetByUUID(c.Context(), paramID)
		if err == nil {
			return song.ID, nil
		}
	case "anime":
		anime, err := h.animeRepo.GetByUUID(c.Context(), paramID)
		if err == nil {
			return anime.ID, nil
		}
	case "artist":
		artist, err := h.artistRepo.GetByUUID(c.Context(), paramID)
		if err == nil {
			return artist.ID, nil
		}
	case "playlist":
		playlist, err := h.playlistRepo.GetByUUID(c.Context(), paramID)
		if err == nil {
			return playlist.ID, nil
		}
	}

	return 0, domain.NewAppError(404, "Entity not found with provided ID or UUID", nil)
}

func (h *AdminHandler) GetDashboard(c *fiber.Ctx) error {
	stats, metrics, err := h.usecase.GetDashboardData(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"stats":   stats,
		"metrics": metrics,
	})
}

func (h *AdminHandler) SnapshotRankingPositions(c *fiber.Ctx) error {
	if err := h.usecase.SnapshotRanking(c.Context()); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Ranking snapshot created successfully"})
}

// USERS
func (h *AdminHandler) GetUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")

	users, total, err := h.usecase.GetUsers(c.Context(), page, limit, search)
	if err != nil {
		return err
	}

	return c.JSON(h.paginatedResponse(c, users, total, page, limit))
}

func (h *AdminHandler) GetRoles(c *fiber.Ctx) error {
	roles, err := h.usecase.GetRoles(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": roles})
}

func (h *AdminHandler) GetUser(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "user")
	if err != nil {
		return err
	}
	user, err := h.usecase.GetUser(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": user})
}

func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	var req struct {
		domain.User
		RoleIDs  []uint64 `json:"role_ids"`
		BadgeIDs []uint64 `json:"badge_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.CreateUser(c.Context(), &req.User, req.RoleIDs, req.BadgeIDs, h.getAuditMetadata(c)); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"data":    req.User,
	})
}

func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "user")
	if err != nil {
		return err
	}
	var req struct {
		domain.User
		RoleIDs  []uint64 `json:"role_ids"`
		BadgeIDs []uint64 `json:"badge_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	req.User.ID = id

	if err := h.usecase.UpdateUser(c.Context(), &req.User, req.RoleIDs, req.BadgeIDs, h.getAuditMetadata(c)); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User updated successfully",
		"data":    req.User,
	})
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "user")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteUser(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
		"data":    nil,
	})
}

func (h *AdminHandler) ResetPassword(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "user")
	if err != nil {
		return err
	}
	newPassword, err := h.usecase.ResetPassword(c.Context(), id, h.getAuditMetadata(c))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password reset successfully",
		"password": newPassword,
	})
}

// ANIMES
func (h *AdminHandler) GetAnimes(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")

	year := c.Query("year", "")
	if year == "" {
		year = c.Query("year_id", "")
	}
	season := c.Query("season", "")
	if season == "" {
		season = c.Query("season_id", "")
	}
	format := c.Query("format", "")
	if format == "" {
		format = c.Query("format_id", "")
	}
	if format == "" {
		format = c.Query("type", "")
	}

	var status *bool
	if statusStr := c.Query("status"); statusStr != "" {
		if statusStr == "true" || statusStr == "1" {
			v := true
			status = &v
		} else if statusStr == "false" || statusStr == "0" || statusStr == "pending" {
			v := false
			status = &v
		}
	}

	genre := c.Query("genre", "")

	animes, total, err := h.usecase.GetAnimes(c.Context(), page, limit, search, year, season, format, genre, status)
	if err != nil {
		return err
	}
	h.usecase.ResolveAnimesURLs(animes)

	return c.JSON(h.paginatedResponse(c, animes, total, page, limit))
}

func (h *AdminHandler) GetAnime(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "anime")
	if err != nil {
		return err
	}
	anime, err := h.usecase.GetAnime(c.Context(), id)
	if err != nil {
		return err
	}
	h.usecase.ResolveAnimeURLs(anime)
	return c.JSON(fiber.Map{"data": anime})
}

func (h *AdminHandler) ToggleAnimeStatus(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "anime")
	if err != nil {
		return err
	}
	if err := h.usecase.ToggleAnimeStatus(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) CreateAnime(c *fiber.Ctx) error {
	var req struct {
		domain.Anime
		SeasonIDRaw interface{} `json:"season_id"`
		YearIDRaw   interface{} `json:"year_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if req.SeasonIDRaw != nil {
		req.Anime.SeasonID = h.resolveUint64(req.SeasonIDRaw)
	}
	if req.YearIDRaw != nil {
		req.Anime.YearID = h.resolveUint64(req.YearIDRaw)
	}

	h.usecase.HandleAnimeImages(c, &req.Anime)

	if err := h.usecase.CreateAnime(c.Context(), &req.Anime, h.getAuditMetadata(c)); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{"data": req.Anime})
}

func (h *AdminHandler) UpdateAnime(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "anime")
	if err != nil {
		return err
	}
	var req struct {
		domain.Anime
		SeasonIDRaw interface{} `json:"season_id"`
		YearIDRaw   interface{} `json:"year_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	req.Anime.ID = id

	if req.SeasonIDRaw != nil {
		req.Anime.SeasonID = h.resolveUint64(req.SeasonIDRaw)
	}
	if req.YearIDRaw != nil {
		req.Anime.YearID = h.resolveUint64(req.YearIDRaw)
	}

	h.usecase.HandleAnimeImages(c, &req.Anime)

	if err := h.usecase.UpdateAnime(c.Context(), &req.Anime, h.getAuditMetadata(c)); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": req.Anime})
}

func (h *AdminHandler) DeleteAnime(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "anime")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteAnime(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) BatchDeleteAnimes(c *fiber.Ctx) error {
	var req struct {
		IDs []uint64 `json:"ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid IDs", nil)
	}
	if err := h.usecase.BatchDeleteAnimes(c.Context(), req.IDs, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) SearchAnilist(c *fiber.Ctx) error {
	// Frontend (Svelte admin) sends ?q=... ; keep ?query= for compatibility
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		query = strings.TrimSpace(c.Query("query"))
	}
	// Empty = all formats (matches admin UI "All Formats"). Do not default to TV.
	format := strings.TrimSpace(c.Query("format"))
	results, err := h.usecase.SearchAnilistAnimes(c.Context(), query, format)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": results})
}

func (h *AdminHandler) CreateAnimeFromAnilist(c *fiber.Ctx) error {
	var req struct {
		AnilistID int `json:"anilist_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid Anilist ID", nil)
	}

	anime, err := h.usecase.CreateAnimeFromAnilist(c.Context(), req.AnilistID, h.getAuditMetadata(c))
	if err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{"data": anime})
}

func (h *AdminHandler) BatchCreateAnimesFromAnilist(c *fiber.Ctx) error {
	var req struct {
		AnilistIDs []int `json:"anilist_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid Anilist IDs", nil)
	}
	if len(req.AnilistIDs) == 0 {
		return domain.NewAppError(400, "No Anilist IDs provided", nil)
	}

	result := h.usecase.BatchCreateAnimesFromAnilist(c.Context(), req.AnilistIDs, h.getAuditMetadata(c))

	// Always 200 with structured body so the client can show partial success / per-ID errors.
	ok := result.Failed == 0 && result.Imported > 0
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": ok,
		"data":    result,
	})
}

func (h *AdminHandler) BatchFetchAnimes(c *fiber.Ctx) error {
	var req struct {
		Season string `json:"season"`
		Year   int    `json:"year"`
		Format string `json:"format"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if req.Format == "" {
		req.Format = "TV"
	}

	if err := h.usecase.BatchFetchAnimes(c.Context(), req.Season, req.Year, req.Format, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Batch fetch started"})
}



func (h *AdminHandler) SyncAnime(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "anime")
	if err != nil {
		return err
	}
	if err := h.usecase.SyncAnime(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Anime synced successfully"})
}

// SONGS
func (h *AdminHandler) GetSongs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	animeIDStr := c.Query("anime", "")
	if animeIDStr == "" {
		animeIDStr = c.Query("anime_id", "")
	}

	var animeID *uint64
	if animeIDStr != "" {
		id, err := strconv.ParseUint(animeIDStr, 10, 64)
		if err == nil {
			animeID = &id
		}
	}

	statusStr := c.Query("status", "")
	var status *bool
	if statusStr != "" {
		var val bool
		if statusStr == "true" || statusStr == "1" {
			val = true
			status = &val
		} else if statusStr == "false" || statusStr == "0" || statusStr == "pending" {
			val = false
			status = &val
		}
	}

	songs, total, err := h.usecase.GetSongs(c.Context(), page, limit, search, animeID, status)
	if err != nil {
		return err
	}
	h.usecase.ResolveSongsURLs(songs)

	return c.JSON(h.paginatedResponse(c, songs, total, page, limit))
}

func (h *AdminHandler) GetLatestSongNumber(c *fiber.Ctx) error {
	animeIDStr := c.Query("anime", "")
	if animeIDStr == "" {
		animeIDStr = c.Query("anime_id", "")
	}
	animeID, _ := strconv.ParseUint(animeIDStr, 10, 64)
	songType := c.Query("type")
	typeIDStr := c.Query("type_id")

	if typeIDStr != "" && typeIDStr != "0" {
		typeID, err := strconv.ParseUint(typeIDStr, 10, 64)
		if err == nil {
			// Find slug for this typeID to keep the usecase clean (it expects slug)
			types, _ := h.songRepo.GetSongTypes(c.Context())
			for _, t := range types {
				if t.ID != nil && *t.ID == typeID {
					if t.Slug != nil {
						songType = *t.Slug
					}
					break
				}
			}
		}
	}

	number, err := h.usecase.GetNextSongNumber(c.Context(), animeID, songType)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"number": number})
}

func (h *AdminHandler) GetSong(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "song")
	if err != nil {
		return err
	}
	song, err := h.usecase.GetSong(c.Context(), id)
	if err != nil {
		return err
	}
	h.usecase.ResolveSongURLs(song)
	return c.JSON(fiber.Map{"data": song})
}

func (h *AdminHandler) CreateSong(c *fiber.Ctx) error {
	var req struct {
		domain.Song
		TypeIDRaw   interface{} `json:"type_id"`
		SeasonIDRaw interface{} `json:"season_id"`
		YearIDRaw   interface{} `json:"year_id"`
		ArtistIDs   []uint64    `json:"artist_ids"`
		ArtistsStr  string      `json:"artists_string"`
	}

	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	// Resolve TypeID if provided as string/UUID
	if req.TypeIDRaw != nil {
		typeIDStr := fmt.Sprintf("%v", req.TypeIDRaw)
		resolvedID, err := h.resolveTypeID(c.Context(), typeIDStr)
		if err == nil {
			req.Song.TypeID = resolvedID
		}
	}

	// Resolve SeasonID and YearID (can be strings from configState)
	if req.SeasonIDRaw != nil {
		req.Song.SeasonID = h.resolveUint64(req.SeasonIDRaw)
	}
	if req.YearIDRaw != nil {
		req.Song.YearID = h.resolveUint64(req.YearIDRaw)
	}

	if err := h.usecase.CreateSong(c.Context(), &req.Song, h.getAuditMetadata(c)); err != nil {
		return err
	}

	// Sync Artists
	if req.ArtistsStr != "" {
		_ = h.usecase.SyncArtistsFromString(c.Context(), req.Song.ID, req.ArtistsStr, h.getAuditMetadata(c))
	} else {
		_ = h.usecase.SyncSongArtists(c.Context(), req.Song.ID, req.ArtistIDs)
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Song created successfully",
		"data":    req.Song,
	})
}

func (h *AdminHandler) UpdateSong(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "song")
	if err != nil {
		return err
	}
	var req struct {
		domain.Song
		TypeIDRaw   interface{} `json:"type_id"`
		SeasonIDRaw interface{} `json:"season_id"`
		YearIDRaw   interface{} `json:"year_id"`
		ArtistIDs   []uint64    `json:"artist_ids"`
		ArtistsStr  string      `json:"artists_string"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	req.Song.ID = id

	// Resolve TypeID if provided as string/UUID
	if req.TypeIDRaw != nil {
		typeIDStr := fmt.Sprintf("%v", req.TypeIDRaw)
		resolvedID, err := h.resolveTypeID(c.Context(), typeIDStr)
		if err == nil {
			req.Song.TypeID = resolvedID
		}
	}

	// Resolve SeasonID and YearID (can be strings from configState)
	if req.SeasonIDRaw != nil {
		req.Song.SeasonID = h.resolveUint64(req.SeasonIDRaw)
	}
	if req.YearIDRaw != nil {
		req.Song.YearID = h.resolveUint64(req.YearIDRaw)
	}

	if err := h.usecase.UpdateSong(c.Context(), &req.Song, h.getAuditMetadata(c)); err != nil {
		return err
	}

	// Sync Artists
	if req.ArtistsStr != "" {
		_ = h.usecase.SyncArtistsFromString(c.Context(), req.Song.ID, req.ArtistsStr, h.getAuditMetadata(c))
	} else {
		_ = h.usecase.SyncSongArtists(c.Context(), req.Song.ID, req.ArtistIDs)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Song updated successfully",
		"data":    req.Song,
	})
}

func (h *AdminHandler) resolveTypeID(ctx context.Context, uuidOrID string) (*uint64, error) {
	if uuidOrID == "" || uuidOrID == "0" {
		return nil, nil
	}

	// Try to parse as uint64 first
	if id, err := strconv.ParseUint(uuidOrID, 10, 64); err == nil {
		return &id, nil
	}

	// Try to resolve as UUID
	types, err := h.songRepo.GetSongTypes(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range types {
		if t.UUID != nil && *t.UUID == uuidOrID {
			return t.ID, nil
		}
	}

	return nil, fmt.Errorf("song type not found: %s", uuidOrID)
}

func (h *AdminHandler) resolveUint64(val interface{}) uint64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return uint64(v)
	case string:
		if v == "" {
			return 0
		}
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			return id
		}
	case int:
		return uint64(v)
	case uint64:
		return v
	}
	return 0
}

func (h *AdminHandler) DeleteSong(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "song")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteSong(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Song deleted successfully",
	})
}

func (h *AdminHandler) ToggleSongStatus(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "song")
	if err != nil {
		return err
	}
	if err := h.usecase.ToggleSongStatus(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Status updated successfully",
	})
}

// SONG VARIANTS
func (h *AdminHandler) GetVariant(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	v, err := h.usecase.GetVariant(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": v})
}

func (h *AdminHandler) GetVariants(c *fiber.Ctx) error {
	return h.getVariantsInternal(c)
}

func (h *AdminHandler) GetVideos(c *fiber.Ctx) error {
	return h.getVariantsInternal(c)
}

func (h *AdminHandler) getVariantsInternal(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")

	var animeID *uint64
	var status *bool

	if aIDStr := c.Query("anime"); aIDStr != "" {
		if id, err := strconv.ParseUint(aIDStr, 10, 64); err == nil && id > 0 {
			animeID = &id
		}
	} else if aIDStr := c.Query("anime_id"); aIDStr != "" {
		if id, err := strconv.ParseUint(aIDStr, 10, 64); err == nil && id > 0 {
			animeID = &id
		}
	}

	if sStr := c.Query("status"); sStr != "" {
		if sStr == "true" || sStr == "1" {
			v := true
			status = &v
		} else if sStr == "false" || sStr == "0" || sStr == "pending" {
			v := false
			status = &v
		}
	}

	variants, total, err := h.usecase.GetVariants(c.Context(), page, limit, search, animeID, status)
	if err != nil {
		return err
	}

	return c.JSON(h.paginatedResponse(c, variants, total, page, limit))
}

func (h *AdminHandler) CreateVariant(c *fiber.Ctx) error {
	var req struct {
		domain.SongVariant
		SeasonIDRaw interface{} `json:"season_id"`
		YearIDRaw   interface{} `json:"year_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if req.SeasonIDRaw != nil {
		req.SongVariant.SeasonID = h.resolveUint64(req.SeasonIDRaw)
	}
	if req.YearIDRaw != nil {
		req.SongVariant.YearID = h.resolveUint64(req.YearIDRaw)
	}

	if err := h.usecase.CreateVariant(c.Context(), &req.SongVariant, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": req.SongVariant})
}

func (h *AdminHandler) UpdateVariant(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var req struct {
		domain.SongVariant
		SeasonIDRaw interface{} `json:"season_id"`
		YearIDRaw   interface{} `json:"year_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	req.SongVariant.ID = id

	if req.SeasonIDRaw != nil {
		req.SongVariant.SeasonID = h.resolveUint64(req.SeasonIDRaw)
	}
	if req.YearIDRaw != nil {
		req.SongVariant.YearID = h.resolveUint64(req.YearIDRaw)
	}

	if err := h.usecase.UpdateVariant(c.Context(), &req.SongVariant, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": req.SongVariant})
}

func (h *AdminHandler) UpdateVariantVideo(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	v, err := h.usecase.GetVariant(c.Context(), id)
	if err != nil {
		return err
	}

	h.usecase.HandleVariantVideo(c, v)
	if err := h.usecase.UpdateVariant(c.Context(), v, h.getAuditMetadata(c)); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": v})
}

func (h *AdminHandler) DeleteVariant(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.DeleteVariant(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) ToggleVariantStatus(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.ToggleVariantStatus(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) ToggleVariantSpoiler(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.ToggleVariantSpoiler(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) ToggleVariantNSFW(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.ToggleVariantNSFW(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}


// ARTISTS
func (h *AdminHandler) GetArtists(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")

	artists, total, err := h.usecase.GetArtists(c.Context(), page, limit, search)
	if err != nil {
		return err
	}
	h.usecase.ResolveArtistsURLs(artists)

	return c.JSON(h.paginatedResponse(c, artists, total, page, limit))
}

func (h *AdminHandler) GetArtist(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "artist")
	if err != nil {
		return err
	}
	artist, err := h.usecase.GetArtist(c.Context(), id)
	if err != nil {
		return err
	}
	h.usecase.ResolveArtistURLs(artist)
	return c.JSON(fiber.Map{"data": artist})
}

func (h *AdminHandler) ToggleArtistStatus(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "artist")
	if err != nil {
		return err
	}
	if err := h.usecase.ToggleArtistStatus(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) CreateArtist(c *fiber.Ctx) error {
	var a domain.Artist
	if err := c.BodyParser(&a); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	file, err := c.FormFile("avatar")
	if err == nil {
		f, _ := file.Open()
		defer f.Close()
		// Handled inside Usecase via direct call or manual extraction
		// For simplicity in this handler, we trust the usecase later
	}

	if err := h.usecase.CreateArtist(c.Context(), &a, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": a})
}

func (h *AdminHandler) GenerateArtistAvatar(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "artist")
	if err != nil {
		return err
	}
	if err := h.usecase.GenerateArtistAvatar(c.Context(), id, true); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Avatar generated!"})
}

func (h *AdminHandler) SyncArtistAvatar(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "artist")
	if err != nil {
		return err
	}
	artist, err := h.usecase.SyncArtistAvatar(c.Context(), id, h.getAuditMetadata(c))
	if err != nil {
		return err
	}
	h.usecase.ResolveArtistURLs(artist)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Avatar synced successfully",
		"data":    artist,
	})
}

func (h *AdminHandler) BatchGenerateArtistAvatars(c *fiber.Ctx) error {
	var body struct {
		ArtistIDs []uint64 `json:"artist_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		// Ignore body parser errors if no body is provided (all artists)
		body.ArtistIDs = []uint64{}
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	progress := make(chan string)

	go func() {
		defer close(progress)
		// Use Background() instead of c.Context() because Fiber context is recycled
		_ = h.usecase.BatchGenerateArtistAvatars(context.Background(), body.ArtistIDs, progress)
	}()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for msg := range progress {
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.Flush()
		}
	})

	return nil
}

func (h *AdminHandler) MergeArtists(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	progress := make(chan string)

	go func() {
		defer close(progress)
		_ = h.usecase.MergeDuplicateArtists(context.Background(), progress)
	}()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for msg := range progress {
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.Flush()
		}
	})

	return nil
}

func (h *AdminHandler) RecountArtistStats(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	progress := make(chan string)

	go func() {
		defer close(progress)
		_ = h.usecase.RecountArtistStats(context.Background(), nil, progress)
	}()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for msg := range progress {
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.Flush()
		}
	})

	return nil
}

func (h *AdminHandler) UpdateArtist(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "artist")
	if err != nil {
		return err
	}
	var a domain.Artist
	if err := c.BodyParser(&a); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	a.ID = id

	// Handle File if any
	file, err := c.FormFile("avatar")
	if err == nil {
		f, _ := file.Open()
		defer f.Close()
		_ = h.usecase.UploadArtistAvatar(c.Context(), a.ID, f, file.Size, file.Header.Get("Content-Type"))
	}

	if err := h.usecase.UpdateArtist(c.Context(), &a, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": a})
}

func (h *AdminHandler) DeleteArtist(c *fiber.Ctx) error {
	id, err := h.resolveID(c, "artist")
	if err != nil {
		return err
	}
	if err := h.usecase.DeleteArtist(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

// TAXONOMIES
func (h *AdminHandler) GetYears(c *fiber.Ctx) error {
	years, err := h.usecase.GetYears(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": years})
}

func (h *AdminHandler) GetSongTypes(c *fiber.Ctx) error {
	types, err := h.usecase.GetSongTypes(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": types})
}

func (h *AdminHandler) GetInitData(c *fiber.Ctx) error {
	ctx := c.Context()

	years, _ := h.usecase.GetYears(ctx)
	seasons, _ := h.usecase.GetSeasons(ctx)
	formats, _ := h.usecase.GetFormats(ctx)
	genres, _ := h.usecase.SearchGenres(ctx, "", 0)
	songTypes, _ := h.usecase.GetSongTypes(ctx)

	mapTaxonomy := func(items interface{}) []fiber.Map {
		v := reflect.ValueOf(items)
		if v.Kind() != reflect.Slice {
			return []fiber.Map{}
		}
		res := make([]fiber.Map, v.Len())
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}
			
			// Helper to get value from field or pointer to field
			getVal := func(fieldName string) string {
				f := item.FieldByName(fieldName)
				if !f.IsValid() {
					return ""
				}
				if f.Kind() == reflect.Ptr {
					if f.IsNil() {
						return ""
					}
					f = f.Elem()
				}
				
				if f.Kind() == reflect.Uint64 {
					return strconv.FormatUint(f.Uint(), 10)
				}
				return f.String()
			}

			res[i] = fiber.Map{
				"id":   getVal("ID"),
				"name": getVal("Name"),
				"slug": getVal("Slug"),
			}
		}
		return res
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"years":      mapTaxonomy(years),
			"seasons":    mapTaxonomy(seasons),
			"formats":    mapTaxonomy(formats),
			"genres":     mapTaxonomy(genres),
			"song_types": mapTaxonomy(songTypes),
		},
	})
}

func (h *AdminHandler) CreateYear(c *fiber.Ctx) error {
	var y domain.Year
	if err := c.BodyParser(&y); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	if err := h.usecase.CreateYear(c.Context(), &y, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": y})
}

func (h *AdminHandler) UpdateYear(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var y domain.Year
	if err := c.BodyParser(&y); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	y.ID = id
	if err := h.usecase.UpdateYear(c.Context(), &y, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": y})
}

func (h *AdminHandler) ToggleYearCurrent(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.ToggleYearCurrent(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) DeleteYear(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.DeleteYear(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) GetSeasons(c *fiber.Ctx) error {
	seasons, err := h.usecase.GetSeasons(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": seasons})
}

func (h *AdminHandler) CreateSeason(c *fiber.Ctx) error {
	var s domain.Season
	if err := c.BodyParser(&s); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	if err := h.usecase.CreateSeason(c.Context(), &s, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": s})
}

func (h *AdminHandler) UpdateSeason(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var s domain.Season
	if err := c.BodyParser(&s); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	s.ID = id
	if err := h.usecase.UpdateSeason(c.Context(), &s, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": s})
}

func (h *AdminHandler) ToggleSeasonCurrent(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.ToggleSeasonCurrent(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) DeleteSeason(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.DeleteSeason(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) CreateFormat(c *fiber.Ctx) error {
	var f domain.Format
	if err := c.BodyParser(&f); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	if err := h.usecase.CreateFormat(c.Context(), &f, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": f})
}

func (h *AdminHandler) UpdateFormat(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var f domain.Format
	if err := c.BodyParser(&f); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	f.ID = id
	if err := h.usecase.UpdateFormat(c.Context(), &f, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": f})
}

func (h *AdminHandler) DeleteFormat(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.DeleteFormat(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) GetFormats(c *fiber.Ctx) error {
	formats, err := h.usecase.GetFormats(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": formats})
}

func (h *AdminHandler) GetGenres(c *fiber.Ctx) error {
	genres, err := h.usecase.SearchGenres(c.Context(), "", 0)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": genres})
}

func (h *AdminHandler) CreateGenre(c *fiber.Ctx) error {
	var g domain.Genre
	if err := c.BodyParser(&g); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	if err := h.usecase.CreateGenre(c.Context(), &g, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": g})
}

func (h *AdminHandler) UpdateGenre(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var g domain.Genre
	if err := c.BodyParser(&g); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	g.ID = id
	if err := h.usecase.UpdateGenre(c.Context(), &g, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": g})
}

func (h *AdminHandler) DeleteGenre(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.DeleteGenre(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *AdminHandler) GetStudios(c *fiber.Ctx) error {
	search := c.Query("search", "")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	studios, err := h.usecase.SearchStudios(c.Context(), search, limit)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": studios})
}

func (h *AdminHandler) GetProducers(c *fiber.Ctx) error {
	search := c.Query("search", "")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	producers, err := h.usecase.SearchProducers(c.Context(), search, limit)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": producers})
}

// AUDIT LOGS
func (h *AdminHandler) GetAuditLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	filters := make(map[string]interface{})
	// Repository expects: user_id, event, auditable_type, auditable_id
	if userID := c.Query("user"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = id
		}
	} else if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = id
		}
	} else if actorID := c.Query("actor_id"); actorID != "" {
		if id, err := strconv.ParseUint(actorID, 10, 64); err == nil {
			filters["user_id"] = id
		}
	}

	if event := c.Query("event"); event != "" {
		filters["event"] = event
	} else if action := c.Query("action"); action != "" {
		filters["event"] = action
	}

	if resource := c.Query("resource"); resource != "" {
		filters["auditable_type"] = resource
	} else if auditableType := c.Query("auditable_type"); auditableType != "" {
		filters["auditable_type"] = auditableType
	} else if resourceType := c.Query("resource_type"); resourceType != "" {
		filters["auditable_type"] = resourceType
	}

	if resourceID := c.Query("resource_id"); resourceID != "" {
		if id, err := strconv.ParseUint(resourceID, 10, 64); err == nil {
			filters["auditable_id"] = id
		}
	} else if auditableID := c.Query("auditable_id"); auditableID != "" {
		if id, err := strconv.ParseUint(auditableID, 10, 64); err == nil {
			filters["auditable_id"] = id
		}
	}

	logs, total, err := h.usecase.GetAuditLogs(c.Context(), page, limit, filters)
	if err != nil {
		return err
	}

	return c.JSON(h.paginatedResponse(c, logs, total, page, limit))
}

func (h *AdminHandler) GetAuditLog(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	log, err := h.usecase.GetAuditLog(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": log})
}

// XP ACTIVITIES
func (h *AdminHandler) GetXPActivities(c *fiber.Ctx) error {
	activities, err := h.usecase.GetAllXPActivities(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": activities})
}

func (h *AdminHandler) UpdateXPActivity(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	var activity domain.XPActivity
	if err := c.BodyParser(&activity); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	activity.ID = id
	if err := h.usecase.UpdateXPActivity(c.Context(), &activity); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "XP Activity updated", "data": activity})
}

// GetApiStatus returns the health status of external APIs
func (h *AdminHandler) GetApiStatus(c *fiber.Ctx) error {
	anilistOnline, anilistMsg := h.usecase.CheckAnilistStatus(c.Context())
	animeThemesOnline, animeThemesMsg := h.usecase.CheckAnimeThemesStatus(c.Context())

	return c.JSON(fiber.Map{
		"anilist": fiber.Map{
			"status":  map[bool]string{true: "online", false: "offline"}[anilistOnline],
			"message": anilistMsg,
		},
		"animethemes": fiber.Map{
			"status":  map[bool]string{true: "online", false: "offline"}[animeThemesOnline],
			"message": animeThemesMsg,
		},
	})
}

package v1

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/playlist"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PlaylistHandler struct {
	usecase *playlist.PlaylistUsecase
}

func NewPlaylistHandler(usecase *playlist.PlaylistUsecase) *PlaylistHandler {
	return &PlaylistHandler{usecase: usecase}
}

func (h *PlaylistHandler) getUserID(c *fiber.Ctx) *uint64 {
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

// GetUserPlaylists retrieves a user's public playlists, or all if requesting self.
// @Summary List User Playlists
// @Description Fetch paginated playlists for a user.
// @Tags Playlists
// @Produce json
// @Param id path int true "Target User ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Playlist
// @Failure 400 {object} domain.AppError
// @Router /playlists/users/{id} [get]
func (h *PlaylistHandler) GetUserPlaylists(c *fiber.Ctx) error {
	targetUserIDParam := c.Params("id")
	targetUserID, err := strconv.ParseUint(targetUserIDParam, 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid user ID", err)
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	requestingUserID := h.getUserID(c)

	playlists, err := h.usecase.GetUserPlaylists(c.Context(), requestingUserID, targetUserID, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": playlists})
}

func (h *PlaylistHandler) GetMyPlaylists(c *fiber.Ctx) error {
	uid := h.getUserID(c)
	if uid == nil {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	userID := *uid

	songID, _ := strconv.ParseUint(c.Query("song_id", "0"), 10, 64)
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	playlists, err := h.usecase.GetMyPlaylists(c.Context(), userID, songID, limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"playlists": playlists})
}

// GetPlaylist retrieves a specific playlist and its songs.
// @Summary Get Playlist
// @Description Fetch details of a playlist including attached songs.
// @Tags Playlists
// @Produce json
// @Param id path int true "Playlist ID"
// @Success 200 {object} object{data=domain.Playlist}
// @Failure 400 {object} domain.AppError
// @Failure 403 {object} domain.AppError
// @Failure 404 {object} domain.AppError
// @Router /playlists/{id} [get]
func (h *PlaylistHandler) GetPlaylist(c *fiber.Ctx) error {
	playlistIDParam := c.Params("id")
	playlistID, err := strconv.ParseUint(playlistIDParam, 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid playlist ID", err)
	}

	requestingUserID := h.getUserID(c)

	playlist, err := h.usecase.GetPlaylist(c.Context(), playlistID, requestingUserID)
	if err != nil {
		return err
	}

	playlistSongs, err := h.usecase.GetPlaylistSongs(c.Context(), playlistID, requestingUserID)
	if err == nil {
		// Flatten PlaylistSong wrappers to Song objects for the frontend
		var songs []domain.Song
		for _, ps := range playlistSongs {
			if ps.Song != nil {
				songs = append(songs, *ps.Song)
			}
		}
		if songs == nil {
			songs = []domain.Song{}
		}
		playlist.Songs = nil // clear PlaylistSong slice

		// Return playlist with songs as flat array
		return c.JSON(fiber.Map{
			"playlist": fiber.Map{
				"id":          playlist.ID,
				"name":        playlist.Name,
				"description": playlist.Description,
				"user_id":     playlist.UserID,
				"is_public":   playlist.IsPublic,
				"created_at":  playlist.CreatedAt,
				"updated_at":  playlist.UpdatedAt,
				"user":        playlist.User,
				"songs":       songs,
			},
		})
	}

	return c.JSON(fiber.Map{"playlist": playlist})
}

// POST /api/playlists
type createPlaylistReq struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsPublic    bool    `json:"is_public"`
}

// Create a new playlist
// @Summary Create Playlist
// @Description Create a new playlist for the authenticated user.
// @Tags Playlists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{name=string,description=string,is_public=boolean} true "Playlist Data"
// @Success 201 {object} object{data=domain.Playlist}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /playlists [post]
func (h *PlaylistHandler) Create(c *fiber.Ctx) error {
	uid := h.getUserID(c)
	if uid == nil {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	userID := *uid

	var req createPlaylistReq
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request payload", err)
	}

	playlist, err := h.usecase.CreatePlaylist(c.Context(), userID, req.Name, req.Description, req.IsPublic)
	if err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{"data": playlist})
}

// Update an existing playlist
// @Summary Update Playlist
// @Description Update an existing playlist by ID.
// @Tags Playlists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Playlist ID"
// @Param request body object{name=string,description=string,is_public=boolean} true "Playlist Data"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Failure 403 {object} domain.AppError
// @Failure 404 {object} domain.AppError
// @Router /playlists/{id} [put]
func (h *PlaylistHandler) Update(c *fiber.Ctx) error {
	uid := h.getUserID(c)
	if uid == nil {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	userID := *uid

	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	var req createPlaylistReq
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.UpdatePlaylist(c.Context(), userID, playlistID, req.Name, req.Description, req.IsPublic); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Playlist updated successfully"})
}

// Delete a playlist
// @Summary Delete Playlist
// @Description Delete an existing playlist by ID.
// @Tags Playlists
// @Security BearerAuth
// @Produce json
// @Param id path int true "Playlist ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Failure 403 {object} domain.AppError
// @Failure 404 {object} domain.AppError
// @Router /playlists/{id} [delete]
func (h *PlaylistHandler) Delete(c *fiber.Ctx) error {
	uid := h.getUserID(c)
	if uid == nil {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	userID := *uid

	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	if err := h.usecase.DeletePlaylist(c.Context(), userID, playlistID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Playlist deleted successfully"})
}

// POST /api/playlists/:id/songs
type addSongReq struct {
	SongID   uint64 `json:"song_id"`
	Position int    `json:"position"`
}

// AddSong adds a song to a playlist
// @Summary Add Song to Playlist
// @Description Append or insert a song at a given position within a playlist.
// @Tags Playlists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Playlist ID"
// @Param request body object{song_id=int,position=int} true "Add Song Data"
// @Success 201 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /playlists/{id}/songs [post]
func (h *PlaylistHandler) AddSong(c *fiber.Ctx) error {
	uid := h.getUserID(c)
	if uid == nil {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	userID := *uid

	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", nil)
	}

	var req addSongReq
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.AddSongToPlaylist(c.Context(), userID, playlistID, req.SongID, req.Position); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{"message": "Song added successfully"})
}

// RemoveSong removes a song from a playlist
// @Summary Remove Song from Playlist
// @Description Remove a specific song from a user's playlist.
// @Tags Playlists
// @Security BearerAuth
// @Produce json
// @Param id path int true "Playlist ID"
// @Param songID path int true "Song ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /playlists/{id}/songs/{songID} [delete]
func (h *PlaylistHandler) RemoveSong(c *fiber.Ctx) error {
	uid := h.getUserID(c)
	if uid == nil {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	userID := *uid

	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid playlist ID", nil)
	}

	songID, err := strconv.ParseUint(c.Params("songID"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid song ID", nil)
	}

	if err := h.usecase.RemoveSongFromPlaylist(c.Context(), userID, playlistID, songID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Song removed successfully"})
}

// PUT /api/playlists/:id/songs/reorder
type reorderReq struct {
	Positions map[uint64]int `json:"positions"` // maps song_id -> position
}

// ReorderSongs updates the positions of songs in a playlist
// @Summary Reorder Playlist Songs
// @Description Update the sort order of songs within a custom playlist.
// @Tags Playlists
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Playlist ID"
// @Param request body object{positions=object} true "Reorder Data (song_id -> position)"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /playlists/{id}/songs/reorder [put]
func (h *PlaylistHandler) ReorderSongs(c *fiber.Ctx) error {
	uid := h.getUserID(c)
	if uid == nil {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	userID := *uid

	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid playlist ID", nil)
	}

	var req reorderReq
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.ReorderPlaylistSongs(c.Context(), userID, playlistID, req.Positions); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Songs reordered successfully"})
}

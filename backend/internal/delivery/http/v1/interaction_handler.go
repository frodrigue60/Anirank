package v1

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	"anirank/api/internal/usecase/interaction"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type InteractionHandler struct {
	usecase         *interaction.InteractionUsecase
	activityUsecase domain.ActivityUsecase
	songRepo        domain.SongRepository
	userRepo        domain.UserRepository
	animeRepo       domain.AnimeRepository
	artistRepo      domain.ArtistRepository
	commentRepo     domain.CommentRepository
}

func (h *InteractionHandler) paginatedResponse(c *fiber.Ctx, items interface{}, total int, page, perPage int) fiber.Map {
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

func NewInteractionHandler(
	usecase *interaction.InteractionUsecase,
	activityUsecase domain.ActivityUsecase,
	songRepo domain.SongRepository,
	userRepo domain.UserRepository,
	animeRepo domain.AnimeRepository,
	artistRepo domain.ArtistRepository,
	commentRepo domain.CommentRepository,
) *InteractionHandler {
	return &InteractionHandler{
		usecase:         usecase,
		activityUsecase: activityUsecase,
		songRepo:        songRepo,
		userRepo:        userRepo,
		animeRepo:       animeRepo,
		artistRepo:      artistRepo,
		commentRepo:     commentRepo,
	}
}

// POST /api/interactions/ratings  (Song-only)
type rateRequest struct {
	SongID string  `json:"song_id"`
	Score  float64 `json:"score"`
}

// Rate allows a user to rate a Song
func (h *InteractionHandler) Rate(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	var req rateRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request payload", err)
	}

	// Resolve Song ID
	songID, err := strconv.ParseUint(req.SongID, 10, 64)
	if err != nil {
		// Try UUID resolution
		song, err := h.songRepo.GetByUUID(c.Context(), req.SongID)
		if err != nil {
			return domain.NewAppError(404, "Song not found", err)
		}
		songID = song.ID
	}

	average, rating, err := h.usecase.RateSong(c.Context(), userID, songID, req.Score)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Rating submitted successfully",
			"average": average,
			"rating":  rating,
		},
	})
}

// POST /api/interactions/reactions
type reactRequest struct {
	EntityID   string      `json:"entity_id"`
	SongID     string      `json:"song_id"`
	EntityType string      `json:"entity_type"`
	Type       interface{} `json:"type"` // Can be string "like"/"dislike" or numeric 1/-1
}

// React allows a user to like/dislike an entity.
func (h *InteractionHandler) React(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	var req reactRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request payload", err)
	}

	// Determine reaction type
	var reactionType int8
	switch v := req.Type.(type) {
	case float64:
		reactionType = int8(v)
	case int:
		reactionType = int8(v)
	case string:
		if v == "like" || v == "1" {
			reactionType = 1
		} else if v == "dislike" || v == "-1" {
			reactionType = -1
		}
	}

	// Prefer SongID for song reactions
	if req.SongID != "" || req.EntityType == "song" {
		target := req.SongID
		if target == "" {
			target = req.EntityID
		}

		songID, err := strconv.ParseUint(target, 10, 64)
		if err != nil {
			song, err := h.songRepo.GetByUUID(c.Context(), target)
			if err != nil {
				return domain.NewAppError(404, "Song not found", err)
			}
			songID = song.ID
		}

		likes, dislikes, err := h.usecase.ReactToSong(c.Context(), userID, songID, reactionType)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"likesCount":    likes,
				"dislikesCount": dislikes,
			},
		})
	}

	// Handle Other Entities (Polymorphic)
	if req.EntityID == "" {
		return domain.NewAppError(400, "entity_id is required", nil)
	}

	entityID, err := strconv.ParseUint(req.EntityID, 10, 64)
	if err != nil {
		// Try resolving based on EntityType
		switch req.EntityType {
		case "comment", "App\\Models\\Comment":
			comment, err := h.commentRepo.GetByUUID(c.Context(), req.EntityID)
			if err != nil {
				return domain.NewAppError(404, "Comment not found", err)
			}
			entityID = comment.ID
		case "anime", "App\\Models\\Anime":
			anime, err := h.animeRepo.GetByUUID(c.Context(), req.EntityID)
			if err != nil {
				return domain.NewAppError(404, "Anime not found", err)
			}
			entityID = anime.ID
		default:
			return domain.NewAppError(400, "Invalid entity ID or unsupported type for UUID resolution", err)
		}
	}

	if req.EntityType == "comment" || req.EntityType == "App\\Models\\Comment" {
		likes, dislikes, err := h.usecase.ReactToComment(c.Context(), userID, entityID, reactionType)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"likesCount":    likes,
				"dislikesCount": dislikes,
			},
		})
	}

	err = h.usecase.ReactToEntity(c.Context(), userID, entityID, req.EntityType, reactionType)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Reaction updated",
		},
	})
}

// POST /api/interactions/favorites
type favoriteRequest struct {
	EntityID   string `json:"entity_id"`
	EntityType string `json:"entity_type"`
}

func (h *InteractionHandler) ToggleFavorite(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	var req favoriteRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid request payload", err)
	}

	// Resolve Entity ID
	entityID, err := strconv.ParseUint(req.EntityID, 10, 64)
	if err != nil {
		switch req.EntityType {
		case "song":
			song, err := h.songRepo.GetByUUID(c.Context(), req.EntityID)
			if err != nil {
				return domain.NewAppError(404, "Song not found", err)
			}
			entityID = song.ID
		case "anime":
			anime, err := h.animeRepo.GetByUUID(c.Context(), req.EntityID)
			if err != nil {
				return domain.NewAppError(404, "Anime not found", err)
			}
			entityID = anime.ID
		case "artist":
			artist, err := h.artistRepo.GetByUUID(c.Context(), req.EntityID)
			if err != nil {
				return domain.NewAppError(404, "Artist not found", err)
			}
			entityID = artist.ID
		default:
			return domain.NewAppError(400, "Unsupported entity type for UUID resolution", err)
		}
	}

	isFavorited, err := h.usecase.ToggleFavorite(c.Context(), userID, entityID, req.EntityType)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"favorited": isFavorited,
		},
	})
}

func (h *InteractionHandler) Feed(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	feed, err := h.usecase.GetGlobalActivity(c.Context(), limit)
	if err != nil {
		return err
	}

	dtoFeed := make([]dto.ActivityItemDTO, len(feed))
	for i, item := range feed {
		dtoFeed[i] = dto.ToActivityItemDTO(item)
	}

	return c.JSON(fiber.Map{"data": dtoFeed})
}

func (h *InteractionHandler) GetComments(c *fiber.Ctx) error {
	var requestingUserID *uint64
	if val := c.Locals("user_id"); val != nil {
		if id, ok := val.(uint64); ok {
			requestingUserID = &id
		} else if f, ok := val.(float64); ok {
			id := uint64(f)
			requestingUserID = &id
		}
	}

	entityID := uint64(c.QueryInt("entity_id", 0))
	entityType := c.Query("entity_type")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	page := (offset / limit) + 1

	if entityID == 0 || entityType == "" {
		return domain.NewAppError(400, "entity_id and entity_type are required", nil)
	}

	comments, total, err := h.usecase.GetCommentsByEntity(c.Context(), requestingUserID, entityID, entityType, limit, offset)
	if err != nil {
		return err
	}

	dtoComments := make([]dto.CommentDTO, len(comments))
	for i, c := range comments {
		dtoComments[i] = dto.ToCommentDTO(&c)
	}

	return c.JSON(h.paginatedResponse(c, dtoComments, total, page, limit))
}

func (h *InteractionHandler) GetSongComments(c *fiber.Ctx) error {
	var requestingUserID *uint64
	if val := c.Locals("user_id"); val != nil {
		if id, ok := val.(uint64); ok {
			requestingUserID = &id
		} else if f, ok := val.(float64); ok {
			id := uint64(f)
			requestingUserID = &id
		}
	}

	songIDParam := c.Params("uuid") // Router updated to :uuid
	songID, err := strconv.ParseUint(songIDParam, 10, 64)
	if err != nil {
		// Try UUID resolution
		song, err := h.songRepo.GetByUUID(c.Context(), songIDParam)
		if err != nil {
			return domain.NewAppError(404, "Song not found", err)
		}
		songID = song.ID
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	page := (offset / limit) + 1

	comments, total, err := h.usecase.GetCommentsByEntity(c.Context(), requestingUserID, songID, "song", limit, offset)
	if err != nil {
		return err
	}

	dtoComments := make([]dto.CommentDTO, len(comments))
	for i, c := range comments {
		dtoComments[i] = dto.ToCommentDTO(&c)
	}

	return c.JSON(h.paginatedResponse(c, dtoComments, total, page, limit))
}

func (h *InteractionHandler) GetReplies(c *fiber.Ctx) error {
	var requestingUserID *uint64
	if val := c.Locals("user_id"); val != nil {
		if id, ok := val.(uint64); ok {
			requestingUserID = &id
		} else if f, ok := val.(float64); ok {
			id := uint64(f)
			requestingUserID = &id
		}
	}

	parentIDParam := c.Params("id")
	parentID, err := strconv.ParseUint(parentIDParam, 10, 64)
	if err != nil {
		// Try UUID resolution
		comment, err := h.commentRepo.GetByUUID(c.Context(), parentIDParam)
		if err != nil {
			return domain.NewAppError(404, "Comment not found", err)
		}
		parentID = comment.ID
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	page := (offset / limit) + 1

	replies, total, err := h.usecase.GetCommentReplies(c.Context(), requestingUserID, parentID, limit, offset)
	if err != nil {
		return err
	}

	dtoReplies := make([]dto.CommentDTO, len(replies))
	for i, r := range replies {
		dtoReplies[i] = dto.ToCommentDTO(&r)
	}

	return c.JSON(h.paginatedResponse(c, dtoReplies, total, page, limit))
}

func (h *InteractionHandler) SongComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Invalid user context", nil)
	}

	type reqBody struct {
		EntityID   string  `json:"entity_id"`
		EntityType string  `json:"entity_type"`
		Content    string  `json:"content"`
		ParentID   *string `json:"parent_id"`
	}

	var req reqBody
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON payload", err)
	}

	// Resolve Entity ID
	entityID, err := strconv.ParseUint(req.EntityID, 10, 64)
	if err != nil {
		switch req.EntityType {
		case "song":
			song, err := h.songRepo.GetByUUID(c.Context(), req.EntityID)
			if err != nil {
				return domain.NewAppError(404, "Song not found", err)
			}
			entityID = song.ID
		default:
			return domain.NewAppError(400, "Invalid entity ID for comment", err)
		}
	}

	var parentID *uint64
	if req.ParentID != nil {
		pid, err := strconv.ParseUint(*req.ParentID, 10, 64)
		if err != nil {
			comment, err := h.commentRepo.GetByUUID(c.Context(), *req.ParentID)
			if err != nil {
				return domain.NewAppError(404, "Parent comment not found", err)
			}
			pid = comment.ID
		}
		parentID = &pid
	}

	comment, err := h.usecase.SongComment(c.Context(), userID, entityID, req.EntityType, req.Content, parentID)
	if err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": dto.ToCommentDTO(comment)})
}

func (h *InteractionHandler) UpdateComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	commentIDParam := c.Params("id")
	commentID, err := strconv.ParseUint(commentIDParam, 10, 64)
	if err != nil {
		comment, err := h.commentRepo.GetByUUID(c.Context(), commentIDParam)
		if err != nil {
			return domain.NewAppError(404, "Comment not found", err)
		}
		commentID = comment.ID
	}

	type reqBody struct {
		Content string `json:"content"`
	}
	var req reqBody
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON payload", err)
	}

	err = h.usecase.UpdateComment(c.Context(), commentID, userID, req.Content)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Comment updated successfully",
		},
	})
}

func (h *InteractionHandler) DeleteComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	commentIDParam := c.Params("id")
	commentID, err := strconv.ParseUint(commentIDParam, 10, 64)
	if err != nil {
		// Try UUID resolution
		comment, err := h.commentRepo.GetByUUID(c.Context(), commentIDParam)
		if err != nil {
			return domain.NewAppError(404, "Comment not found", err)
		}
		commentID = comment.ID
	}
	err = h.usecase.DeleteComment(c.Context(), commentID, userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Comment deleted successfully",
		},
	})
}

// ---- FOLLOWS ----

func (h *InteractionHandler) FollowUser(c *fiber.Ctx) error {
	followerID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	followedTarget := c.Params("id")
	if followedTarget == "" {
		return domain.NewAppError(400, "User identifier is required", nil)
	}

	if err := h.usecase.FollowUser(c.Context(), followerID, followedTarget); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "User followed",
		},
	})
}

func (h *InteractionHandler) UnfollowUser(c *fiber.Ctx) error {
	followerID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	followedTarget := c.Params("id")
	if followedTarget == "" {
		return domain.NewAppError(400, "User identifier is required", nil)
	}

	if err := h.usecase.UnfollowUser(c.Context(), followerID, followedTarget); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "User unfollowed",
		},
	})
}

func (h *InteractionHandler) IsFollowing(c *fiber.Ctx) error {
	followerID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return c.JSON(fiber.Map{"is_following": false})
	}

	followedTarget := c.Params("id")
	if followedTarget == "" {
		return c.JSON(fiber.Map{"is_following": false})
	}

	isFollowing, err := h.usecase.IsFollowing(c.Context(), followerID, followedTarget)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"is_following": isFollowing,
		},
	})
}

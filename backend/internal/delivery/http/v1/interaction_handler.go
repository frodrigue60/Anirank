package v1

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/interaction"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type InteractionHandler struct {
	usecase *interaction.InteractionUsecase
}

func NewInteractionHandler(usecase *interaction.InteractionUsecase) *InteractionHandler {
	return &InteractionHandler{usecase: usecase}
}

// POST /api/interactions/ratings  (Song-only)
type rateRequest struct {
	SongID uint64  `json:"song_id"`
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

	average, rating, err := h.usecase.RateSong(c.Context(), userID, req.SongID, req.Score)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Rating submitted successfully",
		"average": average,
		"rating":  rating,
	})
}

// POST /api/interactions/reactions
type reactRequest struct {
	EntityID   uint64      `json:"entity_id"`
	SongID     uint64      `json:"song_id"`
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

	// Handle Song Reaction (new way)
	if req.SongID != 0 {
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

		likes, dislikes, err := h.usecase.ReactToSong(c.Context(), userID, req.SongID, reactionType)
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"success":       true,
			"likesCount":    likes,
			"dislikesCount": dislikes,
		})
	}

	// Fallback to polymorphic (old way) or EntityType branching
	var typeInt int8 = 0
	switch v := req.Type.(type) {
	case float64:
		typeInt = int8(v)
	case int:
		typeInt = int8(v)
	case string:
		if v == "like" || v == "1" {
			typeInt = 1
		} else if v == "dislike" || v == "-1" {
			typeInt = -1
		}
	}

	entityID := req.EntityID
	if entityID == 0 && req.SongID != 0 {
		entityID = req.SongID
	}

	if req.EntityType == "comment" || req.EntityType == "App\\Models\\Comment" {
		likes, dislikes, err := h.usecase.ReactToComment(c.Context(), userID, entityID, typeInt)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"success":       true,
			"likesCount":    likes,
			"dislikesCount": dislikes,
		})
	}

	err := h.usecase.ReactToEntity(c.Context(), userID, entityID, req.EntityType, typeInt)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Reaction updated",
	})
}

// POST /api/interactions/favorites
type favoriteRequest struct {
	EntityID   uint64 `json:"entity_id"`
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

	isFavorited, err := h.usecase.ToggleFavorite(c.Context(), userID, req.EntityID, req.EntityType)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"favorited": isFavorited,
	})
}

func (h *InteractionHandler) Feed(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	feed, err := h.usecase.GetGlobalActivity(c.Context(), limit)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": feed})
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

	if entityID == 0 || entityType == "" {
		return domain.NewAppError(400, "entity_id and entity_type are required", nil)
	}

	comments, err := h.usecase.GetCommentsByEntity(c.Context(), requestingUserID, entityID, entityType, limit, offset)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": comments})
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

	songIDParam := c.Params("id")
	songID, err := strconv.ParseUint(songIDParam, 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid song ID", err)
	}
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	comments, err := h.usecase.GetCommentsByEntity(c.Context(), requestingUserID, songID, "song", limit, offset)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": comments})
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
		return domain.NewAppError(400, "Invalid comment ID", err)
	}
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	replies, err := h.usecase.GetCommentReplies(c.Context(), requestingUserID, parentID, limit, offset)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": replies})
}

func (h *InteractionHandler) SongComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Invalid user context", nil)
	}

	type reqBody struct {
		EntityID   uint64  `json:"entity_id"`
		EntityType string  `json:"entity_type"`
		Content    string  `json:"content"`
		ParentID   *uint64 `json:"parent_id"`
	}

	var req reqBody
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid JSON payload", err)
	}

	comment, err := h.usecase.SongComment(c.Context(), userID, req.EntityID, req.EntityType, req.Content, req.ParentID)
	if err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": comment})
}

func (h *InteractionHandler) DeleteComment(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}
	commentIDParam := c.Params("id")
	commentID, err := strconv.ParseUint(commentIDParam, 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid comment ID", err)
	}
	err = h.usecase.DeleteComment(c.Context(), commentID, userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Comment deleted successfully"})
}

// ---- FOLLOWS ----

func (h *InteractionHandler) FollowUser(c *fiber.Ctx) error {
	followerID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	followedID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid user ID", err)
	}

	if err := h.usecase.FollowUser(c.Context(), followerID, followedID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User followed",
	})
}

func (h *InteractionHandler) UnfollowUser(c *fiber.Ctx) error {
	followerID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	followedID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid user ID", err)
	}

	if err := h.usecase.UnfollowUser(c.Context(), followerID, followedID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User unfollowed",
	})
}

func (h *InteractionHandler) IsFollowing(c *fiber.Ctx) error {
	followerID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return c.JSON(fiber.Map{"is_following": false})
	}

	followedID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid user ID", err)
	}

	isFollowing, err := h.usecase.IsFollowing(c.Context(), followerID, followedID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"is_following": isFollowing,
	})
}

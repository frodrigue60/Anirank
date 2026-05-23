package v1

import (
	"strconv"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"

	"github.com/gofiber/fiber/v2"
)

type RecommendationHandler struct {
	usecase domain.RecommendationUsecase
}

// NewRecommendationHandler crea una nueva instancia de los controladores de recomendación
func NewRecommendationHandler(u domain.RecommendationUsecase) *RecommendationHandler {
	return &RecommendationHandler{usecase: u}
}

func (h *RecommendationHandler) getUserID(c *fiber.Ctx) *uint64 {
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

// GetSimilarSongs maneja GET /api/v1/songs/:uuid/related
// @Summary Get Related Songs
// @Description Get a list of similar songs based on pgvector content embedding
// @Tags Recommendations
// @Produce json
// @Param uuid path string true "Song UUID"
// @Param limit query int false "Limit"
// @Success 200 {object} object{data=[]dto.SongSlimDTO}
// @Failure 404 {object} domain.AppError
// @Failure 500 {object} domain.AppError
// @Router /v1/songs/{uuid}/related [get]
func (h *RecommendationHandler) GetSimilarSongs(c *fiber.Ctx) error {
	songUUID := c.Params("uuid")
	if songUUID == "" {
		return domain.NewAppError(400, "Missing song UUID", nil)
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	userID := h.getUserID(c)

	songs, err := h.usecase.GetSimilarSongs(c.Context(), userID, songUUID, limit)
	if err != nil {
		return err
	}

	songDTOs := make([]dto.SongSlimDTO, len(songs))
	for i, s := range songs {
		songDTOs[i] = dto.ToSongSlimDTO(&s)
	}

	return c.JSON(fiber.Map{
		"data": songDTOs,
	})
}

// GetPersonalizedRecommendations maneja GET /api/v1/recommendations
// @Summary Get Personalized Recommendations
// @Description Get user-personalized recommendations based on their favorites and ratings
// @Tags Recommendations
// @Produce json
// @Param limit query int false "Limit"
// @Success 200 {object} object{data=[]dto.SongSlimDTO}
// @Failure 401 {object} domain.AppError
// @Failure 500 {object} domain.AppError
// @Router /v1/recommendations [get]
func (h *RecommendationHandler) GetPersonalizedRecommendations(c *fiber.Ctx) error {
	userID := h.getUserID(c)
	if userID == nil {
		return domain.NewAppError(401, "Authentication required for personalized recommendations", nil)
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	songs, err := h.usecase.GetPersonalizedRecommendations(c.Context(), *userID, limit)
	if err != nil {
		return err
	}

	songDTOs := make([]dto.SongSlimDTO, len(songs))
	for i, s := range songs {
		songDTOs[i] = dto.ToSongSlimDTO(&s)
	}

	return c.JSON(fiber.Map{
		"data": songDTOs,
	})
}

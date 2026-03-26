package v1

import (
	"anirank/api/internal/dto"
	"anirank/api/internal/domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ActivityHandler struct {
	usecase domain.ActivityUsecase
}

func NewActivityHandler(u domain.ActivityUsecase) *ActivityHandler {
	return &ActivityHandler{
		usecase: u,
	}
}

// Index handles GET /api/v1/activities?page=1
func (h *ActivityHandler) Index(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit := 20
	offset := (page - 1) * limit

	activities, err := h.usecase.GetFeed(c.Context(), limit, offset)
	if err != nil {
		return err
	}

	total, err := h.usecase.GetCount(c.Context())
	if err != nil {
		total = len(activities) // Fallback
	}

	dtoActivities := make([]dto.ActivityItemDTO, len(activities))
	for i, activity := range activities {
		dtoActivities[i] = dto.ToActivityDTO(activity)
	}

	return c.JSON(paginatedResponse(c, dtoActivities, total, page, limit))
}

// Recent handles GET /api/activities/recent.
// It always returns the newest 10 activities (no pagination).
func (h *ActivityHandler) Recent(c *fiber.Ctx) error {
	limit := 10
	activities, err := h.usecase.GetFeed(c.Context(), limit, 0)
	if err != nil {
		return err
	}

	dtoActivities := make([]dto.ActivityItemDTO, len(activities))
	for i, activity := range activities {
		dtoActivities[i] = dto.ToActivityDTO(activity)
	}

	// For "Recent" we don't necessarily need full pagination info,
	// but standardizing to the same structure with a simple data wrapper if preferred.
	// Actually, the user asked for the "convention", which includes the pagination block even for single pages.
	
	total := len(activities) // For 'recent', just show it as 1 page
	return c.JSON(paginatedResponse(c, dtoActivities, total, 1, limit))
}

package v1

import (
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

	return c.JSON(fiber.Map{
		"data":         activities,
		"current_page": page,
		"per_page":     limit,
	})
}

// Recent handles GET /api/activities/recent.
// It always returns the newest 20 activities (no pagination).
func (h *ActivityHandler) Recent(c *fiber.Ctx) error {
	limit := 20
	activities, err := h.usecase.GetFeed(c.Context(), limit, 0)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data":         activities,
		"current_page": 1,
		"per_page":     limit,
	})
}

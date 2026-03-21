package v1

import (
	"anirank/api/internal/usecase/public"

	"github.com/gofiber/fiber/v2"
)

type DiscoveryHandler struct {
	usecase *public.DiscoveryUsecase
}

func NewDiscoveryHandler(u *public.DiscoveryUsecase) *DiscoveryHandler {
	return &DiscoveryHandler{
		usecase: u,
	}
}

// Init handles GET /api/init loading all initial SPA requirements concurrently
// @Summary Init Data
// @Description Loads essential taxonomy and lookup arrays required for the SPA context (Years, Seasons, Formats, etc).
// @Tags Discovery
// @Produce json
// @Success 200 {object} object{data=map[string]interface{}}
// @Router /init [get]
func (h *DiscoveryHandler) Init(c *fiber.Ctx) error {
	data, err := h.usecase.GetInitData(c.Context())
	if err != nil {
		return err // Let middleware handle fiber errors
	}

	return c.JSON(fiber.Map{
		"data": data,
	})
}

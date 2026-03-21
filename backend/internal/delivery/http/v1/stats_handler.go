package v1

import (
	"anirank/api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type StatsHandler struct {
	statsUsecase domain.StatsUsecase
}

func NewStatsHandler(statsUsecase domain.StatsUsecase) *StatsHandler {
	return &StatsHandler{
		statsUsecase: statsUsecase,
	}
}

func (h *StatsHandler) GetSiteStats(c *fiber.Ctx) error {
	stats, err := h.statsUsecase.GetSiteStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}

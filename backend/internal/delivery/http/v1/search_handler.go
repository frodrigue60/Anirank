package v1

import (
	"anirank/api/internal/usecase/public"

	"github.com/gofiber/fiber/v2"
)

type SearchHandler struct {
	usecase *public.SearchUsecase
}

func NewSearchHandler(u *public.SearchUsecase) *SearchHandler {
	return &SearchHandler{
		usecase: u,
	}
}

// Search handles GET /api/search?q=term
// @Summary Global Search
// @Description Search across Animes, Songs, Artists, and Users using a single query term.
// @Tags Public
// @Produce json
// @Param q query string true "Search Term"
// @Success 200 {object} object{data=object}
// @Failure 500 {object} domain.AppError
// @Router /search [get]
func (h *SearchHandler) Search(c *fiber.Ctx) error {
	term := c.Query("q")

	result, err := h.usecase.GlobalSearch(c.Context(), term)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": result,
	})
}

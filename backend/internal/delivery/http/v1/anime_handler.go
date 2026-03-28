package v1

import (
	"strconv"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	"anirank/api/internal/usecase/public"

	"github.com/gofiber/fiber/v2"
)

type AnimeHandler struct {
	usecase *public.AnimeUsecase
}

func NewAnimeHandler(u *public.AnimeUsecase) *AnimeHandler {
	return &AnimeHandler{
		usecase: u,
	}
}

// Index handles GET /api/animes?page=1
// @Summary List Animes
// @Description Retrieve a paginated list of all Animes ordered by default index rules.
// @Tags Anime
// @Produce json
// @Param page query int false "Page number" default(1)
// @Success 200 {array} dto.AnimeDTO
// @Router /animes [get]
func (h *AnimeHandler) Index(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit := 24
	offset := (page - 1) * limit

	domainFilters := domain.AnimeFilters{
		Search:  c.Query("name", ""),
		Sort:    c.Query("sort", ""),
		Year:    c.Query("year", ""),
		Season:  c.Query("season", ""),
		Format:  c.Query("format", ""),
		Genre:   c.Query("genre", ""),
	}

	animes, total, err := h.usecase.GetPaginatedAnimes(c.Context(), limit, offset, domainFilters)
	if err != nil {
		return err
	}

	lastPage := int(total / limit)
	if total%limit != 0 {
		lastPage++
	}
	if lastPage < 1 {
		lastPage = 1
	}

	animeDTOs := make([]dto.AnimeDTO, len(animes))
	for i, a := range animes {
		animeDTOs[i] = dto.ToAnimeDTO(&a)
	}

	return c.JSON(paginatedResponse(c, animeDTOs, total, page, limit))
}

// Show returns the details of a specific Anime by its slug.
// @Summary Get Anime details
// @Description Retrieves all data associated with an Anime, including related songs and variants based on its slug.
// @Tags Anime
// @Produce json
// @Param slug path string true "Anime Slug"
// @Success 200 {object} object{data=domain.Anime}
// @Failure 404 {object} domain.AppError
// @Router /animes/{slug} [get]
func (h *AnimeHandler) Show(c *fiber.Ctx) error {
	slug := c.Params("slug")

	anime, err := h.usecase.GetAnimeBySlug(c.Context(), slug)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": dto.ToAnimeDTO(anime),
	})
}

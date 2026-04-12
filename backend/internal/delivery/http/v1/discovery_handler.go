package v1

import (
	"anirank/api/internal/dto"
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
		return err
	}

	// Map to DTO for hiding IDs
	yearDTOs := make([]dto.YearDTO, len(data.Years))
	for i, y := range data.Years {
		yearDTOs[i] = dto.ToYearDTO(&y)
	}

	seasonDTOs := make([]dto.SeasonDTO, len(data.Seasons))
	for i, s := range data.Seasons {
		seasonDTOs[i] = dto.ToSeasonDTO(&s)
	}

	formatDTOs := make([]dto.FormatDTO, len(data.Formats))
	for i, f := range data.Formats {
		formatDTOs[i] = dto.ToFormatDTO(&f)
	}

	genreDTOs := make([]dto.GenreDTO, len(data.Genres))
	for i, g := range data.Genres {
		genreDTOs[i] = dto.ToGenreDTO(&g)
	}

	songTypeDTOs := make([]dto.SongTypeDTO, len(data.SongTypes))
	for i, st := range data.SongTypes {
		songTypeDTOs[i] = dto.ToSongTypeDTO(&st)
	}

	return c.JSON(fiber.Map{
		"data": dto.InitDataDTO{
			Years:     yearDTOs,
			Seasons:   seasonDTOs,
			Formats:   formatDTOs,
			Genres:    genreDTOs,
			SongTypes: songTypeDTOs,
		},
	})
}

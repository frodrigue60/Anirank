package v1

import (
	"strconv"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/tournament"

	"github.com/gofiber/fiber/v2"
)

type TournamentHandler struct {
	usecase *tournament.TournamentUsecase
}

func NewTournamentHandler(usecase *tournament.TournamentUsecase) *TournamentHandler {
	return &TournamentHandler{usecase: usecase}
}

// ==== PUBLIC ENDPOINTS ====

// GetActiveTournament retrieves the currently running tournament and its bracket.
// @Summary Get Active Tournament
// @Description Fetch the active tournament and its current matchups/bracket.
// @Tags Tournaments
// @Produce json
// @Success 200 {object} object{data=domain.Tournament}
// @Failure 404 {object} object{error=string}
// @Router /tournaments/active [get]
func (h *TournamentHandler) GetActiveTournament(c *fiber.Ctx) error {
	t, err := h.usecase.GetActiveTournament(c.Context())
	if err != nil {
		if err == domain.ErrNotFound {
			return c.Status(404).JSON(fiber.Map{"error": "No active tournament found"})
		}
		return err
	}

	return c.JSON(fiber.Map{"data": t})
}

// ListTournamentsPublic lists active and completed tournaments for the public.
// @Summary List Public Tournaments
// @Tags Tournaments
// @Produce json
// @Success 200 {object} object{data=[]domain.Tournament}
// @Router /tournaments [get]
func (h *TournamentHandler) ListTournamentsPublic(c *fiber.Ctx) error {
	tt, err := h.usecase.ListPublic(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": tt})
}

// GetTournamentBySlug retrieves any tournament by its slug identifier.
// @Summary Get Tournament by Slug
// @Description Fetch a specific tournament's details and matchups by its slug.
// @Tags Tournaments
// @Produce json
// @Param slug path string true "Tournament Slug"
// @Success 200 {object} object{data=domain.Tournament}
// @Failure 400 {object} domain.AppError
// @Failure 404 {object} object{error=string}
// @Router /tournaments/{slug} [get]
func (h *TournamentHandler) GetTournamentBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return domain.NewAppError(400, "Slug is required", nil)
	}

	var userID *uint64
	if val, ok := c.Locals("user_id").(uint64); ok {
		userID = &val
	}

	t, err := h.usecase.GetTournamentBySlug(c.Context(), slug, userID)
	if err != nil {
		if err == domain.ErrNotFound {
			return c.Status(404).JSON(fiber.Map{"error": "Tournament not found"})
		}
		return err
	}

	return c.JSON(fiber.Map{"data": t})
}

// ==== USER FACING ENDPOINTS ====

// SubmitVote handles a user voting for a song in an active matchup.
// @Summary Submit Vote
// @Description Submit a vote for a specific song in an ongoing tournament matchup.
// @Tags Tournaments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Matchup ID"
// @Param request body object{song_id=int} true "Vote Data"
// @Success 201 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Failure 401 {object} domain.AppError
// @Router /tournaments/matchups/{id}/vote [post]
func (h *TournamentHandler) SubmitVote(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint64)
	if !ok {
		return domain.NewAppError(401, "Unauthorized", nil)
	}

	matchupID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid matchup ID", nil)
	}

	var payload struct {
		SongID uint64 `json:"song_id"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return domain.NewAppError(400, "Invalid JSON payload", err)
	}

	if payload.SongID == 0 {
		return domain.NewAppError(400, "song_id is required", nil)
	}

	ip := c.IP()
	if err := h.usecase.SubmitVote(c.Context(), userID, matchupID, payload.SongID, ip); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{"message": "Vote submitted successfully"})
}

// ==== ADMIN ENDPOINTS ====

// CreateTournament creates a new tournament.
// @Summary Create Tournament
// @Tags Admin/Tournaments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.Tournament true "Tournament Data"
// @Success 201 {object} object{message=string, data=domain.Tournament}
// @Router /admin/tournaments [post]
func (h *TournamentHandler) CreateTournament(c *fiber.Ctx) error {
	var t domain.Tournament
	if err := c.BodyParser(&t); err != nil {
		return domain.NewAppError(400, "Invalid JSON payload", err)
	}

	if err := h.usecase.CreateTournament(c.Context(), &t); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Tournament created successfully",
		"data":    t,
	})
}

// SeedTournament triggers the initial seeding of a draft tournament.
// @Summary Seed Tournament
// @Description Fetch top songs and create initial matchups for a draft tournament.
// @Tags Admin/Tournaments
// @Security BearerAuth
// @Produce json
// @Param id path int true "Tournament ID"
// @Success 200 {object} object{message=string}
// @Router /admin/tournaments/{id}/seed [post]
func (h *TournamentHandler) SeedTournament(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid tournament ID", nil)
	}

	var req domain.SeedRequest
	if err := c.BodyParser(&req); err != nil {
		// If empty body, default to legacy (global top songs)
		req.Method = "top"
	}

	if err := h.usecase.SeedTournament(c.Context(), id, req); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Tournament seeded and started successfully"})
}

// AdvanceTournament manually advances the tournament to the next phase.
// @Summary Advance Tournament Phase
// @Description Manually resolve current active matchups and advance to the next round.
// @Tags Admin/Tournaments
// @Security BearerAuth
// @Produce json
// @Param id path int true "Tournament ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} domain.AppError
// @Router /admin/tournaments/{id}/advance [post]
func (h *TournamentHandler) AdvanceTournament(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid tournament ID", nil)
	}

	if err := h.usecase.AdvanceTournament(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Tournament advanced to the next phase successfully"})
}

// GetTournament retrieves a specific tournament by its ID.
func (h *TournamentHandler) GetTournament(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid tournament ID", nil)
	}
	t, err := h.usecase.GetTournament(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": t})
}

// ListTournaments lists all tournaments for the admin panel.
// @Summary List All Tournaments
// @Tags Admin/Tournaments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} object{data=[]domain.Tournament}
// @Router /admin/tournaments [get]
func (h *TournamentHandler) ListTournaments(c *fiber.Ctx) error {
	tt, err := h.usecase.ListAll(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": tt})
}

// DeleteTournament deletes a tournament.
// @Summary Delete Tournament
// @Tags Admin/Tournaments
// @Security BearerAuth
// @Produce json
// @Param id path int true "Tournament ID"
// @Success 200 {object} object{message=string}
// @Router /admin/tournaments/{id} [delete]
func (h *TournamentHandler) DeleteTournament(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid tournament ID", nil)
	}

	if err := h.usecase.DeleteTournament(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Tournament deleted successfully"})
}

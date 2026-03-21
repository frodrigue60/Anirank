package v1

import (
	"strconv"

	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// Helper function locally, since RespondError is not in domain
func respondError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*domain.AppError); ok {
		return c.Status(appErr.Code).JSON(fiber.Map{
			"error": appErr.Message,
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Internal server error",
	})
}

type BadgeHandler struct {
	usecase domain.BadgeUsecase
}

func NewBadgeHandler(usecase domain.BadgeUsecase) *BadgeHandler {
	return &BadgeHandler{usecase: usecase}
}

func (h *BadgeHandler) getAuditMetadata(c *fiber.Ctx) domain.AuditMetadata {
	actorID, _ := c.Locals("user_id").(uint64)
	return domain.AuditMetadata{
		ActorID:   actorID,
		URL:       c.OriginalURL(),
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
	}
}

func (h *BadgeHandler) GetAll(c *fiber.Ctx) error {
	badges, err := h.usecase.GetAll(c.Context())
	if err != nil {
		return respondError(c, err)
	}
	return c.JSON(fiber.Map{"data": badges})
}

func (h *BadgeHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondError(c, domain.NewAppError(400, "Invalid ID", err))
	}

	badge, err := h.usecase.GetByID(c.Context(), id)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(fiber.Map{"data": badge})
}

func (h *BadgeHandler) Create(c *fiber.Ctx) error {
	var badge domain.Badge

	if c.Get("Content-Type") == "application/json" {
		if err := c.BodyParser(&badge); err != nil {
			return respondError(c, domain.NewAppError(400, "Invalid JSON body", err))
		}
	} else {
		// Form data parsing
		badge.Name = c.FormValue("name")
		description := c.FormValue("description")
		if description != "" {
			badge.Description = &description
		}
		badge.IsActive = c.FormValue("is_active") == "true"
	}

	if err := h.usecase.Create(c.Context(), &badge, h.getAuditMetadata(c)); err != nil {
		return respondError(c, err)
	}

	// Try processing image upload if present
	_ = h.usecase.HandleBadgeIcon(c, &badge)

	// Fetch fresh URL wrapper
	uBadge, _ := h.usecase.GetByID(c.Context(), badge.ID)
	if uBadge == nil {
		uBadge = &badge
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": uBadge})
}

func (h *BadgeHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondError(c, domain.NewAppError(400, "Invalid ID", err))
	}

	var badge domain.Badge
	if c.Get("Content-Type") == "application/json" {
		if err := c.BodyParser(&badge); err != nil {
			return respondError(c, domain.NewAppError(400, "Invalid JSON body", err))
		}
	} else {
		// Form data parsing
		badge.Name = c.FormValue("name")
		description := c.FormValue("description")
		if description != "" {
			badge.Description = &description
		}
		badge.IsActive = c.FormValue("is_active") == "true"
	}
	badge.ID = id

	if err := h.usecase.Update(c.Context(), &badge, h.getAuditMetadata(c)); err != nil {
		return respondError(c, err)
	}

	_ = h.usecase.HandleBadgeIcon(c, &badge)

	uBadge, _ := h.usecase.GetByID(c.Context(), badge.ID)
	if uBadge == nil {
		uBadge = &badge
	}

	return c.JSON(fiber.Map{"data": uBadge})
}

func (h *BadgeHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return respondError(c, domain.NewAppError(400, "Invalid ID", err))
	}

	if err := h.usecase.Delete(c.Context(), id, h.getAuditMetadata(c)); err != nil {
		return respondError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

package v1

import (
	"strconv"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	"anirank/api/internal/usecase/announcement"

	"github.com/gofiber/fiber/v2"
)

type AnnouncementHandler struct {
	usecase *announcement.AnnouncementUsecase
}

func NewAnnouncementHandler(usecase *announcement.AnnouncementUsecase) *AnnouncementHandler {
	return &AnnouncementHandler{usecase: usecase}
}

// GetPublicAnnouncements lists active announcements for the public sidebar.
func (h *AnnouncementHandler) GetPublicAnnouncements(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 5) // Fewer by default for public sidebar
	page, _ := strconv.Atoi(c.Query("page", "1"))

	filters := domain.AnnouncementFilters{
		ActiveOnly: true,
	}

	announcements, err := h.usecase.GetAllAnnouncements(c.Context(), filters, limit, offset)
	if err != nil {
		return err
	}

	total, err := h.usecase.GetCount(c.Context(), filters)
	if err != nil {
		total = len(announcements)
	}

	dtoAnnouncements := make([]dto.AnnouncementDTO, len(announcements))
	for i, a := range announcements {
		dtoAnnouncements[i] = dto.ToAnnouncementDTO(a)
	}

	return c.JSON(paginatedResponse(c, dtoAnnouncements, total, page, limit))
}

// GetAllAnnouncements lists all announcements for admin management.
func (h *AnnouncementHandler) GetAllAnnouncements(c *fiber.Ctx) error {
	limit, offset := parsePagination(c, 20)
	page, _ := strconv.Atoi(c.Query("page", "1"))

	filters := domain.AnnouncementFilters{
		Search: c.Query("search"),
	}
	announcements, err := h.usecase.GetAllAnnouncements(c.Context(), filters, limit, offset)
	if err != nil {
		return err
	}

	total, err := h.usecase.GetCount(c.Context(), filters)
	if err != nil {
		total = len(announcements)
	}

	dtoAnnouncements := make([]dto.AnnouncementDTO, len(announcements))
	for i, a := range announcements {
		dtoAnnouncements[i] = dto.ToAdminAnnouncementDTO(a)
	}

	return c.JSON(paginatedResponse(c, dtoAnnouncements, total, page, limit))
}

// GetAnnouncementByID retrieves a single announcement.
func (h *AnnouncementHandler) GetAnnouncementByID(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	a, err := h.usecase.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": dto.ToAdminAnnouncementDTO(*a)})
}

// CreateAnnouncement handles the creation of a new announcement.
func (h *AnnouncementHandler) CreateAnnouncement(c *fiber.Ctx) error {
	var a domain.Announcement
	if err := c.BodyParser(&a); err != nil {
		return domain.NewAppError(400, "Invalid request body", err)
	}

	// Manual override for dates if they are empty strings in the form
	// (BodyParser with multipart/form fails on empty date strings if not handled)
	a.StartsAt = parseOptionalTime(c.FormValue("starts_at"))
	a.EndsAt = parseOptionalTime(c.FormValue("ends_at"))

	// Handle Image Upload
	file, err := c.FormFile("image_file")
	if err == nil {
		path, err := h.usecase.UploadImage(c.Context(), file)
		if err != nil {
			return domain.NewAppError(500, "Failed to upload image", err)
		}
		a.Image = &path
	}

	if err := h.usecase.Create(c.Context(), &a); err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"data": dto.ToAdminAnnouncementDTO(a)})
}

// UpdateAnnouncement handles updating an existing announcement.
func (h *AnnouncementHandler) UpdateAnnouncement(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	// 1. Get existing announcement to preserve fields
	existing, err := h.usecase.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	var a domain.Announcement
	if err := c.BodyParser(&a); err != nil {
		return domain.NewAppError(400, "Invalid request body", err)
	}
	a.ID = id
	a.UUID = existing.UUID // Preserve UUID

	// Manual override for dates
	a.StartsAt = parseOptionalTime(c.FormValue("starts_at"))
	a.EndsAt = parseOptionalTime(c.FormValue("ends_at"))

	// Handle Image Upload
	file, err := c.FormFile("image_file")
	if err == nil {
		path, err := h.usecase.UploadImage(c.Context(), file)
		if err != nil {
			return domain.NewAppError(500, "Failed to upload image", err)
		}
		a.Image = &path
	} else {
		// No new file uploaded, preserve existing image
		a.Image = existing.Image
	}

	if err := h.usecase.Update(c.Context(), &a); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": dto.ToAdminAnnouncementDTO(a)})
}

func parseOptionalTime(val string) *time.Time {
	if val == "" || val == "null" || val == "undefined" {
		return nil
	}

	formats := []string{
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.Parse(format, val)
		if err == nil {
			return &t
		}
	}

	return nil
}

// DeleteAnnouncement removes an announcement.
func (h *AnnouncementHandler) DeleteAnnouncement(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.Delete(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(204)
}

// ToggleActive switches the is_active status.
func (h *AnnouncementHandler) ToggleActive(c *fiber.Ctx) error {
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	if err := h.usecase.ToggleActive(c.Context(), id); err != nil {
		return err
	}
	return c.SendStatus(200)
}

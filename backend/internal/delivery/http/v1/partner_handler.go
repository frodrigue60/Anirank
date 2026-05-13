package v1

import (
	"anirank/api/internal/domain"
	"strconv"
	"github.com/gofiber/fiber/v2"
)

type PartnerHandler struct {
	usecase domain.PartnerUsecase
}

func NewPartnerHandler(usecase domain.PartnerUsecase) *PartnerHandler {
	return &PartnerHandler{usecase: usecase}
}

// Public Methods

func (h *PartnerHandler) GetActivePartners(c *fiber.Ctx) error {
	partners, err := h.usecase.GetActivePartners(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(partners)
}

// Admin Methods

func (h *PartnerHandler) AdminGetAll(c *fiber.Ctx) error {
	partners, err := h.usecase.AdminGetAllPartners(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(partners)
}

func (h *PartnerHandler) AdminGetByUUID(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	p, err := h.usecase.AdminGetPartnerByUUID(c.Context(), uuid)
	if err != nil {
		return err
	}
	return c.JSON(p)
}

func (h *PartnerHandler) AdminCreate(c *fiber.Ctx) error {
	var p domain.Partner

	// Manually parse form values for better reliability with multipart
	p.Name = c.FormValue("name")
	p.URL = c.FormValue("url")
	p.Description = domain.Ptr(c.FormValue("description"))
	p.Type = c.FormValue("type")
	
	sortOrder, _ := strconv.Atoi(c.FormValue("sort_order"))
	p.SortOrder = sortOrder
	
	isActive := c.FormValue("is_active") == "true"
	p.IsActive = isActive

	// Handle banner upload if present
	if fileHeader, err := c.FormFile("banner_file"); err == nil {
		if file, err := fileHeader.Open(); err == nil {
			defer file.Close()
			if err := h.usecase.UploadBanner(c.Context(), &p, file); err != nil {
				return err
			}
		}
	}

	if err := h.usecase.AdminCreatePartner(c.Context(), &p); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(p)
}

func (h *PartnerHandler) AdminUpdate(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	var p domain.Partner

	// Manually parse form values
	p.Name = c.FormValue("name")
	p.URL = c.FormValue("url")
	p.Description = domain.Ptr(c.FormValue("description"))
	p.Type = c.FormValue("type")
	
	sortOrder, _ := strconv.Atoi(c.FormValue("sort_order"))
	p.SortOrder = sortOrder
	
	isActive := c.FormValue("is_active") == "true"
	p.IsActive = isActive

	// Handle banner upload if present
	if fileHeader, err := c.FormFile("banner_file"); err == nil {
		if file, err := fileHeader.Open(); err == nil {
			defer file.Close()
			if err := h.usecase.UploadBanner(c.Context(), &p, file); err != nil {
				return err
			}
		}
	}

	if err := h.usecase.AdminUpdatePartner(c.Context(), uuid, &p); err != nil {
		return err
	}

	return c.JSON(p)
}

func (h *PartnerHandler) AdminDelete(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if err := h.usecase.AdminDeletePartner(c.Context(), uuid); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

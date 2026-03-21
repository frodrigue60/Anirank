package v1

import (
	"strconv"

	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

type AuditLogHandler struct {
	usecase domain.AuditLogUsecase
}

func NewAuditLogHandler(u domain.AuditLogUsecase) *AuditLogHandler {
	return &AuditLogHandler{usecase: u}
}

func (h *AuditLogHandler) GetAuditLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters["user_id"] = id
		}
	}
	if event := c.Query("event"); event != "" {
		filters["event"] = event
	}
	if auditableType := c.Query("auditable_type"); auditableType != "" {
		filters["auditable_type"] = auditableType
	}
	if auditableID := c.Query("auditable_id"); auditableID != "" {
		if id, err := strconv.ParseUint(auditableID, 10, 64); err == nil {
			filters["auditable_id"] = id
		}
	}

	logs, total, err := h.usecase.GetAuditLogs(c.Context(), page, limit, filters)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"data":         logs,
		"total":        total,
		"current_page": page,
		"per_page":     limit,
	})
}

func (h *AuditLogHandler) GetAuditLog(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return domain.NewAppError(400, "Invalid ID", err)
	}

	log, err := h.usecase.GetAuditLog(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    log,
	})
}

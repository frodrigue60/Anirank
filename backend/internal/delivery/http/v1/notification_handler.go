package v1

import (
	"math"
	"strconv"

	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

type NotificationHandler struct {
	usecase domain.NotificationUsecase
}

func NewNotificationHandler(u domain.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{usecase: u}
}

func (h *NotificationHandler) Index(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	notificationType := c.Query("type", "")
	limit := 20
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	notifications, total, unreadCount, err := h.usecase.GetNotifications(c.Context(), userID, notificationType, limit, offset)
	if err != nil {
		return err
	}

	lastPage := int(math.Ceil(float64(total) / float64(limit)))
	if lastPage < 1 {
		lastPage = 1
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":         notifications,
		"total":        total,
		"unread_count": unreadCount,
		"current_page": page,
		"last_page":    lastPage,
		"per_page":     limit,
	})
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)
	id := c.Params("id")

	if err := h.usecase.MarkAsRead(c.Context(), id, userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)

	if err := h.usecase.MarkAllAsRead(c.Context(), userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *NotificationHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)
	id := c.Params("id")

	if err := h.usecase.Delete(c.Context(), id, userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)
	count, err := h.usecase.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"count": count})
}

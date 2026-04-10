package v1

import (
	"anirank/api/internal/dto"
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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

	dtoNotifications := make([]dto.NotificationDTO, len(notifications))
	for i, n := range notifications {
		dtoNotifications[i] = dto.ToNotificationDTO(n)
	}

	response := paginatedResponse(c, dtoNotifications, total, page, limit)
	response["unread_count"] = unreadCount

	return c.JSON(response)
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)
	id := c.Params("id")

	if err := h.usecase.MarkAsRead(c.Context(), id, userID); err != nil {
		return err
	}

	return c.SendStatus(200)
}

func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)

	if err := h.usecase.MarkAllAsRead(c.Context(), userID); err != nil {
		return err
	}

	return c.SendStatus(200)
}

func (h *NotificationHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)
	id := c.Params("id")

	if err := h.usecase.Delete(c.Context(), id, userID); err != nil {
		return err
	}

	return c.SendStatus(204)
}

func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)
	count, err := h.usecase.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"count": count,
		},
	})
}

func (h *NotificationHandler) Stream(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	userID := c.Locals("user_id").(uint64)

	sub, err := h.usecase.SubscribeToStream(c.Context(), userID)
	if err != nil {
		return err
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer sub.Close()

		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg, ok := <-sub.Channel():
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg.Payload)
				if err := w.Flush(); err != nil {
					return
				}
			case <-ticker.C:
				fmt.Fprintf(w, ": ping\n\n")
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}

func (h *NotificationHandler) GetSettings(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)

	settings, err := h.usecase.GetSettings(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data": settings,
	})
}

func (h *NotificationHandler) UpdateSettings(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint64)

	var payload struct {
		Settings json.RawMessage `json:"settings"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return domain.NewAppError(400, "Invalid settings payload", err)
	}

	if err := h.usecase.UpdateSettings(c.Context(), userID, payload.Settings); err != nil {
		return err
	}

	return c.SendStatus(200)
}

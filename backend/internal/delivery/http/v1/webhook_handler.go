package v1

import (
	"anirank/api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	usecase domain.WebhookUsecase
}

func NewWebhookHandler(usecase domain.WebhookUsecase) *WebhookHandler {
	return &WebhookHandler{usecase: usecase}
}

func (h *WebhookHandler) GetAll(c *fiber.Ctx) error {
	webhooks, err := h.usecase.GetAll(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": webhooks})
}

func (h *WebhookHandler) GetByUUID(c *fiber.Ctx) error {
	webhook, err := h.usecase.GetByUUID(c.Context(), c.Params("uuid"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": webhook})
}

func (h *WebhookHandler) Create(c *fiber.Ctx) error {
	var webhook domain.Webhook
	if err := c.BodyParser(&webhook); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.Create(c.Context(), &webhook); err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{"data": webhook})
}

func (h *WebhookHandler) Update(c *fiber.Ctx) error {
	var webhook domain.Webhook
	if err := c.BodyParser(&webhook); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}
	webhook.UUID = c.Params("uuid")

	if err := h.usecase.Update(c.Context(), &webhook); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": webhook})
}

func (h *WebhookHandler) Delete(c *fiber.Ctx) error {
	if err := h.usecase.Delete(c.Context(), c.Params("uuid")); err != nil {
		return err
	}
	return c.SendStatus(204)
}

func (h *WebhookHandler) Test(c *fiber.Ctx) error {
	if err := h.usecase.TestWebhook(c.Context(), c.Params("uuid")); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Test message sent successfully"})
}

func (h *WebhookHandler) TriggerForAnime(c *fiber.Ctx) error {
	var req struct {
		AnimeID uint64 `json:"anime_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.TriggerForAnime(c.Context(), c.Params("uuid"), req.AnimeID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Notification sent successfully"})
}

func (h *WebhookHandler) TriggerForSong(c *fiber.Ctx) error {
	var req struct {
		SongID uint64 `json:"song_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.TriggerForSong(c.Context(), c.Params("uuid"), req.SongID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Notification sent successfully"})
}

func (h *WebhookHandler) NotifyNewAnime(c *fiber.Ctx) error {
	var req struct {
		AnimeID uint64 `json:"anime_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.NotifyNewAnime(c.Context(), req.AnimeID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Notifications sent successfully to all active webhooks"})
}

func (h *WebhookHandler) NotifyNewSong(c *fiber.Ctx) error {
	var req struct {
		SongID uint64 `json:"song_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.NotifyNewSong(c.Context(), req.SongID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Notifications sent successfully to all active webhooks"})
}

func (h *WebhookHandler) NotifyCustomMessage(c *fiber.Ctx) error {
	var req struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.NotifyCustomMessage(c.Context(), req.Title, req.Message); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Custom notifications sent successfully"})
}

func (h *WebhookHandler) SendCustomMessage(c *fiber.Ctx) error {
	var req struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if err := h.usecase.SendCustomMessage(c.Context(), c.Params("uuid"), req.Title, req.Message); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Custom notification sent successfully"})
}

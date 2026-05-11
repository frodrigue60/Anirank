package v1

import (
	"bufio"
	"context"
	"fmt"
	"anirank/api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func (h *AdminHandler) SearchAnimeThemes(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return domain.NewAppError(400, "Query is required", nil)
	}

	results, err := h.usecase.SearchAnimeThemes(c.Context(), query)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": results})
}

func (h *AdminHandler) HydrateAnimeThemes(c *fiber.Ctx) error {
	var req struct {
		IDs      []uint64 `json:"ids"`
		Language string   `json:"language"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if len(req.IDs) == 0 {
		return domain.NewAppError(400, "IDs are required", nil)
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	progress := make(chan string)
	meta := h.getAuditMetadata(c)

	go func() {
		defer close(progress)
		// Use Background() instead of c.Context() because Fiber context is recycled
		_ = h.usecase.HydrateAnimeThemes(context.Background(), req.IDs, req.Language, meta, progress)
	}()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for msg := range progress {
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.Flush()
		}
	})

	return nil
}

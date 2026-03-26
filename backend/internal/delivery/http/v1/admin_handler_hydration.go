package v1

import (
	"bufio"
	"context"
	"fmt"
	"anirank/api/internal/domain"
	"github.com/gofiber/fiber/v2"
)

func (h *AdminHandler) HydrateAnimeSeason(c *fiber.Ctx) error {
	var req struct {
		Year   int    `json:"year"`
		Season string `json:"season"`
	}
	if err := c.BodyParser(&req); err != nil {
		return domain.NewAppError(400, "Invalid payload", err)
	}

	if req.Year == 0 || req.Season == "" {
		return domain.NewAppError(400, "Year and Season are required", nil)
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
		_ = h.usecase.HydrateAnimeSeason(context.Background(), req.Year, req.Season, meta, progress)
	}()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for msg := range progress {
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.Flush()
		}
	})

	return nil
}

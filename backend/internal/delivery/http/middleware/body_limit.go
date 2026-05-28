package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// BodyLimit restricts the request body size.
// It applies a standard limit to non-multipart requests, allowing multipart/form-data to go up to Fiber's global limit.
func BodyLimit(maxStandardBytes int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		contentType := c.Get("Content-Type")
		if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
			return c.Next()
		}

		// Content-Length check is fast and avoids reading the body if it's already known to be too large
		if c.Request().Header.ContentLength() > maxStandardBytes {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "Request entity too large",
			})
		}
		// Double check actual read body length for chunked requests
		if len(c.Body()) > maxStandardBytes {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "Request entity too large",
			})
		}
		return c.Next()
	}
}

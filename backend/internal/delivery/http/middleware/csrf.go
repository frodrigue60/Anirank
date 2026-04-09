package middleware

import (
	"anirank/api/internal/domain"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

// NewCSRFMiddleware creates a CSRF protection middleware.
// It uses the provided storage (Redis or Memory fallback).
func NewCSRFMiddleware(storage fiber.Storage) fiber.Handler {
	return csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookieSameSite: "Lax",
		CookieHTTPOnly: false, // Standard for SPAs so Axios can read it
		Expiration:     1 * time.Hour,
		Storage:        storage,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Return a clean AppError for CSRF failures
			return domain.NewAppError(403, "Invalid or missing CSRF token", err)
		},
	})
}

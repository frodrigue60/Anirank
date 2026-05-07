package middleware

import (
	"anirank/api/internal/domain"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

// NewCSRFMiddleware creates a CSRF protection middleware.
// It uses the provided storage (Redis or Memory fallback).
func NewCSRFMiddleware(storage fiber.Storage) fiber.Handler {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")

	return csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookieSameSite: "None",
		CookieSecure:   true, // Must be true for SameSite: None
		CookieDomain:   cookieDomain,
		CookieHTTPOnly: false,
		Expiration:     1 * time.Hour,
		Storage:        storage,
		Next: func(c *fiber.Ctx) bool {
			// Skip CSRF for OAuth callbacks as they are external redirects/initiated by OAuth flows
			path := c.Path()
			return strings.HasSuffix(path, "/callback") || strings.HasSuffix(path, "/login-callback")
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Return a clean AppError for CSRF failures
			return domain.NewAppError(403, "Invalid or missing CSRF token", err)
		},
	})
}

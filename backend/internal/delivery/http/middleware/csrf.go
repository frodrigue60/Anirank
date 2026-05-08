package middleware

import (
	"anirank/api/internal/domain"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// NewCSRFMiddleware creates a custom CSRF protection middleware using the
// Double Submit Cookie pattern (OWASP recommended for SPAs with cross-subdomain APIs).
//
// Why custom instead of Fiber's built-in:
//   Fiber's csrf middleware calls refererMatchesHost() on all HTTPS POST/PUT/DELETE
//   requests, which compares refererURL.Host == c.Hostname() literally.
//   Since our frontend (anirank.work) and API (api.anirank.work) are different hosts,
//   this check always fails with ErrBadReferer — regardless of SameSite, TrustedOrigins,
//   or any cookie configuration.
//
// How Double Submit Cookie works:
//   1. On safe requests (GET/HEAD/OPTIONS), a csrf_token cookie is issued if absent.
//   2. The frontend JS reads the cookie value and sends it as the X-CSRF-Token header.
//   3. On unsafe requests (POST/PUT/DELETE/PATCH), we verify cookie == header.
//   4. An attacker on another origin cannot read our cookie (Same-Origin Policy),
//      so they cannot forge the matching header value — CSRF is prevented.
func NewCSRFMiddleware(_ fiber.Storage) fiber.Handler {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")

	return func(c *fiber.Ctx) error {
		method := c.Method()

		// 1. Skip OAuth callbacks — initiated by external OAuth providers
		path := c.Path()
		if strings.HasSuffix(path, "/callback") || strings.HasSuffix(path, "/login-callback") {
			return c.Next()
		}

		// 2. On safe methods: issue the CSRF cookie if missing, then continue
		if method == fiber.MethodGet || method == fiber.MethodHead ||
			method == fiber.MethodOptions || method == fiber.MethodTrace {
			issueCSRFCookie(c, cookieDomain)
			return c.Next()
		}

		// 3. On unsafe methods: validate Double Submit Cookie
		cookieToken := c.Cookies("csrf_token")
		headerToken := c.Get("X-CSRF-Token")

		if cookieToken == "" || headerToken == "" || cookieToken != headerToken {
			return domain.NewAppError(403, "Invalid or missing CSRF token", nil)
		}

		return c.Next()
	}
}

// issueCSRFCookie sets a new csrf_token cookie on the response if one is not already
// present in the request. The cookie is readable by JS (HTTPOnly: false) so the
// frontend can copy its value into the X-CSRF-Token header on subsequent requests.
func issueCSRFCookie(c *fiber.Ctx, cookieDomain string) {
	if c.Cookies("csrf_token") != "" {
		return // Token already exists in the request — no need to regenerate
	}

	cookieSecure := os.Getenv("COOKIE_SECURE") != "false"

	token := uuid.New().String()
	c.Cookie(&fiber.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Domain:   cookieDomain,
		Path:     "/",
		Secure:   cookieSecure,
		HTTPOnly: false, // Must be false — JS needs to read this value
		SameSite: "Lax",  // Lax is safe here: we validate via header, not SameSite
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders adds modern security headers to all responses.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Content Security Policy (CSP)
		// We allow:
		// - Self: Our own domain
		// - https://api.anirank.work: Our API
		// - https://*.google.com, https://*.gstatic.com: Google Auth & Fonts
		// - https://*.anilist.co: AniList Assets
		// - https://*.discord.com: Discord Assets
		// - data: For inline images/base64
		// - blob: For some media handling
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://*.google.com https://*.anilist.co; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"img-src 'self' data: blob: https://*.anirank.work https://*.googleusercontent.com https://*.anilist.co https://*.discordapp.com https://*.media-amazon.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"connect-src 'self' https://api.anirank.work https://*.google.com https://*.anilist.co; " +
			"frame-src 'self' https://www.youtube-nocookie.com https://www.youtube.com https://*.google.com; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none'; " + // Clickjacking protection (equivalent to X-Frame-Options: DENY)
			"upgrade-insecure-requests;"

		c.Set("Content-Security-Policy", csp)

		// 2. HTTP Strict Transport Security (HSTS)
		// 2 years (63072000 seconds), includes subdomains and preload
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		// 3. Cross-Origin Opener Policy (COOP)
		c.Set("Cross-Origin-Opener-Policy", "same-origin")

		// 4. Cross-Origin Resource Policy (CORP)
		c.Set("Cross-Origin-Resource-Policy", "cross-origin")

		// 5. X-Content-Type-Options
		c.Set("X-Content-Type-Options", "nosniff")

		// 6. Referrer Policy
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// 7. Permissions Policy (Minimal for now)
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), interest-cohort=()")

		return c.Next()
	}
}

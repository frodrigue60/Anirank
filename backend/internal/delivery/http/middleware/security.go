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
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://*.google.com https://*.anilist.co https://static.cloudflareinsights.com https://cloudflareinsights.com; " +
			"script-src-elem 'self' 'unsafe-inline' https://*.google.com https://*.anilist.co https://static.cloudflareinsights.com https://cloudflareinsights.com; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"img-src 'self' data: blob: http://localhost:8080 http://localhost:9000 https://*.anirank.work https://*.googleusercontent.com https://*.anilist.co https://*.discordapp.com https://*.media-amazon.com https://*.r2.dev; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"connect-src 'self' http://localhost:8080 ws://localhost:8080 https://api.anirank.work wss://api.anirank.work https://*.google.com https://*.anilist.co https://cloudflareinsights.com; " +
			"frame-src 'self' https://www.youtube-nocookie.com https://www.youtube.com https://*.google.com; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none'; " +
			"media-src 'self' blob: http://localhost:8080 http://localhost:9000 https://*.anirank.work https://*.r2.dev; " +
			"upgrade-insecure-requests; " +
			"trusted-types svelte-trusted-html 'allow-duplicates';"

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
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		return c.Next()
	}
}

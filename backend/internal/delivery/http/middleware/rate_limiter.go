package middleware

import (
	"anirank/api/internal/domain"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// getClientIP extracts the real client IP, considering proxies.
func getClientIP(c *fiber.Ctx) string {
	ip := c.Get("CF-Connecting-IP")
	if ip == "" {
		ip = c.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = c.IP()
	}
	return ip
}

// NewAuthLimiter creates a rate limiter for sensitive authentication routes.
func NewAuthLimiter(storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10, // Increased to 10 for better dev experience
		Expiration: 10 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "limiter:auth:" + getClientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return domain.NewAppError(429, "Too many login attempts. Please try again in 10 minutes.", nil)
		},
		Storage: storage,
	})
}

// NewPublicApiLimiter creates a rate limiter for public catalog endpoints.
func NewPublicApiLimiter(storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100, // Increased for better UX during navigation
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "limiter:public:" + getClientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return domain.NewAppError(429, "Too many requests. Please slow down and try again.", nil)
		},
		Storage: storage,
	})
}

// NewOGLimiter creates a stricter rate limiter for OG image endpoints (CPU/RAM heavy on cache miss).
func NewOGLimiter(storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        30,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "limiter:og:" + getClientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Retry-After", "60")
			return domain.NewAppError(429, "Too many OG image requests. Please try again shortly.", nil)
		},
		Storage: storage,
	})
}

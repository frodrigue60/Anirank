package middleware

import (
	"anirank/api/internal/domain"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// NewAuthLimiter creates a rate limiter for sensitive authentication routes.
// It uses the provided storage (Redis or Memory fallback).
func NewAuthLimiter(storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:               5,
		Expiration:        10 * time.Minute,
		KeyGenerator:      func(c *fiber.Ctx) string {
			// Use CF-Connecting-IP or X-Forwarded-For if behind a proxy
			ip := c.Get("CF-Connecting-IP")
			if ip == "" {
				ip = c.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = c.IP()
			}
			return ip
		},
		LimitReached: func(c *fiber.Ctx) error {
			return domain.NewAppError(429, "Too many login attempts. Please try again in 10 minutes.", nil)
		},
		Storage: storage,
	})
}

// NewPublicApiLimiter creates a rate limiter for public catalog endpoints.
// Limits requests to 60 per minute per IP to prevent scraping and abuse.
func NewPublicApiLimiter(storage fiber.Storage) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:               60,
		Expiration:        1 * time.Minute,
		KeyGenerator:      func(c *fiber.Ctx) string {
			ip := c.Get("CF-Connecting-IP")
			if ip == "" {
				ip = c.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = c.IP()
			}
			return ip
		},
		LimitReached: func(c *fiber.Ctx) error {
			return domain.NewAppError(429, "Too many requests. Please slow down and try again.", nil)
		},
		Storage: storage,
	})
}

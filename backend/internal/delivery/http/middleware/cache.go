package middleware

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
)

// NewResponseCache creates a caching middleware for public routes.
// It bypasses caching if CACHE_ENABLED is not "true" or if a user is authenticated.
func NewResponseCache(storage fiber.Storage, expiration time.Duration) fiber.Handler {
	cacheEnabled := os.Getenv("CACHE_ENABLED") == "true"

	return cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			// Bypass if globally disabled
			if !cacheEnabled {
				return true
			}

			// Bypass if user is authenticated to avoid caching personalized data
			// (user_uuid is set by AuthMiddleware or OptionalAuthMiddleware)
			if c.Locals("user_uuid") != nil {
				return true
			}

			// Only cache GET requests
			return c.Method() != fiber.MethodGet
		},
		Expiration:   expiration,
		CacheHeader:  "X-Cache", // Useful for debugging (HIT/MISS/BYPASS)
		Storage:      storage,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Cache key based on the full URL path and query parameters
			return c.OriginalURL()
		},
	})
}

// LogCacheStatus is an optional middleware to see cache results in console during dev
func LogCacheStatus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		cacheHeader := c.Get("X-Cache")
		if cacheHeader != "" {
			log.Printf("[CACHE] %s %s - %s", c.Method(), c.Path(), cacheHeader)
		}
		return err
	}
}

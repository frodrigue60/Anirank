package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger is a simple middleware that logs incoming requests.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Continue stack
		err := c.Next()

		latency := time.Since(start)
		log.Printf("%s - %s %s - %v - %d",
			c.IP(),
			c.Method(),
			c.Path(),
			latency,
			c.Response().StatusCode(),
		)

		return err
	}
}

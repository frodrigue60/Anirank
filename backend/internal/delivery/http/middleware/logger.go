package middleware

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// requestLogLevel controls which HTTP responses are written to stdout.
// Set via LOG_LEVEL: error (default) | warn | info
//
//	error — status >= 500 only
//	warn  — status >= 400
//	info  — every request (previous behavior)
func requestLogLevel() string {
	level := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	switch level {
	case "info", "warn", "error":
		return level
	default:
		return "error"
	}
}

func shouldLogRequest(status int, level string) bool {
	switch level {
	case "info":
		return true
	case "warn":
		return status >= 400
	default: // error
		return status >= 500
	}
}

// RequestLogger logs HTTP requests according to LOG_LEVEL.
func RequestLogger() fiber.Handler {
	level := requestLogLevel()

	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()

		if !shouldLogRequest(status, level) {
			return err
		}

		log.Printf("%s - %s %s - %v - %d",
			c.IP(),
			c.Method(),
			c.Path(),
			time.Since(start),
			status,
		)

		return err
	}
}

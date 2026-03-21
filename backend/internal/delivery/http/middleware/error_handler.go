package middleware

import (
	"errors"
	"fmt"

	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler is a custom error handler for the Fiber application
func ErrorHandler(c *fiber.Ctx, err error) error {
	fmt.Printf("[ERROR] Path: %s, Error: %v\n", c.Path(), err)
	// Default HTTP status code
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if the error is of type *domain.AppError
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		code = appErr.Code
		message = appErr.Message
	} else if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Send JSON response with error details
	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": message,
	})
}

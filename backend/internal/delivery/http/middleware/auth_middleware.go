package middleware

import (
	"log"
	"strings"

	"anirank/api/internal/domain"
	"anirank/api/internal/usecase/auth"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT context over routes
func AuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return domain.NewAppError(401, "Missing authorization header", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return domain.NewAppError(401, "Invalid authorization header format", nil)
		}

		tokenString := parts[1]
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("Auth Error: %v", err)
			return domain.NewAppError(401, "Invalid or expired token", nil)
		}

		// Store user info in Context Locals for downstream handlers
		c.Locals("user_id", claims.UserID)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

// OptionalAuthMiddleware parses JWT if present, but allows request to continue if not
func OptionalAuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			if claims, err := jwtService.ValidateToken(tokenString); err == nil {
				c.Locals("user_id", claims.UserID)
				c.Locals("user_role", claims.Role)
			}
		}

		return c.Next()
	}
}

// StaffMiddleware only allows users with core administrative type roles to proceed.
func StaffMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok || (role != "admin" && role != "editor" && role != "creator") {
			return domain.NewAppError(403, "Forbidden. Action requires staff privileges", nil)
		}
		return c.Next()
	}
}

// AdminMiddleware strictly only allows top-level admin
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok || role != "admin" {
			return domain.NewAppError(403, "Forbidden. Feature restricted to Administrators", nil)
		}
		return c.Next()
	}
}

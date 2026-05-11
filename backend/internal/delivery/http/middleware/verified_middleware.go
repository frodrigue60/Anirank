package middleware

import (
	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

// VerifiedMiddleware ensures the user has a verified email address.
// It depends on AuthMiddleware or OptionalAuthMiddleware being called before to inject "user_verified" into c.Locals.
func VerifiedMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Staff members (admin, owner, editor, creator) bypass verification for administrative safety
		// but since they are manually assigned, they are expected to be verified.
		roles, _ := c.Locals("user_roles").([]string)
		for _, r := range roles {
			if r == "owner" || r == "admin" || r == "editor" || r == "creator" {
				return c.Next()
			}
		}

		verified, ok := c.Locals("user_verified").(bool)
		if !ok || !verified {
			return domain.NewAppError(403, "Email verification required to perform this action", nil)
		}

		return c.Next()
	}
}

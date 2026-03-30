package middleware

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/pkg/rbac"
	"github.com/gofiber/fiber/v2"
)

// HasPermissionMiddleware creates a Fiber handler that checks for a specific permission.
func HasPermissionMiddleware(permission string, repo domain.UserRepository) fiber.Handler {
	pm := rbac.GetPermissionManager(repo)
	
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("user_role").(string)
		if !ok || role == "" {
			return domain.NewAppError(403, "Access denied. Role not found.", nil)
		}

		// Owner Bypass: Owners always have access.
		if role == "owner" {
			return c.Next()
		}

		// Dynamic check against cached permissions
		if pm.HasPermission(role, permission) {
			return c.Next()
		}

		return domain.NewAppError(403, "Forbidden. Action requires permission: "+permission, nil)
	}
}

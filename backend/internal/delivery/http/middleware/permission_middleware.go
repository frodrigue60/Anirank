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
		roles, ok := c.Locals("user_roles").([]string)
		if !ok || len(roles) == 0 {
			return domain.NewAppError(403, "Access denied. Roles not found.", nil)
		}

		// Owner Bypass: Owners always have access.
		for _, role := range roles {
			if role == "owner" {
				return c.Next()
			}
		}

		// Dynamic check against cached permissions for any of the user's roles
		if pm.HasAnyRolePermission(roles, permission) {
			return c.Next()
		}

		return domain.NewAppError(403, "Forbidden. Action requires permission: "+permission, nil)
	}
}

package middleware

import (
	"net/http/httptest"
	"testing"

	"anirank/api/internal/domain"
	"anirank/api/internal/pkg/rbac"
	"anirank/api/internal/testutil"

	"github.com/gofiber/fiber/v2"
)

func TestHasPermissionMiddleware(t *testing.T) {
	// Setup Mock Repo and RBAC Manager
	mockRepo := &testutil.MockUserRepository{
		GetRolesFunc: func() ([]domain.Role, error) {
			return []domain.Role{
				{ID: 1, Name: "Admin", Slug: "admin"},
				{ID: 2, Name: "Editor", Slug: "editor"},
			}, nil
		},
		GetPermsFunc: func(roleID uint64) ([]domain.Permission, error) {
			if roleID == 1 {
				return []domain.Permission{{ID: 1, Name: "Delete Anime", Slug: "anime.delete"}}, nil
			}
			return nil, nil
		},
	}

	// Reset and initialize RBAC manager with mock
	rbac.Reset()
	rbac.GetPermissionManager(mockRepo)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*domain.AppError); ok {
				return c.Status(e.Code).JSON(fiber.Map{"message": e.Message})
			}
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		},
	})
	
	// Identity simulator (Sets roles in Locals)
	app.Use(func(c *fiber.Ctx) error {
		roles := c.Query("roles")
		if roles != "" {
			c.Locals("user_roles", []string{roles})
		}
		return c.Next()
	})

	app.Get("/delete-anime", HasPermissionMiddleware("anime.delete", mockRepo), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	t.Run("Owner Bypass", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/delete-anime?roles=owner", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200 for owner bypass, got %d", resp.StatusCode)
		}
	})

	t.Run("Role with Permission", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/delete-anime?roles=admin", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200 for admin with permission, got %d", resp.StatusCode)
		}
	})

	t.Run("Role without Permission", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/delete-anime?roles=editor", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 403 {
			t.Errorf("Expected 403 for editor without permission, got %d", resp.StatusCode)
		}
	})

	t.Run("No Roles in Context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/delete-anime", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 403 {
			t.Errorf("Expected 403 for user without roles, got %d", resp.StatusCode)
		}
	})
}

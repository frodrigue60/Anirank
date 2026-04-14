package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"
	"anirank/api/internal/usecase/auth"

	"github.com/gofiber/fiber/v2"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test_secret_key_12345")
	jwtSvc := auth.NewJWTService()
	testUser := &domain.User{ID: 1, UUID: "test-uuid-1"}
	mockRepo := &testutil.MockUserRepository{User: testUser}

	app := testutil.NewTestApp()
	app.Get("/protected", AuthMiddleware(jwtSvc, mockRepo), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		userUUID := c.Locals("user_uuid")
		return c.Status(200).JSON(fiber.Map{
			"user_id":   userID,
			"user_uuid": userUUID,
		})
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 401 {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("Invalid Header Format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		resp, _ := app.Test(req)
		if resp.StatusCode != 401 {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("Valid Token - Success", func(t *testing.T) {
		token, _ := testutil.CreateTestToken(testUser.UUID, []string{"user"})
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		token, _ := testutil.CreateExpiredTestToken(testUser.UUID, []string{"user"})
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		resp, _ := app.Test(req)
		if resp.StatusCode != 401 {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("User Missing in DB", func(t *testing.T) {
		mockRepoMissing := &testutil.MockUserRepository{Err: domain.ErrNotFound}
		appMissing := testutil.NewTestApp()
		appMissing.Get("/protected", AuthMiddleware(jwtSvc, mockRepoMissing), func(c *fiber.Ctx) error {
			return c.SendStatus(200)
		})

		token, _ := testutil.CreateTestToken("missing-uuid", []string{"user"})
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		resp, _ := appMissing.Test(req)
		if resp.StatusCode != 401 {
			t.Errorf("Expected 401 for missing user, got %d", resp.StatusCode)
		}
	})
}

func TestOptionalAuthMiddleware(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key_12345")
	jwtSvc := auth.NewJWTService()
	testUser := &domain.User{ID: 1, UUID: "test-uuid-1"}
	mockRepo := &testutil.MockUserRepository{User: testUser}

	app := testutil.NewTestApp()
	app.Get("/optional", OptionalAuthMiddleware(jwtSvc, mockRepo), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		return c.Status(200).JSON(fiber.Map{"user_id": userID})
	})

	t.Run("No token - Proceed as Guest", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/optional", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Valid token - Identify User", func(t *testing.T) {
		token, _ := testutil.CreateTestToken(testUser.UUID, []string{"user"})
		req := httptest.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestRoleMiddleware(t *testing.T) {
	app := testutil.NewTestApp()
	
	// Dummy endpoint with Staff check
	app.Get("/staff", func(c *fiber.Ctx) error {
		// Simulate AuthMiddleware setting roles
		roles := c.Query("roles")
		if roles != "" {
			c.Locals("user_roles", []string{roles})
		}
		return c.Next()
	}, StaffMiddleware(), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	t.Run("As Admin - Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/staff?roles=admin", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("As Regular User - Forbidden", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/staff?roles=user", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected 403, got %d", resp.StatusCode)
		}
	})
}

package v1

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	"anirank/api/internal/testutil"
	"anirank/api/internal/usecase/auth"
	"anirank/api/internal/delivery/http/middleware"

	"github.com/gofiber/fiber/v2"
)

func TestAuthHandler_Profile(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test_secret_key_12345")
	jwtSvc := auth.NewJWTService()
	testUser := &domain.User{ID: 1, UUID: "test-uuid-handler", Name: "Test User", Email: "test@example.com"}
	mockRepo := &testutil.MockUserRepository{
		User: testUser,
		GetByIDFunc: func(id uint64) (*domain.User, error) {
			if id == 1 { return testUser, nil }
			return nil, domain.ErrNotFound
		},
	}

	// Initialize Usecase with mocks
	usecase := auth.NewAuthUsecase(
		mockRepo,
		jwtSvc,
		&testutil.MockStorageService{},
		&testutil.MockMediaService{},
		&testutil.MockXPUsecase{},
		&testutil.MockBadgeUsecase{},
		nil, nil, nil, "encryption_key",
	)
	handler := NewAuthHandler(usecase)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*domain.AppError); ok {
				return c.Status(e.Code).JSON(fiber.Map{"message": e.Message})
			}
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		},
	})
	// Apply Middleware + Handler
	app.Get("/profile", middleware.AuthMiddleware(jwtSvc, mockRepo), handler.Profile)

	t.Run("Get Profile - Authenticated", func(t *testing.T) {
		token, _ := testutil.CreateTestToken(testUser.UUID, []string{"user"})
		req := httptest.NewRequest("GET", "/profile", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200, got %d", resp.StatusCode)
		}

		var result struct {
			Data dto.UserDTO `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		if result.Data.ID != testUser.UUID {
			t.Errorf("Expected UUID %s, got %s", testUser.UUID, result.Data.ID)
		}
	})

	t.Run("Get Profile - Unauthorized (Missing Token)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/profile", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 401 {
			t.Errorf("Expected 401, got %d", resp.StatusCode)
		}
	})
}

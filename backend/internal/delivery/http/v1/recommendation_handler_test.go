package v1

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type MockRecommendationUsecase struct {
	GetSimilarSongsFunc                func(userID *uint64, songUUID string, limit int) ([]domain.Song, error)
	GetPersonalizedRecommendationsFunc func(userID uint64, limit int) ([]domain.Song, error)
}

func (m *MockRecommendationUsecase) GetSimilarSongs(ctx context.Context, userID *uint64, songUUID string, limit int) ([]domain.Song, error) {
	if m.GetSimilarSongsFunc != nil {
		return m.GetSimilarSongsFunc(userID, songUUID, limit)
	}
	return []domain.Song{}, nil
}

func (m *MockRecommendationUsecase) GetPersonalizedRecommendations(ctx context.Context, userID uint64, limit int) ([]domain.Song, error) {
	if m.GetPersonalizedRecommendationsFunc != nil {
		return m.GetPersonalizedRecommendationsFunc(userID, limit)
	}
	return []domain.Song{}, nil
}

func TestRecommendationHandler_GetSimilarSongs(t *testing.T) {
	mockUsecase := &MockRecommendationUsecase{}
	handler := NewRecommendationHandler(mockUsecase)

	app := fiber.New()
	app.Get("/api/v1/songs/:uuid/related", handler.GetSimilarSongs)

	t.Run("Successfully returns related songs", func(t *testing.T) {
		mockUsecase.GetSimilarSongsFunc = func(userID *uint64, songUUID string, limit int) ([]domain.Song, error) {
			assert.Equal(t, "test-song-uuid", songUUID)
			assert.Equal(t, 5, limit)
			assert.Nil(t, userID)
			return []domain.Song{
				{ID: 10, UUID: "similar-uuid-1", SongRomaji: ptr("Similar Romaji 1")},
			}, nil
		}

		req := httptest.NewRequest("GET", "/api/v1/songs/test-song-uuid/related?limit=5", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data []dto.SongSlimDTO `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Len(t, result.Data, 1)
		assert.Equal(t, "similar-uuid-1", result.Data[0].ID)
	})

	t.Run("Query parameters block bad limit values", func(t *testing.T) {
		mockUsecase.GetSimilarSongsFunc = func(userID *uint64, songUUID string, limit int) ([]domain.Song, error) {
			assert.Equal(t, 10, limit) // Negative parsed fallback
			return []domain.Song{}, nil
		}

		req := httptest.NewRequest("GET", "/api/v1/songs/test-song-uuid/related?limit=-100", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestRecommendationHandler_GetPersonalizedRecommendations(t *testing.T) {
	mockUsecase := &MockRecommendationUsecase{}
	handler := NewRecommendationHandler(mockUsecase)

	app := fiber.New()
	// Middleware a mano para inyectar user_id
	app.Get("/api/v1/recommendations", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint64(42))
		return c.Next()
	}, handler.GetPersonalizedRecommendations)

	t.Run("Successfully returns personalized recommendations", func(t *testing.T) {
		mockUsecase.GetPersonalizedRecommendationsFunc = func(userID uint64, limit int) ([]domain.Song, error) {
			assert.Equal(t, uint64(42), userID)
			assert.Equal(t, 10, limit) // Default limit
			return []domain.Song{
				{ID: 100, UUID: "recs-uuid-1", SongRomaji: ptr("Rec Romaji 1")},
			}, nil
		}

		req := httptest.NewRequest("GET", "/api/v1/recommendations", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data []dto.SongSlimDTO `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		assert.Len(t, result.Data, 1)
		assert.Equal(t, "recs-uuid-1", result.Data[0].ID)
	})
}

func ptr(s string) *string {
	return &s
}

package v1

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"
	"anirank/api/internal/usecase/public"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// MockSearchRepository implements domain.SearchRepository for SQLi endpoint testing.
type MockSearchRepository struct {
	LastQuery string
}

func (m *MockSearchRepository) GlobalSearch(ctx context.Context, term string, limit int) ([]domain.SearchIndexItem, error) {
	m.LastQuery = term
	return []domain.SearchIndexItem{}, nil
}

// MockSongRepositoryForSQLi embeds domain.SongRepository to implement all methods automatically.
type MockSongRepositoryForSQLi struct {
	domain.SongRepository
	LastFilters domain.SongFilters
}

func (m *MockSongRepositoryForSQLi) GetPaginated(ctx context.Context, limit, offset int, filters domain.SongFilters) ([]domain.Song, error) {
	m.LastFilters = filters
	return []domain.Song{}, nil
}

func (m *MockSongRepositoryForSQLi) Count(ctx context.Context, filters domain.SongFilters) (int, error) {
	return 0, nil
}

func (m *MockSongRepositoryForSQLi) GetByAnimeIDAndSlug(ctx context.Context, animeID uint64, slug string) (*domain.Song, error) {
	return nil, domain.ErrNotFound
}

// MockAnimeRepositoryForSQLi embeds domain.AnimeRepository
type MockAnimeRepositoryForSQLi struct {
	domain.AnimeRepository
}

func (m *MockAnimeRepositoryForSQLi) GetBySlug(ctx context.Context, slug string) (*domain.Anime, error) {
	return nil, domain.ErrNotFound
}

// MockUserRepositoryForSQLi embeds domain.UserRepository
type MockUserRepositoryForSQLi struct {
	domain.UserRepository
}

func (m *MockUserRepositoryForSQLi) GetByUUID(ctx context.Context, uuid string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}

// MockCache embeds domain.Cache
type MockCache struct {
	domain.Cache
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	return fmt.Errorf("cache miss")
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

// TestEndpoints_SQLi_Parameters verifies that sending various SQL injection payloads
// in query, route parameters, and header/body values does not crash the server (no 500 errors)
// and that mock layers receive the exact unmodified payload strings, proving no injection is interpreted.
func TestEndpoints_SQLi_Parameters(t *testing.T) {
	// Initialize Mocks
	searchRepo := &MockSearchRepository{}
	songRepo := &MockSongRepositoryForSQLi{}
	animeRepo := &MockAnimeRepositoryForSQLi{}
	userRepo := &MockUserRepositoryForSQLi{}
	mockCache := &MockCache{}

	mediaSvc := &testutil.MockMediaService{}

	// Setup Usecases
	searchUsecase := public.NewSearchUsecase(searchRepo, mediaSvc)
	catalogUsecase := public.NewCatalogUsecase(
		animeRepo,
		songRepo,
		nil, nil, // artistRepo, taxonomyRepo
		userRepo,
		nil, nil, nil, // playlistRepo, interactionRepo, moderationRepo
		nil,      // anilistClient
		mediaSvc,
		mockCache,
		"test_encryption_key",
	)

	// Setup Handlers
	searchHandler := NewSearchHandler(searchUsecase)
	catalogHandler := NewCatalogHandler(catalogUsecase)

	// Setup Fiber App
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*domain.AppError); ok {
				return c.Status(e.Code).JSON(fiber.Map{"message": e.Message})
			}
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		},
	})

	// Routes
	app.Get("/api/search", searchHandler.Search)
	app.Get("/api/songs", catalogHandler.SongIndex)
	app.Get("/api/songs/:anime_slug/:song_slug", catalogHandler.SongShow)
	app.Post("/api/users/favorites", catalogHandler.UserFavorites)

	for _, payload := range testutil.SQLIPayloads {
		t.Run("Endpoint_Search_Payload_"+payload, func(t *testing.T) {
			// Skip search terms shorter than 3 characters (as SearchUsecase rejects them with 400 Bad Request)
			if len(payload) < 3 {
				return
			}
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/search?q=%s", url.QueryEscape(payload)), nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			// Response must NOT be a 500 server error
			assert.NotEqual(t, 500, resp.StatusCode)
			assert.Contains(t, []int{200, 400}, resp.StatusCode) // Either normal safe query or validation block

			// If it reached the repo, it should have passed the exact unchanged payload
			if resp.StatusCode == 200 {
				assert.Equal(t, payload, searchRepo.LastQuery)
			}
		})

		t.Run("Endpoint_SongsQuery_Payload_"+payload, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/songs?name=%s", url.QueryEscape(payload)), nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			assert.NotEqual(t, 500, resp.StatusCode)
			assert.Equal(t, 200, resp.StatusCode)

			// Repo must receive the exact string (indicating no unescaped context shifts)
			assert.Equal(t, payload, songRepo.LastFilters.Search)
		})

		t.Run("Endpoint_SongsRoute_Payload_"+payload, func(t *testing.T) {
			// Route param injection
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/songs/anime-slug/%s", url.PathEscape(payload)), nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			assert.NotEqual(t, 500, resp.StatusCode)
			assert.Contains(t, []int{404, 400}, resp.StatusCode) // Not found (safe lookup block) or bad request
		})
	}
}

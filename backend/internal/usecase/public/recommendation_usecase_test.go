package public

import (
	"context"
	"fmt"
	"testing"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"

	"github.com/stretchr/testify/assert"
)

// MockRecommendationRepository stub for usecase testing
type MockRecommendationRepository struct {
	GetSimilarSongsByVectorFunc func(embedding domain.Vector, excludeSongID uint64, limit int) ([]domain.Song, error)
	UpdateSongEmbeddingFunc    func(songID uint64, embedding domain.Vector) error
	GetSongsWithoutEmbeddingsFunc func(limit int) ([]domain.Song, error)
	GetUserPreferencesVectorFunc func(userID uint64) (domain.Vector, error)
}

func (m *MockRecommendationRepository) GetSimilarSongsByVector(ctx context.Context, embedding domain.Vector, excludeSongID uint64, limit int) ([]domain.Song, error) {
	if m.GetSimilarSongsByVectorFunc != nil {
		return m.GetSimilarSongsByVectorFunc(embedding, excludeSongID, limit)
	}
	return []domain.Song{}, nil
}

func (m *MockRecommendationRepository) UpdateSongEmbedding(ctx context.Context, songID uint64, embedding domain.Vector) error {
	if m.UpdateSongEmbeddingFunc != nil {
		return m.UpdateSongEmbeddingFunc(songID, embedding)
	}
	return nil
}

func (m *MockRecommendationRepository) GetSongsWithoutEmbeddings(ctx context.Context, limit int) ([]domain.Song, error) {
	if m.GetSongsWithoutEmbeddingsFunc != nil {
		return m.GetSongsWithoutEmbeddingsFunc(limit)
	}
	return []domain.Song{}, nil
}

func (m *MockRecommendationRepository) GetUserPreferencesVector(ctx context.Context, userID uint64) (domain.Vector, error) {
	if m.GetUserPreferencesVectorFunc != nil {
		return m.GetUserPreferencesVectorFunc(userID)
	}
	return nil, nil
}

// MockSongRepositoryForUsecaseTest embeds domain.SongRepository to allow testing song methods
type MockSongRepositoryForUsecaseTest struct {
	domain.SongRepository
	GetByUUIDFunc    func(uuid string) (*domain.Song, error)
	GetByAnimeIDFunc func(animeID uint64, isAdmin bool) ([]domain.Song, error)
	GetPaginatedFunc func(limit, offset int, filters domain.SongFilters) ([]domain.Song, error)
	GetByIDFunc      func(id uint64) (*domain.Song, error)
}

func (m *MockSongRepositoryForUsecaseTest) GetByUUID(ctx context.Context, uuid string) (*domain.Song, error) {
	if m.GetByUUIDFunc != nil {
		return m.GetByUUIDFunc(uuid)
	}
	return nil, nil
}

func (m *MockSongRepositoryForUsecaseTest) GetByAnimeID(ctx context.Context, animeID uint64, isAdmin bool) ([]domain.Song, error) {
	if m.GetByAnimeIDFunc != nil {
		return m.GetByAnimeIDFunc(animeID, isAdmin)
	}
	return nil, nil
}

func (m *MockSongRepositoryForUsecaseTest) GetPaginated(ctx context.Context, limit, offset int, filters domain.SongFilters) ([]domain.Song, error) {
	if m.GetPaginatedFunc != nil {
		return m.GetPaginatedFunc(limit, offset, filters)
	}
	return nil, nil
}

func (m *MockSongRepositoryForUsecaseTest) GetByID(ctx context.Context, id uint64) (*domain.Song, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockSongRepositoryForUsecaseTest) GetArtistsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64][]domain.Artist, error) {
	return map[uint64][]domain.Artist{}, nil
}

func (m *MockSongRepositoryForUsecaseTest) GetAverageRatingsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64]float64, error) {
	return map[uint64]float64{}, nil
}

// MockAnimeRepository embeds domain.AnimeRepository
type MockAnimeRepository struct {
	domain.AnimeRepository
}

func (m *MockAnimeRepository) GetMany(ctx context.Context, ids []uint64) ([]domain.Anime, error) {
	return []domain.Anime{}, nil
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

func (m *MockCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *MockCache) IsAvailable() bool {
	return true
}

func TestRecommendationUsecase_GetSimilarSongs(t *testing.T) {
	mockRecRepo := &MockRecommendationRepository{}
	mockSongRepo := &MockSongRepositoryForUsecaseTest{}
	mockAnimeRepo := &MockAnimeRepository{}
	mockCache := &MockCache{}

	usecase := NewRecommendationUsecase(
		mockRecRepo,
		mockSongRepo,
		mockAnimeRepo,
		nil, // interactionRepo
		nil, // moderationRepo
		&testutil.MockMediaService{},
		mockCache,
	)

	t.Run("Related Songs Fallback when no Embedding present", func(t *testing.T) {
		songWithoutVector := &domain.Song{
			ID:      1,
			UUID:    "song-uuid",
			AnimeID: 10,
		}

		mockSongRepo.GetByUUIDFunc = func(uuid string) (*domain.Song, error) {
			if uuid == "song-uuid" {
				return songWithoutVector, nil
			}
			return nil, domain.ErrNotFound
		}

		mockSongRepo.GetByAnimeIDFunc = func(animeID uint64, isAdmin bool) ([]domain.Song, error) {
			assert.Equal(t, uint64(10), animeID)
			return []domain.Song{
				{ID: 1, UUID: "song-uuid"},
				{ID: 2, UUID: "other-song-uuid"},
			}, nil
		}

		songs, err := usecase.GetSimilarSongs(context.Background(), nil, "song-uuid", 5)
		assert.NoError(t, err)
		assert.Len(t, songs, 1)
		assert.Equal(t, uint64(2), songs[0].ID) // Base song ID 1 filtered out
	})

	t.Run("Related Songs using pgvector cosine", func(t *testing.T) {
		songWithVector := &domain.Song{
			ID:        1,
			UUID:      "song-uuid",
			Embedding: domain.Vector{0.1, 0.2, 0.3},
		}

		mockSongRepo.GetByUUIDFunc = func(uuid string) (*domain.Song, error) {
			if uuid == "song-uuid" {
				return songWithVector, nil
			}
			return nil, domain.ErrNotFound
		}

		mockRecRepo.GetSimilarSongsByVectorFunc = func(embedding domain.Vector, excludeSongID uint64, limit int) ([]domain.Song, error) {
			assert.Equal(t, domain.Vector{0.1, 0.2, 0.3}, embedding)
			assert.Equal(t, uint64(1), excludeSongID)
			assert.Equal(t, 5, limit)
			return []domain.Song{
				{ID: 3, UUID: "recommended-1"},
			}, nil
		}

		songs, err := usecase.GetSimilarSongs(context.Background(), nil, "song-uuid", 5)
		assert.NoError(t, err)
		assert.Len(t, songs, 1)
		assert.Equal(t, uint64(3), songs[0].ID)
	})
}

func TestRecommendationUsecase_GetPersonalizedRecommendations(t *testing.T) {
	mockRecRepo := &MockRecommendationRepository{}
	mockSongRepo := &MockSongRepositoryForUsecaseTest{}
	mockAnimeRepo := &MockAnimeRepository{}
	mockCache := &MockCache{}

	usecase := NewRecommendationUsecase(
		mockRecRepo,
		mockSongRepo,
		mockAnimeRepo,
		nil, nil,
		&testutil.MockMediaService{},
		mockCache,
	)

	t.Run("Cold Start Fallback to Global Trends when User has no Preferences", func(t *testing.T) {
		userID := uint64(77)

		// Simula que no hay preferencias
		mockRecRepo.GetUserPreferencesVectorFunc = func(uid uint64) (domain.Vector, error) {
			assert.Equal(t, userID, uid)
			return nil, nil
		}

		mockSongRepo.GetPaginatedFunc = func(limit, offset int, filters domain.SongFilters) ([]domain.Song, error) {
			assert.Equal(t, "favorites", filters.Sort)
			assert.Equal(t, 5, limit)
			return []domain.Song{
				{ID: 10, UUID: "trending-1"},
				{ID: 11, UUID: "trending-2"},
			}, nil
		}

		songs, err := usecase.GetPersonalizedRecommendations(context.Background(), userID, 5)
		assert.NoError(t, err)
		assert.Len(t, songs, 2)
		assert.Equal(t, uint64(10), songs[0].ID)
	})
}

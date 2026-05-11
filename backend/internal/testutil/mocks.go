package testutil

import (
	"context"
	"anirank/api/internal/domain"
)

type MockSongRepository struct {
	GetByIDFunc func(id uint64) (*domain.Song, error)
}

func (m *MockSongRepository) GetByID(ctx context.Context, id uint64) (*domain.Song, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}
func (m *MockSongRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Song, error) { return nil, nil }
func (m *MockSongRepository) GetBySlug(ctx context.Context, slug string) (*domain.Song, error) { return nil, nil }
func (m *MockSongRepository) GetByAnimeIDAndSlug(ctx context.Context, animeID uint64, slug string) (*domain.Song, error) { return nil, nil }
func (m *MockSongRepository) GetPaginated(ctx context.Context, limit, offset int, filters domain.SongFilters) ([]domain.Song, error) { return nil, nil }
func (m *MockSongRepository) GetByAnimeID(ctx context.Context, animeID uint64, isAdmin bool) ([]domain.Song, error) { return nil, nil }
func (m *MockSongRepository) GetByArtistID(ctx context.Context, artistID uint64, limit, offset int, filters domain.SongFilters) ([]domain.Song, error) { return nil, nil }
func (m *MockSongRepository) GetRanking(ctx context.Context, rankingType, songType string, limit, offset int) ([]domain.Song, error) { return nil, nil }
func (m *MockSongRepository) Count(ctx context.Context, filters domain.SongFilters) (int, error) { return 0, nil }
func (m *MockSongRepository) CountByArtistID(ctx context.Context, artistID uint64, filters domain.SongFilters) (int, error) { return 0, nil }
func (m *MockSongRepository) CountFavoritesByUserID(ctx context.Context, userID uint64) (int, error) { return 0, nil }
func (m *MockSongRepository) GetFavoritesByUserID(ctx context.Context, userID uint64, limit, offset int) ([]domain.Song, error) { return nil, nil }
func (m *MockSongRepository) CountRanking(ctx context.Context, rankingType, songType string) (int, error) { return 0, nil }
func (m *MockSongRepository) IncrementViews(ctx context.Context, id uint64) error { return nil }
func (m *MockSongRepository) GetMany(ctx context.Context, ids []uint64) ([]domain.Song, error) { return nil, nil }
func (m *MockSongRepository) Create(ctx context.Context, song *domain.Song) error { return nil }
func (m *MockSongRepository) Update(ctx context.Context, song *domain.Song) error { return nil }
func (m *MockSongRepository) Delete(ctx context.Context, id uint64) error { return nil }
func (m *MockSongRepository) GetVariantsBySongID(ctx context.Context, songID uint64) ([]domain.SongVariant, error) { return nil, nil }
func (m *MockSongRepository) GetVariantsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64][]domain.SongVariant, error) { return nil, nil }
func (m *MockSongRepository) GetArtistsBySongID(ctx context.Context, songID uint64, isAdmin bool) ([]domain.Artist, error) { return nil, nil }
func (m *MockSongRepository) GetArtistsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64][]domain.Artist, error) { return nil, nil }
func (m *MockSongRepository) SyncArtists(ctx context.Context, songID uint64, artistIDs []uint64) error { return nil }
func (m *MockSongRepository) ToggleStatus(ctx context.Context, id uint64) error { return nil }
func (m *MockSongRepository) GetPublicSlugs(ctx context.Context) ([]domain.SitemapItem, error) { return nil, nil }
func (m *MockSongRepository) GetSongTypes(ctx context.Context) ([]domain.SongType, error) { return nil, nil }

type MockCommentRepository struct {
	GetByIDFunc func(id uint64) (*domain.Comment, error)
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id uint64) (*domain.Comment, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}
func (m *MockCommentRepository) GetByEntity(ctx context.Context, userID *uint64, entityID uint64, entityType string, limit, offset int) ([]domain.Comment, error) { return nil, nil }
func (m *MockCommentRepository) GetReplies(ctx context.Context, userID *uint64, parentID uint64, limit, offset int) ([]domain.Comment, error) { return nil, nil }
func (m *MockCommentRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Comment, error) { return nil, nil }
func (m *MockCommentRepository) GetCountByEntity(ctx context.Context, entityID uint64, entityType string) (int, error) { return 0, nil }
func (m *MockCommentRepository) GetCountByUser(ctx context.Context, userID uint64) (int, error) { return 0, nil }
func (m *MockCommentRepository) GetRepliesCount(ctx context.Context, parentID uint64) (int, error) { return 0, nil }
func (m *MockCommentRepository) GetCount(ctx context.Context, songID uint64) (int, error) { return 0, nil }
func (m *MockCommentRepository) Create(ctx context.Context, comment *domain.Comment) error { return nil }
func (m *MockCommentRepository) Update(ctx context.Context, id uint64, content string) error { return nil }
func (m *MockCommentRepository) Delete(ctx context.Context, id, userID uint64) error { return nil }

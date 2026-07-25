package amq

import (
	"context"
	"fmt"
	"io"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/testutil"
)

type stubMediaService struct{}

func (stubMediaService) GetURL(path string) string { return path }
func (stubMediaService) Resolve(path *string) *string {
	return path
}
func (stubMediaService) GeneratePath(prefix string, id uint64, ext string) string { return "" }
func (stubMediaService) UploadImage(ctx context.Context, prefix string, id uint64, file io.Reader, size int64, contentType string) (string, string, error) {
	return "", "", nil
}
func (stubMediaService) UploadVideo(ctx context.Context, prefix string, id uint64, file io.Reader, size int64, contentType string, originalName string) (string, string, error) {
	return "", "", nil
}
func (stubMediaService) UploadImageOptimized(ctx context.Context, prefix string, id uint64, file io.Reader, options infrastructure.ImageOptions) (string, string, error) {
	return "", "", nil
}
func (stubMediaService) UploadWithResolutions(ctx context.Context, prefix string, id uint64, file io.Reader, preset infrastructure.ResolutionPreset) (string, string, error) {
	return "", "", nil
}
func (stubMediaService) GetImageSources(path string) []domain.ImageSource { return nil }
func (stubMediaService) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	return nil, nil
}
func (stubMediaService) DeleteMedia(ctx context.Context, path string) {}
func (stubMediaService) FileExists(ctx context.Context, path string) (bool, error) {
	return false, nil
}

type saveTestSongRepo struct {
	testutil.MockSongRepository
	artistAnchor *domain.AMQSaveThemeAnchor
	genreAnchor  *domain.AMQSaveThemeAnchor
	songIDs      []uint64
	roundType    string
}

func (r *saveTestSongRepo) FindRandomArtistForAMQSave(ctx context.Context, themeTypes []string, minSongs int) (*domain.AMQSaveThemeAnchor, error) {
	if r.artistAnchor != nil {
		return r.artistAnchor, nil
	}
	return &domain.AMQSaveThemeAnchor{Kind: "artist", ArtistUUID: "artist-uuid", ArtistName: "LiSA"}, nil
}

func (r *saveTestSongRepo) FindRandomYearForAMQSave(ctx context.Context, themeTypes []string, minSongs int) (*domain.AMQSaveThemeAnchor, error) {
	return nil, nil
}

func (r *saveTestSongRepo) FindRandomSeasonYearForAMQSave(ctx context.Context, themeTypes []string, minSongs int) (*domain.AMQSaveThemeAnchor, error) {
	return nil, nil
}

func (r *saveTestSongRepo) FindRandomAnimeForAMQSave(ctx context.Context, themeTypes []string, minThemes int) (*domain.AMQSaveThemeAnchor, error) {
	return nil, nil
}

func (r *saveTestSongRepo) FindRandomGenreForAMQSave(ctx context.Context, themeTypes []string, minSongs int) (*domain.AMQSaveThemeAnchor, error) {
	if r.genreAnchor != nil {
		return r.genreAnchor, nil
	}
	return nil, nil
}

func (r *saveTestSongRepo) GetRandomSongIDsForAMQSave(ctx context.Context, anchor domain.AMQSaveThemeAnchor, themeTypes []string, count int) ([]uint64, error) {
	if len(r.songIDs) >= count {
		return r.songIDs[:count], nil
	}
	ids := make([]uint64, count)
	for i := range ids {
		ids[i] = uint64(i + 1)
	}
	return ids, nil
}

func (r *saveTestSongRepo) GetMany(ctx context.Context, ids []uint64) ([]domain.Song, error) {
	songType := r.roundType
	if songType == "" {
		songType = "OP"
	}
	songs := make([]domain.Song, len(ids))
	for i, id := range ids {
		local := "/media/test.mp4"
		songs[i] = domain.Song{
			ID:       id,
			UUID:     fmt.Sprintf("song-uuid-%d", id),
			Type:     songType,
			ThemeNum: fmt.Sprintf("%d", i+1),
			Anime:    &domain.Anime{Title: "Test Anime", UUID: "anime-uuid"},
			Variants: []domain.SongVariant{{
				Video: &domain.SongVariantVideo{LocalUrl: &local},
			}},
		}
	}
	return songs, nil
}

func (r *saveTestSongRepo) GetVariantsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64][]domain.SongVariant, error) {
	m := make(map[uint64][]domain.SongVariant, len(songIDs))
	for _, id := range songIDs {
		local := "/media/test.mp4"
		m[id] = []domain.SongVariant{{
			Video: &domain.SongVariantVideo{LocalUrl: &local},
		}}
	}
	return m, nil
}

func (r *saveTestSongRepo) GetArtistsBySongIDs(ctx context.Context, songIDs []uint64) (map[uint64][]domain.Artist, error) {
	return map[uint64][]domain.Artist{}, nil
}

func (r *saveTestSongRepo) GetRandomSongsForAMQ(ctx context.Context, animeIDs []uint64, themeTypes []string, limit int, excludeIDs []uint64) ([]domain.Song, error) {
	return testSaveSongs(limit), nil
}

func newSaveTestRoom(repo *saveTestSongRepo) *LobbyRoom {
	if repo == nil {
		repo = &saveTestSongRepo{
			songIDs:   []uint64{1, 2, 3, 4, 5, 6, 7, 8},
			roundType: "OP",
		}
	}
	return NewLobbyRoom(
		"TESTROOM",
		domain.AMQConfig{
			GameType:       "save-4",
			MaxRounds:      5,
			PreviewSeconds: 12,
			VoteSeconds:    intPtr(defaultSaveVoteSeconds),
			ThemeType:      "OP",
		},
		nil,
		repo,
		nil,
		nil,
		stubMediaService{},
		nil,
	)
}

func testSaveSongs(n int) []domain.Song {
	songs := make([]domain.Song, n)
	for i := 0; i < n; i++ {
		local := "/media/test.mp4"
		songs[i] = domain.Song{
			ID:       uint64(i + 1),
			UUID:     fmt.Sprintf("song-uuid-%d", i+1),
			Type:     "OP",
			ThemeNum: fmt.Sprintf("%d", i+1),
			Anime:    &domain.Anime{Title: "Anime"},
			Variants: []domain.SongVariant{{
				Video: &domain.SongVariantVideo{LocalUrl: &local},
			}},
		}
	}
	return songs
}

func strPtr(s string) *string { return &s }

func intPtr(n int) *int { return &n }

var _ infrastructure.MediaService = stubMediaService{}

package playlist

import (
	"strings"
	"testing"

	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"
)

type resolvingMediaService struct {
	testutil.MockMediaService
}

func (m *resolvingMediaService) Resolve(path *string) *string {
	if path == nil || *path == "" {
		return nil
	}
	resolved := "https://media.example/" + strings.TrimPrefix(*path, "/")
	return &resolved
}

func TestResolvePlaylistVariantsUsesVideoSrcAndDropsUnplayableVariants(t *testing.T) {
	storageSrc := "videos/2025/winter/theme.webm"
	legacySrc := "https://youtube.com/embed/legacy"
	usecase := &PlaylistUsecase{mediaService: &resolvingMediaService{}}

	result := usecase.resolvePlaylistVariants([]domain.SongVariant{
		{
			Status: true,
			Videos: []domain.SongVariantVideo{{VideoSrc: &storageSrc}},
		},
		{
			Status: true,
			Videos: []domain.SongVariantVideo{{VideoSrc: &legacySrc}},
		},
	})

	if len(result) != 1 {
		t.Fatalf("expected one playable storage variant, got %d", len(result))
	}
	if result[0].Video == nil || result[0].Video.LocalUrl == nil {
		t.Fatal("expected resolved primary video")
	}
	if got := *result[0].Video.LocalUrl; got != "https://media.example/videos/2025/winter/theme.webm" {
		t.Fatalf("unexpected resolved URL: %s", got)
	}
}

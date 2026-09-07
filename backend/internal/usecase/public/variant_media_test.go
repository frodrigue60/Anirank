package public

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

func TestActiveVariantsForSongResolvesCanonicalVideoSrc(t *testing.T) {
	videoSrc := "videos/2026/spring/theme.webm"
	variants := []domain.SongVariant{{
		Status: true,
		Videos: []domain.SongVariantVideo{{
			VideoSrc: videoSrcPointer(videoSrc),
		}},
	}}

	result := activeVariantsForSong(variants, &resolvingMediaService{})
	if len(result) != 1 || len(result[0].Videos) != 1 {
		t.Fatalf("expected one playable variant, got %+v", result)
	}
	got := result[0].Videos[0].LocalUrl
	if got == nil || *got != "https://media.example/videos/2026/spring/theme.webm" {
		t.Fatalf("unexpected resolved URL: %v", got)
	}
}

func videoSrcPointer(value string) *string {
	return &value
}

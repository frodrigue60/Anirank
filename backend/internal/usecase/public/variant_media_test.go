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

func TestSetGeneratedPlaylistThumbnailUsesBannerAndCoverFallback(t *testing.T) {
	banner := "animes/banners/banner.avif"
	cover := "animes/covers/cover.avif"
	usecase := &CatalogUsecase{mediaService: &resolvingMediaService{}}

	withBanner := domain.GeneratedPlaylistDescriptor{}
	usecase.setGeneratedPlaylistThumbnail(&withBanner, []domain.Song{{
		Anime: &domain.Anime{Banner: &banner, Cover: &cover},
	}})
	if withBanner.BannerURL == nil || *withBanner.BannerURL != "https://media.example/animes/banners/banner.avif" {
		t.Fatalf("unexpected banner thumbnail: %v", withBanner.BannerURL)
	}

	withCover := domain.GeneratedPlaylistDescriptor{}
	usecase.setGeneratedPlaylistThumbnail(&withCover, []domain.Song{{
		Anime: &domain.Anime{Cover: &cover},
	}})
	if withCover.BannerURL == nil || *withCover.BannerURL != "https://media.example/animes/covers/cover.avif" {
		t.Fatalf("unexpected cover fallback: %v", withCover.BannerURL)
	}
}

func TestPersonalGeneratedPlaylistDescriptorSupportsFavorites(t *testing.T) {
	descriptor := personalGeneratedPlaylistDescriptor("favorited", 12)

	if descriptor.Kind != "favorited" {
		t.Fatalf("unexpected kind: %s", descriptor.Kind)
	}
	if descriptor.Href != "/playlists/generated/user/favorited" {
		t.Fatalf("unexpected href: %s", descriptor.Href)
	}
	if descriptor.Name != "Favorite Songs" || descriptor.SongCount != 12 {
		t.Fatalf("unexpected descriptor: %+v", descriptor)
	}
}

func videoSrcPointer(value string) *string {
	return &value
}

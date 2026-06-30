package admin

import (
	"testing"

	"anirank/api/internal/domain"
)

func strPtr(s string) *string { return &s }

func TestAnimeFullyEnrichedFromAniList(t *testing.T) {
	complete := &domain.Anime{
		Cover:        strPtr("animes/covers/1.webp"),
		Banner:       strPtr("animes/banners/1.webp"),
		TitleEnglish: strPtr("Title EN"),
		TitleNative:  strPtr("タイトル"),
		Description:  strPtr("Synopsis"),
	}
	if !animeFullyEnrichedFromAniList(complete) {
		t.Fatal("expected fully enriched anime")
	}

	missingBanner := &domain.Anime{
		Cover:        strPtr("animes/covers/1.webp"),
		TitleEnglish: strPtr("Title EN"),
		TitleNative:  strPtr("タイトル"),
		Description:  strPtr("Synopsis"),
	}
	if animeFullyEnrichedFromAniList(missingBanner) {
		t.Fatal("expected incomplete anime without banner")
	}
}

func TestHasStoredImage(t *testing.T) {
	if !hasStoredImage(strPtr(" path.webp ")) {
		t.Fatal("expected trimmed path to count as stored image")
	}
	if hasStoredImage(strPtr("   ")) {
		t.Fatal("expected blank path to be ignored")
	}
}

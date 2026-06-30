package admin

import (
	"testing"

	"anirank/api/internal/domain"
)

func TestAnimeThemesUpsertResultFlags(t *testing.T) {
	created := domain.AnimeThemesUpsertResult{Created: true}
	if !created.Created || created.MergedAnilist || created.DuplicateAnilist {
		t.Fatal("unexpected created result flags")
	}

	dup := domain.AnimeThemesUpsertResult{DuplicateAnilist: true}
	if dup.Created || dup.MergedAnilist || !dup.DuplicateAnilist {
		t.Fatal("unexpected duplicate result flags")
	}
}

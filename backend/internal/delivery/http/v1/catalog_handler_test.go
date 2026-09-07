package v1

import (
	"testing"

	"anirank/api/internal/domain"
	"anirank/api/internal/testutil"
)

func TestGeneratedPlaylistSongDTOsIncludeInteractionState(t *testing.T) {
	const internalID uint64 = 987654321
	songs := []domain.Song{{
		ID:            internalID,
		UUID:          "f7a356ca-1c25-4de8-9069-1796384746be",
		LikesCount:    14,
		DislikesCount: 3,
		IsLiked:       true,
		IsDisliked:    false,
		IsFavorited:   true,
	}}

	result := generatedPlaylistSongDTOs(songs)
	if len(result) != 1 {
		t.Fatalf("expected one song DTO, got %d", len(result))
	}
	if !result[0].IsLiked || result[0].IsDisliked || !result[0].IsFavorited {
		t.Fatalf("interaction state was not preserved: %+v", result[0])
	}
	if result[0].LikesCount != 14 || result[0].DislikesCount != 3 {
		t.Fatalf("interaction counters were not preserved: %+v", result[0])
	}
	testutil.AssertNoInternalIDs(t, result[0], internalID)
}

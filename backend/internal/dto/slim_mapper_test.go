package dto

import (
	"anirank/api/internal/domain"
	"testing"
)

func TestToSongSlimDTO_UserRating(t *testing.T) {
	rating := 8.5
	song := &domain.Song{
		UUID:          "song-uuid",
		Name:          "Test Song",
		AverageRating: 7.0,
		UserRating:    &rating,
	}

	dto := ToSongSlimDTO(song)

	if dto.UserRating == nil {
		t.Fatal("Expected UserRating to be populated, got nil")
	}

	if *dto.UserRating != rating {
		t.Errorf("Expected UserRating %f, got %f", rating, *dto.UserRating)
	}
}

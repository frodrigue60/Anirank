package postgres

import "testing"

func TestSanitizeAnimeSlug(t *testing.T) {
	tests := []struct {
		base string
		atID uint64
		want string
	}{
		{"Bocchi the Rock!", 1, "bocchi-the-rock"},
		{"", 42, "anime-42"},
		{"Season 2", 7, "season-2"},
		{"already-clean", 9, "already-clean"},
	}

	for _, tc := range tests {
		got := sanitizeAnimeSlug(tc.base, tc.atID)
		if got != tc.want {
			t.Fatalf("sanitizeAnimeSlug(%q, %d) = %q, want %q", tc.base, tc.atID, got, tc.want)
		}
	}
}

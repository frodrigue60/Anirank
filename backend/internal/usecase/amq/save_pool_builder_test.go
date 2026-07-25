package amq

import (
	"context"
	"math/rand"
	"testing"

	"anirank/api/internal/domain"
)

func TestIsSaveGameType(t *testing.T) {
	tests := []struct {
		gameType string
		want     bool
	}{
		{"save-4", true},
		{"save-6", true},
		{"type-in", false},
		{"multiple-choice", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsSaveGameType(tt.gameType); got != tt.want {
			t.Errorf("IsSaveGameType(%q) = %v, want %v", tt.gameType, got, tt.want)
		}
	}
}

func TestSaveOptionCount(t *testing.T) {
	if SaveOptionCount("save-6") != 6 {
		t.Fatalf("expected 6 for save-6")
	}
	if SaveOptionCount("save-4") != 4 {
		t.Fatalf("expected 4 for save-4")
	}
}

func TestRoundThemeTypesFilter(t *testing.T) {
	got := roundThemeTypesFilter("OP")
	if len(got) != 1 || got[0] != "OP" {
		t.Fatalf("expected [OP], got %v", got)
	}
}

func TestResolveRoundThemeTypeFixedWhenConfigured(t *testing.T) {
	if resolveRoundThemeType("OP") != "OP" {
		t.Fatal("expected OP when lobby theme is OP")
	}
	if resolveRoundThemeType("ED") != "ED" {
		t.Fatal("expected ED when lobby theme is ED")
	}
}

func TestSanitizeSaveConfigPreviewMax(t *testing.T) {
	cfg := domain.AMQConfig{GameType: "save-6", PreviewSeconds: 99, MaxRounds: 10}
	sanitizeSaveConfig(&cfg)
	if cfg.PreviewSeconds != 15 {
		t.Fatalf("preview seconds capped to 15, got %d", cfg.PreviewSeconds)
	}
}

func TestBuildSaveThemeLabelUsesSingleTypeNeverMixed(t *testing.T) {
	animeAnchor := domain.AMQSaveThemeAnchor{
		Kind:       "anime",
		AnimeTitle: "Kuroko no Basket 3rd SEASON",
	}

	opLabel := buildSaveThemeLabel(animeAnchor, "OP")
	if opLabel != "Best OP of Kuroko no Basket 3rd SEASON" {
		t.Fatalf("unexpected OP label: %q", opLabel)
	}
	if contains(opLabel, "OP/ED") {
		t.Fatalf("label must not contain OP/ED mix: %q", opLabel)
	}

	edLabel := buildSaveThemeLabel(animeAnchor, "ED")
	if edLabel != "Best ED of Kuroko no Basket 3rd SEASON" {
		t.Fatalf("unexpected ED label: %q", edLabel)
	}
}

func TestBuildSaveThemeKeyScopesByRoundType(t *testing.T) {
	anchor := domain.AMQSaveThemeAnchor{
		Kind:       "artist",
		ArtistUUID: "artist-uuid",
	}

	opKey := buildSaveThemeKey(anchor, "OP")
	edKey := buildSaveThemeKey(anchor, "ED")
	if opKey == edKey {
		t.Fatal("OP and ED keys must differ for same artist")
	}
	if opKey != "artist:artist-uuid:OP" {
		t.Fatalf("unexpected key: %q", opKey)
	}
}

func TestSanitizeSaveConfig(t *testing.T) {
	cfg := domain.AMQConfig{
		GameType:          "save-4",
		MaxRounds:         99,
		PreviewSeconds:    5,
		PersonalizedPool:  true,
		ThemeDistribution: "invalid",
	}
	sanitizeSaveConfig(&cfg)

	if cfg.MaxRounds != 30 {
		t.Fatalf("max rounds capped to 30, got %d", cfg.MaxRounds)
	}
	if cfg.PreviewSeconds != 12 {
		t.Fatalf("preview seconds min 12, got %d", cfg.PreviewSeconds)
	}
	if cfg.PersonalizedPool {
		t.Fatal("personalized pool must be disabled for save mode")
	}
	if cfg.ThemeDistribution != "random" {
		t.Fatalf("expected random distribution default, got %q", cfg.ThemeDistribution)
	}
	if saveVoteSecondsValue(cfg.VoteSeconds) != defaultSaveVoteSeconds {
		t.Fatalf("expected default vote seconds %d, got %d", defaultSaveVoteSeconds, saveVoteSecondsValue(cfg.VoteSeconds))
	}

	zero := 0
	cfg.VoteSeconds = &zero
	sanitizeSaveConfig(&cfg)
	if saveVoteSecondsValue(cfg.VoteSeconds) != 0 {
		t.Fatalf("expected instant vote (0), got %d", saveVoteSecondsValue(cfg.VoteSeconds))
	}

	over := 99
	cfg.VoteSeconds = &over
	sanitizeSaveConfig(&cfg)
	if saveVoteSecondsValue(cfg.VoteSeconds) != maxSaveVoteSeconds {
		t.Fatalf("expected vote seconds capped to %d, got %d", maxSaveVoteSeconds, saveVoteSecondsValue(cfg.VoteSeconds))
	}
}

func TestThemeKindsForRoundBalancedPrioritizesSlot(t *testing.T) {
	rand.Seed(1)
	kinds := themeKindsForRound(0, "balanced")
	if kinds[0] != "artist" {
		t.Fatalf("expected artist first for round 0, got %q", kinds[0])
	}
}

func TestBuildSaveRoundPoolDedupThemeKeys(t *testing.T) {
	repo := &saveTestSongRepo{
		artistAnchor: &domain.AMQSaveThemeAnchor{
			Kind:       "artist",
			ArtistUUID: "same-artist",
			ArtistName: "LiSA",
		},
		songIDs: []uint64{1, 2, 3, 4, 5, 6, 7, 8},
	}
	room := newSaveTestRoom(repo)
	room.Config.ThemeType = "OP"
	room.Config.MaxRounds = 3

	rounds, err := room.buildSaveRoundPool(context.Background(), 3, "save-4", "OP")
	if err != nil {
		t.Fatalf("buildSaveRoundPool: %v", err)
	}
	if len(rounds) != 3 {
		t.Fatalf("expected 3 rounds, got %d", len(rounds))
	}

	keys := make(map[string]bool)
	for _, r := range rounds {
		if r.ThemeKey == "" {
			t.Fatal("expected non-empty theme key")
		}
		if keys[r.ThemeKey] {
			t.Fatalf("duplicate theme key in same game: %s", r.ThemeKey)
		}
		keys[r.ThemeKey] = true
		if r.RoundThemeType != "OP" {
			t.Fatalf("round theme type must be OP, got %q", r.RoundThemeType)
		}
		if contains(r.ThemeLabel, "OP/ED") {
			t.Fatalf("label must not mix types: %q", r.ThemeLabel)
		}
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

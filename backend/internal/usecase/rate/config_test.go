package rate

import (
	"testing"

	"anirank/api/internal/domain"
)

func TestSanitizeConfig(t *testing.T) {
	cfg := domain.RateConfig{}
	sanitizeConfig(&cfg)

	if cfg.Name != "Rate Party" {
		t.Fatalf("expected default name, got %q", cfg.Name)
	}
	if cfg.QueueMode != QueueModeHostOnly {
		t.Fatalf("expected host_only, got %q", cfg.QueueMode)
	}
	if cfg.QueueLimitPerUser != DefaultQueueLimitPerUser {
		t.Fatalf("expected limit %d, got %d", DefaultQueueLimitPerUser, cfg.QueueLimitPerUser)
	}
	if cfg.RevealMode != RevealModeBlind {
		t.Fatalf("expected blind, got %q", cfg.RevealMode)
	}
	if cfg.MaxPlayers != DefaultMaxPlayers {
		t.Fatalf("expected max players %d, got %d", DefaultMaxPlayers, cfg.MaxPlayers)
	}
	if cfg.AutoAdvance != AutoAdvanceNever {
		t.Fatalf("expected never, got %q", cfg.AutoAdvance)
	}

	cfg2 := domain.RateConfig{
		Name:              "Custom",
		QueueMode:         QueueModeEveryone,
		QueueLimitPerUser: 99,
		RevealMode:        RevealModeLive,
		MaxPlayers:        8,
		AutoAdvance:       AutoAdvanceAllRated,
	}
	sanitizeConfig(&cfg2)
	if cfg2.QueueLimitPerUser != DefaultQueueLimitPerUser {
		t.Fatalf("expected clamp to default, got %d", cfg2.QueueLimitPerUser)
	}

	cfg3 := domain.RateConfig{QueueLimitPerUser: 5, QueueMode: QueueModeEveryone, RevealMode: RevealModeLive}
	sanitizeConfig(&cfg3)
	if cfg3.QueueLimitPerUser != 5 {
		t.Fatalf("expected 5, got %d", cfg3.QueueLimitPerUser)
	}
}

func TestSessionAvg(t *testing.T) {
	scores := []float64{80, 60, 70}
	var sum float64
	for _, s := range scores {
		sum += s
	}
	avg := sum / float64(len(scores))
	if avg != 70 {
		t.Fatalf("expected 70, got %v", avg)
	}
}

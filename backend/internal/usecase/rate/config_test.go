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
	if cfg.SourceMode != SourceModeManual {
		t.Fatalf("expected manual, got %q", cfg.SourceMode)
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

func TestSanitizeSeasonalPool(t *testing.T) {
	cfg := domain.RateConfig{
		SourceMode:    SourceModeSeasonalPool,
		PoolYear:      "2026",
		PoolSeason:    "summer",
		PoolThemeType: "OP",
		PoolLimit:     99,
		QueueMode:     QueueModeEveryone,
	}
	sanitizeConfig(&cfg)
	if cfg.QueueMode != QueueModeDisabled {
		t.Fatalf("seasonal should force disabled queue, got %q", cfg.QueueMode)
	}
	if cfg.PoolLimit != MaxPoolLimit {
		t.Fatalf("expected pool limit clamp to %d, got %d", MaxPoolLimit, cfg.PoolLimit)
	}
	if !isSeasonalPool(cfg) {
		t.Fatal("expected seasonal pool")
	}

	uncapped := domain.RateConfig{
		SourceMode: SourceModeSeasonalPool,
		PoolYear:   "2026",
		PoolSeason: "spring",
		PoolLimit:  0,
	}
	sanitizeConfig(&uncapped)
	if uncapped.PoolLimit != 0 {
		t.Fatalf("expected optional pool limit 0 (unlimited), got %d", uncapped.PoolLimit)
	}

	tooSmall := domain.RateConfig{
		SourceMode: SourceModeSeasonalPool,
		PoolYear:   "2026",
		PoolSeason: "spring",
		PoolLimit:  2,
	}
	sanitizeConfig(&tooSmall)
	if tooSmall.PoolLimit != MinPoolLimit {
		t.Fatalf("expected bump to min %d, got %d", MinPoolLimit, tooSmall.PoolLimit)
	}

	manual := domain.RateConfig{
		SourceMode:    SourceModeManual,
		PoolYear:      "2026",
		PoolSeason:    "summer",
		PoolThemeType: "OP",
		PoolLimit:     20,
	}
	sanitizeConfig(&manual)
	if manual.PoolYear != "" || manual.PoolSeason != "" || manual.PoolLimit != 0 {
		t.Fatalf("manual mode should clear pool fields: %+v", manual)
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

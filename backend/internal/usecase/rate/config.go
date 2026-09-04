package rate

import "anirank/api/internal/domain"

const (
	QueueModeHostOnly  = "host_only"
	QueueModeEveryone  = "everyone"
	QueueModeDisabled  = "disabled"
	RevealModeBlind    = "blind"
	RevealModeLive     = "live"
	AutoAdvanceNever   = "never"
	AutoAdvanceAllRated = "all_rated"

	DefaultQueueLimitPerUser = 3
	MinQueueLimitPerUser     = 1
	MaxQueueLimitPerUser     = 10
	DefaultMaxPlayers        = 16
	MaxQueueSize             = 50
)

func sanitizeConfig(cfg *domain.RateConfig) {
	if cfg.Name == "" {
		cfg.Name = "Rate Party"
	}
	switch cfg.QueueMode {
	case QueueModeHostOnly, QueueModeEveryone, QueueModeDisabled:
	default:
		cfg.QueueMode = QueueModeHostOnly
	}
	if cfg.QueueLimitPerUser < MinQueueLimitPerUser || cfg.QueueLimitPerUser > MaxQueueLimitPerUser {
		cfg.QueueLimitPerUser = DefaultQueueLimitPerUser
	}
	switch cfg.RevealMode {
	case RevealModeBlind, RevealModeLive:
	default:
		cfg.RevealMode = RevealModeBlind
	}
	if cfg.MaxPlayers < 2 || cfg.MaxPlayers > 32 {
		cfg.MaxPlayers = DefaultMaxPlayers
	}
	switch cfg.AutoAdvance {
	case AutoAdvanceNever, AutoAdvanceAllRated:
	default:
		cfg.AutoAdvance = AutoAdvanceNever
	}
}

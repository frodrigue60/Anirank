package rate

import "anirank/api/internal/domain"

const (
	QueueModeHostOnly   = "host_only"
	QueueModeEveryone   = "everyone"
	QueueModeDisabled   = "disabled"
	RevealModeBlind     = "blind"
	RevealModeLive      = "live"
	AutoAdvanceNever    = "never"
	AutoAdvanceAllRated = "all_rated"

	SourceModeManual       = "manual"
	SourceModeSeasonalPool = "seasonal_pool"

	PoolThemeAll = "all"
	PoolThemeOP  = "OP"
	PoolThemeED  = "ED"

	DefaultQueueLimitPerUser = 3
	MinQueueLimitPerUser     = 1
	MaxQueueLimitPerUser     = 10
	DefaultMaxPlayers        = 16
	MaxQueueSize             = 50
	// PoolLimit 0 = no cap (optional). When set, clamp to [MinPoolLimit, MaxPoolLimit].
	MinPoolLimit             = 5
	MaxPoolLimit             = 50
	// Safety ceiling when loading an uncapped seasonal pool from the catalog.
	UnlimitedPoolFetchLimit  = 500
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

	switch cfg.SourceMode {
	case SourceModeManual, SourceModeSeasonalPool:
	default:
		cfg.SourceMode = SourceModeManual
	}

	if cfg.SourceMode == SourceModeSeasonalPool {
		// Seasonal rooms are pool-driven — no open/host queue adds.
		cfg.QueueMode = QueueModeDisabled
		// 0 = unlimited (optional). Any other value must be within [Min, Max].
		if cfg.PoolLimit < 0 {
			cfg.PoolLimit = 0
		} else if cfg.PoolLimit > 0 && (cfg.PoolLimit < MinPoolLimit || cfg.PoolLimit > MaxPoolLimit) {
			if cfg.PoolLimit < MinPoolLimit {
				cfg.PoolLimit = MinPoolLimit
			} else {
				cfg.PoolLimit = MaxPoolLimit
			}
		}
		switch cfg.PoolThemeType {
		case PoolThemeAll, PoolThemeOP, PoolThemeED:
		default:
			cfg.PoolThemeType = PoolThemeAll
		}
	} else {
		cfg.PoolYear = ""
		cfg.PoolSeason = ""
		cfg.PoolThemeType = ""
		cfg.PoolLimit = 0
	}
}

func isSeasonalPool(cfg domain.RateConfig) bool {
	return cfg.SourceMode == SourceModeSeasonalPool
}

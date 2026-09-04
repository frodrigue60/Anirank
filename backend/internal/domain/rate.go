package domain

import "time"

// RateConfig is the lobby configuration for group rating sessions (/rate).
type RateConfig struct {
	Name               string `json:"name"`
	Private            bool   `json:"private"`
	QueueMode          string `json:"queue_mode"`            // "host_only" | "everyone" | "disabled"
	QueueLimitPerUser  int    `json:"queue_limit_per_user"`  // only for queue_mode=everyone; clamped 1–10
	RevealMode         string `json:"reveal_mode"`           // "blind" | "live"
	MaxPlayers         int    `json:"max_players"`
	AutoAdvance        string `json:"auto_advance"` // "never" | "all_rated"
}

// RatePlayer is a participant in a rate room. UserID is internal-only (never JSON).
type RatePlayer struct {
	SessionID    string     `json:"session_id"`
	UserUUID     string     `json:"user_uuid,omitempty"`
	UserID       uint64     `json:"-"`
	Nickname     string     `json:"nickname"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	ProfileColor *string    `json:"profile_color,omitempty"`
	DeviceID     string     `json:"device_id"`
	IsHost       bool       `json:"is_host"`
	IsSpectator  bool       `json:"is_spectator"`
	Offline      bool       `json:"offline"`
	OfflineSince *time.Time `json:"offline_since,omitempty"`
}

// RateQueueItem is a song waiting to be rated.
type RateQueueItem struct {
	ItemID           string `json:"item_id"`
	SongUUID         string `json:"song_uuid"`
	SongName         string `json:"song_name"`
	AnimeTitle       string `json:"anime_title,omitempty"`
	AnimeSlug        string `json:"anime_slug,omitempty"`
	ThemeLabel       string `json:"theme_label,omitempty"`
	CoverURL         string `json:"cover_url,omitempty"`
	AddedBySessionID string `json:"added_by_session_id"`
	AddedByUserUUID  string `json:"added_by_user_uuid,omitempty"`
	AddedByNickname  string `json:"added_by_nickname"`
}

// RateRoomInfo is the public lobby browser card.
type RateRoomInfo struct {
	RoomID         string `json:"room_id"`
	Name           string `json:"name"`
	HostNickname   string `json:"host_nickname"`
	PlayerCount    int    `json:"player_count"`
	SpectatorCount int    `json:"spectator_count"`
	Status         string `json:"status"` // lobby | waiting | rating | finished
	Private        bool   `json:"private"`
	QueueMode      string `json:"queue_mode"`
	RevealMode     string `json:"reveal_mode"`
	QueueLength    int    `json:"queue_length"`
}

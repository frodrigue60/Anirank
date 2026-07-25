package domain

import (
	"context"
	"time"
)

type AMQConfig struct {
	Name               string `json:"name"`
	MaxRounds          int    `json:"max_rounds"`
	GuessTime          int    `json:"guess_time"`    // in seconds (quiz modes)
	RevealTime         int    `json:"reveal_time"`   // in seconds (quiz modes)
	PreviewSeconds     int    `json:"preview_seconds"` // per-candidate preview/playback (save modes)
	VoteSeconds        *int   `json:"vote_seconds,omitempty"` // post-preview vote window (save modes); 0 = instant tally
	ThemeType          string `json:"theme_type"`  // "OP", "ED", "both"
	GameType           string `json:"game_type"`   // "type-in", "multiple-choice", "save-4", "save-6"
	ThemeDistribution  string `json:"theme_distribution"` // "random" or "balanced" (save modes)
	PersonalizedPool   bool   `json:"personalized_pool"`
	Private            bool   `json:"private"`
}

// AMQSaveRound is a pre-generated thematic round for save-4 / save-6 modes.
type AMQSaveRound struct {
	ThemeKind      string // artist, year, season, anime, genre, fallback
	ThemeKey       string // deduplication key within a game
	ThemeLabel     string
	RoundThemeType string // "OP" or "ED" — single type per round, never mixed
	IsFallback     bool
	Candidates     []Song
}

// AMQSaveRoundResult is persisted in-memory for the finished-screen history.
type AMQSaveRoundResult struct {
	RoundNumber    int            `json:"round_number"`
	ThemeLabel     string         `json:"theme_label"`
	RoundThemeType string         `json:"round_theme_type"`
	IsFallback     bool           `json:"is_fallback"`
	Winners        []string       `json:"winners"`
	VoteCounts     map[string]int `json:"vote_counts"`
	Candidates     []AMQSaveRoundCandidateSummary `json:"candidates"`
}

type AMQSaveRoundCandidateSummary struct {
	SongUUID   string `json:"song_uuid"`
	AnimeTitle string `json:"anime_title"`
	ThemeLabel string `json:"theme_label"`
	VoteCount  int    `json:"vote_count"`
	IsWinner   bool   `json:"is_winner"`
}

type AMQPlayer struct {
	SessionID        string     `json:"session_id"` // WebSocket session identifier
	UserUUID         string     `json:"user_uuid,omitempty"`
	Nickname         string     `json:"nickname"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	ProfileColor     *string    `json:"profile_color,omitempty"`
	DeviceID         string     `json:"device_id"` // For guest session persistence
	Score            int        `json:"score"`
	IsHost           bool       `json:"is_host"`
	IsReady          bool       `json:"is_ready"`
	LastGuess          string `json:"last_guess,omitempty"`
	LastGuessCorrect   bool   `json:"last_guess_correct"`
	SelectedSongUUID   string `json:"selected_song_uuid,omitempty"`
	IsSpectator      bool       `json:"is_spectator"`
	Locked           bool       `json:"locked"`
	Offline          bool       `json:"offline"`
	OfflineSince     *time.Time `json:"offline_since,omitempty"`
}

type AMQRoomInfo struct {
	RoomID         string `json:"room_id"`
	Name           string `json:"name"`
	HostNickname   string `json:"host_nickname"`
	PlayerCount    int    `json:"player_count"`
	SpectatorCount int    `json:"spectator_count"`
	MaxRounds      int    `json:"max_rounds"`
	Status         string `json:"status"` // "lobby", "playing", "reveal", "finished"
	Private        bool   `json:"private"`
	ThemeType         string `json:"theme_type"`
	GameType          string `json:"game_type"`
	PreviewSeconds    int    `json:"preview_seconds,omitempty"`
	VoteSeconds       int    `json:"vote_seconds,omitempty"`
	ThemeDistribution string `json:"theme_distribution,omitempty"`
}

type AMQUsecase interface {
	CreateRoom(ctx context.Context, config AMQConfig, hostUser *User, guestNickname, guestDeviceID string) (string, error)
	ListPublicRooms(ctx context.Context) ([]AMQRoomInfo, error)
}

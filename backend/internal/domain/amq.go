package domain

import (
	"context"
	"time"
)

type AMQConfig struct {
	Name             string `json:"name"`
	MaxRounds        int    `json:"max_rounds"`
	GuessTime        int    `json:"guess_time"`  // in seconds
	RevealTime       int    `json:"reveal_time"` // in seconds
	ThemeType        string `json:"theme_type"`  // "OP", "ED", "both"
	GameType         string `json:"game_type"`   // "type-in", "multiple-choice"
	PersonalizedPool bool   `json:"personalized_pool"`
	Private          bool   `json:"private"`
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
	LastGuess        string     `json:"last_guess,omitempty"`
	LastGuessCorrect bool       `json:"last_guess_correct"`
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
	ThemeType      string `json:"theme_type"`
	GameType       string `json:"game_type"`
}

type AMQUsecase interface {
	CreateRoom(ctx context.Context, config AMQConfig, hostUser *User, guestNickname, guestDeviceID string) (string, error)
	ListPublicRooms(ctx context.Context) ([]AMQRoomInfo, error)
}

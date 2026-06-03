package amq

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"

	"github.com/google/uuid"
)

type LobbyManager struct {
	rooms map[string]*LobbyRoom
	mu    sync.RWMutex

	// Dependencies
	animeRepo    domain.AnimeRepository
	songRepo     domain.SongRepository
	userRepo     domain.UserRepository
	xpUsecase    domain.XPUsecase
	mediaService infrastructure.MediaService
	anilist      anilist.AnilistClient
}

func NewLobbyManager(
	animeRepo domain.AnimeRepository,
	songRepo domain.SongRepository,
	userRepo domain.UserRepository,
	xpUsecase domain.XPUsecase,
	mediaService infrastructure.MediaService,
	anilistClient anilist.AnilistClient,
) *LobbyManager {
	lm := &LobbyManager{
		rooms:        make(map[string]*LobbyRoom),
		animeRepo:    animeRepo,
		songRepo:     songRepo,
		userRepo:     userRepo,
		xpUsecase:    xpUsecase,
		mediaService: mediaService,
		anilist:      anilistClient,
	}
	lm.StartCleanupLoop()
	return lm
}

// CreateRoom implements domain.AMQUsecase.
func (m *LobbyManager) CreateRoom(ctx context.Context, config domain.AMQConfig, hostUser *domain.User, guestNickname, guestDeviceID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Sanitize and validate config (Issue #11)
	if config.MaxRounds < 5 || config.MaxRounds > 50 {
		config.MaxRounds = 10
	}
	if config.GuessTime < 10 || config.GuessTime > 60 {
		config.GuessTime = 20
	}
	if config.RevealTime < 5 || config.RevealTime > 30 {
		config.RevealTime = 10
	}

	// Generate a unique 8-character uppercase RoomID
	var roomID string
	for {
		roomID = strings.ToUpper(uuid.New().String()[:8])
		if _, exists := m.rooms[roomID]; !exists {
			break
		}
	}

	room := NewLobbyRoom(
		roomID,
		config,
		m.animeRepo,
		m.songRepo,
		m.userRepo,
		m.xpUsecase,
		m.mediaService,
		m.anilist,
	)

	// Save in registry and start event loop
	m.rooms[roomID] = room
	room.Start()

	return roomID, nil
}

// ListPublicRooms implements domain.AMQUsecase.
func (m *LobbyManager) ListPublicRooms(ctx context.Context) ([]domain.AMQRoomInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []domain.AMQRoomInfo
	for _, room := range m.rooms {
		info := room.GetRoomInfo()
		if info.Private {
			continue
		}
		list = append(list, info)
	}

	if list == nil {
		list = []domain.AMQRoomInfo{}
	}

	return list, nil
}

// GetRoom retrieves a lobby room by ID
func (m *LobbyManager) GetRoom(roomID string) *LobbyRoom {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[roomID]
}

// JoinRoom upgrades a connection and assigns it to a lobby room
func (m *LobbyManager) JoinRoom(roomID, sessionID string, conn WSConn, user *domain.User, nickname, deviceID string, asSpectator bool) error {
	room := m.GetRoom(roomID)
	if room == nil {
		return errors.New("room not found")
	}

	// Dispatch Player Join event using SendEvent helper
	room.SendEvent(RoomEvent{
		Type: EvJoin,
		Data: &JoinEvent{
			SessionID:   sessionID,
			Conn:        conn,
			User:        user,
			Nickname:    nickname,
			DeviceID:    deviceID,
			AsSpectator: asSpectator,
		},
	})

	return nil
}

// LeaveRoom handles websocket drops or leave clicks
func (m *LobbyManager) LeaveRoom(roomID, sessionID string) {
	room := m.GetRoom(roomID)
	if room == nil {
		return
	}

	// Dispatch Player Leave event using SendEvent helper
	room.SendEvent(RoomEvent{
		Type: EvLeave,
		Data: sessionID,
	})
}

func (m *LobbyManager) StartCleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			m.cleanupRooms()
		}
	}()
}

func (m *LobbyManager) cleanupRooms() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for roomID, room := range m.rooms {
		if room.ShouldDestroy() {
			log.Printf("[AMQ] Cleaning up inactive/empty room %s", roomID)
			room.Close()
			delete(m.rooms, roomID)
		}
	}
}

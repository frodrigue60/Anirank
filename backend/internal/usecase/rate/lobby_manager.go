package rate

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"

	"github.com/google/uuid"
)

// SongRater persists ratings to the global ranking (always-on for /rate).
type SongRater interface {
	RateSongInLiveRoom(ctx context.Context, userID, songID uint64, score float64) (float64, float64, error)
	GetUserSongRating(ctx context.Context, userID, songID uint64) (*domain.Rating, error)
}

type LobbyManager struct {
	rooms map[string]*LobbyRoom
	mu    sync.RWMutex

	songRepo     domain.SongRepository
	userRepo     domain.UserRepository
	mediaService infrastructure.MediaService
	songRater    SongRater
}

func NewLobbyManager(
	songRepo domain.SongRepository,
	userRepo domain.UserRepository,
	mediaService infrastructure.MediaService,
	songRater SongRater,
) *LobbyManager {
	lm := &LobbyManager{
		rooms:        make(map[string]*LobbyRoom),
		songRepo:     songRepo,
		userRepo:     userRepo,
		mediaService: mediaService,
		songRater:    songRater,
	}
	lm.StartCleanupLoop()
	return lm
}

func (m *LobbyManager) CreateRoom(ctx context.Context, config domain.RateConfig, hostUser *domain.User, guestNickname, guestDeviceID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sanitizeConfig(&config)
	if config.SourceMode == SourceModeSeasonalPool && (config.PoolYear == "" || config.PoolSeason == "") {
		return "", errors.New("seasonal pool requires year and season")
	}

	var roomID string
	for {
		roomID = strings.ToUpper(uuid.New().String()[:8])
		if _, exists := m.rooms[roomID]; !exists {
			break
		}
	}

	room := NewLobbyRoom(roomID, config, m.songRepo, m.userRepo, m.mediaService, m.songRater)
	m.rooms[roomID] = room
	room.OnDestroy = func(rid string) {
		m.mu.Lock()
		delete(m.rooms, rid)
		m.mu.Unlock()
		log.Printf("[RATE] Room %s removed from registry", rid)
	}
	room.Start()

	return roomID, nil
}

func (m *LobbyManager) ListPublicRooms(ctx context.Context) ([]domain.RateRoomInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]domain.RateRoomInfo, 0)
	for _, room := range m.rooms {
		info := room.GetRoomInfo()
		if info.Private {
			continue
		}
		list = append(list, info)
	}
	return list, nil
}

func (m *LobbyManager) GetRoom(roomID string) *LobbyRoom {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[roomID]
}

func (m *LobbyManager) JoinRoom(roomID, sessionID string, conn WSConn, user *domain.User, nickname, deviceID string, asSpectator bool) error {
	room := m.GetRoom(roomID)
	if room == nil {
		return errors.New("room not found")
	}
	done := make(chan struct{})
	room.SendEvent(RoomEvent{
		Type: EvJoin,
		Data: &JoinEvent{
			SessionID:   sessionID,
			Conn:        conn,
			User:        user,
			Nickname:    nickname,
			DeviceID:    deviceID,
			AsSpectator: asSpectator,
			Done:        done,
		},
	})
	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("join timed out")
	}
}

func (m *LobbyManager) LeaveRoom(roomID, sessionID string) {
	room := m.GetRoom(roomID)
	if room == nil {
		return
	}
	room.SendEvent(RoomEvent{Type: EvLeave, Data: sessionID})
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
	toClose := make([]*LobbyRoom, 0)
	for roomID, room := range m.rooms {
		if room.ShouldDestroy() {
			log.Printf("[RATE] Cleaning up inactive/empty room %s", roomID)
			toClose = append(toClose, room)
			// Remove under the manager lock first. Close() invokes OnDestroy, which
			// also takes m.mu — calling it while holding the lock deadlocks ListRooms.
			delete(m.rooms, roomID)
		}
	}
	m.mu.Unlock()

	for _, room := range toClose {
		room.OnDestroy = nil
		room.Close()
	}
}

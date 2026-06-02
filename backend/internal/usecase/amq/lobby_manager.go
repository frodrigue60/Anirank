package amq

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"

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
	return &LobbyManager{
		rooms:        make(map[string]*LobbyRoom),
		animeRepo:    animeRepo,
		songRepo:     songRepo,
		userRepo:     userRepo,
		xpUsecase:    xpUsecase,
		mediaService: mediaService,
		anilist:      anilistClient,
	}
}

// CreateRoom implements domain.AMQUsecase.
func (m *LobbyManager) CreateRoom(ctx context.Context, config domain.AMQConfig, hostUser *domain.User, guestNickname, guestDeviceID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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
		if room.Config.Private {
			continue
		}

		// Find host nickname
		hostNick := "Unknown"
		for _, p := range room.Players {
			if p.IsHost {
				hostNick = p.Nickname
				break
			}
		}

		list = append(list, domain.AMQRoomInfo{
			RoomID:       room.RoomID,
			Name:         room.Config.Name,
			HostNickname: hostNick,
			PlayerCount:  len(room.Players),
			MaxRounds:    room.Config.MaxRounds,
			Status:       room.Status,
			Private:      room.Config.Private,
			ThemeType:    room.Config.ThemeType,
			GameType:     room.Config.GameType,
		})
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
func (m *LobbyManager) JoinRoom(roomID, sessionID string, conn WSConn, user *domain.User, nickname, deviceID string) error {
	room := m.GetRoom(roomID)
	if room == nil {
		return errors.New("room not found")
	}

	// Dispatch Player Join event to the room goroutine
	room.EventChan <- RoomEvent{
		Type: EvJoin,
		Data: &JoinEvent{
			SessionID: sessionID,
			Conn:      conn,
			User:      user,
			Nickname:  nickname,
			DeviceID:  deviceID,
		},
	}

	return nil
}

// LeaveRoom handles websocket drops or leave clicks
func (m *LobbyManager) LeaveRoom(roomID, sessionID string) {
	room := m.GetRoom(roomID)
	if room == nil {
		return
	}

	room.EventChan <- RoomEvent{
		Type: EvLeave,
		Data: sessionID,
	}

	// Clean up empty rooms
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// If all players are offline/purged, we can remove the room
	allOffline := true
	if len(room.Players) == 0 {
		allOffline = true
	} else {
		for _, p := range room.Players {
			if !p.Offline {
				allOffline = false
				break
			}
		}
	}

	if allOffline {
		log.Printf("[AMQ] Tearing down empty room %s", roomID)
		close(room.EventChan) // Stop event loop
		delete(m.rooms, roomID)
	}
}

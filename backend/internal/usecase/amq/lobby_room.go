package amq

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
)

type WSConn interface {
	WriteJSON(v interface{}) error
	Close() error
}

type RoomEventType int

const (
	EvJoin RoomEventType = iota
	EvLeave
	EvReady
	EvConfigUpdate
	EvStartGame
	EvSubmitGuess
	EvTimerExpired
	EvSkipSummary
	EvResetToLobby
	EvPoolLoaded
	EvChat
	EvTransferHost
	EvCloseRoom
)

type RoomEvent struct {
	Type RoomEventType
	Data interface{}
}

type TransferHostEvent struct {
	SessionID       string
	TargetSessionID string
}

type JoinEvent struct {
	SessionID   string
	Conn        WSConn
	User        *domain.User
	Nickname    string
	DeviceID    string
	AsSpectator bool
}

type GuessEvent struct {
	SessionID string
	AnimeSlug string
}

type ConfigUpdateEvent struct {
	SessionID string
	Config    domain.AMQConfig
}

type ChatEvent struct {
	SessionID string
	Text      string
}

type LobbyRoom struct {
	RoomID        string
	Config        domain.AMQConfig
	Status        string // "lobby", "playing", "reveal", "finished"
	CurrentRound  int
	Players       map[string]*domain.AMQPlayer // SessionID -> Player
	Conns         map[string]WSConn            // SessionID -> WS connection
	CurrentSong   *domain.Song
	CurrentFakes  []domain.Anime
	CurrentOptions []dto.AnimeMinimalDTO
	SongPool      []domain.Song
	EventChan     chan RoomEvent
	Timer         *time.Timer
	TimerType     string // "guess", "reveal"
	TimerStart    time.Time
	TimerDuration time.Duration
	StartPercent  float64
	CreatedAt     time.Time
	LastActive    time.Time
	Closed        bool

	// Dependencies
	AnimeRepo    domain.AnimeRepository
	SongRepo     domain.SongRepository
	UserRepo     domain.UserRepository
	XPUsecase    domain.XPUsecase
	MediaService infrastructure.MediaService
	Anilist      anilist.AnilistClient

	mu sync.RWMutex
}

func (r *LobbyRoom) SendEvent(ev RoomEvent) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[AMQ] Recovered from panic sending event to room %s: %v", r.RoomID, err)
		}
	}()
	r.EventChan <- ev
}


func NewLobbyRoom(
	roomID string,
	config domain.AMQConfig,
	animeRepo domain.AnimeRepository,
	songRepo domain.SongRepository,
	userRepo domain.UserRepository,
	xpUsecase domain.XPUsecase,
	mediaService infrastructure.MediaService,
	anilistClient anilist.AnilistClient,
) *LobbyRoom {
	now := time.Now()
	return &LobbyRoom{
		RoomID:       roomID,
		Config:       config,
		Status:       "lobby",
		CurrentRound: 0,
		Players:      make(map[string]*domain.AMQPlayer),
		Conns:        make(map[string]WSConn),
		EventChan:    make(chan RoomEvent, 100),
		CreatedAt:    now,
		LastActive:   now,
		AnimeRepo:    animeRepo,
		SongRepo:     songRepo,
		UserRepo:     userRepo,
		XPUsecase:    xpUsecase,
		MediaService: mediaService,
		Anilist:      anilistClient,
	}
}

func (r *LobbyRoom) Start() {
	go r.run()
}

func (r *LobbyRoom) run() {
	// Periodic ticker to clean up stale offline players (e.g., offline for > 60s)
	cleanupTicker := time.NewTicker(10 * time.Second)
	defer cleanupTicker.Stop()

	for {
		stopped := false
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("[AMQ] PANIC recovered in room %s event loop: %v", r.RoomID, rec)
				}
			}()
			select {
			case ev, ok := <-r.EventChan:
				if !ok {
					stopped = true
					return
				}
				r.handleEvent(ev)
				r.mu.Lock()
				r.LastActive = time.Now()
				r.mu.Unlock()
			case <-cleanupTicker.C:
				r.cleanupOfflinePlayers()
			}
		}()
		if stopped {
			return
		}
	}
}

func (r *LobbyRoom) handleEvent(ev RoomEvent) {
	switch ev.Type {
	case EvJoin:
		r.handleJoin(ev.Data.(*JoinEvent))
	case EvLeave:
		r.handleLeave(ev.Data.(string))
	case EvReady:
		r.handleReady(ev.Data.(string))
	case EvConfigUpdate:
		r.handleConfigUpdate(ev.Data.(*ConfigUpdateEvent))
	case EvStartGame:
		r.handleStartGame(ev.Data.(string))
	case EvPoolLoaded:
		r.startRound()
	case EvSubmitGuess:
		r.handleSubmitGuess(ev.Data.(*GuessEvent))
	case EvTimerExpired:
		r.handleTimerExpired(ev.Data.(string))
	case EvSkipSummary:
		r.handleSkipSummary(ev.Data.(string))
	case EvResetToLobby:
		dataStr := ev.Data.(string)
		r.mu.Lock()
		player, exists := r.Players[dataStr]
		isHost := exists && player.IsHost
		r.mu.Unlock()
		if isHost {
			r.handleResetToLobby(dataStr)
		} else {
			r.forceResetToLobby(dataStr)
		}
	case EvChat:
		r.handleChat(ev.Data.(*ChatEvent))
	case EvTransferHost:
		r.handleTransferHost(ev.Data.(*TransferHostEvent))
	case EvCloseRoom:
		r.handleCloseRoom(ev.Data.(string))
	}
}

// WS messages structures
type outMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

func (r *LobbyRoom) broadcast(msgType string, payload interface{}) {
	msg := outMessage{Type: msgType, Payload: payload}
	// Use RLock: broadcast only reads r.Conns, never modifies room state.
	// Using a write lock here caused contention with external goroutines holding RLocks
	// (e.g., LobbyManager.cleanupRooms → ShouldDestroy → RLock), which could deadlock
	// the event loop goroutine when a write lock was pending.
	r.mu.RLock()
	conns := make(map[string]WSConn, len(r.Conns))
	for sid, conn := range r.Conns {
		conns[sid] = conn
	}
	r.mu.RUnlock()

	for sid, conn := range conns {
		if conn == nil {
			continue
		}
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("[AMQ] Error writing to session %s: %v", sid, err)
		}
	}
}

func (r *LobbyRoom) sendTo(sessionID string, msgType string, payload interface{}) {
	msg := outMessage{Type: msgType, Payload: payload}
	// Use RLock: sendTo only reads r.Conns.
	r.mu.RLock()
	conn, ok := r.Conns[sessionID]
	r.mu.RUnlock()

	if ok && conn != nil {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("[AMQ] Error in sendTo session %s: %v", sessionID, err)
		}
	}
}

func (r *LobbyRoom) handleJoin(ev *JoinEvent) {
	r.mu.Lock()
	// Check if this guest already exists in the room by DeviceID
	var existingPlayer *domain.AMQPlayer
	var oldSessionID string

	if ev.User == nil && ev.DeviceID != "" {
		for sid, p := range r.Players {
			if p.UserUUID == "" && p.DeviceID == ev.DeviceID {
				existingPlayer = p
				oldSessionID = sid
				break
			}
		}
	} else if ev.User != nil {
		// Authenticated user: check if they are already in the room
		for sid, p := range r.Players {
			if p.UserUUID == ev.User.UUID {
				existingPlayer = p
				oldSessionID = sid
				break
			}
		}
	}

	if existingPlayer == nil && ev.User == nil {
		// Fallback for guests: match by Nickname (prefer offline, but fallback to online to prevent duplication)
		var match *domain.AMQPlayer
		var matchSid string
		for sid, p := range r.Players {
			if p.UserUUID == "" && p.Nickname == ev.Nickname {
				if p.Offline {
					match = p
					matchSid = sid
					break
				} else {
					match = p
					matchSid = sid
				}
			}
		}
		if match != nil {
			existingPlayer = match
			oldSessionID = matchSid
		}
	}

	if existingPlayer != nil {
		// Reconnection: transfer state to new SessionID
		log.Printf("[AMQ] Reassociating player %s (old session %s, new session %s)", existingPlayer.Nickname, oldSessionID, ev.SessionID)

		// Do NOT call oldConn.Close() here. A forced close from the server side
		// causes the client's WebSocket onclose event to fire with wasClean=false,
		// which incorrectly triggers the "Connection lost. Reconnecting..." UI message.
		// Instead, we just remove the old entry from the maps. The old goroutine's
		// ReadJSON call will fail on its next iteration (the connection is now orphaned)
		// and its deferred LeaveRoom call will be a no-op since the session ID is gone.
		delete(r.Conns, oldSessionID)
		delete(r.Players, oldSessionID)

		// Set new session info
		existingPlayer.SessionID = ev.SessionID
		existingPlayer.Offline = false
		existingPlayer.OfflineSince = nil
		existingPlayer.Nickname = ev.Nickname
		if ev.User != nil && ev.User.AvatarUrl != nil {
			avatar := *ev.User.AvatarUrl
			existingPlayer.AvatarURL = &avatar
		}
		if ev.User != nil && ev.User.ProfileColor != nil {
			color := *ev.User.ProfileColor
			existingPlayer.ProfileColor = &color
		}

		r.Players[ev.SessionID] = existingPlayer
		r.Conns[ev.SessionID] = ev.Conn
	} else {
		// New player
		nonSpectators := 0
		for _, p := range r.Players {
			if !p.IsSpectator {
				nonSpectators++
			}
		}
		isHost := !ev.AsSpectator && nonSpectators == 0
		avatar := ""
		if ev.User != nil && ev.User.AvatarUrl != nil {
			avatar = *ev.User.AvatarUrl
		}
		color := "#683bc9" // Default purple brand color
		if ev.User != nil && ev.User.ProfileColor != nil {
			color = *ev.User.ProfileColor
		}

		userUUID := ""
		if ev.User != nil {
			userUUID = ev.User.UUID
		}

		player := &domain.AMQPlayer{
			SessionID:    ev.SessionID,
			UserUUID:     userUUID,
			Nickname:     ev.Nickname,
			AvatarURL:    &avatar,
			ProfileColor: &color,
			DeviceID:     ev.DeviceID,
			IsHost:       isHost,
			IsReady:      isHost, // Host is ready by default
			IsSpectator:  ev.AsSpectator,
		}

		r.Players[ev.SessionID] = player
		r.Conns[ev.SessionID] = ev.Conn
	}
	r.ensureHostActive()
	r.mu.Unlock()

	// Send current state to the joined user
	r.sendTo(ev.SessionID, "lobby_state_update", r.getRoomStatePayload())

	// Broadcast updated lobby state to all players
	r.broadcast("lobby_state_update", r.getRoomStatePayload())

	role := "player"
	if ev.AsSpectator {
		role = "spectator"
	}
	r.broadcast("chat_message", map[string]interface{}{
		"sender":    "System",
		"text":      fmt.Sprintf("%s joined the room as %s", ev.Nickname, role),
		"type":      "system",
		"timestamp": time.Now(),
	})
}

func (r *LobbyRoom) handleLeave(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists {
		r.mu.Unlock()
		return
	}

	// For guest persistence or authenticated user transient disconnects,
	// mark them offline rather than immediately deleting them
	player.Offline = true
	now := time.Now()
	player.OfflineSince = &now
	r.Conns[sessionID] = nil

	log.Printf("[AMQ] Player %s went offline", player.Nickname)

	r.ensureHostActive()
	r.mu.Unlock()

	// Check if all online players have locked answers (in case this was the last person keeping the timer going)
	if r.Status == "playing" {
		r.checkAllLockedAndProceed()
	}

	r.broadcast("lobby_state_update", r.getRoomStatePayload())

	r.broadcast("chat_message", map[string]interface{}{
		"sender":    "System",
		"text":      fmt.Sprintf("%s went offline", player.Nickname),
		"type":      "system",
		"timestamp": time.Now(),
	})
}

func (r *LobbyRoom) handleReady(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || player.IsSpectator || r.Status != "lobby" {
		r.mu.Unlock()
		return
	}
	// Toggle ready state
	player.IsReady = !player.IsReady
	r.mu.Unlock()
	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) handleConfigUpdate(ev *ConfigUpdateEvent) {
	r.mu.Lock()
	player, exists := r.Players[ev.SessionID]
	if !exists || !player.IsHost || r.Status != "lobby" {
		r.mu.Unlock()
		return
	}
	
	// Sanitize config
	cfg := ev.Config
	if cfg.MaxRounds < 5 || cfg.MaxRounds > 50 {
		cfg.MaxRounds = 10
	}
	if cfg.GuessTime < 10 || cfg.GuessTime > 60 {
		cfg.GuessTime = 20
	}
	if cfg.RevealTime < 5 || cfg.RevealTime > 30 {
		cfg.RevealTime = 10
	}

	r.Config = cfg
	r.mu.Unlock()

	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) handleStartGame(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || !player.IsHost || r.Status != "lobby" {
		r.mu.Unlock()
		return
	}

	// Verify all active players are ready
	for _, p := range r.Players {
		if !p.Offline && !p.IsSpectator && !p.IsReady {
			r.mu.Unlock()
			r.sendTo(sessionID, "error", "Cannot start. All active players must toggle Ready.")
			return
		}
	}

	// 1. Prepare Song Pool
	r.Status = "playing"
	r.CurrentRound = 0

	// Extract variables to prevent data races inside async goroutine
	personalizedPool := r.Config.PersonalizedPool
	maxRounds := r.Config.MaxRounds
	themeTypeConfig := r.Config.ThemeType
	var playerUserUUIDs []string
	for _, p := range r.Players {
		if p.UserUUID != "" && !p.Offline && !p.IsSpectator {
			playerUserUUIDs = append(playerUserUUIDs, p.UserUUID)
		}
	}
	r.mu.Unlock()

	r.broadcast("lobby_state_update", r.getRoomStatePayload())

	go func() {
		anilistCtx, anilistCancel := context.WithTimeout(context.Background(), 8*time.Second)
		dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			anilistCancel()
			dbCancel()
			if err := recover(); err != nil {
				log.Printf("[AMQ] StartGame crashed: %v", err)
				r.SendEvent(RoomEvent{Type: EvResetToLobby, Data: "Game could not start due to an internal error."})
			}
		}()

		// Gather linked AniList IDs
		var watchedAnimeIDs []uint64
		if personalizedPool {
			// Find AniList IDs from authenticated players
			var linkedUserIDs []uint64
			for _, uuidVal := range playerUserUUIDs {
				user, err := r.UserRepo.GetByUUID(dbCtx, uuidVal)
				if err == nil {
					linkedUserIDs = append(linkedUserIDs, user.ID)
				}
			}

			if len(linkedUserIDs) > 0 {
				// Query public AniList identities from DB
				var anilistUserIDs []int64
				for _, uid := range linkedUserIDs {
					sids, err := r.UserRepo.GetSocialIdentitiesByUserID(dbCtx, uid)
					if err == nil {
						for _, si := range sids {
							if si.Provider == "anilist" {
								var aid int64
								if _, errScan := fmt.Sscanf(si.ProviderID, "%d", &aid); errScan == nil {
									anilistUserIDs = append(anilistUserIDs, aid)
								}
							}
						}
					}
				}

				// If we have AniList IDs, query their completed list items from GraphQL AniList API
				var intersectedAnimes = make(map[int]bool)
				firstUser := true

				for _, aid := range anilistUserIDs {
					userAnilistMedia := make(map[int]bool)
					page := 1
					for {
						resp, err := r.Anilist.GetUserMediaList(anilistCtx, aid, "COMPLETED", page, 500)
						if err != nil || resp == nil || len(resp.Data.Page.MediaList) == 0 {
							break
						}
						for _, item := range resp.Data.Page.MediaList {
							userAnilistMedia[item.Media.ID] = true
						}
						if !resp.Data.Page.PageInfo.HasNextPage {
							break
						}
						page++
					}

					// Perform intersection
					if firstUser {
						for mid := range userAnilistMedia {
							intersectedAnimes[mid] = true
						}
						firstUser = false
					} else {
						for mid := range intersectedAnimes {
							if !userAnilistMedia[mid] {
								delete(intersectedAnimes, mid)
							}
						}
					}
				}

				// Convert intersected AniList IDs to DB Anime IDs
				if len(intersectedAnimes) > 0 {
					var rawIDs []int
					for mid := range intersectedAnimes {
						rawIDs = append(rawIDs, mid)
					}
					// Slice batch
					if len(rawIDs) > 100 {
						rawIDs = rawIDs[:100]
					}
					animes, err := r.AnimeRepo.GetByAnilistIDs(dbCtx, rawIDs)
					if err == nil {
						for _, a := range animes {
							watchedAnimeIDs = append(watchedAnimeIDs, a.ID)
						}
					}
				}
			}
		}

		themeTypes := []string{"OP", "ED"}
		if themeTypeConfig == "OP" {
			themeTypes = []string{"OP"}
		} else if themeTypeConfig == "ED" {
			themeTypes = []string{"ED"}
		}

		// Query primary pool songs
		songs, err := r.SongRepo.GetRandomSongsForAMQ(dbCtx, watchedAnimeIDs, themeTypes, maxRounds, nil)
		if err != nil {
			log.Printf("[AMQ] Failed to fetch primary songs: %v", err)
		}

		// Backfill from general pool if needed
		if len(songs) < maxRounds {
			needed := maxRounds - len(songs)
			var excludeIDs []uint64
			for _, s := range songs {
				excludeIDs = append(excludeIDs, s.ID)
			}
			backfill, errBF := r.SongRepo.GetRandomSongsForAMQ(dbCtx, nil, themeTypes, needed, excludeIDs)
			if errBF == nil {
				songs = append(songs, backfill...)
			}
		}

		if len(songs) == 0 {
			log.Printf("[AMQ] Failed to initialize song pool: zero songs fetched")
			r.SendEvent(RoomEvent{Type: EvResetToLobby, Data: "Failed to initialize song pool. Please try again."})
			return
		}

		r.mu.Lock()
		r.SongPool = songs
		r.mu.Unlock()

		r.SendEvent(RoomEvent{Type: EvPoolLoaded, Data: nil})
	}()
}

func (r *LobbyRoom) StartGame(sessionID string) {
	r.SendEvent(RoomEvent{Type: EvStartGame, Data: sessionID})
}

func (r *LobbyRoom) startRound() {
	r.mu.Lock()
	if r.CurrentRound >= len(r.SongPool) || r.CurrentRound >= r.Config.MaxRounds {
		r.mu.Unlock()
		r.endGame()
		return
	}

	r.Status = "playing"
	r.CurrentSong = &r.SongPool[r.CurrentRound]
	r.CurrentFakes = nil

	// Clear player guesses and locks
	for _, p := range r.Players {
		p.LastGuess = ""
		p.Locked = false
		p.LastGuessCorrect = false
	}
	
	gameType := r.Config.GameType
	currentSongAnimeID := r.CurrentSong.AnimeID
	r.mu.Unlock()

	// If multiple choice, fetch 3 fake options from DB (outside lock)
	var fakes []domain.Anime
	if gameType == "multiple-choice" {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		var err error
		fakes, err = r.AnimeRepo.GetRandomAnimes(ctx, 3, []uint64{currentSongAnimeID})
		cancel()
		if err != nil {
			log.Printf("[AMQ] Error getting fake animes: %v", err)
		}
	}

	r.mu.Lock()
	r.CurrentFakes = fakes
	audioURL := r.resolveAudioURL(r.CurrentSong)

	// Generate random starting percent between 0% and 55%
	r.StartPercent = rand.Float64() * 0.55

	// Prepare multiple choice options payload
	r.CurrentOptions = nil
	if gameType == "multiple-choice" && r.CurrentSong.Anime != nil {
		correctDTO := dto.ToAnimeMinimalDTO(r.CurrentSong.Anime)
		r.CurrentOptions = append(r.CurrentOptions, correctDTO)
		for _, fake := range r.CurrentFakes {
			r.CurrentOptions = append(r.CurrentOptions, dto.ToAnimeMinimalDTO(&fake))
		}
		// Shuffle options
		rand.Shuffle(len(r.CurrentOptions), func(i, j int) {
			r.CurrentOptions[i], r.CurrentOptions[j] = r.CurrentOptions[j], r.CurrentOptions[i]
		})
	}

	r.TimerType = "guess"
	r.TimerStart = time.Now()
	r.TimerDuration = time.Duration(r.Config.GuessTime) * time.Second

	if r.Timer != nil {
		r.Timer.Stop()
	}

	r.Timer = time.AfterFunc(r.TimerDuration, func() {
		r.SendEvent(RoomEvent{Type: EvTimerExpired, Data: "guess"})
	})

	// Prepare values for broadcast under lock
	currentRound := r.CurrentRound + 1
	maxRounds := r.Config.MaxRounds
	guessTime := r.Config.GuessTime
	currentOptions := r.CurrentOptions
	startPercent := r.StartPercent
	r.mu.Unlock()

	// Broadcast round start to clients
	r.broadcast("round_start", map[string]interface{}{
		"current_round": currentRound,
		"max_rounds":    maxRounds,
		"guess_time":    guessTime,
		"audio_url":     audioURL,
		"game_type":     gameType,
		"options":       currentOptions,
		"start_percent": startPercent,
	})
	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) handleSubmitGuess(ev *GuessEvent) {
	r.mu.Lock()
	if r.Status != "playing" {
		r.mu.Unlock()
		return
	}

	player, exists := r.Players[ev.SessionID]
	if !exists || player.IsSpectator || player.Locked || player.Offline {
		r.mu.Unlock()
		return
	}

	player.LastGuess = ev.AnimeSlug
	player.Locked = true
	r.mu.Unlock()

	// Check if all online players have locked their answers
	r.checkAllLockedAndProceed()
	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) checkAllLockedAndProceed() {
	r.mu.Lock()
	allLocked := true
	activePlayers := 0
	for _, p := range r.Players {
		if !p.Offline && !p.IsSpectator {
			activePlayers++
			if !p.Locked {
				allLocked = false
			}
		}
	}

	if allLocked && activePlayers > 0 {
		// Cancel timer and transition directly to reveal
		if r.Timer != nil {
			r.Timer.Stop()
		}
		r.mu.Unlock()
		go func() {
			r.SendEvent(RoomEvent{Type: EvTimerExpired, Data: "guess"})
		}()
	} else {
		r.mu.Unlock()
	}
}

func (r *LobbyRoom) handleTimerExpired(timerType string) {
	r.mu.Lock()
	if r.TimerType != timerType {
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()

	if timerType == "guess" {
		r.revealAnswers()
	} else if timerType == "reveal" {
		r.mu.Lock()
		r.CurrentRound++
		r.mu.Unlock()
		r.startRound()
	}
}

func (r *LobbyRoom) revealAnswers() {
	r.mu.Lock()
	r.Status = "reveal"

	// Resolve correct anime
	var correctSlug string
	var correctUUID string
	var correctAnilistID int64

	if r.CurrentSong.Anime != nil {
		correctSlug = r.CurrentSong.Anime.Slug
		correctUUID = r.CurrentSong.Anime.UUID
		if r.CurrentSong.Anime.AnilistID != nil {
			correctAnilistID = *r.CurrentSong.Anime.AnilistID
		}
	}

	// Validate guesses and calculate points
	var correctPlayers []string
	results := make(map[string]map[string]interface{})
	for sid, p := range r.Players {
		if p.IsSpectator {
			continue
		}
		correct := false
		if p.LastGuess != "" {
			guessClean := p.LastGuess
			// Match slug OR UUID OR AniList ID
			if guessClean == correctSlug || guessClean == correctUUID {
				correct = true
			} else if correctAnilistID > 0 && guessClean == fmt.Sprintf("%d", correctAnilistID) {
				correct = true
			}
		}

		if correct {
			p.Score++
			p.LastGuessCorrect = true
			correctPlayers = append(correctPlayers, p.Nickname)
		} else {
			p.LastGuessCorrect = false
		}

		results[sid] = map[string]interface{}{
			"correct": correct,
			"guess":   p.LastGuess,
		}
	}

	// Resolve cover & banner URL using mediaService
	songDTO := dto.ToSongMinimalDTO(r.CurrentSong)
	if songDTO.Anime != nil {
		if r.CurrentSong.Anime.Cover != nil {
			resolved := r.MediaService.GetURL(*r.CurrentSong.Anime.Cover)
			songDTO.Anime.CoverUrl = resolved
		}
		if r.CurrentSong.Anime.Banner != nil {
			resolved := r.MediaService.Resolve(r.CurrentSong.Anime.Banner)
			songDTO.Anime.BannerUrl = resolved
		}
	}

	r.TimerType = "reveal"
	r.TimerStart = time.Now()
	r.TimerDuration = time.Duration(r.Config.RevealTime) * time.Second

	if r.Timer != nil {
		r.Timer.Stop()
	}

	r.Timer = time.AfterFunc(r.TimerDuration, func() {
		r.SendEvent(RoomEvent{Type: EvTimerExpired, Data: "reveal"})
	})
	r.mu.Unlock()

	r.broadcast("round_ended", map[string]interface{}{
		"song":    songDTO,
		"results": results,
	})
	r.broadcast("lobby_state_update", r.getRoomStatePayload())

	var logText string
	if len(correctPlayers) > 0 {
		logText = fmt.Sprintf("Round ended. Correct answers from: %s", strings.Join(correctPlayers, ", "))
	} else {
		logText = "Round ended. No correct answers!"
	}
	r.broadcast("chat_message", map[string]interface{}{
		"sender":    "System",
		"text":      logText,
		"type":      "system",
		"timestamp": time.Now(),
	})
}

func (r *LobbyRoom) handleSkipSummary(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || !player.IsHost || r.Status != "reveal" {
		r.mu.Unlock()
		return
	}

	// Force timer expiration immediately
	if r.Timer != nil {
		r.Timer.Stop()
	}
	r.CurrentRound++
	r.mu.Unlock()
	r.startRound()
}

func (r *LobbyRoom) handleResetToLobby(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || !player.IsHost || r.Status != "finished" {
		r.mu.Unlock()
		return
	}

	r.Status = "lobby"
	r.CurrentRound = 0
	r.CurrentSong = nil
	r.SongPool = nil

	for _, p := range r.Players {
		p.Score = 0
		p.Locked = false
		p.IsReady = p.IsHost // Host ready, others not
		p.LastGuess = ""
		p.LastGuessCorrect = false
	}
	r.mu.Unlock()

	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) forceResetToLobby(errMsg string) {
	r.mu.Lock()
	r.Status = "lobby"
	r.CurrentRound = 0
	r.CurrentSong = nil
	r.SongPool = nil

	for _, p := range r.Players {
		p.Score = 0
		p.Locked = false
		p.IsReady = p.IsHost // Host ready, others not
		p.LastGuess = ""
		p.LastGuessCorrect = false
	}
	r.mu.Unlock()

	r.broadcast("lobby_state_update", r.getRoomStatePayload())
	if errMsg != "" {
		r.broadcast("chat_message", map[string]interface{}{
			"sender":    "System",
			"text":      errMsg,
			"type":      "system",
			"timestamp": time.Now(),
		})
	}
}

func (r *LobbyRoom) endGame() {
	r.mu.Lock()
	r.Status = "finished"

	type playerXPInfo struct {
		UserUUID string
		Score    int
	}
	var playersToAward []playerXPInfo
	for _, p := range r.Players {
		if p.UserUUID != "" && p.Score > 0 && !p.IsSpectator {
			playersToAward = append(playersToAward, playerXPInfo{
				UserUUID: p.UserUUID,
				Score:    p.Score,
			})
		}
	}
	roomID := r.RoomID
	maxRounds := r.Config.MaxRounds
	r.mu.Unlock()

	// Award XP outside lock
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, info := range playersToAward {
		user, err := r.UserRepo.GetByUUID(ctx, info.UserUUID)
		if err == nil {
			_ = r.XPUsecase.AwardXP(ctx, user.ID, "amq_completion", map[string]interface{}{
				"lobby_room": roomID,
				"score":      info.Score,
				"rounds":     maxRounds,
			})
		}
	}

	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) cleanupOfflinePlayers() {
	r.mu.Lock()
	now := time.Now()
	changed := false

	for sid, p := range r.Players {
		if p.Offline && p.OfflineSince != nil {
			if now.Sub(*p.OfflineSince) > 60*time.Second {
				log.Printf("[AMQ] Purging offline player %s", p.Nickname)
				delete(r.Players, sid)
				delete(r.Conns, sid)
				changed = true
			}
		}
	}
	r.mu.Unlock()

	if changed {
		r.broadcast("lobby_state_update", r.getRoomStatePayload())
	}
}

func (r *LobbyRoom) getRoomStatePayload() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	playersList := make([]domain.AMQPlayer, 0, len(r.Players))
	spectatorsList := make([]domain.AMQPlayer, 0)
	for _, p := range r.Players {
		if p.IsSpectator {
			spectatorsList = append(spectatorsList, *p)
		} else {
			playersList = append(playersList, *p)
		}
	}

	// Calculate remaining timer ticks
	var timerLeft int = 0
	if r.Status == "playing" || r.Status == "reveal" {
		elapsed := time.Since(r.TimerStart)
		rem := r.TimerDuration - elapsed
		if rem > 0 {
			timerLeft = int(rem.Seconds())
		}
	}

	var roundData map[string]interface{}
	if (r.Status == "playing" || r.Status == "reveal") && r.CurrentSong != nil {
		audioURL := r.resolveAudioURL(r.CurrentSong)

		roundData = map[string]interface{}{
			"current_round": r.CurrentRound + 1,
			"max_rounds":    r.Config.MaxRounds,
			"guess_time":    r.Config.GuessTime,
			"audio_url":     audioURL,
			"game_type":     r.Config.GameType,
			"options":       r.CurrentOptions,
			"start_percent": r.StartPercent,
		}
	}

	return map[string]interface{}{
		"room_id":       r.RoomID,
		"status":        r.Status,
		"config":        r.Config,
		"current_round": r.CurrentRound + 1,
		"players":       playersList,
		"spectators":    spectatorsList,
		"timer_left":    timerLeft,
		"round_data":    roundData,
	}
}

func (r *LobbyRoom) resolveAudioURL(song *domain.Song) string {
	if song == nil || len(song.Variants) == 0 {
		return ""
	}
	v := song.Variants[0]
	if v.Video != nil {
		if v.Video.LocalUrl != nil {
			return r.MediaService.GetURL(*v.Video.LocalUrl)
		} else if v.Video.EmbedUrl != nil {
			return *v.Video.EmbedUrl
		} else if v.Video.VideoSrc != nil {
			return r.MediaService.GetURL(*v.Video.VideoSrc)
		}
	}
	return ""
}

func (r *LobbyRoom) GetRoomInfo() domain.AMQRoomInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hostNick := "Unknown"
	playerCount := 0
	spectatorCount := 0
	for _, p := range r.Players {
		if p.IsHost {
			hostNick = p.Nickname
		}
		if p.IsSpectator {
			spectatorCount++
		} else {
			playerCount++
		}
	}

	return domain.AMQRoomInfo{
		RoomID:         r.RoomID,
		Name:           r.Config.Name,
		HostNickname:   hostNick,
		PlayerCount:    playerCount,
		SpectatorCount: spectatorCount,
		MaxRounds:      r.Config.MaxRounds,
		Status:         r.Status,
		Private:        r.Config.Private,
		ThemeType:      r.Config.ThemeType,
		GameType:       r.Config.GameType,
	}
}

func (r *LobbyRoom) ShouldDestroy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.Closed {
		return true
	}

	// If there are online players, do not destroy
	for _, p := range r.Players {
		if !p.Offline {
			return false
		}
	}

	// No online players:
	// If the room has never been joined (len == 0 and brand new):
	if len(r.Players) == 0 {
		// Give 3 minutes for initial connection
		if time.Since(r.CreatedAt) > 3*time.Minute {
			return true
		}
	}

	// If it was active but now everyone is offline/purged:
	// If it has been inactive for more than 2 minutes:
	if time.Since(r.LastActive) > 2*time.Minute {
		return true
	}

	return false
}

func (r *LobbyRoom) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Stop the timer
	if r.Timer != nil {
		r.Timer.Stop()
	}

	// Close all active connections
	for _, conn := range r.Conns {
		if conn != nil {
			_ = conn.Close()
		}
	}

	// Close event channel safely
	defer func() {
		recover()
	}()
	close(r.EventChan)
}

func (r *LobbyRoom) handleChat(ev *ChatEvent) {
	r.mu.Lock()
	player, exists := r.Players[ev.SessionID]
	if !exists {
		r.mu.Unlock()
		return
	}
	nickname := player.Nickname
	r.mu.Unlock()

	r.broadcast("chat_message", map[string]interface{}{
		"sender":    nickname,
		"text":      ev.Text,
		"type":      "user",
		"timestamp": time.Now(),
	})
}

// ensureHostActive assumes the room lock (r.mu) is already held.
func (r *LobbyRoom) ensureHostActive() {
	var onlineHost *domain.AMQPlayer
	var firstOnlinePlayer *domain.AMQPlayer

	for _, p := range r.Players {
		if !p.Offline && !p.IsSpectator {
			if firstOnlinePlayer == nil {
				firstOnlinePlayer = p
			}
			if p.IsHost {
				onlineHost = p
			}
		}
	}

	if onlineHost == nil && firstOnlinePlayer != nil {
		for _, p := range r.Players {
			p.IsHost = false
		}
		firstOnlinePlayer.IsHost = true
		firstOnlinePlayer.IsReady = true
		log.Printf("[AMQ] Host assigned/migrated to online player %s", firstOnlinePlayer.Nickname)
	}
}

// handleTransferHost handles transferring the host status manually.
func (r *LobbyRoom) handleTransferHost(ev *TransferHostEvent) {
	r.mu.Lock()

	log.Printf("[AMQ] handleTransferHost: ev.SessionID=%s, ev.TargetSessionID=%s", ev.SessionID, ev.TargetSessionID)

	// Verify requester is the host
	requester, exists := r.Players[ev.SessionID]
	if !exists {
		log.Printf("[AMQ] transfer_host failed: requester session %s not found in room", ev.SessionID)
		r.mu.Unlock()
		return
	}
	if !requester.IsHost {
		log.Printf("[AMQ] transfer_host failed: requester %s is not host (IsHost=%t)", requester.Nickname, requester.IsHost)
		r.mu.Unlock()
		return
	}

	// Verify target player exists and is online/not spectator
	target, exists := r.Players[ev.TargetSessionID]
	if !exists {
		log.Printf("[AMQ] transfer_host failed: target session %s not found in room", ev.TargetSessionID)
		r.mu.Unlock()
		return
	}
	if target.Offline {
		log.Printf("[AMQ] transfer_host failed: target %s is offline", target.Nickname)
		r.mu.Unlock()
		return
	}
	if target.IsSpectator {
		log.Printf("[AMQ] transfer_host failed: target %s is a spectator", target.Nickname)
		r.mu.Unlock()
		return
	}

	// Perform transfer
	requester.IsHost = false
	target.IsHost = true
	target.IsReady = true // Host is always ready
	log.Printf("[AMQ] Host transferred manually from %s to %s", requester.Nickname, target.Nickname)

	requesterName := requester.Nickname
	targetName := target.Nickname

	r.mu.Unlock()

	// Broadcast updated state
	log.Printf("[AMQ] Broadcasting lobby_state_update after host transfer to %s", targetName)
	r.broadcast("lobby_state_update", r.getRoomStatePayload())
	log.Printf("[AMQ] lobby_state_update broadcast complete")

	r.broadcast("chat_message", map[string]interface{}{
		"sender":    "System",
		"text":      fmt.Sprintf("%s transferred host to %s", requesterName, targetName),
		"type":      "system",
		"timestamp": time.Now(),
	})
	log.Printf("[AMQ] chat_message broadcast complete for host transfer")
}

func (r *LobbyRoom) handleCloseRoom(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || !player.IsHost {
		r.mu.Unlock()
		return
	}
	r.Closed = true
	r.mu.Unlock()

	// Broadcast room_closed to everyone
	r.broadcast("room_closed", nil)

	// Close all connections and exit event loop
	r.Close()
}

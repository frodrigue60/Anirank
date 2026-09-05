package rate

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
	"anirank/api/internal/infrastructure"

	"github.com/google/uuid"
)

type WSConn interface {
	WriteJSON(v interface{}) error
	Close() error
}

type RoomEventType int

const (
	EvJoin RoomEventType = iota
	EvLeave
	EvConfigUpdate
	EvStartSession
	EvQueueAdd
	EvQueueRemove
	EvSetSong
	EvNext
	EvSubmitRating
	EvVoteSkip
	EvEndSession
	EvResetToLobby
	EvChat
	EvTransferHost
	EvCloseRoom
	EvPoolLoaded
)

type RoomEvent struct {
	Type RoomEventType
	Data interface{}
}

type JoinEvent struct {
	SessionID   string
	Conn        WSConn
	User        *domain.User
	Nickname    string
	DeviceID    string
	AsSpectator bool
}

type ConfigUpdateEvent struct {
	SessionID string
	Config    domain.RateConfig
}

type QueueAddEvent struct {
	SessionID string
	SongUUID  string
}

type QueueRemoveEvent struct {
	SessionID string
	ItemID    string
}

type SetSongEvent struct {
	SessionID string
	SongUUID  string
}

type RatingEvent struct {
	SessionID string
	Score     float64
}

type TransferHostEvent struct {
	SessionID       string
	TargetSessionID string
}

type ChatEvent struct {
	SessionID string
	Text      string
}

type PoolLoadedEvent struct {
	Songs []domain.Song
	Error string
}

type queuedSong struct {
	Item domain.RateQueueItem
	Song *domain.Song
}

type LobbyRoom struct {
	RoomID     string
	Config     domain.RateConfig
	Status     string // lobby | waiting | rating | finished
	Players    map[string]*domain.RatePlayer
	Conns      map[string]WSConn
	Queue      []queuedSong
	CurrentSong *domain.Song
	// CurrentRatings maps sessionID -> score (0-100) for this song in the live session.
	CurrentRatings map[string]float64
	// PriorScores maps sessionID -> existing global rating (0-100) loaded at beginRating.
	PriorScores map[string]float64
	// SkipVotes maps sessionID -> true when that player voted to skip the current song.
	SkipVotes  map[string]bool
	SongsRated int
	EventChan   chan RoomEvent
	CreatedAt   time.Time
	LastActive  time.Time
	Closed      bool
	OnDestroy   func(roomID string)

	SongRepo     domain.SongRepository
	UserRepo     domain.UserRepository
	MediaService infrastructure.MediaService
	SongRater    SongRater

	mu sync.RWMutex
}

func generateUUID() string {
	return uuid.New().String()
}

func NewLobbyRoom(
	roomID string,
	config domain.RateConfig,
	songRepo domain.SongRepository,
	userRepo domain.UserRepository,
	mediaService infrastructure.MediaService,
	songRater SongRater,
) *LobbyRoom {
	now := time.Now()
	return &LobbyRoom{
		RoomID:         roomID,
		Config:         config,
		Status:         "lobby",
		Players:        make(map[string]*domain.RatePlayer),
		Conns:          make(map[string]WSConn),
		Queue:          make([]queuedSong, 0),
		CurrentRatings: make(map[string]float64),
		PriorScores:    make(map[string]float64),
		SkipVotes:      make(map[string]bool),
		EventChan:      make(chan RoomEvent, 100),
		CreatedAt:      now,
		LastActive:     now,
		SongRepo:       songRepo,
		UserRepo:       userRepo,
		MediaService:   mediaService,
		SongRater:      songRater,
	}
}

func (r *LobbyRoom) SendEvent(ev RoomEvent) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("[RATE] Recovered from panic sending event to room %s: %v", r.RoomID, err)
		}
	}()
	r.EventChan <- ev
}

func (r *LobbyRoom) Start() {
	go r.run()
}

func (r *LobbyRoom) run() {
	cleanupTicker := time.NewTicker(10 * time.Second)
	defer cleanupTicker.Stop()

	for {
		stopped := false
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("[RATE] PANIC recovered in room %s event loop: %v", r.RoomID, rec)
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
	case EvConfigUpdate:
		r.handleConfigUpdate(ev.Data.(*ConfigUpdateEvent))
	case EvStartSession:
		r.handleStartSession(ev.Data.(string))
	case EvQueueAdd:
		r.handleQueueAdd(ev.Data.(*QueueAddEvent))
	case EvQueueRemove:
		r.handleQueueRemove(ev.Data.(*QueueRemoveEvent))
	case EvSetSong:
		r.handleSetSong(ev.Data.(*SetSongEvent))
	case EvNext:
		r.handleNext(ev.Data.(string))
	case EvSubmitRating:
		r.handleSubmitRating(ev.Data.(*RatingEvent))
	case EvVoteSkip:
		r.handleVoteSkip(ev.Data.(string))
	case EvEndSession:
		r.handleEndSession(ev.Data.(string))
	case EvResetToLobby:
		r.handleResetToLobby(ev.Data.(string))
	case EvChat:
		r.handleChat(ev.Data.(*ChatEvent))
	case EvTransferHost:
		r.handleTransferHost(ev.Data.(*TransferHostEvent))
	case EvCloseRoom:
		r.handleCloseRoom(ev.Data.(string))
	case EvPoolLoaded:
		r.handlePoolLoaded(ev.Data.(*PoolLoadedEvent))
	}
}

type outMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

func (r *LobbyRoom) broadcast(msgType string, payload interface{}) {
	msg := outMessage{Type: msgType, Payload: payload}
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
			log.Printf("[RATE] Error writing to session %s: %v", sid, err)
		}
	}
}

func (r *LobbyRoom) sendTo(sessionID string, msgType string, payload interface{}) {
	msg := outMessage{Type: msgType, Payload: payload}
	r.mu.RLock()
	conn, ok := r.Conns[sessionID]
	r.mu.RUnlock()
	if ok && conn != nil {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("[RATE] Error in sendTo session %s: %v", sessionID, err)
		}
	}
}

func (r *LobbyRoom) broadcastPersonalizedState() {
	r.mu.RLock()
	sids := make([]string, 0, len(r.Conns))
	for sid, conn := range r.Conns {
		if conn != nil {
			sids = append(sids, sid)
		}
	}
	r.mu.RUnlock()

	for _, sid := range sids {
		r.sendTo(sid, "lobby_state_update", r.getRoomStatePayload(sid))
	}
}

func (r *LobbyRoom) handleJoin(ev *JoinEvent) {
	r.mu.Lock()

	var existingPlayer *domain.RatePlayer
	var oldSessionID string

	// 1) Authenticated: reclaim seat by user UUID (covers leave/rejoin and multi-tab).
	if ev.User != nil {
		for sid, p := range r.Players {
			if p.UserUUID == ev.User.UUID {
				existingPlayer = p
				oldSessionID = sid
				break
			}
		}
	}

	// 2) Same browser/device: reclaim even if prior seat was auth or guest.
	//    Without this, a reconnect that briefly lacks a JWT creates a ghost guest
	//    while the offline auth seat remains — queue adds then fail (everyone mode
	//    requires user_uuid) and MaxPlayers fills with zombies.
	if existingPlayer == nil && ev.DeviceID != "" {
		var offlineMatch *domain.RatePlayer
		var offlineSID string
		for sid, p := range r.Players {
			if p.DeviceID != ev.DeviceID {
				continue
			}
			if p.Offline {
				offlineMatch = p
				offlineSID = sid
				break
			}
			if existingPlayer == nil {
				existingPlayer = p
				oldSessionID = sid
			}
		}
		if offlineMatch != nil {
			existingPlayer = offlineMatch
			oldSessionID = offlineSID
		}
	}

	// 3) Guest nickname fallback (offline only).
	if existingPlayer == nil && ev.User == nil {
		for sid, p := range r.Players {
			if p.UserUUID == "" && p.Nickname == ev.Nickname && p.Offline {
				existingPlayer = p
				oldSessionID = sid
				break
			}
		}
	}

	onlineNonSpectators := 0
	for sid, p := range r.Players {
		if existingPlayer != nil && sid == oldSessionID {
			continue
		}
		if !p.IsSpectator && !p.Offline {
			onlineNonSpectators++
		}
	}

	isReconnect := existingPlayer != nil

	if existingPlayer != nil {
		delete(r.Conns, oldSessionID)
		delete(r.Players, oldSessionID)

		if score, ok := r.CurrentRatings[oldSessionID]; ok {
			delete(r.CurrentRatings, oldSessionID)
			r.CurrentRatings[ev.SessionID] = score
		}
		if prior, ok := r.PriorScores[oldSessionID]; ok {
			delete(r.PriorScores, oldSessionID)
			r.PriorScores[ev.SessionID] = prior
		}
		if voted, ok := r.SkipVotes[oldSessionID]; ok {
			delete(r.SkipVotes, oldSessionID)
			r.SkipVotes[ev.SessionID] = voted
		}

		existingPlayer.SessionID = ev.SessionID
		existingPlayer.Offline = false
		existingPlayer.OfflineSince = nil
		existingPlayer.Nickname = ev.Nickname
		existingPlayer.DeviceID = ev.DeviceID
		existingPlayer.IsSpectator = ev.AsSpectator
		// Preserve UserID/UserUUID when reconnect arrives without JWT; upgrade when present.
		if ev.User != nil {
			existingPlayer.UserID = ev.User.ID
			existingPlayer.UserUUID = ev.User.UUID
			if ev.User.AvatarUrl != nil {
				avatar := *ev.User.AvatarUrl
				existingPlayer.AvatarURL = &avatar
			}
			if ev.User.ProfileColor != nil {
				color := *ev.User.ProfileColor
				existingPlayer.ProfileColor = &color
			}
		}
		r.Players[ev.SessionID] = existingPlayer
		r.Conns[ev.SessionID] = ev.Conn
	} else {
		if !ev.AsSpectator && onlineNonSpectators >= r.Config.MaxPlayers {
			r.mu.Unlock()
			if ev.Conn != nil {
				_ = ev.Conn.WriteJSON(outMessage{Type: "error", Payload: "Room is full"})
			}
			return
		}

		isHost := !ev.AsSpectator && onlineNonSpectators == 0
		avatar := ""
		if ev.User != nil && ev.User.AvatarUrl != nil {
			avatar = *ev.User.AvatarUrl
		}
		color := "#683bc9"
		if ev.User != nil && ev.User.ProfileColor != nil {
			color = *ev.User.ProfileColor
		}
		userUUID := ""
		var userID uint64
		if ev.User != nil {
			userUUID = ev.User.UUID
			userID = ev.User.ID
		}

		player := &domain.RatePlayer{
			SessionID:    ev.SessionID,
			UserUUID:     userUUID,
			UserID:       userID,
			Nickname:     ev.Nickname,
			AvatarURL:    &avatar,
			ProfileColor: &color,
			DeviceID:     ev.DeviceID,
			IsHost:       isHost,
			IsSpectator:  ev.AsSpectator,
		}
		r.Players[ev.SessionID] = player
		r.Conns[ev.SessionID] = ev.Conn
	}

	// Drop duplicate seats for this user/device left by flaky reconnect races.
	r.dedupePlayerSeatsLocked(ev.SessionID, ev.DeviceID, ev.User)

	r.ensureHostActive()
	status := r.Status
	song := r.CurrentSong
	needsPrior := false
	var joinUserID uint64
	if p := r.Players[ev.SessionID]; p != nil {
		_, hasPrior := r.PriorScores[ev.SessionID]
		if !hasPrior && status == "rating" && song != nil && p.UserID > 0 && !p.IsSpectator {
			needsPrior = true
			joinUserID = p.UserID
		}
	}
	r.mu.Unlock()

	if needsPrior && r.SongRater != nil && song != nil {
		rating, err := r.SongRater.GetUserSongRating(context.Background(), joinUserID, song.ID)
		if err == nil && rating != nil && rating.Rating > 0 {
			r.mu.Lock()
			if r.PriorScores == nil {
				r.PriorScores = make(map[string]float64)
			}
			r.PriorScores[ev.SessionID] = rating.Rating
			r.mu.Unlock()
		}
	}

	r.sendTo(ev.SessionID, "lobby_state_update", r.getRoomStatePayload(ev.SessionID))
	r.broadcastPersonalizedState()

	// Only announce first joins — reconnects/reassociations spam the chat otherwise.
	if !isReconnect {
		role := "player"
		if ev.AsSpectator {
			role = "spectator"
		}
		r.broadcast("chat_message", map[string]interface{}{
			"sender":    "System",
			"text":      fmt.Sprintf("%s joined as %s", ev.Nickname, role),
			"type":      "system",
			"timestamp": time.Now(),
		})
	}
}

// dedupePlayerSeatsLocked removes leftover seats that share this device or user UUID.
// Caller must hold r.mu.
func (r *LobbyRoom) dedupePlayerSeatsLocked(keepSessionID, deviceID string, user *domain.User) {
	for sid, p := range r.Players {
		if sid == keepSessionID {
			continue
		}
		sameDevice := deviceID != "" && p.DeviceID == deviceID
		sameUser := user != nil && user.UUID != "" && p.UserUUID == user.UUID
		if !sameDevice && !sameUser {
			continue
		}
		delete(r.Players, sid)
		delete(r.Conns, sid)
		if score, ok := r.CurrentRatings[sid]; ok {
			delete(r.CurrentRatings, sid)
			if _, exists := r.CurrentRatings[keepSessionID]; !exists {
				r.CurrentRatings[keepSessionID] = score
			}
		}
		if prior, ok := r.PriorScores[sid]; ok {
			delete(r.PriorScores, sid)
			if _, exists := r.PriorScores[keepSessionID]; !exists {
				r.PriorScores[keepSessionID] = prior
			}
		}
		if voted, ok := r.SkipVotes[sid]; ok {
			delete(r.SkipVotes, sid)
			if _, exists := r.SkipVotes[keepSessionID]; !exists {
				r.SkipVotes[keepSessionID] = voted
			}
		}
	}
}

func (r *LobbyRoom) handleLeave(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists {
		r.mu.Unlock()
		return
	}
	player.Offline = true
	now := time.Now()
	player.OfflineSince = &now
	r.Conns[sessionID] = nil
	r.ensureHostActive()
	nick := player.Nickname
	r.mu.Unlock()

	r.broadcastPersonalizedState()
	r.broadcast("chat_message", map[string]interface{}{
		"sender":    "System",
		"text":      fmt.Sprintf("%s went offline", nick),
		"type":      "system",
		"timestamp": time.Now(),
	})
}

func (r *LobbyRoom) handleConfigUpdate(ev *ConfigUpdateEvent) {
	r.mu.Lock()
	player, exists := r.Players[ev.SessionID]
	if !exists || !player.IsHost || r.Status != "lobby" {
		r.mu.Unlock()
		return
	}
	prevSeasonal := isSeasonalPool(r.Config)
	cfg := ev.Config
	sanitizeConfig(&cfg)
	if cfg.SourceMode == SourceModeSeasonalPool && (cfg.PoolYear == "" || cfg.PoolSeason == "") {
		r.mu.Unlock()
		r.sendTo(ev.SessionID, "error", "Seasonal pool requires year and season")
		return
	}
	r.Config = cfg
	// Mode switch clears pending queue so manual vs pool never mix unexpectedly.
	if prevSeasonal != isSeasonalPool(cfg) {
		r.Queue = make([]queuedSong, 0)
		r.CurrentSong = nil
		r.CurrentRatings = make(map[string]float64)
		r.PriorScores = make(map[string]float64)
		r.SkipVotes = make(map[string]bool)
	}
	r.mu.Unlock()
	r.broadcastPersonalizedState()
}

func (r *LobbyRoom) handleStartSession(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || !player.IsHost || r.Status != "lobby" {
		r.mu.Unlock()
		return
	}
	seasonal := isSeasonalPool(r.Config)
	if seasonal && (r.Config.PoolYear == "" || r.Config.PoolSeason == "") {
		r.mu.Unlock()
		r.sendTo(sessionID, "error", "Seasonal pool requires year and season")
		return
	}
	r.Status = "waiting"
	r.SongsRated = 0
	r.mu.Unlock()
	r.broadcastPersonalizedState()

	if seasonal {
		r.broadcast("chat_message", map[string]interface{}{
			"sender":    "System",
			"text":      "Loading seasonal pool…",
			"type":      "system",
			"timestamp": time.Now(),
		})
		r.loadSeasonalPoolAsync()
		return
	}

	r.broadcast("chat_message", map[string]interface{}{
		"sender":    "System",
		"text":      "Session started — add a song or play from the queue",
		"type":      "system",
		"timestamp": time.Now(),
	})
}

func (r *LobbyRoom) handleQueueAdd(ev *QueueAddEvent) {
	if ev.SongUUID == "" {
		return
	}

	r.mu.RLock()
	player, exists := r.Players[ev.SessionID]
	status := r.Status
	queueMode := r.Config.QueueMode
	limit := r.Config.QueueLimitPerUser
	queueLen := len(r.Queue)
	seasonal := isSeasonalPool(r.Config)
	r.mu.RUnlock()

	if !exists || player.IsSpectator || player.Offline {
		r.sendTo(ev.SessionID, "error", "You cannot add songs right now — reconnect if this persists")
		return
	}
	if seasonal {
		r.sendTo(ev.SessionID, "error", "Manual theme adds are disabled in seasonal pool mode. Host can switch to manual in lobby settings.")
		return
	}
	if status != "lobby" && status != "waiting" && status != "rating" {
		return
	}
	if queueMode == QueueModeDisabled {
		r.sendTo(ev.SessionID, "error", "Queue is disabled for this room")
		return
	}
	if queueMode == QueueModeHostOnly && !player.IsHost {
		r.sendTo(ev.SessionID, "error", "Only the host can add songs to the queue")
		return
	}
	if queueMode == QueueModeEveryone && player.UserUUID == "" {
		r.sendTo(ev.SessionID, "error", "You must be logged in to add songs")
		return
	}
	if queueLen >= MaxQueueSize {
		r.sendTo(ev.SessionID, "error", "Queue is full")
		return
	}

	if queueMode == QueueModeEveryone {
		r.mu.RLock()
		count := 0
		for _, q := range r.Queue {
			if q.Item.AddedByUserUUID == player.UserUUID {
				count++
			}
		}
		r.mu.RUnlock()
		if count >= limit {
			r.sendTo(ev.SessionID, "error", fmt.Sprintf("Queue limit reached (%d songs per user)", limit))
			return
		}
	}

	song, err := r.loadSong(ev.SongUUID)
	if err != nil || song == nil {
		r.sendTo(ev.SessionID, "error", rateSongLoadErrorMessage(err))
		return
	}

	item := r.buildQueueItem(song, player)
	r.mu.Lock()
	r.Queue = append(r.Queue, queuedSong{Item: item, Song: song})
	r.mu.Unlock()

	r.broadcastPersonalizedState()
}

func (r *LobbyRoom) handleQueueRemove(ev *QueueRemoveEvent) {
	r.mu.Lock()
	player, exists := r.Players[ev.SessionID]
	if !exists || player.IsSpectator {
		r.mu.Unlock()
		return
	}

	idx := -1
	for i, q := range r.Queue {
		if q.Item.ItemID == ev.ItemID {
			idx = i
			break
		}
	}
	if idx < 0 {
		r.mu.Unlock()
		return
	}

	item := r.Queue[idx].Item
	canRemove := player.IsHost ||
		(item.AddedBySessionID == ev.SessionID) ||
		(player.UserUUID != "" && item.AddedByUserUUID == player.UserUUID)
	if !canRemove {
		r.mu.Unlock()
		r.sendTo(ev.SessionID, "error", "You cannot remove this queue item")
		return
	}

	r.Queue = append(r.Queue[:idx], r.Queue[idx+1:]...)
	r.mu.Unlock()
	r.broadcastPersonalizedState()
}

func (r *LobbyRoom) handleSetSong(ev *SetSongEvent) {
	r.mu.RLock()
	player, exists := r.Players[ev.SessionID]
	status := r.Status
	seasonal := isSeasonalPool(r.Config)
	r.mu.RUnlock()

	if !exists || !player.IsHost {
		return
	}
	if seasonal {
		r.sendTo(ev.SessionID, "error", "Manual theme picks are disabled in seasonal pool mode. Use Next to advance the pool.")
		return
	}
	if status != "waiting" && status != "rating" {
		r.sendTo(ev.SessionID, "error", "Start the session before setting a song")
		return
	}

	song, err := r.loadSong(ev.SongUUID)
	if err != nil || song == nil {
		r.sendTo(ev.SessionID, "error", rateSongLoadErrorMessage(err))
		return
	}

	r.beginRating(song)
}

func (r *LobbyRoom) handleNext(sessionID string) {
	r.mu.RLock()
	player, exists := r.Players[sessionID]
	status := r.Status
	queueLen := len(r.Queue)
	r.mu.RUnlock()

	if !exists || !player.IsHost {
		return
	}
	if status != "waiting" && status != "rating" {
		return
	}

	if queueLen == 0 {
		r.mu.Lock()
		if r.CurrentSong != nil {
			r.SongsRated++
		}
		r.CurrentSong = nil
		r.CurrentRatings = make(map[string]float64)
		r.PriorScores = make(map[string]float64)
		r.SkipVotes = make(map[string]bool)
		r.Status = "waiting"
		r.mu.Unlock()
		r.broadcastPersonalizedState()
		return
	}

	r.mu.Lock()
	next := r.Queue[0]
	r.Queue = r.Queue[1:]
	r.mu.Unlock()

	r.beginRating(next.Song)
}

func (r *LobbyRoom) beginRating(song *domain.Song) {
	type lookup struct {
		sessionID string
		userID    uint64
	}

	r.mu.Lock()
	if r.CurrentSong != nil {
		r.SongsRated++
	}
	r.CurrentSong = song
	r.CurrentRatings = make(map[string]float64)
	r.PriorScores = make(map[string]float64)
	r.SkipVotes = make(map[string]bool)
	r.Status = "rating"
	lookups := make([]lookup, 0, len(r.Players))
	for sid, p := range r.Players {
		if p.UserID > 0 && !p.IsSpectator && !p.Offline {
			lookups = append(lookups, lookup{sessionID: sid, userID: p.UserID})
		}
	}
	r.mu.Unlock()

	priors := make(map[string]float64, len(lookups))
	if r.SongRater != nil && song != nil && song.ID > 0 {
		for _, l := range lookups {
			rating, err := r.SongRater.GetUserSongRating(context.Background(), l.userID, song.ID)
			if err != nil || rating == nil {
				continue
			}
			if rating.Rating > 0 {
				priors[l.sessionID] = rating.Rating
			}
		}
	}

	r.mu.Lock()
	r.PriorScores = priors
	sessionIDs := make([]string, 0, len(r.Conns))
	for sid, conn := range r.Conns {
		if conn != nil {
			sessionIDs = append(sessionIDs, sid)
		}
	}
	r.mu.Unlock()

	basePayload := r.buildSongStartedPayload(song)
	for _, sid := range sessionIDs {
		payload := map[string]interface{}{
			"song":      basePayload["song"],
			"audio_url": basePayload["audio_url"],
		}
		if prior, ok := priors[sid]; ok {
			payload["prior_score"] = prior
		}
		r.sendTo(sid, "song_started", payload)
	}
	r.broadcastPersonalizedState()
}

func (r *LobbyRoom) handleSubmitRating(ev *RatingEvent) {
	if ev.Score < 0 || ev.Score > 100 {
		r.sendTo(ev.SessionID, "error", "Score must be between 0 and 100")
		return
	}

	r.mu.RLock()
	player, exists := r.Players[ev.SessionID]
	status := r.Status
	song := r.CurrentSong
	r.mu.RUnlock()

	if !exists {
		r.sendTo(ev.SessionID, "error", "You are not in this room")
		return
	}
	if player.IsSpectator {
		r.sendTo(ev.SessionID, "error", "Spectators cannot rate")
		return
	}
	if player.Offline {
		r.sendTo(ev.SessionID, "error", "Reconnect before rating")
		return
	}
	if status != "rating" || song == nil {
		r.sendTo(ev.SessionID, "error", "No song is being rated right now")
		return
	}
	if player.UserUUID == "" || player.UserID == 0 {
		r.sendTo(ev.SessionID, "error", "You must be logged in to rate")
		return
	}

	_, _, err := r.SongRater.RateSongInLiveRoom(context.Background(), player.UserID, song.ID, ev.Score)
	if err != nil {
		log.Printf("[RATE] RateSong failed for user %d song %d: %v", player.UserID, song.ID, err)
		msg := "Failed to save rating"
		var appErr *domain.AppError
		if errors.As(err, &appErr) {
			msg = appErr.Message
		}
		r.sendTo(ev.SessionID, "error", msg)
		return
	}

	r.mu.Lock()
	r.CurrentRatings[ev.SessionID] = ev.Score
	autoAdvance := r.Config.AutoAdvance == AutoAdvanceAllRated && r.allEligibleRatedLocked()
	hostSID := ""
	if autoAdvance {
		for sid, p := range r.Players {
			if p.IsHost {
				hostSID = sid
				break
			}
		}
	}
	r.mu.Unlock()

	r.broadcastPersonalizedRatingUpdate()

	if autoAdvance && hostSID != "" {
		r.SendEvent(RoomEvent{Type: EvNext, Data: hostSID})
	}
}

func (r *LobbyRoom) handleVoteSkip(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || player.IsSpectator || player.Offline {
		r.mu.Unlock()
		r.sendTo(sessionID, "error", "You cannot vote to skip right now")
		return
	}
	if !r.Config.VoteSkip {
		r.mu.Unlock()
		r.sendTo(sessionID, "error", "Vote skip is disabled for this room")
		return
	}
	if r.Status != "rating" || r.CurrentSong == nil {
		r.mu.Unlock()
		r.sendTo(sessionID, "error", "No song is playing to skip")
		return
	}
	if r.SkipVotes == nil {
		r.SkipVotes = make(map[string]bool)
	}
	r.SkipVotes[sessionID] = true

	eligible := 0
	votes := 0
	for _, p := range r.Players {
		if p.IsSpectator || p.Offline {
			continue
		}
		eligible++
		if r.SkipVotes[p.SessionID] {
			votes++
		}
	}
	majority := eligible > 0 && votes*2 > eligible
	hostSID := ""
	if majority {
		for sid, p := range r.Players {
			if p.IsHost {
				hostSID = sid
				break
			}
		}
	}
	r.mu.Unlock()

	r.broadcastPersonalizedState()
	if majority {
		r.broadcast("chat_message", map[string]interface{}{
			"sender":    "System",
			"text":      "Skip vote passed — advancing",
			"type":      "system",
			"timestamp": time.Now(),
		})
		if hostSID != "" {
			r.SendEvent(RoomEvent{Type: EvNext, Data: hostSID})
		}
	}
}

func (r *LobbyRoom) broadcastPersonalizedRatingUpdate() {
	r.mu.RLock()
	sids := make([]string, 0, len(r.Conns))
	for sid, conn := range r.Conns {
		if conn != nil {
			sids = append(sids, sid)
		}
	}
	r.mu.RUnlock()

	for _, sid := range sids {
		r.sendTo(sid, "rating_update", r.buildRatingSnapshot(sid))
		r.sendTo(sid, "lobby_state_update", r.getRoomStatePayload(sid))
	}
}

func (r *LobbyRoom) allEligibleRatedLocked() bool {
	for _, p := range r.Players {
		if p.IsSpectator || p.Offline || p.UserUUID == "" {
			continue
		}
		if _, ok := r.CurrentRatings[p.SessionID]; !ok {
			return false
		}
	}
	return true
}

func (r *LobbyRoom) handleEndSession(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || !player.IsHost {
		r.mu.Unlock()
		return
	}
	r.Status = "finished"
	r.CurrentSong = nil
	r.CurrentRatings = make(map[string]float64)
	r.PriorScores = make(map[string]float64)
	r.SkipVotes = make(map[string]bool)
	r.mu.Unlock()
	r.broadcastPersonalizedState()
}

func (r *LobbyRoom) handleResetToLobby(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || !player.IsHost {
		r.mu.Unlock()
		return
	}
	r.Status = "lobby"
	r.CurrentSong = nil
	r.CurrentRatings = make(map[string]float64)
	r.PriorScores = make(map[string]float64)
	r.SkipVotes = make(map[string]bool)
	r.Queue = make([]queuedSong, 0)
	r.SongsRated = 0
	r.mu.Unlock()
	r.broadcastPersonalizedState()
}

func (r *LobbyRoom) handleChat(ev *ChatEvent) {
	r.mu.RLock()
	player, exists := r.Players[ev.SessionID]
	r.mu.RUnlock()
	if !exists || player.Offline || ev.Text == "" {
		return
	}
	text := ev.Text
	if len(text) > 300 {
		text = text[:300]
	}
	r.broadcast("chat_message", map[string]interface{}{
		"sender":    player.Nickname,
		"text":      text,
		"type":      "user",
		"timestamp": time.Now(),
	})
}

func (r *LobbyRoom) handleTransferHost(ev *TransferHostEvent) {
	r.mu.Lock()
	player, exists := r.Players[ev.SessionID]
	target, targetExists := r.Players[ev.TargetSessionID]
	if !exists || !player.IsHost || !targetExists || target.Offline || target.IsSpectator {
		r.mu.Unlock()
		return
	}
	player.IsHost = false
	target.IsHost = true
	targetNick := target.Nickname
	r.mu.Unlock()

	r.broadcastPersonalizedState()
	r.broadcast("chat_message", map[string]interface{}{
		"sender":    "System",
		"text":      fmt.Sprintf("%s is now the host", targetNick),
		"type":      "system",
		"timestamp": time.Now(),
	})
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

	r.broadcast("room_closed", nil)
	r.Close()
}

func (r *LobbyRoom) Close() {
	r.mu.Lock()
	if r.Closed && r.EventChan == nil {
		r.mu.Unlock()
		return
	}
	r.Closed = true
	ch := r.EventChan
	r.EventChan = nil
	onDestroy := r.OnDestroy
	roomID := r.RoomID
	r.mu.Unlock()

	if ch != nil {
		close(ch)
	}
	if onDestroy != nil {
		onDestroy(roomID)
	}
}

func (r *LobbyRoom) cleanupOfflinePlayers() {
	r.mu.Lock()
	now := time.Now()
	changed := false
	for sid, p := range r.Players {
		if p.Offline && p.OfflineSince != nil && now.Sub(*p.OfflineSince) > 60*time.Second {
			delete(r.Players, sid)
			delete(r.Conns, sid)
			delete(r.CurrentRatings, sid)
			delete(r.SkipVotes, sid)
			changed = true
		}
	}
	if changed {
		r.ensureHostActive()
	}
	r.mu.Unlock()
	if changed {
		r.broadcastPersonalizedState()
	}
}

func (r *LobbyRoom) ensureHostActive() {
	var host *domain.RatePlayer
	var firstOnline *domain.RatePlayer
	for _, p := range r.Players {
		if p.IsHost {
			host = p
		}
		if firstOnline == nil && !p.Offline && !p.IsSpectator {
			firstOnline = p
		}
	}
	if host != nil && !host.Offline {
		return
	}
	if host != nil {
		host.IsHost = false
	}
	if firstOnline != nil {
		firstOnline.IsHost = true
	}
}

func (r *LobbyRoom) ShouldDestroy() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.Closed {
		return true
	}
	for _, p := range r.Players {
		if !p.Offline {
			return false
		}
	}
	if len(r.Players) == 0 {
		return time.Since(r.CreatedAt) > 3*time.Minute
	}
	return time.Since(r.LastActive) > 5*time.Minute
}

func (r *LobbyRoom) GetRoomInfo() domain.RateRoomInfo {
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
	return domain.RateRoomInfo{
		RoomID:         r.RoomID,
		Name:           r.Config.Name,
		HostNickname:   hostNick,
		PlayerCount:    playerCount,
		SpectatorCount: spectatorCount,
		Status:         r.Status,
		Private:        r.Config.Private,
		QueueMode:      r.Config.QueueMode,
		RevealMode:     r.Config.RevealMode,
		QueueLength:    len(r.Queue),
		SourceMode:     r.Config.SourceMode,
		PoolYear:       r.Config.PoolYear,
		PoolSeason:     r.Config.PoolSeason,
		PoolThemeType:  r.Config.PoolThemeType,
		PoolFormat:     r.Config.PoolFormat,
	}
}

func (r *LobbyRoom) loadSeasonalPoolAsync() {
	r.mu.RLock()
	year := r.Config.PoolYear
	season := r.Config.PoolSeason
	themeType := r.Config.PoolThemeType
	format := r.Config.PoolFormat
	limit := r.Config.PoolLimit
	r.mu.RUnlock()

	go func() {
		ctx := context.Background()
		statusTrue := true
		filters := domain.SongFilters{
			Year:   year,
			Season: season,
			Status: &statusTrue,
			Sort:   "views",
		}
		if themeType != "" && themeType != PoolThemeAll {
			filters.Type = themeType
		}
		if format != "" && format != PoolFormatAll {
			filters.Format = format
		}
		if limit <= 0 {
			limit = UnlimitedPoolFetchLimit
		}

		songs, err := r.SongRepo.GetPaginated(ctx, limit, 0, filters)
		if err != nil {
			r.SendEvent(RoomEvent{Type: EvPoolLoaded, Data: &PoolLoadedEvent{Error: "Failed to load seasonal pool"}})
			return
		}
		if len(songs) == 0 {
			r.SendEvent(RoomEvent{Type: EvPoolLoaded, Data: &PoolLoadedEvent{Error: "No themes found for that season"}})
			return
		}

		enriched := make([]domain.Song, 0, len(songs))
		seen := make(map[string]struct{}, len(songs))
		for i := range songs {
			s := songs[i]
			if s.UUID == "" {
				continue
			}
			if _, ok := seen[s.UUID]; ok {
				continue
			}
			seen[s.UUID] = struct{}{}
			full, loadErr := r.loadSong(s.UUID)
			if loadErr != nil || full == nil {
				continue
			}
			enriched = append(enriched, *full)
		}

		if len(enriched) == 0 {
			r.SendEvent(RoomEvent{Type: EvPoolLoaded, Data: &PoolLoadedEvent{
				Error: "No playable themes found for that season (inactive or missing video)",
			}})
			return
		}

		rand.Shuffle(len(enriched), func(i, j int) {
			enriched[i], enriched[j] = enriched[j], enriched[i]
		})

		r.SendEvent(RoomEvent{Type: EvPoolLoaded, Data: &PoolLoadedEvent{Songs: enriched}})
	}()
}

func (r *LobbyRoom) handlePoolLoaded(ev *PoolLoadedEvent) {
	if ev.Error != "" {
		r.broadcast("chat_message", map[string]interface{}{
			"sender":    "System",
			"text":      ev.Error,
			"type":      "system",
			"timestamp": time.Now(),
		})
		r.mu.Lock()
		r.Status = "lobby"
		r.mu.Unlock()
		r.broadcastPersonalizedState()
		return
	}

	r.mu.RLock()
	seasonal := isSeasonalPool(r.Config)
	status := r.Status
	cfg := r.Config
	r.mu.RUnlock()
	if !seasonal || status != "waiting" {
		return
	}

	queue := make([]queuedSong, 0, len(ev.Songs))
	for i := range ev.Songs {
		s := ev.Songs[i]
		item := r.buildQueueItemFromPool(&s, cfg)
		queue = append(queue, queuedSong{Item: item, Song: &ev.Songs[i]})
	}

	r.mu.Lock()
	r.Queue = queue
	r.mu.Unlock()

	label := fmt.Sprintf("%s %s", cfg.PoolSeason, cfg.PoolYear)
	if cfg.PoolThemeType != "" && cfg.PoolThemeType != PoolThemeAll {
		label = fmt.Sprintf("%s (%s)", label, cfg.PoolThemeType)
	}
	if cfg.PoolFormat != "" && cfg.PoolFormat != PoolFormatAll {
		label = fmt.Sprintf("%s · %s", label, cfg.PoolFormat)
	}
	r.broadcastPersonalizedState()
	r.broadcast("chat_message", map[string]interface{}{
		"sender":    "System",
		"text":      fmt.Sprintf("Loaded %d themes from %s — use Next to start", len(queue), label),
		"type":      "system",
		"timestamp": time.Now(),
	})
}

func (r *LobbyRoom) buildQueueItemFromPool(song *domain.Song, cfg domain.RateConfig) domain.RateQueueItem {
	name := song.Name
	if name == "" && song.SongRomaji != nil {
		name = *song.SongRomaji
	}
	if name == "" && song.SongEN != nil {
		name = *song.SongEN
	}
	if name == "" {
		name = "Unknown Theme"
	}
	animeTitle := ""
	animeSlug := ""
	coverURL := ""
	if song.Anime != nil {
		animeTitle = song.Anime.Title
		animeSlug = song.Anime.Slug
		if song.Anime.Cover != nil {
			coverURL = r.MediaService.GetURL(*song.Anime.Cover)
		}
	}
	themeLabel := song.Type + song.ThemeNum
	return domain.RateQueueItem{
		ItemID:           generateUUID(),
		SongUUID:         song.UUID,
		SongName:         name,
		AnimeTitle:       animeTitle,
		AnimeSlug:        animeSlug,
		ThemeLabel:       themeLabel,
		CoverURL:         coverURL,
		AddedBySessionID: "",
		AddedByUserUUID:  "",
		AddedByNickname:  fmt.Sprintf("Pool · %s %s", cfg.PoolSeason, cfg.PoolYear),
	}
}

func (r *LobbyRoom) loadSong(songUUID string) (*domain.Song, error) {
	ctx := context.Background()
	song, err := r.SongRepo.GetByUUID(ctx, songUUID)
	if err != nil {
		return nil, err
	}
	variants, err := r.SongRepo.GetVariantsBySongID(ctx, song.ID)
	if err == nil {
		song.Variants = variants
	}
	artists, err := r.SongRepo.GetArtistsBySongID(ctx, song.ID, false)
	if err == nil {
		song.Artists = artists
	}

	audioURL := resolvePlayableAudioURL(r.MediaService, song)
	if err := validateSongForRate(song, audioURL); err != nil {
		return nil, err
	}
	return song, nil
}

func (r *LobbyRoom) buildQueueItem(song *domain.Song, player *domain.RatePlayer) domain.RateQueueItem {
	name := song.Name
	if name == "" && song.SongRomaji != nil {
		name = *song.SongRomaji
	}
	if name == "" && song.SongEN != nil {
		name = *song.SongEN
	}
	if name == "" {
		name = "Unknown Theme"
	}
	animeTitle := ""
	animeSlug := ""
	coverURL := ""
	if song.Anime != nil {
		animeTitle = song.Anime.Title
		animeSlug = song.Anime.Slug
		if song.Anime.Cover != nil {
			coverURL = r.MediaService.GetURL(*song.Anime.Cover)
		}
	}
	themeLabel := song.Type + song.ThemeNum
	return domain.RateQueueItem{
		ItemID:           generateUUID(),
		SongUUID:         song.UUID,
		SongName:         name,
		AnimeTitle:       animeTitle,
		AnimeSlug:        animeSlug,
		ThemeLabel:       themeLabel,
		CoverURL:         coverURL,
		AddedBySessionID: player.SessionID,
		AddedByUserUUID:  player.UserUUID,
		AddedByNickname:  player.Nickname,
	}
}

func (r *LobbyRoom) resolveAudioURL(song *domain.Song) string {
	return resolvePlayableAudioURL(r.MediaService, song)
}

func (r *LobbyRoom) buildSongDTO(song *domain.Song) dto.SongMinimalDTO {
	songDTO := dto.ToSongMinimalDTO(song)
	if song.Anime != nil {
		if song.Anime.Cover != nil {
			resolved := r.MediaService.GetURL(*song.Anime.Cover)
			songDTO.Anime.CoverUrl = resolved
		}
		if song.Anime.Banner != nil {
			resolved := r.MediaService.Resolve(song.Anime.Banner)
			songDTO.Anime.BannerUrl = resolved
		}
	}
	return songDTO
}

func (r *LobbyRoom) buildSongStartedPayload(song *domain.Song) map[string]interface{} {
	return map[string]interface{}{
		"song":      r.buildSongDTO(song),
		"audio_url": r.resolveAudioURL(song),
	}
}

func (r *LobbyRoom) buildRatingSnapshot(forSessionID string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	live := r.Config.RevealMode == RevealModeLive
	ratings := make(map[string]map[string]interface{})
	var sum float64
	ratedCount := 0
	eligible := 0

	for _, p := range r.Players {
		if p.IsSpectator {
			continue
		}
		if p.UserUUID != "" && !p.Offline {
			eligible++
		}
		entry := map[string]interface{}{"rated": false}
		if score, ok := r.CurrentRatings[p.SessionID]; ok {
			entry["rated"] = true
			ratedCount++
			sum += score
			if live || p.SessionID == forSessionID {
				entry["score"] = score
			}
		}
		ratings[p.SessionID] = entry
	}

	var sessionAvg *float64
	if ratedCount > 0 {
		avg := sum / float64(ratedCount)
		sessionAvg = &avg
	}

	var myScore *float64
	if score, ok := r.CurrentRatings[forSessionID]; ok {
		s := score
		myScore = &s
	}

	var priorScore *float64
	if myScore == nil {
		if prior, ok := r.PriorScores[forSessionID]; ok {
			p := prior
			priorScore = &p
		}
	}

	songUUID := ""
	if r.CurrentSong != nil {
		songUUID = r.CurrentSong.UUID
	}

	return map[string]interface{}{
		"song_uuid":     songUUID,
		"session_avg":   sessionAvg,
		"rated_count":   ratedCount,
		"player_count":  eligible,
		"ratings":       ratings,
		"my_score":      myScore,
		"prior_score":   priorScore,
		"reveal_mode":   r.Config.RevealMode,
	}
}

func (r *LobbyRoom) getRoomStatePayload(forSessionID string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	playersList := make([]domain.RatePlayer, 0)
	spectatorsList := make([]domain.RatePlayer, 0)
	for _, p := range r.Players {
		// copy without leaking UserID (already json:"-")
		cp := *p
		if p.IsSpectator {
			spectatorsList = append(spectatorsList, cp)
		} else {
			playersList = append(playersList, cp)
		}
	}

	queueItems := make([]domain.RateQueueItem, 0, len(r.Queue))
	for _, q := range r.Queue {
		queueItems = append(queueItems, q.Item)
	}

	var currentSong interface{}
	var audioURL string
	if r.CurrentSong != nil && (r.Status == "rating" || r.Status == "waiting") {
		// unlock temporarily not needed — buildSongDTO only reads song + media
		dtoSong := dto.ToSongMinimalDTO(r.CurrentSong)
		if r.CurrentSong.Anime != nil {
			if r.CurrentSong.Anime.Cover != nil {
				resolved := r.MediaService.GetURL(*r.CurrentSong.Anime.Cover)
				dtoSong.Anime.CoverUrl = resolved
			}
			if r.CurrentSong.Anime.Banner != nil {
				resolved := r.MediaService.Resolve(r.CurrentSong.Anime.Banner)
				dtoSong.Anime.BannerUrl = resolved
			}
		}
		currentSong = dtoSong
		audioURL = r.resolveAudioURL(r.CurrentSong)
	}

	// rating snapshot without re-locking: inline
	live := r.Config.RevealMode == RevealModeLive
	ratings := make(map[string]map[string]interface{})
	var sum float64
	ratedCount := 0
	eligible := 0
	for _, p := range r.Players {
		if p.IsSpectator {
			continue
		}
		if p.UserUUID != "" && !p.Offline {
			eligible++
		}
		entry := map[string]interface{}{"rated": false}
		if score, ok := r.CurrentRatings[p.SessionID]; ok {
			entry["rated"] = true
			ratedCount++
			sum += score
			if live || p.SessionID == forSessionID {
				entry["score"] = score
			}
		}
		ratings[p.SessionID] = entry
	}
	var sessionAvg *float64
	if ratedCount > 0 {
		avg := sum / float64(ratedCount)
		sessionAvg = &avg
	}
	var myScore *float64
	if score, ok := r.CurrentRatings[forSessionID]; ok {
		s := score
		myScore = &s
	}
	var priorScore *float64
	if myScore == nil {
		if prior, ok := r.PriorScores[forSessionID]; ok {
			p := prior
			priorScore = &p
		}
	}

	songUUID := ""
	if r.CurrentSong != nil {
		songUUID = r.CurrentSong.UUID
	}

	return map[string]interface{}{
		"room_id":    r.RoomID,
		"status":     r.Status,
		"config":     r.Config,
		"players":    playersList,
		"spectators": spectatorsList,
		"queue":      queueItems,
		"current_song": currentSong,
		"audio_url":    audioURL,
		"songs_rated":  r.SongsRated,
		"rating_data": map[string]interface{}{
			"song_uuid":    songUUID,
			"session_avg":  sessionAvg,
			"rated_count":  ratedCount,
			"player_count": eligible,
			"ratings":      ratings,
			"my_score":     myScore,
			"prior_score":  priorScore,
			"reveal_mode":  r.Config.RevealMode,
		},
		"skip_vote":     r.buildSkipVoteSnapshotLocked(forSessionID),
		"my_session_id": forSessionID,
	}
}

func (r *LobbyRoom) buildSkipVoteSnapshotLocked(forSessionID string) map[string]interface{} {
	if !r.Config.VoteSkip {
		return map[string]interface{}{
			"enabled":  false,
			"count":    0,
			"needed":   0,
			"my_voted": false,
		}
	}
	eligible := 0
	count := 0
	for _, p := range r.Players {
		if p.IsSpectator || p.Offline {
			continue
		}
		eligible++
		if r.SkipVotes != nil && r.SkipVotes[p.SessionID] {
			count++
		}
	}
	needed := 0
	if eligible > 0 {
		needed = eligible/2 + 1
	}
	myVoted := r.SkipVotes != nil && r.SkipVotes[forSessionID]
	return map[string]interface{}{
		"enabled":  true,
		"count":    count,
		"needed":   needed,
		"my_voted": myVoted,
	}
}

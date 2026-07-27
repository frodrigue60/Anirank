package amq

import (
	"math/rand"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/dto"
)

type SelectCandidateEvent struct {
	SessionID string
	SongUUID  string
}

const (
	defaultSaveVoteSeconds    = 10
	maxSaveVoteSeconds        = 60
	defaultMediaBufferSeconds = 8
)

func normalizeSaveVoteSeconds(voteSeconds *int) *int {
	if voteSeconds == nil {
		v := defaultSaveVoteSeconds
		return &v
	}
	v := *voteSeconds
	if v < 0 {
		v = defaultSaveVoteSeconds
	}
	if v > maxSaveVoteSeconds {
		v = maxSaveVoteSeconds
	}
	return &v
}

func saveVoteSecondsValue(voteSeconds *int) int {
	if voteSeconds == nil {
		return defaultSaveVoteSeconds
	}
	return *voteSeconds
}

func (r *LobbyRoom) effectiveVoteSeconds() int {
	return saveVoteSecondsValue(r.Config.VoteSeconds)
}

func (r *LobbyRoom) startSaveRound() {
	r.mu.Lock()
	if r.CurrentRound >= len(r.SaveRounds) || r.CurrentRound >= r.Config.MaxRounds {
		r.mu.Unlock()
		r.endGame()
		return
	}

	saveRound := r.SaveRounds[r.CurrentRound]
	r.Status = "playing"
	r.RoundPhase = "media_buffer"
	r.PreviewIndex = 0
	r.WinnerPlayIndex = 0
	r.SaveCandidates = saveRound.Candidates
	r.RoundWinners = nil
	r.RoundVoteCounts = make(map[string]int)
	r.StartPercents = make([]float64, len(saveRound.Candidates))
	for i := range r.StartPercents {
		r.StartPercents[i] = rand.Float64() * 0.55
	}

	for _, p := range r.Players {
		p.SelectedSongUUID = ""
	}

	currentRound := r.CurrentRound + 1
	maxRounds := r.Config.MaxRounds
	previewSeconds := r.Config.PreviewSeconds
	r.mu.Unlock()

	r.broadcastSaveRoundStart(currentRound, maxRounds, previewSeconds, &saveRound)
	r.scheduleMediaBufferTimeout()
}

func (r *LobbyRoom) scheduleMediaBufferTimeout() {
	r.mu.Lock()
	if r.RoundPhase != "media_buffer" {
		r.mu.Unlock()
		return
	}
	bufferSecs := defaultMediaBufferSeconds
	r.TimerType = "media_buffer"
	r.TimerStart = time.Now()
	r.TimerDuration = time.Duration(bufferSecs) * time.Second
	if r.Timer != nil {
		r.Timer.Stop()
	}
	r.Timer = time.AfterFunc(r.TimerDuration, func() {
		r.SendEvent(RoomEvent{Type: EvTimerExpired, Data: "media_buffer"})
	})
	r.mu.Unlock()

	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) beginPreviewSelectAfterBuffer() {
	r.mu.Lock()
	if r.RoundPhase != "media_buffer" {
		r.mu.Unlock()
		return
	}
	r.RoundPhase = "preview_select"
	r.PreviewIndex = 0
	if r.Timer != nil {
		r.Timer.Stop()
	}
	previewSeconds := r.Config.PreviewSeconds
	if previewSeconds <= 0 {
		previewSeconds = 12
	}
	r.mu.Unlock()

	r.broadcast("phase_change", map[string]interface{}{
		"round_phase":     "preview_select",
		"preview_index":   0,
		"preview_seconds": previewSeconds,
	})
	r.schedulePreviewStepTimer()
}

func (r *LobbyRoom) handleMediaReady(ev *MediaReadyEvent) {
	r.mu.Lock()
	player, exists := r.Players[ev.SessionID]
	if !exists || !player.IsHost || player.IsSpectator || player.Offline {
		r.mu.Unlock()
		return
	}
	if r.RoundPhase != "media_buffer" || r.Status != "playing" {
		r.mu.Unlock()
		return
	}
	expectedRound := r.CurrentRound + 1
	if ev.RoundNumber > 0 && ev.RoundNumber != expectedRound {
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()

	r.beginPreviewSelectAfterBuffer()
}

func (r *LobbyRoom) schedulePreviewStepTimer() {
	r.mu.Lock()
	previewSeconds := r.Config.PreviewSeconds
	if previewSeconds <= 0 {
		previewSeconds = 12
	}
	r.TimerType = "preview_step"
	r.TimerStart = time.Now()
	r.TimerDuration = time.Duration(previewSeconds) * time.Second
	if r.Timer != nil {
		r.Timer.Stop()
	}
	r.Timer = time.AfterFunc(r.TimerDuration, func() {
		r.SendEvent(RoomEvent{Type: EvTimerExpired, Data: "preview_step"})
	})
	r.mu.Unlock()

	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) handlePreviewStepExpired() {
	r.mu.Lock()
	if r.RoundPhase != "preview_select" {
		r.mu.Unlock()
		return
	}

	optionCount := len(r.SaveCandidates)
	r.PreviewIndex++

	if r.PreviewIndex < optionCount {
		nextIndex := r.PreviewIndex
		r.mu.Unlock()
		r.schedulePreviewStepTimer()
		r.broadcast("phase_change", map[string]interface{}{
			"round_phase":   "preview_select",
			"preview_index": nextIndex,
		})
		return
	}

	r.mu.Unlock()
	r.startVoteSelectPhase()
}

func (r *LobbyRoom) startVoteSelectPhase() {
	voteSecs := r.effectiveVoteSeconds()
	if voteSecs == 0 {
		r.tallySaveRoundVotes()
		return
	}

	r.mu.Lock()
	r.RoundPhase = "vote_select"
	if r.Timer != nil {
		r.Timer.Stop()
	}
	r.mu.Unlock()

	r.scheduleVoteStepTimer(voteSecs)
	r.broadcast("phase_change", map[string]interface{}{
		"round_phase":  "vote_select",
		"vote_seconds": voteSecs,
	})
	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) scheduleVoteStepTimer(voteSecs int) {
	r.mu.Lock()
	r.TimerType = "vote_step"
	r.TimerStart = time.Now()
	r.TimerDuration = time.Duration(voteSecs) * time.Second
	if r.Timer != nil {
		r.Timer.Stop()
	}
	r.Timer = time.AfterFunc(r.TimerDuration, func() {
		r.SendEvent(RoomEvent{Type: EvTimerExpired, Data: "vote_step"})
	})
	r.mu.Unlock()

	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) handleVoteStepExpired() {
	r.mu.Lock()
	if r.RoundPhase != "vote_select" {
		r.mu.Unlock()
		return
	}
	r.mu.Unlock()
	r.tallySaveRoundVotes()
}

func (r *LobbyRoom) tallySaveRoundVotes() {
	r.mu.Lock()
	if r.RoundVoteCounts == nil {
		r.RoundVoteCounts = make(map[string]int)
	}
	for _, p := range r.Players {
		if p.IsSpectator || p.Offline || p.SelectedSongUUID == "" {
			continue
		}
		r.RoundVoteCounts[p.SelectedSongUUID]++
	}

	validUUIDs := make(map[string]bool, len(r.SaveCandidates))
	for _, s := range r.SaveCandidates {
		validUUIDs[s.UUID] = true
	}

	maxVotes := 0
	for uuid, count := range r.RoundVoteCounts {
		if !validUUIDs[uuid] {
			delete(r.RoundVoteCounts, uuid)
			continue
		}
		if count > maxVotes {
			maxVotes = count
		}
	}

	winners := make([]string, 0)
	if maxVotes > 0 {
		for uuid, count := range r.RoundVoteCounts {
			if count == maxVotes {
				winners = append(winners, uuid)
			}
		}
	}

	r.RoundWinners = winners
	r.RoundPhase = "winner_playback"
	r.WinnerPlayIndex = 0
	r.PreviewIndex = 0

	if r.Timer != nil {
		r.Timer.Stop()
	}

	voteSnapshot := make(map[string]int, len(r.RoundVoteCounts))
	for k, v := range r.RoundVoteCounts {
		voteSnapshot[k] = v
	}
	winnersSnapshot := append([]string(nil), r.RoundWinners...)
	r.mu.Unlock()

	r.broadcast("round_results", map[string]interface{}{
		"votes":      voteSnapshot,
		"winners":    winnersSnapshot,
		"next_round": r.buildNextSaveRoundPreviewPayload(),
	})
	r.startWinnerPlaybackStep()
}

func (r *LobbyRoom) startWinnerPlaybackStep() {
	r.mu.Lock()
	if r.RoundPhase != "winner_playback" {
		r.mu.Unlock()
		return
	}

	if len(r.RoundWinners) == 0 {
		r.mu.Unlock()
		r.finishSaveRound()
		return
	}

	if r.WinnerPlayIndex >= len(r.RoundWinners) {
		r.mu.Unlock()
		r.finishSaveRound()
		return
	}

	previewSeconds := r.Config.PreviewSeconds
	if previewSeconds <= 0 {
		previewSeconds = 12
	}
	r.TimerType = "winner_step"
	r.TimerStart = time.Now()
	r.TimerDuration = time.Duration(previewSeconds) * time.Second
	if r.Timer != nil {
		r.Timer.Stop()
	}
	winnerPlayIndex := r.WinnerPlayIndex
	voteSnapshot := make(map[string]int, len(r.RoundVoteCounts))
	for k, v := range r.RoundVoteCounts {
		voteSnapshot[k] = v
	}
	winnersSnapshot := append([]string(nil), r.RoundWinners...)
	r.Timer = time.AfterFunc(r.TimerDuration, func() {
		r.SendEvent(RoomEvent{Type: EvTimerExpired, Data: "winner_step"})
	})
	r.mu.Unlock()

	r.broadcast("phase_change", map[string]interface{}{
		"round_phase":       "winner_playback",
		"winner_play_index": winnerPlayIndex,
		"votes":             voteSnapshot,
		"winners":           winnersSnapshot,
		"next_round":        r.buildNextSaveRoundPreviewPayload(),
	})
	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) handleWinnerStepExpired() {
	r.mu.Lock()
	if r.RoundPhase != "winner_playback" {
		r.mu.Unlock()
		return
	}
	r.WinnerPlayIndex++
	done := r.WinnerPlayIndex >= len(r.RoundWinners)
	r.mu.Unlock()

	if done {
		r.finishSaveRound()
		return
	}
	r.startWinnerPlaybackStep()
}

func (r *LobbyRoom) finishSaveRound() {
	r.mu.Lock()
	r.recordSaveRoundResultLocked()
	r.CurrentRound++
	currentRound := r.CurrentRound
	maxRounds := r.Config.MaxRounds
	r.RoundPhase = ""
	r.SaveCandidates = nil
	r.RoundWinners = nil
	r.RoundVoteCounts = nil
	if r.Timer != nil {
		r.Timer.Stop()
	}
	r.mu.Unlock()

	if currentRound >= maxRounds || currentRound >= len(r.SaveRounds) {
		r.endGame()
		return
	}
	r.startSaveRound()
}

func (r *LobbyRoom) handleSelectCandidate(ev *SelectCandidateEvent) {
	r.mu.Lock()
	if (r.RoundPhase != "preview_select" && r.RoundPhase != "vote_select") || r.Status != "playing" {
		r.mu.Unlock()
		return
	}

	player, exists := r.Players[ev.SessionID]
	if !exists || player.IsSpectator || player.Offline {
		r.mu.Unlock()
		return
	}

	if ev.SongUUID == "" {
		player.SelectedSongUUID = ""
		r.mu.Unlock()
		r.broadcast("lobby_state_update", r.getRoomStatePayload())
		return
	}

	valid := false
	for _, s := range r.SaveCandidates {
		if s.UUID == ev.SongUUID {
			valid = true
			break
		}
	}
	if !valid {
		r.mu.Unlock()
		return
	}

	if player.SelectedSongUUID == ev.SongUUID {
		player.SelectedSongUUID = ""
	} else {
		player.SelectedSongUUID = ev.SongUUID
	}
	r.mu.Unlock()
	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) handleSkipSavePlayback(sessionID string) {
	r.mu.Lock()
	player, exists := r.Players[sessionID]
	if !exists || !player.IsHost || r.RoundPhase != "winner_playback" {
		r.mu.Unlock()
		return
	}
	if r.Timer != nil {
		r.Timer.Stop()
	}
	r.mu.Unlock()
	r.finishSaveRound()
}

func (r *LobbyRoom) broadcastSaveRoundStart(currentRound, maxRounds, previewSeconds int, saveRound *domain.AMQSaveRound) {
	payload := map[string]interface{}{
		"current_round":        currentRound,
		"max_rounds":           maxRounds,
		"game_type":            r.Config.GameType,
		"round_phase":          "media_buffer",
		"preview_index":        0,
		"preview_seconds":      previewSeconds,
		"media_buffer_seconds": defaultMediaBufferSeconds,
		"vote_seconds":         r.effectiveVoteSeconds(),
		"theme_label":      saveRound.ThemeLabel,
		"round_theme_type": saveRound.RoundThemeType,
		"is_fallback":      saveRound.IsFallback,
		"candidates":      r.buildSaveCandidatesPayload(saveRound.Candidates, nil, nil),
		"winners":         []string{},
	}
	r.broadcast("round_start", payload)
	r.broadcast("lobby_state_update", r.getRoomStatePayload())
}

func (r *LobbyRoom) buildNextSaveRoundPreviewPayload() map[string]interface{} {
	nextIdx := r.CurrentRound + 1
	if nextIdx >= len(r.SaveRounds) || nextIdx >= r.Config.MaxRounds {
		return nil
	}
	saveRound := r.SaveRounds[nextIdx]
	candidates := make([]map[string]interface{}, 0, len(saveRound.Candidates))
	for _, song := range saveRound.Candidates {
		animeTitle := ""
		if song.Anime != nil {
			animeTitle = truncateAnimeTitle(song.Anime.Title, 32)
		}
		candidates = append(candidates, map[string]interface{}{
			"song_uuid":   song.UUID,
			"audio_url":   r.resolveAudioURL(&song),
			"anime_title": animeTitle,
			"theme_label": songThemeLabel(&song),
		})
	}
	return map[string]interface{}{
		"round_number": nextIdx + 1,
		"candidates":   candidates,
	}
}

func (r *LobbyRoom) buildSaveCandidatesPayload(candidates []domain.Song, voteCounts map[string]int, winners []string) []map[string]interface{} {
	winnerSet := make(map[string]bool, len(winners))
	for _, w := range winners {
		winnerSet[w] = true
	}

	out := make([]map[string]interface{}, 0, len(candidates))
	for i, song := range candidates {
		animeTitle := ""
		if song.Anime != nil {
			animeTitle = truncateAnimeTitle(song.Anime.Title, 32)
		}
		count := 0
		if voteCounts != nil {
			count = voteCounts[song.UUID]
		}
		startPercent := 0.0
		if i < len(r.StartPercents) {
			startPercent = r.StartPercents[i]
		}
		out = append(out, map[string]interface{}{
			"song_uuid":     song.UUID,
			"audio_url":     r.resolveAudioURL(&song),
			"start_percent": startPercent,
			"anime_title":   animeTitle,
			"theme_label":   songThemeLabel(&song),
			"vote_count":    count,
			"is_winner":     winnerSet[song.UUID],
		})
	}
	return out
}

func (r *LobbyRoom) buildSaveRoundDataPayload() map[string]interface{} {
	if len(r.SaveCandidates) == 0 {
		return nil
	}

	var saveRound *domain.AMQSaveRound
	if r.CurrentRound < len(r.SaveRounds) {
		saveRound = &r.SaveRounds[r.CurrentRound]
	}

	themeLabel := "Save Round"
	roundThemeType := ""
	isFallback := false
	if saveRound != nil {
		themeLabel = saveRound.ThemeLabel
		roundThemeType = saveRound.RoundThemeType
		isFallback = saveRound.IsFallback
	}

	activeWinnerUUID := ""
	if r.RoundPhase == "winner_playback" && r.WinnerPlayIndex < len(r.RoundWinners) {
		activeWinnerUUID = r.RoundWinners[r.WinnerPlayIndex]
	}

	return map[string]interface{}{
		"current_round":       r.CurrentRound + 1,
		"max_rounds":          r.Config.MaxRounds,
		"game_type":           r.Config.GameType,
		"round_phase":         r.RoundPhase,
		"preview_index":       r.PreviewIndex,
		"winner_play_index":   r.WinnerPlayIndex,
		"preview_seconds":     r.Config.PreviewSeconds,
		"media_buffer_seconds": defaultMediaBufferSeconds,
		"vote_seconds":        r.effectiveVoteSeconds(),
		"theme_label":         themeLabel,
		"round_theme_type":    roundThemeType,
		"is_fallback":         isFallback,
		"candidates":          r.buildSaveCandidatesPayload(r.SaveCandidates, r.RoundVoteCounts, r.RoundWinners),
		"winners":             r.RoundWinners,
		"active_winner_uuid":  activeWinnerUUID,
		"audio_url":           r.resolveSaveActiveAudioURL(),
		"start_percent":       r.resolveSaveActiveStartPercent(),
	}
}

func (r *LobbyRoom) resolveSaveActiveAudioURL() string {
	if len(r.SaveCandidates) == 0 {
		return ""
	}

	if r.RoundPhase == "winner_playback" && r.WinnerPlayIndex < len(r.RoundWinners) {
		targetUUID := r.RoundWinners[r.WinnerPlayIndex]
		for i := range r.SaveCandidates {
			if r.SaveCandidates[i].UUID == targetUUID {
				return r.resolveAudioURL(&r.SaveCandidates[i])
			}
		}
	}

	if r.PreviewIndex >= 0 && r.PreviewIndex < len(r.SaveCandidates) {
		return r.resolveAudioURL(&r.SaveCandidates[r.PreviewIndex])
	}
	return r.resolveAudioURL(&r.SaveCandidates[0])
}

func (r *LobbyRoom) resolveSaveActiveStartPercent() float64 {
	if len(r.SaveCandidates) == 0 {
		return 0
	}

	idx := r.PreviewIndex
	if r.RoundPhase == "winner_playback" && r.WinnerPlayIndex < len(r.RoundWinners) {
		targetUUID := r.RoundWinners[r.WinnerPlayIndex]
		for i := range r.SaveCandidates {
			if r.SaveCandidates[i].UUID == targetUUID {
				idx = i
				break
			}
		}
	}
	if idx >= 0 && idx < len(r.StartPercents) {
		return r.StartPercents[idx]
	}
	return 0
}

func (r *LobbyRoom) resetSaveStateLocked() {
	r.SaveRounds = nil
	r.SaveCandidates = nil
	r.RoundPhase = ""
	r.PreviewIndex = 0
	r.WinnerPlayIndex = 0
	r.RoundWinners = nil
	r.RoundVoteCounts = nil
	r.StartPercents = nil
	r.SaveRoundHistory = nil
}

func sanitizeSaveConfig(cfg *domain.AMQConfig) {
	if cfg.MaxRounds < 5 {
		cfg.MaxRounds = 5
	}
	if cfg.MaxRounds > 30 {
		cfg.MaxRounds = 30
	}
	if cfg.PreviewSeconds < 10 {
		cfg.PreviewSeconds = 12
	}
	if cfg.PreviewSeconds > 15 {
		cfg.PreviewSeconds = 15
	}
	if cfg.GameType != "save-4" && cfg.GameType != "save-6" {
		cfg.GameType = "save-4"
	}
	if cfg.ThemeDistribution != "balanced" {
		cfg.ThemeDistribution = "random"
	}
	cfg.VoteSeconds = normalizeSaveVoteSeconds(cfg.VoteSeconds)
	cfg.PersonalizedPool = false
}

func (r *LobbyRoom) recordSaveRoundResultLocked() {
	if len(r.SaveCandidates) == 0 {
		return
	}

	var saveRound *domain.AMQSaveRound
	if r.CurrentRound < len(r.SaveRounds) {
		saveRound = &r.SaveRounds[r.CurrentRound]
	}

	themeLabel := "Save Round"
	roundThemeType := ""
	isFallback := false
	if saveRound != nil {
		themeLabel = saveRound.ThemeLabel
		roundThemeType = saveRound.RoundThemeType
		isFallback = saveRound.IsFallback
	}

	winnerSet := make(map[string]bool, len(r.RoundWinners))
	for _, w := range r.RoundWinners {
		winnerSet[w] = true
	}

	voteSnapshot := make(map[string]int, len(r.RoundVoteCounts))
	for k, v := range r.RoundVoteCounts {
		voteSnapshot[k] = v
	}

	candidates := make([]domain.AMQSaveRoundCandidateSummary, 0, len(r.SaveCandidates))
	for _, song := range r.SaveCandidates {
		animeTitle := ""
		if song.Anime != nil {
			animeTitle = truncateAnimeTitle(song.Anime.Title, 48)
		}
		count := 0
		if r.RoundVoteCounts != nil {
			count = r.RoundVoteCounts[song.UUID]
		}
		candidates = append(candidates, domain.AMQSaveRoundCandidateSummary{
			SongUUID:   song.UUID,
			AnimeTitle: animeTitle,
			ThemeLabel: songThemeLabel(&song),
			VoteCount:  count,
			IsWinner:   winnerSet[song.UUID],
		})
	}

	r.SaveRoundHistory = append(r.SaveRoundHistory, domain.AMQSaveRoundResult{
		RoundNumber:    r.CurrentRound + 1,
		ThemeLabel:     themeLabel,
		RoundThemeType: roundThemeType,
		IsFallback:     isFallback,
		Winners:        append([]string(nil), r.RoundWinners...),
		VoteCounts:     voteSnapshot,
		Candidates:     candidates,
	})
}

func buildSaveRoundHistoryPayload(history []domain.AMQSaveRoundResult) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(history))
	for _, h := range history {
		candidates := make([]map[string]interface{}, 0, len(h.Candidates))
		for _, c := range h.Candidates {
			candidates = append(candidates, map[string]interface{}{
				"song_uuid":   c.SongUUID,
				"anime_title": c.AnimeTitle,
				"theme_label": c.ThemeLabel,
				"vote_count":  c.VoteCount,
				"is_winner":   c.IsWinner,
			})
		}
		out = append(out, map[string]interface{}{
			"round_number":     h.RoundNumber,
			"theme_label":      h.ThemeLabel,
			"round_theme_type": h.RoundThemeType,
			"is_fallback":      h.IsFallback,
			"winners":          h.Winners,
			"vote_counts":      h.VoteCounts,
			"candidates":       candidates,
		})
	}
	return out
}

func buildSavePlayedSongsDTO(r *LobbyRoom, limit int) []interface{} {
	var played []interface{}
	for i := 0; i < limit && i < len(r.SaveRounds); i++ {
		sr := r.SaveRounds[i]
		for _, s := range sr.Candidates {
			sDTO := dto.ToSongMinimalDTO(&s)
			if s.Anime != nil && s.Anime.Cover != nil {
				resolved := r.MediaService.GetURL(*s.Anime.Cover)
				sDTO.Anime.CoverUrl = resolved
			}
			played = append(played, sDTO)
		}
	}
	return played
}

package amq

import (
	"testing"

	"anirank/api/internal/domain"
)

func TestPreviewStepAdvancesBeforeTally(t *testing.T) {
	room := newSaveTestRoom(nil)
	room.RoundPhase = "preview_select"
	room.Status = "playing"
	room.SaveCandidates = testSaveSongs(4)
	room.PreviewIndex = 0
	room.RoundVoteCounts = make(map[string]int)

	room.handlePreviewStepExpired()

	if room.PreviewIndex != 1 {
		t.Fatalf("expected preview index 1, got %d", room.PreviewIndex)
	}
	if room.RoundPhase != "preview_select" {
		t.Fatalf("expected still in preview_select, got %q", room.RoundPhase)
	}
}

func TestTallyWithNilVoteCountsMap(t *testing.T) {
	room := setupRoomForTally(t)
	room.RoundVoteCounts = nil
	room.Players["p1"].SelectedSongUUID = "song-uuid-1"

	room.handlePreviewStepExpired()

	if room.RoundVoteCounts == nil {
		t.Fatal("vote counts map must be initialized")
	}
	if room.RoundVoteCounts["song-uuid-1"] != 1 {
		t.Fatalf("expected 1 vote, got %d", room.RoundVoteCounts["song-uuid-1"])
	}
}

func TestHandleSelectCandidateToggleAndValidate(t *testing.T) {
	room := newSaveTestRoom(nil)
	room.RoundPhase = "preview_select"
	room.Status = "playing"
	room.SaveCandidates = testSaveSongs(4)

	room.Players["p1"] = &domain.AMQPlayer{SessionID: "p1", Nickname: "Alice"}
	room.Players["spec"] = &domain.AMQPlayer{SessionID: "spec", IsSpectator: true}

	room.handleSelectCandidate(&SelectCandidateEvent{SessionID: "p1", SongUUID: "song-uuid-2"})
	if room.Players["p1"].SelectedSongUUID != "song-uuid-2" {
		t.Fatalf("expected selection song-uuid-2")
	}

	room.handleSelectCandidate(&SelectCandidateEvent{SessionID: "p1", SongUUID: "song-uuid-2"})
	if room.Players["p1"].SelectedSongUUID != "" {
		t.Fatalf("expected deselect on second click")
	}

	room.handleSelectCandidate(&SelectCandidateEvent{SessionID: "spec", SongUUID: "song-uuid-1"})
	if room.Players["spec"].SelectedSongUUID != "" {
		t.Fatal("spectator must not select")
	}

	room.handleSelectCandidate(&SelectCandidateEvent{SessionID: "p1", SongUUID: "invalid"})
	if room.Players["p1"].SelectedSongUUID != "" {
		t.Fatal("invalid candidate must be ignored")
	}
}

func TestTallyVotesSingleWinner(t *testing.T) {
	room := setupRoomForTally(t)
	room.Players["p1"].SelectedSongUUID = "song-uuid-1"
	room.Players["p2"].SelectedSongUUID = "song-uuid-1"
	room.Players["p3"].SelectedSongUUID = "song-uuid-2"

	room.handlePreviewStepExpired()

	if len(room.RoundWinners) != 1 || room.RoundWinners[0] != "song-uuid-1" {
		t.Fatalf("expected song-uuid-1 winner, got %v", room.RoundWinners)
	}
	if room.RoundVoteCounts["song-uuid-1"] != 2 {
		t.Fatalf("expected 2 votes, got %d", room.RoundVoteCounts["song-uuid-1"])
	}
	if room.RoundPhase != "winner_playback" {
		t.Fatalf("expected winner_playback, got %q", room.RoundPhase)
	}
}

func TestTallyVotesTie(t *testing.T) {
	room := setupRoomForTally(t)
	room.Players["p1"].SelectedSongUUID = "song-uuid-1"
	room.Players["p2"].SelectedSongUUID = "song-uuid-2"

	room.handlePreviewStepExpired()

	if len(room.RoundWinners) != 2 {
		t.Fatalf("expected 2 tied winners, got %v", room.RoundWinners)
	}
}

func TestTallyVotesZeroVotesSkipsWinnerPlayback(t *testing.T) {
	room := setupRoomForTally(t)

	room.handlePreviewStepExpired()

	if len(room.RoundWinners) != 0 {
		t.Fatalf("zero votes must not produce winners, got %v", room.RoundWinners)
	}
	if room.RoundPhase != "" {
		t.Fatalf("expected round to finish without winner playback, got phase %q", room.RoundPhase)
	}
	if room.CurrentRound != 1 {
		t.Fatalf("expected round to advance, got currentRound=%d", room.CurrentRound)
	}
}

func TestStartSaveRoundClearsSelections(t *testing.T) {
	room := newSaveTestRoom(nil)
	room.Config.GameType = "save-4"
	room.Config.MaxRounds = 1
	room.Config.PreviewSeconds = 12
	room.SaveRounds = []domain.AMQSaveRound{{
		ThemeLabel:     "Best OP of Test Anime",
		RoundThemeType: "OP",
		Candidates:     testSaveSongs(4),
	}}
	room.Players["p1"] = &domain.AMQPlayer{SessionID: "p1", SelectedSongUUID: "old-vote"}

	room.startSaveRound()

	if room.Players["p1"].SelectedSongUUID != "" {
		t.Fatal("selection must be cleared at round start")
	}
	if room.PreviewIndex != 0 {
		t.Fatalf("preview must start at index 0, got %d", room.PreviewIndex)
	}
	if room.RoundPhase != "preview_select" {
		t.Fatalf("expected preview_select, got %q", room.RoundPhase)
	}
}

func TestOfflinePlayersDoNotVote(t *testing.T) {
	room := setupRoomForTally(t)
	room.Players["p1"].SelectedSongUUID = "song-uuid-1"
	room.Players["p2"].SelectedSongUUID = "song-uuid-2"
	room.Players["p2"].Offline = true

	room.handlePreviewStepExpired()

	if room.RoundVoteCounts["song-uuid-2"] != 0 {
		t.Fatal("offline player vote must not count")
	}
	if len(room.RoundWinners) != 1 || room.RoundWinners[0] != "song-uuid-1" {
		t.Fatalf("expected only online vote to win, got %v", room.RoundWinners)
	}
}

func TestHandleSkipSavePlayback(t *testing.T) {
	room := setupRoomForTally(t)
	room.RoundPhase = "winner_playback"
	room.RoundWinners = []string{"song-uuid-1", "song-uuid-2"}
	room.WinnerPlayIndex = 0
	room.SaveRounds = []domain.AMQSaveRound{{ThemeLabel: "Test", Candidates: testSaveSongs(4)}}
	room.Config.MaxRounds = 1
	room.Players["host"] = &domain.AMQPlayer{SessionID: "host", IsHost: true}

	room.handleSkipSavePlayback("host")

	if room.CurrentRound != 1 {
		t.Fatalf("expected round advanced to 1, got %d", room.CurrentRound)
	}
}

func TestRecordSaveRoundHistory(t *testing.T) {
	room := setupRoomForTally(t)
	room.SaveRounds = []domain.AMQSaveRound{{
		ThemeLabel:     "Best OP of Test",
		RoundThemeType: "OP",
		Candidates:     testSaveSongs(4),
	}}
	room.RoundWinners = []string{"song-uuid-1"}
	room.RoundVoteCounts["song-uuid-1"] = 2

	room.mu.Lock()
	room.recordSaveRoundResultLocked()
	room.mu.Unlock()

	if len(room.SaveRoundHistory) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(room.SaveRoundHistory))
	}
	if room.SaveRoundHistory[0].ThemeLabel != "Best OP of Test" {
		t.Fatalf("unexpected theme label %q", room.SaveRoundHistory[0].ThemeLabel)
	}
	if len(room.SaveRoundHistory[0].Candidates) != 4 {
		t.Fatalf("expected 4 candidates in history, got %d", len(room.SaveRoundHistory[0].Candidates))
	}
}

func setupRoomForTally(t *testing.T) *LobbyRoom {
	t.Helper()
	room := newSaveTestRoom(nil)
	room.RoundPhase = "preview_select"
	room.Status = "playing"
	room.SaveCandidates = testSaveSongs(4)
	room.PreviewIndex = 4 // preview finished
	room.RoundVoteCounts = make(map[string]int)

	room.Players["p1"] = &domain.AMQPlayer{SessionID: "p1"}
	room.Players["p2"] = &domain.AMQPlayer{SessionID: "p2"}
	room.Players["p3"] = &domain.AMQPlayer{SessionID: "p3"}
	return room
}
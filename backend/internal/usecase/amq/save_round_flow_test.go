package amq

import (
	"testing"

	"anirank/api/internal/domain"
)

func TestSaveRoundLifecycle(t *testing.T) {
	room := newSaveTestRoom(nil)
	room.Config.GameType = "save-4"
	room.Config.MaxRounds = 1
	room.Config.PreviewSeconds = 12
	room.SaveRounds = []domain.AMQSaveRound{{
		ThemeLabel:     "Best OP of Test",
		RoundThemeType: "OP",
		Candidates:     testSaveSongs(4),
	}}
	room.Players["p1"] = &domain.AMQPlayer{SessionID: "p1"}
	room.Players["host"] = &domain.AMQPlayer{SessionID: "host", IsHost: true}

	room.startSaveRound()
	if room.RoundPhase != "preview_select" {
		t.Fatalf("expected preview_select, got %q", room.RoundPhase)
	}

	room.handleSelectCandidate(&SelectCandidateEvent{SessionID: "p1", SongUUID: "song-uuid-2"})
	for i := 0; i < 4; i++ {
		room.handlePreviewStepExpired()
	}

	if room.RoundPhase != "winner_playback" {
		t.Fatalf("expected winner_playback after previews, got %q", room.RoundPhase)
	}
	if len(room.RoundWinners) != 1 || room.RoundWinners[0] != "song-uuid-2" {
		t.Fatalf("expected song-uuid-2 winner, got %v", room.RoundWinners)
	}

	room.handleSkipSavePlayback("host")
	if len(room.SaveRoundHistory) != 1 {
		t.Fatalf("expected history after round, got %d entries", len(room.SaveRoundHistory))
	}
	if room.Status != "finished" {
		t.Fatalf("expected finished after single-round game, got %q", room.Status)
	}
}

func TestBuildSaveRoundDataPayloadMidPreview(t *testing.T) {
	room := newSaveTestRoom(nil)
	room.Status = "playing"
	room.RoundPhase = "preview_select"
	room.PreviewIndex = 2
	room.SaveCandidates = testSaveSongs(4)
	room.SaveRounds = []domain.AMQSaveRound{{
		ThemeLabel:     "Best OP of Test",
		RoundThemeType: "OP",
		Candidates:     testSaveSongs(4),
	}}
	room.RoundVoteCounts = make(map[string]int)
	room.StartPercents = []float64{0.1, 0.2, 0.3, 0.4}
	room.Players["p1"] = &domain.AMQPlayer{SessionID: "p1", SelectedSongUUID: "song-uuid-3"}

	payload := room.buildSaveRoundDataPayload()
	if payload == nil {
		t.Fatal("expected round data payload")
	}
	if payload["preview_index"] != 2 {
		t.Fatalf("expected preview_index 2, got %v", payload["preview_index"])
	}
	if payload["round_phase"] != "preview_select" {
		t.Fatalf("expected preview_select, got %v", payload["round_phase"])
	}
}

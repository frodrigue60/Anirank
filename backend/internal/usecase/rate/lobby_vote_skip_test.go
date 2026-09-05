package rate

import (
	"sync"
	"testing"
	"time"

	"anirank/api/internal/domain"
)

func TestVoteSkipMajorityAdvances(t *testing.T) {
	cfg := testRateConfig()
	cfg.VoteSkip = true
	cfg.MaxPlayers = 8
	room := NewLobbyRoom("SKIP", cfg, nil, nil, nil, nil)
	room.Start()
	defer room.Close()

	conn := nopConn{}
	room.handleJoin(&JoinEvent{SessionID: "h", Conn: conn, Nickname: "Host", DeviceID: "dh", User: &domain.User{ID: 1, UUID: "uh", Name: "Host"}})
	room.handleJoin(&JoinEvent{SessionID: "p1", Conn: conn, Nickname: "P1", DeviceID: "d1", User: &domain.User{ID: 2, UUID: "u1", Name: "P1"}})
	room.handleJoin(&JoinEvent{SessionID: "p2", Conn: conn, Nickname: "P2", DeviceID: "d2", User: &domain.User{ID: 3, UUID: "u2", Name: "P2"}})

	room.mu.Lock()
	room.Status = "rating"
	room.CurrentSong = &domain.Song{ID: 99, UUID: "song-1", SongRomaji: strPtr("Theme")}
	room.Queue = []queuedSong{{
		Item: domain.RateQueueItem{ItemID: "q1", SongUUID: "song-2", SongName: "Next"},
		Song: &domain.Song{ID: 100, UUID: "song-2", SongRomaji: strPtr("Next")},
	}}
	room.mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Drain next event from the event loop by invoking handler via SendEvent path.
		// handleVoteSkip schedules EvNext asynchronously; wait briefly then check.
		time.Sleep(50 * time.Millisecond)
	}()

	room.handleVoteSkip("p1")
	room.mu.RLock()
	count := len(room.SkipVotes)
	status := room.Status
	room.mu.RUnlock()
	if count != 1 {
		t.Fatalf("expected 1 skip vote, got %d", count)
	}
	if status != "rating" {
		t.Fatalf("majority not reached yet, status should stay rating, got %s", status)
	}

	room.handleVoteSkip("p2")
	// Majority 2/3 → EvNext queued; process via handleNext directly after unlock path.
	// handleVoteSkip already SendEvent'd EvNext; give loop a moment.
	time.Sleep(80 * time.Millisecond)

	room.mu.RLock()
	defer room.mu.RUnlock()
	if room.Status != "rating" {
		// After next with queue item, beginRating sets rating again — OK either way if song advanced.
	}
	if room.CurrentSong == nil || room.CurrentSong.UUID != "song-2" {
		t.Fatalf("expected advance to song-2, got %#v", room.CurrentSong)
	}
	if len(room.SkipVotes) != 0 {
		t.Fatalf("skip votes should reset on new song, got %d", len(room.SkipVotes))
	}
	wg.Wait()
}

func TestVoteSkipDisabledRejected(t *testing.T) {
	cfg := testRateConfig()
	cfg.VoteSkip = false
	room := NewLobbyRoom("NOSKIP", cfg, nil, nil, nil, nil)
	room.Start()
	defer room.Close()

	conn := nopConn{}
	room.handleJoin(&JoinEvent{SessionID: "h", Conn: conn, Nickname: "Host", DeviceID: "dh", User: &domain.User{ID: 1, UUID: "uh", Name: "Host"}})
	room.mu.Lock()
	room.Status = "rating"
	room.CurrentSong = &domain.Song{ID: 1, UUID: "s", SongRomaji: strPtr("X")}
	room.mu.Unlock()

	room.handleVoteSkip("h")
	room.mu.RLock()
	defer room.mu.RUnlock()
	if len(room.SkipVotes) != 0 {
		t.Fatalf("expected no votes when disabled")
	}
}

func strPtr(s string) *string { return &s }

package rate

import (
	"testing"
	"time"

	"anirank/api/internal/domain"
)

// Regression: cleanup used to call Close() while holding m.mu, and Close → OnDestroy
// re-acquired m.mu, deadlocking ListPublicRooms forever (browser saw a fake CORS error).
func TestCleanupRoomsDoesNotDeadlockOnDestroy(t *testing.T) {
	m := NewLobbyManager(nil, nil, nil, nil)

	cfg := domain.RateConfig{
		Name:              "test",
		QueueMode:         QueueModeHostOnly,
		QueueLimitPerUser: 3,
		RevealMode:        RevealModeBlind,
		MaxPlayers:        16,
		AutoAdvance:       AutoAdvanceNever,
		SourceMode:        SourceModeManual,
	}
	sanitizeConfig(&cfg)

	room := NewLobbyRoom("DEADLOCK", cfg, nil, nil, nil, nil)
	room.CreatedAt = time.Now().Add(-10 * time.Minute)
	room.OnDestroy = func(rid string) {
		m.mu.Lock()
		delete(m.rooms, rid)
		m.mu.Unlock()
	}
	m.mu.Lock()
	m.rooms[room.RoomID] = room
	m.mu.Unlock()
	room.Start()

	done := make(chan struct{})
	go func() {
		m.cleanupRooms()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cleanupRooms deadlocked (OnDestroy re-entered manager lock)")
	}

	m.mu.RLock()
	_, stillThere := m.rooms[room.RoomID]
	m.mu.RUnlock()
	if stillThere {
		t.Fatal("expected room removed from registry")
	}
}

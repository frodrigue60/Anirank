package rate

import (
	"testing"
	"time"

	"anirank/api/internal/domain"
)

type nopConn struct{}

func (nopConn) WriteJSON(v interface{}) error { return nil }
func (nopConn) Close() error                  { return nil }

func testRateConfig() domain.RateConfig {
	cfg := domain.RateConfig{
		Name:              "test",
		QueueMode:         QueueModeEveryone,
		QueueLimitPerUser: 3,
		RevealMode:        RevealModeBlind,
		MaxPlayers:        2,
		AutoAdvance:       AutoAdvanceNever,
		SourceMode:        SourceModeManual,
	}
	sanitizeConfig(&cfg)
	return cfg
}

func TestJoinReclaimsAuthSeatByDeviceWhenJWTMissing(t *testing.T) {
	room := NewLobbyRoom("RECLAIM", testRateConfig(), nil, nil, nil, nil)
	room.Start()
	defer room.Close()

	user := &domain.User{ID: 7, UUID: "user-7", Name: "Luis"}
	conn := nopConn{}

	room.handleJoin(&JoinEvent{
		SessionID: "s1",
		Conn:      conn,
		User:      user,
		Nickname:  "Luis",
		DeviceID:  "device-a",
	})

	room.handleLeave("s1")

	// Reconnect without JWT (token race) but same device — must reclaim auth seat.
	room.handleJoin(&JoinEvent{
		SessionID: "s2",
		Conn:      conn,
		User:      nil,
		Nickname:  "Luis",
		DeviceID:  "device-a",
	})

	room.mu.RLock()
	defer room.mu.RUnlock()
	if len(room.Players) != 1 {
		t.Fatalf("expected 1 player seat, got %d", len(room.Players))
	}
	p := room.Players["s2"]
	if p == nil {
		t.Fatal("expected player on session s2")
	}
	if p.Offline {
		t.Fatal("reclaimed seat should be online")
	}
	if p.UserUUID != "user-7" || p.UserID != 7 {
		t.Fatalf("expected preserved auth identity, got uuid=%q id=%d", p.UserUUID, p.UserID)
	}
}

func TestJoinDedupesGhostGuestAndAuthSeats(t *testing.T) {
	room := NewLobbyRoom("DEDUPE", testRateConfig(), nil, nil, nil, nil)
	room.Start()
	defer room.Close()

	conn := nopConn{}
	user := &domain.User{ID: 3, UUID: "user-3", Name: "Ada"}

	// Simulate pre-fix ghost: offline auth + online guest same device.
	now := time.Now()
	room.mu.Lock()
	room.Players["auth-old"] = &domain.RatePlayer{
		SessionID:    "auth-old",
		UserUUID:     "user-3",
		UserID:       3,
		Nickname:     "Ada",
		DeviceID:     "device-b",
		Offline:      true,
		OfflineSince: &now,
	}
	room.Players["guest-new"] = &domain.RatePlayer{
		SessionID: "guest-new",
		Nickname:  "Ada",
		DeviceID:  "device-b",
	}
	room.mu.Unlock()

	room.handleJoin(&JoinEvent{
		SessionID: "s3",
		Conn:      conn,
		User:      user,
		Nickname:  "Ada",
		DeviceID:  "device-b",
	})

	room.mu.RLock()
	defer room.mu.RUnlock()
	if len(room.Players) != 1 {
		t.Fatalf("expected deduped to 1 seat, got %d", len(room.Players))
	}
	if room.Players["s3"] == nil || room.Players["s3"].UserUUID != "user-3" {
		t.Fatalf("expected single auth seat on s3, players=%v", room.Players)
	}
}

func TestOfflinePlayersDoNotCountTowardMaxPlayers(t *testing.T) {
	room := NewLobbyRoom("CAP", testRateConfig(), nil, nil, nil, nil)
	room.Start()
	defer room.Close()

	conn := nopConn{}
	room.handleJoin(&JoinEvent{SessionID: "a", Conn: conn, Nickname: "A", DeviceID: "d1", User: &domain.User{ID: 1, UUID: "u1", Name: "A"}})
	room.handleJoin(&JoinEvent{SessionID: "b", Conn: conn, Nickname: "B", DeviceID: "d2", User: &domain.User{ID: 2, UUID: "u2", Name: "B"}})
	room.handleLeave("a")
	room.handleLeave("b")

	// Room max is 2; both seats offline — a new player must still be able to join.
	room.handleJoin(&JoinEvent{SessionID: "c", Conn: conn, Nickname: "C", DeviceID: "d3", User: &domain.User{ID: 3, UUID: "u3", Name: "C"}})

	room.mu.RLock()
	defer room.mu.RUnlock()
	if room.Players["c"] == nil {
		t.Fatal("expected new player to join while others are offline")
	}
}

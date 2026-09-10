package mmo

import (
	"path/filepath"
	"testing"
	"time"
)

func testPlayer(id, name string) *Player {
	return &Player{ID: id, Name: name, send: make(chan []byte, 64)}
}

func TestMultiplayerRoomFlow(t *testing.T) {
	t.Setenv("ULAR_MATCH_STORE", filepath.Join(t.TempDir(), "m.json"))
	prev := ularCountdown
	ularCountdown = time.Millisecond
	t.Cleanup(func() { ularCountdown = prev })
	h := &Hub{Lobby: NewUlarLobby()}
	a, b, c, d, e := testPlayer("ua", "Andi"), testPlayer("ub", "Budi"), testPlayer("uc", "Citra"), testPlayer("ud", "Deni"), testPlayer("ue", "Eka")
	h.Lobby.Connect(a)
	room, errc := h.Lobby.CreateSized(a, 4, "")
	if errc != "" || room == nil || len(room.RoomCode) != 6 {
		t.Fatalf("create %s %+v", errc, room)
	}
	for _, p := range []*Player{b, c, d} {
		h.Lobby.Connect(p)
		if _, errc := h.Lobby.Join(p, room.RoomCode); errc != "" {
			t.Fatal(errc)
		}
	}
	if len(room.Players) != 4 {
		t.Fatalf("n=%d", len(room.Players))
	}
	h.Lobby.Connect(e)
	if _, errc := h.Lobby.Join(e, room.RoomCode); errc != ErrRoomFull {
		t.Fatalf("want full got %s", errc)
	}
	for _, id := range []string{"ua", "ub", "uc", "ud"} {
		if _, errc := h.Lobby.SetReady(id, true); errc != "" {
			t.Fatal(errc)
		}
	}
	if errc := h.startMatch("ub"); errc != ErrNotHost {
		t.Fatalf("host %s", errc)
	}
	if errc := h.startMatch("ua"); errc != "" {
		t.Fatal(errc)
	}
	time.Sleep(20 * time.Millisecond)
	h.Lobby.mu.Lock()
	st := room.Status
	cur := room.Match.CurrentPlayerID
	h.Lobby.mu.Unlock()
	if st != UlarPlaying || cur != "ua" {
		t.Fatalf("play %s %s", st, cur)
	}
	if errc := h.rollMatch(b); errc != ErrNotYourTurn {
		t.Fatalf("turn %s", errc)
	}
	if errc := h.rollMatch(a); errc != "" {
		t.Fatal(errc)
	}
	if errc := h.rollMatch(a); errc != ErrAlreadyRolled && errc != ErrLocked {
		t.Fatalf("double %s", errc)
	}
	h.Lobby.Disconnect(c)
	h.Lobby.mu.Lock()
	offline := h.Lobby.player(room, "uc").ConnState
	h.Lobby.mu.Unlock()
	if offline != "DISCONNECTED" {
		t.Fatal(offline)
	}
	h.Lobby.Connect(c)
	deadline := time.Now().Add(2 * time.Second)
	var pos int
	for time.Now().Before(deadline) {
		h.Lobby.mu.Lock()
		pos = h.Lobby.player(room, "ua").Position
		h.Lobby.mu.Unlock()
		if pos >= MIN_POSITION {
			break
		}
		time.Sleep(40 * time.Millisecond)
	}
	if pos < MIN_POSITION {
		t.Fatal("pos reset")
	}
}

func TestMultiplayerIgnoresClientUserID(t *testing.T) {
	t.Setenv("ULAR_MATCH_STORE", filepath.Join(t.TempDir(), "m.json"))
	h := &Hub{Lobby: NewUlarLobby()}
	a := testPlayer("real-id", "Andi")
	h.Lobby.Connect(a)
	room, _ := h.Lobby.Create(a)
	if room.Players[0].UserID != "real-id" {
		t.Fatal(room.Players[0].UserID)
	}
}

func TestMatchStorePersistsWinner(t *testing.T) {
	dir := t.TempDir()
	s := OpenMatchStore(filepath.Join(dir, "m.json"))
	s.Append(StoredMatch{ID: "mt1", RoomCode: "ABC123", Status: "FINISHED", WinnerID: "ua", Players: []StoredMatchPlayer{{UserID: "ua", FinalPosition: 100, FinishOrder: 1}}})
	s2 := OpenMatchStore(filepath.Join(dir, "m.json"))
	if len(s2.All()) != 1 || s2.All()[0].WinnerID != "ua" {
		t.Fatal("persist")
	}
}

func TestLobbyDisconnectFreesSeatImmediately(t *testing.T) {
	h := &Hub{Lobby: NewUlarLobby()}
	a, b, c := testPlayer("ua", "Andi"), testPlayer("ub", "Budi"), testPlayer("uc", "Citra")
	h.Lobby.Connect(a)
	room, errc := h.Lobby.CreateSized(a, 2, "")
	if errc != "" {
		t.Fatal(errc)
	}
	h.Lobby.Connect(b)
	if _, errc := h.Lobby.Join(b, room.RoomCode); errc != "" {
		t.Fatal(errc)
	}
	if len(room.Players) != 2 {
		t.Fatalf("want 2 got %d", len(room.Players))
	}
	h.Lobby.Disconnect(b)
	// Kursi di-hold sebentar untuk reconnect; Join pemain lain membuang ghost.
	h.Lobby.Connect(c)
	if _, errc := h.Lobby.Join(c, room.RoomCode); errc != "" {
		t.Fatalf("join after disconnect: %s", errc)
	}
	if len(room.Players) != 2 {
		t.Fatalf("n=%d", len(room.Players))
	}
	if h.Lobby.player(room, "ub") != nil {
		t.Fatal("ghost seat should be pruned on join")
	}
}

func TestStaleDisconnectDoesNotWipeReconnect(t *testing.T) {
	h := &Hub{Lobby: NewUlarLobby()}
	oldP := testPlayer("ua", "Andi")
	newP := testPlayer("ua", "Andi")
	h.Lobby.Connect(oldP)
	room, errc := h.Lobby.CreateSized(oldP, 2, "")
	if errc != "" {
		t.Fatal(errc)
	}
	h.Lobby.Connect(newP)
	if h.Lobby.online["ua"] != newP {
		t.Fatal("online should be new session")
	}
	h.Lobby.Disconnect(oldP)
	if h.Lobby.online["ua"] != newP {
		t.Fatal("stale disconnect wiped new session")
	}
	pl := h.Lobby.player(room, "ua")
	if pl == nil || !pl.IsConnected {
		t.Fatal("seat should stay connected after stale disconnect")
	}
}

func TestPlayingLeaveFreesSeat(t *testing.T) {
	prev := ularCountdown
	ularCountdown = time.Millisecond
	t.Cleanup(func() { ularCountdown = prev })
	h := &Hub{Lobby: NewUlarLobby()}
	a, b, c := testPlayer("ua", "Andi"), testPlayer("ub", "Budi"), testPlayer("uc", "Citra")
	h.Lobby.Connect(a)
	room, _ := h.Lobby.CreateSized(a, 3, "")
	h.Lobby.Connect(b)
	_, _ = h.Lobby.Join(b, room.RoomCode)
	_, _ = h.Lobby.SetReady("ua", true)
	_, _ = h.Lobby.SetReady("ub", true)
	if errc := h.startMatch("ua"); errc != "" {
		t.Fatal(errc)
	}
	time.Sleep(20 * time.Millisecond)
	h.Lobby.Leave("ub")
	h.Lobby.mu.Lock()
	n := len(room.Players)
	status := room.Status
	h.Lobby.mu.Unlock()
	if n != 1 {
		t.Fatalf("playing leave should free seat, n=%d status=%s", n, status)
	}
	h.Lobby.Connect(c)
	if _, errc := h.Lobby.Join(c, room.RoomCode); errc != "" {
		t.Fatalf("late join after leave: %s", errc)
	}
}

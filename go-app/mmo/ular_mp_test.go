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
	room, errc := h.Lobby.Create(a)
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

package mmo

import "testing"

func TestPhase1BoardConstants(t *testing.T) {
	if BOARD_SIZE != 100 || OFFBOARD_START != 0 || MIN_POSITION != 1 || MAX_POSITION != 100 || MAX_PLAYERS != 4 {
		t.Fatal("board constants")
	}
	if DefaultSnakes[97] != 78 || DefaultLadders[4] != 25 {
		t.Fatal("snake/ladder config")
	}
	if AdventureGameplayEnabled {
		t.Fatal("adventure must stay off in phase 1")
	}
}

func TestPhase1RoomFoundation(t *testing.T) {
	lobby := NewUlarLobby()
	host := &Player{ID: "u1", Name: "Host", send: make(chan []byte, 4)}
	guest := &Player{ID: "u2", Name: "Guest", send: make(chan []byte, 4)}
	room, errc := lobby.Create(host)
	if errc != "" || room == nil || room.MaxPlayers != 4 || room.Status != UlarWaiting {
		t.Fatal("create")
	}
	if room.Players[0].Position != OFFBOARD_START {
		t.Fatal("default position")
	}
	joined, msg := lobby.Join(guest, room.RoomCode)
	if msg != "" || len(joined.Players) != 2 {
		t.Fatalf("join %s", msg)
	}
	lobby.SetReady("u1", true)
	ready, errc := lobby.SetReady("u2", true)
	if errc != "" || ready.Status != UlarReady {
		t.Fatal("ready")
	}
}

func TestPhase1QuestionModel(t *testing.T) {
	q := UlarQuestion{ID: "q1", Category: "umum", Active: true, CorrectAnswer: "A"}
	if q.ID == "" || !q.Active {
		t.Fatal("question")
	}
}

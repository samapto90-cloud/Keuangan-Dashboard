package mmo

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var DefaultHub = NewHub()

func init() {
	go DefaultHub.Run()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			return true
		}
		return strings.Contains(origin, "localhost") ||
			strings.Contains(origin, "127.0.0.1") ||
			strings.Contains(origin, "sakubijak.com")
	},
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("mmo upgrade: %v", err)
		return
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(8 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		return
	}
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil || env.Type != TypeAuth {
		_ = conn.WriteMessage(websocket.TextMessage, marshal(TypeAuthFail, ErrorOut{Message: "auth diperlukan"}))
		return
	}
	var auth AuthIn
	_ = json.Unmarshal(env.Data, &auth)
	gameSess, fail := DefaultHub.Accounts.AuthenticateWS(auth)
	if fail != "" || gameSess == nil {
		_ = conn.WriteMessage(websocket.TextMessage, marshal(TypeAuthFail, ErrorOut{Message: fail}))
		return
	}
	player := &Player{
		ID:        gameSess.PlayerID,
		SessionID: gameSess.Token,
		Name:      gameSess.Username,
		Class:     "WARRIOR",
		Level:     1,
		HP:        BaseMaxHP,
		MaxHP:     BaseMaxHP,
		State:     "IDLE",
		send:      make(chan []byte, 64),
	}
	player.initCombat()

	go writePump(conn, player)
	select {
	case player.send <- marshal(TypeAuthOk, AuthOkOut{Token: gameSess.Token, PlayerID: gameSess.PlayerID, SessionID: gameSess.Token}):
	default:
	}

	joined := false
	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout))
		return nil
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		_ = conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout))
		var msg Envelope
		if json.Unmarshal(raw, &msg) != nil {
			continue
		}
		switch msg.Type {
		case TypeAuth:
			continue
		case TypeJoinWorld:
			if !joined {
				joined = true
				DefaultHub.Join(player)
			}
		default:
			DefaultHub.Inbound(player, msg)
		}
	}
	if joined {
		DefaultHub.Leave(player)
	} else {
		player.CloseSend()
	}
}

func writePump(conn *websocket.Conn, p *Player) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-p.send:
			_ = conn.SetWriteDeadline(time.Now().Add(8 * time.Second))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(8 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

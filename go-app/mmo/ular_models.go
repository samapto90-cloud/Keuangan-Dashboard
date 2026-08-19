package mmo

import "time"

type UlarPlayer struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	Color       string `json:"color"`
	Position    int    `json:"position"`
	IsReady     bool   `json:"isReady"`
	IsConnected bool   `json:"isConnected"`
}

func NewUlarPlayer(userID, username string) *UlarPlayer {
	return &UlarPlayer{
		ID:          userID,
		UserID:      userID,
		Username:    username,
		Avatar:      "default",
		Color:       "#1aa6a0",
		Position:    MIN_POSITION,
		IsReady:     false,
		IsConnected: true,
	}
}

type UlarRoom struct {
	ID         string          `json:"id"`
	RoomCode   string          `json:"roomCode"`
	HostID     string          `json:"hostId"`
	Status     UlarMatchStatus `json:"status"`
	MaxPlayers int             `json:"maxPlayers"`
	CreatedAt  time.Time       `json:"createdAt"`
	Players    []*UlarPlayer   `json:"players"`
}

type UlarMatch struct {
	ID              string          `json:"id"`
	RoomID          string          `json:"roomId"`
	Status          UlarMatchStatus `json:"status"`
	CurrentPlayerID string          `json:"currentPlayerId"`
	WinnerID        string          `json:"winnerId,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	FinishedAt      *time.Time      `json:"finishedAt,omitempty"`
}

type UlarQuestion struct {
	ID            string `json:"id"`
	Category      string `json:"category"`
	Subject       string `json:"subject"`
	Grade         string `json:"grade"`
	Difficulty    string `json:"difficulty"`
	Question      string `json:"question"`
	OptionA       string `json:"optionA"`
	OptionB       string `json:"optionB"`
	OptionC       string `json:"optionC"`
	OptionD       string `json:"optionD"`
	CorrectAnswer string `json:"correctAnswer"`
	Explanation   string `json:"explanation"`
	Active        bool   `json:"active"`
}

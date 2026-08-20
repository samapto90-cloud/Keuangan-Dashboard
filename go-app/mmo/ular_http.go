package mmo

import (
	"encoding/json"
	"net/http"
)

type ularResolveIn struct {
	Position int    `json:"position"`
	Player   string `json:"player"`
}

func HandleUlarBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	cfg := DefaultBoardConfig()
	if err := ValidateBoardConfig(cfg); err != nil {
		writeAccountErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func HandleUlarResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAccountErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in ularResolveIn
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}
	if in.Position < OFFBOARD_START || in.Position > MAX_POSITION {
		writeAccountErr(w, http.StatusBadRequest, "posisi tidak valid")
		return
	}
	dice, err := RollDiceSecure()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Terjadi kesalahan. Silakan coba lagi."})
		return
	}
	cfg := DefaultBoardConfig()
	out := ResolveMove(cfg, in.Position, dice)
	if in.Player != "" {
		out.Log = in.Player + " " + out.Log
	}
	writeJSON(w, http.StatusOK, out)
}

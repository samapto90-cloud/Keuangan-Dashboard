package mmo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleUlarBoardAndResolve(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cahaya/api/ular/board", nil)
	rec := httptest.NewRecorder()
	HandleUlarBoard(rec, req)
	if rec.Code != 200 {
		t.Fatalf("board %d %s", rec.Code, rec.Body.String())
	}
	var cfg BoardConfig
	if json.Unmarshal(rec.Body.Bytes(), &cfg) != nil || cfg.Snakes[88] != 48 || cfg.Ladders[28] != 55 {
		t.Fatalf("cfg %+v", cfg)
	}
	body, _ := json.Marshal(map[string]any{"position": 1, "player": "Budi"})
	req = httptest.NewRequest(http.MethodPost, "/cahaya/api/ular/resolve", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	HandleUlarResolve(rec, req)
	if rec.Code != 200 {
		t.Fatalf("resolve %d %s", rec.Code, rec.Body.String())
	}
	var out MoveResult
	if json.Unmarshal(rec.Body.Bytes(), &out) != nil || out.Dice < 1 || out.Dice > 6 || out.Final < 1 {
		t.Fatalf("out %+v", out)
	}
}

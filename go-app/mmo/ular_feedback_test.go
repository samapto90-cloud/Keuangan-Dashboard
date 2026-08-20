package mmo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHandleFeedback(t *testing.T) {
	dir := t.TempDir()
	DefaultHub = &Hub{Ops: OpenOpsStore(filepath.Join(dir, "ops.json"))}
	body, _ := json.Marshal(map[string]string{"category": "BUG", "message": "Tombol dadu tidak merespons"})
	req := httptest.NewRequest(http.MethodPost, "/cahaya/api/feedback", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	HandleFeedback(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}

func TestHandleFeedbackRateLimit(t *testing.T) {
	dir := t.TempDir()
	DefaultHub = &Hub{Ops: OpenOpsStore(filepath.Join(dir, "ops2.json"))}
	for i := 0; i < 6; i++ {
		body, _ := json.Marshal(map[string]string{"category": "OTHER", "message": "test message long enough"})
		req := httptest.NewRequest(http.MethodPost, "/cahaya/api/feedback", bytes.NewReader(body))
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		HandleFeedback(rec, req)
		if i < 5 && rec.Code != http.StatusOK {
			t.Fatalf("ok %d: %d", i, rec.Code)
		}
		if i == 5 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 got %d", rec.Code)
		}
	}
}

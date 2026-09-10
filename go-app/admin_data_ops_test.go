package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTahunAnggaranDefaultAndSet(t *testing.T) {
	prev := getSystemSettingsCopy()
	t.Cleanup(func() {
		systemSettingsMu.Lock()
		systemSettings = prev
		systemSettingsMu.Unlock()
	})

	systemSettingsMu.Lock()
	systemSettings.TahunAnggaran = 0
	systemSettingsMu.Unlock()
	if got := getTahunAnggaran(); got != 2026 {
		t.Fatalf("default tahun = %d, want 2026", got)
	}

	if err := setTahunAnggaran(2027, false); err != nil {
		t.Fatal(err)
	}
	if got := getTahunAnggaran(); got != 2027 {
		t.Fatalf("tahun = %d, want 2027", got)
	}
	if err := setTahunAnggaran(1999, false); err == nil {
		t.Fatal("expected invalid tahun error")
	}
}

func TestHandlePublicTahun(t *testing.T) {
	_ = setTahunAnggaran(2028, false)
	req := httptest.NewRequest(http.MethodGet, "/data/tahun", nil)
	rr := httptest.NewRecorder()
	handlePublicTahun(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if int(out["tahun_anggaran"].(float64)) != 2028 {
		t.Fatalf("got %#v", out)
	}
}

func TestDataOpsClearRequiresConfirm(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"confirm":            "salah",
		"clear_transactions": true,
	})
	req := httptest.NewRequest(http.MethodPost, "/data/admin/data-ops/clear", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleAdminDataClear(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d want 400", rr.Code)
	}
}

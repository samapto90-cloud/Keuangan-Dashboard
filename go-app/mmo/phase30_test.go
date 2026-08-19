package mmo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPhase30LandminesKept(t *testing.T) {
	if GameTitle != "Ular Tangga Nusantara" || GamePhase != "ULAR/1" {
		t.Fatal("phase 1 identity")
	}
	if AdventureGameplayEnabled {
		t.Fatal("adventure still enabled")
	}
}

func TestPhase30ReleaseViewAndEndgame(t *testing.T) {
	w, p := testVillagePlayer()
	rel := phase30ReleaseView()
	if rel["version"] != GameVersion || rel["siluman"] != 0 || rel["status"] != "PHASE_1_FOUNDATION" {
		t.Fatal("release view")
	}
	view := w.endgameView(p)
	if view["version"] != GameVersion || view["phase"] != GamePhase {
		t.Fatal("endgame version")
	}
}

func TestPhase30SecurityRejects(t *testing.T) {
	w, p := testVillagePlayer()
	if !rejectAction(w.ApplyPhase30(p, Envelope{Type: TypeGiveGold}), TypeGiveGold) {
		t.Fatal("give gold")
	}
	if !rejectAction(w.ApplyPhase30(p, Envelope{Type: TypeGiveItem}), TypeGiveItem) {
		t.Fatal("give item")
	}
	if !rejectAction(w.ApplyPhase30(p, Envelope{Type: TypeSetDamage}), TypeSetDamage) {
		t.Fatal("set damage")
	}
	if !rejectAction(w.ApplyExplore(p.ID, Envelope{Type: TypeTeleport, Data: []byte(`{"x":99,"z":99}`)}), TypeTeleport) {
		t.Fatal("teleport")
	}
}

func TestPhase30PersistBackupAndReady(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CAHAYA_SOCIAL_STORE", filepath.Join(dir, "cahaya-social.json"))
	ready := PersistReady()
	if ready["status"] != "ready" || ready["writable"] != true {
		t.Fatal("ready")
	}
	w, p := testVillagePlayer()
	w.SaveRuntime()
	path, err := BackupRuntime()
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		raw := []byte(`{"audit":[]}`)
		if err := os.WriteFile(filepath.Join(dir, "cahaya-social.json"), raw, 0o644); err != nil {
			t.Fatal(err)
		}
		path, err = BackupRuntime()
		if err != nil || path == "" {
			t.Fatalf("backup %s %v", path, err)
		}
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	_ = p
}

func TestPhase30EducationStillRewards(t *testing.T) {
	if questionByID["q-add-2-4"].Correct != 1 || questionByID["q-read-buku"].Correct != 0 {
		t.Fatal("phase29 education")
	}
	if questionByID["q-apel-3-2"].Correct != 1 {
		t.Fatal("keep apel")
	}
}

func TestPhase30RestoreRuntimeRoundtrip(t *testing.T) {
	dir := t.TempDir()
	store := filepath.Join(dir, "cahaya-social.json")
	t.Setenv("CAHAYA_SOCIAL_STORE", store)
	original := []byte(`{"audit":[{"note":"phase30-original"}]}`)
	if err := os.WriteFile(store, original, 0o644); err != nil {
		t.Fatal(err)
	}
	bak, err := BackupRuntime()
	if err != nil || bak == "" {
		t.Fatalf("backup %s %v", bak, err)
	}
	if err := os.WriteFile(store, []byte(`{"audit":[{"note":"mutated"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	safety, err := RestoreRuntime(bak)
	if err != nil {
		t.Fatal(err)
	}
	if safety == "" {
		t.Fatal("expected safety snapshot of mutated live file")
	}
	got, err := os.ReadFile(store)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("restore mismatch %s", got)
	}
	if _, err := os.Stat(bak); err != nil {
		t.Fatal("backup must remain")
	}
	if _, err := os.Stat(safety); err != nil {
		t.Fatal("safety snapshot missing")
	}
}

func TestPhase30TwoPlayerPartySmoke(t *testing.T) {
	w, a := testVillagePlayer()
	b := testOnline("p_b", "Sinta")
	w.Add(b)
	if rejectAction(w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_b"}`)}), TypePartyInvite) {
		t.Fatal("invite")
	}
	if rejectAction(w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)}), TypePartyAccept) {
		t.Fatal("accept")
	}
	if a.PartyID == "" || a.PartyID != b.PartyID {
		t.Fatalf("party %s %s", a.PartyID, b.PartyID)
	}
	if rejectAction(w.ApplyParty(a.ID, Envelope{Type: TypePartyReady, Data: []byte(`{}`)}), TypePartyReady) {
		t.Fatal("leader ready")
	}
	if rejectAction(w.ApplyParty(b.ID, Envelope{Type: TypePartyLeave, Data: []byte(`{}`)}), TypePartyLeave) {
		t.Fatal("leave")
	}
	if b.PartyID != "" {
		t.Fatal("leave must clear party")
	}
}

func TestPhase30SaveLoadProgressSmoke(t *testing.T) {
	dir := t.TempDir()
	store := filepath.Join(dir, "cahaya-social.json")
	t.Setenv("CAHAYA_SOCIAL_STORE", store)
	w, p := testVillagePlayer()
	w.Audit = append(w.Audit, AuditEntry{ID: "qa-30", Kind: "smoke", Player: p.ID, Detail: "save-load"})
	w.SaveRuntime()
	raw, err := os.ReadFile(store)
	if err != nil || !strings.Contains(string(raw), "save-load") {
		t.Fatalf("persist %s %v", raw, err)
	}
	bak, err := BackupRuntime()
	if err != nil || bak == "" {
		t.Fatalf("backup %v", err)
	}
	w2 := NewWorldState()
	w2.loadRuntimeFile()
	found := false
	for _, a := range w2.Audit {
		if a.ID == "qa-30" && a.Detail == "save-load" {
			found = true
		}
	}
	if !found {
		t.Fatal("audit did not reload")
	}
}


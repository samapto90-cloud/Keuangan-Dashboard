package mmo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAccountPasswordNeverPlaintext(t *testing.T) {
	dir := t.TempDir()
	s := OpenAccountStore(filepath.Join(dir, "accounts.json"))
	sess, msg := s.Register("RakaSatu", "raka1@example.com", "Rahasia1", "Rahasia1")
	if msg != "" || sess == nil {
		t.Fatalf("register: %s", msg)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "accounts.json"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if strings.Contains(body, "Rahasia1") {
		t.Fatal("password tersimpan plain text")
	}
	if !strings.Contains(body, "$2") {
		t.Fatal("hash bcrypt tidak ada")
	}
	if _, msg := s.Login("RakaSatu", "salah1234"); msg == "" {
		t.Fatal("password salah harus ditolak")
	}
	if _, msg := s.Login("raka1@example.com", "Rahasia1"); msg != "" {
		t.Fatalf("login email: %s", msg)
	}
}

func TestAccountUsernameAndEmailUnique(t *testing.T) {
	s := OpenAccountStore(filepath.Join(t.TempDir(), "a.json"))
	if _, msg := s.Register("PemainA", "a@example.com", "Rahasia1", "Rahasia1"); msg != "" {
		t.Fatal(msg)
	}
	if _, msg := s.Register("PemainA", "b@example.com", "Rahasia1", "Rahasia1"); msg == "" {
		t.Fatal("username duplikat")
	}
	if _, msg := s.Register("PemainB", "a@example.com", "Rahasia1", "Rahasia1"); msg == "" {
		t.Fatal("email duplikat")
	}
}

func TestPlayerIsolationChaptersAndSave(t *testing.T) {
	dir := t.TempDir()
	store := OpenDiskPlayerStore(dir)
	w := NewWorldState()
	w.QuestRepo = store.AsQuest()
	w.InvRepo = store.AsInv()
	w.JourneyRepo = store

	a := &Player{ID: "p_iso_a", Name: "PemainA", send: make(chan []byte, 8)}
	b := &Player{ID: "p_iso_b", Name: "PemainB", send: make(chan []byte, 8)}
	a.initCombat()
	b.initCombat()
	w.Add(a)
	w.Add(b)
	a.ensureLog().StoryChapter = "st-ch03"
	a.ensureLog().StoryCheckpoint = "camp-river"
	a.X, a.Z = 1, 82
	a.Level = 8
	a.Bag.add("potion_heal", 2)
	b.ensureLog().StoryChapter = "st-ch07"
	b.ensureLog().StoryCheckpoint = "camp-siluman"
	b.X, b.Z = -1, 150
	b.Level = 20
	w.persist(a)
	w.persistGear(a)
	w.persist(b)
	w.persistGear(b)

	w2 := NewWorldState()
	w2.QuestRepo = store.AsQuest()
	w2.InvRepo = store.AsInv()
	w2.JourneyRepo = store
	a2 := &Player{ID: "p_iso_a", Name: "PemainA", send: make(chan []byte, 8)}
	b2 := &Player{ID: "p_iso_b", Name: "PemainB", send: make(chan []byte, 8)}
	a2.initCombat()
	b2.initCombat()
	w2.Add(a2)
	w2.Add(b2)
	if a2.ensureLog().StoryChapter != "st-ch03" {
		t.Fatalf("A chapter %s", a2.ensureLog().StoryChapter)
	}
	if b2.ensureLog().StoryChapter != "st-ch07" {
		t.Fatalf("B chapter %s", b2.ensureLog().StoryChapter)
	}
	if a2.Z < 70 {
		t.Fatalf("A harus continue sungai, z=%v", a2.Z)
	}
	if b2.Z < 140 {
		t.Fatalf("B harus continue desa siluman, z=%v", b2.Z)
	}
	if a2.Bag.count("potion_heal") < 2 {
		t.Fatal("inventory A hilang setelah restart")
	}
	if b2.Bag.count("potion_heal") >= 2 && a2.ID != b2.ID {
		if b2.ensureLog().StoryChapter == a2.ensureLog().StoryChapter {
			t.Fatal("data tertukar")
		}
	}
}

func TestTenAccountsStayIsolated(t *testing.T) {
	acc := OpenAccountStore(filepath.Join(t.TempDir(), "acc.json"))
	store := OpenDiskPlayerStore(t.TempDir())
	want := map[string]string{}
	for i := 0; i < 10; i++ {
		name := "Pemain" + itoa(i)
		sess, msg := acc.Register(name, name+"@example.com", "Rahasia1", "Rahasia1")
		if msg != "" {
			t.Fatal(msg)
		}
		ch := "st-ch10"
		if i < 9 {
			ch = "st-ch0" + itoa(i+1)
		}
		log := newPlayerLog()
		log.StoryChapter = ch
		store.AsQuest().Save(sess.PlayerID, log)
		want[sess.PlayerID] = ch
	}
	seen := map[string]bool{}
	for id, ch := range want {
		if seen[id] {
			t.Fatal("player id duplikat")
		}
		seen[id] = true
		got := store.LoadLog(id)
		if got == nil || got.StoryChapter != ch {
			t.Fatalf("%s chapter %v want %s", id, got, ch)
		}
		acc.AccountByID(id)
	}
}

func TestSessionCannotReadOtherPlayer(t *testing.T) {
	s := OpenAccountStore(filepath.Join(t.TempDir(), "s.json"))
	a, _ := s.Register("IsoA", "isoa@example.com", "Rahasia1", "Rahasia1")
	b, _ := s.Register("IsoB", "isob@example.com", "Rahasia1", "Rahasia1")
	sa := s.Lookup(a.Token)
	if sa.PlayerID != a.PlayerID || sa.PlayerID == b.PlayerID {
		t.Fatal("sesi A tidak boleh mengikat B")
	}
	if s.Lookup(b.Token).PlayerID != b.PlayerID {
		t.Fatal("sesi B")
	}
}

func TestCheckpointContinueAfterLogout(t *testing.T) {
	store := OpenDiskPlayerStore(t.TempDir())
	w := NewWorldState()
	w.QuestRepo = store.AsQuest()
	w.InvRepo = store.AsInv()
	w.JourneyRepo = store
	p := &Player{ID: "p_cp", Name: "Raka", send: make(chan []byte, 8)}
	p.initCombat()
	w.Add(p)
	p.X, p.Z = 0, 34
	evs := w.useCheckpoint(p, InteractDef{ID: "camp-forest", X: 0, Z: 34})
	if len(evs) == 0 {
		t.Fatal("checkpoint save")
	}
	w.Remove(p)
	w2 := NewWorldState()
	w2.QuestRepo = store.AsQuest()
	w2.InvRepo = store.AsInv()
	w2.JourneyRepo = store
	p2 := &Player{ID: "p_cp", Name: "Raka", send: make(chan []byte, 8)}
	p2.initCombat()
	w2.Add(p2)
	if p2.ensureLog().StoryCheckpoint != "camp-forest" {
		t.Fatalf("checkpoint %s", p2.ensureLog().StoryCheckpoint)
	}
	if p2.Z < 30 {
		t.Fatalf("continue z=%v", p2.Z)
	}
}

func TestStoryHasTenJourneyChapters(t *testing.T) {
	want := []string{
		"Awal Perjalanan", "Hutan Larangan", "Sungai Kehidupan", "Gunung Kabut", "Kuil Tua",
		"Gurun Panjang", "Desa Para Siluman", "Dungeon Kegelapan", "Wilayah Terakhir", "Perjalanan Menuju Masjid",
	}
	if len(storyChapterCatalog) != 10 {
		t.Fatalf("chapters %d", len(storyChapterCatalog))
	}
	for i, title := range want {
		if storyChapterCatalog[i].Title != title {
			t.Fatalf("ch %d %s", i+1, storyChapterCatalog[i].Title)
		}
	}
}

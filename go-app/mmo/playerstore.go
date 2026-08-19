package mmo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var playerFileRE = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type JourneySave struct {
	HasPos         bool    `json:"hasPos"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
	Z              float64 `json:"z"`
	Yaw            float64 `json:"yaw"`
	CheckpointID   string  `json:"checkpointId"`
	CheckpointName string  `json:"checkpointName"`
	Chapter        string  `json:"chapter"`
	SavedAt        int64   `json:"savedAt"`
}

type playerBlob struct {
	Log     *PlayerLog   `json:"log"`
	Gear    *GearSave    `json:"gear"`
	Journey *JourneySave `json:"journey"`
}

type JourneyRepository interface {
	LoadJourney(playerID string) *JourneySave
	SaveJourney(playerID string, j *JourneySave)
}

type DiskPlayerStore struct {
	mu   sync.Mutex
	dir  string
	logs map[string]*PlayerLog
	gear map[string]*GearSave
	pos  map[string]*JourneySave
}

func playerStoreDir() string {
	if p := strings.TrimSpace(os.Getenv("CAHAYA_PLAYER_STORE")); p != "" {
		return p
	}
	return filepath.Join("data", "cahaya-players")
}

func OpenDiskPlayerStore(dir string) *DiskPlayerStore {
	_ = os.MkdirAll(dir, 0o755)
	return &DiskPlayerStore{
		dir:  dir,
		logs: map[string]*PlayerLog{},
		gear: map[string]*GearSave{},
		pos:  map[string]*JourneySave{},
	}
}

func (s *DiskPlayerStore) AsQuest() QuestRepository { return diskQuest{s} }
func (s *DiskPlayerStore) AsInv() InventoryRepository { return diskInv{s} }

func safePlayerFile(id string) string {
	id = strings.TrimSpace(id)
	if !playerFileRE.MatchString(id) {
		return ""
	}
	return id + ".json"
}

func (s *DiskPlayerStore) filePath(id string) string {
	name := safePlayerFile(id)
	if name == "" {
		return ""
	}
	return filepath.Join(s.dir, name)
}

func (s *DiskPlayerStore) hydrateLocked(id string) {
	if _, ok := s.logs[id]; ok {
		return
	}
	path := s.filePath(id)
	if path == "" {
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) == 0 {
		return
	}
	var blob playerBlob
	if json.Unmarshal(raw, &blob) != nil {
		return
	}
	if blob.Log != nil {
		s.logs[id] = cloneLog(blob.Log)
	}
	if blob.Gear != nil {
		s.gear[id] = blob.Gear
	}
	if blob.Journey != nil {
		cp := *blob.Journey
		s.pos[id] = &cp
	}
}

func (s *DiskPlayerStore) flushLocked(id string) {
	path := s.filePath(id)
	if path == "" {
		return
	}
	blob := playerBlob{Log: s.logs[id], Gear: s.gear[id], Journey: s.pos[id]}
	_ = writeAtomicJSON(path, blob)
}

func (s *DiskPlayerStore) LoadLog(id string) *PlayerLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hydrateLocked(id)
	if s.logs[id] == nil {
		return nil
	}
	return cloneLog(s.logs[id])
}

func (s *DiskPlayerStore) LoadGear(id string) *GearSave {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hydrateLocked(id)
	return cloneGearSave(s.gear[id])
}

func (s *DiskPlayerStore) LoadJourney(playerID string) *JourneySave {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hydrateLocked(playerID)
	j := s.pos[playerID]
	if j == nil {
		return nil
	}
	cp := *j
	return &cp
}

func (s *DiskPlayerStore) SaveJourney(playerID string, j *JourneySave) {
	if j == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hydrateLocked(playerID)
	cp := *j
	s.pos[playerID] = &cp
	s.flushLocked(playerID)
}

func (s *DiskPlayerStore) saveLog(id string, log *PlayerLog) {
	if log == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hydrateLocked(id)
	s.logs[id] = cloneLog(log)
	s.flushLocked(id)
}

func (s *DiskPlayerStore) saveGear(id string, g *GearSave) {
	if g == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hydrateLocked(id)
	s.gear[id] = cloneGearSave(g)
	s.flushLocked(id)
}

type diskQuest struct{ s *DiskPlayerStore }

func (d diskQuest) Load(playerID string) *PlayerLog { return d.s.LoadLog(playerID) }
func (d diskQuest) Save(playerID string, log *PlayerLog) { d.s.saveLog(playerID, log) }

type diskInv struct{ s *DiskPlayerStore }

func (d diskInv) Load(playerID string) *GearSave { return d.s.LoadGear(playerID) }
func (d diskInv) Save(playerID string, g *GearSave) { d.s.saveGear(playerID, g) }

func cloneGearSave(s *GearSave) *GearSave {
	if s == nil {
		return nil
	}
	cp := *s
	cp.Inv = cloneInv(s.Inv)
	cp.Bank = cloneInv(s.Bank)
	cp.UnlockedSkills = append([]string{}, s.UnlockedSkills...)
	cp.UnlockedForms = append([]string{}, s.UnlockedForms...)
	cp.LoadoutSkills = append([]string{}, s.LoadoutSkills...)
	cp.GearUp = copyIntMapVal(s.GearUp)
	cp.GearInst = copyStrMapVal(s.GearInst)
	return &cp
}

func journeyFromPlayer(p *Player) *JourneySave {
	if p == nil {
		return nil
	}
	log := p.ensureLog()
	ch := log.StoryChapter
	if ch == "" {
		ch = "st-ch01"
	}
	name := log.StoryCheckpoint
	if name == "" {
		name = "Desa Awal"
	}
	return &JourneySave{
		HasPos: true, X: p.X, Y: p.Y, Z: p.Z, Yaw: p.Yaw,
		CheckpointID: log.StoryCheckpoint, CheckpointName: checkpointDisplayName(name),
		Chapter: ch, SavedAt: time.Now().UnixMilli(),
	}
}

func checkpointDisplayName(id string) string {
	if id == "" {
		return "Desa Awal"
	}
	if lm, ok := landmarkByID[id]; ok && lm.Name != "" {
		return lm.Name
	}
	switch id {
	case "checkpoint-1":
		return "Alun-Alun Desa"
	case "camp-forest":
		return "Camp Hutan Larangan"
	case "camp-river":
		return "Camp Sungai Kehidupan"
	case "camp-mountain":
		return "Shrine Gunung Kabut"
	case "camp-temple":
		return "Kuil Tua"
	case "camp-desert":
		return "Camp Gurun Panjang"
	case "camp-siluman":
		return "Desa Para Siluman"
	case "camp-dungeon":
		return "Ambang Dungeon"
	case "camp-dark":
		return "Wilayah Terakhir"
	case "camp-masjid":
		return "Halaman Masjid"
	default:
		return id
	}
}

func (w *WorldState) persistJourney(p *Player) {
	if w.JourneyRepo == nil || p == nil {
		return
	}
	w.JourneyRepo.SaveJourney(p.ID, journeyFromPlayer(p))
}

func (w *WorldState) restoreJourney(p *Player) {
	if w.JourneyRepo == nil || p == nil {
		return
	}
	j := w.JourneyRepo.LoadJourney(p.ID)
	if j == nil || !j.HasPos {
		return
	}
	if j.X != j.X || j.Z != j.Z || j.Y != j.Y {
		return
	}
	p.X, p.Y, p.Z, p.Yaw = j.X, j.Y, j.Z, j.Yaw
}

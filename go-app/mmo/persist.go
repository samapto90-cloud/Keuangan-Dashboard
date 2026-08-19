package mmo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type runtimeBlob struct {
	Guilds         *GuildHub           `json:"guilds"`
	Social         *SocialHub          `json:"social"`
	Audit          []AuditEntry        `json:"audit"`
	Reports        []Report            `json:"reports"`
	TxDone         map[string][][]byte `json:"txDone"`
	Pvp            *pvpPersist         `json:"pvp"`
	Community      *CommunityLive      `json:"community"`
	HorizonLB      []HorizonScore      `json:"horizonLb"`
	HorizonLBWeek  string              `json:"horizonLbWeek"`
	HorizonHistory []HorizonScore      `json:"horizonHistory"`
	Market         *MarketHub          `json:"market"`
	Housing        *HousingHub         `json:"housing"`
	EconomyLog     []EconomyTx         `json:"economyLog"`
	ChatLog        []ChatOut           `json:"chatLog"`
	DungeonBoard   []DungeonBoardRow   `json:"dungeonBoard"`
	DungeonHistory []DungeonRunRow     `json:"dungeonHistory"`
}

type pvpPersist struct {
	History []PvpHistoryRow           `json:"history"`
	Abuse   map[string]*PvpAbuse      `json:"abuse"`
	LB      LeaderboardCache          `json:"leaderboard"`
	Reports []PvpReport               `json:"reports"`
	Replays map[string][]ReplayEvent  `json:"replays"`
}

func runtimePath() string {
	if p := os.Getenv("CAHAYA_SOCIAL_STORE"); p != "" {
		return p
	}
	return filepath.Join("data", "cahaya-social.json")
}

func (w *WorldState) exportRuntime() runtimeBlob {
	blob := runtimeBlob{Guilds: w.Guilds, Social: w.Social, Audit: w.Audit, Reports: w.Reports, TxDone: w.TxDone,
		Community: w.Community, HorizonLB: w.HorizonLB, HorizonLBWeek: w.HorizonLBWeek, HorizonHistory: w.HorizonHistory,
		Market: w.Market, Housing: w.Housing, EconomyLog: w.EconomyLog, ChatLog: w.ChatLog,
		DungeonBoard: w.DungeonBoard, DungeonHistory: w.DungeonHistory}
	if w.PvP != nil {
		blob.Pvp = &pvpPersist{History: w.PvP.history, Abuse: w.PvP.abuse, LB: w.PvP.lb, Reports: w.PvP.reports, Replays: w.PvP.replays}
	}
	return blob
}

func (w *WorldState) importRuntime(b runtimeBlob) {
	if b.Guilds != nil {
		w.Guilds = b.Guilds
		if w.Guilds.ByID == nil {
			w.Guilds.ByID = map[string]*Guild{}
		}
		if w.Guilds.ByName == nil {
			w.Guilds.ByName = map[string]string{}
		}
		if w.Guilds.ByTag == nil {
			w.Guilds.ByTag = map[string]string{}
		}
		if w.Guilds.ByPlayer == nil {
			w.Guilds.ByPlayer = map[string]string{}
		}
		if w.Guilds.Invites == nil {
			w.Guilds.Invites = map[string]string{}
		}
		for id, g := range w.Guilds.ByID {
			if g == nil {
				continue
			}
			if g.Members == nil {
				g.Members = map[string]*GuildMember{}
			}
			for pid := range g.Members {
				w.Guilds.ByPlayer[pid] = id
			}
		}
	}
	if b.Social != nil {
		w.Social = b.Social
		if w.Social.Friends == nil {
			w.Social.Friends = map[string]map[string]FriendRel{}
		}
		if w.Social.Incoming == nil {
			w.Social.Incoming = map[string]map[string]bool{}
		}
		if w.Social.Outgoing == nil {
			w.Social.Outgoing = map[string]map[string]bool{}
		}
		if w.Social.Blocked == nil {
			w.Social.Blocked = map[string]map[string]bool{}
		}
	}
	if b.Audit != nil {
		w.Audit = b.Audit
	}
	if b.Reports != nil {
		w.Reports = b.Reports
	}
	if b.TxDone != nil {
		w.TxDone = b.TxDone
	}
	if b.Pvp != nil {
		if w.PvP == nil {
			w.PvP = NewPvpHub()
		}
		w.PvP.history = b.Pvp.History
		if b.Pvp.Abuse != nil {
			w.PvP.abuse = b.Pvp.Abuse
		}
		w.PvP.lb = b.Pvp.LB
		if b.Pvp.Reports != nil {
			w.PvP.reports = b.Pvp.Reports
		}
		if b.Pvp.Replays != nil {
			w.PvP.replays = b.Pvp.Replays
		}
		w.PvP.ensure()
	}
	if b.Community != nil {
		w.Community = b.Community
		if w.Community.Granted == nil {
			w.Community.Granted = map[string]bool{}
		}
		if w.Community.Participants == nil {
			w.Community.Participants = map[string]int{}
		}
	}
	w.HorizonLB = b.HorizonLB
	w.HorizonLBWeek = b.HorizonLBWeek
	w.HorizonHistory = b.HorizonHistory
	if b.Market != nil {
		w.Market = b.Market
		if w.Market.History == nil {
			w.Market.History = map[string][]MarketHistory{}
		}
	}
	if b.Housing != nil {
		w.Housing = b.Housing
		if w.Housing.ByOwner == nil {
			w.Housing.ByOwner = map[string]*HouseInstance{}
		}
		if w.Housing.ByID == nil {
			w.Housing.ByID = map[string]*HouseInstance{}
		}
		for _, h := range w.Housing.ByOwner {
			if h != nil {
				w.Housing.ByID[h.ID] = h
			}
		}
	}
	if b.EconomyLog != nil {
		w.EconomyLog = b.EconomyLog
	}
	if b.ChatLog != nil {
		w.ChatLog = b.ChatLog
	}
	if b.DungeonBoard != nil {
		w.DungeonBoard = b.DungeonBoard
	}
	if b.DungeonHistory != nil {
		w.DungeonHistory = b.DungeonHistory
	}
}

func writeAtomicJSON(path string, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (w *WorldState) loadRuntimeFile() {
	raw, err := os.ReadFile(runtimePath())
	if err != nil || len(raw) == 0 {
		return
	}
	var blob runtimeBlob
	if json.Unmarshal(raw, &blob) != nil {
		return
	}
	w.importRuntime(blob)
}

func (w *WorldState) saveRuntimeLocked() {
	blob := w.exportRuntime()
	raw, err := json.Marshal(blob)
	if err != nil {
		return
	}
	path := runtimePath()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := path + ".tmp"
	if os.WriteFile(tmp, raw, 0o644) != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

func (w *WorldState) SaveRuntime() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.saveRuntimeLocked()
}

// BackupRuntime copies the JSON runtime store next to the live file.
// It never overwrites the live store and never deletes data.
func BackupRuntime() (string, error) {
	src := runtimePath()
	raw, err := os.ReadFile(src)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	dst := src + ".bak-" + time.Now().UTC().Format("20060102-150405")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dst, raw, 0o644); err != nil {
		return "", err
	}
	return dst, nil
}

// RestoreRuntime copies a backup JSON onto the live store.
// The current live file is snapshotted first (BackupRuntime) so the restore can be rolled back.
// It never deletes backup files.
func RestoreRuntime(backupPath string) (safety string, err error) {
	backupPath = strings.TrimSpace(backupPath)
	if backupPath == "" {
		return "", fmt.Errorf("backup path required")
	}
	raw, err := os.ReadFile(backupPath)
	if err != nil {
		return "", err
	}
	if !json.Valid(raw) {
		return "", fmt.Errorf("backup is not valid JSON")
	}
	dest := runtimePath()
	if _, err := os.Stat(dest); err == nil {
		safety, err = BackupRuntime()
		if err != nil {
			return "", err
		}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return safety, err
	}
	tmp := dest + ".restore-tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return safety, err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return safety, err
	}
	return safety, nil
}

func PersistReady() map[string]any {
	path := runtimePath()
	writable := true
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		writable = false
	}
	_, statErr := os.Stat(path)
	return map[string]any{
		"status":   "ready",
		"version":  GameVersion,
		"persist":  path,
		"exists":   statErr == nil,
		"writable": writable,
	}
}

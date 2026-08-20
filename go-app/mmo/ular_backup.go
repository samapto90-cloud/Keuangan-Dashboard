package mmo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"
	"strings"
)

var (
	ularBackupMu           sync.Mutex
	ularLastBackupAt      time.Time
	ularLastBackupErr     string
	ularLastRestoreAt     time.Time
	ularLastRestoreErr    string
	ularLastRestoreSrc    string
	ularRestoreCountTotal int
)

func ularBackupRoot() string {
	// Default under `data/` to match other runtime stores.
	if p := filepath.ToSlash(strings.TrimSpace(os.Getenv("ULAR_BACKUP_ROOT"))); p != "" {
		return p
	}
	return filepath.Join("data", "ular-backups")
}

// ularSnapshotName uses the same parse-friendly layout across platforms.
func ularSnapshotName(t time.Time) string {
	// Example: 20260820-133055 (UTC).
	return t.UTC().Format("20060102-150405")
}

func parseUlarSnapshotName(name string) (time.Time, error) {
	// Must be UTC to make pruning consistent.
	return time.ParseInLocation("20060102-150405", name, time.UTC)
}

func ularBackupFiles() map[string]string {
	// map[absolutePath]logicalName
	m := map[string]string{}

	paths := map[string]string{
		matchStorePath():    "matches",
		attemptStorePath():  "attempts",
		progressStorePath(): "progress",
		socialStorePath():   "social",
		opsStorePath():      "ops",
		accountStorePath():  "accounts",
	}

	// Optional: if a question external store is configured.
	if qp := questionStorePath(); qp != "" {
		paths[qp] = "questions"
	}

	for p, name := range paths {
		if strings.TrimSpace(p) == "" {
			continue
		}
		m[filepath.FromSlash(p)] = name
	}
	return m
}

func copyFile(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	// Atomic-ish: write temp then rename.
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func ularLatestSnapshotDir(root string) (string, time.Time, error) {
	ents, err := os.ReadDir(root)
	if err != nil {
		return "", time.Time{}, err
	}
	type item struct {
		name string
		t    time.Time
	}
	items := make([]item, 0, len(ents))
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		t, err := parseUlarSnapshotName(e.Name())
		if err != nil {
			continue
		}
		items = append(items, item{name: e.Name(), t: t})
	}
	if len(items) == 0 {
		return "", time.Time{}, os.ErrNotExist
	}
	sort.Slice(items, func(i, j int) bool { return items[i].t.After(items[j].t) })
	return filepath.Join(root, items[0].name), items[0].t, nil
}

func UlarBackupAll() (string, error) {
	ularBackupMu.Lock()
	defer ularBackupMu.Unlock()

	root := ularBackupRoot()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}

	snapDir := filepath.Join(root, ularSnapshotName(time.Now()))

	files := ularBackupFiles()
	copied := []string{}
	ularLastBackupErr = ""
	for livePath, logical := range files {
		if _, err := os.Stat(livePath); err != nil {
			continue
		}
		dst := filepath.Join(snapDir, logical+".json")
		if err := copyFile(livePath, dst); err != nil {
			ularLastBackupErr = err.Error()
			return "", fmt.Errorf("backup %s (%s): %w", livePath, logical, err)
		}
		copied = append(copied, logical)
	}
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return "", err
	}
	_ = copied

	// Prune old snapshots for retention control.
	pruneUlarBackupsLocked(root, time.Now().UTC())
	ularLastBackupAt = time.Now()
	return filepath.Base(snapDir), nil
}

func pruneUlarBackupsLocked(root string, now time.Time) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}

	type snap struct {
		name string
		t    time.Time
	}
	snaps := []snap{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		t, err := parseUlarSnapshotName(e.Name())
		if err != nil {
			continue
		}
		snaps = append(snaps, snap{name: e.Name(), t: t})
	}
	if len(snaps) == 0 {
		return
	}

	// Keep set.
	keep := map[string]bool{}
	// Daily: last 7 days.
	latestDaily := now.Add(-7 * 24 * time.Hour)
	for _, s := range snaps {
		if s.t.After(latestDaily) || s.t.Equal(latestDaily) {
			keep[s.name] = true
		}
	}

	// Weekly: keep latest snapshot per ISO week for last 28 days.
	type wk struct {
		year int
		week int
	}
	wMap := map[wk]snap{}
	for _, s := range snaps {
		if s.t.Before(now.Add(-28 * 24 * time.Hour)) {
			continue
		}
		y, w := s.t.ISOWeek()
		k := wk{year: y, week: w}
		if cur, ok := wMap[k]; !ok || s.t.After(cur.t) {
			wMap[k] = s
		}
	}
	for _, s := range wMap {
		keep[s.name] = true
	}

	// Monthly: keep latest snapshot per month for last 90 days.
	type mo struct {
		year  int
		month int
	}
	mMap := map[mo]snap{}
	for _, s := range snaps {
		if s.t.Before(now.Add(-90 * 24 * time.Hour)) {
			continue
		}
		k := mo{year: s.t.Year(), month: int(s.t.Month())}
		if cur, ok := mMap[k]; !ok || s.t.After(cur.t) {
			mMap[k] = s
		}
	}
	for _, s := range mMap {
		keep[s.name] = true
	}

	// Delete anything not in keep.
	for _, s := range snaps {
		if keep[s.name] {
			continue
		}
		_ = os.RemoveAll(filepath.Join(root, s.name))
	}
}

// UlarRestoreSnapshot restores all live stores from a given snapshot name.
// Snapshot name must match `ularSnapshotName()` output, without root prefix.
func UlarRestoreSnapshot(snapshotName string) (string, []string, error) {
	ularBackupMu.Lock()
	defer ularBackupMu.Unlock()

	root := ularBackupRoot()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", nil, err
	}

	var snapDir string
	var err error
	if snapshotName == "" || snapshotName == "LATEST" {
		snapDir, _, err = ularLatestSnapshotDir(root)
		if err != nil {
			return "", nil, err
		}
	} else {
		snapDir = filepath.Join(root, snapshotName)
		if _, err := os.Stat(snapDir); err != nil {
			return "", nil, err
		}
	}

	// Safety snapshot before restore.
	safetyDir := filepath.Join(root, "safety-"+ularSnapshotName(time.Now()))
	if err := os.MkdirAll(safetyDir, 0o755); err != nil {
		return "", nil, err
	}
	files := ularBackupFiles()
	safetyCopied := []string{}
	for livePath, logical := range files {
		if _, err := os.Stat(livePath); err != nil {
			continue
		}
		safetyDst := filepath.Join(safetyDir, logical+".json")
		if err := copyFile(livePath, safetyDst); err != nil {
			// If safety fails, abort restore.
			return "", nil, err
		}
		safetyCopied = append(safetyCopied, logical)
	}
	_ = safetyCopied

	restored := []string{}
	for livePath, logical := range files {
		src := filepath.Join(snapDir, logical+".json")
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := copyFile(src, livePath); err != nil {
			return safetyDir, restored, err
		}
		restored = append(restored, logical)
	}
	ularLastRestoreAt = time.Now()
	return safetyDir, restored, nil
}

func reloadUlarStoresFromDisk() {
	// Best-effort reload of persistent stores.
	// Used by admin restore-test. This is not intended for restore during active gameplay.
	if DefaultHub == nil {
		return
	}
	DefaultHub.Accounts = OpenAccountStore(accountStorePath())
	DefaultHub.Progress = OpenProgressStore(progressStorePath())
	DefaultHub.Social = OpenSocialStore(socialStorePath())
	DefaultHub.Ops = OpenOpsStore(opsStorePath())
	if DefaultHub.Lobby != nil {
		DefaultHub.Lobby.Store = OpenMatchStore(matchStorePath())
		DefaultHub.Lobby.Attempts = OpenAttemptStore(attemptStorePath())
	}
}

func startUlarDailyBackupLoop() {
	enabled := os.Getenv("ULAR_BACKUP_ENABLED")
	if strings.TrimSpace(enabled) != "1" {
		return
	}

	intervalHours := 24
	if raw := strings.TrimSpace(os.Getenv("ULAR_BACKUP_INTERVAL_HOURS")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			intervalHours = n
		}
	}
	tick := time.NewTicker(time.Duration(intervalHours) * time.Hour)

	// First run immediately on boot.
	go func() {
		for {
			_, _ = UlarBackupAll()
			// Record errors/status in a deferred wrapper below.
			// We intentionally keep loop simple and resilient.
			<-tick.C
		}
	}()
}

func UlarBackupStatus() map[string]any {
	return map[string]any{
		"lastBackupAt":    ularLastBackupAt.UnixMilli(),
		"lastBackupErr":   ularLastBackupErr,
		"lastRestoreAt":   ularLastRestoreAt.UnixMilli(),
		"lastRestoreErr":  ularLastRestoreErr,
		"lastRestoreFrom": ularLastRestoreSrc,
		"restoreCount":    ularRestoreCountTotal,
	}
}

// adminUlarRestoreTest is protected via RBAC through HandleAdminAPI.
func adminUlarRestoreTest(w http.ResponseWriter, r *http.Request, a *adminCtx) {
	var in struct {
		Snapshot string `json:"snapshot"` // optional: "LATEST" or snapshotName
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeAccountErr(w, http.StatusBadRequest, "payload")
		return
	}

	safety, restored, err := UlarRestoreSnapshot(strings.TrimSpace(in.Snapshot))
	ularLastRestoreAt = time.Now()
	ularLastRestoreSrc = in.Snapshot
	ularRestoreCountTotal++
	if err != nil {
		ularLastRestoreErr = err.Error()
		writeAccountErr(w, http.StatusBadRequest, "restore gagal")
		return
	}
	ularLastRestoreErr = ""

	reloadUlarStoresFromDisk()

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"restored":   restored,
		"safetyBackup": safety,
	})
}

// Note: needs to be called from server start (or via init with env guard).
func init() {
	// Keep unit tests stable: only start loop when explicitly enabled.
	startUlarDailyBackupLoop()
}


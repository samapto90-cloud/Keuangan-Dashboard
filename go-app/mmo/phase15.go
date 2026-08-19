package mmo

import (
	"strings"
	"time"
)

// DungeonHub is DungeonService + RaidService. Phase 15 extends it; do not add a duplicate service.

const (
	QueueLeaveWindow   = 12 * time.Second
	QueuePenaltyAfter  = 3
	QueuePenaltyDur    = 90 * time.Second
	RaidEnrageDefault  = 8 * time.Minute
	DungeonInactiveTTL = 8 * time.Minute
	PhaseLockDuration  = 2 * time.Second
	ReviveRaidCooldown = 8.0
)

type DungeonRunRow struct {
	PlayerID   string  `json:"playerId"`
	Name       string  `json:"name"`
	DungeonID  string  `json:"dungeonId"`
	NameDun    string  `json:"nameDun,omitempty"`
	Kind       string  `json:"kind"`
	Difficulty string  `json:"difficulty"`
	Rating     string  `json:"rating"`
	Elapsed    int     `json:"elapsed"`
	Deaths     int     `json:"deaths"`
	Wipes      int     `json:"wipes"`
	Guild      string  `json:"guild,omitempty"`
	Party      string  `json:"party,omitempty"`
	Bosses     int     `json:"bosses"`
	At         int64   `json:"at"`
	FirstClear bool    `json:"firstClear,omitempty"`
}

type DungeonBoardRow struct {
	DungeonID  string `json:"dungeonId"`
	Name       string `json:"name"`
	PlayerID   string `json:"playerId"`
	Player     string `json:"player"`
	Guild      string `json:"guild,omitempty"`
	Party      string `json:"party,omitempty"`
	Elapsed    int    `json:"elapsed"`
	Difficulty string `json:"difficulty"`
	Rating     string `json:"rating"`
	At         int64  `json:"at"`
}

type RaidShopItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Cost     int    `json:"cost"`
	Kind     string `json:"kind"`
	RewardID string `json:"rewardId"`
}

type PendingLootRow struct {
	ClaimID string         `json:"claimId"`
	Items   []LootItemView `json:"items"`
}

var raidShopCatalog = []RaidShopItem{
	{ID: "raid-cloak", Name: "Mistwood Cloak", Cost: 20, Kind: "cosmetic", RewardID: "cloak-mistwood"},
	{ID: "raid-banner", Name: "Celestial Banner", Cost: 35, Kind: "cosmetic", RewardID: "banner-celestial"},
	{ID: "raid-title", Name: "Celestial Raider", Cost: 50, Kind: "title", RewardID: "celestial-raider"},
	{ID: "raid-frame", Name: "Light Avatar Frame", Cost: 40, Kind: "cosmetic", RewardID: "frame-light"},
}

func (h *DungeonHub) ensurePhase15() {
	if h.PenaltyUntil == nil {
		h.PenaltyUntil = map[string]time.Time{}
	}
	if h.QueueStrikes == nil {
		h.QueueStrikes = map[string]int{}
	}
	if h.LastQueueJoin == nil {
		h.LastQueueJoin = map[string]time.Time{}
	}
}

func normalizeDungeonDiff(diff string) string {
	d := strings.ToUpper(strings.TrimSpace(diff))
	switch d {
	case "HARD":
		return "HARD"
	case "HEROIC", "MYTHIC":
		return d
	default:
		return "NORMAL"
	}
}

func prototypeDiffOK(diff string) bool {
	return diff == "NORMAL" || diff == "HARD"
}

func (w *WorldState) pickRandomDungeon(p *Player, diff string) (DungeonDef, bool) {
	var picks []DungeonDef
	for _, def := range dungeonCatalog {
		if dungeonKind(def) != "DUNGEON" {
			continue
		}
		if p.Level < def.MinimumLevel {
			continue
		}
		if def.ChapterID != "" && chapterStatus(p, chapterByID[def.ChapterID]) == "LOCKED" {
			continue
		}
		if def.ID == horizonCfg.DungeonID && !p.endgameUnlocked() {
			continue
		}
		if w.raidLocked(p, def) {
			continue
		}
		picks = append(picks, def)
	}
	if len(picks) == 0 {
		return DungeonDef{}, false
	}
	seed := time.Now().UnixNano() + int64(len(p.ID))*17 + int64(p.Level)
	if diff == "HARD" {
		seed += 41
	}
	return picks[int(seed%int64(len(picks)))], true
}

func dungeonScale(inst *DungeonInstance) (hp, dmg, telegraph float64) {
	hp, dmg, telegraph = 1, 1, 1
	if inst == nil {
		return
	}
	switch strings.ToUpper(inst.Difficulty) {
	case "HARD":
		hp, dmg, telegraph = 1.35, 1.22, 0.82
	case "HEROIC":
		hp, dmg, telegraph = 1.7, 1.4, 0.72
	case "MYTHIC":
		hp, dmg, telegraph = 2.1, 1.65, 0.62
	}
	n := len(inst.Players)
	if n < 1 {
		n = 1
	}
	extra := float64(n-1) * 0.11
	hp *= 1 + extra
	dmg *= 1 + extra*0.45
	telegraph *= 1 - extra*0.035
	if telegraph < 0.55 {
		telegraph = 0.55
	}
	return hp, dmg, telegraph
}

func applyInstanceScale(inst *DungeonInstance, e *Enemy) {
	if inst == nil || e == nil {
		return
	}
	hpM, dmgM, _ := dungeonScale(inst)
	e.MaxHP = int(float64(e.MaxHP) * hpM)
	if e.MaxHP < 1 {
		e.MaxHP = 1
	}
	e.HP = e.MaxHP
	e.Def.Attack = int(float64(e.Def.Attack) * dmgM)
	if e.Def.Attack < 1 {
		e.Def.Attack = 1
	}
}

func hardTimeLimit(def DungeonDef, diff string) int {
	sec := def.TimeLimit
	if sec < 1 {
		sec = 1800
	}
	if strings.ToUpper(diff) == "HARD" {
		return sec * 25 / 30
	}
	return sec
}

func raidResetAt(now time.Time) time.Time {
	now = now.UTC()
	wd := int(now.Weekday())
	if wd == 0 {
		wd = 7
	}
	days := 8 - wd
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, days)
	if !start.After(now) {
		start = start.Add(7 * 24 * time.Hour)
	}
	return start
}

func raidLockoutLabel(now time.Time) string {
	left := raidResetAt(now).Sub(now)
	if left < 0 {
		left = 0
	}
	d := int(left.Hours()) / 24
	h := int(left.Hours()) % 24
	return itoa(d) + "d " + itoa(h) + "h"
}

func filterAFKLoot(items []LootItemView) []LootItemView {
	var out []LootItemView
	for _, it := range items {
		r := strings.ToUpper(it.Rarity)
		if r == "EPIC" || r == "LEGENDARY" {
			continue
		}
		def := itemByID[it.ItemID]
		if def.Type == "EQUIPMENT" || def.Type == "WEAPON" || def.Type == "ARMOR" {
			continue
		}
		out = append(out, it)
	}
	return out
}

func (w *WorldState) applyPhase15Rewards(p *Player, inst *DungeonInstance, elapsed float64, deaths int) {
	if p == nil || inst == nil {
		return
	}
	def := dungeonByID[inst.DefID]
	log := p.ensureLog()
	if log.FirstClears == nil {
		log.FirstClears = map[string]bool{}
	}
	first := !log.FirstClears[def.ID]
	if first {
		log.FirstClears[def.ID] = true
		w.giveExp(p, def.Rewards.Exp/5+40)
		w.giveCurrency(p, def.Rewards.Coin/10+15, 0)
		if def.ID == "dun-mistwood" {
			p.grantCosmetic("cloak-mistwood")
		}
	}
	if log.WeeklyChallengeProg >= 3 && log.WeeklyDungeonBonus != raidWeekKey() {
		log.WeeklyDungeonBonus = raidWeekKey()
		w.giveCurrency(p, 80, 1)
		w.addCurrency(p, "raid", 5, "weekly_dungeon")
	}
	kind := dungeonKind(def)
	if kind == "RAID" {
		w.addCurrency(p, "raid", 25, "raid_clear")
		p.grantTitle("celestial-raider")
		if inst.WipeCount == 0 {
			log.Flags["ach_dungeon_no_death"] = true
		}
	}
	if def.ID == "dun-mistwood" {
		p.grantTitle("cavern-explorer")
	}
	if def.ID == "dun-crimson" || def.ID == "dun-ember" {
		p.grantTitle("temple-breaker")
	}
	if inst.Rating == "S+" || inst.Rating == "S" {
		p.grantTitle("speed-runner")
		p.grantCosmetic("badge-speed-run")
		log.Flags["ach_dungeon_fast"] = true
	}
	if inst.WipeCount == 0 && deaths == 0 {
		log.Flags["ach_dungeon_no_death"] = true
	}
	if inst.PuzzleStep > 0 && inst.PuzzleStep >= len(inst.PuzzleNeed) {
		log.Flags["ach_dungeon_puzzle"] = true
	}
	log.Flags["ach_dungeon_boss"] = true
	row := DungeonRunRow{
		PlayerID: p.ID, Name: p.Name, DungeonID: def.ID, NameDun: def.Name, Kind: kind,
		Difficulty: inst.Difficulty, Rating: inst.Rating, Elapsed: int(elapsed), Deaths: deaths,
		Wipes: inst.WipeCount, Guild: p.GuildTag, Party: inst.PartyID, Bosses: len(def.Bosses),
		At: time.Now().UnixMilli(), FirstClear: first,
	}
	log.DungeonRuns = append(log.DungeonRuns, row)
	if len(log.DungeonRuns) > 40 {
		log.DungeonRuns = log.DungeonRuns[len(log.DungeonRuns)-40:]
	}
	w.recordDungeonBoard(row)
	w.applyPhase23Rewards(p, inst, elapsed, deaths)
	_ = w.refreshAchievements(p)
}

func (w *WorldState) recordDungeonBoard(row DungeonRunRow) {
	if w.DungeonBoard == nil {
		w.DungeonBoard = []DungeonBoardRow{}
	}
	entry := DungeonBoardRow{
		DungeonID: row.DungeonID, Name: row.NameDun, PlayerID: row.PlayerID, Player: row.Name,
		Guild: row.Guild, Party: row.Party, Elapsed: row.Elapsed, Difficulty: row.Difficulty,
		Rating: row.Rating, At: row.At,
	}
	replaced := false
	for i, cur := range w.DungeonBoard {
		if cur.DungeonID == row.DungeonID && cur.PlayerID == row.PlayerID {
			if row.Elapsed > 0 && (cur.Elapsed == 0 || row.Elapsed < cur.Elapsed) {
				w.DungeonBoard[i] = entry
			}
			replaced = true
			break
		}
	}
	if !replaced {
		w.DungeonBoard = append(w.DungeonBoard, entry)
	}
	if len(w.DungeonBoard) > 200 {
		w.DungeonBoard = w.DungeonBoard[len(w.DungeonBoard)-200:]
	}
	w.DungeonHistory = append(w.DungeonHistory, row)
	if len(w.DungeonHistory) > 300 {
		w.DungeonHistory = w.DungeonHistory[len(w.DungeonHistory)-300:]
	}
}

func (w *WorldState) dungeonHistoryFor(p *Player) []DungeonRunRow {
	log := p.ensureLog()
	out := append([]DungeonRunRow{}, log.DungeonRuns...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}

func (w *WorldState) dungeonBoardFor(dungeonID string) []DungeonBoardRow {
	var out []DungeonBoardRow
	for _, row := range w.DungeonBoard {
		if dungeonID != "" && row.DungeonID != dungeonID {
			continue
		}
		out = append(out, row)
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Elapsed > 0 && (out[i].Elapsed == 0 || out[j].Elapsed < out[i].Elapsed) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if len(out) > 15 {
		out = out[:15]
	}
	return out
}

func (w *WorldState) queuePenaltyActive(id string) bool {
	if w.Dungeons == nil {
		return false
	}
	w.Dungeons.ensurePhase15()
	until := w.Dungeons.PenaltyUntil[id]
	return time.Now().Before(until)
}

func (w *WorldState) noteQueueJoin(id string) {
	w.Dungeons.ensurePhase15()
	w.Dungeons.LastQueueJoin[id] = time.Now()
}

func (w *WorldState) noteQueueLeave(id string) {
	w.Dungeons.ensurePhase15()
	joined := w.Dungeons.LastQueueJoin[id]
	if joined.IsZero() || time.Since(joined) > QueueLeaveWindow {
		w.Dungeons.QueueStrikes[id] = 0
		return
	}
	w.Dungeons.QueueStrikes[id]++
	if w.Dungeons.QueueStrikes[id] >= QueuePenaltyAfter {
		w.Dungeons.PenaltyUntil[id] = time.Now().Add(QueuePenaltyDur)
		w.Dungeons.QueueStrikes[id] = 0
	}
}

func (w *WorldState) raidTokenExchange(p *Player, shopID string) [][]byte {
	var item RaidShopItem
	found := false
	for _, it := range raidShopCatalog {
		if it.ID == shopID {
			item, found = it, true
			break
		}
	}
	if !found {
		return rejectFor(p.ID, TypeRaidExchange, "item")
	}
	if !w.removeCurrency(p, "raid", item.Cost, "raid_shop") {
		return rejectFor(p.ID, TypeRaidExchange, "currency")
	}
	switch item.Kind {
	case "title":
		p.grantTitle(item.RewardID)
	default:
		p.grantCosmetic(item.RewardID)
	}
	p.markDirty()
	w.persist(p)
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Raid token ditukar.", nil)), marshal(TypeDungeonList, w.dungeonList(p))}
}

func (w *WorldState) tickRaidMechanics(inst *DungeonInstance, now time.Time) [][]byte {
	if inst == nil || dungeonKind(dungeonByID[inst.DefID]) != "RAID" || inst.State != DunBoss {
		return nil
	}
	if inst.GuideHP <= 0 && inst.GuideAlive {
		inst.GuideAlive = false
		return w.wipeToCheckpoint(inst)
	}
	if inst.MechanicUntil.IsZero() || now.After(inst.MechanicUntil) {
		if inst.Mechanic == "STACK" {
			inst.Mechanic = "SPREAD"
		} else {
			inst.Mechanic = "STACK"
		}
		inst.MechanicUntil = now.Add(18 * time.Second)
		return [][]byte{marshal(TypeBossTelegraph, BossTelegraphOut{
			InstanceID: inst.ID, Skill: inst.Mechanic, X: 0, Z: 18, Radius: 4, Until: inst.MechanicUntil.UnixMilli(),
			VFX: "raid-" + strings.ToLower(inst.Mechanic), Shape: "circle", Pulse: true,
		})}
	}
	var events [][]byte
	alive := []*Player{}
	for _, id := range inst.Players {
		p := w.players[id]
		if p != nil && p.Connected && p.alive() {
			alive = append(alive, p)
		}
	}
	if len(alive) < 2 {
		return events
	}
	fail := false
	if inst.Mechanic == "STACK" {
		maxD := 0.0
		for i := 0; i < len(alive); i++ {
			for j := i + 1; j < len(alive); j++ {
				d := dist2(alive[i].X, alive[i].Z, alive[j].X, alive[j].Z)
				if d > maxD {
					maxD = d
				}
			}
		}
		fail = maxD > 8
	}
	if inst.Mechanic == "SPREAD" {
		minD := 99.0
		for i := 0; i < len(alive); i++ {
			for j := i + 1; j < len(alive); j++ {
				d := dist2(alive[i].X, alive[i].Z, alive[j].X, alive[j].Z)
				if d < minD {
					minD = d
				}
			}
		}
		fail = minD < 2.2
	}
	if fail && now.UnixMilli()%900 < 50 {
		for _, p := range alive {
			events = append(events, w.hitPlayer(inst.BossID, "Raid Mechanic", p, 8, 0, 0.4, "raid")...)
		}
	}
	return events
}

func (w *WorldState) cleanupIdleInstances(now time.Time) {
	if w.Dungeons == nil {
		return
	}
	for id, inst := range w.Dungeons.instances {
		if inst == nil {
			continue
		}
		live := 0
		for _, pid := range inst.Players {
			p := w.players[pid]
			if p != nil && p.Connected {
				live++
			}
		}
		stale := inst.State == DunClosing || inst.State == DunAbandoned
		idle := live == 0 && now.Sub(inst.StartedAt) > DungeonInactiveTTL
		if stale || idle {
			for _, e := range inst.Enemies {
				if e != nil {
					e.Alive = false
				}
			}
			inst.Enemies = map[string]*Enemy{}
			inst.Objects = nil
			inst.Loot = map[string][]LootItemView{}
			delete(w.Dungeons.instances, id)
		}
	}
}

func (w *WorldState) seedRaidGuide(inst *DungeonInstance) {
	if inst == nil || dungeonKind(dungeonByID[inst.DefID]) != "RAID" {
		return
	}
	inst.GuideHP = 100
	inst.GuideAlive = true
	inst.CrystalOrder = []string{"crystal-1", "crystal-2", "crystal-3"}
	inst.Objects = append(inst.Objects, ObjectSnapshot{ID: "celestial-guide", Kind: "npc", X: 3, Z: 14, Text: "Celestial Guide"})
}

func (w *WorldState) activateRaidCrystal(inst *DungeonInstance, crystalID string) bool {
	if inst == nil || len(inst.CrystalOrder) == 0 {
		return false
	}
	need := inst.CrystalOrder[0]
	if crystalID != need {
		inst.CrystalOrder = []string{"crystal-1", "crystal-2", "crystal-3"}
		return false
	}
	inst.CrystalOrder = inst.CrystalOrder[1:]
	return true
}

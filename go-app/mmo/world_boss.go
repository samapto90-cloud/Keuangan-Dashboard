package mmo

import (
	"strconv"
	"time"

	_ "embed"
)

//go:embed data/worldBoss.json
var worldBossJSON []byte

type WorldBossDef struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Region      string    `json:"region"`
	X           float64   `json:"x"`
	Z           float64   `json:"z"`
	Level       int       `json:"level"`
	MaxHP       int       `json:"maxHp"`
	MaxPlayers  int       `json:"maxPlayers"`
	Announce    string    `json:"announce"`
	DurationSec float64   `json:"durationSec"`
	Rewards     RewardDef `json:"rewards"`
}

type WorldBossLive struct {
	Def     WorldBossDef
	State   string
	SpawnID string
	EnemyID string
	HP      int
	MaxHP   int
	Until   time.Time
	Damage  map[string]int
	Heal    map[string]int
	Support map[string]int
	Revive  map[string]int
	Claimed map[string]bool
	Seq     int
}

var worldBossDef WorldBossDef

func init() {
	mustJSON("worldBoss.json", worldBossJSON, &worldBossDef)
	if worldBossDef.ID == "" {
		worldBossDef = WorldBossDef{
			ID: "ancient-gate-guardian", Name: "Guardian of the Ancient Gate", Region: "valley",
			X: 0, Z: 58, Level: 18, MaxHP: 8000, MaxPlayers: WorldBossMaxPop,
			Announce: "Ancient Guardian has appeared in Stone Valley!", DurationSec: 180,
			Rewards: RewardDef{Exp: 120, Coin: 50, Crystal: 2},
		}
	}
	if worldBossDef.MaxPlayers < 1 {
		worldBossDef.MaxPlayers = WorldBossMaxPop
	}
}

func (b *WorldBossLive) phase() (int, string) {
	if b == nil || b.MaxHP <= 0 {
		return 1, "AOE"
	}
	pct := 100 * b.HP / b.MaxHP
	switch {
	case pct > 70:
		return 1, "AOE"
	case pct > 40:
		return 2, "SUMMON"
	case pct > 10:
		return 3, "SHIELD"
	default:
		return 4, "FINAL"
	}
}

func (b *WorldBossLive) score(id string) int {
	return b.Damage[id] + b.Heal[id] + b.Support[id]*20 + b.Revive[id]*40
}

func (b *WorldBossLive) view() WorldBossView {
	ph, name := b.phase()
	n := 0
	seen := map[string]bool{}
	for id := range b.Damage {
		seen[id] = true
	}
	for id := range b.Support {
		seen[id] = true
	}
	n = len(seen)
	return WorldBossView{
		ID: b.Def.ID, Name: b.Def.Name, Region: b.Def.Region, State: b.State, Announce: b.Def.Announce,
		HP: b.HP, MaxHP: b.MaxHP, Phase: ph, PhaseName: name, Until: b.Until.UnixMilli(), Players: n,
		X: b.Def.X, Z: b.Def.Z,
	}
}

func (w *WorldState) spawnWorldBoss() [][]byte {
	if w.Boss != nil && (w.Boss.State == "ACTIVE" || w.Boss.State == "ANNOUNCED") {
		return nil
	}
	def := worldBossDef
	def = w.rotatedWorldBoss(def)
	hp := def.MaxHP
	if hp < 500 {
		hp = 8000
	}
	eDef := EnemyDef{
		ID: def.ID, Name: def.Name, Level: def.Level, MaxHP: hp, Attack: 14, Defense: 8, Speed: 1.4,
		AttackRange: 2.4, AggroRange: 16, Leash: 22, AttackCooldown: 1.8, ExpReward: 80, Behavior: "boss", Rank: "boss",
	}
	e := spawnEnemy(eDef, def.X, def.Z)
	e.NoRespawn = true
	w.enemies[e.ID] = e
	dur := def.DurationSec
	if dur < 30 {
		dur = 180
	}
	w.Boss = &WorldBossLive{
		Def: def, State: "ACTIVE", SpawnID: strconv.FormatInt(time.Now().UnixNano(), 10),
		EnemyID: e.ID, HP: e.HP, MaxHP: e.MaxHP, Until: time.Now().Add(time.Duration(dur * float64(time.Second))),
		Damage: map[string]int{}, Heal: map[string]int{}, Support: map[string]int{}, Revive: map[string]int{}, Claimed: map[string]bool{},
	}
	v := w.Boss.view()
	return [][]byte{
		marshal(TypeWorldBossAnnounce, v),
		marshal(TypeWorldBossState, v),
		marshal(TypeChatMessage, ChatOut{Channel: "SYSTEM", From: "SYSTEM", Text: def.Announce, System: true}),
	}
}

func (w *WorldState) tickWorldBoss(now time.Time) [][]byte {
	b := w.Boss
	if b == nil || (b.State != "ACTIVE" && b.State != "ANNOUNCED") {
		return nil
	}
	if e := w.enemies[b.EnemyID]; e != nil {
		b.HP = e.HP
		if e.HP <= 0 && e.Alive {
			e.Alive = false
		}
		if !e.Alive || e.HP <= 0 {
			b.State = "SUCCESS"
			b.HP = 0
			v := b.view()
			return [][]byte{marshal(TypeWorldBossState, v)}
		}
	}
	if now.After(b.Until) {
		b.State = "EXPIRED"
		if e := w.enemies[b.EnemyID]; e != nil {
			e.Alive = false
			e.HP = 0
			delete(w.enemies, e.ID)
		}
		v := b.view()
		return [][]byte{marshal(TypeWorldBossState, v)}
	}
	return nil
}

func (w *WorldState) noteWorldBossHit(p *Player, e *Enemy, dealt int) {
	b := w.Boss
	if b == nil || b.State != "ACTIVE" || e == nil || e.ID != b.EnemyID || dealt <= 0 {
		return
	}
	if b.Damage == nil {
		b.Damage = map[string]int{}
	}
	b.Damage[p.ID] += dealt
	b.HP = e.HP
	if b.Support[p.ID] == 0 {
		if b.Support == nil {
			b.Support = map[string]int{}
		}
		b.Support[p.ID] = 1
	}
}

func (w *WorldState) claimWorldBoss(p *Player, bossID string, clientScore int, tx string) [][]byte {
	_ = clientScore
	if prev, ok := w.txSeen(tx); ok {
		return prev
	}
	b := w.Boss
	if b == nil || (bossID != "" && b.Def.ID != bossID) {
		return rejectFor(p.ID, TypeClaimWorldBoss, "boss")
	}
	if b.State != "SUCCESS" && b.State != "ACTIVE" {
		return rejectFor(p.ID, TypeClaimWorldBoss, "state")
	}
	pid := "wb:" + b.Def.ID + ":" + b.SpawnID + ":" + p.ID
	log := p.ensureLog()
	if log.WorldBossClaims == nil {
		log.WorldBossClaims = map[string]bool{}
	}
	if b.Claimed[p.ID] || log.WorldBossClaims[pid] {
		return rejectFor(p.ID, TypeClaimWorldBoss, "claimed")
	}
	if b.score(p.ID) < 1 {
		return rejectFor(p.ID, TypeClaimWorldBoss, "participation")
	}
	b.Claimed[p.ID] = true
	log.WorldBossClaims[pid] = true
	r := b.Def.Rewards
	events := w.giveExp(p, r.Exp)
	w.giveCurrency(p, r.Coin, r.Crystal)
	chest := worldBossChest(b.score(p.ID))
	w.noteActivity(p, "WORLD_BOSS", b.Def.ID, 1)
	p.markDirty()
	w.persist(p)
	w.audit("worldBossRewarded", p.ID, pid)
	events = append(events, marshal(TypeEventReward, map[string]any{
		"eventId": b.Def.ID, "participationId": pid, "exp": r.Exp, "coin": r.Coin, "toId": p.ID, "score": b.score(p.ID), "chest": chest,
	}))
	return w.rememberTx(tx, events)
}

func (w *WorldState) worldBossViewFor(p *Player) *WorldBossView {
	b := w.Boss
	if b == nil || (b.State != "ACTIVE" && b.State != "ANNOUNCED") {
		return nil
	}
	if p != nil && zoneAt(p.X, p.Z).ID != b.Def.Region && b.Def.Region != "" {
		return nil
	}
	v := b.view()
	return &v
}

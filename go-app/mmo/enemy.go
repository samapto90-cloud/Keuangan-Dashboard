package mmo

import (
	"encoding/json"
	"log"
	"math"
	"sync/atomic"
	"time"

	_ "embed"
)

//go:embed data/enemies.json
var enemiesJSON []byte

//go:embed data/skills.json
var skillsJSON []byte

type EnemyDef struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Level          int     `json:"level"`
	MaxHP          int     `json:"maxHp"`
	Attack         int     `json:"attack"`
	Defense        int     `json:"defense"`
	Speed          float64 `json:"speed"`
	AttackRange    float64 `json:"attackRange"`
	AggroRange     float64 `json:"aggroRange"`
	Leash          float64 `json:"leash"`
	AttackCooldown float64 `json:"attackCooldown"`
	ExpReward      int     `json:"expReward"`
	Behavior       string  `json:"behavior"`
	LootTableID    string  `json:"lootTableId"`
	Rank           string  `json:"rank"`
}

type SkillDef struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Kind          string  `json:"kind"`
	Type          string  `json:"type"`
	Target        string  `json:"target"`
	Damage        int     `json:"damage"`
	EnergyCost    int     `json:"energyCost"`
	StaminaCost   float64 `json:"staminaCost"`
	Cooldown      float64 `json:"cooldown"`
	Range         float64 `json:"range"`
	Radius        float64 `json:"radius"`
	RequiredLevel int     `json:"requiredLevel"`
	RequiredSkill string  `json:"requiredSkill"`
	Animation     string  `json:"animation"`
	VFX           string  `json:"vfx"`
	Effect        string  `json:"effect"`
	Element       string  `json:"element"`
}

type Enemy struct {
	Def          EnemyDef
	ID           string
	X, Y, Z, Yaw float64
	SX, SY, SZ   float64
	VX, VZ       float64
	HP, MaxHP    int
	State        string
	TargetID     string
	LastHitBy    string
	NextAttack   time.Time
	DeadUntil    time.Time
	Alive        bool
	InstanceID   string
	NoRespawn    bool
	Elite        bool
	IsBoss       bool
	Stagger      int
	StaggerMax   int
	Guarding     bool
	StatusUntil  map[string]time.Time
}

var (
	enemyCatalog []EnemyDef
	skillCatalog = map[string]SkillDef{}
	enemySeq     atomic.Uint64
)

func init() {
	if err := json.Unmarshal(enemiesJSON, &enemyCatalog); err != nil {
		log.Printf("mmo enemies.json: %v", err)
	}
	var skills []SkillDef
	if err := json.Unmarshal(skillsJSON, &skills); err != nil {
		log.Printf("mmo skills.json: %v", err)
	}
	for _, s := range skills {
		skillCatalog[s.ID] = s
	}
}

func (e *Enemy) Snap() EnemySnapshot {
	return EnemySnapshot{
		ID:    e.ID,
		Kind:  e.Def.ID,
		Name:  e.Def.Name,
		Level: e.Def.Level,
		X:     e.X,
		Y:     e.Y,
		Z:     e.Z,
		Yaw:   e.Yaw,
		HP:    e.HP,
		MaxHP: e.MaxHP,
		State: e.State,
		Rank:  e.Def.Rank,
	}
}

func (e *Enemy) Dist(x, z float64) float64 {
	return math.Hypot(e.X-x, e.Z-z)
}

func spawnEnemy(def EnemyDef, x, z float64) *Enemy {
	id := def.ID + "-" + itoa(int(enemySeq.Add(1)))
	return &Enemy{
		Def:   def,
		ID:    id,
		X:     x,
		Y:     0,
		Z:     z,
		SX:    x,
		SY:    0,
		SZ:    z,
		HP:    def.MaxHP,
		MaxHP: def.MaxHP,
		State: "IDLE",
		Alive: true,
	}
}

func (w *WorldState) seedEnemies() {
	spots := []struct {
		kind string
		x, z float64
	}{
		{"training_dummy", -2.8, 9.1},
		{"training_dummy", 0, 9.6},
		{"training_dummy", 2.8, 9.1},
		{"forest_fang", -2.4, 19.4},
		{"forest_fang", 2.4, 19.4},
		{"forest_fang", 0, 20.2},
		{"forest_fang", -6, 26},
		{"forest_fang", 6, 26},
		{"forest_fang", 0, 28.5},
		{"shadow_imp", -7, 34},
		{"shadow_imp", 7, 34},
		{"stone_beast", 0, 38},
		{"stone_beast", -3.5, 36},
	}
	byID := map[string]EnemyDef{}
	for _, d := range enemyCatalog {
		byID[d.ID] = d
	}
	for _, s := range spots {
		def, ok := byID[s.kind]
		if !ok {
			continue
		}
		e := spawnEnemy(def, s.x, s.z)
		w.enemies[e.ID] = e
	}
}

func (w *WorldState) enemyPool(p *Player) map[string]*Enemy {
	if p != nil && p.InstanceID != "" {
		if inst := w.dungeonOf(p.ID); inst != nil {
			return inst.Enemies
		}
		return map[string]*Enemy{}
	}
	return w.enemies
}

func (w *WorldState) enemyByID(id string) *Enemy {
	if e := w.enemies[id]; e != nil {
		return e
	}
	if w.Dungeons != nil {
		for _, inst := range w.Dungeons.instances {
			if e := inst.Enemies[id]; e != nil {
				return e
			}
		}
	}
	return nil
}

func (w *WorldState) nearestEnemy(x, z, maxDist float64) *Enemy {
	return w.nearestIn(w.enemies, x, z, maxDist)
}

func (w *WorldState) nearestEnemyFor(p *Player, maxDist float64) *Enemy {
	return w.nearestIn(w.enemyPool(p), p.X, p.Z, maxDist)
}

func (w *WorldState) nearestIn(pool map[string]*Enemy, x, z, maxDist float64) *Enemy {
	var best *Enemy
	bestD := maxDist
	for _, e := range pool {
		if !e.Alive || e.HP <= 0 {
			continue
		}
		d := e.Dist(x, z)
		if d <= bestD {
			bestD = d
			best = e
		}
	}
	return best
}

func (w *WorldState) enemiesInRadius(x, z, radius float64) []*Enemy {
	return w.enemiesInRadiusPool(w.enemies, x, z, radius)
}

func (w *WorldState) enemiesInRadiusFor(p *Player, radius float64) []*Enemy {
	return w.enemiesInRadiusPool(w.enemyPool(p), p.X, p.Z, radius)
}

func (w *WorldState) enemiesInRadiusPool(pool map[string]*Enemy, x, z, radius float64) []*Enemy {
	out := make([]*Enemy, 0, 4)
	for _, e := range pool {
		if !e.Alive || e.HP <= 0 {
			continue
		}
		if e.Dist(x, z) <= radius {
			out = append(out, e)
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

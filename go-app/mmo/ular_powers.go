package mmo

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	PowerBomb      = "bomb"
	PowerThunder   = "thunder"
	PowerSuperman  = "superman"
	TypeUsePower   = "power:use"
	TypePowerUsed  = "power:used"
	TypePowerGrant = "power:grant"
	TypePowerPickup = "power:pickup"
)

var powerBoardLimits = map[string]int{
	PowerBomb:     1,
	PowerThunder:  5,
	PowerSuperman: 3,
}

func GrantRandomPower() string {
	// legacy helper — unused for board spawns
	opts := []string{PowerBomb, PowerThunder, PowerSuperman}
	var b [1]byte
	if _, err := rand.Read(b[:]); err != nil {
		return PowerBomb
	}
	return opts[int(b[0])%len(opts)]
}

func clampBoardPos(p int) int {
	if p < OFFBOARD_START {
		return OFFBOARD_START
	}
	if p > MAX_POSITION {
		return MAX_POSITION
	}
	return p
}

func (p *UlarPlayer) ensureItems() {
	if p.Items == nil {
		p.Items = map[string]int{}
	}
}

func (p *UlarPlayer) addItem(kind string) {
	p.ensureItems()
	p.Items[kind]++
}

func (p *UlarPlayer) takeItem(kind string) bool {
	p.ensureItems()
	if p.Items[kind] <= 0 {
		return false
	}
	p.Items[kind]--
	return true
}

func randIntn(n int) int {
	if n <= 0 {
		return 0
	}
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int(binary.BigEndian.Uint64(b[:]) % uint64(n))
}

/** Spawn 1 bom, 5 petir, 3 pesawat di kotak acak. */
func SpawnPowerCells(snakes, ladders map[int]int) map[int]string {
	blocked := map[int]bool{1: true, 100: true}
	for a, b := range snakes {
		blocked[a] = true
		blocked[b] = true
	}
	for a, b := range ladders {
		blocked[a] = true
		blocked[b] = true
	}
	pool := make([]int, 0, 90)
	for p := 2; p <= 99; p++ {
		if !blocked[p] {
			pool = append(pool, p)
		}
	}
	for i := len(pool) - 1; i > 0; i-- {
		j := randIntn(i + 1)
		pool[i], pool[j] = pool[j], pool[i]
	}
	out := map[int]string{}
	idx := 0
	place := func(kind string, n int) {
		for k := 0; k < n && idx < len(pool); k++ {
			out[pool[idx]] = kind
			idx++
		}
	}
	place(PowerBomb, powerBoardLimits[PowerBomb])
	place(PowerThunder, powerBoardLimits[PowerThunder])
	place(PowerSuperman, powerBoardLimits[PowerSuperman])
	return out
}

func (h *Hub) tryPickupPowerLocked(room *UlarRoom, pl *UlarPlayer) (string, bool) {
	if room == nil || room.Match == nil || pl == nil {
		return "", false
	}
	m := room.Match
	if m.PowerCells == nil {
		return "", false
	}
	kind, ok := m.PowerCells[pl.Position]
	if !ok || kind == "" {
		return "", false
	}
	delete(m.PowerCells, pl.Position)
	pl.addItem(kind)
	return kind, true
}

type UsePowerIn struct {
	Item     string `json:"item"`
	TargetID string `json:"targetId"`
}

type PowerUsedOut struct {
	UlarEnvelope
	UserID      string         `json:"userId"`
	Username    string         `json:"username"`
	Item        string         `json:"item"`
	TargetID    string         `json:"targetId"`
	TargetName  string         `json:"targetName"`
	FromPos     int            `json:"fromPos"`
	ToPos       int            `json:"toPos"`
	ActorItems  map[string]int `json:"actorItems,omitempty"`
	TargetItems map[string]int `json:"targetItems,omitempty"`
	PowerCells  map[int]string `json:"powerCells,omitempty"`
	Message     string         `json:"message"`
}

type PowerPickupOut struct {
	UlarEnvelope
	UserID     string         `json:"userId"`
	Username   string         `json:"username"`
	Item       string         `json:"item"`
	Position   int            `json:"position"`
	Items      map[string]int `json:"items,omitempty"`
	PowerCells map[int]string `json:"powerCells,omitempty"`
	Message    string         `json:"message"`
}

func (h *Hub) usePower(p *Player, item, targetID string) string {
	l := h.Lobby
	l.mu.Lock()
	defer l.mu.Unlock()
	room := l.rooms[l.inRoom[p.ID]]
	if room == nil || room.Match == nil || room.Status != UlarPlaying {
		return ErrNotInRoom
	}
	m := room.Match
	if m.WaitingAnim || m.CurrentQuestionID != "" {
		return ErrLocked
	}
	if m.CurrentPlayerID != p.ID {
		return ErrNotYourTurn
	}
	actor := l.player(room, p.ID)
	if actor == nil {
		return ErrNotInRoom
	}
	item = normalizePower(item)
	if item == "" {
		return ErrInvalidRequest
	}
	if !actor.takeItem(item) {
		return "no_item"
	}
	if targetID == "" {
		targetID = p.ID
	}
	target := l.player(room, targetID)
	if target == nil {
		actor.addItem(item)
		return ErrInvalidRequest
	}
	if item != PowerSuperman && targetID == p.ID {
		actor.addItem(item)
		return ErrInvalidRequest
	}
	from := target.Position
	to := from
	msg := ""
	switch item {
	case PowerBomb:
		to = OFFBOARD_START
		msg = fmt.Sprintf("%s melempar 💣 Bom ke %s — kembali ke START!", actor.Username, target.Username)
	case PowerThunder:
		to = clampBoardPos(from - 3)
		msg = fmt.Sprintf("%s menyambar ⚡ Petir ke %s — mundur 3 langkah!", actor.Username, target.Username)
	case PowerSuperman:
		to = clampBoardPos(from + 3)
		if targetID == p.ID {
			msg = fmt.Sprintf("%s memakai ✈️ Pesawat — naik 3 kotak!", actor.Username)
		} else {
			msg = fmt.Sprintf("%s memberi ✈️ Pesawat ke %s — naik 3 kotak!", actor.Username, target.Username)
		}
	default:
		actor.addItem(item)
		return ErrInvalidRequest
	}
	target.Position = to
	out := PowerUsedOut{
		UserID: p.ID, Username: actor.Username, Item: item,
		TargetID: target.UserID, TargetName: target.Username,
		FromPos: from, ToPos: to, Message: msg,
		ActorItems:  cloneItemMap(actor.Items),
		TargetItems: cloneItemMap(target.Items),
		PowerCells:  clonePowerCells(m.PowerCells),
	}
	seq, evid, at := l.nextSeq(room)
	out.UlarEnvelope = UlarEnvelope{Seq: seq, EventID: evid, At: at}
	h.emitLocked(room, TypePowerUsed, out)
	snap := l.wrap(room, l.snapshotLocked(room, ""))
	h.emitLocked(room, TypeGameState, snap)
	return ""
}

func normalizePower(s string) string {
	switch s {
	case PowerBomb, "BOMB", "bom":
		return PowerBomb
	case PowerThunder, "THUNDER", "petir", "lightning":
		return PowerThunder
	case PowerSuperman, "SUPERMAN", "supermen", "pesawat", "PESAWAT", "plane":
		return PowerSuperman
	default:
		return ""
	}
}

func cloneItemMap(src map[string]int) map[string]int {
	if src == nil {
		return nil
	}
	out := make(map[string]int, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func clonePowerCells(src map[int]string) map[int]string {
	if src == nil {
		return nil
	}
	out := make(map[int]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

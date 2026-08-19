package mmo

import (
	"strings"
	"time"
)

func (w *WorldState) ensureHousing() *HousingHub {
	if w.Housing == nil {
		w.Housing = &HousingHub{ByOwner: map[string]*HouseInstance{}, ByID: map[string]*HouseInstance{}}
	}
	if w.Housing.ByOwner == nil {
		w.Housing.ByOwner = map[string]*HouseInstance{}
	}
	if w.Housing.ByID == nil {
		w.Housing.ByID = map[string]*HouseInstance{}
	}
	return w.Housing
}

func (w *WorldState) ensureHouse(ownerID string) *HouseInstance {
	hub := w.ensureHousing()
	h := hub.ByOwner[ownerID]
	if h != nil {
		return h
	}
	h = &HouseInstance{
		ID: "house-" + ownerID, OwnerID: ownerID, Type: "house-small", Access: "PRIVATE",
		Items: []HouseItem{}, Return: map[string][3]float64{},
		LocationID: "village", LayoutID: "small", Name: "Rumah",
		CreatedAt: time.Now().UnixMilli(), Wall: "dawn", Floor: "wood", Roof: "tile", Light: "warm", Color: "cream",
		Plots: seedGardenPlots(4), Votes: map[string]int{}, VoteCat: map[string]string{},
	}
	hub.ByOwner[ownerID] = h
	hub.ByID[h.ID] = h
	return h
}

func (w *WorldState) houseOfPlayer(id string) *HouseInstance {
	if w.Housing == nil {
		return nil
	}
	if p := w.players[id]; p != nil && strings.HasPrefix(p.InstanceID, "house-") {
		if h := w.Housing.ByID[p.InstanceID]; h != nil {
			return h
		}
	}
	if h := w.Housing.ByOwner[id]; h != nil {
		return h
	}
	return nil
}

func houseID(id string) bool {
	return strings.HasPrefix(id, "house-")
}

func (w *WorldState) houseCanVisit(visitor *Player, h *HouseInstance) bool {
	if h == nil || visitor == nil {
		return false
	}
	if visitor.ID == h.OwnerID {
		return true
	}
	if h.GuildHall {
		g := w.guildOf(visitor.ID)
		return g != nil && h.OwnerID == "guild:"+g.ID
	}
	switch strings.ToUpper(h.Access) {
	case "PUBLIC":
		return true
	case "FRIENDS":
		return w.Social != nil && w.Social.Friends[h.OwnerID] != nil && func() bool { _, ok := w.Social.Friends[h.OwnerID][visitor.ID]; return ok }()
	case "GUILD":
		og := w.guildOf(h.OwnerID)
		vg := w.guildOf(visitor.ID)
		return og != nil && vg != nil && og.ID == vg.ID
	default:
		return false
	}
}

func (w *WorldState) houseEnter(p *Player, ownerID string) [][]byte {
	if ownerID == "" {
		ownerID = p.ID
	}
	if p.InstanceID != "" && !houseID(p.InstanceID) {
		return rejectFor(p.ID, TypeHouseEnter, "instance")
	}
	h := w.ensureHouse(ownerID)
	w.ensureHouseLife(h, p)
	if h.Locked && p.ID != h.OwnerID {
		return rejectFor(p.ID, TypeHouseEnter, "locked")
	}
	if !w.houseCanVisit(p, h) {
		return rejectFor(p.ID, TypeHouseEnter, "access")
	}
	if h.Return == nil {
		h.Return = map[string][3]float64{}
	}
	if p.InstanceID != h.ID {
		h.Return[p.ID] = [3]float64{p.X, p.Y, p.Z}
	}
	p.InstanceID = h.ID
	p.X, p.Y, p.Z = 0, 0, 0
	if p.ID == h.OwnerID && len(h.Items) == 0 && p.Bag.count("decor-chair") == 0 {
		w.giveItem(p, "decor-chair", 1)
		w.giveItem(p, "decor-plant", 1)
	}
	vis := map[string]bool{}
	for _, id := range h.Visitors {
		vis[id] = true
	}
	if !vis[p.ID] {
		h.Visitors = append(h.Visitors, p.ID)
	}
	if p.ID != h.OwnerID {
		log := p.ensurePhase22()
		h.GuestLog = append(h.GuestLog, HouseGuest{PlayerID: p.ID, At: time.Now().UnixMilli()})
		if len(h.GuestLog) > 20 {
			h.GuestLog = h.GuestLog[len(h.GuestLog)-20:]
		}
		if !log.MeetPlayers[h.OwnerID] {
			log.MeetPlayers[h.OwnerID] = true
			p.credit("MEET", "player", 1)
		}
		p.credit("VISIT", "house", 1)
		log.Flags["daily_visit"] = true
	}
	w.syncHouseDummy(h)
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeHouseState, w.houseView(h, p)), marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase))}
}

func (w *WorldState) houseLeave(p *Player) [][]byte {
	h := w.houseOfPlayer(p.ID)
	if h == nil || p.InstanceID != h.ID {
		p.InstanceID = ""
		return [][]byte{marshal(TypeHouseState, map[string]any{"left": true, "toId": p.ID})}
	}
	if pos, ok := h.Return[p.ID]; ok {
		p.X, p.Y, p.Z = pos[0], pos[1], pos[2]
		delete(h.Return, p.ID)
	}
	p.InstanceID = ""
	live := h.Visitors[:0]
	for _, id := range h.Visitors {
		if id != p.ID {
			live = append(live, id)
		}
	}
	h.Visitors = live
	return [][]byte{marshal(TypeHouseState, map[string]any{"left": true, "toId": p.ID})}
}

func (w *WorldState) housePlace(p *Player, slot int, itemID string, x, z, yaw float64) [][]byte {
	return w.housePlaceTx(p, slot, itemID, x, z, yaw, "")
}

func (w *WorldState) housePlaceTx(p *Player, slot int, itemID string, x, z, yaw float64, tx string) [][]byte {
	if tx != "" {
		if prev, ok := w.txSeen(tx); ok {
			return prev
		}
	}
	h := w.houseOfPlayer(p.ID)
	if h == nil {
		h = w.ensureHouse(p.ID)
	}
	if p.InstanceID != h.ID {
		return rejectFor(p.ID, TypeHousePlace, "inside")
	}
	if !w.canDecorateHouse(p, h) {
		return rejectFor(p.ID, TypeHousePlace, "owner")
	}
	p.ensurePhase22()
	p.ensureGear()
	limit := socialCfg.HouseObjectLimit
	if limit < 1 {
		limit = 40
	}
	if len(h.Items) >= limit {
		return rejectFor(p.ID, TypeHousePlace, "limit")
	}
	if slot < 0 || slot >= len(p.Bag.Slots) {
		return rejectFor(p.ID, TypeHousePlace, "slot")
	}
	s := p.Bag.Slots[slot]
	if s.ItemID == "" || (itemID != "" && s.ItemID != itemID) {
		return rejectFor(p.ID, TypeHousePlace, "owned")
	}
	def, ok := itemByID[s.ItemID]
	if !ok || def.Type != "DECOR" {
		return rejectFor(p.ID, TypeHousePlace, "decor")
	}
	x = snapHouseGrid(x)
	z = snapHouseGrid(z)
	if x < -8 {
		x = -8
	}
	if x > 8 {
		x = 8
	}
	if z < -8 {
		z = -8
	}
	if z > 8 {
		z = 8
	}
	if houseFurnitureBlocked(h, "", x, z) {
		return rejectFor(p.ID, TypeHousePlace, "overlap")
	}
	if !p.Bag.removeAt(slot, 1) {
		return rejectFor(p.ID, TypeHousePlace, "item")
	}
	item := HouseItem{ID: randomID("dec_"), ItemID: s.ItemID, X: x, Y: 0, Z: z, Yaw: yaw}
	h.Items = append(h.Items, item)
	p.ensureLog().FurnitureOwned[s.ItemID] = true
	p.ensureLog().Collections["furn:"+s.ItemID] = true
	p.credit("PLACE", "furniture", 1)
	w.syncHouseDummy(h)
	w.persistGear(p)
	w.persist(p)
	w.saveRuntimeLocked()
	out := [][]byte{marshal(TypeInventoryUpdated, p.loadout("Placed.", []int{slot})), marshal(TypeHouseState, w.houseView(h, p)), marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase))}
	if tx != "" {
		return w.rememberTx(tx, out)
	}
	return out
}

func (w *WorldState) houseRemove(p *Player, decorID string) [][]byte {
	h := w.houseOfPlayer(p.ID)
	if h == nil {
		h = w.ensureHouse(p.ID)
	}
	if !w.canDecorateHouse(p, h) {
		return rejectFor(p.ID, TypeHouseRemove, "owner")
	}
	idx := -1
	var it HouseItem
	for i, d := range h.Items {
		if d.ID == decorID {
			idx, it = i, d
			break
		}
	}
	if idx < 0 {
		return rejectFor(p.ID, TypeHouseRemove, "item")
	}
	if !w.giveItem(p, it.ItemID, 1) {
		return rejectFor(p.ID, TypeHouseRemove, "space")
	}
	h.Items = append(h.Items[:idx], h.Items[idx+1:]...)
	w.syncHouseDummy(h)
	w.persistGear(p)
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeInventoryUpdated, p.loadout("Removed.", nil)), marshal(TypeHouseState, w.houseView(h, p))}
}

func (w *WorldState) houseSetAccess(p *Player, access string) [][]byte {
	h := w.ensureHouse(p.ID)
	switch strings.ToUpper(access) {
	case "PRIVATE", "FRIENDS", "GUILD", "PUBLIC":
		h.Access = strings.ToUpper(access)
	default:
		return rejectFor(p.ID, TypeSetHouseAccess, "access")
	}
	w.saveRuntimeLocked()
	return [][]byte{marshal(TypeHouseState, w.houseView(h, p))}
}

func (w *WorldState) houseView(h *HouseInstance, watcher *Player) map[string]any {
	if h == nil {
		return map[string]any{"toId": watcher.ID}
	}
	to := ""
	if watcher != nil {
		to = watcher.ID
	}
	loc := h.LocationID
	if loc == "" {
		loc = "village"
	}
	name := h.Name
	if name == "" {
		name = "Rumah"
	}
	items := make([]map[string]any, 0, len(h.Items))
	for _, it := range h.Items {
		items = append(items, map[string]any{"id": it.ID, "itemId": it.ItemID, "x": it.X, "y": it.Y, "z": it.Z, "yaw": it.Yaw})
	}
	plots := make([]map[string]any, 0, len(h.Plots))
	for _, p := range h.Plots {
		plots = append(plots, map[string]any{"id": p.ID, "plant": p.Plant, "state": p.State, "plantedAt": p.PlantedAt, "readyAt": p.ReadyAt})
	}
	view := map[string]any{
		"instanceId": h.ID, "ownerId": h.OwnerID, "type": h.Type, "access": h.Access,
		"district": housingCfg.District, "items": items, "visitors": h.Visitors,
		"limit": socialCfg.HouseObjectLimit, "toId": to,
		"houseId": h.ID, "locationId": loc, "layoutId": h.LayoutID, "name": name,
		"createdAt": h.CreatedAt, "locked": h.Locked, "guildHall": h.GuildHall,
		"wall": h.Wall, "floor": h.Floor, "roof": h.Roof, "light": h.Light, "color": h.Color,
		"plots": plots, "rooms": houseRooms(h), "locations": houseLocations(),
		"guestLog": h.GuestLog, "grid": HouseGrid, "sign": name,
	}
	if watcher != nil && (watcher.ID == h.OwnerID || w.canDecorateHouse(watcher, h)) {
		if h.Storage != nil {
			view["storage"] = h.Storage.views()
		}
	}
	return view
}

func (w *WorldState) syncHouseDummy(h *HouseInstance) {
	if h == nil {
		return
	}
	has := false
	var at HouseItem
	for _, it := range h.Items {
		if it.ItemID == "decor-dummy" {
			has, at = true, it
			break
		}
	}
	for id, e := range w.enemies {
		if e != nil && e.InstanceID == h.ID && e.Def.ID == "training_dummy" && !has {
			delete(w.enemies, id)
		}
	}
	if !has {
		return
	}
	for _, e := range w.enemies {
		if e != nil && e.InstanceID == h.ID && e.Def.ID == "training_dummy" {
			e.X, e.Z = at.X, at.Z
			e.Alive = true
			e.HP = e.MaxHP
			return
		}
	}
	def := EnemyDef{ID: "training_dummy", Name: "Training Dummy", Level: 1, MaxHP: 400, Attack: 0, Defense: 4, Speed: 0, AttackRange: 0, AggroRange: 0, Behavior: "dummy"}
	for _, d := range enemyCatalog {
		if d.ID == "training_dummy" {
			def = d
			break
		}
	}
	e := spawnEnemy(def, at.X, at.Z)
	e.InstanceID = h.ID
	e.NoRespawn = true
	w.enemies[e.ID] = e
}

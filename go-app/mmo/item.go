package mmo

import (
	"encoding/json"
	"log"

	_ "embed"
)

//go:embed data/items.json
var itemsJSON []byte

type ItemDef struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Type             string      `json:"type"`
	Slot             string      `json:"slot"`
	Rarity           string      `json:"rarity"`
	Stackable        bool        `json:"stackable"`
	MaxStack         int         `json:"maxStack"`
	Icon             string      `json:"icon"`
	Value            int         `json:"value"`
	LevelRequirement int         `json:"levelRequirement"`
	Effects          ItemEffects `json:"effects"`
	Untradable       bool        `json:"untradable"`
}

var (
	itemCatalog []ItemDef
	itemByID    = map[string]ItemDef{}
)

func init() {
	if err := json.Unmarshal(itemsJSON, &itemCatalog); err != nil {
		log.Printf("mmo items.json: %v", err)
	}
	for _, it := range itemCatalog {
		itemByID[it.ID] = it
	}
	registerGuardianItems()
}

func (d ItemDef) view() ItemDefView {
	v := ItemDefView{
		ID: d.ID, Name: d.Name, Description: d.Description, Type: d.Type, Slot: d.Slot,
		Rarity: d.Rarity, Stackable: d.Stackable, MaxStack: d.MaxStack, Icon: d.Icon,
		Value: d.Value, LevelRequirement: d.LevelRequirement, Effects: d.Effects, Tradable: !d.Untradable,
		Bind: itemBind(d), ItemLevel: 1,
	}
	if dawnSetPieces[d.ID] {
		v.SetID = DawnSetID
	}
	if d.ID == "dawn_blade" {
		v.Lore = loreByID["lore-dawn-blade"].Text
	}
	if d.ID == "light_staff" {
		v.Lore = loreByID["lore-light-staff"].Text
	}
	return v
}

func catalogViews() []ItemDefView {
	out := make([]ItemDefView, 0, len(itemCatalog))
	for _, it := range itemCatalog {
		out = append(out, it.view())
	}
	return out
}

const (
	InvCapacity     = 50
	ChatGlobal      = "GLOBAL"
	ChatWorld       = "WORLD"
	ChatParty       = "PARTY"
	ChatPrivate     = "PRIVATE"
	PartyMaxSize    = 8
	PartyShareRange = 50.0
	NearbyRange     = 50.0
)

type InvSlot struct {
	ItemID     string
	Qty        int
	Locked     bool
	Favorite   bool
	InstanceID string
	ItemLevel  int
	Upgrade    int
	Enchants   []string
	Durability int
}

type Inventory struct {
	ID       string
	Capacity int
	Version  int
	Slots    []InvSlot
}

type EquipmentSet struct {
	HEAD, BODY, LEGS, WEAPON, ACCESSORY_1, ACCESSORY_2, ACCESSORY_3 string
}

func newInventory(playerID string) *Inventory {
	inv := &Inventory{ID: "inv-" + playerID, Capacity: InvCapacity, Slots: make([]InvSlot, InvCapacity)}
	inv.add("potion_heal", 3)
	inv.add("potion_energy", 1)
	inv.add("training_gloves", 1)
	return inv
}

func (inv *Inventory) bump() {
	inv.Version++
}

func (inv *Inventory) count(itemID string) int {
	n := 0
	for _, s := range inv.Slots {
		if s.ItemID == itemID {
			n += s.Qty
		}
	}
	return n
}

func (inv *Inventory) emptySlots() int {
	n := 0
	for _, s := range inv.Slots {
		if s.ItemID == "" {
			n++
		}
	}
	return n
}

func (inv *Inventory) add(itemID string, qty int) (changed []int, ok bool) {
	def, exists := itemByID[itemID]
	if !exists || qty <= 0 {
		return nil, false
	}
	max := def.MaxStack
	if max < 1 {
		max = 1
	}
	if def.Type == "MATERIAL" && def.Stackable && socialCfg.MaxStack > max {
		max = socialCfg.MaxStack
	}
	left := qty
	if def.Stackable {
		for i := range inv.Slots {
			if left <= 0 {
				break
			}
			s := &inv.Slots[i]
			if s.ItemID != itemID {
				continue
			}
			space := max - s.Qty
			if space <= 0 {
				continue
			}
			take := left
			if take > space {
				take = space
			}
			s.Qty += take
			left -= take
			changed = append(changed, i)
		}
	}
	for i := range inv.Slots {
		if left <= 0 {
			break
		}
		s := &inv.Slots[i]
		if s.ItemID != "" {
			continue
		}
		take := left
		if def.Stackable {
			if take > max {
				take = max
			}
		} else {
			take = 1
		}
		s.ItemID = itemID
		s.Qty = take
		if s.InstanceID == "" {
			s.InstanceID = randomID("itm_")
		}
		if s.ItemLevel < 1 {
			s.ItemLevel = 1
		}
		left -= take
		changed = append(changed, i)
	}
	if left > 0 {
		return changed, false
	}
	inv.bump()
	return changed, true
}

func (inv *Inventory) removeAt(index, qty int) bool {
	if index < 0 || index >= len(inv.Slots) || qty <= 0 {
		return false
	}
	s := &inv.Slots[index]
	if s.ItemID == "" || s.Qty < qty {
		return false
	}
	s.Qty -= qty
	if s.Qty <= 0 {
		*s = InvSlot{}
	}
	inv.bump()
	return true
}

func (inv *Inventory) takeItem(itemID string, qty int) bool {
	if inv.count(itemID) < qty {
		return false
	}
	left := qty
	for i := range inv.Slots {
		if left <= 0 {
			break
		}
		s := &inv.Slots[i]
		if s.ItemID != itemID {
			continue
		}
		take := s.Qty
		if take > left {
			take = left
		}
		inv.removeAt(i, take)
		left -= take
	}
	return left == 0
}

func (inv *Inventory) firstIndex(itemID string) int {
	for i, s := range inv.Slots {
		if s.ItemID == itemID && s.Qty > 0 {
			return i
		}
	}
	return -1
}

func cloneInv(src *Inventory) *Inventory {
	if src == nil {
		return nil
	}
	out := &Inventory{ID: src.ID, Capacity: src.Capacity, Version: src.Version, Slots: make([]InvSlot, len(src.Slots))}
	copy(out.Slots, src.Slots)
	return out
}

func (eq EquipmentSet) get(slot string) string {
	switch slot {
	case "HEAD":
		return eq.HEAD
	case "BODY":
		return eq.BODY
	case "LEGS":
		return eq.LEGS
	case "WEAPON":
		return eq.WEAPON
	case "ACCESSORY_1":
		return eq.ACCESSORY_1
	case "ACCESSORY_2":
		return eq.ACCESSORY_2
	case "ACCESSORY_3":
		return eq.ACCESSORY_3
	}
	return ""
}

func (eq *EquipmentSet) set(slot, itemID string) {
	switch slot {
	case "HEAD":
		eq.HEAD = itemID
	case "BODY":
		eq.BODY = itemID
	case "LEGS":
		eq.LEGS = itemID
	case "WEAPON":
		eq.WEAPON = itemID
	case "ACCESSORY_1":
		eq.ACCESSORY_1 = itemID
	case "ACCESSORY_2":
		eq.ACCESSORY_2 = itemID
	case "ACCESSORY_3":
		eq.ACCESSORY_3 = itemID
	}
}

func (eq EquipmentSet) view() EquipmentView {
	return EquipmentView{HEAD: eq.HEAD, BODY: eq.BODY, LEGS: eq.LEGS, WEAPON: eq.WEAPON, ACCESSORY_1: eq.ACCESSORY_1, ACCESSORY_2: eq.ACCESSORY_2, ACCESSORY_3: eq.ACCESSORY_3}
}

func (eq EquipmentSet) ids() []string {
	return []string{eq.HEAD, eq.BODY, eq.LEGS, eq.WEAPON, eq.ACCESSORY_1, eq.ACCESSORY_2, eq.ACCESSORY_3}
}

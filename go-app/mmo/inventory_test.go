package mmo

import (
	"encoding/json"
	"testing"
)

func TestStarterInventoryAndStack(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureGear()
	if p.Bag.count("potion_heal") != 3 {
		t.Fatalf("starter potion %d", p.Bag.count("potion_heal"))
	}
	w.useItem(p, p.Bag.firstIndex("potion_heal"))
	if p.Bag.count("potion_heal") != 2 {
		t.Fatalf("after use %d", p.Bag.count("potion_heal"))
	}
}

func TestGiveItemCheatRejected(t *testing.T) {
	w, p := testVillagePlayer()
	evs := w.ApplyInventory(p.ID, Envelope{Type: TypeGiveItem, Data: []byte(`{"itemId":"traveler_staff","quantity":9999}`)})
	if !rejectAction(evs, TypeGiveItem) {
		t.Fatal("GIVE_ITEM harus ditolak")
	}
	evs = w.ApplyInventory(p.ID, Envelope{Type: TypeGiveCurrency, Data: []byte(`{"coin":9999}`)})
	if !rejectAction(evs, TypeGiveCurrency) {
		t.Fatal("GIVE_CURRENCY harus ditolak")
	}
}

func TestPickupRequiresDrop(t *testing.T) {
	w, p := testVillagePlayer()
	raw, _ := json.Marshal(PickupIn{ItemID: "traveler_staff", Quantity: 9999})
	evs := w.ApplyInventory(p.ID, Envelope{Type: TypePickupItem, Data: raw})
	if !rejectAction(evs, TypePickupItem) {
		t.Fatal("pickup tanpa drop harus ditolak")
	}
}

func TestEquipTrainingGlovesRaisesAttack(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensureGear()
	p.applyDerived()
	before := p.finalStats().Attack
	idx := p.Bag.firstIndex("training_gloves")
	if idx < 0 {
		t.Fatal("gloves")
	}
	w.equipItem(p, idx)
	after := p.finalStats().Attack
	if after <= before {
		t.Fatalf("attack %d -> %d", before, after)
	}
	w.unequipItem(p, "ACCESSORY_1")
	if p.finalStats().Attack != before {
		t.Fatalf("unequip attack %d want %d", p.finalStats().Attack, before)
	}
}

func TestInventoryPersistsOnRelog(t *testing.T) {
	w, p := testVillagePlayer()
	w.giveItem(p, "potion_stamina", 2)
	w.persistGear(p)
	w.Remove(p)
	q := &Player{ID: p.ID, Name: "Raka", send: make(chan []byte, 8)}
	q.initCombat()
	ok, isNew := w.Add(q)
	if !ok || !isNew {
		t.Fatal("rejoin")
	}
	if q.Bag.count("potion_stamina") != 2 {
		t.Fatalf("persist stamina potion %d", q.Bag.count("potion_stamina"))
	}
}

func TestPartyInviteAcceptLeaderLeave(t *testing.T) {
	w, a := testVillagePlayer()
	b := &Player{ID: "p_b", Name: "Sinta", send: make(chan []byte, 8)}
	b.initCombat()
	w.Add(b)
	c := &Player{ID: "p_c", Name: "Sam", send: make(chan []byte, 8)}
	c.initCombat()
	w.Add(c)
	if rejectAction(w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_b"}`)}), TypePartyInvite) {
		t.Fatal("invite B")
	}
	w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_c"}`)})
	w.ApplyParty(c.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	pt := w.Parties.Of(a.ID)
	if pt == nil || len(pt.Members) != 3 {
		t.Fatalf("party size %+v", pt)
	}
	w.ApplyParty(a.ID, Envelope{Type: TypePartyLeave, Data: []byte(`{}`)})
	pt = w.Parties.Of(b.ID)
	if pt == nil || pt.LeaderID != b.ID {
		t.Fatalf("leader should pass to B, got %+v", pt)
	}
}

func TestPartyKickRejectedForMember(t *testing.T) {
	w, a := testVillagePlayer()
	b := &Player{ID: "p_b", Name: "Sinta", send: make(chan []byte, 8)}
	b.initCombat()
	w.Add(b)
	w.ApplyParty(a.ID, Envelope{Type: TypePartyInvite, Data: []byte(`{"targetId":"p_b"}`)})
	w.ApplyParty(b.ID, Envelope{Type: TypePartyAccept, Data: []byte(`{}`)})
	evs := w.ApplyParty(b.ID, Envelope{Type: TypePartyKick, Data: []byte(`{"targetId":"p_a"}`)})
	if !rejectAction(evs, TypePartyKick) {
		t.Fatal("member tidak boleh kick")
	}
	evs = w.ApplyParty(b.ID, Envelope{Type: TypePartyDisband, Data: []byte(`{}`)})
	if !rejectAction(evs, TypePartyDisband) {
		t.Fatal("member tidak boleh disband")
	}
}

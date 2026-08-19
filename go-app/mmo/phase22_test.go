package mmo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPhase22LandminesKept(t *testing.T) {
	if questionByID["q-add-4-3"].ID == "" || questionByID["q-add-4-3"].Correct != 1 {
		t.Fatal("keep q-add-4-3")
	}
	if questionByID["q-kursi-4-2"].Prompt == "" || questionByID["q-kursi-4-2"].Correct != 1 {
		t.Fatal("4 kursi + 2 = 6 index B")
	}
	if regionByID["village"].Name != "Dawn City" {
		t.Fatal("keep Dawn City")
	}
	if mountByID["wind-runner"].Name != "Wind Runner" {
		t.Fatal("keep wind-runner")
	}
	if shopCatalog.ID != "dawn-merchant" {
		t.Fatal("shops.json shape")
	}
	if eventByID["mount-festival"].ID == "" || eventByID["festival-karya"].ID == "" {
		t.Fatal("keep festivals")
	}
	if eventByID["open-house"].Name != "OPEN HOUSE" {
		t.Fatal("open house")
	}
	if MaxPlayers < 100 {
		t.Fatal("100 players")
	}
}

func TestPhase22CreateHouseRejected(t *testing.T) {
	w, p := testVillagePlayer()
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeCreateHouse, Data: []byte(`{}`)}), TypeCreateHouse) {
		t.Fatal("create house")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeAddPet, Data: []byte(`{"petId":"dawn-pup"}`)}), TypeAddPet) {
		t.Fatal("add pet")
	}
}

func TestPhase22FurnitureDuplicateRejected(t *testing.T) {
	w, p := testVillagePlayer()
	w.giveItem(p, "decor-chair", 2)
	if rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeHouseEnter, Data: []byte(`{}`)}), TypeHouseEnter) {
		t.Fatal("enter")
	}
	slot := p.Bag.firstIndex("decor-chair")
	place, _ := json.Marshal(map[string]any{"slot": slot, "itemId": "decor-chair", "x": 1.0, "z": 1.0, "yaw": 0, "transactionId": "pl1"})
	if rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeHousePlace, Data: place}), TypeHousePlace) {
		t.Fatal("place1")
	}
	slot = p.Bag.firstIndex("decor-chair")
	dup, _ := json.Marshal(map[string]any{"slot": slot, "itemId": "decor-chair", "x": 1.0, "z": 1.0, "yaw": 0, "transactionId": "pl2"})
	if !rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeHousePlace, Data: dup}), TypeHousePlace) {
		t.Fatal("overlap should reject")
	}
	if len(w.ensureHouse(p.ID).Items) != 1 {
		t.Fatal("dup furniture")
	}
}

func TestPhase22PetClaimOnce(t *testing.T) {
	w, p := testVillagePlayer()
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypePetClaim, Data: []byte(`{"petId":"dawn-pup","transactionId":"pet1"}`)}), TypePetClaim) {
		t.Fatal("claim")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypePetClaim, Data: []byte(`{"petId":"dawn-pup","transactionId":"pet2"}`)}), TypePetClaim) {
		t.Fatal("second claim")
	}
	if len(p.ensureLog().Pets) != 1 {
		t.Fatal("dup pet")
	}
}

func TestPhase22FarmingPersistAndHarvestOnce(t *testing.T) {
	w, p := testVillagePlayer()
	p.addMaterial("mat-dawn-berry", 2)
	if rejectAction(w.ApplyPhase14(p.ID, Envelope{Type: TypeHouseEnter, Data: []byte(`{}`)}), TypeHouseEnter) {
		t.Fatal("enter")
	}
	h := w.ensureHouse(p.ID)
	plot := h.Plots[0].ID
	plant, _ := json.Marshal(map[string]any{"plotId": plot, "plantId": "dawn-berry", "transactionId": "gp1"})
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGardenPlant, Data: plant}), TypeGardenPlant) {
		t.Fatal("plant")
	}
	blob := w.exportRuntime()
	w2 := NewWorldState()
	w2.importRuntime(blob)
	h2 := w2.Housing.ByOwner[p.ID]
	if h2 == nil || len(h2.Plots) < 1 || h2.Plots[0].State != "SEEDED" {
		t.Fatal("plant missing after reconnect")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGardenHarvest, Data: []byte(`{"plotId":"` + plot + `","transactionId":"gh0"}`)}), TypeGardenHarvest) {
		t.Fatal("early harvest")
	}
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGardenWater, Data: []byte(`{"plotId":"` + plot + `"}`)}), TypeGardenWater) {
		t.Fatal("water")
	}
	time.Sleep(4 * time.Millisecond)
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGardenHarvest, Data: []byte(`{"plotId":"` + plot + `","transactionId":"gh1"}`)}), TypeGardenHarvest) {
		t.Fatal("harvest")
	}
	got := p.materialCount("mat-dawn-berry")
	if got != 2 {
		t.Fatalf("harvest bag %d", got)
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeGardenHarvest, Data: []byte(`{"plotId":"` + plot + `","transactionId":"gh2"}`)}), TypeGardenHarvest) {
		t.Fatal("second harvest")
	}
	if p.materialCount("mat-dawn-berry") != 2 {
		t.Fatal("dup harvest")
	}
	if farmLevel(p.ensureLog().LifeFarmXP) < 1 {
		t.Fatal("farming xp")
	}
}

func TestPhase22HousePermissionAndVisitorSecurity(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_v", "Kira")
	if rejectAction(w.ApplyPhase14(a.ID, Envelope{Type: TypeHouseEnter, Data: []byte(`{}`)}), TypeHouseEnter) {
		t.Fatal("owner enter")
	}
	w.ApplyPhase14(a.ID, Envelope{Type: TypeHouseLeave, Data: []byte(`{}`)})
	if !rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeHouseVisit, Data: []byte(`{"ownerId":"p_a"}`)}), TypeHouseEnter) {
		t.Fatal("private visit")
	}
	if rejectAction(w.ApplyPhase14(a.ID, Envelope{Type: TypeSetHouseAccess, Data: []byte(`{"access":"PUBLIC"}`)}), TypeSetHouseAccess) {
		t.Fatal("public")
	}
	if rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeHouseVisit, Data: []byte(`{"ownerId":"p_a"}`)}), TypeHouseEnter) {
		t.Fatal("public visit")
	}
	if b.InstanceID == "" {
		t.Fatal("visitor missing")
	}
	h := w.ensureHouse(a.ID)
	found := false
	for _, id := range h.Visitors {
		if id == b.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("guest not listed")
	}
	if !rejectAction(w.ApplyEconomy(b.ID, Envelope{Type: TypeHouseTake, Data: []byte(`{"slot":0,"qty":1}`)}), TypeHouseTake) {
		t.Fatal("visitor take")
	}
	if !rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeHouseRemove, Data: []byte(`{"decorId":"x"}`)}), TypeHouseRemove) {
		t.Fatal("visitor remove")
	}
}

func TestPhase22FriendVisitGuestLog(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_fr", "Nara")
	w.ApplySocial(a.ID, Envelope{Type: TypeFriendRequest, Data: []byte(`{"targetId":"p_fr"}`)})
	w.ApplySocial(b.ID, Envelope{Type: TypeAcceptFriend, Data: []byte(`{"targetId":"p_a"}`)})
	w.ApplyPhase14(a.ID, Envelope{Type: TypeSetHouseAccess, Data: []byte(`{"access":"FRIENDS"}`)})
	if rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeHouseVisit, Data: []byte(`{"ownerId":"p_a"}`)}), TypeHouseEnter) {
		t.Fatal("friend visit")
	}
	h := w.ensureHouse(a.ID)
	if len(h.GuestLog) < 1 || h.GuestLog[0].PlayerID != b.ID {
		t.Fatal("guest log")
	}
}

func TestPhase22GuildStoragePermission(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_mem", "Sam")
	a.Level = 10
	a.ensureLog().Coin = 1000
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildCreate, Data: []byte(`{"name":"Hall Guard","tag":"HALL"}`)})
	w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildInvite, Data: []byte(`{"targetId":"p_mem"}`)})
	w.ApplyEconomy(b.ID, Envelope{Type: TypeGuildAccept, Data: []byte(`{}`)})
	w.giveItem(a, "iron_ore", 2)
	slot := a.Bag.firstIndex("iron_ore")
	dep, _ := json.Marshal(map[string]any{"slot": slot, "qty": 1})
	w.ApplyPhase14(a.ID, Envelope{Type: TypeGuildDeposit, Data: dep})
	g := w.guildOf(a.ID)
	st := g.Storage.firstIndex("iron_ore")
	wd, _ := json.Marshal(map[string]any{"slot": st, "qty": 1})
	if !rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeGuildWithdraw, Data: wd}), TypeGuildWithdraw) {
		t.Fatal("member withdraw")
	}
	if rejectAction(w.ApplyEconomy(a.ID, Envelope{Type: TypeGuildHallEnter, Data: []byte(`{}`)}), TypeGuildHallEnter) {
		t.Fatal("hall enter")
	}
	if !houseID(a.InstanceID) || w.Housing.ByID[a.InstanceID] == nil || !w.Housing.ByID[a.InstanceID].GuildHall {
		t.Fatal("guild hall instance")
	}
}

func TestPhase22PetFollowAndCare(t *testing.T) {
	w, p := testVillagePlayer()
	w.ApplyEconomy(p.ID, Envelope{Type: TypePetClaim, Data: []byte(`{"petId":"dawn-pup","transactionId":"pc1"}`)})
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypePetSummon, Data: []byte(`{"petId":"dawn-pup"}`)}), TypePetSummon) {
		t.Fatal("summon")
	}
	if p.PetID != "dawn-pup" || p.ensureLog().ActivePet != "dawn-pup" {
		t.Fatal("follow")
	}
	before := p.ensureLog().PetHappy["dawn-pup"]
	p.addMaterial("mat-dawn-berry", 1)
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypePetCare, Data: []byte(`{"petId":"dawn-pup","action":"feed"}`)}), TypePetCare) {
		t.Fatal("feed")
	}
	if p.ensureLog().PetHappy["dawn-pup"] <= before {
		t.Fatal("happiness")
	}
}

func TestPhase22DailyQuestOnce(t *testing.T) {
	w, p := testVillagePlayer()
	p.ensurePhase22().Flags["daily_plant"] = true
	if rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeClaimDailyLife, Data: []byte(`{"questId":"plant"}`)}), TypeClaimDailyLife) {
		t.Fatal("claim")
	}
	if !rejectAction(w.ApplyEconomy(p.ID, Envelope{Type: TypeClaimDailyLife, Data: []byte(`{"questId":"plant"}`)}), TypeClaimDailyLife) {
		t.Fatal("second daily")
	}
}

func TestPhase22LifeQuizAndVote(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_vote", "Bima")
	edu := a.ensureLog().EduToken
	if rejectAction(w.ApplyEconomy(a.ID, Envelope{Type: TypeLifeQuiz, Data: []byte(`{"questionId":"q-kursi-4-2","choice":1}`)}), TypeLifeQuiz) {
		t.Fatal("quiz")
	}
	if a.ensureLog().EduToken <= edu {
		t.Fatal("knowledge token")
	}
	w.ApplyPhase14(a.ID, Envelope{Type: TypeHouseEnter, Data: []byte(`{}`)})
	w.ApplyPhase14(a.ID, Envelope{Type: TypeHouseLeave, Data: []byte(`{}`)})
	if rejectAction(w.ApplyEconomy(b.ID, Envelope{Type: TypeHouseVote, Data: []byte(`{"ownerId":"p_a","category":"COZY","score":5}`)}), TypeHouseVote) {
		t.Fatal("vote")
	}
	if w.ensureHouse(a.ID).Votes[b.ID] != 5 {
		t.Fatal("vote stored")
	}
	if !rejectAction(w.ApplyEconomy(b.ID, Envelope{Type: TypeHouseVote, Data: []byte(`{"ownerId":"p_a","category":"COZY","score":4}`)}), TypeHouseVote) {
		t.Fatal("double vote")
	}
}

func TestPhase22HouseLockAndLockedVisit(t *testing.T) {
	w, a := testVillagePlayer()
	b := addWorldPlayer(w, "p_lk", "Lio")
	w.ApplyPhase14(a.ID, Envelope{Type: TypeSetHouseAccess, Data: []byte(`{"access":"PUBLIC"}`)})
	if rejectAction(w.ApplyEconomy(a.ID, Envelope{Type: TypeHouseLock, Data: []byte(`{"on":true}`)}), TypeHouseLock) {
		t.Fatal("lock")
	}
	if !rejectAction(w.ApplyPhase14(b.ID, Envelope{Type: TypeHouseVisit, Data: []byte(`{"ownerId":"p_a"}`)}), TypeHouseEnter) {
		t.Fatal("locked visit")
	}
}

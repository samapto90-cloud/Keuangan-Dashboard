package mmo

import (
	"math"
	"time"
)

func (w *WorldState) markGuardianDefeated(p *Player, bossID string) [][]byte {
	g, ok := guardianByID[bossID]
	if !ok {
		for _, cand := range guardianCatalog {
			if cand.ChapterID != "" && cand.ID == bossID {
				g, ok = cand, true
				break
			}
		}
	}
	if !ok {
		return nil
	}
	log := p.ensureLog()
	if log.Guardians == nil {
		log.Guardians = map[string]bool{}
	}
	if log.Guardians[g.ID] {
		return nil
	}
	log.Guardians[g.ID] = true
	log.Flags["guardian_"+g.ID+"_defeated"] = true
	w.noteActivity(p, "GUARDIAN", g.ID, 1)
	if log.GuardianTimes == nil {
		log.GuardianTimes = map[string]int64{}
	}
	log.GuardianTimes[g.ID] = time.Now().UnixMilli()
	log.GuardianTokens++
	p.credit("BOSS_DEFEATED", g.ID, 1)
	n := 0
	for _, ok := range log.Guardians {
		if ok {
			n++
		}
	}
	w.grantGuardianCollection(p, n)
	if g.ChapterID != "" {
		w.unlockChapter(p, g.ChapterID)
	}
	if g.ID == "avaron" {
		log.Flags["celestial_gate_unlocked"] = true
		log.Flags["masjid_path_open"] = true
		log.Flags["openedMountainGate"] = true
	}
	if g.UniqueItem != "" {
		w.giveItem(p, g.UniqueItem, 1)
	}
	events := w.refreshAchievements(p)
	p.markDirty()
	w.persist(p)
	w.persistGear(p)
	w.audit("guardianDefeated", p.ID, g.ID)
	out := [][]byte{
		marshal(TypeGuardianDefeated, GuardianView{
			ID: g.ID, Name: g.Name, Title: g.Title, Status: "DEFEATED", ChapterID: g.ChapterID, Region: g.Region, PlayerID: p.ID,
			Personality: g.Personality, Weakness: g.Weakness, Story: g.Story, DefeatedAt: log.GuardianTimes[g.ID],
		}),
		marshal(TypeChatMessage, ChatOut{Channel: "SYSTEM", From: "SYSTEM", Text: g.Name + " has been defeated!", System: true}),
	}
	if g.ID == "avaron" {
		out = append(out, marshal(TypeChatMessage, ChatOut{Channel: "SYSTEM", From: "SYSTEM", Text: "Avaron: Perjalananmu belum selesai.", System: true}))
	}
	out = append(out, events...)
	out = append(out, w.afterGuardianStory(p, g)...)
	return out
}

func (w *WorldState) seedWorldGuardians() {
	byEnemy := map[string]EnemyDef{}
	for _, d := range enemyCatalog {
		byEnemy[d.ID] = d
	}
	for i, g := range guardianCatalog {
		def, ok := byEnemy[g.ID]
		if !ok {
			def = EnemyDef{ID: g.ID, Name: g.Name, Level: g.Level, MaxHP: 80 + g.Level*20, Attack: 8 + g.Level, Defense: 4, Speed: 1.6, AttackRange: 1.8, AggroRange: 8, Leash: 14, AttackCooldown: 1.6, ExpReward: 40 + g.Level*4, Behavior: "boss", Rank: "boss"}
		}
		x, z := g.X, g.Z
		if math.Abs(x) < 0.2 {
			x = float64((i%2)*2-1) * (5.5 + float64(i%3)*0.8)
		}
		if z < 22 {
			z = 26 + float64(i%5)*1.4
		}
		e := spawnEnemy(def, x, z)
		e.NoRespawn = true
		w.enemies[e.ID] = e
	}
}

func elementalBonus(element, attackType string) float64 {
	if element == "CELESTIAL" || element == "LIGHT" {
		if attackType == "ENERGY" {
			return 0.08
		}
	}
	if element == "STONE" || element == "IRON" {
		if attackType == "PHYSICAL" || attackType == "light" || attackType == "heavy" {
			return -0.08
		}
	}
	return 0
}

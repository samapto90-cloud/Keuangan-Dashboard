package mmo

import "os"

// Phase 30 overlay: integration, security, ops, and release metadata.
// Reuses WorldService, QuestService, StoryService, BossService, DungeonService,
// RaidService, EventService, PartyService, GuildService, CombatService,
// TransformationService, AchievementService, LeaderboardService,
// NotificationService, NPCService, RewardService, EconomyService.
// Do not add duplicates. Do not change gameplay balance.

const (
	GameVersion = "0.1.0-phase1"
	GamePhase   = "ULAR/1"
	GameTitle   = "Ular Tangga Nusantara"
)

func init() {
	registerPhase30()
}

func registerPhase30() {
	if os.Getenv("SIPKEU_ENV") == "production" || os.Getenv("CAHAYA_ENV") == "production" {
		_ = os.Setenv("CAHAYA_RELEASE", GameVersion)
	}
}

func phase30Cheat(t string) bool {
	switch t {
	case TypeGiveGold, TypeSetGold, TypeAddGold, TypeGiveItem, TypeAddItem, TypeSetQuantity,
		TypeSetDamage, TypeSetEnergy, TypeSetLevel, TypeSetBossHP, TypeSetBossDead,
		TypeCompleteDungeon, TypeUnlockAchievement, TypeSetSeasonXP, TypeContributeEvent,
		TypeSetStoryFlag, TypeDefeatSiluman, TypeClaimStoryReward, TypeTeleport,
		TypeSetMountSpeed, TypeSetCoin, TypeDuplicateInstance:
		return true
	default:
		return false
	}
}

func (w *WorldState) ApplyPhase30(p *Player, env Envelope) [][]byte {
	if p == nil {
		return nil
	}
	if phase30Cheat(env.Type) {
		return rejectFor(p.ID, env.Type, "server_authoritative")
	}
	switch env.Type {
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func phase30EnrichEndgameView(view map[string]any) {
	if view == nil {
		return
	}
	view["version"] = GameVersion
	view["phase"] = GamePhase
	view["release"] = "READY_FOR_EXTERNAL_QA"
	view["title"] = GameTitle
}

func phase30ReleaseView() map[string]any {
	return map[string]any{
		"title":   GameTitle,
		"version": GameVersion,
		"phase":   GamePhase,
		"status":     "PHASE_1_FOUNDATION",
		"forms":      []string{},
		"siluman":    0,
		"regions":    0,
		"chapters":   0,
		"boardSize":  BOARD_SIZE,
		"maxPlayers": MAX_PLAYERS,
	}
}

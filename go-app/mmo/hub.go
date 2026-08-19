package mmo

import (
	"encoding/json"
	"log"
	"time"
)

type inbound struct {
	player *Player
	env    Envelope
}

type Hub struct {
	Auth     *AuthStore
	Accounts *AccountStore
	Players  *DiskPlayerStore
	World    *WorldState
	Lobby    *UlarLobby
	join     chan *Player
	leave    chan *Player
	in       chan inbound
}

func NewHub() *Hub {
	world := NewWorldState()
	store := OpenDiskPlayerStore(playerStoreDir())
	world.QuestRepo = store.AsQuest()
	world.InvRepo = store.AsInv()
	world.JourneyRepo = store
	world.loadRuntimeFile()
	return &Hub{
		Auth:     NewAuthStore(),
		Accounts: OpenAccountStore(accountStorePath()),
		Players:  store,
		World:    world,
		Lobby:    NewUlarLobby(),
		join:     make(chan *Player, 16),
		leave:    make(chan *Player, 16),
		in:       make(chan inbound, 256),
	}
}

func (h *Hub) Run() {
	tick := time.NewTicker(time.Second / time.Duration(ServerTickRate))
	defer tick.Stop()
	dt := 1.0 / float64(ServerTickRate)
	for {
		select {
		case p := <-h.join:
			if !AdventureGameplayEnabled {
				h.lobbyJoin(p)
				continue
			}
			ok, isNew := h.World.Add(p)
			if !ok {
				select {
				case p.send <- marshal(TypeError, ErrorOut{Message: "channel penuh"}):
				default:
				}
				continue
			}
			welcome := WelcomeOut{
				WorldID:     h.World.WorldID,
				Channel:     h.World.ChannelID,
				TickRate:    ServerTickRate,
				PlayerID:    p.ID,
				SessionID:   p.SessionID,
				Self:        p.Spawn(),
				Players:     h.World.SpawnsExcept(p.ID),
				Snapshot:    h.World.SnapshotFor(p.ID),
				Progress:    h.World.ProgressFor(p.ID),
				Loadout:     h.World.LoadoutFor(p.ID),
				Catalog:     catalogViews(),
				Social:      h.World.SocialFor(p.ID),
				Progression: h.World.ProgressionFor(p.ID),
			}
			select {
			case p.send <- marshal(TypeWelcome, welcome):
			default:
			}
			if isNew {
				h.World.BroadcastScope(p.ID, marshal(TypePlayerSpawn, p.Spawn()))
			}
			log.Printf("mmo spawn %s (%s) world=%s channel=%s online=%d new=%v", p.ID, p.Name, h.World.WorldID, h.World.ChannelID, h.World.Count(), isNew)
		case p := <-h.leave:
			if !AdventureGameplayEnabled {
				h.lobbyLeave(p)
				continue
			}
			if h.World.Remove(p) != nil {
				h.World.BroadcastScope(p.ID, marshal(TypePlayerDespawn, DespawnOut{PlayerID: p.ID}))
				log.Printf("mmo despawn %s online=%d", p.ID, h.World.Count())
			}
			h.World.SaveRuntime()
			p.CloseSend()
		case msg := <-h.in:
			h.handle(msg)
		case <-tick.C:
			if !AdventureGameplayEnabled {
				continue
			}
			events := h.World.Simulate(dt)
			for _, ev := range events {
				h.World.BroadcastScope("", ev)
			}
			h.World.BroadcastSnapshots()
			for _, stale := range h.World.DropTimedOut() {
				h.World.BroadcastScope(stale.ID, marshal(TypePlayerDespawn, DespawnOut{PlayerID: stale.ID}))
				log.Printf("mmo timeout %s online=%d", stale.ID, h.World.Count())
				stale.CloseSend()
			}
		}
	}
}

func (h *Hub) handle(msg inbound) {
	if !AdventureGameplayEnabled {
		h.handlePhase1(msg)
		return
	}
	p := msg.player
	live := h.World.Get(p.ID)
	if live != nil && live != p {
		return
	}
	switch msg.env.Type {
	case TypeMoveInput:
		if live == nil {
			return
		}
		var body MoveInput
		if json.Unmarshal(msg.env.Data, &body) != nil {
			return
		}
		live.SetInput(body)
	case TypePlayerAttack, TypePlayerSkill, TypePlayerDodge, TypePlayerCombo, TypePlayerRespawn, TypePlayerBlock, TypePlayerCounter, TypePlayerCharge, TypeSetEnergy, TypeSetCooldown, TypeUnlockTransform:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyCombat(p.ID, msg.env) {
			h.World.BroadcastScope(p.ID, ev)
		}
		if payload := h.World.FlushQuest(p.ID); payload != nil {
			h.World.SendTo(p.ID, payload)
		}
		h.flushPartyQuest(p.ID)
	case TypeInteract, TypeQuestAccept, TypeQuestDecline, TypeQuestClaim,
		TypeEducationAnswer, TypeHeal, TypeShopOpen, TypeQuestComplete,
		TypeCollectItem, TypeForestUnlock, TypeEducationCorrect:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyWorld(p.ID, msg.env) {
			if worldBroadcast(ev) {
				h.World.BroadcastAll(ev)
			} else {
				h.World.SendTo(p.ID, ev)
			}
		}
		if payload := h.World.FlushQuest(p.ID); payload != nil {
			h.World.SendTo(p.ID, payload)
		}
	case TypeFastTravel, TypeGetWorldJournal, TypeRequestMount, TypeDismount,
		TypeSkipCinematic, TypeJoinWorldEvent, TypeClaimEventReward,
		TypeUnlockGuardian, TypeUnlockRegion, TypeSetWeather, TypeTeleport,
		TypeSwitchChannel, TypeChannelList, TypeClaimWorldBoss, TypeGetCollections, TypeManualSave,
		TypeTriggerWorldBoss, TypeSetWorldBossHP, TypeSetWorldTime, TypeSpawnTreasure, TypeStartWorldEvent,
		TypeSetLanguage, TypeGetStory, TypeStoryChoice, TypeClaimStoryChapter, TypeReplayCinematic,
		TypeReplayChapter, TypeStartNGPlus, TypeCinematicDone, TypeSetStoryFlag, TypeUnlockStoryChapter,
		TypeDefeatSiluman, TypeClaimStoryReward, TypeGetMounts, TypeFavoriteMount, TypeEquipMount,
		TypeSetMountCosmetic, TypeMountEmote, TypeUnlockMount, TypeGrantMount, TypeClaimMount,
		TypeSetMountSpeed, TypeRaceStart, TypeRaceCheckpoint, TypeRaceFinish, TypeTravelEvent,
		TypeInspectLandmark, TypeGetOpenWorld, TypeGetAdventure, TypeSetRelationship,
		TypeAddRelationship, TypeSetNpcMemory, TypeUnlockLore, TypeCompleteQuest,
		TypeSetWorldState, TypeClaimEducationReward:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyExplore(p.ID, msg.env) {
			if worldBroadcast(ev) {
				h.World.BroadcastScope(p.ID, ev)
			} else {
				h.World.SendTo(p.ID, ev)
			}
		}
		if payload := h.World.FlushQuest(p.ID); payload != nil {
			h.World.SendTo(p.ID, payload)
		}
	case TypePickupItem, TypeUseItem, TypeEquipItem, TypeUnequipItem, TypeDiscardItem,
		TypeGetInventory, TypeGiveItem, TypeGiveCurrency, TypeAddItem, TypeSetQuantity,
		TypeSplitStack, TypeUpgradeItem, TypeEnchantItem, TypeExpandBag, TypeSaveGearLoadout,
		TypeLoadGearLoadout, TypeClaimTempLoot, TypeSalvageItems, TypeToggleCosmetic,
		TypeTraceItem, TypeSetItemStats, TypeSetItemLevel, TypeDuplicateInstance, TypeSetInstanceID:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyInventory(p.ID, msg.env) {
			h.World.SendTo(p.ID, ev)
		}
	case TypePartyInvite, TypePartyAccept, TypePartyDecline, TypePartyLeave,
		TypePartyKick, TypePartyDisband, TypePartySetTarget, TypePartyTransfer, TypePartySetRole, TypePartyReady,
		TypePartyFinderList, TypePartyFinderCreate, TypePartyFinderJoin, TypePartySetWaypoint, TypeFollowParty, TypePartyCreate:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyParty(p.ID, msg.env) {
			h.routeSocial(p.ID, ev)
		}
	case TypeFriendRequest, TypeAcceptFriend, TypeDeclineFriend, TypeRemoveFriend,
		TypeBlockPlayer, TypeUnblockPlayer, TypeInspectPlayer, TypeGetSocial,
		TypeNearbyPlayers, TypeChat, TypeSetPrivacy, TypeGetPrivacy, TypeSetPresence, TypeLocalMute, TypeUnmuteLocal,
		TypeReportMessage, TypeSetFriendship, TypeSetPartyLeader, TypeSetTradeItem, TypeSetChatSender, TypeSetGuildMaster:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplySocial(p.ID, msg.env) {
			h.routeSocial(p.ID, ev)
		}
	case TypeGuildCreate, TypeGuildInvite, TypeGuildAccept, TypeGuildDecline, TypeGuildLeave,
		TypeGuildKick, TypeGuildDisband, TypeGuildTransfer, TypeGuildAnnounce, TypeSetGuildRank, TypeGetGuild,
		TypeTradeRequest, TypeTradeAccept, TypeTradeDecline, TypeTradeOffer, TypeTradeReady, TypeTradeConfirm, TypeTradeCancel,
		TypeShopBuy, TypeShopSell, TypeSetCoin, TypeSetTitle, TypeSetCosmetic,
		TypeReportPlayer, TypeSearchPlayer, TypeGetNotifies, TypeMutePlayer, TypeModKick, TypeModBan,
		TypeGetMarket, TypeMarketSearch, TypeMarketList, TypeMarketBuy, TypeMarketCancel,
		TypeBankDeposit, TypeBankWithdraw, TypeGetBank, TypeLockItem, TypeFavoriteItem,
		TypeHouseEnter, TypeHouseLeave, TypeHouseVisit, TypeHousePlace, TypeHouseRemove, TypeSetHouseAccess, TypeGetHouse,
		TypeGuildApply, TypeGuildReview, TypeGuildDeposit, TypeGuildWithdraw, TypeGetGuildLog, TypeGuildSetEmblem, TypeGuildSetDesc,
		TypeSocialEmote, TypeGetPlayerCard, TypeSetName, TypeSetGuildXP, TypeSetOwned, TypeGetEconomy,
		TypeGather, TypeCraft, TypeGetCrafting, TypeSetProfession, TypeResetProfession,
		TypeFishStart, TypeFishCatch, TypeNpcShopOpen, TypeNpcShopBuy, TypeNpcShopSell, TypeNpcRepair,
		TypeStallOpen, TypeStallList, TypeStallBuy, TypeStallClose, TypeCraftOrder, TypeCraftOrderAccept,
		TypeGetWorkshop, TypeAddGold, TypeAddMaterial,
		TypeSetPrice, TypeSetGold, TypeGiveGold, TypeCreateRecipe, TypeGuildContribute, TypeGetGoldLog,
		TypeCreateHouse, TypeHouseLock, TypeHouseRename, TypeHouseStyle, TypeHouseMove, TypeHouseDecorate,
		TypeHouseStore, TypeHouseTake, TypeGardenPlant, TypeGardenWater, TypeGardenHarvest,
		TypePetClaim, TypePetSummon, TypePetDismiss, TypePetCare, TypePetName, TypeGetPets, TypeAddPet,
		TypeGuildHallEnter, TypeGuildHallLeave, TypeGuildHost, TypeGetLife, TypeClaimDailyLife,
		TypeHouseVote, TypeLifeQuiz, TypeClaimCollection:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyEconomy(p.ID, msg.env) {
			h.routeSocial(p.ID, ev)
		}
	case TypeAllocateAttribute, TypeResetAttributes, TypeUnlockSkill, TypeGetProgression,
		TypeRequestTransformation, TypeSetTransformation, TypeSetLevel, TypeSetSkillPoints,
		TypeSetCombatStyle, TypeSaveBuild, TypeLoadBuild, TypeSwitchBuild, TypeSetLoadout, TypeResetSkills, TypeGetBuilds:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyProgression(p.ID, msg.env) {
			if worldBroadcast(ev) {
				h.World.BroadcastAll(ev)
			} else {
				h.World.SendTo(p.ID, ev)
			}
		}
		if payload := h.World.FlushQuest(p.ID); payload != nil {
			h.World.SendTo(p.ID, payload)
		}
	case TypeDungeonEnter, TypeDungeonReady, TypeDungeonLeave, TypeDungeonAbandon, TypeDungeonRetry,
		TypeClaimLoot, TypeGetChapters, TypeUnlockChapter, TypeSetBossHP, TypeSetChapter,
		TypeGetDungeons, TypeQueueJoin, TypeQueueLeave, TypeDungeonJoin, TypeDungeonFill,
		TypeDungeonRevive, TypeDungeonVote, TypeDungeonTaunt, TypeSkipDungeonIntro,
		TypeSetObjective, TypeGiveLoot, TypeObjectiveComplete, TypeGetDungeonHistory, TypeRaidExchange,
		TypeSetBossDead, TypeSpawnBoss, TypeCompleteDungeon, TypeSkipMechanic, TypeDamageBoss:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyDungeon(p.ID, msg.env) {
			h.routeDungeon(p.ID, ev)
		}
		if payload := h.World.FlushQuest(p.ID); payload != nil {
			h.World.SendTo(p.ID, payload)
		}
	case TypeGetPvp, TypePvpQueueJoin, TypePvpQueueLeave, TypePvpReady, TypePvpDecline, TypePvpLeave,
		TypePvpEmote, TypePvpSpectate, TypePvpReport, TypePvpLeaderboard, TypePvpHistory, TypePvpShopBuy,
		TypePvpTraining, TypePvpDuel, TypePvpDuelAccept, TypePvpDuelDecline, TypeGetReplay,
		TypeSetRating, TypeSetRank, TypePvpWin, TypeSetDamage:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyPvp(p.ID, msg.env) {
			h.routeDungeon(p.ID, ev)
		}
	case TypeGetEndgame, TypeClaimDaily, TypeClaimWeekly, TypeClaimChallenge, TypeClaimSeason, TypeGetSeason,
		TypeSetSeasonXP, TypeGetHorizon, TypeGetCalendar, TypeGetLiveEvent, TypeContributeEvent,
		TypeUnlockAchievement, TypeUnlockCosmetic, TypeSetAchievement, TypeSetShowcase, TypeGetPublicProfile,
		TypeGetLearning, TypeGetLoreBook, TypeGetLeaderboards:
		if live == nil {
			return
		}
		for _, ev := range h.World.ApplyEndgame(p.ID, msg.env) {
			h.World.SendTo(p.ID, ev)
		}
	case TypePing:
		var ping PingIn
		_ = json.Unmarshal(msg.env.Data, &ping)
		now := time.Now()
		if live != nil {
			live.LastHeard = now
			if ping.T > 0 {
				delta := now.UnixMilli() - ping.T
				if delta < 0 {
					delta = -delta
				}
				live.PingMs = int(delta)
			}
		} else {
			p.LastHeard = now
		}
		select {
		case p.send <- marshal(TypePong, PongOut{T: ping.T, St: time.Now().UnixMilli()}):
		default:
		}
	default:
		return
	}
}

func (h *Hub) flushPartyQuest(id string) {
	w := h.World
	w.mu.Lock()
	p := w.players[id]
	var ids []string
	if p != nil && w.Parties != nil {
		if pt := w.Parties.Of(id); pt != nil {
			ids = append([]string{}, pt.Members...)
		}
	}
	w.mu.Unlock()
	for _, mid := range ids {
		if mid == id {
			continue
		}
		if payload := w.FlushQuest(mid); payload != nil {
			w.SendTo(mid, payload)
		}
	}
}
func (h *Hub) Join(p *Player) { h.join <- p }
func (h *Hub) Leave(p *Player) {
	select {
	case h.leave <- p:
	default:
		go func() { h.leave <- p }()
	}
}
func (h *Hub) Inbound(p *Player, env Envelope) {
	select {
	case h.in <- inbound{player: p, env: env}:
	default:
	}
}

func worldBroadcast(payload []byte) bool {
	var env Envelope
	if json.Unmarshal(payload, &env) != nil {
		return false
	}
	switch env.Type {
	case TypePlayerLevelUp, TypeAttackResult, TypeDamageResult, TypePlayerDeath, TypeEnemyDeath,
		TypeTransformationStarted, TypeTransformationUpdated, TypeTransformationEnded, TypeSkillUsed,
		TypeMountUpdated, TypeWorldEvent, TypeGuardianDefeated, TypeWeatherUpdated, TypeChatMessage,
		TypeWorldBossAnnounce, TypeWorldBossState, TypeChapterComplete, TypeEmotePlayed, TypeMountEmote, TypeEnemySpawn:
		return true
	default:
		return false
	}
}

func (h *Hub) routeSocial(fromID string, ev []byte) {
	var env Envelope
	if json.Unmarshal(ev, &env) != nil {
		h.World.SendTo(fromID, ev)
		return
	}
	switch env.Type {
	case TypePartyInvite, TypeFriendRequest, TypeGuildInvite, TypeTradeRequest:
		var body struct {
			ToID string `json:"toId"`
		}
		_ = json.Unmarshal(env.Data, &body)
		if body.ToID != "" {
			h.World.SendTo(body.ToID, ev)
			return
		}
	case TypeSocialNotification:
		var n SocialNote
		_ = json.Unmarshal(env.Data, &n)
		if n.ToID != "" {
			h.World.SendTo(n.ToID, ev)
			return
		}
	case TypePartyUpdated, TypePartyMemberJoined, TypePartyMemberLeft, TypeChatMessage, TypeGuildUpdated, TypeTradeUpdated, TypeHouseState, TypeMarketListings, TypeEmotePlayed:
		var view PartyView
		_ = json.Unmarshal(env.Data, &view)
		seen := map[string]bool{}
		for _, id := range view.NotifyIDs {
			h.World.SendTo(id, ev)
			seen[id] = true
		}
		for _, m := range view.Members {
			if seen[m.PlayerID] {
				continue
			}
			h.World.SendTo(m.PlayerID, ev)
			seen[m.PlayerID] = true
		}
		if !seen[fromID] {
			h.World.SendTo(fromID, ev)
		}
		return
	case TypeFriendUpdated:
		var body struct {
			ToID string `json:"toId"`
		}
		_ = json.Unmarshal(env.Data, &body)
		if body.ToID != "" {
			h.World.SendTo(body.ToID, ev)
			return
		}
	}
	h.World.SendTo(fromID, ev)
}

func (h *Hub) routeDungeon(fromID string, ev []byte) {
	var env Envelope
	if json.Unmarshal(ev, &env) != nil {
		h.World.SendTo(fromID, ev)
		return
	}
	var body struct {
		ToID       string `json:"toId"`
		InstanceID string `json:"instanceId"`
	}
	_ = json.Unmarshal(env.Data, &body)
	if body.ToID != "" {
		h.World.SendTo(body.ToID, ev)
		return
	}
	if env.Type == TypeDungeonReadyCheck {
		var ready DungeonReadyOut
		_ = json.Unmarshal(env.Data, &ready)
		seen := map[string]bool{}
		for _, m := range ready.Members {
			h.World.SendTo(m.PlayerID, ev)
			seen[m.PlayerID] = true
		}
		if !seen[fromID] {
			h.World.SendTo(fromID, ev)
		}
		return
	}
	if body.InstanceID != "" {
		h.World.BroadcastInstance(body.InstanceID, ev)
		h.World.SendTo(fromID, ev)
		return
	}
	h.World.BroadcastScope(fromID, ev)
}

func (w *WorldState) BroadcastSnapshots() {
	w.mu.Lock()
	ids := make([]string, 0, len(w.players))
	for _, p := range w.players {
		if p.Connected {
			ids = append(ids, p.ID)
		}
	}
	w.mu.Unlock()
	for _, id := range ids {
		w.SendTo(id, marshal(TypeWorldSnapshot, w.SnapshotFor(id)))
	}
}

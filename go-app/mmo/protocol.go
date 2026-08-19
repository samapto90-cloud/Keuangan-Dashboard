package mmo

import "encoding/json"

const (
	TypeAuth                   = "AUTH"
	TypeAuthOk                 = "AUTH_OK"
	TypeAuthFail               = "AUTH_FAIL"
	TypeJoinWorld              = "JOIN_WORLD"
	TypeWelcome                = "WELCOME"
	TypePlayerSpawn            = "PLAYER_SPAWN"
	TypePlayerDespawn          = "PLAYER_DESPAWN"
	TypePlayerJoin             = "PLAYER_JOIN"
	TypePlayerLeave            = "PLAYER_LEAVE"
	TypeMoveInput              = "MOVE_INPUT"
	TypeWorldSnapshot          = "WORLD_SNAPSHOT"
	TypeError                  = "ERROR"
	TypePing                   = "PING"
	TypePong                   = "PONG"
	TypePlayerAttack           = "PLAYER_ATTACK"
	TypePlayerSkill            = "PLAYER_SKILL"
	TypePlayerDodge            = "PLAYER_DODGE"
	TypePlayerCombo            = "PLAYER_COMBO"
	TypePlayerRespawn          = "PLAYER_RESPAWN"
	TypeAttackResult           = "ATTACK_RESULT"
	TypeDamageResult           = "DAMAGE_RESULT"
	TypePlayerHit              = "PLAYER_HIT"
	TypeEnemyHit               = "ENEMY_HIT"
	TypeEnemyDeath             = "ENEMY_DEATH"
	TypeEnemySpawn             = "ENEMY_SPAWN"
	TypePlayerDeath            = "PLAYER_DEATH"
	TypePlayerLevelUp          = "PLAYER_LEVEL_UP"
	TypeActionReject           = "ACTION_REJECT"
	TypeInteract               = "INTERACT"
	TypeInteractResult         = "INTERACT_RESULT"
	TypeManualSave             = "MANUAL_SAVE"
	TypeSaveOk                 = "SAVE_OK"
	TypeQuestAccept            = "QUEST_ACCEPT"
	TypeQuestDecline           = "QUEST_DECLINE"
	TypeQuestClaim             = "QUEST_CLAIM"
	TypeQuestUpdated           = "QUEST_UPDATED"
	TypeQuestReward            = "QUEST_REWARD"
	TypeQuestComplete          = "QUEST_COMPLETE"
	TypeCollectItem            = "COLLECT_ITEM"
	TypeForestUnlock           = "FOREST_UNLOCK"
	TypeEducationAnswer        = "EDUCATION_ANSWER"
	TypeEducationQuestion      = "EDUCATION_QUESTION"
	TypeEducationFeedback      = "EDUCATION_FEEDBACK"
	TypeEducationCorrect       = "EDUCATION_CORRECT"
	TypeHeal                   = "HEAL"
	TypeShopOpen               = "SHOP_OPEN"
	TypeFastTravel             = "FAST_TRAVEL"
	TypePickupItem             = "PICKUP_ITEM"
	TypeUseItem                = "USE_ITEM"
	TypeEquipItem              = "EQUIP_ITEM"
	TypeUnequipItem            = "UNEQUIP_ITEM"
	TypeDiscardItem            = "DISCARD_ITEM"
	TypeGetInventory           = "GET_INVENTORY"
	TypeInventoryUpdated       = "INVENTORY_UPDATED"
	TypeItemAdded              = "ITEM_ADDED"
	TypeItemRemoved            = "ITEM_REMOVED"
	TypeItemUsed               = "ITEM_USED"
	TypeItemConsumed           = "ITEM_CONSUMED"
	TypeEquipmentUpdated       = "EQUIPMENT_UPDATED"
	TypePlayerStatsUpdated     = "PLAYER_STATS_UPDATED"
	TypeGiveItem               = "GIVE_ITEM"
	TypeGiveCurrency           = "GIVE_CURRENCY"
	TypeAddItem                = "ADD_ITEM"
	TypeSetQuantity            = "SET_QUANTITY"
	TypeSplitStack             = "SPLIT_STACK"
	TypeUpgradeItem            = "UPGRADE_ITEM"
	TypeEnchantItem            = "ENCHANT_ITEM"
	TypeExpandBag              = "EXPAND_BAG"
	TypeSaveGearLoadout        = "SAVE_GEAR_LOADOUT"
	TypeLoadGearLoadout        = "LOAD_GEAR_LOADOUT"
	TypeClaimTempLoot          = "CLAIM_TEMP_LOOT"
	TypeSalvageItems           = "SALVAGE_ITEMS"
	TypeToggleCosmetic         = "TOGGLE_COSMETIC"
	TypeTraceItem              = "TRACE_ITEM"
	TypeSetItemStats           = "SET_ITEM_STATS"
	TypeSetItemLevel           = "SET_ITEM_LEVEL"
	TypeDuplicateInstance      = "DUPLICATE_ITEM"
	TypeSetInstanceID          = "SET_INSTANCE"
	TypePartyInvite            = "PARTY_INVITE"
	TypePartyAccept            = "PARTY_ACCEPT"
	TypePartyDecline           = "PARTY_DECLINE"
	TypePartyLeave             = "PARTY_LEAVE"
	TypePartyKick              = "PARTY_KICK"
	TypePartyDisband           = "PARTY_DISBAND"
	TypePartySetTarget         = "PARTY_SET_TARGET"
	TypePartyUpdated           = "PARTY_UPDATED"
	TypePartyMemberJoined      = "PARTY_MEMBER_JOINED"
	TypePartyMemberLeft        = "PARTY_MEMBER_LEFT"
	TypeFriendRequest          = "FRIEND_REQUEST"
	TypeAcceptFriend           = "ACCEPT_FRIEND"
	TypeDeclineFriend          = "DECLINE_FRIEND"
	TypeRemoveFriend           = "REMOVE_FRIEND"
	TypeBlockPlayer            = "BLOCK_PLAYER"
	TypeUnblockPlayer          = "UNBLOCK_PLAYER"
	TypeInspectPlayer          = "INSPECT_PLAYER"
	TypeInspectResult          = "INSPECT_RESULT"
	TypeGetSocial              = "GET_SOCIAL"
	TypeFriendUpdated          = "FRIEND_UPDATED"
	TypeSocialNotification     = "SOCIAL_NOTIFICATION"
	TypeChat                   = "CHAT"
	TypeNearbyPlayers          = "NEARBY_PLAYERS"
	TypeDungeonEnter           = "DUNGEON_ENTER"
	TypeDungeonReady           = "DUNGEON_READY"
	TypeDungeonLeave           = "DUNGEON_LEAVE"
	TypeDungeonAbandon         = "DUNGEON_ABANDON"
	TypeDungeonRetry           = "DUNGEON_RETRY"
	TypeDungeonOffer           = "DUNGEON_OFFER"
	TypeDungeonReadyCheck      = "DUNGEON_READY_CHECK"
	TypeDungeonLoading         = "DUNGEON_LOADING"
	TypeDungeonStarted         = "DUNGEON_STARTED"
	TypeDungeonState           = "DUNGEON_STATE"
	TypeDungeonWave            = "DUNGEON_WAVE"
	TypeDungeonObjective       = "DUNGEON_OBJECTIVE"
	TypeDungeonComplete        = "DUNGEON_COMPLETE"
	TypeDungeonFailed          = "DUNGEON_FAILED"
	TypeDungeonLeft            = "DUNGEON_LEFT"
	TypeGetDungeons            = "GET_DUNGEONS"
	TypeDungeonList            = "DUNGEON_LIST"
	TypeQueueJoin              = "QUEUE_JOIN"
	TypeQueueLeave             = "QUEUE_LEAVE"
	TypeQueueUpdate            = "QUEUE_UPDATE"
	TypeDungeonJoin            = "DUNGEON_JOIN"
	TypeDungeonFill            = "DUNGEON_FILL"
	TypeDungeonRevive          = "DUNGEON_REVIVE"
	TypeDungeonVote            = "DUNGEON_VOTE"
	TypeDungeonVoteUpdate      = "DUNGEON_VOTE_UPDATE"
	TypeDungeonTaunt           = "DUNGEON_TAUNT"
	TypeDungeonWipe            = "DUNGEON_WIPE"
	TypePlayerDowned           = "PLAYER_DOWNED"
	TypePlayerRevived          = "PLAYER_REVIVED"
	TypeBossInterrupt          = "BOSS_INTERRUPT"
	TypeBossLock               = "BOSS_LOCK"
	TypeSkipDungeonIntro       = "SKIP_DUNGEON_INTRO"
	TypeGetPvp                 = "GET_PVP"
	TypePvpLobby               = "PVP_LOBBY"
	TypePvpQueueJoin           = "PVP_QUEUE_JOIN"
	TypePvpQueueLeave          = "PVP_QUEUE_LEAVE"
	TypePvpQueueUpdate         = "PVP_QUEUE_UPDATE"
	TypePvpReady               = "PVP_READY"
	TypePvpDecline             = "PVP_DECLINE"
	TypePvpReadyCheck          = "PVP_READY_CHECK"
	TypePvpLoading             = "PVP_LOADING"
	TypePvpCountdown           = "PVP_COUNTDOWN"
	TypePvpState               = "PVP_STATE"
	TypePvpKillFeed            = "PVP_KILL_FEED"
	TypePvpCapture             = "PVP_CAPTURE"
	TypePvpResult              = "PVP_RESULT"
	TypePvpLeaderboard         = "PVP_LEADERBOARD"
	TypePvpHistory             = "PVP_HISTORY"
	TypePvpEmote               = "PVP_EMOTE"
	TypePvpSpectate            = "PVP_SPECTATE"
	TypePvpReport              = "PVP_REPORT"
	TypePvpLeave               = "PVP_LEAVE"
	TypePvpLeft                = "PVP_LEFT"
	TypePvpAfk                 = "PVP_AFK"
	TypePvpShopBuy             = "PVP_SHOP_BUY"
	TypePvpTraining            = "PVP_TRAINING"
	TypePvpDuel                = "PVP_DUEL"
	TypePvpDuelAccept          = "PVP_DUEL_ACCEPT"
	TypePvpDuelDecline         = "PVP_DUEL_DECLINE"
	TypePvpDuelRequest         = "PVP_DUEL_REQUEST"
	TypeGetReplay              = "GET_REPLAY"
	TypePvpReplay              = "PVP_REPLAY"
	TypeSetRating              = "SET_RATING"
	TypeSetRank                = "SET_RANK"
	TypePvpWin                 = "PVP_WIN"
	TypeSetDamage              = "SET_DAMAGE"
	TypeSetEnergy              = "SET_ENERGY"
	TypeSetCooldown            = "SET_COOLDOWN"
	TypeUnlockTransform        = "UNLOCK_TRANSFORM"
	TypePlayerCharge           = "PLAYER_CHARGE"
	TypeBossSpawn              = "BOSS_SPAWN"
	TypeBossPhase              = "BOSS_PHASE"
	TypeBossTelegraph          = "BOSS_TELEGRAPH"
	TypeBossAOE                = "BOSS_AOE"
	TypeBossEnrage             = "BOSS_ENRAGE"
	TypeBossDefeated           = "BOSS_DEFEATED"
	TypeBossReset              = "BOSS_RESET"
	TypeLootResult             = "LOOT_RESULT"
	TypeClaimLoot              = "CLAIM_LOOT"
	TypeGetChapters            = "GET_CHAPTERS"
	TypeChapterList            = "CHAPTER_LIST"
	TypeUnlockChapter          = "UNLOCK_CHAPTER"
	TypeSetBossHP              = "SET_BOSS_HP"
	TypeSetChapter             = "SET_CHAPTER"
	TypeSetObjective           = "SET_OBJECTIVE"
	TypeGiveLoot               = "GIVE_LOOT"
	TypeObjectiveComplete      = "OBJECTIVE_COMPLETE"
	TypeSetBossDead            = "SET_BOSS_DEAD"
	TypeSpawnBoss              = "SPAWN_BOSS"
	TypeCompleteDungeon        = "COMPLETE_DUNGEON"
	TypeSkipMechanic           = "SKIP_MECHANIC"
	TypeDamageBoss             = "DAMAGE_BOSS"
	TypeGetOpenWorld           = "GET_OPEN_WORLD"
	TypeOpenWorld              = "OPEN_WORLD"
	TypeGetDungeonHistory      = "GET_DUNGEON_HISTORY"
	TypeRaidExchange           = "RAID_EXCHANGE"
	TypeAllocateAttribute      = "ALLOCATE_ATTRIBUTE"
	TypeResetAttributes        = "RESET_ATTRIBUTES"
	TypeUnlockSkill            = "UNLOCK_SKILL"
	TypeGetProgression         = "GET_PROGRESSION"
	TypeRequestTransformation  = "REQUEST_TRANSFORMATION"
	TypeSetTransformation      = "SET_TRANSFORMATION"
	TypeSetLevel               = "SET_LEVEL"
	TypeSetSkillPoints         = "SET_SKILL_POINTS"
	TypeProgressionState       = "PROGRESSION_STATE"
	TypeSkillUnlocked          = "SKILL_UNLOCKED"
	TypeSkillUsed              = "SKILL_USED"
	TypeTransformationStarted  = "TRANSFORMATION_STARTED"
	TypeTransformationUpdated  = "TRANSFORMATION_UPDATED"
	TypeTransformationEnded    = "TRANSFORMATION_ENDED"
	TypeTransformationRejected = "TRANSFORMATION_REJECTED"
	TypePowerRatingUpdated     = "POWER_RATING_UPDATED"
	TypeGetWorldJournal        = "GET_WORLD_JOURNAL"
	TypeWorldJournal           = "WORLD_JOURNAL"
	TypeZoneDiscovered         = "ZONE_DISCOVERED"
	TypeLandmarkDiscovered     = "LANDMARK_DISCOVERED"
	TypeLoreDiscovered         = "LORE_DISCOVERED"
	TypeRequestMount           = "REQUEST_MOUNT"
	TypeDismount               = "DISMOUNT"
	TypeMountUpdated           = "MOUNT_UPDATED"
	TypeWorldEvent             = "WORLD_EVENT"
	TypeJoinWorldEvent         = "JOIN_WORLD_EVENT"
	TypeClaimEventReward       = "CLAIM_EVENT_REWARD"
	TypeEventReward            = "EVENT_REWARD"
	TypeGuardianDefeated       = "GUARDIAN_DEFEATED"
	TypeUnlockGuardian         = "UNLOCK_GUARDIAN"
	TypeUnlockRegion           = "UNLOCK_REGION"
	TypeSetWeather             = "SET_WEATHER"
	TypeTeleport               = "TELEPORT"
	TypeSkipCinematic          = "SKIP_CINEMATIC"
	TypeCinematicSkipped       = "CINEMATIC_SKIPPED"
	TypeFastTravelOk           = "FAST_TRAVEL_OK"
	TypeWeatherUpdated         = "WEATHER_UPDATED"
	TypeSwitchChannel          = "SWITCH_CHANNEL"
	TypeChannelList            = "CHANNEL_LIST"
	TypeWorldBossState         = "WORLD_BOSS_STATE"
	TypeWorldBossAnnounce      = "WORLD_BOSS_ANNOUNCE"
	TypeTriggerWorldBoss       = "TRIGGER_WORLD_BOSS"
	TypeClaimWorldBoss         = "CLAIM_WORLD_BOSS"
	TypeGetCollections         = "GET_COLLECTIONS"
	TypeCollectionBook         = "COLLECTION_BOOK"
	TypeChapterComplete        = "CHAPTER_COMPLETE"
	TypeEnterMasjid            = "ENTER_MASJID"
	TypeSetWorldBossHP         = "SET_WORLD_BOSS_HP"
	TypeSetWorldTime           = "SET_WORLD_TIME"
	TypeSpawnTreasure          = "SPAWN_TREASURE"
	TypeStartWorldEvent        = "START_WORLD_EVENT"
	TypeRandomEncounter        = "RANDOM_ENCOUNTER"
	TypeSetLanguage            = "SET_LANGUAGE"
	TypeGetStory               = "GET_STORY"
	TypeStoryState             = "STORY_STATE"
	TypeStoryChoice            = "STORY_CHOICE"
	TypeClaimStoryChapter      = "CLAIM_STORY_CHAPTER"
	TypeReplayCinematic        = "REPLAY_CINEMATIC"
	TypeReplayChapter          = "REPLAY_CHAPTER"
	TypeCinematicStart         = "CINEMATIC_START"
	TypeCinematicDone          = "CINEMATIC_DONE"
	TypeStartNGPlus            = "START_NG_PLUS"
	TypeSetStoryFlag           = "SET_STORY_FLAG"
	TypeUnlockStoryChapter     = "UNLOCK_STORY_CHAPTER"
	TypeDefeatSiluman          = "DEFEAT_SILUMAN"
	TypeClaimStoryReward       = "CLAIM_STORY_REWARD"
	TypeGetAdventure           = "GET_ADVENTURE"
	TypeSetRelationship        = "SET_RELATIONSHIP"
	TypeAddRelationship        = "ADD_RELATIONSHIP"
	TypeSetNpcMemory           = "SET_NPC_MEMORY"
	TypeUnlockLore             = "UNLOCK_LORE"
	TypeCompleteQuest          = "COMPLETE_QUEST"
	TypeSetWorldState          = "SET_WORLD_STATE"
	TypeClaimEducationReward   = "CLAIM_EDUCATION_REWARD"
	TypeGetMounts              = "GET_MOUNTS"
	TypeMountCollection        = "MOUNT_COLLECTION"
	TypeFavoriteMount          = "FAVORITE_MOUNT"
	TypeEquipMount             = "EQUIP_MOUNT"
	TypeSetMountCosmetic       = "SET_MOUNT_COSMETIC"
	TypeMountEmote             = "MOUNT_EMOTE"
	TypeUnlockMount            = "UNLOCK_MOUNT"
	TypeGrantMount             = "GRANT_MOUNT"
	TypeClaimMount             = "CLAIM_MOUNT"
	TypeSetMountSpeed          = "SET_MOUNT_SPEED"
	TypeRaceStart              = "RACE_START"
	TypeRaceCheckpoint         = "RACE_CHECKPOINT"
	TypeRaceFinish             = "RACE_FINISH"
	TypeRaceUpdated            = "RACE_UPDATED"
	TypeTravelEvent            = "TRAVEL_EVENT"
	TypeInspectLandmark        = "INSPECT_LANDMARK"
	TypePartySetWaypoint       = "PARTY_SET_WAYPOINT"
	TypeFollowParty            = "FOLLOW_PARTY"
	TypeSetCombatStyle         = "SET_COMBAT_STYLE"
	TypeSaveBuild              = "SAVE_BUILD"
	TypeLoadBuild              = "LOAD_BUILD"
	TypeSwitchBuild            = "SWITCH_BUILD"
	TypeGetBuilds              = "GET_BUILDS"
	TypeBuildList              = "BUILD_LIST"
	TypeBuildSaved             = "BUILD_SAVED"
	TypeBuildLoaded            = "BUILD_LOADED"
	TypeSetLoadout             = "SET_LOADOUT"
	TypeResetSkills            = "RESET_SKILLS"
	TypePlayerBlock            = "PLAYER_BLOCK"
	TypePlayerCounter          = "PLAYER_COUNTER"
	TypeTrainingMeter          = "TRAINING_METER"
	TypeTravelSuggestion       = "TRAVEL_SUGGESTION"
	TypeGetEndgame             = "GET_ENDGAME"
	TypeEndgameState           = "ENDGAME_STATE"
	TypeClaimDaily             = "CLAIM_DAILY"
	TypeClaimWeekly            = "CLAIM_WEEKLY"
	TypeClaimChallenge         = "CLAIM_CHALLENGE"
	TypeClaimSeason            = "CLAIM_SEASON"
	TypeGetSeason              = "GET_SEASON"
	TypeSetSeasonXP            = "SET_SEASON_XP"
	TypeGetHorizon             = "GET_HORIZON"
	TypeGetCalendar            = "GET_CALENDAR"
	TypeGetLiveEvent           = "GET_LIVE_EVENT"
	TypeContributeEvent        = "CONTRIBUTE_EVENT"
	TypeUnlockAchievement      = "UNLOCK_ACHIEVEMENT"
	TypeUnlockCosmetic         = "UNLOCK_COSMETIC"
	TypeSetAchievement         = "SET_ACHIEVEMENT"
	TypeSetShowcase            = "SET_SHOWCASE"
	TypeGetPublicProfile       = "GET_PUBLIC_PROFILE"
	TypeGetLearning            = "GET_LEARNING"
	TypeGetLoreBook            = "GET_LORE_BOOK"
	TypeGetLeaderboards        = "GET_LEADERBOARDS"
	TypeAchievementUnlocked    = "ACHIEVEMENT_UNLOCKED"
	TypeChatMessage            = "CHAT_MESSAGE"
	TypeGuildCreate            = "GUILD_CREATE"
	TypeGuildInvite            = "GUILD_INVITE"
	TypeGuildAccept            = "GUILD_ACCEPT"
	TypeGuildDecline           = "GUILD_DECLINE"
	TypeGuildLeave             = "GUILD_LEAVE"
	TypeGuildKick              = "GUILD_KICK"
	TypeGuildDisband           = "GUILD_DISBAND"
	TypeGuildTransfer          = "GUILD_TRANSFER"
	TypeGuildAnnounce          = "GUILD_ANNOUNCE"
	TypeGuildUpdated           = "GUILD_UPDATED"
	TypeSetGuildRank           = "SET_GUILD_RANK"
	TypeGetGuild               = "GET_GUILD"
	TypeTradeRequest           = "TRADE_REQUEST"
	TypeTradeAccept            = "TRADE_ACCEPT"
	TypeTradeDecline           = "TRADE_DECLINE"
	TypeTradeOffer             = "TRADE_OFFER"
	TypeTradeReady             = "TRADE_READY"
	TypeTradeConfirm           = "TRADE_CONFIRM"
	TypeTradeCancel            = "TRADE_CANCEL"
	TypeTradeUpdated           = "TRADE_UPDATED"
	TypeShopBuy                = "SHOP_BUY"
	TypeShopSell               = "SHOP_SELL"
	TypeShopCatalog            = "SHOP_CATALOG"
	TypeSetCoin                = "SET_COIN"
	TypeSetTitle               = "SET_TITLE"
	TypeSetCosmetic            = "SET_COSMETIC"
	TypeReportPlayer           = "REPORT_PLAYER"
	TypeSearchPlayer           = "SEARCH_PLAYER"
	TypeSearchResult           = "SEARCH_RESULT"
	TypePartyFinderList        = "PARTY_FINDER_LIST"
	TypePartyFinderCreate      = "PARTY_FINDER_CREATE"
	TypePartyFinderJoin        = "PARTY_FINDER_JOIN"
	TypePartyTransfer          = "PARTY_TRANSFER"
	TypePartySetRole           = "PARTY_SET_ROLE"
	TypePartyReady             = "PARTY_READY"
	TypeNotifyList             = "NOTIFY_LIST"
	TypeGetNotifies            = "GET_NOTIFIES"
	TypeMutePlayer             = "MUTE_PLAYER"
	TypeModKick                = "MOD_KICK"
	TypeModBan                 = "MOD_BAN"
	TypeSetPrivacy             = "SET_PRIVACY"
	TypeGetPrivacy             = "GET_PRIVACY"
	TypePrivacyUpdated         = "PRIVACY_UPDATED"
	TypeSetPresence            = "SET_PRESENCE"
	TypeLocalMute              = "LOCAL_MUTE"
	TypeUnmuteLocal            = "UNMUTE_LOCAL"
	TypeReportMessage          = "REPORT_MESSAGE"
	TypePartyCreate            = "PARTY_CREATE"
	TypeSetFriendship          = "SET_FRIENDSHIP"
	TypeSetPartyLeader         = "SET_PARTY_LEADER"
	TypeSetTradeItem           = "SET_TRADE_ITEM"
	TypeSetChatSender          = "SET_CHAT_SENDER"
	TypeSetGuildMaster         = "SET_GUILD_MASTER"
	TypeGetMarket              = "GET_MARKET"
	TypeMarketSearch           = "MARKET_SEARCH"
	TypeMarketList             = "MARKET_LIST"
	TypeMarketBuy              = "MARKET_BUY"
	TypeMarketCancel           = "MARKET_CANCEL"
	TypeMarketListings         = "MARKET_LISTINGS"
	TypeBankDeposit            = "BANK_DEPOSIT"
	TypeBankWithdraw           = "BANK_WITHDRAW"
	TypeGetBank                = "GET_BANK"
	TypeBankUpdated            = "BANK_UPDATED"
	TypeLockItem               = "LOCK_ITEM"
	TypeFavoriteItem           = "FAVORITE_ITEM"
	TypeHouseEnter             = "HOUSE_ENTER"
	TypeHouseLeave             = "HOUSE_LEAVE"
	TypeHouseVisit             = "HOUSE_VISIT"
	TypeHousePlace             = "HOUSE_PLACE"
	TypeHouseRemove            = "HOUSE_REMOVE"
	TypeSetHouseAccess         = "SET_HOUSE_ACCESS"
	TypeGetHouse               = "GET_HOUSE"
	TypeHouseState             = "HOUSE_STATE"
	TypeGuildApply             = "GUILD_APPLY"
	TypeGuildReview            = "GUILD_REVIEW"
	TypeGuildDeposit           = "GUILD_DEPOSIT"
	TypeGuildWithdraw          = "GUILD_WITHDRAW"
	TypeGetGuildLog            = "GET_GUILD_LOG"
	TypeGuildLog               = "GUILD_LOG"
	TypeGuildSetEmblem         = "GUILD_SET_EMBLEM"
	TypeGuildSetDesc           = "GUILD_SET_DESC"
	TypeSocialEmote            = "SOCIAL_EMOTE"
	TypeEmotePlayed            = "EMOTE_PLAYED"
	TypeGetPlayerCard          = "GET_PLAYER_CARD"
	TypePlayerCard             = "PLAYER_CARD"
	TypeSetName                = "SET_NAME"
	TypeSetGuildXP             = "SET_GUILD_XP"
	TypeSetOwned               = "SET_OWNED"
	TypeGetEconomy             = "GET_ECONOMY"
	TypeGather                 = "GATHER"
	TypeGatherResult           = "GATHER_RESULT"
	TypeCraft                  = "CRAFT"
	TypeCraftResult            = "CRAFT_RESULT"
	TypeGetCrafting            = "GET_CRAFTING"
	TypeCraftingState          = "CRAFTING_STATE"
	TypeSetProfession          = "SET_PROFESSION"
	TypeResetProfession        = "RESET_PROFESSION"
	TypeFishStart              = "FISH_START"
	TypeFishCatch              = "FISH_CATCH"
	TypeFishState              = "FISH_STATE"
	TypeNpcShopOpen            = "NPC_SHOP_OPEN"
	TypeNpcShopBuy             = "NPC_SHOP_BUY"
	TypeNpcShopSell            = "NPC_SHOP_SELL"
	TypeNpcRepair              = "NPC_REPAIR"
	TypeStallOpen              = "STALL_OPEN"
	TypeStallList              = "STALL_LIST"
	TypeStallBuy               = "STALL_BUY"
	TypeStallClose             = "STALL_CLOSE"
	TypeCraftOrder             = "CRAFT_ORDER"
	TypeCraftOrderAccept       = "CRAFT_ORDER_ACCEPT"
	TypeGetWorkshop            = "GET_WORKSHOP"
	TypeAddGold                = "ADD_GOLD"
	TypeAddMaterial            = "ADD_MATERIAL"
	TypeSetPrice               = "SET_PRICE"
	TypeSetGold                = "SET_GOLD"
	TypeGiveGold               = "GIVE_GOLD"
	TypeCreateRecipe           = "CREATE_RECIPE"
	TypeGuildContribute        = "GUILD_CONTRIBUTE"
	TypeGetGoldLog             = "GET_GOLD_LOG"
	TypeCreateHouse            = "CREATE_HOUSE"
	TypeHouseLock              = "HOUSE_LOCK"
	TypeHouseRename            = "HOUSE_RENAME"
	TypeHouseStyle             = "HOUSE_STYLE"
	TypeHouseMove              = "HOUSE_MOVE"
	TypeHouseDecorate          = "HOUSE_DECORATE"
	TypeHouseStore             = "HOUSE_STORE"
	TypeHouseTake              = "HOUSE_TAKE"
	TypeGardenPlant            = "GARDEN_PLANT"
	TypeGardenWater            = "GARDEN_WATER"
	TypeGardenHarvest          = "GARDEN_HARVEST"
	TypePetClaim               = "PET_CLAIM"
	TypePetSummon              = "PET_SUMMON"
	TypePetDismiss             = "PET_DISMISS"
	TypePetCare                = "PET_CARE"
	TypePetName                = "PET_NAME"
	TypeGetPets                = "GET_PETS"
	TypeAddPet                 = "ADD_PET"
	TypeGuildHallEnter         = "GUILD_HALL_ENTER"
	TypeGuildHallLeave         = "GUILD_HALL_LEAVE"
	TypeGuildHost              = "GUILD_HOST"
	TypeGetLife                = "GET_LIFE"
	TypeClaimDailyLife         = "CLAIM_DAILY_LIFE"
	TypeHouseVote              = "HOUSE_VOTE"
	TypeLifeQuiz               = "LIFE_QUIZ"
	TypeClaimCollection        = "CLAIM_COLLECTION"
	TypeLifeState              = "LIFE_STATE"
	TypePetState               = "PET_STATE"
)

type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type AuthIn struct {
	Name     string `json:"name"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthOkOut struct {
	Token     string `json:"token"`
	PlayerID  string `json:"playerId"`
	SessionID string `json:"sessionId"`
}

type JoinWorldIn struct {
	WorldID string `json:"worldId"`
}

type MoveInput struct {
	Seq    uint32  `json:"seq"`
	AX     float64 `json:"ax"`
	AZ     float64 `json:"az"`
	Yaw    float64 `json:"yaw"`
	Sprint bool    `json:"sprint"`
	Jump   bool    `json:"jump"`
}

type PingIn struct {
	T int64 `json:"t"`
}

type PongOut struct {
	T  int64 `json:"t"`
	St int64 `json:"st"`
}

type PlayerSnapshot struct {
	PlayerID    string  `json:"id"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           float64 `json:"z"`
	Yaw         float64 `json:"yaw"`
	VX          float64 `json:"vx"`
	VY          float64 `json:"vy"`
	VZ          float64 `json:"vz"`
	State       string  `json:"st"`
	CombatState string  `json:"cs"`
	HP          int     `json:"hp"`
	MaxHP       int     `json:"maxHp"`
	Energy      int     `json:"energy"`
	MaxEnergy   int     `json:"maxEnergy"`
	Stamina     int     `json:"stamina"`
	Level       int     `json:"level"`
	Exp         int     `json:"exp"`
	ExpToNext   int     `json:"expToNext"`
	Seq         uint32  `json:"seq"`
	FormID      string  `json:"formId,omitempty"`
	Transform   string  `json:"transform,omitempty"`
	MountID     string  `json:"mountId,omitempty"`
	Mounted     bool    `json:"mounted,omitempty"`
	MountState  string  `json:"mountState,omitempty"`
	Swimming    bool    `json:"swimming,omitempty"`
	ZoneID      string  `json:"zoneId,omitempty"`
	GuildTag    string  `json:"guildTag,omitempty"`
	Title       string  `json:"title,omitempty"`
	PetID       string  `json:"petId,omitempty"`
}

type PlayerSpawn struct {
	PlayerID    string  `json:"playerId"`
	Name        string  `json:"name"`
	Level       int     `json:"level"`
	Class       string  `json:"class"`
	HP          int     `json:"hp"`
	MaxHP       int     `json:"maxHp"`
	Energy      int     `json:"energy"`
	MaxEnergy   int     `json:"maxEnergy"`
	Exp         int     `json:"exp"`
	ExpToNext   int     `json:"expToNext"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           float64 `json:"z"`
	Yaw         float64 `json:"yaw"`
	State       string  `json:"state"`
	CombatState string  `json:"combatState"`
	FormID      string  `json:"formId,omitempty"`
	GuildTag    string  `json:"guildTag,omitempty"`
	Title       string  `json:"title,omitempty"`
}

type WelcomeOut struct {
	WorldID     string            `json:"worldId"`
	Channel     string            `json:"channel"`
	TickRate    int               `json:"tickRate"`
	PlayerID    string            `json:"playerId"`
	SessionID   string            `json:"sessionId"`
	Self        PlayerSpawn       `json:"self"`
	Players     []PlayerSpawn     `json:"players"`
	Snapshot    WorldSnapshot     `json:"snapshot"`
	Progress    PlayerProgressOut `json:"progress"`
	Loadout     InventoryUpdated  `json:"loadout"`
	Catalog     []ItemDefView     `json:"catalog"`
	Social      SocialState       `json:"social"`
	Progression ProgressionView   `json:"progression"`
}

type WorldSnapshot struct {
	WorldID    string           `json:"worldId"`
	Channel    string           `json:"channel"`
	T          int64            `json:"t"`
	Online     int              `json:"online"`
	Players    []PlayerSnapshot `json:"players"`
	NPCs       []NPCSnapshot    `json:"npcs"`
	Enemies    []EnemySnapshot  `json:"enemies"`
	Objects    []ObjectSnapshot `json:"objects"`
	Drops      []DropSnapshot   `json:"drops"`
	TimeOfDay  string           `json:"timeOfDay"`
	Weather    string           `json:"weather,omitempty"`
	ZoneID     string           `json:"zoneId,omitempty"`
	Event      *WorldEventView  `json:"event,omitempty"`
	WorldBoss  *WorldBossView   `json:"worldBoss,omitempty"`
	InstanceID string           `json:"instanceId,omitempty"`
	Dungeon    *DungeonView     `json:"dungeon,omitempty"`
	Pvp        *PvpView         `json:"pvp,omitempty"`
	Clock      string           `json:"clock,omitempty"`
	ClockLabel string           `json:"clockLabel,omitempty"`
	WorldName  string           `json:"worldName,omitempty"`
}

type NPCSnapshot struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	Type        string  `json:"type"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Z           float64 `json:"z"`
	Yaw         float64 `json:"yaw"`
	Activity    string  `json:"activity,omitempty"`
	VoiceLineID string  `json:"voiceLineId,omitempty"`
	Marker      string  `json:"marker,omitempty"`
}

type ObjectSnapshot struct {
	ID   string  `json:"id"`
	Kind string  `json:"kind"`
	X    float64 `json:"x"`
	Z    float64 `json:"z"`
	Text string  `json:"text,omitempty"`
}

type EnemySnapshot struct {
	ID    string  `json:"id"`
	Kind  string  `json:"kind"`
	Name  string  `json:"name"`
	Level int     `json:"level"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Yaw   float64 `json:"yaw"`
	HP    int     `json:"hp"`
	MaxHP int     `json:"maxHp"`
	State string  `json:"st"`
	Rank  string  `json:"rank"`
}

type AttackIn struct {
	AttackType string  `json:"attackType"`
	TargetID   string  `json:"targetId"`
	Timestamp  int64   `json:"timestamp"`
	Direction  float64 `json:"direction"`
}

type SkillIn struct {
	SkillID   string  `json:"skillId"`
	TargetID  string  `json:"targetId"`
	Timestamp int64   `json:"timestamp"`
	Direction float64 `json:"direction"`
}

type DodgeIn struct {
	Timestamp int64   `json:"timestamp"`
	AX        float64 `json:"ax"`
	AZ        float64 `json:"az"`
	Yaw       float64 `json:"yaw"`
}

type RespawnIn struct {
	Timestamp int64 `json:"timestamp"`
}

type AttackResult struct {
	AttackerID string   `json:"attackerId"`
	AttackType string   `json:"attackType"`
	SkillID    string   `json:"skillId,omitempty"`
	TargetIDs  []string `json:"targetIds"`
	Timestamp  int64    `json:"timestamp"`
	ComboHits  int      `json:"comboHits,omitempty"`
	Finisher   bool     `json:"finisher,omitempty"`
}

type DamageResult struct {
	AttackerID  string  `json:"attackerId"`
	TargetID    string  `json:"targetId"`
	Damage      int     `json:"damage"`
	IsCritical  bool    `json:"isCritical"`
	Blocked     bool    `json:"blocked,omitempty"`
	HitX        float64 `json:"hitX"`
	HitY        float64 `json:"hitY"`
	HitZ        float64 `json:"hitZ"`
	AttackType  string  `json:"attackType"`
	Timestamp   int64   `json:"timestamp"`
	TargetHP    int     `json:"targetHp"`
	TargetMaxHP int     `json:"targetMaxHp"`
	Killed      bool    `json:"killed"`
	Kind        string  `json:"kind"`
}

type DeathOut struct {
	PlayerID  string `json:"playerId"`
	RespawnAt int64  `json:"respawnAt"`
}

type RespawnOut struct {
	PlayerID string  `json:"playerId"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Z        float64 `json:"z"`
	HP       int     `json:"hp"`
}

type LevelUpOut struct {
	PlayerID        string `json:"playerId"`
	FromLevel       int    `json:"fromLevel"`
	NewLevel        int    `json:"newLevel"`
	MaxHP           int    `json:"maxHp"`
	AttributePoints int    `json:"attributePoints"`
	SkillPoints     int    `json:"skillPoints"`
	Reward          string `json:"reward"`
}

type EnemyDeathOut struct {
	EnemyID  string `json:"enemyId"`
	KillerID string `json:"killerId"`
	Exp      int    `json:"exp"`
}

type RejectOut struct {
	Action   string `json:"action"`
	Reason   string `json:"reason"`
	PlayerID string `json:"playerId"`
}

type InteractIn struct {
	TargetID string `json:"targetId"`
	Kind     string `json:"kind"`
}

type QuestActionIn struct {
	QuestID string `json:"questId"`
}

type CollectItemIn struct {
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

type EducationAnswerIn struct {
	QuestionID string `json:"questionId"`
	Choice     int    `json:"choice"`
}

type DialogOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type QuestionOut struct {
	ID       string   `json:"id"`
	Index    int      `json:"index"`
	Total    int      `json:"total"`
	Category string   `json:"category"`
	Prompt   string   `json:"prompt"`
	Choices  []string `json:"choices"`
	ToID     string   `json:"toId,omitempty"`
}

type RewardView struct {
	Exp      int  `json:"exp"`
	Coin     int  `json:"coin"`
	Crystal  int  `json:"crystal,omitempty"`
	Potion   int  `json:"potion,omitempty"`
	EduToken int  `json:"eduToken,omitempty"`
	Perfect  bool `json:"perfect,omitempty"`
}

type ShopItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type InteractResult struct {
	Kind        string         `json:"kind"`
	TargetID    string         `json:"targetId"`
	Title       string         `json:"title"`
	Speaker     string         `json:"speaker"`
	Role        string         `json:"role"`
	Text        string         `json:"text"`
	Marker      string         `json:"marker"`
	Options     []DialogOption `json:"options"`
	Toast       string         `json:"toast,omitempty"`
	Locked      bool           `json:"locked,omitempty"`
	Shop        []ShopItem     `json:"shop,omitempty"`
	Question    *QuestionOut   `json:"question,omitempty"`
	Rewards     *RewardView    `json:"rewards,omitempty"`
	Subtitle    string         `json:"subtitle,omitempty"`
	CinematicID string         `json:"cinematicId,omitempty"`
	Emotion     string         `json:"emotion,omitempty"`
	Gesture     string         `json:"gesture,omitempty"`
	VoiceID     string         `json:"voiceId,omitempty"`
	History     []string       `json:"history,omitempty"`
}

type EducationFeedback struct {
	Correct  bool         `json:"correct"`
	Explain  string       `json:"explain"`
	Retry    bool         `json:"retry"`
	Toast    string       `json:"toast,omitempty"`
	Question *QuestionOut `json:"question,omitempty"`
	ToID     string       `json:"toId,omitempty"`
}

type QuestView struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Kind        string          `json:"kind"`
	State       string          `json:"state"`
	Description string          `json:"description"`
	NPC         string          `json:"npc"`
	NPCName     string          `json:"npcName"`
	Location    string          `json:"location"`
	Objectives  []ObjectiveView `json:"objectives"`
	Rewards     RewardView      `json:"rewards"`
}

type ObjectiveView struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	Text     string `json:"text"`
	Count    int    `json:"count"`
	Progress int    `json:"progress"`
}

type PlayerProgressOut struct {
	PlayerID          string          `json:"playerId"`
	Quests            []QuestView     `json:"quests"`
	Flags             map[string]bool `json:"flags"`
	Coin              int             `json:"coin"`
	Potion            int             `json:"potion"`
	Crystal           int             `json:"crystal"`
	EduToken          int             `json:"eduToken"`
	EnergyPotion      int             `json:"energyPotion"`
	ForestUnlocked    bool            `json:"forestUnlocked"`
	Claimed           []string        `json:"claimed"`
	FastTravel        []string        `json:"fastTravel"`
	TimeOfDay         string          `json:"timeOfDay"`
	ActiveQuestID     string          `json:"activeQuestId"`
	Chapters          []ChapterView   `json:"chapters,omitempty"`
	ZoneID            string          `json:"zoneId,omitempty"`
	Weather           string          `json:"weather,omitempty"`
	KnowledgePoints   int             `json:"knowledgePoints,omitempty"`
	Clock             string          `json:"clock,omitempty"`
	ClockLabel        string          `json:"clockLabel,omitempty"`
	RegionReputation  map[string]int  `json:"regionReputation,omitempty"`
	FactionReputation map[string]int  `json:"factionReputation,omitempty"`
	Journey           *JourneyView    `json:"journey,omitempty"`
}

type ItemEffects struct {
	HealPct        float64 `json:"healPct"`
	EnergyPct      float64 `json:"energyPct"`
	StaminaPct     float64 `json:"staminaPct"`
	Attack         int     `json:"attack"`
	Defense        int     `json:"defense"`
	MaxHP          int     `json:"maxHp"`
	MaxEnergy      int     `json:"maxEnergy"`
	Strength       int     `json:"strength"`
	Agility        int     `json:"agility"`
	EnergyPower    int     `json:"energyPower"`
	CriticalChance float64 `json:"criticalChance"`
	MovementSpeed  float64 `json:"movementSpeed"`
	Range          float64 `json:"range"`
	Dodge          int     `json:"dodge,omitempty"`
	EnergyRegen    float64 `json:"energyRegen,omitempty"`
}

type ItemDefView struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Type             string      `json:"type"`
	Slot             string      `json:"slot,omitempty"`
	Rarity           string      `json:"rarity"`
	Stackable        bool        `json:"stackable"`
	MaxStack         int         `json:"maxStack"`
	Icon             string      `json:"icon"`
	Value            int         `json:"value"`
	LevelRequirement int         `json:"levelRequirement"`
	Effects          ItemEffects `json:"effects"`
	Tradable         bool        `json:"tradable"`
	Bind             string      `json:"bind,omitempty"`
	ItemLevel        int         `json:"itemLevel,omitempty"`
	SetID            string      `json:"setId,omitempty"`
	Lore             string      `json:"lore,omitempty"`
}

type InvSlotView struct {
	Index    int          `json:"index"`
	Item     *ItemDefView `json:"item"`
	Qty      int          `json:"qty"`
	Locked     bool         `json:"locked,omitempty"`
	Favorite   bool         `json:"favorite,omitempty"`
	InstanceID string       `json:"itemInstanceId,omitempty"`
	Upgrade    int          `json:"upgrade,omitempty"`
	ItemLevel  int          `json:"itemLevel,omitempty"`
}

type EquipmentView struct {
	HEAD        string `json:"HEAD"`
	BODY        string `json:"BODY"`
	LEGS        string `json:"LEGS"`
	WEAPON      string `json:"WEAPON"`
	ACCESSORY_1 string `json:"ACCESSORY_1"`
	ACCESSORY_2 string `json:"ACCESSORY_2"`
	ACCESSORY_3 string `json:"ACCESSORY_3,omitempty"`
}

type StatsView struct {
	Level           int     `json:"level"`
	Class           string  `json:"class"`
	HP              int     `json:"hp"`
	MaxHP           int     `json:"maxHp"`
	Energy          int     `json:"energy"`
	MaxEnergy       int     `json:"maxEnergy"`
	Stamina         int     `json:"stamina"`
	Attack          int     `json:"attack"`
	Defense         int     `json:"defense"`
	Strength        int     `json:"strength"`
	Agility         int     `json:"agility"`
	EnergyPower     int     `json:"energyPower"`
	Vitality        int     `json:"vitality"`
	CriticalChance  float64 `json:"criticalChance"`
	MoveSpeed       float64 `json:"moveSpeed"`
	AttributePoints int     `json:"attributePoints"`
	SkillPoints     int     `json:"skillPoints"`
	PowerRating     int     `json:"powerRating"`
	FormID          string  `json:"formId,omitempty"`
	TransformState  string  `json:"transformState,omitempty"`
	Dodge           int     `json:"dodge,omitempty"`
	EnergyRegen     float64 `json:"energyRegen,omitempty"`
	ShowCosmetic    bool    `json:"showCosmetic,omitempty"`
}

type InventoryUpdated struct {
	PlayerID      string        `json:"playerId"`
	Version       int           `json:"inventoryVersion"`
	ChangedSlots  []InvSlotView `json:"changedSlots"`
	Slots         []InvSlotView `json:"slots,omitempty"`
	Equipment     EquipmentView `json:"equipment"`
	Stats         StatsView     `json:"stats"`
	Coin          int           `json:"coin"`
	Crystal       int           `json:"crystal"`
	EduToken      int           `json:"eduToken"`
	BattleToken   int           `json:"battleToken,omitempty"`
	GuardianToken int           `json:"guardianToken,omitempty"`
	RaidToken     int           `json:"raidToken,omitempty"`
	Bank          []InvSlotView `json:"bank,omitempty"`
	Toast         string        `json:"toast,omitempty"`
	TempLoot      []TempLootRow `json:"tempLoot,omitempty"`
	SetPieces     int           `json:"setPieces,omitempty"`
	BagCapacity   int           `json:"bagCapacity,omitempty"`
	ShowCosmetic  bool          `json:"showCosmetic,omitempty"`
	ItemHist      []ItemHistRow `json:"itemHistory,omitempty"`
}

type PickupIn struct {
	DropID   string `json:"dropId"`
	ItemID   string `json:"itemId"`
	Quantity int    `json:"quantity"`
}

type SlotIn struct {
	Slot      int    `json:"slot"`
	ItemID    string `json:"itemId"`
	EquipSlot string `json:"equipSlot"`
}

type PartyActionIn struct {
	TargetID  string `json:"targetId"`
	PartyID   string `json:"partyId"`
	Role      string `json:"role"`
	Activity  string `json:"activity"`
	MinLevel  int    `json:"minLevel"`
	ListingID string `json:"listingId"`
	LandmarkID string `json:"landmarkId"`
}

type InspectOut struct {
	PlayerID          string        `json:"playerId"`
	Name              string        `json:"name"`
	Level             int           `json:"level"`
	Class             string        `json:"class"`
	Stats             StatsView     `json:"stats"`
	Equipment         EquipmentView `json:"equipment"`
	PowerRating       int           `json:"powerRating"`
	Guild             string        `json:"guild,omitempty"`
	GuildTag          string        `json:"guildTag,omitempty"`
	Title             string        `json:"title,omitempty"`
	Rank              string        `json:"rank,omitempty"`
	Season            string        `json:"season,omitempty"`
	SeasonLevel       int           `json:"seasonLevel,omitempty"`
	Badge             string        `json:"badge,omitempty"`
	Aura              string        `json:"aura,omitempty"`
	Mount             string        `json:"mount,omitempty"`
	Avatar            string        `json:"avatar,omitempty"`
	Status            string        `json:"status,omitempty"`
	Region            string        `json:"region,omitempty"`
	AchievementScore  int           `json:"achievementScore,omitempty"`
}

type PartyMemberView struct {
	PlayerID  string  `json:"playerId"`
	Name      string  `json:"name"`
	Level     int     `json:"level"`
	Class     string  `json:"class"`
	HP        int     `json:"hp"`
	MaxHP     int     `json:"maxHp"`
	Energy    int     `json:"energy"`
	MaxEnergy int     `json:"maxEnergy"`
	Distance  float64 `json:"distance"`
	Online    bool    `json:"online"`
	Role      string  `json:"role"`
	Leader    bool    `json:"leader"`
	Ready     bool    `json:"ready,omitempty"`
	Status    string  `json:"status,omitempty"`
}

type PartyView struct {
	PartyID    string            `json:"partyId"`
	LeaderID   string            `json:"leaderId"`
	Members    []PartyMemberView `json:"members"`
	TargetID   string            `json:"targetId,omitempty"`
	TargetName string            `json:"targetName,omitempty"`
	TargetHP   int               `json:"targetHp,omitempty"`
	TargetMax  int               `json:"targetMaxHp,omitempty"`
	TargetLv   int               `json:"targetLevel,omitempty"`
	NotifyIDs  []string          `json:"notifyIds,omitempty"`
	Activity   string            `json:"activity,omitempty"`
	MinLevel   int               `json:"minLevel,omitempty"`
	NeedRole   string            `json:"requiredRole,omitempty"`
	State      string            `json:"state,omitempty"`
	Ready      map[string]bool   `json:"ready,omitempty"`
	WaypointID string            `json:"waypointId,omitempty"`
	WaypointX  float64           `json:"waypointX,omitempty"`
	WaypointZ  float64           `json:"waypointZ,omitempty"`
}

type FriendView struct {
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	Class    string `json:"class"`
	Online   bool   `json:"online"`
	LastSeen int64  `json:"lastSeen"`
	Status   string `json:"status,omitempty"`
	Guild    string `json:"guild,omitempty"`
	Title    string `json:"title,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Region   string `json:"region,omitempty"`
}

type NearbyView struct {
	PlayerID string  `json:"playerId"`
	Name     string  `json:"name"`
	Level    int     `json:"level"`
	Class    string  `json:"class"`
	Distance float64 `json:"distance"`
	Online   bool    `json:"online"`
}

type SocialState struct {
	Party    *PartyView   `json:"party"`
	Friends  []FriendView `json:"friends"`
	Pending  []FriendView `json:"pending"`
	Outgoing []FriendView `json:"outgoing"`
	Blocked  []string     `json:"blocked"`
	Nearby   []NearbyView `json:"nearby"`
	Notifies []NotifyView `json:"notifies,omitempty"`
	Guild    *GuildView   `json:"guild,omitempty"`
	Wallet   WalletView   `json:"wallet"`
	ToID     string       `json:"toId,omitempty"`
	Privacy  *PrivacyView `json:"privacy,omitempty"`
}

type PrivacyView struct {
	Friend string `json:"friend"`
	Party  string `json:"party"`
	Trade  string `json:"trade"`
	PM     string `json:"pm"`
}

type SocialNote struct {
	Kind     string `json:"kind"`
	Text     string `json:"text"`
	FromID   string `json:"fromId"`
	From     string `json:"from"`
	ToID     string `json:"toId,omitempty"`
	Priority string `json:"priority,omitempty"`
}

type DropSnapshot struct {
	ID     string  `json:"id"`
	ItemID string  `json:"itemId"`
	Name   string  `json:"name"`
	Qty    int     `json:"qty"`
	X      float64 `json:"x"`
	Z      float64 `json:"z"`
	Rarity string  `json:"rarity"`
}

type ErrorOut struct {
	Message string `json:"message"`
}

type DespawnOut struct {
	PlayerID string `json:"playerId"`
}

type DungeonActionIn struct {
	DungeonID  string `json:"dungeonId"`
	InstanceID string `json:"instanceId"`
	ClaimID    string `json:"claimId"`
	Ready      bool   `json:"ready"`
	BossHP     int    `json:"bossHp"`
	Role       string `json:"role"`
	Vote       string `json:"vote"`
	TargetID   string `json:"targetId"`
	Difficulty string `json:"difficulty"`
	Region     string `json:"region"`
	ShopItemID string `json:"shopItemId"`
	Language   string `json:"language"`
}

type DungeonOffer struct {
	DungeonID        string     `json:"dungeonId"`
	Name             string     `json:"name"`
	ChapterID        string     `json:"chapterId"`
	Kind             string     `json:"kind"`
	Description      string     `json:"description"`
	Difficulty       string     `json:"difficulty"`
	RecommendedLevel int        `json:"recommendedLevel"`
	MinPlayers       int        `json:"minPlayers"`
	MaxPlayers       int        `json:"maxPlayers"`
	TimeLimit        int        `json:"timeLimit"`
	Rewards          RewardView `json:"rewards"`
	Status           string     `json:"status"`
	Region           string     `json:"region,omitempty"`
	Difficulties     []string   `json:"difficulties,omitempty"`
	LockoutResetAt   int64      `json:"lockoutResetAt,omitempty"`
	LockoutLabel     string     `json:"lockoutLabel,omitempty"`
}

type DungeonReadyMember struct {
	PlayerID string `json:"playerId"`
	Ready    bool   `json:"ready"`
}

type DungeonReadyOut struct {
	DungeonID string               `json:"dungeonId"`
	LeaderID  string               `json:"leaderId"`
	Until     int64                `json:"until"`
	Members   []DungeonReadyMember `json:"members"`
	Cancelled bool                 `json:"cancelled,omitempty"`
	FromQueue bool                 `json:"fromQueue,omitempty"`
}

type DungeonLoading struct {
	InstanceID string `json:"instanceId"`
	DungeonID  string `json:"dungeonId"`
	Name       string `json:"name"`
	ToID       string `json:"toId,omitempty"`
}

type DungeonMemberView struct {
	PlayerID       string  `json:"playerId"`
	Name           string  `json:"name"`
	Level          int     `json:"level"`
	HP             int     `json:"hp"`
	MaxHP          int     `json:"maxHp"`
	Dead           bool    `json:"dead"`
	Downed         bool    `json:"downed"`
	Online         bool    `json:"online"`
	Distance       float64 `json:"distance"`
	Role           string  `json:"role"`
	ReviveProgress int     `json:"reviveProgress"`
	ReviveToken    int     `json:"reviveToken"`
	Energy         int     `json:"energy,omitempty"`
	MaxEnergy      int     `json:"maxEnergy,omitempty"`
}

type BossView struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Level   int    `json:"level"`
	HP      int    `json:"hp"`
	MaxHP   int    `json:"maxHp"`
	Phase   int    `json:"phase"`
	Enraged bool   `json:"enraged"`
	Alive   bool   `json:"alive"`
}

type DungeonView struct {
	InstanceID    string              `json:"instanceId"`
	DungeonID     string              `json:"dungeonId"`
	ChapterID     string              `json:"chapterId"`
	Kind          string              `json:"kind"`
	Name          string              `json:"name"`
	Title         string              `json:"title"`
	State         string              `json:"state"`
	Wave          int                 `json:"wave"`
	WaveTotal     int                 `json:"waveTotal"`
	Encounter     int                 `json:"encounter"`
	Enemies       int                 `json:"enemies"`
	Objective     string              `json:"objective"`
	ObjectiveType string              `json:"objectiveType"`
	Progress      int                 `json:"progress"`
	Count         int                 `json:"count"`
	TimeLeft      int                 `json:"timeLeft"`
	Rating        string              `json:"rating"`
	ClaimID       string              `json:"claimId,omitempty"`
	Chest         bool                `json:"chest"`
	Elapsed       int                 `json:"elapsed"`
	Toast         string              `json:"toast,omitempty"`
	BossLocked    bool                `json:"bossLocked"`
	CrystalShield bool                `json:"crystalShield"`
	PuzzleStep    int                 `json:"puzzleStep"`
	WipeCount     int                 `json:"wipeCount"`
	Synergy       bool                `json:"synergy"`
	Boss          *BossView           `json:"boss,omitempty"`
	Members       []DungeonMemberView `json:"members,omitempty"`
	Loot          []LootItemView      `json:"loot,omitempty"`
	Votes         map[string]string   `json:"votes,omitempty"`
	Difficulty    string              `json:"difficulty,omitempty"`
	Mechanic      string              `json:"mechanic,omitempty"`
	GuideHP       int                 `json:"guideHp,omitempty"`
	EduShield     bool                `json:"eduShield,omitempty"`
	LockoutLabel  string              `json:"lockoutLabel,omitempty"`
	Room          string              `json:"room,omitempty"`
	Checkpoint    string              `json:"checkpoint,omitempty"`
}

type QueueView struct {
	State      string `json:"state"`
	DungeonID  string `json:"dungeonId"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	Players    int    `json:"players"`
	MinPlayers int    `json:"minPlayers"`
	MaxPlayers int    `json:"maxPlayers"`
	WaitMs     int64  `json:"waitMs"`
	ToID       string `json:"toId,omitempty"`
}

type DungeonListOut struct {
	Dungeons       []DungeonOffer    `json:"dungeons"`
	Queue          *QueueView        `json:"queue,omitempty"`
	History        []DungeonRunRow   `json:"history,omitempty"`
	Board          []DungeonBoardRow `json:"board,omitempty"`
	RaidShop       []RaidShopItem    `json:"raidShop,omitempty"`
	RaidTokens     int               `json:"raidTokens,omitempty"`
	LockoutResetAt int64             `json:"lockoutResetAt,omitempty"`
	LockoutLabel   string            `json:"lockoutLabel,omitempty"`
}

type LootItemView struct {
	ItemID string `json:"itemId"`
	Name   string `json:"name"`
	Qty    int    `json:"qty"`
	Rarity string `json:"rarity"`
}

type LootResult struct {
	PlayerID string         `json:"playerId"`
	ClaimID  string         `json:"claimId"`
	Items    []LootItemView `json:"items"`
	Exp      int            `json:"exp"`
	Coin     int            `json:"coin"`
	Crystal  int            `json:"crystal"`
	ToID     string         `json:"toId,omitempty"`
}

type BossTelegraphOut struct {
	InstanceID    string  `json:"instanceId"`
	Skill         string  `json:"skill"`
	X             float64 `json:"x"`
	Z             float64 `json:"z"`
	Radius        float64 `json:"radius"`
	Until         int64   `json:"until"`
	VFX           string  `json:"vfx"`
	Shape         string  `json:"shape"`
	Interruptible bool    `json:"interruptible"`
	Pulse         bool    `json:"pulse"`
}

type BossAOEOut struct {
	InstanceID string  `json:"instanceId"`
	X          float64 `json:"x"`
	Z          float64 `json:"z"`
	Radius     float64 `json:"radius"`
	Damage     int     `json:"damage"`
}

type BossPhaseOut struct {
	InstanceID string `json:"instanceId"`
	Phase      int    `json:"phase"`
	Label      string `json:"label"`
}

type ChapterView struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Region        string     `json:"region"`
	Story         string     `json:"story"`
	BossID        string     `json:"bossId"`
	BossName      string     `json:"bossName"`
	RequiredLevel int        `json:"requiredLevel"`
	Reward        RewardView `json:"reward"`
	Status        string     `json:"status"`
	DungeonID     string     `json:"dungeonId"`
}

type ChapterListOut struct {
	Chapters []ChapterView `json:"chapters"`
}

type AttributeIn struct {
	Stat string `json:"stat"`
}

type SkillUnlockIn struct {
	NodeID  string `json:"nodeId"`
	SkillID string `json:"skillId"`
}

type TransformIn struct {
	FormID string `json:"formId"`
}

type TransformReject struct {
	Reason   string `json:"reason"`
	PlayerID string `json:"playerId"`
	FormID   string `json:"formId,omitempty"`
}

type TransformView struct {
	PlayerID  string `json:"playerId"`
	FormID    string `json:"formId"`
	Name      string `json:"name"`
	Visual    string `json:"visual"`
	State     string `json:"state"`
	Energy    int    `json:"energy"`
	MaxEnergy int    `json:"maxEnergy"`
	Until     int64  `json:"until"`
	AuraColor string `json:"auraColor"`
	Particles string `json:"particles"`
	Reason    string `json:"reason,omitempty"`
}

type FormView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Visual      string `json:"visual,omitempty"`
	Unlocked    bool   `json:"unlocked"`
	Active      bool   `json:"active"`
	EnergyCost  int    `json:"energyCost,omitempty"`
	Level       int    `json:"level,omitempty"`
	StoryName   string `json:"storyName,omitempty"`
	Passive     string `json:"passive,omitempty"`
	Mastery     int    `json:"mastery,omitempty"`
}

type SkillNodeView struct {
	ID            string  `json:"id"`
	SkillID       string  `json:"skillId"`
	Branch        string  `json:"branch"`
	BranchName    string  `json:"branchName,omitempty"`
	Cost          int     `json:"cost"`
	RequiredLevel int     `json:"requiredLevel"`
	Prerequisite  string  `json:"prerequisite,omitempty"`
	Unlocked      bool    `json:"unlocked"`
	Available     bool    `json:"available"`
	Name          string  `json:"name,omitempty"`
	Description   string  `json:"description,omitempty"`
	EnergyCost    int     `json:"energyCost,omitempty"`
	Cooldown      float64 `json:"cooldown,omitempty"`
	Effect        string  `json:"effect,omitempty"`
	Damage        int     `json:"damage,omitempty"`
	Range         float64 `json:"range,omitempty"`
}

type ProgressionView struct {
	PlayerID        string          `json:"playerId"`
	Level           int             `json:"level"`
	Exp             int             `json:"exp"`
	ExpToNext       int             `json:"expToNext"`
	AttributePoints int             `json:"attributePoints"`
	SkillPoints     int             `json:"skillPoints"`
	SpentSTR        int             `json:"spentStr"`
	SpentDEF        int             `json:"spentDef"`
	SpentAGI        int             `json:"spentAgi"`
	SpentENG        int             `json:"spentEng"`
	SpentVIT        int             `json:"spentVit"`
	UnlockedSkills  []string        `json:"unlockedSkills"`
	UnlockedForms   []string        `json:"unlockedForms"`
	FormID          string          `json:"formId"`
	TransformState  string          `json:"transformState"`
	TransEnergy     int             `json:"transEnergy"`
	MaxTransEnergy  int             `json:"maxTransEnergy"`
	PowerRating     int             `json:"powerRating"`
	Forms           []FormView      `json:"forms"`
	Nodes           []SkillNodeView `json:"nodes"`
	TransformReady  bool            `json:"transformReady"`
	CombatStyle     string          `json:"combatStyle,omitempty"`
	StyleMastery    map[string]int  `json:"styleMastery,omitempty"`
	FormMastery     map[string]int  `json:"formMastery,omitempty"`
	UltCharge       int             `json:"ultCharge,omitempty"`
	MaxUltCharge    int             `json:"maxUltCharge,omitempty"`
	Loadout         []string        `json:"loadout,omitempty"`
	Builds          []CombatBuild   `json:"builds,omitempty"`
	ActiveBuild     int             `json:"activeBuild,omitempty"`
	CombatRating    int             `json:"combatRating,omitempty"`
	AttrResetLeft   int             `json:"attrResetLeft,omitempty"`
	Training        *TrainingRecord `json:"training,omitempty"`
	Styles          []CombatStyleDef `json:"styles,omitempty"`
}

type StyleIn struct {
	StyleID string `json:"styleId"`
}

type BuildIn struct {
	Slot  int    `json:"slot"`
	Name  string `json:"name"`
	Style string `json:"style"`
}

type LoadoutIn struct {
	Skills    []string `json:"skills"`
	Ultimate  string   `json:"ultimate"`
}

type BlockIn struct {
	On        bool  `json:"on"`
	Timestamp int64 `json:"timestamp"`
}

type BuildListView struct {
	PlayerID string        `json:"playerId"`
	Slots    []CombatBuild `json:"slots"`
	Active   int           `json:"active"`
	ToID     string        `json:"toId,omitempty"`
}

type FastTravelIn struct {
	LandmarkID string `json:"landmarkId"`
}

type MountIn struct {
	MountID string `json:"mountId"`
}

type EventClaimIn struct {
	EventID string `json:"eventId"`
	Score   int    `json:"score"`
}

type ZoneDiscovered struct {
	PlayerID string `json:"playerId"`
	ZoneID   string `json:"zoneId"`
	Name     string `json:"name"`
	Exp      int    `json:"exp"`
	Toast    string `json:"toast"`
}

type LoreView struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	Region      string `json:"region,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Discovered  bool   `json:"discovered,omitempty"`
	Personality string `json:"personality,omitempty"`
	Mechanic    string `json:"mechanic,omitempty"`
}

type NpcRelView struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	XP           int    `json:"xp"`
	NextReward   string `json:"nextReward,omitempty"`
	Memory       bool   `json:"memory,omitempty"`
	Role         string `json:"role,omitempty"`
	Trait        string `json:"trait,omitempty"`
}

type ChoiceHistView struct {
	ID     string `json:"id"`
	Choice string `json:"choice"`
	Impact string `json:"impact,omitempty"`
}

type OverlayChapterView struct {
	Index  int    `json:"index"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Locked bool   `json:"locked,omitempty"`
}

type EnemyLoreView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	Personality string `json:"personality"`
	Mechanic    string `json:"mechanic"`
	Lore        string `json:"lore,omitempty"`
	Encountered bool   `json:"encountered"`
	Defeated    bool   `json:"defeated"`
	Discovered  bool   `json:"discovered"`
}

type LandmarkView struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Region     string  `json:"region"`
	Discovered bool    `json:"discovered"`
	X          float64 `json:"x"`
	Z          float64 `json:"z"`
}

type GuardianView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	ChapterID   string `json:"chapterId"`
	Region      string `json:"region"`
	Index       int    `json:"index,omitempty"`
	PlayerID    string `json:"playerId,omitempty"`
	Personality string `json:"personality,omitempty"`
	Weakness    string `json:"weakness,omitempty"`
	Story       string `json:"story,omitempty"`
	DefeatedAt  int64  `json:"defeatedAt,omitempty"`
	StoryName   string `json:"storyName,omitempty"`
	CodexStatus string `json:"codexStatus,omitempty"`
	Ally        bool   `json:"ally,omitempty"`
	MiniBoss    bool   `json:"miniBoss,omitempty"`
}

type WorldBossView struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Region    string  `json:"region"`
	State     string  `json:"state"`
	Announce  string  `json:"announce"`
	HP        int     `json:"hp"`
	MaxHP     int     `json:"maxHp"`
	Phase     int     `json:"phase"`
	PhaseName string  `json:"phaseName,omitempty"`
	Until     int64   `json:"until"`
	Players   int     `json:"players"`
	X         float64 `json:"x,omitempty"`
	Z         float64 `json:"z,omitempty"`
}

type CollectionBook struct {
	PlayerID       string   `json:"playerId"`
	Guardians      int      `json:"guardians"`
	GuardiansTotal int      `json:"guardiansTotal"`
	Tokens         int      `json:"tokens"`
	Locations      int      `json:"locations"`
	Titles         []string `json:"titles"`
	Cosmetics      []string `json:"cosmetics"`
	Achievements   []string `json:"achievements"`
	StoryCompleted bool     `json:"storyCompleted"`
	ExplorerMode   bool     `json:"explorerMode"`
	Aura           int      `json:"aura"`
	AuraTotal      int      `json:"auraTotal"`
	Mount          int      `json:"mount"`
	MountTotal     int      `json:"mountTotal"`
}

type ChannelIn struct {
	Channel string `json:"channel"`
}

type WorldBossClaimIn struct {
	BossID        string `json:"bossId"`
	TransactionID string `json:"transactionId"`
	Contribution  int    `json:"contribution"`
}

type RegionView struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Title             string `json:"title,omitempty"`
	Discovered        bool   `json:"discovered"`
	Unlocked          bool   `json:"unlocked"`
	Completion        int    `json:"completion"`
	RecommendedLevel  int    `json:"recommendedLevel,omitempty"`
	MinimumLevel      int    `json:"minimumLevel,omitempty"`
	EnemyTier         string `json:"enemyTier,omitempty"`
	ResourceTier      string `json:"resourceTier,omitempty"`
}

type MountView struct {
	PlayerID string `json:"playerId"`
	MountID  string `json:"mountId"`
	Mounted  bool   `json:"mounted"`
	Name     string `json:"name,omitempty"`
	Reason   string `json:"reason,omitempty"`
	State    string `json:"state,omitempty"`
}

type WorldEventView struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Kind         string  `json:"kind,omitempty"`
	State        string  `json:"state"`
	Phase        string  `json:"phase,omitempty"`
	Region       string  `json:"region"`
	Announce     string  `json:"announce"`
	AnnounceJV   string  `json:"announceJv,omitempty"`
	Objective    string  `json:"objective,omitempty"`
	Progress     int     `json:"progress,omitempty"`
	Need         int     `json:"need,omitempty"`
	GateHP       int     `json:"gateHp"`
	MaxGateHP    int     `json:"maxGateHp"`
	Until        int64   `json:"until"`
	StartsAt     int64   `json:"startsAt,omitempty"`
	StartsIn     int     `json:"startsIn,omitempty"`
	EndsIn       int     `json:"endsIn,omitempty"`
	Participants int     `json:"participants,omitempty"`
	Success      bool    `json:"success,omitempty"`
	X            float64 `json:"x,omitempty"`
	Z            float64 `json:"z,omitempty"`
}

type WorldJournal struct {
	PlayerID          string             `json:"playerId"`
	Guardians         []GuardianView     `json:"guardians"`
	Regions           []RegionView       `json:"regions"`
	Lore              []LoreView         `json:"lore"`
	Landmarks         []LandmarkView     `json:"landmarks"`
	GuardiansDefeated int                `json:"guardiansDefeated"`
	GuardiansTotal    int                `json:"guardiansTotal"`
	RegionsDiscovered int                `json:"regionsDiscovered"`
	RegionsTotal      int                `json:"regionsTotal"`
	Mounts            []string           `json:"mounts"`
	CelestialGate     bool               `json:"celestialGate"`
	Objective         string             `json:"objective"`
	Tokens            int                `json:"tokens,omitempty"`
	StoryCompleted    bool               `json:"storyCompleted,omitempty"`
	ExplorerMode      bool               `json:"explorerMode,omitempty"`
	Achievements      []string           `json:"achievements,omitempty"`
	Channel           string             `json:"channel,omitempty"`
	Language          string             `json:"language,omitempty"`
	StoryChapter      string             `json:"storyChapter,omitempty"`
	StoryState        string             `json:"storyState,omitempty"`
	StoryChapters     []StoryChapterView `json:"storyChapters,omitempty"`
	Allies            []string           `json:"allies,omitempty"`
	EndingID          string             `json:"endingId,omitempty"`
	NGPlus            int                `json:"ngPlus,omitempty"`
	Markers           []MapMarkerView    `json:"markers,omitempty"`
	NextWorldBoss     *WorldBossPreview  `json:"nextWorldBoss,omitempty"`
	WorldMood         string             `json:"worldMood,omitempty"`
	NpcBook           []NpcRelView       `json:"npcBook,omitempty"`
	ChoiceHistory     []ChoiceHistView   `json:"choiceHistory,omitempty"`
	OverlayChapters   []OverlayChapterView `json:"overlayChapters,omitempty"`
	EnemyLore         []EnemyLoreView    `json:"enemyLore,omitempty"`
	LoreCards         []LoreView         `json:"loreCards,omitempty"`
}

type WalletView struct {
	Coins           int `json:"coins"`
	Crystals        int `json:"crystals"`
	EducationTokens int `json:"educationTokens"`
	GuildTokens     int `json:"guildTokens"`
	BattleTokens    int `json:"battleTokens"`
	GuardianTokens  int `json:"guardianTokens"`
	RaidTokens      int `json:"raidTokens,omitempty"`
}

type PvpActionIn struct {
	Mode          string `json:"mode"`
	Emote         string `json:"emote"`
	TargetID      string `json:"targetId"`
	MatchID       string `json:"matchId"`
	Category      string `json:"category"`
	Board         string `json:"board"`
	ShopItemID    string `json:"shopItemId"`
	TransactionID string `json:"transactionId"`
	Ready         bool   `json:"ready"`
}

type PvpModeView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	TeamSize int    `json:"teamSize"`
	MinLevel int    `json:"minLevel"`
	Duration int    `json:"duration"`
	Map      string `json:"map"`
	Enabled  bool   `json:"enabled"`
	Status   string `json:"status"`
}

type PvpProfileView struct {
	Rating        int     `json:"rating"`
	Rank          string  `json:"rank"`
	RankName      string  `json:"rankName"`
	Division      string  `json:"division,omitempty"`
	RankVisual    string  `json:"rankVisual,omitempty"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	WinRate       float64 `json:"winRate"`
	PlacementLeft int     `json:"placementLeft"`
	WinStreak     int     `json:"winStreak"`
	LossStreak    int     `json:"lossStreak"`
	HighestRank   string  `json:"highestRank"`
	BattleToken   int     `json:"battleToken"`
	Season        string  `json:"season"`
	SeasonID      string  `json:"seasonId"`
}

type PvpSeasonView struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Number int    `json:"number"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Weeks  int    `json:"weeks"`
}

type PvpShopView struct {
	ShopItemID string `json:"shopItemId"`
	ItemID     string `json:"itemId"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Price      int    `json:"price"`
	Currency   string `json:"currency"`
}

type PvpRewardView struct {
	ID       string `json:"id"`
	Rank     string `json:"rank"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Unlocked bool   `json:"unlocked"`
}

type PvpQueueView struct {
	State     string `json:"state"`
	Mode      string `json:"mode"`
	Name      string `json:"name"`
	Players   int    `json:"players"`
	Need      int    `json:"need"`
	WaitMs    int64  `json:"waitMs"`
	WaitEstMs int64  `json:"waitEstMs,omitempty"`
	WaitNote  string `json:"waitNote,omitempty"`
	ToID      string `json:"toId,omitempty"`
}

type PvpMemberView struct {
	PlayerID   string `json:"playerId"`
	Name       string `json:"name"`
	Team       int    `json:"team"`
	Ready      bool   `json:"ready,omitempty"`
	HP         int    `json:"hp"`
	MaxHP      int    `json:"maxHp"`
	Alive      bool   `json:"alive"`
	Kills      int    `json:"kills"`
	Deaths     int    `json:"deaths"`
	Assists    int    `json:"assists"`
	Damage     int    `json:"damage"`
	PingMs     int    `json:"pingMs,omitempty"`
	SpectateID string `json:"spectateId,omitempty"`
}

type PvpPointView struct {
	ID        string  `json:"id"`
	Owner     int     `json:"owner"`
	Contested bool    `json:"contested"`
	ProgressA int     `json:"progressA"`
	ProgressB int     `json:"progressB"`
	X         float64 `json:"x"`
	Z         float64 `json:"z"`
}

type PvpView struct {
	MatchID    string          `json:"matchId"`
	Mode       string          `json:"mode"`
	Map        string          `json:"map"`
	State      string          `json:"state"`
	TimeLeft   int             `json:"timeLeft"`
	ScoreA     int             `json:"scoreA"`
	ScoreB     int             `json:"scoreB"`
	Members    []PvpMemberView `json:"members"`
	Points     []PvpPointView  `json:"points,omitempty"`
	KillFeed   []string        `json:"killFeed,omitempty"`
	Team       int             `json:"team"`
	InstanceID string          `json:"instanceId,omitempty"`
	Countdown  int64           `json:"countdown,omitempty"`
}

type PvpReadyOut struct {
	MatchID    string          `json:"matchId"`
	Mode       string          `json:"mode"`
	Name       string          `json:"name"`
	Until      int64           `json:"until"`
	Members    []PvpMemberView `json:"members"`
	InstanceID string          `json:"instanceId,omitempty"`
}

type PvpNearbyView struct {
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	Cosmetic string `json:"cosmetic,omitempty"`
}

type PvpLobbyOut struct {
	Modes         []PvpModeView         `json:"modes"`
	Profile       PvpProfileView        `json:"profile"`
	Season        PvpSeasonView         `json:"season"`
	Shop          []PvpShopView         `json:"shop"`
	Rewards       []PvpRewardView       `json:"rewards"`
	History       []PvpHistoryRow       `json:"history"`
	SeasonHistory []PvpSeasonHistoryRow `json:"seasonHistory,omitempty"`
	Nearby        []PvpNearbyView       `json:"nearby,omitempty"`
	Queue         *PvpQueueView         `json:"queue,omitempty"`
	Match         PvpView               `json:"match,omitempty"`
	Training      bool                  `json:"training,omitempty"`
	ToID          string                `json:"toId,omitempty"`
}

type PvpResultOut struct {
	MatchID      string   `json:"matchId"`
	Mode         string   `json:"mode"`
	Title        string   `json:"title"`
	Victory      bool     `json:"victory"`
	Draw         bool     `json:"draw,omitempty"`
	Kills        int      `json:"kills"`
	Deaths       int      `json:"deaths"`
	Assists      int      `json:"assists"`
	Damage       int      `json:"damage"`
	Objective    int      `json:"objective,omitempty"`
	Duration     int      `json:"duration,omitempty"`
	Mvp          bool     `json:"mvp,omitempty"`
	MvpName      string   `json:"mvpName,omitempty"`
	RatingChange int      `json:"ratingChange"`
	Rating       int      `json:"rating"`
	Rank         string   `json:"rank"`
	RankName     string   `json:"rankName"`
	Promoted     bool     `json:"promoted"`
	Demoted      bool     `json:"demoted"`
	BattleToken  int      `json:"battleToken"`
	SeasonXP     int      `json:"seasonXP,omitempty"`
	Rewards      []string `json:"rewards,omitempty"`
	InstanceID   string   `json:"instanceId,omitempty"`
	ToID         string   `json:"toId,omitempty"`
}

type NotifyView struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Read     bool   `json:"read"`
	At       int64  `json:"timestamp"`
	Priority string `json:"priority,omitempty"`
}

type ChatOut struct {
	Channel   string   `json:"channel"`
	FromID    string   `json:"fromId"`
	From      string   `json:"from"`
	Text      string   `json:"text"`
	ToID      string   `json:"toId,omitempty"`
	System    bool     `json:"system,omitempty"`
	NotifyIDs []string `json:"notifyIds,omitempty"`
}

type GuildMemberView struct {
	PlayerID     string `json:"playerId"`
	Name         string `json:"name"`
	Rank         string `json:"rank"`
	Contribution int    `json:"contribution"`
	JoinedAt     int64  `json:"joinedAt"`
}

type GuildView struct {
	ID           string            `json:"guildId"`
	Name         string            `json:"name"`
	Tag          string            `json:"tag"`
	LeaderID     string            `json:"leader"`
	Level        int               `json:"level"`
	Exp          int               `json:"exp"`
	Members      []GuildMemberView `json:"members"`
	Announcement string            `json:"announcement"`
	EmblemID     string            `json:"emblemId"`
	Quest        string            `json:"quest,omitempty"`
	Description  string            `json:"description,omitempty"`
	Capacity     int               `json:"capacity,omitempty"`
	Storage      []InvSlotView     `json:"storage,omitempty"`
	Logs         []GuildLog        `json:"logs,omitempty"`
	Apps         []string          `json:"applications,omitempty"`
	PendingLead  string            `json:"pendingLeader,omitempty"`
	QuestProg    int               `json:"questProgress,omitempty"`
	NotifyIDs    []string          `json:"notifyIds,omitempty"`
}

type TradeOfferView struct {
	PlayerID string          `json:"playerId"`
	Slots    []TradeSlotView `json:"slots"`
	Coin     int             `json:"coin"`
	Ready    bool            `json:"ready"`
	Confirm  bool            `json:"confirm"`
}

type TradeSlotView struct {
	Slot   int    `json:"slot"`
	ItemID string `json:"itemId"`
	Qty    int    `json:"qty"`
}

type TradeView struct {
	ID        string         `json:"tradeId"`
	TxID      string         `json:"transactionId"`
	A         TradeOfferView `json:"a"`
	B         TradeOfferView `json:"b"`
	State     string         `json:"state"`
	Result    string         `json:"result,omitempty"`
	NotifyIDs []string       `json:"notifyIds,omitempty"`
}

type ShopItemView struct {
	ShopItemID    string `json:"shopItemId"`
	ItemID        string `json:"itemId"`
	Name          string `json:"name"`
	Price         int    `json:"price"`
	Currency      string `json:"currency"`
	Stock         int    `json:"stock"`
	PurchaseLimit int    `json:"purchaseLimit"`
	Bought        int    `json:"bought"`
}

type FinderListing struct {
	ID       string `json:"id"`
	Leader   string `json:"leader"`
	Activity string `json:"activity"`
	Level    int    `json:"level"`
	Role     string `json:"requiredRole"`
	Players  int    `json:"players"`
	Cap      int    `json:"cap"`
}

type SearchHit struct {
	PlayerID string `json:"playerId"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	Guild    string `json:"guild"`
	Status   string `json:"status"`
}

func marshal(typ string, data any) []byte {
	raw, _ := json.Marshal(data)
	out, _ := json.Marshal(Envelope{Type: typ, Data: raw})
	return out
}

package mmo

import "sync"

const (
	QuestLocked    = "LOCKED"
	QuestAvailable = "AVAILABLE"
	QuestActive    = "ACTIVE"
	QuestCompleted = "COMPLETED"
	QuestFailed    = "FAILED"
	QuestClaimed   = "CLAIMED"
)

type QuestLog struct {
	ID       string
	State    string
	Progress []int
	Perfect  bool
	Wrongs   int
}

type QuizSession struct {
	QuestID string
	Index   int
	Active  bool
}

type PlayerLog struct {
	Quests              map[string]*QuestLog
	Flags               map[string]bool
	Coin                int
	Potion              int
	Crystal             int
	EduToken            int
	EnergyPotion        int
	ForestUnlocked      bool
	Claimed             map[string]bool
	FastTravel          map[string]bool
	DiscoveredZones     map[string]bool
	Landmarks           map[string]bool
	Guardians           map[string]bool
	Lore                map[string]bool
	Mounts              []string
	GuildID             string
	GuildTokens         int
	Titles              []string
	ActiveTitle         string
	Cosmetics           []string
	ActiveCosmetic      string
	Notices             []NotifyView
	BuyCounts           map[string]int
	BuyDay              string
	Quiz                QuizSession
	RaidLockout         map[string]string
	DailyDungeonDay     string
	DailyDungeonID      string
	DailyDungeonCount   int
	WeeklyChallengeWeek string
	WeeklyChallengeID   string
	WeeklyChallengeProg int
	PvpRating           int
	PvpWins             int
	PvpLosses           int
	PvpMatches          int
	PvpRankedMatches    int
	PvpPlacementLeft    int
	PvpWinStreak        int
	PvpLossStreak       int
	PvpHighestRank      string
	PvpSeasonID         string
	BattleToken         int
	PvpDamage           int
	PvpKills            int
	PvpDeaths           int
	PvpHistory          []PvpHistoryRow
	PvpClaimed          map[string]bool
	PvpCaptures         int
	PvpSeasonHistory    []PvpSeasonHistoryRow
	GuardianTokens      int
	Achievements        []string
	EventClaims         map[string]bool
	WorldBossClaims     map[string]bool
	GuardianTimes       map[string]int64
	DailyDay            string
	DailyProgress       map[string]int
	DailyClaimed        map[string]bool
	WeeklyWeek          string
	WeeklyProgress      map[string]int
	WeeklyClaimed       map[string]bool
	WeeklyBonusClaimed  bool
	ChallengeProgress   map[string]int
	ChallengeClaimed    map[string]bool
	SeasonTrackID       string
	SeasonXP            int
	SeasonLevel         int
	SeasonClaimed       map[string]bool
	SeasonXPDay         string
	SeasonXPToday       int
	SeasonHistory       []SeasonHistoryRow
	EduCorrect          int
	EduAnswered         int
	EduStreak           int
	EduLastDay          string
	EduCats             map[string]int
	DeathlessKills      int
	HorizonBest         int
	HorizonWeek         string
	HorizonClaimWeek    string
	LiveContrib         map[string]int
	LiveDay             string
	LiveDayAmt          int
	HorizonFragments    int
	ShowcaseBadge       string
	ShowcaseAura        string
	ShowcaseMount       string
	LastQuestion        string
	FirstClears         map[string]bool
	RaidTokens          int
	DungeonRuns         []DungeonRunRow
	PendingLoot         map[string][]LootItemView
	WeeklyDungeonBonus  string
	KnowledgePoints     int
	RegionRep           map[string]int
	FactionRep          map[string]int
	POI                 map[string]bool
	EncounterAt         int64
	DialogueChoices     map[string]string
	StoryCheckpoint     string
	Language            string
	StoryChapter        string
	StoryState          string
	StoryChoices        map[string]string
	SilumanProg         map[string]string
	StoryAllies         []string
	CinematicsSeen      map[string]bool
	ChapterRewards      map[string]bool
	PendingCinematic    string
	EndingID            string
	NGPlus              int
	HelpScore           int
	PowerScore          int
	RedeemScore         int
	ActiveMount         string
	MountFavorite       string
	MountFavorites      map[string]bool
	MountCosmetics      map[string]string
	RaceBest            map[string]int64
	RaceClaimed         map[string]bool
	FestivalTokens      int
	ExplorationXP       int
	TravelEvent         string
	CombatStyle         string
	StyleMastery        map[string]int
	FormMastery         map[string]int
	Builds              []CombatBuild
	ActiveBuild         int
	SkillResetAt        int64
	AttrResetUsed       int
	AttrResetUntil      int64
	StatusResist        map[string]float64
	Training            TrainingRecord
	Materials           map[string]int
	MaterialCodex       map[string]bool
	ProfessionXP        map[string]int
	ActiveGather        []string
	ActiveCraft         []string
	ProfessionResets    int
	Recipes             map[string]string
	RecipeCrafts        map[string]int
	CraftQueue          []CraftJob
	CraftUntil          int64
	FishSpot            string
	FishTargetA         int
	FishTargetB         int
	StallOpen           bool
	StallListings       []StallListing
	GearDurability      int
	Pets                []string
	PetNames            map[string]string
	PetHappy            map[string]int
	PetMood             map[string]string
	PetCosmetic         map[string]string
	ActivePet           string
	LifeFarmXP          int
	LifeXP              int
	LifeDailyDay        string
	LifeDailyClaimed    map[string]bool
	Collections         map[string]bool
	CollectionClaimed   map[string]bool
	MeetPlayers         map[string]bool
	FurnitureOwned      map[string]bool
	VoteDay             map[string]string
	NpcRel              map[string]int
	NpcMemory           map[string]bool
	DialogueHistory     []string
	EduAskAt            int64
	CombatHits          int
	PerfectDodges       int
	PrivacyFriend       string
	PrivacyParty        string
	PrivacyTrade        string
	PrivacyPM           string
	PresenceMode        string
	LocalMutes          map[string]bool
	NotifyFriendLogin   bool
	NotifyGuildLogin    bool
	BagBonus            int
	AutoLoot            bool
	AutoLootMin         string
	TempLoot            []TempLootRow
	GearLoadouts        []GearLoadout
	ItemHist            []ItemHistRow
	GoldHist            []GoldHistRow
	GoldToday           int
	GoldDay             string
	NodeFound           map[string]bool
}

type QuestRepository interface {
	Load(playerID string) *PlayerLog
	Save(playerID string, log *PlayerLog)
}

type MemoryQuestRepo struct {
	mu   sync.Mutex
	logs map[string]*PlayerLog
}

func NewMemoryQuestRepo() *MemoryQuestRepo {
	return &MemoryQuestRepo{logs: map[string]*PlayerLog{}}
}

func (r *MemoryQuestRepo) Load(playerID string) *PlayerLog {
	r.mu.Lock()
	defer r.mu.Unlock()
	src := r.logs[playerID]
	if src == nil {
		return nil
	}
	return cloneLog(src)
}

func (r *MemoryQuestRepo) Save(playerID string, log *PlayerLog) {
	if log == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logs[playerID] = cloneLog(log)
}

func cloneLog(src *PlayerLog) *PlayerLog {
	out := &PlayerLog{
		Quests:              map[string]*QuestLog{},
		Flags:               map[string]bool{},
		Coin:                src.Coin,
		Potion:              src.Potion,
		Crystal:             src.Crystal,
		EduToken:            src.EduToken,
		EnergyPotion:        src.EnergyPotion,
		ForestUnlocked:      src.ForestUnlocked,
		Claimed:             map[string]bool{},
		FastTravel:          map[string]bool{},
		DiscoveredZones:     map[string]bool{},
		Landmarks:           map[string]bool{},
		Guardians:           map[string]bool{},
		Lore:                map[string]bool{},
		Mounts:              append([]string{}, src.Mounts...),
		GuildID:             src.GuildID,
		GuildTokens:         src.GuildTokens,
		Titles:              append([]string{}, src.Titles...),
		ActiveTitle:         src.ActiveTitle,
		Cosmetics:           append([]string{}, src.Cosmetics...),
		ActiveCosmetic:      src.ActiveCosmetic,
		Notices:             append([]NotifyView{}, src.Notices...),
		BuyCounts:           map[string]int{},
		BuyDay:              src.BuyDay,
		Quiz:                src.Quiz,
		RaidLockout:         map[string]string{},
		DailyDungeonDay:     src.DailyDungeonDay,
		DailyDungeonID:      src.DailyDungeonID,
		DailyDungeonCount:   src.DailyDungeonCount,
		WeeklyChallengeWeek: src.WeeklyChallengeWeek,
		WeeklyChallengeID:   src.WeeklyChallengeID,
		WeeklyChallengeProg: src.WeeklyChallengeProg,
		PvpRating:           src.PvpRating,
		PvpWins:             src.PvpWins,
		PvpLosses:           src.PvpLosses,
		PvpMatches:          src.PvpMatches,
		PvpRankedMatches:    src.PvpRankedMatches,
		PvpPlacementLeft:    src.PvpPlacementLeft,
		PvpWinStreak:        src.PvpWinStreak,
		PvpLossStreak:       src.PvpLossStreak,
		PvpHighestRank:      src.PvpHighestRank,
		PvpSeasonID:         src.PvpSeasonID,
		BattleToken:         src.BattleToken,
		PvpDamage:           src.PvpDamage,
		PvpKills:            src.PvpKills,
		PvpDeaths:           src.PvpDeaths,
		PvpHistory:          append([]PvpHistoryRow{}, src.PvpHistory...),
		PvpClaimed:          map[string]bool{},
		PvpCaptures:         src.PvpCaptures,
		PvpSeasonHistory:    append([]PvpSeasonHistoryRow{}, src.PvpSeasonHistory...),
		GuardianTokens:      src.GuardianTokens,
		Achievements:        append([]string{}, src.Achievements...),
		EventClaims:         map[string]bool{},
		WorldBossClaims:     map[string]bool{},
		GuardianTimes:       map[string]int64{},
		DailyDay:            src.DailyDay,
		DailyProgress:       map[string]int{},
		DailyClaimed:        map[string]bool{},
		WeeklyWeek:          src.WeeklyWeek,
		WeeklyProgress:      map[string]int{},
		WeeklyClaimed:       map[string]bool{},
		WeeklyBonusClaimed:  src.WeeklyBonusClaimed,
		ChallengeProgress:   map[string]int{},
		ChallengeClaimed:    map[string]bool{},
		SeasonTrackID:       src.SeasonTrackID,
		SeasonXP:            src.SeasonXP,
		SeasonLevel:         src.SeasonLevel,
		SeasonClaimed:       map[string]bool{},
		SeasonXPDay:         src.SeasonXPDay,
		SeasonXPToday:       src.SeasonXPToday,
		SeasonHistory:       append([]SeasonHistoryRow{}, src.SeasonHistory...),
		EduCorrect:          src.EduCorrect,
		EduAnswered:         src.EduAnswered,
		EduStreak:           src.EduStreak,
		EduLastDay:          src.EduLastDay,
		EduCats:             map[string]int{},
		DeathlessKills:      src.DeathlessKills,
		HorizonBest:         src.HorizonBest,
		HorizonWeek:         src.HorizonWeek,
		HorizonClaimWeek:    src.HorizonClaimWeek,
		LiveContrib:         map[string]int{},
		LiveDay:             src.LiveDay,
		LiveDayAmt:          src.LiveDayAmt,
		HorizonFragments:    src.HorizonFragments,
		ShowcaseBadge:       src.ShowcaseBadge,
		ShowcaseAura:        src.ShowcaseAura,
		ShowcaseMount:       src.ShowcaseMount,
		LastQuestion:        src.LastQuestion,
		RaidTokens:          src.RaidTokens,
		DungeonRuns:         append([]DungeonRunRow{}, src.DungeonRuns...),
		WeeklyDungeonBonus:  src.WeeklyDungeonBonus,
		FirstClears:         map[string]bool{},
		PendingLoot:         map[string][]LootItemView{},
		KnowledgePoints:     src.KnowledgePoints,
		RegionRep:           map[string]int{},
		FactionRep:          map[string]int{},
		POI:                 map[string]bool{},
		EncounterAt:         src.EncounterAt,
		DialogueChoices:     map[string]string{},
		StoryCheckpoint:     src.StoryCheckpoint,
		Language:            src.Language,
		StoryChapter:        src.StoryChapter,
		StoryState:          src.StoryState,
		StoryChoices:        map[string]string{},
		SilumanProg:         map[string]string{},
		StoryAllies:         append([]string{}, src.StoryAllies...),
		CinematicsSeen:      map[string]bool{},
		ChapterRewards:      map[string]bool{},
		PendingCinematic:    src.PendingCinematic,
		EndingID:            src.EndingID,
		NGPlus:              src.NGPlus,
		HelpScore:           src.HelpScore,
		PowerScore:          src.PowerScore,
		RedeemScore:         src.RedeemScore,
		ActiveMount:         src.ActiveMount,
		MountFavorite:       src.MountFavorite,
		MountFavorites:      map[string]bool{},
		MountCosmetics:      map[string]string{},
		RaceBest:            map[string]int64{},
		RaceClaimed:         map[string]bool{},
		FestivalTokens:      src.FestivalTokens,
		ExplorationXP:       src.ExplorationXP,
		TravelEvent:         src.TravelEvent,
		CombatStyle:         src.CombatStyle,
		ActiveBuild:         src.ActiveBuild,
		SkillResetAt:        src.SkillResetAt,
		AttrResetUsed:       src.AttrResetUsed,
		AttrResetUntil:      src.AttrResetUntil,
		Training:            src.Training,
		ProfessionResets:    src.ProfessionResets,
		CraftUntil:          src.CraftUntil,
		FishSpot:            src.FishSpot,
		FishTargetA:         src.FishTargetA,
		FishTargetB:         src.FishTargetB,
		StallOpen:           src.StallOpen,
		GearDurability:      src.GearDurability,
		ActiveGather:        append([]string{}, src.ActiveGather...),
		ActiveCraft:         append([]string{}, src.ActiveCraft...),
		CraftQueue:          append([]CraftJob{}, src.CraftQueue...),
		StallListings:       append([]StallListing{}, src.StallListings...),
		Materials:           map[string]int{},
		MaterialCodex:       map[string]bool{},
		ProfessionXP:        map[string]int{},
		Recipes:             map[string]string{},
		RecipeCrafts:        map[string]int{},
		StyleMastery:        map[string]int{},
		FormMastery:         map[string]int{},
		StatusResist:        map[string]float64{},
		Builds:              append([]CombatBuild{}, src.Builds...),
		Pets:                append([]string{}, src.Pets...),
		ActivePet:           src.ActivePet,
		LifeFarmXP:          src.LifeFarmXP,
		LifeXP:              src.LifeXP,
		LifeDailyDay:        src.LifeDailyDay,
		PetNames:            map[string]string{},
		PetHappy:            map[string]int{},
		PetMood:             map[string]string{},
		PetCosmetic:         map[string]string{},
		LifeDailyClaimed:    map[string]bool{},
		Collections:         map[string]bool{},
		CollectionClaimed:   map[string]bool{},
		MeetPlayers:         map[string]bool{},
		FurnitureOwned:      map[string]bool{},
		VoteDay:             map[string]string{},
		NpcRel:              map[string]int{},
		NpcMemory:           map[string]bool{},
		DialogueHistory:     append([]string{}, src.DialogueHistory...),
		EduAskAt:            src.EduAskAt,
		CombatHits:          src.CombatHits,
		PerfectDodges:       src.PerfectDodges,
		PrivacyFriend:       src.PrivacyFriend,
		PrivacyParty:        src.PrivacyParty,
		PrivacyTrade:        src.PrivacyTrade,
		PrivacyPM:           src.PrivacyPM,
		PresenceMode:        src.PresenceMode,
		LocalMutes:          map[string]bool{},
		NotifyFriendLogin:   src.NotifyFriendLogin,
		NotifyGuildLogin:    src.NotifyGuildLogin,
		BagBonus:            src.BagBonus,
		AutoLoot:            src.AutoLoot,
		AutoLootMin:         src.AutoLootMin,
		TempLoot:            append([]TempLootRow{}, src.TempLoot...),
		ItemHist:            append([]ItemHistRow{}, src.ItemHist...),
		GoldHist:            append([]GoldHistRow{}, src.GoldHist...),
		GoldToday:           src.GoldToday,
		GoldDay:             src.GoldDay,
	}
	if len(src.GearLoadouts) > 0 {
		out.GearLoadouts = make([]GearLoadout, len(src.GearLoadouts))
		for i, g := range src.GearLoadouts {
			out.GearLoadouts[i] = GearLoadout{Slot: g.Slot, Gear: g.Gear, Up: copyIntMapVal(g.Up)}
		}
	}
	for k, v := range src.Flags {
		out.Flags[k] = v
	}
	for k, v := range src.Claimed {
		out.Claimed[k] = v
	}
	for k, v := range src.FastTravel {
		out.FastTravel[k] = v
	}
	for k, v := range src.DiscoveredZones {
		out.DiscoveredZones[k] = v
	}
	for k, v := range src.Landmarks {
		out.Landmarks[k] = v
	}
	for k, v := range src.Guardians {
		out.Guardians[k] = v
	}
	for k, v := range src.Lore {
		out.Lore[k] = v
	}
	for k, v := range src.BuyCounts {
		out.BuyCounts[k] = v
	}
	for k, v := range src.RaidLockout {
		out.RaidLockout[k] = v
	}
	for k, v := range src.PvpClaimed {
		out.PvpClaimed[k] = v
	}
	if src.EventClaims != nil {
		for k, v := range src.EventClaims {
			out.EventClaims[k] = v
		}
	}
	if src.WorldBossClaims != nil {
		for k, v := range src.WorldBossClaims {
			out.WorldBossClaims[k] = v
		}
	}
	if src.GuardianTimes != nil {
		for k, v := range src.GuardianTimes {
			out.GuardianTimes[k] = v
		}
	}
	copyIntMap(src.DailyProgress, out.DailyProgress)
	copyBoolMap(src.DailyClaimed, out.DailyClaimed)
	copyIntMap(src.WeeklyProgress, out.WeeklyProgress)
	copyBoolMap(src.WeeklyClaimed, out.WeeklyClaimed)
	copyIntMap(src.ChallengeProgress, out.ChallengeProgress)
	copyBoolMap(src.ChallengeClaimed, out.ChallengeClaimed)
	copyBoolMap(src.SeasonClaimed, out.SeasonClaimed)
	copyIntMap(src.EduCats, out.EduCats)
	copyIntMap(src.LiveContrib, out.LiveContrib)
	if src.FirstClears != nil {
		for k, v := range src.FirstClears {
			out.FirstClears[k] = v
		}
	}
	if src.PendingLoot != nil {
		for k, v := range src.PendingLoot {
			out.PendingLoot[k] = append([]LootItemView{}, v...)
		}
	}
	for id, q := range src.Quests {
		cp := *q
		cp.Progress = append([]int{}, q.Progress...)
		out.Quests[id] = &cp
	}
	copyIntMap(src.RegionRep, out.RegionRep)
	copyIntMap(src.FactionRep, out.FactionRep)
	copyBoolMap(src.POI, out.POI)
	if src.DialogueChoices != nil {
		for k, v := range src.DialogueChoices {
			out.DialogueChoices[k] = v
		}
	}
	if src.StoryChoices != nil {
		for k, v := range src.StoryChoices {
			out.StoryChoices[k] = v
		}
	}
	if src.MountFavorites != nil {
		for k, v := range src.MountFavorites {
			out.MountFavorites[k] = v
		}
	}
	if src.MountCosmetics != nil {
		for k, v := range src.MountCosmetics {
			out.MountCosmetics[k] = v
		}
	}
	if src.RaceBest != nil {
		for k, v := range src.RaceBest {
			out.RaceBest[k] = v
		}
	}
	if src.RaceClaimed != nil {
		for k, v := range src.RaceClaimed {
			out.RaceClaimed[k] = v
		}
	}
	if src.SilumanProg != nil {
		for k, v := range src.SilumanProg {
			out.SilumanProg[k] = v
		}
	}
	if src.CinematicsSeen != nil {
		for k, v := range src.CinematicsSeen {
			out.CinematicsSeen[k] = v
		}
	}
	if src.ChapterRewards != nil {
		for k, v := range src.ChapterRewards {
			out.ChapterRewards[k] = v
		}
	}
	if src.StyleMastery != nil {
		for k, v := range src.StyleMastery {
			out.StyleMastery[k] = v
		}
	}
	if src.FormMastery != nil {
		for k, v := range src.FormMastery {
			out.FormMastery[k] = v
		}
	}
	if src.Materials != nil {
		for k, v := range src.Materials {
			out.Materials[k] = v
		}
	}
	if src.MaterialCodex != nil {
		for k, v := range src.MaterialCodex {
			out.MaterialCodex[k] = v
		}
	}
	if src.ProfessionXP != nil {
		for k, v := range src.ProfessionXP {
			out.ProfessionXP[k] = v
		}
	}
	if src.Recipes != nil {
		for k, v := range src.Recipes {
			out.Recipes[k] = v
		}
	}
	if src.RecipeCrafts != nil {
		for k, v := range src.RecipeCrafts {
			out.RecipeCrafts[k] = v
		}
	}
	if src.StatusResist != nil {
		for k, v := range src.StatusResist {
			out.StatusResist[k] = v
		}
	}
	if src.PetNames != nil {
		for k, v := range src.PetNames {
			out.PetNames[k] = v
		}
	}
	copyIntMap(src.PetHappy, out.PetHappy)
	if src.PetMood != nil {
		for k, v := range src.PetMood {
			out.PetMood[k] = v
		}
	}
	if src.PetCosmetic != nil {
		for k, v := range src.PetCosmetic {
			out.PetCosmetic[k] = v
		}
	}
	copyBoolMap(src.LifeDailyClaimed, out.LifeDailyClaimed)
	copyBoolMap(src.Collections, out.Collections)
	copyBoolMap(src.CollectionClaimed, out.CollectionClaimed)
	copyBoolMap(src.MeetPlayers, out.MeetPlayers)
	copyBoolMap(src.FurnitureOwned, out.FurnitureOwned)
	if src.VoteDay != nil {
		for k, v := range src.VoteDay {
			out.VoteDay[k] = v
		}
	}
	copyIntMap(src.NpcRel, out.NpcRel)
	copyBoolMap(src.NpcMemory, out.NpcMemory)
	copyBoolMap(src.LocalMutes, out.LocalMutes)
	if src.NodeFound != nil {
		out.NodeFound = map[string]bool{}
		copyBoolMap(src.NodeFound, out.NodeFound)
	}
	return out
}

func newPlayerLog() *PlayerLog {
	log := &PlayerLog{
		Quests:          map[string]*QuestLog{},
		Flags:           map[string]bool{},
		Claimed:         map[string]bool{},
		FastTravel:      map[string]bool{},
		DiscoveredZones: map[string]bool{},
		Landmarks:       map[string]bool{},
		Guardians:       map[string]bool{},
		Lore:            map[string]bool{},
		RegionRep:       map[string]int{},
		FactionRep:      map[string]int{},
		POI:             map[string]bool{},
		DialogueChoices: map[string]string{},
		Language:        "jv",
		StoryChapter:    "st-ch01",
		StoryState:      StoryNotStarted,
		StoryChoices:    map[string]string{},
		SilumanProg:     map[string]string{},
		CinematicsSeen:  map[string]bool{},
		ChapterRewards:  map[string]bool{},
		NpcRel:          map[string]int{},
		NpcMemory:       map[string]bool{},
	}
	for k, v := range defaultFlags {
		log.Flags[k] = v
	}
	for _, def := range questCatalog {
		ql := &QuestLog{ID: def.ID, State: QuestLocked, Progress: make([]int, len(def.Objectives)), Perfect: true}
		if len(def.Prereq) == 0 {
			ql.State = QuestAvailable
		}
		log.Quests[def.ID] = ql
	}
	return log
}

func (p *Player) ensureLog() *PlayerLog {
	if p.Log == nil {
		p.Log = newPlayerLog()
	}
	if p.Log.Quests == nil {
		p.Log.Quests = map[string]*QuestLog{}
	}
	if p.Log.Flags == nil {
		p.Log.Flags = map[string]bool{}
	}
	if p.Log.Claimed == nil {
		p.Log.Claimed = map[string]bool{}
	}
	if p.Log.FastTravel == nil {
		p.Log.FastTravel = map[string]bool{}
	}
	if p.Log.DiscoveredZones == nil {
		p.Log.DiscoveredZones = map[string]bool{}
	}
	if p.Log.Landmarks == nil {
		p.Log.Landmarks = map[string]bool{}
	}
	if p.Log.Guardians == nil {
		p.Log.Guardians = map[string]bool{}
	}
	if p.Log.Lore == nil {
		p.Log.Lore = map[string]bool{}
	}
	if p.Log.BuyCounts == nil {
		p.Log.BuyCounts = map[string]int{}
	}
	if p.Log.RaidLockout == nil {
		p.Log.RaidLockout = map[string]string{}
	}
	if p.Log.PvpClaimed == nil {
		p.Log.PvpClaimed = map[string]bool{}
	}
	if p.Log.EventClaims == nil {
		p.Log.EventClaims = map[string]bool{}
	}
	if p.Log.WorldBossClaims == nil {
		p.Log.WorldBossClaims = map[string]bool{}
	}
	if p.Log.GuardianTimes == nil {
		p.Log.GuardianTimes = map[string]int64{}
	}
	if p.Log.DailyProgress == nil {
		p.Log.DailyProgress = map[string]int{}
	}
	if p.Log.DailyClaimed == nil {
		p.Log.DailyClaimed = map[string]bool{}
	}
	if p.Log.WeeklyProgress == nil {
		p.Log.WeeklyProgress = map[string]int{}
	}
	if p.Log.WeeklyClaimed == nil {
		p.Log.WeeklyClaimed = map[string]bool{}
	}
	if p.Log.ChallengeProgress == nil {
		p.Log.ChallengeProgress = map[string]int{}
	}
	if p.Log.ChallengeClaimed == nil {
		p.Log.ChallengeClaimed = map[string]bool{}
	}
	if p.Log.SeasonClaimed == nil {
		p.Log.SeasonClaimed = map[string]bool{}
	}
	if p.Log.EduCats == nil {
		p.Log.EduCats = map[string]int{}
	}
	if p.Log.LiveContrib == nil {
		p.Log.LiveContrib = map[string]int{}
	}
	if p.Log.FirstClears == nil {
		p.Log.FirstClears = map[string]bool{}
	}
	if p.Log.PendingLoot == nil {
		p.Log.PendingLoot = map[string][]LootItemView{}
	}
	if p.Log.PvpHistory == nil {
		p.Log.PvpHistory = []PvpHistoryRow{}
	}
	if p.Log.StoryChoices == nil {
		p.Log.StoryChoices = map[string]string{}
	}
	if p.Log.SilumanProg == nil {
		p.Log.SilumanProg = map[string]string{}
	}
	if p.Log.CinematicsSeen == nil {
		p.Log.CinematicsSeen = map[string]bool{}
	}
	if p.Log.ChapterRewards == nil {
		p.Log.ChapterRewards = map[string]bool{}
	}
	if p.Log.Language == "" {
		p.Log.Language = "jv"
	}
	if p.Log.StoryChapter == "" {
		p.Log.StoryChapter = "st-ch01"
	}
	if p.Log.StoryState == "" {
		p.Log.StoryState = StoryNotStarted
	}
	if p.Log.NpcRel == nil {
		p.Log.NpcRel = map[string]int{}
	}
	if p.Log.NpcMemory == nil {
		p.Log.NpcMemory = map[string]bool{}
	}
	return p.Log
}

func (p *Player) quest(id string) *QuestLog {
	log := p.ensureLog()
	q := log.Quests[id]
	if q != nil {
		return q
	}
	def, ok := questByID[id]
	if !ok {
		return nil
	}
	q = &QuestLog{ID: id, State: QuestLocked, Progress: make([]int, len(def.Objectives)), Perfect: true}
	if len(def.Prereq) == 0 {
		q.State = QuestAvailable
	}
	log.Quests[id] = q
	return q
}

func (p *Player) markDirty() {
	p.questDirty = true
}

func (w *WorldState) persist(p *Player) {
	if p == nil {
		return
	}
	if w.QuestRepo != nil {
		w.QuestRepo.Save(p.ID, p.ensureLog())
	}
	w.persistJourney(p)
}

func prereqMet(p *Player, def QuestDef) bool {
	for _, id := range def.Prereq {
		q := p.quest(id)
		if q == nil || q.State != QuestClaimed {
			return false
		}
	}
	return true
}

func (w *WorldState) refreshAvailability(p *Player) {
	for _, def := range questCatalog {
		q := p.quest(def.ID)
		if q == nil || q.State != QuestLocked {
			continue
		}
		if prereqMet(p, def) {
			q.State = QuestAvailable
			p.markDirty()
		}
	}
}

func objectivesDone(q *QuestLog, def QuestDef) bool {
	if q == nil {
		return false
	}
	for i, obj := range def.Objectives {
		need := obj.Count
		if need < 1 {
			need = 1
		}
		got := 0
		if i < len(q.Progress) {
			got = q.Progress[i]
		}
		if got < need {
			return false
		}
	}
	return true
}

func (p *Player) credit(kind, target string, amount int) bool {
	if amount <= 0 {
		return false
	}
	changed := false
	for _, def := range questCatalog {
		q := p.quest(def.ID)
		if q == nil || q.State != QuestActive {
			continue
		}
		for i, obj := range def.Objectives {
			if obj.Type != kind || obj.Target != target {
				continue
			}
			if i >= len(q.Progress) {
				q.Progress = append(q.Progress, make([]int, i+1-len(q.Progress))...)
			}
			need := obj.Count
			if need < 1 {
				need = 1
			}
			if q.Progress[i] >= need {
				continue
			}
			q.Progress[i] += amount
			if q.Progress[i] > need {
				q.Progress[i] = need
			}
			changed = true
		}
		if changed && objectivesDone(q, def) {
			q.State = QuestCompleted
		}
	}
	if changed {
		p.markDirty()
	}
	return changed
}

func (w *WorldState) creditKill(p *Player, kind string) {
	if p == nil {
		return
	}
	p.credit("KILL", kind, 1)
	w.noteActivity(p, "KILL", kind, 1)
	w.persist(p)
}

func applyReward(p *Player, r RewardDef) {
	p.ensureLog()
	p.Log.Coin += r.Coin
	p.Log.Potion += r.Potion
	p.Log.Crystal += r.Crystal
	p.Log.EduToken += r.EduToken
	p.markDirty()
}

func rewardView(r RewardDef, perfect bool) RewardView {
	return RewardView{Exp: r.Exp, Coin: r.Coin, Crystal: r.Crystal, Potion: r.Potion, EduToken: r.EduToken, Perfect: perfect}
}

func (p *Player) progressOut(timeOfDay string) PlayerProgressOut {
	log := p.ensureLog()
	views := make([]QuestView, 0, len(questCatalog))
	active := ""
	for _, def := range questCatalog {
		q := p.quest(def.ID)
		if q == nil {
			continue
		}
		if (q.State == QuestActive || q.State == QuestCompleted) && active == "" {
			active = def.ID
		}
		npcName := def.NPC
		if n, ok := npcByID[def.NPC]; ok {
			npcName = n.Name
		}
		objs := make([]ObjectiveView, 0, len(def.Objectives))
		for i, obj := range def.Objectives {
			prog := 0
			if i < len(q.Progress) {
				prog = q.Progress[i]
			}
			objs = append(objs, ObjectiveView{Type: obj.Type, Target: obj.Target, Text: obj.Text, Count: obj.Count, Progress: prog})
		}
		views = append(views, QuestView{
			ID:          def.ID,
			Title:       def.Title,
			Kind:        def.Kind,
			State:       q.State,
			Description: def.Description,
			NPC:         def.NPC,
			NPCName:     npcName,
			Location:    def.Location,
			Objectives:  objs,
			Rewards:     rewardView(def.Rewards, false),
		})
	}
	claimed := make([]string, 0, len(log.Claimed))
	for id, ok := range log.Claimed {
		if ok {
			claimed = append(claimed, id)
		}
	}
	ft := make([]string, 0, len(log.FastTravel))
	for id, ok := range log.FastTravel {
		if ok {
			ft = append(ft, id)
		}
	}
	flags := map[string]bool{}
	for k, v := range log.Flags {
		flags[k] = v
	}
	j := p.journeyView()
	return PlayerProgressOut{
		PlayerID:          p.ID,
		Quests:            views,
		Flags:             flags,
		Coin:              log.Coin,
		Potion:            log.Potion,
		Crystal:           log.Crystal,
		EduToken:          log.EduToken,
		EnergyPotion:      log.EnergyPotion,
		ForestUnlocked:    log.ForestUnlocked,
		Claimed:           claimed,
		FastTravel:        ft,
		TimeOfDay:         timeOfDay,
		ActiveQuestID:     active,
		Chapters:          playerChapterViews(p),
		ZoneID:            p.ZoneID,
		Weather:           "",
		KnowledgePoints:   log.KnowledgePoints,
		Clock:             "",
		ClockLabel:        "",
		RegionReputation:  log.RegionRep,
		FactionReputation: log.FactionRep,
		Journey:           &j,
	}
}

func npcMarker(p *Player, npc NPCDef) string {
	for _, id := range npc.QuestIDs {
		q := p.quest(id)
		if q == nil {
			continue
		}
		def := questByID[id]
		if q.State == QuestCompleted && def.ClaimAt == npc.ID {
			return "?"
		}
		if q.State == QuestAvailable && (def.NPC == npc.ID || def.ClaimAt == npc.ID) {
			return "!"
		}
	}
	for _, def := range questCatalog {
		q := p.quest(def.ID)
		if q == nil {
			continue
		}
		if q.State == QuestCompleted && def.ClaimAt == npc.ID {
			return "?"
		}
		if q.State == QuestAvailable && def.NPC == npc.ID {
			return "!"
		}
	}
	return ""
}

func (w *WorldState) FlushQuest(id string) []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.questDirty {
		return nil
	}
	p.questDirty = false
	out := p.progressOut(w.Time.Phase)
	out.Clock = w.Time.ClockText()
	out.ClockLabel = w.Time.ClockLabel()
	out.Weather = w.weatherForWatcher(p)
	return marshal(TypeQuestUpdated, out)
}

func (w *WorldState) acceptQuest(p *Player, questID string) [][]byte {
	def, ok := questByID[questID]
	if !ok {
		return rejectFor(p.ID, TypeQuestAccept, "quest")
	}
	q := p.quest(questID)
	if q == nil || q.State != QuestAvailable {
		return rejectFor(p.ID, TypeQuestAccept, "not_available")
	}
	if !prereqMet(p, def) {
		return rejectFor(p.ID, TypeQuestAccept, "prerequisite")
	}
	q.State = QuestActive
	p.markDirty()
	if len(def.Objectives) > 0 && def.Objectives[0].Type == "TALK" {
		p.credit("TALK", def.Objectives[0].Target, 1)
	}
	w.refreshAvailability(p)
	w.persist(p)
	if def.Kind == "education" {
		p.questDirty = false
		return w.startQuiz(p)
	}
	text := "Misi diterima."
	if line, ok := dialogueCatalog[npcByID[def.NPC].DialogueID+"_offer"]; ok {
		text = line.Text
	}
	return [][]byte{
		marshal(TypeInteractResult, InteractResult{
			Kind: "quest", TargetID: def.NPC, Title: def.Title, Speaker: npcByID[def.NPC].Name,
			Role: npcByID[def.NPC].Role, Text: text, Marker: npcMarker(p, npcByID[def.NPC]),
			Options: claimOrClose(p, def),
		}),
		marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
	}
}

func (w *WorldState) declineQuest(p *Player, questID string) [][]byte {
	q := p.quest(questID)
	if q == nil || q.State != QuestAvailable {
		return rejectFor(p.ID, TypeQuestDecline, "not_available")
	}
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "npc", TargetID: questByID[questID].NPC, Title: questByID[questID].Title,
		Speaker: npcByID[questByID[questID].NPC].Name, Text: "Baik. Kembalilah jika kau siap.",
		Options: []DialogOption{{ID: "close", Label: "Tutup"}},
	})}
}

func claimOrClose(p *Player, def QuestDef) []DialogOption {
	q := p.quest(def.ID)
	if q != nil && q.State == QuestCompleted {
		return []DialogOption{{ID: "claim:" + def.ID, Label: "Ambil hadiah"}}
	}
	return []DialogOption{{ID: "close", Label: "Tutup"}}
}

func (w *WorldState) claimQuest(p *Player, questID string) [][]byte {
	def, ok := questByID[questID]
	if !ok {
		return rejectFor(p.ID, TypeQuestClaim, "quest")
	}
	q := p.quest(questID)
	if q == nil || q.State != QuestCompleted {
		return rejectFor(p.ID, TypeQuestClaim, "not_complete")
	}
	npc := npcByID[def.ClaimAt]
	need := npc.InteractionRange
	if need <= 0 {
		need = 2.6
	}
	if hypot2(p.X, p.Z, npc.X, npc.Z) > (need+0.6)*(need+0.6) {
		return rejectFor(p.ID, TypeQuestClaim, "distance")
	}
	q.State = QuestClaimed
	events := w.giveRewardBundle(p, def.Rewards)
	if def.PerfectBonus != nil && q.Perfect && q.Wrongs == 0 {
		events = append(events, w.giveRewardBundle(p, *def.PerfectBonus)...)
	}
	for _, flag := range def.FlagsOnClaim {
		p.ensureLog().Flags[flag] = true
	}
	if def.UnlockForest {
		p.ensureLog().ForestUnlocked = true
		p.ensureLog().Flags["forest_unlocked"] = true
	}
	p.markDirty()
	w.refreshAvailability(p)
	w.persist(p)
	rv := rewardView(def.Rewards, q.Perfect && q.Wrongs == 0)
	if def.PerfectBonus != nil && rv.Perfect {
		rv.Exp += def.PerfectBonus.Exp
		rv.Coin += def.PerfectBonus.Coin
		rv.EduToken += def.PerfectBonus.EduToken
	}
	out := [][]byte{
		marshal(TypeQuestReward, InteractResult{
			Kind: "reward", TargetID: def.ClaimAt, Title: "QUEST COMPLETE", Speaker: npc.Name,
			Text: def.Title + " selesai.", Rewards: &rv,
			Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		}),
		marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
	}
	out = append(out, events...)
	w.audit("questCompleted", p.ID, questID)
	if p.ensureLog().Flags["storyCompleted"] {
		out = append(out, w.completeStoryChapter(p)...)
	}
	out = append(out, w.afterStoryQuestClaim(p, def)...)
	out = append(out, w.afterMountQuestClaim(p, def)...)
	return out
}

func hypot2(x1, z1, x2, z2 float64) float64 {
	dx, dz := x1-x2, z1-z2
	return dx*dx + dz*dz
}

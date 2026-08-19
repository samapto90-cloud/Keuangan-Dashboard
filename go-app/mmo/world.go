package mmo

import (
	"sync"
	"time"
)

type Player struct {
	ID, SessionID, Name, Class        string
	Level, HP, MaxHP                  int
	Energy, MaxEnergy                 int
	Exp, ExpToNext                    int
	Strength, Defense                 int
	Stamina, MaxStamina               float64
	X, Y, Z, Yaw                      float64
	VX, VY, VZ                        float64
	AX, AZ, CamYaw                    float64
	Sprint, JumpQueued                bool
	Grounded                          bool
	State                             string
	CombatState                       string
	Seq                               uint32
	Connected                         bool
	LastInputAt                       time.Time
	LastHeard                         time.Time
	LastUpdate                        time.Time
	DroppedInputs                     int
	InvulnUntil                       time.Time
	IFrameUntil                       time.Time
	DodgeUntil                        time.Time
	DodgeCDUntil                      time.Time
	HitUntil                          time.Time
	StunUntil                         time.Time
	SilenceUntil                      time.Time
	SlowUntil                         time.Time
	SlowFactor                        float64
	PvpNoActUntil                     time.Time
	AttackCDUntil                     time.Time
	ComboUntil                        time.Time
	InCombatUntil                     time.Time
	RespawnAt                         time.Time
	ComboStep                         int
	ComboKind                         string
	SkillCD                           map[string]time.Time
	hpRegenAcc                        float64
	energyAcc                         float64
	Log                               *PlayerLog
	questDirty                        bool
	Bag                               *Inventory
	Bank                              *Inventory
	Gear                              EquipmentSet
	BaseMaxHP, BaseMaxEnergy          int
	BaseStrength, BaseDefense         int
	BaseAgility, BaseEnergyPower      int
	Agility, EnergyPower, EquipAttack int
	CritChance                        float64
	MoveSpeedBonus                    float64
	AttackRange                       float64
	PartyID                           string
	InstanceID                        string
	AttributePoints, SkillPoints      int
	SpentSTR, SpentDEF, SpentAGI      int
	SpentENG, SpentVIT                int
	UnlockedSkills                    []string
	UnlockedForms                     []string
	FormID, TransformState            string
	TransEnergy, MaxTransEnergy       int
	TransEnergyAcc                    float64
	TransformUntil                    time.Time
	TransformCDUntil                  time.Time
	ComboChain                        string
	ExpBoostPct                       float64
	ExpBoostUntil                     time.Time
	ZoneID                            string
	Mounted                           bool
	MountID                           string
	MountState                        string
	PetID                             string
	Swimming                          bool
	TravelOK                          bool
	FollowParty                       bool
	RaceID                            string
	RaceStep                          int
	RaceStartedAt                     int64
	GuildID, GuildTag, Title          string
	Region                            string
	PingMs                            int
	Suspicious                        float64
	Channel                           string
	CombatStyle                       string
	Blocking                          bool
	Charging                          bool
	ChargeUntil                       time.Time
	CombatStreak                      int
	GuardUntil                        time.Time
	PerfectDodgeUntil                 time.Time
	UltCharge                         int
	Stagger, StaggerMax               int
	LoadoutSkills                     []string
	LoadoutUlt                        string
	AttrResetUsed                     int
	SkillResetUntil                   time.Time
	StatusUntil                       map[string]time.Time
	TrainingHits, TrainingDmg         int
	TrainingStart                     time.Time
	ComboMasteryReady                 bool
	GearUp                            map[string]int
	GearInst                          map[string]string
	ShowCosmetic                      bool
	send                              chan []byte
	closeOnce                         sync.Once
}

func (p *Player) CloseSend() {
	p.closeOnce.Do(func() { close(p.send) })
}

func (p *Player) Spawn() PlayerSpawn {
	return PlayerSpawn{
		PlayerID:    p.ID,
		Name:        p.Name,
		Level:       p.Level,
		Class:       p.Class,
		HP:          p.HP,
		MaxHP:       p.MaxHP,
		Energy:      p.Energy,
		MaxEnergy:   p.MaxEnergy,
		Exp:         p.Exp,
		ExpToNext:   p.ExpToNext,
		X:           p.X,
		Y:           p.Y,
		Z:           p.Z,
		Yaw:         p.Yaw,
		State:       p.State,
		CombatState: p.CombatState,
		FormID:      p.FormID,
		GuildTag:    p.GuildTag,
		Title:       p.Title,
	}
}

func (p *Player) Snap() PlayerSnapshot {
	return PlayerSnapshot{
		PlayerID:    p.ID,
		X:           p.X,
		Y:           p.Y,
		Z:           p.Z,
		Yaw:         p.Yaw,
		VX:          p.VX,
		VY:          p.VY,
		VZ:          p.VZ,
		State:       p.State,
		CombatState: p.CombatState,
		HP:          p.HP,
		MaxHP:       p.MaxHP,
		Energy:      p.Energy,
		MaxEnergy:   p.MaxEnergy,
		Stamina:     int(p.Stamina + 0.5),
		Level:       p.Level,
		Exp:         p.Exp,
		ExpToNext:   p.ExpToNext,
		Seq:         p.Seq,
		FormID:      p.FormID,
		Transform:   p.TransformState,
		MountID:     p.MountID,
		Mounted:     p.Mounted,
		MountState:  p.MountState,
		Swimming:    p.Swimming,
		ZoneID:      p.ZoneID,
		GuildTag:    p.GuildTag,
		Title:       p.Title,
		PetID:       p.PetID,
	}
}

type WorldState struct {
	WorldID   string
	ChannelID string
	Time      WorldTimeSystem
	Events    WorldEventManager
	QuestRepo   QuestRepository
	InvRepo     InventoryRepository
	JourneyRepo JourneyRepository
	Parties   *PartyHub
	Social    *SocialHub
	Dungeons  *DungeonHub
	PvP       *PvpHub
	Guilds    *GuildHub
	Trades    *TradeHub
	Finder    *FinderHub
	Audit     []AuditEntry
	Reports   []Report
	TxDone    map[string][][]byte
	Mutes     map[string]time.Time
	limits    map[string][]time.Time
	mu        sync.Mutex
	players   map[string]*Player
	enemies   map[string]*Enemy
	drops     map[string]*WorldDrop
	spawnIdx  int
	Boss      *WorldBossLive
	Community *CommunityLive
	HorizonLB []HorizonScore
	HorizonLBWeek string
	HorizonHistory []HorizonScore
	Market    *MarketHub
	Housing   *HousingHub
	EconomyLog []EconomyTx
	ChatLog   []ChatOut
	lastChat  map[string]string
	DungeonBoard   []DungeonBoardRow
	DungeonHistory []DungeonRunRow
	npcPos         map[string]npcLive
	NodeCooldown   map[string]int64
	Stalls         map[string]*PlayerStall
	CraftOrders    map[string]*CraftOrder
	TradeLogs      []TradeLogRow
	GoldCreated    int
	GoldRemoved    int
	TradeVolume    int
	CraftVolume    int
	ItemHist       []ItemHistRow
	GuildContrib   []GuildContribRow
}

func NewWorldState() *WorldState {
	w := &WorldState{
		WorldID:   WorldID,
		ChannelID: ChannelID,
		Time:      NewWorldTimeSystem(),
		Events:    NewWorldEventManager(),
		QuestRepo: NewMemoryQuestRepo(),
		InvRepo:   NewMemoryInvRepo(),
		Parties:   NewPartyHub(),
		Social:    NewSocialHub(),
		Dungeons:  NewDungeonHub(),
		PvP:       NewPvpHub(),
		Guilds:    NewGuildHub(),
		Trades:    NewTradeHub(),
		Finder:    NewFinderHub(),
		TxDone:    map[string][][]byte{},
		Mutes:     map[string]time.Time{},
		limits:    map[string][]time.Time{},
		players:   map[string]*Player{},
		enemies:   map[string]*Enemy{},
		drops:     map[string]*WorldDrop{},
		Market:    &MarketHub{History: map[string][]MarketHistory{}},
		Housing:   &HousingHub{ByOwner: map[string]*HouseInstance{}, ByID: map[string]*HouseInstance{}},
		lastChat:     map[string]string{},
		NodeCooldown: map[string]int64{},
		Stalls:       map[string]*PlayerStall{},
		CraftOrders:  map[string]*CraftOrder{},
	}
	w.seedEnemies()
	w.seedWorldGuardians()
	return w
}

func (w *WorldState) Add(p *Player) (ok bool, isNew bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if cur, exists := w.players[p.ID]; exists {
		p.X, p.Y, p.Z = cur.X, cur.Y, cur.Z
		p.Yaw, p.VX, p.VY, p.VZ = cur.Yaw, cur.VX, cur.VY, cur.VZ
		p.AX, p.AZ, p.CamYaw = cur.AX, cur.AZ, cur.CamYaw
		p.State, p.Grounded = cur.State, cur.Grounded
		p.Level, p.Class, p.HP, p.MaxHP = cur.Level, cur.Class, cur.HP, cur.MaxHP
		p.Energy, p.MaxEnergy, p.Exp, p.ExpToNext = cur.Energy, cur.MaxEnergy, cur.Exp, cur.ExpToNext
		p.Strength, p.Defense = cur.Strength, cur.Defense
		p.Stamina, p.MaxStamina = cur.Stamina, cur.MaxStamina
		p.CombatState = cur.CombatState
		p.SkillCD = cur.SkillCD
		p.ComboStep, p.ComboKind, p.ComboUntil = cur.ComboStep, cur.ComboKind, cur.ComboUntil
		p.RespawnAt = cur.RespawnAt
		p.Name = cur.Name
		p.Log = cur.Log
		p.questDirty = cur.questDirty
		p.Bag = cur.Bag
		p.Gear = cur.Gear
		p.Bank = cur.Bank
		p.BaseMaxHP, p.BaseMaxEnergy = cur.BaseMaxHP, cur.BaseMaxEnergy
		p.BaseStrength, p.BaseDefense = cur.BaseStrength, cur.BaseDefense
		p.BaseAgility, p.BaseEnergyPower = cur.BaseAgility, cur.BaseEnergyPower
		p.PartyID = cur.PartyID
		p.InstanceID = cur.InstanceID
		p.AttributePoints, p.SkillPoints = cur.AttributePoints, cur.SkillPoints
		p.SpentSTR, p.SpentDEF, p.SpentAGI, p.SpentENG, p.SpentVIT = cur.SpentSTR, cur.SpentDEF, cur.SpentAGI, cur.SpentENG, cur.SpentVIT
		p.UnlockedSkills = append([]string{}, cur.UnlockedSkills...)
		p.UnlockedForms = append([]string{}, cur.UnlockedForms...)
		p.FormID, p.TransformState = cur.FormID, cur.TransformState
		p.TransEnergy, p.MaxTransEnergy = cur.TransEnergy, cur.MaxTransEnergy
		p.TransformUntil, p.TransformCDUntil = cur.TransformUntil, cur.TransformCDUntil
		p.ComboChain, p.ComboUntil = cur.ComboChain, cur.ComboUntil
		p.ExpBoostPct, p.ExpBoostUntil = cur.ExpBoostPct, cur.ExpBoostUntil
		p.ZoneID, p.Mounted, p.MountID = cur.ZoneID, cur.Mounted, cur.MountID
		p.GuildID, p.GuildTag, p.Title = cur.GuildID, cur.GuildTag, cur.Title
		p.Region, p.PingMs = cur.Region, cur.PingMs
		p.Channel = cur.Channel
		p.CombatStyle, p.UltCharge = cur.CombatStyle, cur.UltCharge
		p.LoadoutSkills = append([]string{}, cur.LoadoutSkills...)
		p.LoadoutUlt = cur.LoadoutUlt
		p.AttrResetUsed, p.SkillResetUntil = cur.AttrResetUsed, cur.SkillResetUntil
		p.initCombat()
		p.applyDerived()
		p.Seq = cur.Seq
		p.Connected = true
		p.LastHeard = time.Now()
		p.LastUpdate = time.Now()
		w.phase25SafeReconnect(p)
		w.phase26RestoreSocial(p)
		w.players[p.ID] = p
		return true, false
	}
	if len(w.players) >= MaxPlayers {
		return false, false
	}
	pt := spawnPoints[w.spawnIdx%len(spawnPoints)]
	w.spawnIdx++
	p.X, p.Y, p.Z = pt[0], pt[1], pt[2]
	p.Yaw = 0
	w.restoreJourney(p)
	p.State = "IDLE"
	p.CombatState = "IDLE"
	p.initCombat()
	p.Grounded = true
	p.Connected = true
	p.LastHeard = time.Now()
	p.LastUpdate = time.Now()
	p.LastInputAt = time.Now()
	if p.Region == "" {
		p.Region = "ASIA"
	}
	if saved := w.QuestRepo.Load(p.ID); saved != nil {
		p.Log = saved
	} else {
		p.Log = newPlayerLog()
		w.QuestRepo.Save(p.ID, p.Log)
	}
	p.GuildID = p.ensureLog().GuildID
	if w.Guilds != nil && p.GuildID != "" {
		if g := w.Guilds.ByID[p.GuildID]; g != nil {
			p.GuildTag = g.Tag
			w.Guilds.ByPlayer[p.ID] = g.ID
		}
	}
	w.refreshTitles(p)
	if gear := w.InvRepo.Load(p.ID); gear != nil {
		p.applyGearSave(gear)
	} else {
		p.ensureGear()
		p.applyDerived()
		w.InvRepo.Save(p.ID, p.gearSave())
	}
	if pt := w.Parties.Of(p.ID); pt != nil {
		p.PartyID = pt.ID
	}
	w.phase26RestoreSocial(p)
	if w.Dungeons != nil {
		if slot := w.Dungeons.pending[p.ID]; slot != nil && time.Now().Before(slot.Until) {
			if inst := w.Dungeons.instances[slot.InstanceID]; inst != nil && inst.State != DunClosing && inst.State != DunCompleted && inst.State != DunFailed {
				p.InstanceID = inst.ID
				p.X, p.Y, p.Z = slot.X, slot.Y, slot.Z
				w.Dungeons.byPlayer[p.ID] = inst.ID
				delete(inst.OfflineSince, p.ID)
				if !containsID(inst.Players, p.ID) {
					inst.Players = append(inst.Players, p.ID)
				}
			}
			delete(w.Dungeons.pending, p.ID)
		}
	}
	if w.PvP != nil {
		if slot := w.PvP.pending[p.ID]; slot != nil && time.Now().Before(slot.Until) {
			if inst := w.PvP.instances[slot.InstanceID]; inst != nil && inst.State != PvpCompleted && inst.State != PvpCancelled && inst.State != PvpEnding {
				p.InstanceID = inst.ID
				p.X, p.Y, p.Z = slot.X, slot.Y, slot.Z
				w.PvP.byPlayer[p.ID] = inst.ID
				delete(inst.Offline, p.ID)
			}
			delete(w.PvP.pending, p.ID)
		}
	}
	if p.Channel == "" {
		p.Channel = "1"
		if w.channelCount("1") >= ChannelMaxPop {
			p.Channel = w.leastChannel()
		}
	}
	if w.channelCount(p.Channel) >= ChannelMaxPop {
		p.Channel = w.leastChannel()
	}
	if w.channelCount(p.Channel) >= ChannelMaxPop {
		return false, false
	}
	w.players[p.ID] = p
	return true, true
}

func (w *WorldState) Remove(p *Player) *Player {
	w.mu.Lock()
	defer w.mu.Unlock()
	cur := w.players[p.ID]
	if cur != p {
		return nil
	}
	w.persist(cur)
	w.persistGear(cur)
	w.cancelTradeOf(cur.ID)
	w.phase26MarkPartyOffline(cur.ID)
	cur.ensureLog().PresenceMode = "OFFLINE"
	if cur.InstanceID != "" && w.Dungeons != nil {
		if inst := w.Dungeons.instances[cur.InstanceID]; inst != nil {
			inst.OfflineSince[cur.ID] = time.Now()
			w.Dungeons.pending[cur.ID] = &rejoinSlot{InstanceID: cur.InstanceID, X: cur.X, Y: cur.Y, Z: cur.Z, Until: time.Now().Add(DungeonRejoinTimeout)}
		}
	}
	if cur.InstanceID != "" && w.PvP != nil {
		if inst := w.PvP.instances[cur.InstanceID]; inst != nil {
			inst.Offline[cur.ID] = time.Now()
			w.PvP.pending[cur.ID] = &rejoinSlot{InstanceID: cur.InstanceID, X: cur.X, Y: cur.Y, Z: cur.Z, Until: time.Now().Add(pvpReconnectWindow())}
		}
	}
	delete(w.players, p.ID)
	return p
}

func (w *WorldState) Get(id string) *Player {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.players[id]
}

func (w *WorldState) Simulate(dt float64) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	var events [][]byte
	if w.Time.Tick(now) {
		events = append(events, marshal(TypeWeatherUpdated, map[string]any{"phase": w.Time.Phase, "weather": w.Time.Weather, "clock": w.Time.ClockText(), "label": w.Time.ClockLabel()}))
	}
	w.tickNPCs(dt)
	events = append(events, w.tickWorldEvents(now)...)
	events = append(events, w.tickWorldBoss(now)...)
	for _, p := range w.players {
		if !p.Connected {
			continue
		}
		p.Simulate(dt)
		if !p.SlowUntil.IsZero() && now.Before(p.SlowUntil) && p.SlowFactor > 0 && p.SlowFactor < 1 {
			p.VX *= p.SlowFactor
			p.VZ *= p.SlowFactor
		}
		if p.InstanceID != "" {
			lim := DungeonLimit
			if w.pvpOf(p.ID) != nil {
				lim = w.pvpLimit(p)
				w.validatePvpMove(p, dt)
			}
			p.X = clamp(p.X, -lim, lim)
			p.Z = clamp(p.Z, -lim, lim)
			continue
		}
		maxZ := p.maxExploreZ()
		if p.Z > maxZ {
			p.Z = maxZ
			if p.VZ > 0 {
				p.VZ = 0
			}
		}
		events = append(events, w.tickExplore(p)...)
		events = append(events, w.tickTravel(p)...)
		events = append(events, w.tickJourney(p)...)
		events = append(events, w.autoDiscoverNearby(p)...)
		if p.InstanceID == "" {
			p.X, p.Z = resolveWorldXZ(p.X, p.Z, PlayerRadius)
			p.X, p.Z = resolvePlayerNPCPush(p.X, p.Z, PlayerRadius, w.npcPos, activeNPCList())
		}
	}
	events = append(events, w.tickCombat(dt)...)
	w.tickDungeons(dt, now)
	events = append(events, w.tickPvp(now)...)
	return events
}

func (w *WorldState) Snapshot() WorldSnapshot {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.snapshotLocked("")
}

func (w *WorldState) SnapshotFor(id string) WorldSnapshot {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.snapshotLocked(id)
}

func (w *WorldState) snapshotLocked(watcher string) WorldSnapshot {
	watchInst := ""
	wx, wz := 0.0, 0.0
	party := map[string]bool{}
	var watchPlayer *Player
	if watcher != "" {
		if p := w.players[watcher]; p != nil {
			watchInst = p.InstanceID
			wx, wz = p.X, p.Z
			watchPlayer = p
			if p.PartyID != "" && w.Parties != nil {
				if pt := w.Parties.Of(p.ID); pt != nil {
					for _, id := range pt.Members {
						party[id] = true
					}
				}
			}
		}
	}
	radius := worldCfg.InterestRadius
	if radius < 8 {
		radius = 52
	}
	see := func(x, z float64, id string) bool {
		if watcher == "" || id == watcher || party[id] {
			return true
		}
		return nearby(wx, wz, x, z, radius)
	}
	players := make([]PlayerSnapshot, 0, len(w.players))
	online := 0
	for _, p := range w.players {
		if !p.Connected {
			continue
		}
		online++
		if p.InstanceID != watchInst {
			continue
		}
		if watchPlayer != nil && p.channelID() != watchPlayer.channelID() && !party[p.ID] {
			continue
		}
		if !see(p.X, p.Z, p.ID) {
			continue
		}
		snap := p.Snap()
		if watcher != "" && p.ID != watcher && !party[p.ID] {
			d2 := hypot2(p.X, p.Z, wx, wz)
			if d2 > 18*18 {
				snap.CombatState = ""
			}
			if d2 > 32*32 {
				if p.VX*p.VX+p.VZ*p.VZ > 0.25 {
					snap.State = "WALK"
				} else {
					snap.State = "IDLE"
				}
			}
		}
		players = append(players, snap)
	}
	var enemies []EnemySnapshot
	var npcs []NPCSnapshot
	var objects []ObjectSnapshot
	var dungeon *DungeonView
	var pvpView *PvpView
	if watchInst != "" && w.Dungeons != nil {
		inst := w.Dungeons.instances[watchInst]
		if inst != nil {
			enemies = make([]EnemySnapshot, 0, len(inst.Enemies))
			for _, e := range inst.Enemies {
				enemies = append(enemies, e.Snap())
			}
			objects = append(objects, inst.Objects...)
			dungeon = new(DungeonView)
			*dungeon = w.dungeonView(inst, watcher)
		}
	}
	if watchInst != "" && w.PvP != nil {
		if inst := w.PvP.instances[watchInst]; inst != nil {
			v := w.pvpView(inst, watcher)
			pvpView = &v
			for _, s := range inst.Shrines {
				objects = append(objects, ObjectSnapshot{ID: "shrine-" + s.ID, Kind: "shrine", X: s.X, Z: s.Z, Text: s.ID})
			}
		}
	}
	if watchInst == "" {
		enemies = make([]EnemySnapshot, 0, len(w.enemies))
		for _, e := range w.enemies {
			if watcher != "" && !nearby(wx, wz, e.X, e.Z, radius) {
				continue
			}
			enemies = append(enemies, e.Snap())
		}
		npcs = make([]NPCSnapshot, 0, len(npcCatalog)+len(crowdNPCExtra))
		eventOn := w.Events.Active != nil && (w.Events.Active.State == "ACTIVE" || w.Events.Active.State == "FINAL")
		for _, n := range activeNPCList() {
			snap := npcSnap(n)
			snap.X, snap.Z = w.npcLive(n)
			snap.Yaw = w.npcLiveYaw(n)
			snap.Activity = npcActivity(n, w.Time, eventOn)
			if watcher != "" && !adventureNPCVisible(n, wx, wz) {
				continue
			}
			npcs = append(npcs, snap)
		}
		objects = make([]ObjectSnapshot, 0, len(interactCatalog)+len(landmarkCatalog))
		for _, o := range interactCatalog {
			if watcher != "" && !nearby(wx, wz, o.X, o.Z, radius) {
				continue
			}
			objects = append(objects, objectSnap(o))
		}
		objects = append(objects, exploreObjects(wx, wz, radius, watcher != "")...)
	}
	zoneID := ""
	if watchPlayer != nil {
		zoneID = watchPlayer.ZoneID
	}
	worldName := worldCfg.Name
	if worldName == "" {
		worldName = "World of Dawn"
	}
	return WorldSnapshot{
		WorldID: w.WorldID, Channel: w.ChannelID, T: time.Now().UnixMilli(), Online: online,
		Players: players, NPCs: npcs, Enemies: enemies, Objects: objects,
		Drops: w.dropSnapsFor(watchInst), TimeOfDay: w.Time.Phase, Weather: w.weatherForWatcher(watchPlayer),
		ZoneID: zoneID, Event: w.eventViewFor(watchPlayer), WorldBoss: w.worldBossViewFor(watchPlayer), InstanceID: watchInst, Dungeon: dungeon, Pvp: pvpView,
		Clock: w.Time.ClockText(), ClockLabel: w.Time.ClockLabel(), WorldName: worldName,
	}
}

func (w *WorldState) SpawnsExcept(id string) []PlayerSpawn {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]PlayerSpawn, 0, len(w.players))
	selfInst := ""
	if s := w.players[id]; s != nil {
		selfInst = s.InstanceID
	}
	for _, p := range w.players {
		if p.ID == id || !p.Connected || p.InstanceID != selfInst {
			continue
		}
		out = append(out, p.Spawn())
	}
	return out
}

func (w *WorldState) Broadcast(except string, payload []byte) {
	w.mu.Lock()
	list := make([]*Player, 0, len(w.players))
	for _, p := range w.players {
		if p.ID != except && p.Connected {
			list = append(list, p)
		}
	}
	w.mu.Unlock()
	for _, p := range list {
		select {
		case p.send <- payload:
		default:
		}
	}
}

func (w *WorldState) BroadcastAll(payload []byte) {
	w.Broadcast("", payload)
}

func (w *WorldState) BroadcastInstance(instanceID string, payload []byte) {
	w.mu.Lock()
	list := make([]*Player, 0, 4)
	if w.Dungeons != nil {
		if inst := w.Dungeons.instances[instanceID]; inst != nil {
			for _, id := range inst.Players {
				if p := w.players[id]; p != nil && p.Connected {
					list = append(list, p)
				}
			}
		}
	}
	if w.PvP != nil {
		if inst := w.PvP.instances[instanceID]; inst != nil {
			for _, id := range inst.Players {
				if p := w.players[id]; p != nil && p.Connected {
					list = append(list, p)
				}
			}
		}
	}
	w.mu.Unlock()
	for _, p := range list {
		select {
		case p.send <- payload:
		default:
		}
	}
}

func (w *WorldState) BroadcastScope(fromID string, payload []byte) {
	w.mu.Lock()
	inst := ""
	if p := w.players[fromID]; p != nil {
		inst = p.InstanceID
	}
	list := make([]*Player, 0, len(w.players))
	for _, p := range w.players {
		if p.Connected && p.InstanceID == inst {
			list = append(list, p)
		}
	}
	w.mu.Unlock()
	for _, p := range list {
		select {
		case p.send <- payload:
		default:
		}
	}
}

func (w *WorldState) SendTo(id string, payload []byte) {
	if payload == nil {
		return
	}
	w.mu.Lock()
	p := w.players[id]
	w.mu.Unlock()
	if p == nil || !p.Connected {
		return
	}
	select {
	case p.send <- payload:
	default:
	}
}

func (w *WorldState) ProgressFor(id string) PlayerProgressOut {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil {
		return PlayerProgressOut{}
	}
	out := p.progressOut(w.Time.Phase)
	out.Weather = w.Time.Weather
	out.ZoneID = p.ZoneID
	return out
}

func (w *WorldState) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := 0
	for _, p := range w.players {
		if p.Connected {
			n++
		}
	}
	return n
}

func (w *WorldState) DropTimedOut() []*Player {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	var out []*Player
	for id, p := range w.players {
		if p.Connected && now.Sub(p.LastHeard) > HeartbeatTimeout {
			w.persist(p)
			w.persistGear(p)
			w.cancelTradeOf(p.ID)
			if p.InstanceID != "" && w.Dungeons != nil {
				if inst := w.Dungeons.instances[p.InstanceID]; inst != nil {
					inst.OfflineSince[p.ID] = now
					w.Dungeons.pending[p.ID] = &rejoinSlot{InstanceID: p.InstanceID, X: p.X, Y: p.Y, Z: p.Z, Until: now.Add(DungeonRejoinTimeout)}
				}
			}
			if p.InstanceID != "" && w.PvP != nil {
				if inst := w.PvP.instances[p.InstanceID]; inst != nil {
					inst.Offline[p.ID] = now
					w.PvP.pending[p.ID] = &rejoinSlot{InstanceID: p.InstanceID, X: p.X, Y: p.Y, Z: p.Z, Until: now.Add(pvpReconnectWindow())}
				}
			}
			p.Connected = false
			delete(w.players, id)
			out = append(out, p)
		}
	}
	return out
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

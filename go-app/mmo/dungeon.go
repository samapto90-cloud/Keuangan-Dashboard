package mmo

import (
	"math"
	"strings"
	"time"
)

const (
	DungeonLimit         = 28.0
	DungeonReadyTimeout  = 20 * time.Second
	DungeonRejoinTimeout = 5 * time.Minute
	DungeonMaxPlayers    = 10
	DungeonRespawnDelay  = 5 * time.Second
	DungeonDownedWindow  = 10 * time.Second
	DungeonWipePenalty   = 45 * time.Second
	ReviveChannelTime    = 3.0
	QueueFillWait        = 8 * time.Second
)

const (
	DunWaiting    = "WAITING"
	DunStarting   = "STARTING"
	DunLoading    = "LOADING"
	DunActive     = "ACTIVE"
	DunBoss       = "BOSS"
	DunCompleted  = "COMPLETED"
	DunFailed     = "FAILED"
	DunAbandoned  = "ABANDONED"
	DunClosing    = "CLOSING"
)

const (
	QueueIdle     = "IDLE"
	QueueQueued   = "QUEUED"
	QueueMatching = "MATCHING"
	QueueReady    = "READY"
	QueueMatched  = "MATCHED"
	QueueCancel   = "CANCELLED"
)

type DungeonParticipant struct {
	PlayerID string
	Ready    bool
	Offline  bool
}

type ReadyCheck struct {
	DungeonID  string
	LeaderID   string
	Members    map[string]bool
	Until      time.Time
	FromQueue  bool
	PartyID    string
	Difficulty string
}

type QueueEntry struct {
	PlayerID   string
	Role       string
	DungeonID  string
	Difficulty string
	Region     string
	PartyID    string
	JoinedAt   time.Time
	State      string
}

type DungeonQueue struct {
	Entries []QueueEntry
}

type DungeonHub struct {
	instances     map[string]*DungeonInstance
	byPlayer      map[string]string
	ready         map[string]*ReadyCheck
	pending       map[string]*rejoinSlot
	Queue         *DungeonQueue
	PenaltyUntil  map[string]time.Time
	QueueStrikes  map[string]int
	LastQueueJoin map[string]time.Time
}

type rejoinSlot struct {
	InstanceID string
	X, Y, Z    float64
	Until      time.Time
}

type DungeonInstance struct {
	ID, DefID, ChapterID, PartyID string
	Kind                          string
	Players                       []string
	CreatedAt, StartedAt, EndsAt  time.Time
	State                         string
	WaveIndex                     int
	ObjIndex                      int
	ObjProgress, ObjNeed          int
	EncounterIndex                int
	Enemies                       map[string]*Enemy
	BossID                        string
	BossPhase                     int
	BossLocked                    bool
	Enraged                       bool
	EnrageAt                      time.Time
	SkillCD                       map[string]time.Time
	Telegraph                     *BossTelegraph
	CheckpointX, CheckpointZ      float64
	CheckpointWave                int
	Deaths                        map[string]int
	Claims                        map[string]bool
	RewardClaimID                 string
	Loot                          map[string][]LootItemView
	Rating                        string
	ChestReady                    bool
	Objects                       []ObjectSnapshot
	Return                        map[string][3]float64
	OfflineSince                  map[string]time.Time
	ReviveToken                   map[string]int
	DownedAt                      map[string]time.Time
	ReviveProgress                map[string]float64
	Reviving                      map[string]string
	Threat                        map[string]float64
	TauntUntil                    map[string]time.Time
	Votes                         map[string]string
	WipeCount                     int
	PuzzleStep                    int
	PuzzleNeed                    []string
	CrystalShield                 bool
	SkipIntro                     map[string]bool
	DamageDealt                   map[string]int
	Suspicious                    map[string]int
	LastHitID                     string
	LastHitAt                     time.Time
	Roles                         map[string]string
	Modifiers                     []string
	HorizonLevel                  int
	HorizonScore                  int
	Difficulty                    string
	EduShield                     bool
	EduQuestion                   string
	PhaseLockUntil                time.Time
	Mechanic                      string
	MechanicUntil                 time.Time
	GuideHP                       int
	GuideAlive                    bool
	CrystalOrder                  []string
}

type BossTelegraph struct {
	SkillID       string
	X, Z          float64
	Radius        float64
	Until         time.Time
	Damage        int
	Shape         string
	Yaw           float64
	Interruptible bool
}

func NewDungeonHub() *DungeonHub {
	return &DungeonHub{
		instances: map[string]*DungeonInstance{},
		byPlayer:  map[string]string{},
		ready:     map[string]*ReadyCheck{},
		pending:       map[string]*rejoinSlot{},
		Queue:         &DungeonQueue{},
		PenaltyUntil:  map[string]time.Time{},
		QueueStrikes:  map[string]int{},
		LastQueueJoin: map[string]time.Time{},
	}
}

func (w *WorldState) dungeonOf(playerID string) *DungeonInstance {
	if w.Dungeons == nil {
		return nil
	}
	return w.Dungeons.instances[w.Dungeons.byPlayer[playerID]]
}

func (w *WorldState) ApplyDungeon(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	if w.Dungeons == nil {
		w.Dungeons = NewDungeonHub()
	}
	var in DungeonActionIn
	_ = unmarshal(env.Data, &in)
	switch env.Type {
	case TypeUnlockChapter, TypeSetBossHP, TypeSetChapter, TypeSetObjective, TypeGiveLoot, TypeObjectiveComplete, TypeSetBossDead, TypeSpawnBoss, TypeCompleteDungeon, TypeSkipMechanic, TypeDamageBoss:
		return rejectFor(p.ID, env.Type, "server_authoritative")
	case TypeGetChapters:
		return [][]byte{marshal(TypeChapterList, w.chapterList(p))}
	case TypeGetDungeons, TypeGetDungeonHistory:
		return [][]byte{marshal(TypeDungeonList, w.dungeonList(p))}
	case TypeRaidExchange:
		return w.raidTokenExchange(p, in.ShopItemID)
	case TypeQueueJoin:
		return w.queueJoin(p, in)
	case TypeQueueLeave:
		return w.queueLeave(p)
	case TypeDungeonEnter:
		return w.dungeonEnter(p, in.DungeonID, in.Difficulty)
	case TypeDungeonReady:
		return w.dungeonReady(p, in.Ready)
	case TypeDungeonLeave:
		return w.dungeonLeave(p, false)
	case TypeDungeonAbandon:
		return w.dungeonAbandon(p)
	case TypeDungeonRetry:
		return w.dungeonRetry(p)
	case TypeDungeonJoin:
		return w.dungeonJoin(p, in.InstanceID)
	case TypeDungeonFill:
		return w.dungeonFill(p, in.InstanceID)
	case TypeDungeonRevive:
		return w.dungeonRevive(p, in.TargetID)
	case TypeDungeonVote:
		return w.dungeonVote(p, in.Vote)
	case TypeDungeonTaunt:
		return w.dungeonTaunt(p)
	case TypeSkipDungeonIntro:
		return w.skipDungeonIntro(p)
	case TypeClaimLoot:
		return w.claimDungeonLoot(p, in.ClaimID)
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) offerDungeon(p *Player, dungeonID string) [][]byte {
	def, ok := dungeonByID[dungeonID]
	if !ok {
		dungeonID = "dun-ch01"
		def = dungeonByID[dungeonID]
	}
	ch := chapterByID[def.ChapterID]
	status := chapterStatus(p, ch)
	if status == "LOCKED" {
		return [][]byte{marshal(TypeInteractResult, InteractResult{
			Kind: "dungeon", TargetID: "dungeon-portal-1", Title: def.Name, Speaker: "Petualangan",
			Text: "Chapter ini masih terkunci.", Locked: true,
			Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		})}
	}
	speaker := ch.Title
	if speaker == "" {
		speaker = "Petualangan"
	}
	text := def.Name + "\nLevel disarankan: " + itoa(def.RecommendedLevel) + "\nPemain: " + itoa(dungeonMinPlayers(def)) + "–" + itoa(def.MaxPlayers) + "\nBatas waktu: " + itoa(def.TimeLimit/60) + " menit\nHadiah: EXP " + itoa(def.Rewards.Exp) + " · Koin " + itoa(def.Rewards.Coin)
	kind := dungeonKind(def)
	return [][]byte{
		marshal(TypeDungeonOffer, DungeonOffer{
			DungeonID: def.ID, Name: def.Name, ChapterID: def.ChapterID, Kind: kind, Description: def.Description,
			Difficulty: dungeonDifficulty(def), RecommendedLevel: def.RecommendedLevel,
			MinPlayers: dungeonMinPlayers(def), MaxPlayers: def.MaxPlayers,
			TimeLimit: def.TimeLimit, Rewards: rewardView(def.Rewards, false),
			Status: status,
		}),
		marshal(TypeDungeonList, w.dungeonList(p)),
		marshal(TypeInteractResult, InteractResult{
			Kind: "dungeon", TargetID: "dungeon-portal-1", Title: "ENTER DUNGEON", Speaker: speaker,
			Text: text, Options: []DialogOption{
				{ID: "dungeon:" + def.ID, Label: "ENTER DUNGEON"},
				{ID: "queue:" + def.ID, Label: "QUEUE MATCHMAKING"},
				{ID: "close", Label: "Tutup"},
			},
		}),
	}
}

func (w *WorldState) dungeonEnter(p *Player, dungeonID, difficulty string) [][]byte {
	if strings.EqualFold(dungeonID, "RANDOM") {
		picked, ok := w.pickRandomDungeon(p, normalizeDungeonDiff(difficulty))
		if !ok {
			return rejectFor(p.ID, TypeDungeonEnter, "dungeon")
		}
		dungeonID = picked.ID
	}
	if dungeonID == "" {
		dungeonID = "dun-ch01"
	}
	def, ok := dungeonByID[dungeonID]
	if !ok {
		return rejectFor(p.ID, TypeDungeonEnter, "dungeon")
	}
	if p.InstanceID != "" {
		return rejectFor(p.ID, TypeDungeonEnter, "already")
	}
	if p.Level < def.MinimumLevel {
		return rejectFor(p.ID, TypeDungeonEnter, "level")
	}
	if def.ID == horizonCfg.DungeonID && !p.endgameUnlocked() {
		return rejectFor(p.ID, TypeDungeonEnter, "endgame")
	}
	if w.raidLocked(p, def) {
		return rejectFor(p.ID, TypeDungeonEnter, "lockout")
	}
	ch := chapterByID[def.ChapterID]
	if def.ChapterID != "" && chapterStatus(p, ch) == "LOCKED" {
		return rejectFor(p.ID, TypeDungeonEnter, "locked")
	}
	members := []string{p.ID}
	leader := p.ID
	if pt := w.Parties.Of(p.ID); pt != nil {
		if pt.LeaderID != p.ID {
			return rejectFor(p.ID, TypeDungeonEnter, "not_leader")
		}
		members = append([]string{}, pt.Members...)
		leader = pt.LeaderID
	}
	if len(members) > def.MaxPlayers {
		return rejectFor(p.ID, TypeDungeonEnter, "full")
	}
	if len(members) < dungeonMinPlayers(def) {
		return rejectFor(p.ID, TypeDungeonEnter, "size")
	}
	for _, id := range members {
		m := w.players[id]
		if m == nil || !m.Connected {
			return rejectFor(p.ID, TypeDungeonEnter, "offline")
		}
		if m.InstanceID != "" {
			return rejectFor(p.ID, TypeDungeonEnter, "busy")
		}
		if m.Level < def.MinimumLevel {
			return rejectFor(p.ID, TypeDungeonEnter, "level")
		}
		if w.raidLocked(m, def) {
			return rejectFor(p.ID, TypeDungeonEnter, "lockout")
		}
	}
	diff := normalizeDungeonDiff(difficulty)
	if difficulty != "" && !prototypeDiffOK(diff) {
		return rejectFor(p.ID, TypeDungeonEnter, "difficulty")
	}
	if len(members) == 1 {
		return w.startDungeon(def, members, p.PartyID, diff)
	}
	ready := &ReadyCheck{DungeonID: def.ID, LeaderID: leader, Members: map[string]bool{leader: true}, Until: time.Now().Add(DungeonReadyTimeout), Difficulty: diff}
	w.Dungeons.ready[leader] = ready
	for _, id := range members {
		if id != leader {
			ready.Members[id] = false
		}
	}
	return [][]byte{marshal(TypeDungeonReadyCheck, DungeonReadyOut{
		DungeonID: def.ID, LeaderID: leader, Until: ready.Until.UnixMilli(), Members: readyMembers(members, ready.Members),
	})}
}

func (w *WorldState) dungeonReady(p *Player, ready bool) [][]byte {
	var check *ReadyCheck
	for _, c := range w.Dungeons.ready {
		if c.LeaderID == p.ID {
			check = c
			break
		}
		if _, ok := c.Members[p.ID]; ok {
			check = c
			break
		}
		if pt := w.Parties.Of(c.LeaderID); pt != nil && containsID(pt.Members, p.ID) {
			check = c
			break
		}
	}
	if check == nil {
		return rejectFor(p.ID, TypeDungeonReady, "check")
	}
	if time.Now().After(check.Until) {
		delete(w.Dungeons.ready, check.LeaderID)
		out := [][]byte{marshal(TypeDungeonReadyCheck, DungeonReadyOut{DungeonID: check.DungeonID, Cancelled: true, FromQueue: check.FromQueue})}
		if check.FromQueue {
			out = append(out, w.requeueMembers(check)...)
		}
		return out
	}
	check.Members[p.ID] = ready
	if !ready && check.FromQueue {
		delete(w.Dungeons.ready, check.LeaderID)
		out := [][]byte{marshal(TypeDungeonReadyCheck, DungeonReadyOut{DungeonID: check.DungeonID, Cancelled: true, FromQueue: true})}
		return append(out, w.requeueMembers(check)...)
	}
	members := readyMemberIDs(check)
	if !check.FromQueue {
		if pt := w.Parties.Of(check.LeaderID); pt != nil {
			members = append([]string{}, pt.Members...)
		}
	}
	out := [][]byte{marshal(TypeDungeonReadyCheck, DungeonReadyOut{
		DungeonID: check.DungeonID, LeaderID: check.LeaderID, Until: check.Until.UnixMilli(), Members: readyMembers(members, check.Members), FromQueue: check.FromQueue,
	})}
	all := true
	for _, id := range members {
		if !check.Members[id] {
			all = false
			break
		}
	}
	if !all {
		return out
	}
	delete(w.Dungeons.ready, check.LeaderID)
	def := dungeonByID[check.DungeonID]
	partyID := check.PartyID
	if partyID == "" {
		if pt := w.Parties.Of(check.LeaderID); pt != nil {
			partyID = pt.ID
		}
	}
	return append(out, w.startDungeon(def, members, partyID, check.Difficulty)...)
}

func (w *WorldState) startDungeon(def DungeonDef, members []string, partyID, difficulty string) [][]byte {
	diff := normalizeDungeonDiff(difficulty)
	if !prototypeDiffOK(diff) {
		diff = dungeonDifficulty(def)
	}
	limit := hardTimeLimit(def, diff)
	inst := &DungeonInstance{
		ID: randomID("dun_"), DefID: def.ID, ChapterID: def.ChapterID, PartyID: partyID, Kind: dungeonKind(def),
		Players: append([]string{}, members...), CreatedAt: time.Now(), StartedAt: time.Now(),
		EndsAt: time.Now().Add(time.Duration(limit) * time.Second),
		State:  DunStarting, Enemies: map[string]*Enemy{}, SkillCD: map[string]time.Time{},
		CheckpointX: 0, CheckpointZ: 2, Deaths: map[string]int{}, Claims: map[string]bool{},
		Loot: map[string][]LootItemView{}, Return: map[string][3]float64{}, OfflineSince: map[string]time.Time{},
		BossPhase: 1, EnrageAt: time.Now().Add(RaidEnrageDefault),
		ReviveToken: map[string]int{}, DownedAt: map[string]time.Time{}, ReviveProgress: map[string]float64{},
		Reviving: map[string]string{}, Threat: map[string]float64{}, TauntUntil: map[string]time.Time{},
		Votes: map[string]string{}, SkipIntro: map[string]bool{}, DamageDealt: map[string]int{},
		Suspicious: map[string]int{}, Roles: map[string]string{}, PuzzleNeed: []string{"wind", "stone", "light"},
		Objects: []ObjectSnapshot{
			{ID: "dungeon-checkpoint", Kind: "checkpoint", X: 0, Z: 2},
		},
		RewardClaimID: "", Difficulty: diff,
	}
	if b := bossByID[def.BossID]; b.EnrageTime > 0 {
		inst.EnrageAt = time.Now().Add(time.Duration(b.EnrageTime) * time.Second)
	}
	w.Dungeons.instances[inst.ID] = inst
	spots := [][2]float64{{0, 2}, {1.5, 2.2}, {-1.5, 2.2}, {0, 3.4}, {2.2, 3}, {-2.2, 3}, {1, 4}, {-1, 4}, {2.4, 4.2}, {-2.4, 4.2}}
	var out [][]byte
	for i, id := range members {
		m := w.players[id]
		if m == nil {
			continue
		}
		inst.Return[id] = [3]float64{m.X, m.Y, m.Z}
		m.InstanceID = inst.ID
		w.Dungeons.byPlayer[id] = inst.ID
		sp := spots[i%len(spots)]
		m.X, m.Y, m.Z = sp[0], 0, sp[1]
		m.VX, m.VZ = 0, 0
		inst.ReviveToken[id] = 1
		if pt := w.Parties.Of(id); pt != nil && pt.Roles != nil {
			inst.Roles[id] = pt.Roles[id]
		}
		if inst.Roles[id] == "" {
			inst.Roles[id] = "FLEX"
		}
		if m.alive() {
			m.HP = m.MaxHP
		}
		out = append(out, marshal(TypeDungeonLoading, DungeonLoading{InstanceID: inst.ID, DungeonID: def.ID, Name: def.Name, ToID: id}))
		if def.EducationQuestion != "" && !m.ensureLog().Flags["edu_gate_"+def.EducationQuestion] {
			out = append(out, w.offerDungeonQuiz(m, def.EducationQuestion)...)
		}
	}
	w.applyHorizonStart(inst, def)
	w.seedDungeonMechanics(inst)
	w.seedRaidGuide(inst)
	if dungeonKind(def) != "RAID" {
		w.spawnDungeonWave(inst, 1)
	}
	w.setObjective(inst, 0)
	inst.State = DunActive
	inst.CheckpointWave = inst.WaveIndex
	view := w.dungeonView(inst, members[0])
	view.InstanceID = inst.ID
	out = append(out, marshal(TypeDungeonStarted, view), marshal(TypeDungeonState, view))
	return out
}

func (w *WorldState) dungeonLeave(p *Player, kicked bool) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypeDungeonLeave, "not_in_dungeon")
	}
	instID := inst.ID
	w.extractPlayer(inst, p, true)
	out := [][]byte{marshal(TypeDungeonLeft, map[string]any{"playerId": p.ID, "kicked": kicked, "instanceId": instID, "toId": p.ID})}
	if len(inst.Players) == 0 {
		inst.State = DunClosing
		delete(w.Dungeons.instances, inst.ID)
		return out
	}
	return append(out, marshal(TypeDungeonState, w.dungeonView(inst, inst.Players[0])))
}

func (w *WorldState) dungeonAbandon(p *Player) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypeDungeonAbandon, "not_in_dungeon")
	}
	if inst.PartyID != "" {
		pt := w.Parties.Get(inst.PartyID)
		if pt != nil && pt.LeaderID != p.ID {
			return rejectFor(p.ID, TypeDungeonAbandon, "not_leader")
		}
	}
	return w.closeDungeon(inst, DunAbandoned, "Dungeon ditutup.")
}

func (w *WorldState) dungeonRetry(p *Player) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil || (inst.State != DunFailed && inst.State != DunCompleted) {
		return rejectFor(p.ID, TypeDungeonRetry, "state")
	}
	def := dungeonByID[inst.DefID]
	members := append([]string{}, inst.Players...)
	w.closeDungeon(inst, DunClosing, "")
	return w.startDungeon(def, members, inst.PartyID, inst.Difficulty)
}

func (w *WorldState) extractPlayer(inst *DungeonInstance, p *Player, toWorld bool) {
	inst.Players = removeID(inst.Players, p.ID)
	delete(w.Dungeons.byPlayer, p.ID)
	p.InstanceID = ""
	if toWorld {
		if ret, ok := inst.Return[p.ID]; ok {
			p.X, p.Y, p.Z = ret[0], ret[1], ret[2]
		} else {
			p.X, p.Y, p.Z = 0, 0, 24.6
		}
		if p.HP <= 0 {
			p.HP = p.MaxHP
			p.CombatState = "IDLE"
		}
	}
}

func (w *WorldState) closeDungeon(inst *DungeonInstance, state, toast string) [][]byte {
	inst.State = state
	var out [][]byte
	ids := append([]string{}, inst.Players...)
	for _, id := range ids {
		if m := w.players[id]; m != nil {
			w.extractPlayer(inst, m, true)
			if toast != "" {
				out = append(out, marshal(TypeSocialNotification, SocialNote{Kind: "dungeon", Text: toast, ToID: id}))
			}
			out = append(out, marshal(TypeDungeonLeft, map[string]any{"playerId": id, "instanceId": inst.ID, "toId": id}))
		}
	}
	delete(w.Dungeons.instances, inst.ID)
	return out
}

func (w *WorldState) SendToLocked(id string, payload []byte) {
	p := w.players[id]
	if p == nil || !p.Connected || payload == nil {
		return
	}
	select {
	case p.send <- payload:
	default:
	}
}

func readyMembers(ids []string, ready map[string]bool) []DungeonReadyMember {
	out := make([]DungeonReadyMember, 0, len(ids))
	for _, id := range ids {
		out = append(out, DungeonReadyMember{PlayerID: id, Ready: ready[id]})
	}
	return out
}

func containsID(list []string, id string) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

func (w *WorldState) sameInstance(a, b string) bool {
	if w.Dungeons == nil {
		return a == "" && b == ""
	}
	return w.Dungeons.byPlayer[a] == w.Dungeons.byPlayer[b]
}

func dist2(ax, az, bx, bz float64) float64 {
	return math.Hypot(ax-bx, az-bz)
}

func (w *WorldState) offerDungeonQuiz(p *Player, questionID string) [][]byte {
	if p == nil || questionID == "" {
		return nil
	}
	log := p.ensureLog()
	if log.Quiz.Active {
		return nil
	}
	idx := -1
	for i, q := range questionCatalog {
		if q.ID == questionID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil
	}
	log.Quiz = QuizSession{QuestID: "dungeon", Index: idx, Active: true}
	out := questionOut(idx)
	if out == nil {
		return nil
	}
	out.Index = 1
	out.Total = 1
	out.ToID = p.ID
	return [][]byte{marshal(TypeEducationQuestion, out)}
}

func (w *WorldState) answerDungeonQuestion(p *Player, in EducationAnswerIn) [][]byte {
	log := p.ensureLog()
	def, ok := questionByID[in.QuestionID]
	if !ok {
		return rejectFor(p.ID, TypeEducationAnswer, "question")
	}
	if inst := w.dungeonOf(p.ID); inst != nil && inst.EduShield {
		want := inst.EduQuestion
		if want == "" {
			want = dungeonByID[inst.DefID].EducationBoss
		}
		if want != "" && in.QuestionID != want {
			return rejectFor(p.ID, TypeEducationAnswer, "question")
		}
		if in.Choice != def.Correct {
			q := questionOut(log.Quiz.Index)
			if q != nil {
				q.ToID = p.ID
			}
			return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
				Correct: false, Explain: def.Explain, Retry: true, Toast: "Perisai boss tetap aktif.", Question: q,
			}), marshal(TypeDungeonState, w.dungeonView(inst, p.ID))}
		}
		inst.EduShield = false
		inst.CrystalShield = false
		log.Quiz.Active = false
		log.Flags["edu_gate_"+in.QuestionID] = true
		log.LastQuestion = in.QuestionID
		p.markDirty()
		events := [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: true, Explain: def.Explain, Toast: "Perisai pecah. Serang boss.",
		}), marshal(TypeDungeonState, w.dungeonView(inst, p.ID))}
		events = append(events, w.giveExp(p, 8)...)
		w.giveCurrency(p, 2, 0)
		grantWhisperEducation(p, in.QuestionID)
		events = append(events, w.recordEducation(p, def.Category, true)...)
		return events
	}
	if log.Quiz.Index >= len(questionCatalog) || questionCatalog[log.Quiz.Index].ID != in.QuestionID {
		return rejectFor(p.ID, TypeEducationAnswer, "order")
	}
	if in.Choice != def.Correct {
		q := questionOut(log.Quiz.Index)
		if q != nil {
			q.ToID = p.ID
		}
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: false, Explain: def.Explain, Retry: true, Toast: "Coba lagi.", Question: q,
		})}
	}
	log.Quiz.Active = false
	p.ensureLog().Flags["edu_gate_"+in.QuestionID] = true
	log.LastQuestion = in.QuestionID
	p.markDirty()
	events := [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
		Correct: true, Explain: def.Explain, Toast: "Bonus dungeon: +8 EXP · +2 koin",
	})}
	events = append(events, w.giveExp(p, 8)...)
	w.giveCurrency(p, 2, 0)
	events = append(events, w.recordEducation(p, def.Category, true)...)
	return events
}

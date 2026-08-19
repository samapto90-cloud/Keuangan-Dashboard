package mmo

import (
	"math"
	"strings"
	"time"
)

func (w *WorldState) ApplyWorld(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	switch env.Type {
	case TypeCollectItem, TypeForestUnlock, TypeEducationCorrect:
		return rejectCheat(p.ID, env)
	case TypeInteract:
		var in InteractIn
		if unmarshal(env.Data, &in) != nil || in.TargetID == "" {
			return rejectFor(p.ID, TypeInteract, "payload")
		}
		return w.handleInteract(p, in.TargetID, in.Kind)
	case TypeQuestAccept:
		var in QuestActionIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypeQuestAccept, "payload")
		}
		return w.acceptQuest(p, in.QuestID)
	case TypeQuestDecline:
		var in QuestActionIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypeQuestDecline, "payload")
		}
		return w.declineQuest(p, in.QuestID)
	case TypeQuestClaim:
		var in QuestActionIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypeQuestClaim, "payload")
		}
		return w.claimQuest(p, in.QuestID)
	case TypeEducationAnswer:
		var in EducationAnswerIn
		if unmarshal(env.Data, &in) != nil {
			return rejectFor(p.ID, TypeEducationAnswer, "payload")
		}
		return w.answerQuestion(p, in)
	case TypeHeal:
		return w.healAtNara(p)
	case TypeShopOpen:
		return w.openShop(p)
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func rejectCheat(playerID string, env Envelope) [][]byte {
	if env.Type == TypeCollectItem {
		var in CollectItemIn
		_ = unmarshal(env.Data, &in)
		if in.Quantity > 1 {
			return rejectFor(playerID, TypeCollectItem, "quantity")
		}
	}
	return rejectFor(playerID, env.Type, "server_authoritative")
}

func (w *WorldState) handleInteract(p *Player, targetID, kind string) [][]byte {
	if p.InstanceID != "" {
		if d := w.drops[targetID]; d != nil {
			return w.pickupDrop(p, PickupIn{DropID: targetID})
		}
		return w.handleDungeonInteract(p, targetID, kind)
	}
	if npc, ok := npcByID[targetID]; ok {
		nx, nz := w.npcLive(npc)
		rangeMax := npc.InteractionRange
		if rangeMax < 2 {
			rangeMax = 2.6
		}
		if rangeMax > 4 {
			rangeMax = 4
		}
		if math.Hypot(p.X-nx, p.Z-nz) > rangeMax+0.5 {
			return rejectFor(p.ID, TypeInteract, "distance")
		}
		if kind == "quiz" {
			return w.startQuiz(p)
		}
		if kind == "quiz-apel" {
			return w.startStoryQuiz(p, "stq-edu-apel", "q-apel-3-2")
		}
		if kind == "quiz-latihan" {
			return w.startStoryQuiz(p, "stq-latihan", "q-add-6-2")
		}
		if kind == "quiz-prof" {
			return w.startProfessionQuiz(p)
		}
		if kind == "quiz-econ" {
			return w.startEconQuiz(p)
		}
		if kind == "quiz-craft" {
			return w.startCraftQuiz(p)
		}
		if kind == "repair" {
			return w.npcRepair(p, "")
		}
		if strings.HasPrefix(kind, "npc-shop:") {
			return w.npcShopOpen(p, strings.TrimPrefix(kind, "npc-shop:"))
		}
		if kind == "fish-ui" {
			return w.fishStart(p, "spot-village")
		}
		if strings.HasPrefix(kind, "choice") {
			if npc.ID == "mira" || npc.Type == "TEACHER" {
				return w.npcChoice(p, npc, kind)
			}
			if phase24NPC(npc.ID) {
				return w.phase24Choice(p, npc, kind)
			}
			return w.storyNpcChoice(p, npc, kind)
		}
		if kind == "quiz-guru" {
			return w.startPhase24Quiz(p, npc)
		}
		if kind == "quiz-combat" {
			return w.startPhase25Quiz(p, npc)
		}
		if strings.HasPrefix(kind, "cin:") {
			return w.startCinematic(p, strings.TrimPrefix(kind, "cin:"), true)
		}
		if kind == "forest" {
			return [][]byte{marshal(TypeInteractResult, InteractResult{
				Kind: "npc", TargetID: npc.ID, Speaker: npc.Name, Role: npc.Role, Title: npc.Name,
				Text:    lineOr("elder_forest", "Forest Fang terlihat di dekat gerbang. Latihlah dirimu, lalu bantu kami mengusir mereka."),
				Options: elderOptions(p, npc),
			})}
		}
		return w.talkNPC(p, npc)
	}
	if d := w.drops[targetID]; d != nil {
		return w.pickupDrop(p, PickupIn{DropID: targetID})
	}
	if obj, ok := interactByID[targetID]; ok {
		if math.Hypot(p.X-obj.X, p.Z-obj.Z) > 2.6 {
			return rejectFor(p.ID, TypeInteract, "distance")
		}
		switch obj.Kind {
		case "chest":
			return w.openChest(p, obj)
		case "crystal":
			return w.collectCrystal(p, obj)
		case "checkpoint":
			return w.useCheckpoint(p, obj)
		case "gate":
			return w.useGate(p, obj)
		case "sign":
			return [][]byte{marshal(TypeInteractResult, InteractResult{
				Kind: "sign", TargetID: obj.ID, Title: "Papan Desa", Speaker: "Village of Dawn",
				Text: obj.Text, Options: []DialogOption{{ID: "close", Label: "Tutup"}},
			})}
		case "dungeon":
			if obj.ID == "cave-whispers-door" {
				return w.offerDungeon(p, WhispersDungeonID)
			}
			if obj.ID == "gerbang-33-door" {
				return w.offerDungeon(p, GerbangRaidID)
			}
			return w.offerDungeon(p, "dun-ch01")
		case "edu-shrine":
			return w.startShrineQuiz(p, obj)
		case "training-quiz":
			return w.startTrainingQuiz(p, obj)
		case "travel-puzzle":
			return w.handleTravelPuzzle(p, obj)
		case "lever", "pressure-plate":
			return w.openTravelPath(p, obj)
		case "gather-wood", "gather-stone", "gather-ore", "gather-herb", "gather-fruit", "gather-fiber":
			return w.gatherNode(p, obj.ID, "")
		case "forge", "workbench", "cooking-fire", "alchemy":
			return w.stationInteract(p, obj)
		case "fishing-spot":
			return w.fishStart(p, obj.ID)
		}
	}
	if lm, ok := landmarkByID[targetID]; ok {
		if math.Hypot(p.X-lm.X, p.Z-lm.Z) > 3.2 {
			return rejectFor(p.ID, TypeInteract, "distance")
		}
		events := w.discoverLandmark(p, lm)
		kind := "landmark"
		if lm.Rest {
			kind = "checkpoint"
		}
		text := lm.LoreJv
		sub := lm.LoreID
		if text == "" {
			text = lm.Lore
		}
		if sub == "" {
			sub = lm.Lore
		}
		events = append(events, marshal(TypeInteractResult, InteractResult{
			Kind: kind, TargetID: lm.ID, Title: lm.Name, Speaker: lm.Name,
			Text: text, Subtitle: sub, Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		}))
		return events
	}
	if lore, ok := loreByID[targetID]; ok {
		if math.Hypot(p.X-lore.X, p.Z-lore.Z) > 3.2 {
			return rejectFor(p.ID, TypeInteract, "distance")
		}
		return append(w.discoverLore(p, lore), marshal(TypeInteractResult, InteractResult{
			Kind: "lore", TargetID: lore.ID, Title: lore.Title, Speaker: "Lore",
			Text: lore.Text, Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		}))
	}
	if events := w.trySecret(p, targetID); events != nil {
		return events
	}
	if p.InstanceID != "" {
		return w.handleDungeonInteract(p, targetID, kind)
	}
	return rejectFor(p.ID, TypeInteract, "target")
}

func (w *WorldState) handleDungeonInteract(p *Player, targetID, kind string) [][]byte {
	inst := w.dungeonOf(p.ID)
	if inst == nil {
		return rejectFor(p.ID, TypeInteract, "instance")
	}
	if targetID == "mist-gate" || targetID == "dungeon-gate" {
		if dist2(p.X, p.Z, 0, 16) > 2.8 {
			return rejectFor(p.ID, TypeInteract, "distance")
		}
		objs := dungeonObjByDun[inst.DefID]
		if inst.ObjIndex < len(objs) && objs[inst.ObjIndex].Type == "REACH" {
			return w.advanceObjective(inst)
		}
	}
	if targetID == "dungeon-chest" && inst.ChestReady {
		return w.claimDungeonLoot(p, inst.RewardClaimID)
	}
	if targetID == "dungeon-checkpoint" {
		inst.CheckpointX, inst.CheckpointZ = p.X, p.Z
		inst.CheckpointWave = inst.WaveIndex
		p.HP = p.MaxHP
		return [][]byte{marshal(TypeInteractResult, InteractResult{
			Kind: "checkpoint", TargetID: targetID, Title: "Checkpoint", Speaker: "Cahaya",
			Text: "Titik kembali dungeon tersimpan.", Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		})}
	}
	if strings.HasPrefix(targetID, "puzzle-") {
		part := strings.TrimPrefix(targetID, "puzzle-")
		need := inst.PuzzleNeed
		if inst.PuzzleStep < len(need) && need[inst.PuzzleStep] == part {
			inst.PuzzleStep++
			text := "Puzzle: " + part + " (" + itoa(inst.PuzzleStep) + "/" + itoa(len(need)) + ")"
			if inst.PuzzleStep >= len(need) {
				text = "Puzzle Wind / Stone / Light selesai."
			}
			return [][]byte{marshal(TypeDungeonState, w.dungeonView(inst, p.ID)), marshal(TypeInteractResult, InteractResult{
				Kind: "puzzle", TargetID: targetID, Title: "Puzzle", Speaker: "Cahaya", Text: text,
				Options: []DialogOption{{ID: "close", Label: "Tutup"}},
			})}
		}
		inst.PuzzleStep = 0
		return [][]byte{marshal(TypeInteractResult, InteractResult{
			Kind: "puzzle", TargetID: targetID, Title: "Puzzle", Speaker: "Cahaya",
			Text: "Urutan salah. Mulai lagi: Wind, Stone, Light.", Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		})}
	}
	if strings.HasPrefix(targetID, "crystal-") {
		ok := w.activateRaidCrystal(inst, targetID)
		text := "Urutan kristal salah. Ikuti petunjuk cahaya."
		if ok {
			left := len(inst.CrystalOrder)
			if left == 0 {
				text = "Semua kristal aktif. Mekanik selesai."
			} else {
				text = "Kristal benar. Sisa " + itoa(left) + "."
			}
		}
		return [][]byte{marshal(TypeDungeonState, w.dungeonView(inst, p.ID)), marshal(TypeInteractResult, InteractResult{
			Kind: "crystal", TargetID: targetID, Title: "Raid Crystal", Speaker: "Cahaya",
			Text: text, Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		})}
	}
	for _, id := range inst.Players {
		if id == targetID && id != p.ID {
			return w.dungeonRevive(p, id)
		}
	}
	_ = kind
	return rejectFor(p.ID, TypeInteract, "target")
}

func (w *WorldState) talkNPC(p *Player, npc NPCDef) [][]byte {
	p.ensureLog()
	p.credit("TALK", npc.ID, 1)
	if npc.Type == "MENTOR" {
		return w.talkMbahJagat(p, npc)
	}
	if npc.Type == "VILLAGER" {
		return w.talkStoryVillager(p, npc)
	}
	marker := npcMarker(p, npc)
	res := InteractResult{
		Kind: "npc", TargetID: npc.ID, Title: npc.Name, Speaker: npc.Name, Role: npc.Role,
		Marker: marker, Options: []DialogOption{{ID: "close", Label: "Tutup"}},
	}
	switch npc.Type {
	case "ELDER":
		if p.ensureLog().Flags["chapter1_complete"] {
			res.Text = lineOr("elder_mist_cleared", "Kabut telah menghilang.")
		} else {
			res.Text = lineOr("elder_offer", "Raka, jalan menuju hutan tidak lagi aman.")
		}
		res.Options = elderOptions(p, npc)
	case "TEACHER":
		res.Text = lineOr("mira_welcome", "Selamat datang di Dawn Village. Aku Mira, pemandu desa.")
		if w.Events.Active != nil && (w.Events.Active.State == "ACTIVE" || w.Events.Active.State == "FINAL") {
			res.Text = "Makhluk bayangan dekat! Ikuti penjaga ke alun-alun."
		}
		p.credit("TALK", npc.ID, 1)
		res.Options = miraOptions(p, npc)
	case "GUIDE":
		jv, id, _, _ := p.journeyHint()
		res.Text = jv
		res.Subtitle = id
		res.Options = []DialogOption{{ID: "close", Label: "Tutup"}}
	case "TRAINER":
		res.Text = "Skill perlu dilatih. Aku jelaskan, tapi kekuatan tidak diberikan gratis."
		res.Options = []DialogOption{{ID: "close", Label: "Mengerti"}}
	case "SCHOLAR":
		res.Text = lineOr("scholar_nur", "Tujuan perjalanan bukan hanya kekuatan.")
		res.Options = append(questOption(p, "mq020", "Aku siap merenung"), DialogOption{ID: "close", Label: "Tutup"})
	case "CHILD":
		res.Text = lineOr("child_lina", "Aku tersesat tadi. Terima kasih sudah mencari.")
		res.Options = []DialogOption{{ID: "close", Label: "Tutup"}}
	case "TRAVELER", "ARENA":
		res.Text = lineOr(npc.DialogueID, npc.Role+" menunggu di sini.")
		res.Options = []DialogOption{{ID: "close", Label: "Tutup"}}
	case "MERCHANT":
		if npc.ID == "dawn_merchant" {
			res.Text = lineOr("dawn_merchant", "Ramuan, bahan, dan perlengkapan perjalanan.")
		} else {
			res.Text = lineOr("lio_intro", "Kristal daganganku hilang di sekitar desa.")
		}
		if w.Time.ClockHour() < 8 || w.Time.ClockHour() >= 20 {
			res.Text = "Toko tutup malam. Kembali pagi, pukul 08:00–20:00."
		}
		res.Options = lioOptions(p, npc)
	case "GUILD":
		res.Text = lineOr("guild_clerk", "Selamat datang di Guild Hall Dawn City.")
		res.Options = []DialogOption{{ID: "guild", Label: "Guild Hall"}, {ID: "close", Label: "Tutup"}}
	case "GUARD":
		jv, id, _, _ := p.journeyHint()
		if p.ensureLog().Flags["chapter1_complete"] {
			res.Text = lineOr("raven_valley_open", "Jalan ke lembah kini terbuka.")
			res.Subtitle = id
		} else if p.ensureLog().ForestUnlocked {
			res.Text = jv
			res.Subtitle = id
		} else {
			res.Text = lineOr("raven_locked", "Gerbang ini terkunci. Selesaikan misi yang diperlukan.")
			res.Subtitle = "Ikuti jalan tanah ke utara setelah misi desa selesai."
		}
		res.Options = ravenOptions(p, npc)
	case "HEALER":
		res.Text = lineOr("nara_intro", "Istirahatlah di sini jika terluka.")
		res.Options = naraOptions(p, npc)
	case "QUEST_BOARD":
		if npc.ID == "party_board" {
			res.Title = "Party Board"
			res.Text = lineOr("party_board", "Cari party untuk Dungeon, Guardian, atau World Event.")
			res.Options = []DialogOption{{ID: "finder", Label: "Party Finder"}, {ID: "close", Label: "Tutup"}}
		} else {
			res.Title = "Papan Misi"
			res.Text = boardText(p)
			res.Options = []DialogOption{{ID: "close", Label: "Tutup"}}
		}
	case "BLACKSMITH":
		res.Text = lineOr("mbah_karya", "Le, nek arep ndandani gegamanmu, gawa mrene.")
		res.Subtitle = "Nak, kalau ingin memperbaiki perlengkapanmu, bawa kemari."
		res.Options = professionNpcOptions(p, npc, "pq-mine-1", "Sinau nambang")
	case "COOK":
		res.Text = lineOr("mbok_rasa", "Le, yen ngelih, masak ing kene. Aja mangan barang mbebayani.")
		res.Subtitle = "Nak, kalau lapar, masak di sini. Jangan makan barang berbahaya."
		res.Options = professionNpcOptions(p, npc, "pq-edu-1", "Sinau ngetung")
	case "FISHER":
		res.Text = lineOr("pak_jala", "Le, iwak ana ing kali lan tlaga. Tunggu wektune, banjur tarik.")
		res.Subtitle = "Nak, ikan ada di sungai dan danau. Tunggu waktunya, lalu tarik."
		res.Options = professionNpcOptions(p, npc, "", "")
	case "MINER":
		res.Text = lineOr("mbah_batu", "Le, watu lan bijih ana ing Stone Valley. Aja kesusu.")
		res.Subtitle = "Nak, batu dan bijih ada di Stone Valley. Jangan terburu-buru."
		res.Options = professionNpcOptions(p, npc, "pq-mine-1", "Sinau nambang")
	case "PEDAGANG":
		return merchantInteract(p)
	}
	w.applyPhase24Talk(p, npc, &res)
	if res.Text == "" {
		res.Text = lineOr(npc.DialogueID, npc.Name+" nyapa.")
	}
	return [][]byte{marshal(TypeInteractResult, res)}
}

func lineOr(id, fallback string) string {
	if d, ok := dialogueCatalog[id]; ok && d.Text != "" {
		return d.Text
	}
	return fallback
}

func questOption(p *Player, id, acceptLabel string) []DialogOption {
	q := p.quest(id)
	if q == nil {
		return nil
	}
	switch q.State {
	case QuestAvailable:
		return []DialogOption{
			{ID: "accept:" + id, Label: acceptLabel},
			{ID: "decline:" + id, Label: "Nanti dulu"},
		}
	case QuestCompleted:
		return []DialogOption{{ID: "claim:" + id, Label: "Ambil hadiah"}}
	}
	return nil
}

func elderOptions(p *Player, npc NPCDef) []DialogOption {
	opts := []DialogOption{{ID: "talk:forest", Label: "Apa yang terjadi di hutan?"}}
	for _, id := range []string{"mq001", "mq002", "mq003", "mq005"} {
		opts = append(opts, questOption(p, id, "Siap membantu")...)
	}
	opts = append(opts, DialogOption{ID: "close", Label: "Tutup"})
	return opts
}

func miraOptions(p *Player, npc NPCDef) []DialogOption {
	opts := []DialogOption{
		{ID: "choice-a", Label: "A. Ceritakan pelan-pelan"},
		{ID: "choice-b", Label: "B. Di mana kuil belajar?"},
		{ID: "choice-c", Label: "C. Aku ingin cepat ke hutan"},
	}
	opts = append(opts, questOption(p, "oq001", "Terima misi hutan")...)
	opts = append(opts, questOption(p, "eq001", "Mulai ujian")...)
	q := p.quest("eq001")
	if q != nil && q.State == QuestActive {
		opts = append(opts, DialogOption{ID: "quiz", Label: "Jawab pertanyaan"})
	}
	opts = append(opts, DialogOption{ID: "close", Label: "Tutup"})
	return opts
}

func lioOptions(p *Player, npc NPCDef) []DialogOption {
	opts := questOption(p, "sq001", "Aku akan mencari")
	opts = append(opts, DialogOption{ID: "shop", Label: "Lihat toko"}, DialogOption{ID: "close", Label: "Tutup"})
	return opts
}

func ravenOptions(p *Player, npc NPCDef) []DialogOption {
	opts := questOption(p, "mq004", "Aku siap ke hutan")
	if p.ensureLog().ForestUnlocked {
		opts = append(opts, DialogOption{ID: "close", Label: "Masuk hutan"})
	} else {
		opts = append(opts, DialogOption{ID: "close", Label: "Tutup"})
	}
	return opts
}

func naraOptions(p *Player, npc NPCDef) []DialogOption {
	opts := questOption(p, "sq002", "Aku akan membantu")
	opts = append(opts, DialogOption{ID: "heal", Label: "HEAL"}, DialogOption{ID: "close", Label: "Tutup"})
	return opts
}

func boardText(p *Player) string {
	text := "Misi desa:\n"
	for _, def := range questCatalog {
		q := p.quest(def.ID)
		if q == nil {
			continue
		}
		text += "- " + def.Title + " [" + q.State + "]\n"
	}
	return text
}

func (w *WorldState) openChest(p *Player, obj InteractDef) [][]byte {
	log := p.ensureLog()
	if log.Claimed[obj.ID] {
		return [][]byte{marshal(TypeInteractResult, InteractResult{
			Kind: "chest", TargetID: obj.ID, Title: "Peti", Text: "Peti ini sudah kau buka.",
			Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		})}
	}
	log.Claimed[obj.ID] = true
	bundle := RewardDef{Coin: obj.Loot["coin"], Potion: obj.Loot["potion"], Crystal: obj.Loot["crystal"]}
	w.giveRewardBundle(p, bundle)
	p.markDirty()
	w.persist(p)
	return [][]byte{
		marshal(TypeInteractResult, InteractResult{
			Kind: "chest", TargetID: obj.ID, Title: "Peti", Text: "Kau menemukan persediaan desa.",
			Rewards: &RewardView{Coin: obj.Loot["coin"], Potion: obj.Loot["potion"], Crystal: obj.Loot["crystal"]},
			Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		}),
		marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
	}
}

func (w *WorldState) collectCrystal(p *Player, obj InteractDef) [][]byte {
	log := p.ensureLog()
	if log.Claimed[obj.ID] {
		return rejectFor(p.ID, TypeInteract, "already")
	}
	log.Claimed[obj.ID] = true
	w.giveItem(p, "crystal_shard", 1)
	p.credit("COLLECT", "crystal", 1)
	p.markDirty()
	w.persist(p)
	w.persistGear(p)
	return [][]byte{
		marshal(TypeInteractResult, InteractResult{
			Kind: "crystal", TargetID: obj.ID, Title: "Kristal Cahaya",
			Text: "Kristal berkilau di tanganmu.", Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		}),
		marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
	}
}

func (w *WorldState) useCheckpoint(p *Player, obj InteractDef) [][]byte {
	p.HP = p.MaxHP
	p.Energy = p.MaxEnergy
	p.Stamina = p.MaxStamina
	log := p.ensureLog()
	log.FastTravel[obj.ID] = true
	log.StoryCheckpoint = obj.ID
	p.markDirty()
	w.persist(p)
	w.persistGear(p)
	name := checkpointDisplayName(obj.ID)
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "checkpoint", TargetID: obj.ID, Title: name,
		Text:    "Perjalanan tersimpan. HP, Energy, dan Stamina pulih. Lanjut dari sini saat login berikutnya.",
		Options: []DialogOption{{ID: "close", Label: "Tutup"}},
	}), marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)), marshal(TypeSaveOk, map[string]string{"checkpoint": name})}
}

func (w *WorldState) useGate(p *Player, obj InteractDef) [][]byte {
	if p.ensureLog().ForestUnlocked {
		return [][]byte{marshal(TypeInteractResult, InteractResult{
			Kind: "gate", TargetID: obj.ID, Title: "Village Gate",
			Text:    "Gerbang terbuka menuju Whisper Forest.",
			Options: []DialogOption{{ID: "close", Label: "Tutup"}},
		})}
	}
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "gate", TargetID: obj.ID, Title: "Village Gate", Locked: true,
		Text:    "Complete the required quest.",
		Options: []DialogOption{{ID: "close", Label: "Tutup"}},
	})}
}

func (w *WorldState) healAtNara(p *Player) [][]byte {
	nara := npcByID["nara"]
	if math.Hypot(p.X-nara.X, p.Z-nara.Z) > nara.InteractionRange+0.5 {
		return rejectFor(p.ID, TypeHeal, "distance")
	}
	amount := p.MaxHP
	if inst := w.dungeonOf(p.ID); inst != nil {
		amount = horizonHealAmount(inst, p.MaxHP)
	}
	p.HP += amount
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
	p.Energy = p.MaxEnergy
	p.Stamina = p.MaxStamina
	return [][]byte{marshal(TypeInteractResult, InteractResult{
		Kind: "heal", TargetID: "nara", Speaker: nara.Name, Role: nara.Role,
		Title: nara.Name, Text: "Luka-lukamu mereda. Istirahatlah sejenak.",
		Options: []DialogOption{{ID: "close", Label: "Tutup"}},
	})}
}

func (w *WorldState) openShop(p *Player) [][]byte {
	for _, id := range []string{"lio", "dawn_merchant"} {
		n := npcByID[id]
		if n.ID == "" {
			continue
		}
		r := n.InteractionRange
		if r <= 0 {
			r = 2.6
		}
		if math.Hypot(p.X-n.X, p.Z-n.Z) <= r+0.8 {
			return w.shopCatalogFor(p)
		}
	}
	return rejectFor(p.ID, TypeShopOpen, "distance")
}

func (w *WorldState) answerQuestion(p *Player, in EducationAnswerIn) [][]byte {
	log := p.ensureLog()
	if log.Quiz.Active && (log.Quiz.QuestID == "pq-edu-1" || log.Quiz.QuestID == "fish-edu" || log.Quiz.QuestID == "craft-edu" || log.Quiz.QuestID == "econ-edu") {
		return w.answerProfessionQuiz(p, in)
	}
	if inst := w.dungeonOf(p.ID); inst != nil && inst.EduShield {
		return w.answerDungeonQuestion(p, in)
	}
	if log.Quiz.Active && log.Quiz.QuestID == "dungeon" {
		return w.answerDungeonQuestion(p, in)
	}
	if log.Quiz.Active && log.Quiz.QuestID == "edu-shrine" {
		return w.answerShrine(p, in)
	}
	if log.Quiz.Active && log.Quiz.QuestID == "ctq001" {
		return w.answerTrainingQuiz(p, in)
	}
	if log.Quiz.Active && log.Quiz.QuestID == "travel-puzzle" {
		return w.answerTravelPuzzle(p, in)
	}
	if log.Quiz.Active && strings.HasPrefix(log.Quiz.QuestID, "nq-guru") {
		return w.answerPhase24Edu(p, in)
	}
	if log.Quiz.Active && log.Quiz.QuestID == "nq-edu-combat" {
		return w.answerPhase25Edu(p, in)
	}
	if log.Quiz.Active && isStoryQuiz(log.Quiz.QuestID) {
		return w.answerStoryQuiz(p, in)
	}
	q := p.quest("eq001")
	if q == nil || q.State != QuestActive {
		return rejectFor(p.ID, TypeEducationAnswer, "not_active")
	}
	if !log.Quiz.Active {
		return rejectFor(p.ID, TypeEducationAnswer, "no_session")
	}
	def, ok := questionByID[in.QuestionID]
	if !ok {
		return rejectFor(p.ID, TypeEducationAnswer, "question")
	}
	if log.Quiz.Index >= len(questionCatalog) || questionCatalog[log.Quiz.Index].ID != in.QuestionID {
		return rejectFor(p.ID, TypeEducationAnswer, "order")
	}
	if in.Choice != def.Correct {
		q.Wrongs++
		q.Perfect = false
		p.markDirty()
		w.persist(p)
		return [][]byte{marshal(TypeEducationFeedback, EducationFeedback{
			Correct: false, Explain: def.Explain, Retry: true, Toast: "Coba lagi.",
			Question: questionOut(log.Quiz.Index),
		})}
	}
	p.credit("ANSWER", "quiz_mira", 1)
	w.grantEducationBonus(p)
	w.breakEducationShield(p)
	log.LastQuestion = def.ID
	events := w.grantExp(p, 8)
	events = append(events, w.recordEducation(p, def.Category, true)...)
	events = append(events, w.refreshAwakening(p)...)
	log.Quiz.Index++
	if log.Quiz.Index >= len(questionCatalog) {
		log.Quiz.Active = false
		if objectivesDone(q, questByID["eq001"]) {
			q.State = QuestCompleted
		}
		p.markDirty()
		w.persist(p)
		fb := EducationFeedback{Correct: true, Explain: def.Explain, Toast: "Semua pertanyaan selesai. Kembali ke Mira untuk hadiah."}
		if q.Perfect && q.Wrongs == 0 {
			fb.Toast = "PERFECT!"
		}
		out := [][]byte{
			marshal(TypeEducationFeedback, fb),
			marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
		}
		return append(events, out...)
	}
	p.markDirty()
	w.persist(p)
	out := [][]byte{
		marshal(TypeEducationFeedback, EducationFeedback{
			Correct: true, Explain: def.Explain, Question: questionOut(log.Quiz.Index),
		}),
		marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
	}
	return append(events, out...)
}

func questionOut(index int) *QuestionOut {
	if index < 0 || index >= len(questionCatalog) {
		return nil
	}
	q := questionCatalog[index]
	return &QuestionOut{
		ID: q.ID, Index: index + 1, Total: len(questionCatalog),
		Category: q.Category, Prompt: q.Prompt, Choices: q.Choices,
	}
}

func (w *WorldState) startQuiz(p *Player) [][]byte {
	q := p.quest("eq001")
	if q == nil || (q.State != QuestActive && q.State != QuestAvailable) {
		return rejectFor(p.ID, TypeInteract, "quest")
	}
	if q.State == QuestAvailable {
		q.State = QuestActive
	}
	p.ensureLog().Quiz = QuizSession{QuestID: "eq001", Index: 0, Active: true}
	p.markDirty()
	w.persist(p)
	return [][]byte{
		marshal(TypeEducationQuestion, questionOut(0)),
		marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)),
	}
}

func (w *WorldState) manualSave(p *Player) [][]byte {
	if p == nil || !p.alive() {
		return rejectFor(p.ID, TypeManualSave, "dead")
	}
	near := false
	for _, lm := range landmarkCatalog {
		if !lm.Rest {
			continue
		}
		if math.Hypot(p.X-lm.X, p.Z-lm.Z) > 6 {
			continue
		}
		near = true
		if p.ensureLog().StoryCheckpoint == "" {
			p.ensureLog().StoryCheckpoint = lm.ID
		}
		break
	}
	if !near {
		return rejectFor(p.ID, TypeManualSave, "checkpoint")
	}
	p.markDirty()
	w.persist(p)
	w.persistGear(p)
	name := checkpointDisplayName(p.ensureLog().StoryCheckpoint)
	return [][]byte{marshal(TypeSaveOk, map[string]string{"checkpoint": name})}
}

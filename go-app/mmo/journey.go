package mmo

import "math"

// Journey is the single current travel objective shown to the player.
type JourneyView struct {
	Title          string  `json:"title"`
	Objective      string  `json:"objective"`
	Hint           string  `json:"hint"`
	Region         string  `json:"region"`
	NextRegion     string  `json:"nextRegion"`
	NavX           float64 `json:"navX"`
	NavZ           float64 `json:"navZ"`
	Landmark       string  `json:"landmark,omitempty"`
	Cardinal       string  `json:"cardinal,omitempty"`
	SubObjective   string  `json:"subObjective,omitempty"`
	Optional       string  `json:"optional,omitempty"`
	OptionalX      float64 `json:"optionalX,omitempty"`
	OptionalZ      float64 `json:"optionalZ,omitempty"`
	HintJv         string  `json:"hintJv,omitempty"`
	HintId         string  `json:"hintId,omitempty"`
}

func (p *Player) journeyView() JourneyView {
	log := p.ensureLog()
	z := zoneAt(p.X, p.Z)
	j := JourneyView{Region: z.ID, Title: "PERJALANAN"}
	switch {
	case !log.Flags["village_intro_complete"] && !log.Claimed["mq001"] && p.quest("mq001") != nil && p.quest("mq001").State != QuestClaimed:
		j.Objective = "Temui Mbah Jagat dan Elder Ardan."
		j.Hint = "Dengarkan mereka sebelum meninggalkan desa."
		j.SubObjective = "Cari alun-alun di tengah desa."
		j.Landmark = "Elder Ardan"
		j.NavX, j.NavZ = 0, 3.2
		j.NextRegion = "forest"
	case !log.Flags["training_complete"] && p.quest("mq002") != nil && p.quest("mq002").State != QuestClaimed:
		j.Objective = "Latihan di tanah desa, lalu menuju gerbang."
		j.Hint = "Uji pukulan pada patung latihan."
		j.Landmark = "Tanah Latihan"
		j.NavX, j.NavZ = 0, 9.4
		j.NextRegion = "forest"
	case !log.ForestUnlocked:
		j.Objective = "Temukan jalan menuju Hutan Larangan."
		j.Hint = "Ikuti jalan tanah ke gerbang utara. Temui Guard Raven."
		j.SubObjective = "Ikuti lampu di tepi jalan hingga gerbang."
		j.Landmark = "Gerbang Desa"
		j.NavX, j.NavZ = 0, 19.6
		j.NextRegion = "forest"
	case z.ID == "village":
		j.Objective = "Tinggalkan desa. Ikuti jalan menuju hutan."
		j.Hint = "Gerbang sudah terbuka."
		j.Landmark = "Hutan Larangan"
		j.NavX, j.NavZ = 0, 26
		j.NextRegion = "forest"
	case z.ID == "forest":
		j.Objective = "Temukan jalan menuju Gunung Kabut."
		j.Hint = "Jalan utama di tengah. Sisi hutan menyimpan rahasia."
		j.SubObjective = "Ikuti sungai hingga menemukan jembatan tua."
		j.Optional = "Temukan air terjun kabut."
		j.OptionalX, j.OptionalZ = -11, 42
		j.Landmark = "Gunung Kabut"
		j.NavX, j.NavZ = 0, 48
		j.NextRegion = "valley"
	case z.ID == "valley":
		j.Objective = "Temukan jalan menuju kuil di balik tebing."
		j.Hint = "Jembatan batu menandai jalur utama."
		j.SubObjective = "Lewati gerbang batu di tengah lembah."
		j.Landmark = "Gerbang Batu"
		j.NavX, j.NavZ = 0, 70
		j.NextRegion = "plains"
	case z.ID == "plains":
		j.Objective = "Ikuti sungai cahaya menuju dataran merah."
		j.SubObjective = "Jangan tinggalkan tepi sungai terlalu lama."
		j.Landmark = "Sungai Cahaya"
		j.NavX, j.NavZ = 0, 92
		j.NextRegion = "canyon"
	case z.ID == "canyon":
		j.Objective = "Lewati kuil merah menuju pegunungan."
		j.Landmark = "Puncak Fajar"
		j.NavX, j.NavZ = 0, 112
		j.NextRegion = "temple"
	case z.ID == "temple":
		j.Objective = "Daki pegunungan menuju gurun kuno."
		j.Landmark = "Menara Langit"
		j.NavX, j.NavZ = 0, 128
		j.NextRegion = "ruins"
	case z.ID == "ruins":
		j.Objective = "Seberangi gurun menuju cakrawala suci."
		j.Landmark = "Gerbang Cakrawala"
		j.NavX, j.NavZ = 0, 144
		j.NextRegion = "celestial"
	case z.ID == "celestial":
		j.Objective = "Temukan Masjid Cahaya di ujung jalan."
		j.Hint = "Cahaya di utara adalah tujuanmu."
		j.Landmark = "Masjid Cahaya"
		j.NavX, j.NavZ = 0, 166
		j.NextRegion = "masjid"
	case z.ID == "masjid":
		j.Objective = "Masuki Masjid Cahaya. Perjalanan hampir selesai."
		j.Landmark = "Masjid Cahaya"
		j.NavX, j.NavZ = 0, 166
		j.NextRegion = "horizon"
	default:
		j.Objective = "Jelajahi jalan di depanmu."
		j.Landmark = "Jalan Utama"
		j.NavX, j.NavZ = 0, p.Z+12
	}
	applyJourneyGuidance(p, &j)
	return j
}

func landmarkNameForNav(_, z float64, _ string) string {
	switch {
	case z < 22:
		return "Gerbang Desa"
	case z < 50:
		return "Gunung Kabut"
	case z < 72:
		return "Gerbang Batu"
	case z < 158:
		return "Masjid Cahaya"
	default:
		return "Masjid Cahaya"
	}
}

func (w *WorldState) tickJourney(p *Player) [][]byte {
	if p == nil || p.InstanceID != "" {
		return nil
	}
	var events [][]byte
	notify := false
	if p.Z >= 17.8 && math.Abs(p.X) < 6 {
		notify = p.credit("REACH", "village-gate", 1) || notify
	}
	if p.ensureLog().ForestUnlocked && p.Z >= 22 {
		notify = p.credit("REACH", "forest-path", 1) || notify
	}
	if p.Z >= 48 {
		notify = p.credit("REACH", "valley-gate", 1) || notify
	}
	if notify {
		events = append(events, marshal(TypeQuestUpdated, p.progressOut(w.Time.Phase)))
	}
	return events
}

func regionIntroCinematic(zoneID string) string {
	switch zoneID {
	case "forest":
		return "cin-region-forest"
	case "valley":
		return "cin-region-valley"
	case "plains":
		return "cin-region-river"
	case "canyon":
		return "cin-region-canyon"
	case "temple":
		return "cin-region-mountain"
	case "ruins":
		return "cin-region-desert"
	case "celestial":
		return "cin-region-horizon"
	case "masjid":
		return "cin-region-masjid"
	default:
		return ""
	}
}

func adventureNPCVisible(n NPCDef, wx, wz float64) bool {
	if len(n.ID) >= 6 && n.ID[:6] == "crowd_" {
		return false
	}
	if n.Type == "QUEST_BOARD" {
		return false
	}
	home := zoneAt(n.X, n.Z)
	here := zoneAt(wx, wz)
	if here.ID == "village" {
		if villageJourneyNPC[n.ID] {
			return nearby(wx, wz, n.X, n.Z, 20)
		}
		return math.Hypot(n.X-wx, n.Z-wz) < 3.2
	}
	if n.Type == "ARENA" || n.Type == "GUILD" {
		return math.Hypot(n.X-wx, n.Z-wz) < 8
	}
	if home.ID != here.ID {
		return nearby(wx, wz, n.X, n.Z, 14)
	}
	return nearby(wx, wz, n.X, n.Z, 16)
}

var villageJourneyNPC = map[string]bool{
	"elder_ardan": true,
	"mira":        true,
	"lio":         true,
	"raven":       true,
	"nara":        true,
	"kian":        true,
	"child_lina":  true,
	"mbah_jagat":  true,
	"ibu_desa":    true,
	"petani":      true,
}

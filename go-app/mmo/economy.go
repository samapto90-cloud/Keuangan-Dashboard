package mmo

import (
	"strings"
	"time"

	_ "embed"
)

//go:embed data/guild.json
var guildCfgJSON []byte

//go:embed data/guildQuests.json
var guildQuestJSON []byte

//go:embed data/shops.json
var shopsJSON []byte

//go:embed data/titles.json
var titlesJSON []byte

//go:embed data/cosmetics.json
var cosmeticsJSON []byte

//go:embed data/profanity.json
var profanityJSON []byte

type GuildCfg struct {
	CreateLevel int      `json:"createLevel"`
	CreateCost  int      `json:"createCost"`
	TagMin      int      `json:"tagMin"`
	TagMax      int      `json:"tagMax"`
	NameMin     int      `json:"nameMin"`
	NameMax     int      `json:"nameMax"`
	Reserved        []string `json:"reserved"`
	Levels          []int    `json:"levels"`
	MemberBase      int      `json:"memberBase"`
	MemberPerLevel  int      `json:"memberPerLevel"`
	DescMax         int      `json:"descMax"`
}

type GuildQuestDef struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Kind    string `json:"kind"`
	Target  string `json:"target"`
	Count   int    `json:"count"`
	Reset   string `json:"reset"`
	Rewards struct {
		GuildExp     int `json:"guildExp"`
		Contribution int `json:"contribution"`
		Coin         int `json:"coin"`
	} `json:"rewards"`
}

type ShopDef struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	NPC   string        `json:"npc"`
	Items []ShopItemDef `json:"items"`
}

type ShopItemDef struct {
	ShopItemID    string `json:"shopItemId"`
	ItemID        string `json:"itemId"`
	Price         int    `json:"price"`
	Currency      string `json:"currency"`
	Stock         int    `json:"stock"`
	PurchaseLimit int    `json:"purchaseLimit"`
}

type TitleDef struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Source string `json:"source"`
	Rarity string `json:"rarity"`
}

type CosmeticDef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type AuditEntry struct {
	ID     string
	Kind   string
	Player string
	Detail string
	At     time.Time
}

var (
	guildCfg       GuildCfg
	guildQuestList []GuildQuestDef
	shopCatalog    ShopDef
	titleCatalog   []TitleDef
	titleByID      = map[string]TitleDef{}
	cosmeticCat    []CosmeticDef
	cosmeticByID   = map[string]CosmeticDef{}
	badWords       []string
)

func initEconomy() {
	mustJSON("guild.json", guildCfgJSON, &guildCfg)
	if guildCfg.CreateLevel < 1 {
		guildCfg.CreateLevel = 10
	}
	if guildCfg.CreateCost < 1 {
		guildCfg.CreateCost = 1000
	}
	mustJSON("guildQuests.json", guildQuestJSON, &guildQuestList)
	mustJSON("shops.json", shopsJSON, &shopCatalog)
	mustJSON("titles.json", titlesJSON, &titleCatalog)
	for _, t := range titleCatalog {
		titleByID[t.ID] = t
	}
	mustJSON("cosmetics.json", cosmeticsJSON, &cosmeticCat)
	for _, c := range cosmeticCat {
		cosmeticByID[c.ID] = c
	}
	mustJSON("profanity.json", profanityJSON, &badWords)
}

func init() {
	initEconomy()
}

func (w *WorldState) addCurrency(p *Player, kind string, amount int, reason string) bool {
	if amount == 0 {
		return true
	}
	log := p.ensureLog()
	if amount < 0 {
		have := currencyOf(log, kind)
		if have < -amount {
			return false
		}
	}
	switch kind {
	case "crystal":
		log.Crystal += amount
	case "edu", "education":
		log.EduToken += amount
	case "guild":
		log.GuildTokens += amount
	case "battle":
		log.BattleToken += amount
	case "guardian":
		log.GuardianTokens += amount
	case "raid":
		log.RaidTokens += amount
	default:
		log.Coin += amount
		if amount > 0 {
			log.GoldToday += amount
			p.recordGoldHist(reason, amount, p.ID)
		} else if amount < 0 {
			p.recordGoldHist(reason, amount, "sink")
		}
		capGold(log)
	}
	if log.Coin < 0 || log.Crystal < 0 || log.EduToken < 0 || log.GuildTokens < 0 || log.BattleToken < 0 || log.GuardianTokens < 0 || log.RaidTokens < 0 {
		return false
	}
	p.markDirty()
	w.audit("CURRENCY_CHANGE", p.ID, kind+":"+itoa(amount)+":"+reason)
	return true
}

func (w *WorldState) removeCurrency(p *Player, kind string, amount int, reason string) bool {
	if amount < 0 {
		return false
	}
	return w.addCurrency(p, kind, -amount, reason)
}

func currencyOf(log *PlayerLog, kind string) int {
	switch kind {
	case "crystal":
		return log.Crystal
	case "edu", "education":
		return log.EduToken
	case "guild":
		return log.GuildTokens
	case "battle":
		return log.BattleToken
	case "guardian":
		return log.GuardianTokens
	case "raid":
		return log.RaidTokens
	default:
		return log.Coin
	}
}

func (p *Player) wallet() WalletView {
	log := p.ensureLog()
	return WalletView{Coins: log.Coin, Crystals: log.Crystal, EducationTokens: log.EduToken, GuildTokens: log.GuildTokens, BattleTokens: log.BattleToken, GuardianTokens: log.GuardianTokens, RaidTokens: log.RaidTokens}
}

func (w *WorldState) audit(kind, player, detail string) {
	if w.Audit == nil {
		w.Audit = []AuditEntry{}
	}
	w.Audit = append(w.Audit, AuditEntry{ID: randomID("aud_"), Kind: kind, Player: player, Detail: detail, At: time.Now()})
	if len(w.Audit) > 400 {
		w.Audit = w.Audit[len(w.Audit)-300:]
	}
}

func (w *WorldState) rememberTx(id string, result [][]byte) [][]byte {
	if id == "" {
		return result
	}
	if w.TxDone == nil {
		w.TxDone = map[string][][]byte{}
	}
	if prev, ok := w.TxDone[id]; ok {
		return prev
	}
	w.TxDone[id] = result
	return result
}

func (w *WorldState) txSeen(id string) ([][]byte, bool) {
	if id == "" || w.TxDone == nil {
		return nil, false
	}
	v, ok := w.TxDone[id]
	return v, ok
}

func (w *WorldState) limited(key string, n int, win time.Duration) bool {
	if w.limits == nil {
		w.limits = map[string][]time.Time{}
	}
	now := time.Now()
	cut := now.Add(-win)
	live := w.limits[key][:0]
	for _, t := range w.limits[key] {
		if t.After(cut) {
			live = append(live, t)
		}
	}
	if len(live) >= n {
		w.limits[key] = live
		return true
	}
	w.limits[key] = append(live, now)
	return false
}

func (w *WorldState) notify(p *Player, kind, text string) []byte {
	if p == nil {
		return nil
	}
	log := p.ensureLog()
	n := NotifyView{ID: randomID("nt_"), Type: kind, Message: text, At: time.Now().UnixMilli(), Priority: notifyPriority(kind)}
	log.Notices = append(log.Notices, n)
	if len(log.Notices) > 40 {
		log.Notices = log.Notices[len(log.Notices)-30:]
	}
	p.markDirty()
	return marshal(TypeSocialNotification, SocialNote{Kind: kind, Text: text, ToID: p.ID, Priority: notifyPriority(kind)})
}

func filterChat(s string) string {
	s = htmlStrip(s)
	s = phase26StripLinks(s)
	out := s
	low := strings.ToLower(s)
	for _, w := range badWords {
		if w == "" {
			continue
		}
		if strings.Contains(low, strings.ToLower(w)) {
			out = strings.ReplaceAll(out, w, "***")
			out = strings.ReplaceAll(out, strings.ToUpper(w), "***")
		}
	}
	return out
}

func htmlStrip(s string) string {
	b := make([]rune, 0, len(s))
	skip := false
	for _, r := range s {
		if r == '<' {
			skip = true
			continue
		}
		if r == '>' {
			skip = false
			continue
		}
		if !skip {
			b = append(b, r)
		}
	}
	return string(b)
}

func (w *WorldState) muted(id string) bool {
	if w.Mutes == nil {
		return false
	}
	until := w.Mutes[id]
	return time.Now().Before(until)
}

func (w *WorldState) ApplyEconomy(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	p.LastHeard = time.Now()
	switch env.Type {
	case TypeChat:
		return w.ApplyChat(id, env)
	case TypeGuildCreate, TypeGuildInvite, TypeGuildAccept, TypeGuildDecline, TypeGuildLeave,
		TypeGuildKick, TypeGuildDisband, TypeGuildTransfer, TypeGuildAnnounce, TypeSetGuildRank, TypeGetGuild:
		return w.ApplyGuild(id, env)
	case TypeTradeRequest, TypeTradeAccept, TypeTradeDecline, TypeTradeOffer, TypeTradeReady, TypeTradeConfirm, TypeTradeCancel:
		return w.ApplyTrade(id, env)
	case TypeShopOpen, TypeShopBuy, TypeShopSell, TypeSetCoin:
		return w.ApplyShop(id, env)
	case TypeSetTitle, TypeSetCosmetic:
		return w.ApplyTitle(id, env)
	case TypeReportPlayer, TypeSearchPlayer, TypeGetNotifies, TypeMutePlayer, TypeModKick, TypeModBan, TypeReportMessage:
		return w.ApplyModeration(id, env)
	case TypeGather, TypeCraft, TypeGetCrafting, TypeSetProfession, TypeResetProfession,
		TypeFishStart, TypeFishCatch, TypeNpcShopOpen, TypeNpcShopBuy, TypeNpcShopSell, TypeNpcRepair,
		TypeStallOpen, TypeStallList, TypeStallBuy, TypeStallClose, TypeCraftOrder, TypeCraftOrderAccept,
		TypeGetWorkshop, TypeAddGold, TypeAddMaterial,
		TypeSetPrice, TypeSetGold, TypeGiveGold, TypeCreateRecipe, TypeGuildContribute, TypeGetGoldLog:
		return w.ApplyPhase21(p, env)
	case TypeCreateHouse, TypeHouseLock, TypeHouseRename, TypeHouseStyle, TypeHouseMove, TypeHouseDecorate,
		TypeHouseStore, TypeHouseTake, TypeGardenPlant, TypeGardenWater, TypeGardenHarvest,
		TypePetClaim, TypePetSummon, TypePetDismiss, TypePetCare, TypePetName, TypeGetPets, TypeAddPet,
		TypeGuildHallEnter, TypeGuildHallLeave, TypeGuildHost, TypeGetLife, TypeClaimDailyLife,
		TypeHouseVote, TypeLifeQuiz, TypeClaimCollection:
		return w.ApplyPhase22(p, env)
	default:
		return w.applyPhase14(p, env)
	}
}

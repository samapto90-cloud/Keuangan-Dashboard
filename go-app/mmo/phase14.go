package mmo

import (
	_ "embed"
	"strings"
	"time"
)

//go:embed data/social.json
var socialJSON []byte

//go:embed data/housing.json
var housingJSON []byte

//go:embed data/emblems.json
var emblemsJSON []byte

type SocialCfg struct {
	FriendLimit        int `json:"friendLimit"`
	PartyMax           int `json:"partyMax"`
	PartyInviteSec     int `json:"partyInviteSec"`
	ChatBurst          int `json:"chatBurst"`
	ChatWindowSec      int `json:"chatWindowSec"`
	InventorySlots     int `json:"inventorySlots"`
	BankSlots          int `json:"bankSlots"`
	MaxStack           int `json:"maxStack"`
	MarketFeePct       int `json:"marketFeePct"`
	MarketPage         int `json:"marketPage"`
	MarketMinPrice     int `json:"marketMinPrice"`
	MarketMaxPrice     int `json:"marketMaxPrice"`
	HouseObjectLimit   int `json:"houseObjectLimit"`
	EmoteRate          int `json:"emoteRate"`
	EmoteWindowSec     int `json:"emoteWindowSec"`
	GuildMemberBase     int `json:"guildMemberBase"`
	GuildMemberPerLevel int `json:"guildMemberPerLevel"`
	GuildDescMax        int `json:"guildDescMax"`
	BlockLimit          int `json:"blockLimit"`
}

type HousingCfg struct {
	District string      `json:"district"`
	SafeZone bool        `json:"safeZone"`
	Houses   []HouseDef  `json:"houses"`
	Decor    []DecorDef  `json:"decor"`
}

type HouseDef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slots int    `json:"slots"`
}

type DecorDef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EmblemDef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type EconomyTx struct {
	ID, Player, Currency, Reason string
	Amount                       int
	At                           time.Time
}

type MarketListing struct {
	ID, SellerID, Seller, ItemID, Category, Rarity string
	Qty, Price, Level                              int
	Created                                        int64
}

type MarketHistory struct {
	ID, Kind, ItemID, Other string
	Qty, Price              int
	At                      int64
}

type MarketHub struct {
	Listings []MarketListing
	History  map[string][]MarketHistory
	Flags    []string
}

type HouseItem struct {
	ID, ItemID string
	X, Y, Z    float64
	Yaw        float64
}

type HouseGuest struct {
	PlayerID string `json:"playerId"`
	At       int64  `json:"at"`
}

type GardenPlot struct {
	ID, Plant, State              string
	PlantedAt, WateredAt, ReadyAt int64
}

type HouseInstance struct {
	ID, OwnerID, Type, Access           string
	Items                               []HouseItem
	Visitors                            []string
	Return                              map[string][3]float64
	LocationID, LayoutID, Name          string
	CreatedAt                           int64
	Locked                              bool
	Wall, Floor, Roof, Light, Color     string
	Plots                               []GardenPlot
	GuestLog                            []HouseGuest
	Storage                             *Inventory
	Votes                               map[string]int
	VoteCat                             map[string]string
	GuildHall                           bool
}

type HousingHub struct {
	ByOwner map[string]*HouseInstance
	ByID    map[string]*HouseInstance
}

type GuildLog struct {
	ID, Player, ItemID, Action string
	Qty                        int
	At                         int64
}

type NameChangeService struct {
	AdminOnly bool
}

var (
	socialCfg   SocialCfg
	housingCfg  HousingCfg
	emblemCat   []EmblemDef
	emblemByID  = map[string]EmblemDef{}
	nameChange  = NameChangeService{AdminOnly: true}
)

func init() {
	mustJSON("social.json", socialJSON, &socialCfg)
	if socialCfg.FriendLimit < 1 {
		socialCfg.FriendLimit = 100
	}
	if socialCfg.PartyMax < 1 {
		socialCfg.PartyMax = 5
	}
	if socialCfg.PartyInviteSec < 1 {
		socialCfg.PartyInviteSec = 30
	}
	if socialCfg.InventorySlots < 30 {
		socialCfg.InventorySlots = 50
	}
	if socialCfg.BankSlots < 1 {
		socialCfg.BankSlots = 100
	}
	if socialCfg.MarketPage < 1 {
		socialCfg.MarketPage = 20
	}
	if socialCfg.MarketFeePct < 0 {
		socialCfg.MarketFeePct = 5
	}
	if socialCfg.HouseObjectLimit < 1 {
		socialCfg.HouseObjectLimit = 40
	}
	mustJSON("housing.json", housingJSON, &housingCfg)
	mustJSON("emblems.json", emblemsJSON, &emblemCat)
	for _, e := range emblemCat {
		emblemByID[e.ID] = e
	}
}

func partyCap() int {
	if socialCfg.PartyMax > 0 && socialCfg.PartyMax < PartyMaxSize {
		return socialCfg.PartyMax
	}
	if socialCfg.PartyMax > 0 {
		return socialCfg.PartyMax
	}
	return PartyMaxSize
}

func guildMemberCap(level int) int {
	base := guildCfg.MemberBase
	if base < 1 {
		base = socialCfg.GuildMemberBase
	}
	if base < 1 {
		base = 30
	}
	per := guildCfg.MemberPerLevel
	if per < 1 {
		per = socialCfg.GuildMemberPerLevel
	}
	if level < 1 {
		level = 1
	}
	return base + per*(level-1)
}

func friendCount(s *SocialHub, id string) int {
	if s == nil || s.Friends[id] == nil {
		return 0
	}
	return len(s.Friends[id])
}

func (w *WorldState) ApplyPhase14(id string, env Envelope) [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.players[id]
	if p == nil || !p.Connected {
		return nil
	}
	return w.applyPhase14(p, env)
}

func (w *WorldState) applyPhase14(p *Player, env Envelope) [][]byte {
	switch env.Type {
	case TypeGetMarket, TypeMarketSearch:
		var in struct {
			Category, Rarity, Sort string
			MinPrice, MaxPrice, Level, Page int
		}
		_ = unmarshal(env.Data, &in)
		return w.marketSearch(p, in.Category, in.Rarity, in.Sort, in.MinPrice, in.MaxPrice, in.Level, in.Page)
	case TypeMarketList:
		var in struct {
			ItemID, TransactionID string
			Slot, Qty, Price      int
		}
		_ = unmarshal(env.Data, &in)
		return w.marketList(p, in.Slot, in.ItemID, in.Qty, in.Price, in.TransactionID)
	case TypeMarketBuy:
		var in struct{ ListingID, TransactionID string }
		_ = unmarshal(env.Data, &in)
		return w.marketBuy(p, in.ListingID, in.TransactionID)
	case TypeMarketCancel:
		var in struct{ ListingID string }
		_ = unmarshal(env.Data, &in)
		return w.marketCancel(p, in.ListingID)
	case TypeBankDeposit:
		var in struct{ Slot, Qty int }
		_ = unmarshal(env.Data, &in)
		return w.bankMove(p, true, in.Slot, in.Qty)
	case TypeBankWithdraw:
		var in struct{ Slot, Qty int }
		_ = unmarshal(env.Data, &in)
		return w.bankMove(p, false, in.Slot, in.Qty)
	case TypeGetBank:
		return [][]byte{marshal(TypeBankUpdated, w.bankView(p))}
	case TypeLockItem:
		var in struct{ Slot int; On bool }
		_ = unmarshal(env.Data, &in)
		return w.toggleItemFlag(p, in.Slot, true, in.On)
	case TypeFavoriteItem:
		var in struct{ Slot int; On bool }
		_ = unmarshal(env.Data, &in)
		return w.toggleItemFlag(p, in.Slot, false, in.On)
	case TypeHouseEnter:
		return w.houseEnter(p, p.ID)
	case TypeHouseLeave:
		return w.houseLeave(p)
	case TypeHouseVisit:
		var in struct{ OwnerID string }
		_ = unmarshal(env.Data, &in)
		return w.houseEnter(p, in.OwnerID)
	case TypeHousePlace:
		var in struct {
			ItemID, TransactionID string
			Slot                  int
			X, Z, Yaw             float64
		}
		_ = unmarshal(env.Data, &in)
		return w.housePlaceTx(p, in.Slot, in.ItemID, in.X, in.Z, in.Yaw, in.TransactionID)
	case TypeHouseRemove:
		var in struct{ DecorID string }
		_ = unmarshal(env.Data, &in)
		return w.houseRemove(p, in.DecorID)
	case TypeSetHouseAccess:
		var in struct{ Access string }
		_ = unmarshal(env.Data, &in)
		return w.houseSetAccess(p, in.Access)
	case TypeGetHouse:
		return [][]byte{marshal(TypeHouseState, w.houseView(w.houseOfPlayer(p.ID), p))}
	case TypeGuildApply:
		var in struct{ GuildID string }
		_ = unmarshal(env.Data, &in)
		return w.guildApply(p, in.GuildID)
	case TypeGuildReview:
		var in struct {
			TargetID string
			Accept   bool
		}
		_ = unmarshal(env.Data, &in)
		return w.guildReview(p, in.TargetID, in.Accept)
	case TypeGuildDeposit:
		var in struct{ Slot, Qty int }
		_ = unmarshal(env.Data, &in)
		return w.guildStorageMove(p, true, in.Slot, in.Qty)
	case TypeGuildWithdraw:
		var in struct{ Slot, Qty int }
		_ = unmarshal(env.Data, &in)
		return w.guildStorageMove(p, false, in.Slot, in.Qty)
	case TypeGetGuildLog:
		return w.guildLogView(p)
	case TypeGuildSetEmblem:
		var in struct{ EmblemID string }
		_ = unmarshal(env.Data, &in)
		return w.guildSetEmblem(p, in.EmblemID)
	case TypeGuildSetDesc:
		var in struct{ Text string }
		_ = unmarshal(env.Data, &in)
		return w.guildSetDesc(p, in.Text)
	case TypeSocialEmote:
		var in struct{ Emote string }
		_ = unmarshal(env.Data, &in)
		return w.socialEmote(p, in.Emote)
	case TypeGetPlayerCard:
		var in struct{ TargetID string }
		_ = unmarshal(env.Data, &in)
		return w.playerCard(p, in.TargetID)
	case TypeSetName:
		return rejectFor(p.ID, TypeSetName, "admin_only")
	case TypeSetGuildXP, TypeSetOwned, TypeGetEconomy:
		return rejectFor(p.ID, env.Type, "server_authoritative")
		return rejectFor(p.ID, env.Type, "server_authoritative")
	default:
		return rejectFor(p.ID, env.Type, "unknown")
	}
}

func (w *WorldState) economyCredit(p *Player, kind string, amount int, reason, tx string) bool {
	if prev, ok := w.txSeen(tx); ok {
		_ = prev
		return true
	}
	if amount < 0 {
		return false
	}
	ok := w.addCurrency(p, kind, amount, reason)
	if ok {
		w.recordEconomy(p.ID, kind, amount, reason, tx)
		if tx != "" {
			w.rememberTx(tx, [][]byte{marshal(TypeInventoryUpdated, p.loadout("", nil))})
		}
	}
	return ok
}

func (w *WorldState) economyDebit(p *Player, kind string, amount int, reason, tx string) bool {
	if prev, ok := w.txSeen(tx); ok {
		_ = prev
		return true
	}
	if amount < 0 {
		return false
	}
	ok := w.removeCurrency(p, kind, amount, reason)
	if ok {
		w.recordEconomy(p.ID, kind, -amount, reason, tx)
		if tx != "" {
			w.rememberTx(tx, [][]byte{marshal(TypeInventoryUpdated, p.loadout("", nil))})
		}
	}
	return ok
}

func (w *WorldState) recordEconomy(player, kind string, amount int, reason, tx string) {
	if w.EconomyLog == nil {
		w.EconomyLog = []EconomyTx{}
	}
	w.EconomyLog = append(w.EconomyLog, EconomyTx{
		ID: tx, Player: player, Currency: kind, Amount: amount, Reason: reason, At: time.Now(),
	})
	if len(w.EconomyLog) > 400 {
		w.EconomyLog = w.EconomyLog[len(w.EconomyLog)-300:]
	}
}

func (w *WorldState) socialEmote(p *Player, emote string) [][]byte {
	emote = strings.ToLower(strings.TrimSpace(emote))
	ok := false
	for _, e := range []string{"wave", "cheer", "bow", "sit", "laugh", "clap"} {
		if e == emote {
			ok = true
			break
		}
	}
	if !ok {
		return rejectFor(p.ID, TypeSocialEmote, "emote")
	}
	n, win := socialCfg.EmoteRate, time.Duration(socialCfg.EmoteWindowSec)*time.Second
	if n < 1 {
		n = 4
	}
	if win < time.Second {
		win = 10 * time.Second
	}
	if w.limited("emote:"+p.ID, n, win) {
		return rejectFor(p.ID, TypeSocialEmote, "rate")
	}
	out := map[string]any{"playerId": p.ID, "emote": emote, "x": p.X, "z": p.Z, "notifyIds": []string{p.ID}}
	ids := []string{p.ID}
	for _, o := range w.players {
		if o == nil || o.ID == p.ID || !o.Connected {
			continue
		}
		if hypot2(p.X, p.Z, o.X, o.Z) > 30*30 {
			continue
		}
		ids = append(ids, o.ID)
	}
	out["notifyIds"] = ids
	return [][]byte{marshal(TypeEmotePlayed, out)}
}

func (w *WorldState) playerCard(viewer *Player, targetID string) [][]byte {
	if targetID == "" {
		targetID = viewer.ID
	}
	t := w.players[targetID]
	if t == nil {
		return rejectFor(viewer.ID, TypeGetPlayerCard, "target")
	}
	log := t.ensureLog()
	w.refreshTitles(t)
	gname, gtag := "", t.GuildTag
	if g := w.guildOf(t.ID); g != nil {
		gname, gtag = g.Name, g.Tag
	}
	return [][]byte{marshal(TypePlayerCard, map[string]any{
		"playerId": t.ID, "name": t.Name, "level": t.Level, "class": t.Class,
		"guild": gname, "guildTag": gtag, "title": t.Title, "rank": log.PvpHighestRank,
		"achievements": len(log.Achievements), "cosmetic": log.ActiveCosmetic,
		"season": seasonTrack.Name, "status": w.presenceOf(t), "toId": viewer.ID,
	})}
}

func reportCategoryOK(cat string) bool {
	switch strings.ToUpper(cat) {
	case "HARASSMENT", "SPAM", "CHEATING", "ABUSE", "INAPPROPRIATE_NAME", "PLAYER",
		"INAPPROPRIATE", "SCAM", "OTHER":
		return true
	}
	return false
}

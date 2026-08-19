import { escapeHtml } from "../dialogue/DialogueUI";
import type { LoadoutStore } from "../inventory/LoadoutStore";
import { FriendList } from "./FriendList";
import { NearbyPlayers } from "./NearbyPlayers";
import { PlayerInspect } from "./PlayerInspect";
import { PlayerCard } from "./PlayerCard";
import { marketHtml } from "./MarketUI";
import { housingHtml } from "./HousingUI";
import type { HouseState, MarketState, SearchHit, TradeView } from "../network/NetworkMessage";

export type SocialTab = "FRIENDS" | "PARTY" | "GUILD" | "CHAT" | "BLOCKED" | "TRADE" | "MARKET" | "HOUSING" | "PROFILE";

const TABS: SocialTab[] = ["FRIENDS", "PARTY", "GUILD", "CHAT", "BLOCKED", "TRADE", "MARKET", "HOUSING", "PROFILE"];

export class SocialUI {
  readonly root: HTMLElement;
  readonly inspect: PlayerInspect;
  readonly card: PlayerCard;
  blocking = false;
  tab: SocialTab = "FRIENDS";
  friendTab: "ONLINE" | "AWAY" | "OFFLINE" | "REQUESTS" = "ONLINE";
  guildTab: "OVERVIEW" | "MEMBERS" | "QUEST" | "CHAT" | "HALL" | "SETTINGS" = "OVERVIEW";
  chatHint = "AREA / WORLD / PARTY / GUILD / PRIVATE";
  searchHits: SearchHit[] = [];
  market: MarketState | null = null;
  house: HouseState | null = null;
  trade: TradeView | null = null;
  onInvite: ((id: string) => void) | null = null;
  onFriend: ((id: string) => void) | null = null;
  onInspect: ((id: string) => void) | null = null;
  onCard: ((id: string) => void) | null = null;
  onAcceptFriend: ((id: string) => void) | null = null;
  onDeclineFriend: ((id: string) => void) | null = null;
  onRemoveFriend: ((id: string) => void) | null = null;
  onBlock: ((id: string) => void) | null = null;
  onUnblock: ((id: string) => void) | null = null;
  onLeaveParty: (() => void) | null = null;
  onPartyReady: (() => void) | null = null;
  onPartyCreate: (() => void) | null = null;
  onGuildHall: (() => void) | null = null;
  onPrivacy: ((key: string, value: string) => void) | null = null;
  onRefresh: (() => void) | null = null;
  onMessage: ((id: string, name: string) => void) | null = null;
  onTrade: ((id: string) => void) | null = null;
  onSearch: ((name: string) => void) | null = null;
  onReport: ((id: string, category: string) => void) | null = null;
  onMarketBuy: ((id: string) => void) | null = null;
  onMarketCancel: ((id: string) => void) | null = null;
  onMarketList: ((slot: number) => void) | null = null;
  onMarketRefresh: (() => void) | null = null;
  onHouseEnter: (() => void) | null = null;
  onHouseLeave: (() => void) | null = null;
  onHousePlace: ((slot: number) => void) | null = null;
  onHouseRemove: ((id: string) => void) | null = null;
  onHouseAccess: ((access: string) => void) | null = null;
  onHouseCmd: ((cmd: string, id: string) => void) | null = null;
  onEmote: ((emote: string) => void) | null = null;
  onGuildInvite: ((id: string) => void) | null = null;
  onGuildDeposit: ((slot: number) => void) | null = null;
  onGuildWithdraw: ((slot: number) => void) | null = null;

  constructor(
    host: HTMLElement,
    private readonly store: LoadoutStore,
  ) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay social-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.inspect = new PlayerInspect(host);
    this.card = new PlayerCard(host);
    this.root.addEventListener("click", (e) => this.onClick(e));
  }

  get overlayOpen(): boolean {
    return this.blocking || this.inspect.blocking || this.card.blocking;
  }

  toggle(): void {
    if (this.blocking) this.close();
    else this.open();
  }

  open(tab: SocialTab = this.tab): void {
    this.tab = tab;
    this.blocking = true;
    this.root.hidden = false;
    this.onRefresh?.();
    this.render();
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
    this.inspect.close();
    this.card.close();
  }

  render(): void {
    if (!this.blocking) return;
    const nav = TABS.map((t) => `<button type="button" data-tab="${t}" class="${this.tab === t ? "on" : ""}">${t}</button>`).join("");
    this.root.innerHTML = `
      <div class="rpg-card social-card social-hub">
        <header><h3>SOCIAL</h3><button type="button" data-soc="close">Close</button></header>
        <nav class="social-nav">${nav}</nav>
        <form class="social-search"><input name="q" placeholder="Username only" /><button type="submit">SEARCH</button></form>
        ${this.searchHits.map((h) => `
          <div class="social-row">
            <span>${escapeHtml(h.name)} Lv${h.level} ${escapeHtml(h.status || "")}</span>
            <div class="social-acts">
              <button type="button" data-soc="card" data-id="${escapeHtml(h.playerId)}">CARD</button>
              <button type="button" data-soc="friend" data-id="${escapeHtml(h.playerId)}">FRIEND</button>
            </div>
          </div>`).join("")}
        <div class="social-body">${this.body()}</div>
      </div>`;
    this.root.querySelector("form")?.addEventListener("submit", (e) => {
      e.preventDefault();
      const name = (this.root.querySelector("[name=q]") as HTMLInputElement)?.value.trim() || "";
      if (name) this.onSearch?.(name);
    });
  }

  private body(): string {
    const party = this.store.party;
    const guild = this.store.social.guild;
    const wallet = this.store.social.wallet;
    if (this.tab === "FRIENDS") {
      return FriendList.html(this.store.social.friends, this.store.social.pending, this.store.social.outgoing, this.store.social.blocked, this.friendTab)
        + NearbyPlayers.html(this.store.social.nearby);
    }
    if (this.tab === "PARTY") {
      const rows = (party?.members ?? []).map((m) =>
        `<div class="social-row"><span>${escapeHtml(m.name)} Lv${m.level} ${escapeHtml(m.role || "")} HP ${m.hp}/${m.maxHp} EN ${m.energy}/${m.maxEnergy} ${escapeHtml(m.status || "")} ${Number(m.distance || 0).toFixed(0)}m ${m.ready ? "READY" : "WAITING"}</span>
         <button type="button" data-soc="card" data-id="${escapeHtml(m.playerId)}">CARD</button></div>`).join("");
      return `<h4>PARTY ${escapeHtml(party?.state || "WAITING")}</h4>${rows || "<p>Tidak dalam party.</p>"}
        <div class="inv-actions">
          <button type="button" data-soc="create-party">CREATE PARTY</button>
          <button type="button" data-soc="ready">READY</button>
          <button type="button" data-soc="leave">LEAVE</button>
        </div>`;
    }
    if (this.tab === "GUILD") {
      if (!guild?.guildId) return `<p>Belum dalam guild. Buka GUILD (G) untuk membuat. Syarat level 10.</p>`;
      const gnav = ["OVERVIEW", "MEMBERS", "QUEST", "CHAT", "HALL", "SETTINGS"].map((t) =>
        `<button type="button" data-gtab="${t}" class="${this.guildTab === t ? "on" : ""}">${t}</button>`).join("");
      const members = [...(guild.members ?? [])].sort((a, b) => a.rank.localeCompare(b.rank)).map((m) =>
        `<div class="social-row"><span>[${escapeHtml(m.rank === "LEADER" ? "MASTER" : m.rank)}] ${escapeHtml(m.name)}</span></div>`).join("");
      const questTitle = guild.quest === "gq-desa-1" ? "Pertahanan Desa." : (guild.quest || "—");
      const logs = (guild.logs ?? []).slice(-8).reverse().map((l) =>
        `<div class="social-row"><span>${escapeHtml(l.player)} ${escapeHtml(l.action)}ed ${l.qty} ${escapeHtml(l.itemId)}</span></div>`).join("");
      const bag = this.store.slots.filter((s) => s.item).slice(0, 6).map((s) =>
        `<button type="button" data-soc="gdep" data-slot="${s.index}">Deposit ${escapeHtml(s.item?.name || "")}</button>`).join("");
      const storage = (guild.storage ?? []).filter((s) => s.item).map((s) =>
        `<button type="button" data-soc="gwdr" data-slot="${s.index}">Withdraw ${escapeHtml(s.item?.name || "")} x${s.qty}</button>`).join("");
      let inner = "";
      if (this.guildTab === "MEMBERS") inner = `<h4>MEMBERS</h4>${members}`;
      else if (this.guildTab === "QUEST") inner = `<h4>GUILD QUEST</h4><p>Objective: ${escapeHtml(questTitle)}</p><p>Progress ${guild.questProgress || 0}</p><p>Reward: Guild EXP · Guild Token</p>`;
      else if (this.guildTab === "CHAT") inner = `<p>Gunakan channel GUILD pada chat panel.</p>`;
      else if (this.guildTab === "HALL") inner = `<p>Balai guild: rapat, latihan, NPC, papan kabar.</p><button type="button" data-soc="hall">ENTER HALL</button>`;
      else if (this.guildTab === "SETTINGS") inner = `<p>Emblem preset. Deskripsi maks 200 karakter. Disband butuh konfirmasi di panel Guild (G).</p>`;
      else inner = `<div class="guild-emblem">${escapeHtml(guild.emblemId || "light")}</div>
        <div class="char-name">[${escapeHtml(guild.tag)}] ${escapeHtml(guild.name)}</div>
        <p>Level ${guild.level} · EXP ${guild.exp} · ${guild.members?.length || 0}/${guild.capacity || 50}</p>
        <p>${escapeHtml(guild.announcement || "No announcement.")}</p>
        <h4>STORAGE</h4>${storage || "<p>Kosong.</p>"}${bag}
        <h4>LOGS</h4>${logs || "<p>Kosong.</p>"}`;
      return `<div class="social-tabs">${gnav}</div>${inner}`;
    }
    if (this.tab === "BLOCKED") {
      return FriendList.html(this.store.social.friends, this.store.social.pending, this.store.social.outgoing, this.store.social.blocked, "BLOCKED");
    }
    if (this.tab === "CHAT") {
      return `<h4>CHAT</h4><p>${escapeHtml(this.chatHint)}</p>
        <p>Channels: AREA WORLD PARTY GUILD PRIVATE. Maks 200 karakter.</p>
        <div class="inv-actions">
          ${["wave", "cheer", "bow", "sit", "laugh", "clap"].map((e) => `<button type="button" data-soc="emote" data-id="${e}">${e.toUpperCase()}</button>`).join("")}
        </div>`;
    }
    if (this.tab === "TRADE") {
      const a = this.trade?.a;
      const b = this.trade?.b;
      return `<h4>TRADE WINDOW</h4>
        <div class="trade-cols">
          <div><h5>YOU</h5><p>Coin ${a?.coin ?? 0} · ${a?.ready ? "READY" : "OPEN"} · ${a?.confirm ? "CONFIRM" : ""}</p></div>
          <div><h5>THEM</h5><p>Coin ${b?.coin ?? 0} · ${b?.ready ? "READY" : "OPEN"} · ${b?.confirm ? "CONFIRM" : ""}</p></div>
        </div>
        <p>Gunakan jendela Trade yang sudah ada, atau kirim request dari Player Card.</p>`;
    }
    if (this.tab === "MARKET") return marketHtml(this.market, this.store);
    if (this.tab === "HOUSING") return housingHtml(this.house, this.store);
    const priv = this.store.social.privacy;
    const opt = (key: string, cur: string) => ["EVERYONE", "FRIENDS", "NONE"].map((v) =>
      `<button type="button" data-soc="privacy" data-id="${key}" data-name="${v}" class="${(cur || "EVERYONE") === v ? "on" : ""}">${v}</button>`).join("");
    return `<h4>PROFILE</h4>
      <p>${escapeHtml(this.store.stats?.class || "WARRIOR")} Lv${this.store.stats?.level ?? 1}</p>
      <p>Guild ${escapeHtml(guild?.name || "—")}</p>
      <h4>PRIVACY</h4>
      <p>Friend ${opt("friend", priv?.friend || "EVERYONE")}</p>
      <p>Party ${opt("party", priv?.party || "EVERYONE")}</p>
      <p>Trade ${opt("trade", priv?.trade || "EVERYONE")}</p>
      <p>Private chat ${opt("pm", priv?.pm || "EVERYONE")}</p>
      <h4>WALLET</h4>
      <p>COIN ${wallet?.coins ?? this.store.coin}</p>
      <p>BATTLE TOKEN ${wallet?.battleTokens ?? this.store.battleToken}</p>
      <p>GUARDIAN TOKEN ${wallet?.guardianTokens ?? this.store.guardianToken}</p>
      <p>EDUCATION TOKEN ${wallet?.educationTokens ?? this.store.eduToken}</p>`;
  }

  private onClick(e: MouseEvent): void {
    const t = e.target as HTMLElement;
    const ftab = t.closest("[data-ftab]") as HTMLElement | null;
    if (ftab?.dataset.ftab) {
      this.friendTab = ftab.dataset.ftab as "ONLINE" | "AWAY" | "OFFLINE" | "REQUESTS";
      this.render();
      return;
    }
    const gtab = t.closest("[data-gtab]") as HTMLElement | null;
    if (gtab?.dataset.gtab) {
      this.guildTab = gtab.dataset.gtab as "OVERVIEW" | "MEMBERS" | "QUEST" | "CHAT" | "HALL" | "SETTINGS";
      this.render();
      return;
    }
    const tab = t.closest("[data-tab]") as HTMLElement | null;
    if (tab?.dataset.tab) {
      this.tab = tab.dataset.tab as SocialTab;
      this.onRefresh?.();
      this.render();
      return;
    }
    const act = t.closest("[data-soc]") as HTMLElement | null;
    if (!act) return;
    const id = act.dataset.id ?? "";
    const slot = Number(act.dataset.slot || "0");
    if (act.dataset.soc === "close") this.close();
    if (act.dataset.soc === "invite") this.onInvite?.(id);
    if (act.dataset.soc === "friend") this.onFriend?.(id);
    if (act.dataset.soc === "inspect") this.onInspect?.(id);
    if (act.dataset.soc === "card") this.onCard?.(id);
    if (act.dataset.soc === "accept-friend") this.onAcceptFriend?.(id);
    if (act.dataset.soc === "decline-friend") this.onDeclineFriend?.(id);
    if (act.dataset.soc === "remove-friend") this.onRemoveFriend?.(id);
    if (act.dataset.soc === "block") this.onBlock?.(id);
    if (act.dataset.soc === "unblock") this.onUnblock?.(id);
    if (act.dataset.soc === "leave") this.onLeaveParty?.();
    if (act.dataset.soc === "ready") this.onPartyReady?.();
    if (act.dataset.soc === "create-party") this.onPartyCreate?.();
    if (act.dataset.soc === "privacy") this.onPrivacy?.(id, act.dataset.name || "EVERYONE");
    if (act.dataset.soc === "hall") this.onGuildHall?.();
    if (act.dataset.soc === "message") this.onMessage?.(id, act.dataset.name || "");
    if (act.dataset.soc === "trade") this.onTrade?.(id);
    if (act.dataset.soc === "report") this.onReport?.(id, "HARASSMENT");
    if (act.dataset.soc === "market-buy") this.onMarketBuy?.(id);
    if (act.dataset.soc === "market-cancel") this.onMarketCancel?.(id);
    if (act.dataset.soc === "market-list") this.onMarketList?.(slot);
    if (act.dataset.soc === "market-refresh") this.onMarketRefresh?.();
    if (act.dataset.soc === "house-enter") this.onHouseEnter?.();
    if (act.dataset.soc === "house-leave") this.onHouseLeave?.();
    if (act.dataset.soc === "house-place") this.onHousePlace?.(slot);
    if (act.dataset.soc === "house-remove") this.onHouseRemove?.(id);
    if (act.dataset.soc === "house-access") this.onHouseAccess?.(id);
    if (act.dataset.soc === "hcmd") this.onHouseCmd?.(act.dataset.cmd || "", id);
    if (act.dataset.soc === "emote") this.onEmote?.(id);
    if (act.dataset.soc === "gdep") this.onGuildDeposit?.(slot);
    if (act.dataset.soc === "gwdr") this.onGuildWithdraw?.(slot);
  }
}

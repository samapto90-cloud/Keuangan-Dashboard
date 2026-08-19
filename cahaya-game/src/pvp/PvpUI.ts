import { escapeHtml } from "../dialogue/DialogueUI";
import type {
  PvpLobbyOut,
  PvpNearbyView,
  PvpQueueView,
  PvpReadyOut,
  PvpResultOut,
  PvpView,
  LBEntry,
} from "../network/NetworkMessage";

type PvpTab = "TRAINING" | "DUEL" | "ARENA" | "BATTLEGROUND" | "RANKED" | "LEADERBOARD" | "HISTORY";

export class PvpUI {
  readonly overlay: HTMLElement;
  blocking = false;
  lobby: PvpLobbyOut | null = null;
  tab: PvpTab = "ARENA";
  onQueue: ((mode: string) => void) | null = null;
  onLeaveQueue: (() => void) | null = null;
  onReady: ((ok: boolean) => void) | null = null;
  onShop: ((id: string) => void) | null = null;
  onBoard: ((board: string) => void) | null = null;
  onEmote: ((emote: string) => void) | null = null;
  onTraining: (() => void) | null = null;
  onDuel: ((targetId: string) => void) | null = null;
  onDuelAccept: ((ok: boolean) => void) | null = null;
  onClose: (() => void) | null = null;

  constructor(host: HTMLElement) {
    this.overlay = document.createElement("div");
    this.overlay.className = "pvp-overlay";
    this.overlay.hidden = true;
    host.appendChild(this.overlay);
  }

  open(): void {
    this.overlay.hidden = false;
    this.blocking = true;
    this.render();
  }

  close(): void {
    this.overlay.hidden = true;
    this.blocking = false;
  }

  toggle(): void {
    if (this.blocking) this.close();
    else this.open();
  }

  applyLobby(data: PvpLobbyOut): void {
    this.lobby = data;
    if (this.blocking) this.render();
  }

  showQueue(q: PvpQueueView): void {
    if (!this.lobby) this.lobby = emptyLobby();
    this.lobby.queue = q;
    if (this.blocking) this.render();
  }

  showReady(ready: PvpReadyOut): void {
    this.overlay.hidden = false;
    this.blocking = true;
    const names = (ready.members || []).map((m) => `${escapeHtml(m.name)} ${m.ready ? "READY" : "..."}`).join("<br>");
    this.overlay.innerHTML = `<div class="pvp-card">
      <div class="pvp-kicker">MATCH FOUND</div>
      <h3>${escapeHtml(ready.name || ready.mode)}</h3>
      <p>Semua pemain harus READY.</p>
      <p>${names}</p>
      <button type="button" class="pvp-btn" id="pvp-accept">ACCEPT</button>
      <button type="button" id="pvp-decline">DECLINE</button>
    </div>`;
    this.overlay.querySelector("#pvp-accept")?.addEventListener("click", () => this.onReady?.(true));
    this.overlay.querySelector("#pvp-decline")?.addEventListener("click", () => this.onReady?.(false));
  }

  showDuelRequest(from: string, level: number): void {
    this.overlay.hidden = false;
    this.blocking = true;
    this.overlay.innerHTML = `<div class="pvp-card">
      <div class="pvp-kicker">DUEL REQUEST</div>
      <h3>${escapeHtml(from)} menantangmu</h3>
      <p>Level ${level} · 1 vs 1 · tanpa rating</p>
      <button type="button" class="pvp-btn" id="pvp-duel-ok">ACCEPT</button>
      <button type="button" id="pvp-duel-no">DECLINE</button>
    </div>`;
    this.overlay.querySelector("#pvp-duel-ok")?.addEventListener("click", () => this.onDuelAccept?.(true));
    this.overlay.querySelector("#pvp-duel-no")?.addEventListener("click", () => this.onDuelAccept?.(false));
  }

  render(): void {
    const L = this.lobby;
    const p = L?.profile;
    const rank = p?.rankName || "UNRANKED";
    const visual = p?.rankVisual || "bronze";
    const rating = p?.rating ?? 1000;
    const wr = Math.round((p?.winRate || 0) * 100);
    const tabs: PvpTab[] = ["TRAINING", "DUEL", "ARENA", "BATTLEGROUND", "RANKED", "LEADERBOARD", "HISTORY"];
    const tabBtns = tabs.map((t) => `<button type="button" class="pvp-tab${this.tab === t ? " on" : ""}" data-tab="${t}">${t}</button>`).join("");
    let body = "";
    if (this.tab === "LEADERBOARD") body = this.renderBoard();
    else if (this.tab === "HISTORY") body = this.renderHistory();
    else if (this.tab === "TRAINING") body = this.renderTraining();
    else if (this.tab === "DUEL") body = this.renderDuel();
    else body = this.renderModes();
    const q = L?.queue;
    const est = q?.waitEstMs ? ` · perkiraan ${Math.round((q.waitEstMs || 0) / 1000)}s` : "";
    const queueLine = q && q.state === "QUEUED" ? `<p class="pvp-queue">${escapeHtml(q.name || q.mode || "")} · ${q.players}/${q.need} · ${Math.round((q.waitMs || 0) / 1000)}s${est}</p><button type="button" id="pvp-leaveq">CANCEL</button>` : "";
    this.overlay.innerHTML = `<div class="pvp-card pvp-wide">
      <div class="pvp-kicker">PVP</div>
      <h3>Petualangan Menuju Cahaya</h3>
      <div class="pvp-rank pvp-rank-${escapeHtml(visual)}"><strong>${escapeHtml(rank)}</strong> · MMR ${rating} · ${p?.wins ?? 0}W ${p?.losses ?? 0}L · WR ${wr}%</div>
      <div class="pvp-bar"><i style="width:${Math.max(4, ((rating % 500) / 500) * 100)}%"></i></div>
      <div class="pvp-tabs">${tabBtns}</div>
      ${queueLine}
      ${body}
      <div class="pvp-emotes">
        <button type="button" data-emote="wave">WAVE</button>
        <button type="button" data-emote="cheer">CHEER</button>
        <button type="button" data-emote="bow">BOW</button>
        <button type="button" data-emote="laugh">LAUGH</button>
      </div>
      <button type="button" id="pvp-close">Tutup</button>
    </div>`;
    this.overlay.querySelectorAll(".pvp-tab").forEach((el) => {
      el.addEventListener("click", () => {
        this.tab = (el.getAttribute("data-tab") || "ARENA") as PvpTab;
        if (this.tab === "LEADERBOARD") this.onBoard?.("GLOBAL");
        this.render();
      });
    });
    this.overlay.querySelectorAll("[data-mode]").forEach((el) => {
      el.addEventListener("click", () => this.onQueue?.(el.getAttribute("data-mode") || ""));
    });
    this.overlay.querySelectorAll("[data-shop]").forEach((el) => {
      el.addEventListener("click", () => this.onShop?.(el.getAttribute("data-shop") || ""));
    });
    this.overlay.querySelectorAll("[data-board]").forEach((el) => {
      el.addEventListener("click", () => this.onBoard?.(el.getAttribute("data-board") || "GLOBAL"));
    });
    this.overlay.querySelectorAll("[data-emote]").forEach((el) => {
      el.addEventListener("click", () => this.onEmote?.(el.getAttribute("data-emote") || "wave"));
    });
    this.overlay.querySelectorAll("[data-duel]").forEach((el) => {
      el.addEventListener("click", () => this.onDuel?.(el.getAttribute("data-duel") || ""));
    });
    this.overlay.querySelector("#pvp-train")?.addEventListener("click", () => this.onTraining?.());
    this.overlay.querySelector("#pvp-leaveq")?.addEventListener("click", () => this.onLeaveQueue?.());
    this.overlay.querySelector("#pvp-close")?.addEventListener("click", () => {
      this.close();
      this.onClose?.();
    });
  }

  private renderTraining(): string {
    return `<p>Training Arena memakai dummy yang sudah ada. Tanpa ranking.</p>
      <p>Uji combo, skill, dodge, transformasi, dan damage.</p>
      <button type="button" id="pvp-train">MASUK TRAINING ARENA</button>`;
  }

  private renderDuel(): string {
    const nearby = this.lobby?.nearby || [];
    if (!nearby.length) return "<p>Tidak ada pemain di dekatmu untuk duel.</p>";
    return `<p>Friendly Duel · 1 vs 1 · 5 menit · tanpa rating.</p>` + nearby
      .map((n: PvpNearbyView) => `<button type="button" data-duel="${n.playerId}">${escapeHtml(n.name)} · LV ${n.level}${n.cosmetic ? " · " + escapeHtml(n.cosmetic) : ""}</button>`)
      .join("");
  }

  private renderModes(): string {
    const kind = this.tab === "ARENA" ? "CASUAL" : this.tab;
    const modes = (this.lobby?.modes || []).filter((m) => {
      if (this.tab === "ARENA") return m.kind === "CASUAL";
      if (this.tab === "RANKED") return m.kind === "RANKED";
      if (this.tab === "BATTLEGROUND") return m.kind === "BATTLEGROUND";
      return m.kind === kind;
    });
    if (!modes.length) return "<p>Mode belum tersedia.</p>";
    const extra = this.tab === "RANKED" ? this.renderRankedExtras() : this.tab === "BATTLEGROUND" ? "<p>Dawn Arena · 3 Light Shrines · 1000 poin · respawn 8 detik.</p>" : "";
    return extra + modes
      .map((m) => {
        const locked = m.status !== "AVAILABLE" || !m.enabled;
        const map = mapLabel(m.map);
        return `<button type="button" data-mode="${m.id}" ${locked ? "disabled" : ""}>${escapeHtml(m.name)} · ${map} · LV ${m.minLevel} · ${m.teamSize}v${m.teamSize}${locked ? " (" + m.status + ")" : ""}</button>`;
      })
      .join("");
  }

  private renderRankedExtras(): string {
    const s = this.lobby?.season;
    const rewards = (this.lobby?.rewards || []).map((r) => `<li>${escapeHtml(r.name)} · ${r.rank} ${r.unlocked ? "UNLOCKED" : "LOCKED"}</li>`).join("");
    const shop = (this.lobby?.shop || []).map((x) => `<button type="button" data-shop="${x.shopItemId}">${escapeHtml(x.name)} · ${x.price} Token</button>`).join("");
    return `<p>Musim ${escapeHtml(s?.name || "Dawn Awakening")} · ${s?.weeks || 8} minggu · 5 placement.</p>
      <p>Reward kosmetik. Tidak ada keunggulan tempur permanen.</p>
      <ul>${rewards}</ul>
      <h4>Arena Shop</h4>${shop}`;
  }

  private renderBoard(): string {
    const entries = this.lobby?.leaderboard || [];
    const rows = entries
      .map(
        (e: LBEntry) =>
          `<tr><td>${e.rank}</td><td>${escapeHtml(e.player)}</td><td>${e.rating}</td><td>${e.wins}</td><td>${e.losses}</td><td>${Math.round((e.winRate || 0) * 100)}%</td></tr>`,
      )
      .join("");
    return `<div class="pvp-boards">
      <button type="button" data-board="GLOBAL">GLOBAL</button>
      <button type="button" data-board="REGION">REGION</button>
    </div>
    <table class="pvp-table"><thead><tr><th>Rank</th><th>Player</th><th>Rating</th><th>W</th><th>L</th><th>WR</th></tr></thead><tbody>${rows || "<tr><td colspan=6>Kosong</td></tr>"}</tbody></table>
    <p>Username saja. Tidak ada email atau nama asli.</p>`;
  }

  private renderHistory(): string {
    const hist = (this.lobby?.history || [])
      .map((h) => {
        const label = h.result === "VICTORY" ? "VICTORY" : h.result === "DRAW" ? "DRAW" : "DEFEAT";
        const kda = `${h.kills ?? 0}/${h.deaths ?? 0}/${h.assists ?? 0}`;
        return `<li><strong>${label}</strong> · ${escapeHtml(h.mode)} vs ${escapeHtml(h.opponent)} · K/D/A ${kda} · DMG ${h.damage ?? 0} · OBJ ${h.objective ?? 0} · ${h.ratingChange >= 0 ? "+" : ""}${h.ratingChange}</li>`;
      })
      .join("");
    const seasons = (this.lobby?.seasonHistory || [])
      .map((s) => `<li>${escapeHtml(s.season)} · ${escapeHtml(s.highestRank)} · ${s.finalRating} MMR · ${s.wins}W ${s.losses}L</li>`)
      .join("");
    return `<h4>Match History</h4><ul>${hist || "<li>Belum ada pertandingan</li>"}</ul>
      <h4>Season History</h4><ul>${seasons || "<li>Musim berjalan</li>"}</ul>`;
  }
}

export class PvpHud {
  readonly root: HTMLElement;
  readonly result: HTMLElement;
  view: PvpView | null = null;
  private promoTimer = 0;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "pvp-hud";
    this.root.hidden = true;
    this.result = document.createElement("div");
    this.result.className = "pvp-result";
    this.result.hidden = true;
    host.appendChild(this.root);
    host.appendChild(this.result);
  }

  apply(view: PvpView | null | undefined, selfId: string): void {
    this.view = view || null;
    if (!view || !view.matchId) {
      this.root.hidden = true;
      return;
    }
    this.root.hidden = false;
    const enemy = (view.members || []).find((m) => m.playerId !== selfId && m.team !== view.team) || (view.members || []).find((m) => m.playerId !== selfId);
    const self = (view.members || []).find((m) => m.playerId === selfId);
    const mm = Math.floor((view.timeLeft || 0) / 60);
    const ss = String((view.timeLeft || 0) % 60).padStart(2, "0");
    const feed = (view.killFeed || []).slice(-4).map((l) => `<div>${escapeHtml(l)}</div>`).join("");
    const ping = pingColor(self?.pingMs ?? 0);
    if (isBattleground(view.map)) {
      const pts = (view.points || []).map((p) => {
        const own = p.owner === 1 ? "A" : p.owner === 2 ? "B" : "NEUTRAL";
        const st = p.contested ? "CONTESTED" : own;
        return `<span class="pvp-pt">Shrine ${p.id} ${st} ${p.progressA}/${p.progressB}</span>`;
      }).join("");
      this.root.innerHTML = `<div class="pvp-top"><span>Team A ${view.scoreA}</span><strong>${mm}:${ss}</strong><span>Team B ${view.scoreB}</span></div>
        <div class="pvp-pts">${pts}</div>
        <div class="pvp-self">${escapeHtml(self?.name || "You")} ${self?.hp ?? 0}/${self?.maxHp ?? 0} · ping ${ping}</div>
        <div class="pvp-feed">${feed}</div>`;
      return;
    }
    const hp = enemy ? Math.round(((enemy.hp ?? 0) / Math.max(1, enemy.maxHp ?? 1)) * 100) : 0;
    this.root.innerHTML = `<div class="pvp-top"><span class="pvp-en">${escapeHtml(enemy?.name || "Lawan")} HP ${enemy?.hp ?? 0}/${enemy?.maxHp ?? 1}</span><strong>${mm}:${ss}</strong></div>
      <div class="pvp-ehp"><i style="width:${hp}%"></i></div>
      <div class="pvp-self">${escapeHtml(self?.name || "You")} ${self?.hp ?? 0}/${self?.maxHp ?? 0} · ping ${ping}</div>
      <div class="pvp-feed">${feed}</div>`;
  }

  showResult(r: PvpResultOut): void {
    this.result.hidden = false;
    const chg = r.ratingChange >= 0 ? `+${r.ratingChange}` : String(r.ratingChange);
    const headline = r.draw ? "DRAW" : r.victory ? "VICTORY" : "DEFEAT";
    const promo = r.promoted ? `<p class="pvp-promo">Naik peringkat · ${escapeHtml(r.rankName)}</p>` : r.demoted ? `<p>Peringkat disesuaikan · ${escapeHtml(r.rankName)}</p>` : "";
    const mvp = r.mvp ? "<p>MVP · bonus Season XP</p>" : r.mvpName ? `<p>MVP: ${escapeHtml(r.mvpName)}</p>` : "";
    this.result.innerHTML = `<div class="pvp-card">
      <div class="pvp-kicker">${headline}</div>
      <h3>${escapeHtml(r.title)}</h3>
      <p>Durasi ${r.duration ?? 0}s · K/D/A ${r.kills}/${r.deaths}/${r.assists} · Damage ${r.damage} · Objective ${r.objective ?? 0}</p>
      <p>Rating ${chg} → ${r.rating} · ${escapeHtml(r.rankName)}</p>
      ${promo}${mvp}
      <p>Battle Token ${r.battleToken} · Season XP ${r.seasonXP ?? 0}</p>
      <button type="button" id="pvp-res-ok">Lanjut</button>
    </div>`;
    if (r.promoted) {
      window.clearTimeout(this.promoTimer);
      this.promoTimer = window.setTimeout(() => {
        /* celebration 2–4s */
      }, 3000);
    }
    this.result.querySelector("#pvp-res-ok")?.addEventListener("click", () => {
      this.result.hidden = true;
    });
  }

  clear(): void {
    this.root.hidden = true;
    this.view = null;
  }
}

function emptyLobby(): PvpLobbyOut {
  return {
    modes: [],
    profile: { rating: 1000, rank: "BRONZE", rankName: "UNRANKED", wins: 0, losses: 0, winRate: 0, placementLeft: 5, winStreak: 0, lossStreak: 0, highestRank: "", battleToken: 0, season: "", seasonId: "" },
    season: { id: "s1", name: "Dawn Awakening", number: 1, start: "", end: "", weeks: 8 },
    shop: [],
    rewards: [],
    history: [],
  };
}

function isBattleground(map: string): boolean {
  return map === "valley-of-dawn" || map === "dawn-arena";
}

function mapLabel(map: string): string {
  if (map === "dawn-arena" || map === "valley-of-dawn") return "Dawn Arena";
  if (map === "mistwood-battlefield") return "Mistwood Battlefield";
  return "Celestial Courtyard";
}

function pingColor(ms: number): string {
  if (ms < 80) return `<span class="pvp-ping green">${ms}ms</span>`;
  if (ms < 140) return `<span class="pvp-ping yellow">${ms}ms</span>`;
  return `<span class="pvp-ping red">${ms}ms</span>`;
}

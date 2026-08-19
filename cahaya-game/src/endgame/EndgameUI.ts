import { escapeHtml } from "../dialogue/DialogueUI";
import type { EndgameState } from "../network/NetworkMessage";

const TABS = [
  "DASHBOARD",
  "CHALLENGES",
  "DUNGEON",
  "RAID",
  "WORLD BOSS",
  "PVP",
  "SEASON",
  "ACHIEVEMENTS",
  "COLLECTIONS",
  "EXPLORATION",
  "COSMETICS",
] as const;

type EndgameTab = (typeof TABS)[number];

export class EndgameUI {
  readonly overlay: HTMLElement;
  blocking = false;
  state: EndgameState | null = null;
  tab: EndgameTab = "DASHBOARD";
  onClaimDaily: ((id: string) => void) | null = null;
  onClaimWeekly: ((id: string) => void) | null = null;
  onClaimChallenge: ((id: string) => void) | null = null;
  onClaimSeason: ((level: number) => void) | null = null;
  onClose: (() => void) | null = null;

  constructor(host: HTMLElement) {
    this.overlay = document.createElement("div");
    this.overlay.className = "endgame-overlay";
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

  apply(data: EndgameState): void {
    this.state = data;
    if (this.blocking) this.render();
  }

  render(): void {
    const s = this.state;
    const locked = s && s.unlocked === false;
    const nav = TABS.map(
      (t) => `<button type="button" class="eg-tab${this.tab === t ? " on" : ""}" data-tab="${t}">${t}</button>`,
    ).join("");
    this.overlay.innerHTML = `<div class="eg-shell">
      <aside class="eg-side">
        <div class="eg-kicker">HALL OF HORIZON</div>
        <h2>ENDGAME</h2>
        <p class="eg-sub">${escapeHtml(s?.seasonName || "DAWN OF THE 33")}</p>
        <nav class="eg-nav">${nav}</nav>
        <button type="button" class="eg-close" id="eg-close">Tutup</button>
      </aside>
      <div class="eg-main">
        ${locked ? `<p class="eg-lock">Selesaikan cerita atau capai level yang ditentukan untuk membuka Hall of Horizon.</p>` : this.body(s)}
      </div>
    </div>`;
    this.overlay.querySelectorAll("[data-tab]").forEach((el) => {
      el.addEventListener("click", () => {
        this.tab = ((el as HTMLElement).dataset.tab || "DASHBOARD") as EndgameTab;
        this.render();
      });
    });
    this.overlay.querySelector("#eg-close")?.addEventListener("click", () => {
      this.close();
      this.onClose?.();
    });
    this.overlay.querySelectorAll("[data-daily]").forEach((el) => {
      el.addEventListener("click", () => this.onClaimDaily?.((el as HTMLElement).dataset.daily || ""));
    });
    this.overlay.querySelectorAll("[data-weekly]").forEach((el) => {
      el.addEventListener("click", () => this.onClaimWeekly?.((el as HTMLElement).dataset.weekly || ""));
    });
    this.overlay.querySelectorAll("[data-ch]").forEach((el) => {
      el.addEventListener("click", () => this.onClaimChallenge?.((el as HTMLElement).dataset.ch || ""));
    });
    this.overlay.querySelector("#eg-season-claim")?.addEventListener("click", () => {
      const lv = Number((this.overlay.querySelector("#eg-season-claim") as HTMLElement).dataset.level || 0);
      this.onClaimSeason?.(lv);
    });
  }

  private body(s: EndgameState | null): string {
    switch (this.tab) {
      case "CHALLENGES":
        return this.questList("Challenges", s?.challenges, "ch");
      case "DUNGEON":
        return `<h3>Horizon Challenge</h3><p>Dungeon endgame dengan modifier acak dari server.</p>
          <p>Best score: ${s?.horizon?.best ?? 0} · Week ${escapeHtml(String(s?.horizon?.week || ""))}</p>
          ${this.board(s?.horizon?.board)}`;
      case "RAID":
        return this.phase29Raid(s);
      case "WORLD BOSS":
        return this.phase29WorldBoss(s);
      case "PVP":
        return `<h3>Arena</h3><p>Arena Master tetap di Dawn City. Musim PvP terpisah dari Season XP endgame.</p>`;
      case "SEASON":
        return this.season(s);
      case "ACHIEVEMENTS":
        return `<h3>Achievements</h3><ul class="eg-list">${(s?.achievements || []).map((a) => `<li>${escapeHtml(a)}</li>`).join("") || "<li>Belum ada</li>"}</ul>`;
      case "COLLECTIONS":
        return this.collection(s);
      case "EXPLORATION":
        return `<h3>Lore Book</h3><ul class="eg-list">${(s?.lore || []).map((l) => `<li><strong>${escapeHtml(l.title)}</strong><span>${escapeHtml(l.text || "")}</span></li>`).join("") || "<li>Jelajahi dunia untuk menemukan fragmen.</li>"}</ul>`;
      case "COSMETICS":
        return `<h3>Cosmetics</h3><p>Preview tidak memberikan item. Equip hanya jika owned.</p>
          <ul class="eg-list">${(s?.cosmetics || []).map((c) => `<li>${escapeHtml(c)}</li>`).join("") || "<li>Kosong</li>"}</ul>`;
      default:
        return this.dashboard(s);
    }
  }

  private dashboard(s: EndgameState | null): string {
    const xp = s?.seasonXP ?? 0;
    const need = s?.seasonXPNeed || 100;
    const pct = Math.min(100, Math.round((xp / need) * 100));
    const c = s?.community;
    const p29 = s?.phase29;
    return `<div class="eg-dash">
      <article class="eg-card"><div class="eg-kicker">VERSION</div><strong>${escapeHtml(s?.version || "1.0.0-beta")}</strong><span>${escapeHtml(s?.phase || "30/30")}</span></article>
      <article class="eg-card"><div class="eg-kicker">LEVEL</div><strong>${s?.level ?? 1}</strong></article>
      <article class="eg-card"><div class="eg-kicker">SEASON LEVEL</div><strong>${s?.seasonLevel ?? 1}</strong>
        <div class="eg-bar"><i style="width:${pct}%"></i></div><span>${xp} / ${need} Season XP</span></article>
      <article class="eg-card eg-wide"><div class="eg-kicker">DAILY</div>${this.questList("", s?.daily, "daily")}</article>
      <article class="eg-card eg-wide"><div class="eg-kicker">WEEKLY</div>${this.questList("", s?.weekly, "weekly")}</article>
      <article class="eg-card"><div class="eg-kicker">CHALLENGES</div><strong>${(s?.challenges || []).filter((q) => q.claimed).length}/${(s?.challenges || []).length}</strong></article>
      <article class="eg-card"><div class="eg-kicker">ACHIEVEMENTS</div><strong>${(s?.achievements || []).length}</strong></article>
      <article class="eg-card"><div class="eg-kicker">COLLECTION</div>
        <p>Aura ${s?.collection?.aura ?? 0}/${s?.collection?.auraTotal ?? 0}</p>
        <p>Mount ${s?.collection?.mount ?? 0}/${s?.collection?.mountTotal ?? 0}</p>
        <p>Titles ${s?.collection?.titles ?? 0}/${s?.collection?.titleTotal ?? 0}</p></article>
      <article class="eg-card eg-wide"><div class="eg-kicker">WORLD PROGRESS</div>
        <strong>${c?.name || "Light Festival"}</strong>
        <div class="eg-bar"><i style="width:${Math.min(100, Math.round(100 * (c?.points || 0) / (c?.target || 1)))}%"></i></div>
        <span>${c?.points ?? 0} / ${c?.target ?? 1000000}</span></article>
      <article class="eg-card eg-wide"><div class="eg-kicker">PHASE 29</div>
        <strong>${escapeHtml(p29?.mainStory?.title || "Petualangan Menuju Cahaya")}</strong>
        <p>${escapeHtml(p29?.mainStory?.objective || "Lanjutkan perjalananmu.")}</p>
        <p>World Event: ${escapeHtml(p29?.worldEvent?.name || "Bayangan Ing Alas")} · ${p29?.worldEvent?.globalProgress ?? 0}%</p>
        <p>Final Journey: ${escapeHtml(p29?.finalJourney?.questName || "Lampah Pungkasan.")} · party ${p29?.finalJourney?.partyRequired ?? 10}+ </p>
      </article>
    </div>`;
  }

  private season(s: EndgameState | null): string {
    const xp = s?.seasonXP ?? 0;
    const need = s?.seasonXPNeed || 100;
    const cur = s?.currentReward;
    const next = s?.nextReward;
    return `<h3>${escapeHtml(s?.seasonName || "DAWN OF THE 33")}</h3>
      <p>Level ${s?.seasonLevel ?? 1} · Free Track · Berakhir ${escapeHtml(s?.seasonEnd || "")}</p>
      <div class="eg-bar"><i style="width:${Math.min(100, Math.round((xp / need) * 100))}%"></i></div>
      <p>XP ${xp} / ${need}</p>
      <p>Current: ${escapeHtml(String(cur?.kind || "—"))} ${escapeHtml(String(cur?.id || ""))}</p>
      <p>Next: ${escapeHtml(String(next?.kind || "—"))} ${escapeHtml(String(next?.id || ""))}</p>
      <button type="button" class="eg-btn" id="eg-season-claim" data-level="${s?.seasonLevel ?? 1}">Claim Level Reward</button>
      <h4>History</h4>
      <ul class="eg-list">${(s?.history || []).map((h) => `<li>${escapeHtml(h.name)} · Lv ${h.level} · Rank ${escapeHtml(h.rank || "—")}</li>`).join("") || "<li>Belum ada</li>"}</ul>`;
  }

  private collection(s: EndgameState | null): string {
    const c = s?.collection;
    return `<h3>Collection</h3>
      <p>Aura ${c?.aura ?? 0}/${c?.auraTotal ?? 0}</p>
      <p>Mount ${c?.mount ?? 0}/${c?.mountTotal ?? 0}</p>
      <p>Titles ${c?.titles ?? 0}/${c?.titleTotal ?? 0}</p>
      <p>Learning: ${s?.learning?.correct ?? 0}/${s?.learning?.answered ?? 0} benar · akurasi ${s?.learning?.accuracy ?? 0}%</p>`;
  }

  private calendar(s: EndgameState | null): string {
    const cal = s?.calendar;
    const today = cal?.today as { worldBossName?: string; event?: string } | undefined;
    return `<p>Timezone server: ${escapeHtml(String(cal?.timezone || "UTC"))}</p>
      <h4>TODAY</h4><p>${escapeHtml(today?.worldBossName || "—")} · ${escapeHtml(today?.event || "—")}</p>
      <h4>THIS WEEK</h4>${this.board((cal?.week || []) as Array<{ name?: string; rank?: number; score?: number; level?: number }>)}
      <h4>UPCOMING</h4><ul class="eg-list">${((cal?.upcoming || []) as Array<{ name?: string; when?: string }>).map((u) => `<li>${escapeHtml(u.when || "")} ${escapeHtml(u.name || "")}</li>`).join("")}</ul>`;
  }

  private phase29Raid(s: EndgameState | null): string {
    const raid = s?.phase29?.raid;
    const finalJ = s?.phase29?.finalJourney;
    return `<h3>${escapeHtml(raid?.name || "Lampah Pungkasan")}</h3>
      <p>Raid 10-20 pemain. Checkpoint ${raid?.checkpoint ? "aktif" : "mati"} · Difficulty ${(raid?.difficulties || []).join(" / ") || "HARD / NIGHTMARE"}</p>
      <p>Bosses: ${(raid?.bosses || []).map((b) => escapeHtml(b)).join(" · ") || "3 bosses"}</p>
      <p>Final Gate: ${finalJ?.gateLocked ? "LOCKED" : "OPEN"} · Quest ${escapeHtml(finalJ?.questName || "Lampah Pungkasan.")}</p>
      <p>Reward raid, title, cosmetic, lan checkpoint disimpan server-side.</p>
      ${this.multiBoard("LEADERBOARD", s?.leaderboards)}
    `;
  }

  private phase29WorldBoss(s: EndgameState | null): string {
    const boss = s?.phase29?.worldBoss;
    const event = s?.phase29?.worldEvent;
    const seasonal = s?.phase29?.seasonal;
    return `<h3>${escapeHtml(boss?.name || "RAKSA CAHAYA PETENG")}</h3>
      <p>Status: ${escapeHtml(boss?.state || "SCHEDULED")} · limit reward ${escapeHtml(boss?.dailyLimit || "server_authoritative")}</p>
      <p>World Event: ${escapeHtml(event?.name || "Bayangan Ing Alas")} · ${event?.active ? "ACTIVE" : "ROTATING"} · ${event?.globalProgress ?? 0}%</p>
      <p>Season: ${escapeHtml(seasonal?.seasonName || s?.seasonName || "")} · Festival ${escapeHtml(seasonal?.festival || "Festival Cahaya")}</p>
      ${this.calendar(s)}`;
  }

  private multiBoard(title: string, rows?: EndgameState["leaderboards"]): string {
    const section = (label: string, data: unknown[] | undefined) => `<h4>${label}</h4>${this.board(data as Array<{ rank?: number; name?: string; score?: number; level?: number }> | undefined)}`;
    return `<div class="eg-multi-board">
      <h3>${escapeHtml(title)}</h3>
      ${section("Level", rows?.level as unknown[] | undefined)}
      ${section("Combat", rows?.combat as unknown[] | undefined)}
      ${section("Boss Defeated", rows?.bossDefeated as unknown[] | undefined)}
      ${section("Quest", rows?.quest as unknown[] | undefined)}
      ${section("Crafting", rows?.crafting as unknown[] | undefined)}
      ${section("Gathering", rows?.gathering as unknown[] | undefined)}
      ${section("Guild", rows?.guilds as unknown[] | undefined)}
    </div>`;
  }

  private questList(title: string, rows: EndgameState["daily"], kind: "daily" | "weekly" | "ch"): string {
    const attr = kind === "ch" ? "data-ch" : kind === "weekly" ? "data-weekly" : "data-daily";
    return `<h3>${title}</h3><div class="eg-quests">${(rows || [])
      .map(
        (q) => `<article class="eg-q">
        <strong>${escapeHtml(q.title)}</strong>
        <span>${q.tier ? q.tier + " · " : ""}${q.progress}/${q.count}</span>
        <button type="button" class="eg-btn" ${attr}="${q.id}" ${q.claimed || q.progress < q.count ? "disabled" : ""}>Claim</button>
      </article>`,
      )
      .join("")}</div>`;
  }

  private board(rows?: Array<{ rank?: number; name?: string; score?: number; level?: number }>): string {
    if (!rows?.length) return "<p>Belum ada peringkat.</p>";
    return `<table class="eg-table"><tr><th>Rank</th><th>Player</th><th>Score</th><th>Level</th></tr>${rows
      .map((r, i) => `<tr><td>${r.rank || i + 1}</td><td>${escapeHtml(r.name || "")}</td><td>${r.score ?? 0}</td><td>${r.level ?? 0}</td></tr>`)
      .join("")}</table>`;
  }
}

import { escapeHtml } from "../dialogue/DialogueUI";
import type { DungeonListOut, DungeonOffer, DungeonReadyOut, DungeonView, QueueView } from "../network/NetworkMessage";
import { DungeonMap } from "./DungeonMap";
import { DungeonObjectiveUI } from "./DungeonObjectiveUI";
import { DungeonRewardUI } from "./DungeonRewardUI";
import { DungeonTimerUI } from "./DungeonTimerUI";

export class DungeonUI {
  readonly overlay: HTMLElement;
  readonly timer: DungeonTimerUI;
  readonly objective: DungeonObjectiveUI;
  readonly rewards: DungeonRewardUI;
  readonly map: DungeonMap;
  readonly reviveBtn: HTMLButtonElement;
  readonly voteBar: HTMLElement;
  blocking = false;
  instanceId = "";
  view: DungeonView | null = null;
  queue: QueueView | null = null;
  selfId = "";
  onEnter: ((dungeonId: string, difficulty?: string) => void) | null = null;
  onQueue: ((dungeonId: string, role: string, difficulty?: string) => void) | null = null;
  onQueueLeave: (() => void) | null = null;
  onReady: ((ready: boolean) => void) | null = null;
  onLeave: (() => void) | null = null;
  onRetry: (() => void) | null = null;
  onRevive: ((targetId: string) => void) | null = null;
  onVote: ((vote: string) => void) | null = null;
  onSkip: (() => void) | null = null;
  onExchange: ((shopId: string) => void) | null = null;
  private difficulty = "NORMAL";
  private role = "FLEX";

  constructor(host: HTMLElement) {
    this.overlay = document.createElement("div");
    this.overlay.className = "dun-overlay";
    this.overlay.hidden = true;
    host.appendChild(this.overlay);
    this.timer = new DungeonTimerUI(host);
    this.objective = new DungeonObjectiveUI(host);
    this.rewards = new DungeonRewardUI(host);
    this.map = new DungeonMap(host);
    this.reviveBtn = document.createElement("button");
    this.reviveBtn.type = "button";
    this.reviveBtn.className = "dun-revive-btn";
    this.reviveBtn.hidden = true;
    this.reviveBtn.textContent = "REVIVE";
    host.appendChild(this.reviveBtn);
    this.voteBar = document.createElement("div");
    this.voteBar.className = "dun-vote";
    this.voteBar.hidden = true;
    this.voteBar.innerHTML = `<button type="button" data-vote="restart">VOTE RESTART</button><button type="button" data-vote="leave">VOTE LEAVE</button>`;
    host.appendChild(this.voteBar);
    this.overlay.addEventListener("click", (e) => {
      const t = (e.target as HTMLElement).closest("[data-dun]") as HTMLElement | null;
      if (!t) return;
      if (t.dataset.dun === "enter" && t.dataset.id) this.onEnter?.(t.dataset.id, this.difficulty);
      if (t.dataset.dun === "queue" && t.dataset.id) this.onQueue?.(t.dataset.id, t.dataset.role || this.role, this.difficulty);
      if (t.dataset.dun === "leaveq") this.onQueueLeave?.();
      if (t.dataset.dun === "ready") this.onReady?.(true);
      if (t.dataset.dun === "notready") this.onReady?.(false);
      if (t.dataset.dun === "leave") {
        if (window.confirm("Yakin meninggalkan dungeon?")) this.onLeave?.();
      }
      if (t.dataset.dun === "retry") this.onRetry?.();
      if (t.dataset.dun === "skip") this.onSkip?.();
      if (t.dataset.dun === "close") this.hideOverlay();
      if (t.dataset.dun === "diff" && t.dataset.id) this.difficulty = t.dataset.id;
      if (t.dataset.dun === "role" && t.dataset.id) this.role = t.dataset.id;
      if (t.dataset.dun === "exchange" && t.dataset.id) this.onExchange?.(t.dataset.id);
    });
    this.reviveBtn.addEventListener("click", () => {
      const target = this.nearbyDowned();
      if (target) this.onRevive?.(target);
    });
    this.voteBar.addEventListener("click", (e) => {
      const t = (e.target as HTMLElement).closest("[data-vote]") as HTMLElement | null;
      if (!t?.dataset.vote) return;
      if (t.dataset.vote === "leave" && !window.confirm("Yakin meninggalkan dungeon?")) return;
      this.onVote?.(t.dataset.vote);
    });
  }

  get overlayOpen(): boolean {
    return !this.overlay.hidden;
  }

  showOffer(offer: DungeonOffer): void {
    this.blocking = true;
    this.overlay.hidden = false;
    const min = offer.minPlayers || 1;
    this.overlay.innerHTML = `
      <div class="dun-card">
        <div class="dun-kicker">${escapeHtml(offer.kind || offer.chapterId || "DUNGEON")}</div>
        <h3>ENTER DUNGEON</h3>
        <p>${escapeHtml(offer.name)}</p>
        <p>${escapeHtml(offer.description || "")}</p>
        <p>Level ${offer.recommendedLevel} · ${min}–${offer.maxPlayers} pemain · ${Math.round(offer.timeLimit / 60)} menit · ${escapeHtml(offer.difficulty || "NORMAL")}</p>
        <button type="button" class="dun-btn" data-dun="enter" data-id="${escapeHtml(offer.dungeonId)}">ENTER DUNGEON</button>
        <button type="button" data-dun="queue" data-id="${escapeHtml(offer.dungeonId)}" data-role="FLEX">QUEUE MATCHMAKING</button>
        <button type="button" data-dun="close">Tutup</button>
      </div>`;
  }

  showList(list: DungeonListOut): void {
    this.blocking = true;
    this.overlay.hidden = false;
    const rows = (list.dungeons ?? [])
      .map((d) => {
        const min = d.minPlayers || 1;
        const reward = d.rewards ? `EXP ${d.rewards.exp || 0}` : "";
        const lock = d.status === "LOCKOUT" && d.lockoutLabel ? ` · reset ${escapeHtml(d.lockoutLabel)}` : "";
        return `<div class="dun-row">
          <div><b>${escapeHtml(d.name)}</b> <span>${escapeHtml(d.kind || "")} LV ${d.recommendedLevel} · ${min}–${d.maxPlayers} · ${Math.round((d.timeLimit || 1800) / 60)}m · ${escapeHtml(d.difficulty || "NORMAL")} · ${escapeHtml(d.status)}${lock} · ${reward}</span></div>
          <button type="button" data-dun="enter" data-id="${escapeHtml(d.dungeonId)}">Enter</button>
          <button type="button" data-dun="queue" data-id="${escapeHtml(d.dungeonId)}">Queue</button>
        </div>`;
      })
      .join("");
    const q = list.queue ? `<p>Queue: ${escapeHtml(list.queue.name || list.queue.dungeonId)} (${list.queue.players || 0}/${list.queue.maxPlayers || 4})</p><button type="button" data-dun="leaveq">LEAVE QUEUE</button>` : "";
    const hist = (list.history ?? []).slice(0, 5).map((h) => `<div>${escapeHtml(h.nameDun || h.dungeonId)} · ${h.rating} · ${h.elapsed}s · ${h.deaths} death</div>`).join("") || "<div>Belum ada riwayat.</div>";
    const board = (list.board ?? []).slice(0, 5).map((b) => `<div>${escapeHtml(b.player)}${b.guild ? ` [${escapeHtml(b.guild)}]` : ""} · ${escapeHtml(b.name)} · ${b.elapsed}s</div>`).join("") || "<div>Belum ada leaderboard.</div>";
    const shop = (list.raidShop ?? []).map((s) => `<button type="button" data-dun="exchange" data-id="${escapeHtml(s.id)}">${escapeHtml(s.name)} · ${s.cost} RT</button>`).join("");
    this.overlay.innerHTML = `<div class="dun-card dun-wide"><div class="dun-kicker">DUNGEON FINDER</div><h3>Dungeon & Raid</h3>
      <p>Difficulty: <button type="button" data-dun="diff" data-id="NORMAL">NORMAL</button> <button type="button" data-dun="diff" data-id="HARD">HARD</button> · Role: <button type="button" data-dun="role" data-id="DPS">DPS</button> <button type="button" data-dun="role" data-id="TANK">TANK</button> <button type="button" data-dun="role" data-id="SUPPORT">SUPPORT</button> <button type="button" data-dun="role" data-id="FLEX">FLEX</button></p>
      <p>Raid lockout reset: ${escapeHtml(list.lockoutLabel || "-")} · Raid Token: ${list.raidTokens || 0}</p>
      ${q}<div class="dun-list">${rows}</div>
      <h4>History</h4><div class="dun-list">${hist}</div>
      <h4>Fastest Clear</h4><div class="dun-list">${board}</div>
      <h4>Raid Token Shop</h4><div class="dun-list">${shop}</div>
      <button type="button" data-dun="close">Tutup</button></div>`;
  }

  showQueue(view: QueueView): void {
    this.queue = view;
    if (view.state === "CANCELLED" || view.state === "IDLE") return;
    this.blocking = true;
    this.overlay.hidden = false;
    this.overlay.innerHTML = `
      <div class="dun-card">
        <div class="dun-kicker">${escapeHtml(view.state)}</div>
        <h3>QUEUE</h3>
        <p>${escapeHtml(view.name || view.dungeonId)}</p>
        <p>${view.players || 0} / ${view.maxPlayers || 4} pemain · role ${escapeHtml(view.role || "FLEX")}</p>
        <button type="button" data-dun="leaveq">LEAVE QUEUE</button>
      </div>`;
  }

  showReady(check: DungeonReadyOut): void {
    this.blocking = true;
    this.overlay.hidden = false;
    if (check.cancelled) {
      this.hideOverlay();
      return;
    }
    const rows = (check.members ?? [])
      .map((m) => `<div>${escapeHtml(m.playerId)} — ${m.ready ? "READY" : "NOT READY"}</div>`)
      .join("");
    this.overlay.innerHTML = `
      <div class="dun-card">
        <div class="dun-kicker">READY CHECK</div>
        <h3>${escapeHtml(check.dungeonId)}</h3>
        <div class="dun-ready">${rows}</div>
        <button type="button" class="dun-btn" data-dun="ready">READY</button>
        <button type="button" data-dun="notready">NOT READY</button>
      </div>`;
  }

  showLoading(name: string): void {
    this.blocking = true;
    this.overlay.hidden = false;
    this.overlay.innerHTML = `
      <div class="dun-card">
        <div class="dun-kicker">INSTANCE</div>
        <h3>ENTERING DUNGEON...</h3>
        <p>${escapeHtml(name)}</p>
        <button type="button" data-dun="skip">SKIP INTRO</button>
      </div>`;
  }

  applyState(view: DungeonView, selfX = 0, selfZ = 0): void {
    this.view = view;
    this.instanceId = view.instanceId;
    if (view.state === "STARTING" || view.state === "LOADING") this.showLoading(view.name);
    else if (view.state !== "COMPLETED" && view.state !== "FAILED") this.hideOverlay();
    this.timer.set(view.timeLeft, view.state === "ACTIVE" || view.state === "BOSS" || view.state === "LOADING");
    this.objective.apply(view);
    this.map.apply(view, selfX, selfZ);
    this.voteBar.hidden = !(view.state === "ACTIVE" || view.state === "BOSS");
    this.updateRevive();
    if (view.state === "COMPLETED") this.showComplete(view);
    if (view.state === "FAILED") this.showFail(view);
  }

  showComplete(view: DungeonView): void {
    this.blocking = true;
    this.overlay.hidden = false;
    this.voteBar.hidden = true;
    this.reviveBtn.hidden = true;
    const loot = (view.loot ?? []).map((it) => `<div>${escapeHtml(it.name)} x${it.qty}</div>`).join("");
    const kind = view.kind === "RAID" ? "RAID CLEAR" : "DUNGEON COMPLETE";
    this.overlay.innerHTML = `
      <div class="dun-card">
        <div class="dun-kicker">RATING ${escapeHtml(view.rating || "C")}</div>
        <h3>${kind}</h3>
        <p>${escapeHtml(view.name)}</p>
        <p>Personal loot · anti-duplicate claim</p>
        <div class="dun-loot">${loot}</div>
        <button type="button" class="dun-btn" data-dun="retry">RETRY</button>
        <button type="button" data-dun="leave">LEAVE</button>
      </div>`;
  }

  showFail(view: DungeonView): void {
    this.blocking = true;
    this.overlay.hidden = false;
    this.voteBar.hidden = true;
    this.overlay.innerHTML = `
      <div class="dun-card">
        <div class="dun-kicker">FAILED</div>
        <h3>DUNGEON FAILED</h3>
        <p>${escapeHtml(view.toast || view.name)}</p>
        <button type="button" class="dun-btn" data-dun="retry">RETRY</button>
        <button type="button" data-dun="leave">LEAVE</button>
      </div>`;
  }

  showWipe(view: DungeonView): void {
    this.view = view;
    this.blocking = true;
    this.overlay.hidden = false;
    this.timer.set(view.timeLeft, true);
    this.objective.apply(view);
    this.overlay.innerHTML = `
      <div class="dun-card">
        <div class="dun-kicker">WIPE ${view.wipeCount || 1}</div>
        <h3>PARTY DEFEATED</h3>
        <p>${escapeHtml(view.toast || "Kembali ke checkpoint.")}</p>
        <button type="button" class="dun-btn" data-dun="close">REVIVE</button>
        <button type="button" data-dun="leave">RETURN</button>
      </div>`;
  }

  hideOverlay(): void {
    this.blocking = false;
    this.overlay.hidden = true;
  }

  nearbyDowned(): string {
    const me = (this.view?.members ?? []).find((m) => m.playerId === this.selfId);
    const downed = (this.view?.members ?? []).find((m) => m.downed && m.playerId !== this.selfId && m.distance <= 3);
    if (!me || me.dead || me.downed) return "";
    return downed?.playerId || "";
  }

  updateRevive(): void {
    const target = this.nearbyDowned();
    const downedSelf = (this.view?.members ?? []).some((m) => m.playerId === this.selfId && m.downed);
    this.reviveBtn.hidden = !target && !downedSelf;
    if (downedSelf) {
      this.reviveBtn.textContent = "DOWNED";
      this.reviveBtn.disabled = true;
    } else {
      const prog = (this.view?.members ?? []).find((m) => m.playerId === target)?.reviveProgress || 0;
      this.reviveBtn.textContent = prog > 0 ? `REVIVE ${prog}%` : "REVIVE";
      this.reviveBtn.disabled = !target;
    }
  }

  clear(): void {
    this.view = null;
    this.instanceId = "";
    this.hideOverlay();
    this.timer.set(0, false);
    this.objective.apply(null);
    this.map.apply(null, 0, 0);
    this.rewards.hide();
    this.reviveBtn.hidden = true;
    this.voteBar.hidden = true;
  }
}

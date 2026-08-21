import { truncateName, avatarSpriteHtml } from "../assets/registry";
import { loadPrefs } from "./prefs";
import { toast } from "./chrome";
import { bagTotal, pointsFromPosition, powerIconHtml, type PowerBag } from "../game/powers";

export type PremiumSeat = {
  id: string;
  username: string;
  color: string;
  position: number;
  bag?: PowerBag;
  isConnected?: boolean;
};

export type PremiumDockHandlers = {
  onRoll?: () => void;
  onExit?: () => void;
  onHistory?: () => void;
  onInventory?: () => void;
  onChat?: () => void;
  onSettings?: () => void;
};

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export function activateBoardTheme(): () => void {
  document.body.classList.add("board-theme-fantasy");
  document.documentElement.dataset.board = "fantasy";
  return () => {
    document.body.classList.remove("board-theme-fantasy");
    delete document.documentElement.dataset.board;
  };
}

/** Shell papan — rapi, terbaca. */
export function premiumBoardShell(extra?: string): string {
  return `
    <div class="premium-board-root mountain-mode">
      <div class="premium-bg" aria-hidden="true"></div>
      <h1 class="premium-title">ULAR TANGGA NUSANTARA</h1>

      <header class="premium-header">
        <div class="premium-player-strip" id="player-strip" role="list"></div>
        <div class="premium-top-actions">
          <button type="button" class="premium-icon-btn" data-premium="settings" aria-label="Pengaturan">⚙</button>
          <button type="button" class="premium-icon-btn" data-premium="chat" aria-label="Chat">💬</button>
        </div>
      </header>

      <div class="premium-main game-container">
        <div class="premium-board-stage">
          <div class="iso-wrap">
            <div class="board-frame premium-frame" id="board-frame">
              <div class="board-playfield" id="board-playfield">
                <div class="start-pad" aria-label="Zona start di luar papan">
                  <span class="start-pad-label">START</span>
                  <span class="start-pad-hint">Masuk →</span>
                </div>
                <div class="board-grid" id="board-grid"></div>
                <svg class="route-layer" id="route-layer" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true"></svg>
                <div class="token-layer" id="token-layer"></div>
              </div>
            </div>
          </div>
        </div>

        <aside class="premium-side control-panel">
          <div class="premium-turn-hud status" id="turn-hud" aria-live="polite"></div>
          <div class="dice premium-dice" id="dice" data-face="1">1</div>
          <button type="button" class="dock-roll is-ready" id="roll-btn" aria-label="Kocok dadu">
            <span class="dock-roll-label">KOCOK DADU</span>
          </button>
          <button type="button" class="dock-btn" data-dock="inventory"><span class="dock-ico">🎒</span> INVENTORY</button>
          <button type="button" class="dock-btn" data-dock="history"><span class="dock-ico">📄</span> RIWAYAT</button>
          <button type="button" class="dock-btn dock-exit" data-dock="exit"><span class="dock-ico">🚪</span> KELUAR</button>
        </aside>
      </div>

      <p class="conn-line premium-conn" id="conn-line" hidden></p>
      <p class="dice-result premium-dice-result" id="dice-result"></p>
      <p class="nt-error premium-error" id="board-error" hidden role="alert"></p>
      <div id="power-host"></div>
      <aside class="board-chat" id="board-chat" hidden>
        <header class="board-chat-head">
          <strong>💬 Chat</strong>
          <button type="button" class="power-close" data-chat="close" aria-label="Tutup">✕</button>
        </header>
        <div class="board-chat-log" id="board-chat-log"></div>
        <form class="board-chat-form" id="board-chat-form">
          <input name="msg" maxlength="200" placeholder="Tulis pesan…" autocomplete="off" />
          <button type="submit" class="nt-btn">Kirim</button>
        </form>
      </aside>
      ${extra || ""}
    </div>`;
}

export function paintPremiumStrip(el: HTMLElement | null, seats: PremiumSeat[], turnId: string): void {
  if (!el) return;
  el.innerHTML = seats
    .map((s, i) => {
      const pts = pointsFromPosition(s.position);
      const bag = s.bag;
      const items = bag && bagTotal(bag) > 0
        ? `<p class="pcard-bag">${bag.bomb ? `${powerIconHtml("bomb", "power-ico-img sm")}<span>×${bag.bomb}</span> ` : ""}${bag.thunder ? `${powerIconHtml("thunder", "power-ico-img sm")}<span>×${bag.thunder}</span> ` : ""}${bag.superman ? `${powerIconHtml("superman", "power-ico-img sm")}<span>×${bag.superman}</span>` : ""}</p>`
        : "";
      const on = s.id === turnId;
      const initial = esc(s.username.replace(/^🤖\s*/, "").slice(0, 1).toUpperCase());
      return `
        <article class="premium-pcard ${on ? "is-turn" : ""} ${s.isConnected === false ? "is-off" : ""}" role="listitem" style="--pc:${s.color}">
          <div class="pcard-avatar is-sprite">${avatarSpriteHtml(i, s.username)}<span class="pcard-fallback" aria-hidden="true">${initial}</span></div>
          <div class="pcard-body">
            <strong>${esc(truncateName(s.username, 10))}</strong>
            <p class="pcard-stars"><span class="star">★</span> ${pts} poin</p>
            ${items}
          </div>
        </article>`;
    })
    .join("");
}

export function paintPremiumTurnHud(
  el: HTMLElement | null,
  username: string,
  opts?: { timerSec?: number; position?: number; isMine?: boolean },
): void {
  if (!el) return;
  const name = truncateName(username.replace(/^🤖\s*/, ""), 14);
  const t = opts?.timerSec;
  const timer = t !== undefined ? ` · ⏱ ${t}` : "";
  const pts = opts?.position !== undefined ? pointsFromPosition(opts.position) : undefined;
  el.innerHTML = `
    <p class="turn-hud-kicker">GILIRAN</p>
    <p class="turn-hud-name ${opts?.isMine ? "is-mine" : ""}">${esc(name)}</p>
    ${pts !== undefined
      ? `<p class="turn-hud-pos">${pts <= 0 ? "0 poin · luar papan" : `${pts} poin · kotak ${opts!.position}`}${timer}</p>`
      : `<p class="turn-hud-pos">${timer || "Siap bermain"}</p>`}`;
  el.style.borderLeftColor = opts?.isMine ? "#00ffff" : "#ff00ff";
}

export function bindPremiumDock(root: ParentNode, handlers: PremiumDockHandlers): void {
  root.querySelector('[data-dock="exit"]')?.addEventListener("click", () => handlers.onExit?.());
  root.querySelector('[data-dock="history"]')?.addEventListener("click", () => {
    if (handlers.onHistory) handlers.onHistory();
    else toast("Riwayat match tersedia di profil.", "info");
  });
  root.querySelector('[data-dock="inventory"]')?.addEventListener("click", () => {
    if (handlers.onInventory) handlers.onInventory();
    else toast("Inventory kosong.", "info");
  });
  root.querySelector('[data-premium="chat"]')?.addEventListener("click", () => {
    if (handlers.onChat) handlers.onChat();
    else {
      const panel = root.querySelector<HTMLElement>("#board-chat");
      if (panel) panel.hidden = !panel.hidden;
    }
  });
  root.querySelector('[data-chat="close"]')?.addEventListener("click", () => {
    const panel = root.querySelector<HTMLElement>("#board-chat");
    if (panel) panel.hidden = true;
  });
  root.querySelector('[data-premium="settings"]')?.addEventListener("click", () => {
    if (handlers.onSettings) handlers.onSettings();
    else document.dispatchEvent(new CustomEvent("ular:open-settings"));
  });
  const roll = root.querySelector<HTMLElement>("#roll-btn");
  const fireRoll = (): void => {
    if (!roll || roll.classList.contains("is-disabled") || !handlers.onRoll) return;
    handlers.onRoll();
  };
  roll?.addEventListener("click", fireRoll);
  roll?.addEventListener("keydown", (e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      fireRoll();
    }
  });
}

export function applyPremiumRollState(btn: HTMLElement | null, label: string, enabled: boolean): void {
  if (!btn) return;
  btn.classList.toggle("is-ready", enabled);
  btn.classList.toggle("is-disabled", !enabled);
  btn.setAttribute("aria-disabled", enabled ? "false" : "true");
  btn.tabIndex = enabled ? 0 : -1;
  const lab = btn.querySelector(".dock-roll-label");
  if (lab) lab.textContent = label.toUpperCase().includes("ROLL") || label.toUpperCase().includes("KOCOK")
    ? "KOCOK DADU"
    : label.toUpperCase();
  if (enabled) {
    if (lab) lab.textContent = "KOCOK DADU";
  }
  btn.setAttribute("aria-label", label);
}

export function usePremiumBoardLayout(): boolean {
  return loadPrefs().quality !== "low";
}

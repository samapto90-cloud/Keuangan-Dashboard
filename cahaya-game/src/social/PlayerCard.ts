import { escapeHtml } from "../dialogue/DialogueUI";
import type { PlayerCardView } from "../network/NetworkMessage";

export class PlayerCard {
  readonly root: HTMLElement;
  blocking = false;
  data: PlayerCardView | null = null;
  onInspect: ((id: string) => void) | null = null;
  onInvite: ((id: string) => void) | null = null;
  onFriend: ((id: string) => void) | null = null;
  onWhisper: ((id: string, name: string) => void) | null = null;
  onTrade: ((id: string) => void) | null = null;
  onReport: ((id: string, category: string) => void) | null = null;
  onBlock: ((id: string) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay card-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = (e.target as HTMLElement).closest("[data-card]") as HTMLElement | null;
      if (!t) return;
      const id = this.data?.playerId || "";
      const name = this.data?.name || "";
      if (t.dataset.card === "close") this.close();
      if (t.dataset.card === "inspect") this.onInspect?.(id);
      if (t.dataset.card === "invite") this.onInvite?.(id);
      if (t.dataset.card === "friend") this.onFriend?.(id);
      if (t.dataset.card === "whisper") this.onWhisper?.(id, name);
      if (t.dataset.card === "trade") this.onTrade?.(id);
      if (t.dataset.card === "report") this.onReport?.(id, t.dataset.cat || "HARASSMENT");
      if (t.dataset.card === "block") this.onBlock?.(id);
    });
  }

  show(data: PlayerCardView): void {
    this.data = data;
    this.blocking = true;
    this.root.hidden = false;
    this.root.innerHTML = `
      <div class="rpg-card inspect-card">
        <header><h3>PLAYER CARD</h3><button type="button" data-card="close">Close</button></header>
        <div class="char-name">${escapeHtml(data.title ? `[${data.title}] ${data.name}` : data.name)}</div>
        <p>Lv${data.level} · ${escapeHtml(data.class)} · ${escapeHtml(data.status || "ONLINE")}</p>
        <p>Guild ${escapeHtml(data.guild ? `[${data.guildTag || ""}] ${data.guild}` : "—")}</p>
        <p>Rank ${escapeHtml(data.rank || "—")} · AP ${data.achievements ?? 0}</p>
        <p>Season ${escapeHtml(data.season || "—")} · Cosmetic ${escapeHtml(data.cosmetic || "—")}</p>
        <div class="social-acts">
          <button type="button" data-card="inspect">INSPECT</button>
          <button type="button" data-card="invite">INVITE</button>
          <button type="button" data-card="friend">FRIEND</button>
          <button type="button" data-card="whisper">WHISPER</button>
          <button type="button" data-card="trade">TRADE</button>
          <button type="button" data-card="block">BLOCK</button>
        </div>
        <p>Report</p>
        <div class="social-acts">
          <button type="button" data-card="report" data-cat="HARASSMENT">HARASSMENT</button>
          <button type="button" data-card="report" data-cat="SPAM">SPAM</button>
          <button type="button" data-card="report" data-cat="CHEATING">CHEATING</button>
          <button type="button" data-card="report" data-cat="ABUSE">ABUSE</button>
          <button type="button" data-card="report" data-cat="INAPPROPRIATE_NAME">NAME</button>
        </div>
      </div>`;
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }
}

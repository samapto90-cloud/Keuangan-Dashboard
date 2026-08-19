import { escapeHtml } from "../dialogue/DialogueUI";
import type { GuildView } from "../network/NetworkMessage";

export class GuildUI {
  readonly root: HTMLElement;
  blocking = false;
  guild: GuildView | null = null;
  confirmDisband = false;
  onCreate: ((name: string, tag: string) => void) | null = null;
  onInvite: ((id: string) => void) | null = null;
  onLeave: (() => void) | null = null;
  onDisband: (() => void) | null = null;
  onAnnounce: ((text: string) => void) | null = null;
  onRefresh: (() => void) | null = null;
  onHall: (() => void) | null = null;
  onContribute: (() => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay guild-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => this.onClick(e));
  }

  toggle(): void {
    if (this.blocking) this.close();
    else this.open();
  }

  open(): void {
    this.blocking = true;
    this.root.hidden = false;
    this.onRefresh?.();
    this.render();
  }

  close(): void {
    this.blocking = false;
    this.confirmDisband = false;
    this.root.hidden = true;
  }

  apply(guild: GuildView | null | undefined): void {
    this.guild = guild?.guildId ? guild : null;
    if (this.blocking) this.render();
  }

  render(): void {
    if (!this.blocking) return;
    const g = this.guild;
    const members = (g?.members ?? []).map((m) =>
      `<div class="social-row"><span>[${escapeHtml(m.rank === "LEADER" ? "MASTER" : m.rank)}] ${escapeHtml(m.name)} · contrib ${m.contribution}</span></div>`).join("");
    this.root.innerHTML = `
      <div class="rpg-card guild-card">
        <header><h3>GUILD HALL</h3><button type="button" data-g="close">Close</button></header>
        ${g ? `
          <div class="guild-emblem">${escapeHtml(g.emblemId || "emblem-dawn")}</div>
          <div class="char-name">[${escapeHtml(g.tag)}] ${escapeHtml(g.name)}</div>
          <p>Level ${g.level} · EXP ${g.exp}</p>
          <p>Quest: ${escapeHtml(g.quest === "gq-desa-1" ? "Pertahanan Desa." : (g.quest || "—"))} (${g.questProgress || 0})</p>
          <p class="guild-ann">${escapeHtml(g.announcement || "No announcement.")}</p>
          <h4>MEMBERS</h4>
          <div class="social-body">${members}</div>
          <label>Invite player id <input name="invite" /></label>
          <label>Announcement <input name="ann" maxlength="200" /></label>
          <footer class="inv-actions">
            <button type="button" data-g="hall">ENTER HALL</button>
            <button type="button" data-g="contrib">SUMBANG BAHAN</button>
            <button type="button" data-g="invite">INVITE</button>
            <button type="button" data-g="announce">ANNOUNCE</button>
            <button type="button" data-g="leave">LEAVE</button>
            <button type="button" data-g="disband">${this.confirmDisband ? "CONFIRM DISBAND" : "DISBAND"}</button>
            ${this.confirmDisband ? `<button type="button" data-g="cancel-disband">CANCEL</button>` : ""}
          </footer>
        ` : `
          <p>Belum dalam guild. Level 10 · 1000 Coin.</p>
          <label>Name <input name="name" maxlength="24" /></label>
          <label>Tag <input name="tag" maxlength="5" /></label>
          <footer class="inv-actions"><button type="button" data-g="create">CREATE GUILD</button></footer>
        `}
      </div>`;
  }

  private onClick(e: MouseEvent): void {
    const t = (e.target as HTMLElement).closest("[data-g]") as HTMLElement | null;
    if (!t) return;
    if (t.dataset.g === "close") this.close();
    if (t.dataset.g === "create") {
      const name = (this.root.querySelector("[name=name]") as HTMLInputElement)?.value || "";
      const tag = (this.root.querySelector("[name=tag]") as HTMLInputElement)?.value || "";
      this.onCreate?.(name, tag);
    }
    if (t.dataset.g === "hall") this.onHall?.();
    if (t.dataset.g === "contrib") this.onContribute?.();
    if (t.dataset.g === "invite") {
      const id = (this.root.querySelector("[name=invite]") as HTMLInputElement)?.value.trim() || "";
      if (id) this.onInvite?.(id);
    }
    if (t.dataset.g === "announce") {
      const text = (this.root.querySelector("[name=ann]") as HTMLInputElement)?.value || "";
      this.onAnnounce?.(text);
    }
    if (t.dataset.g === "leave") this.onLeave?.();
    if (t.dataset.g === "cancel-disband") {
      this.confirmDisband = false;
      this.render();
    }
    if (t.dataset.g === "disband") {
      if (!this.confirmDisband) {
        this.confirmDisband = true;
        this.render();
        return;
      }
      this.onDisband?.();
      this.confirmDisband = false;
    }
  }
}

import { escapeHtml } from "../dialogue/DialogueUI";
import type { InspectOut } from "../network/NetworkMessage";
import { EQUIP_SLOTS } from "../inventory/LoadoutStore";
import type { LoadoutStore } from "../inventory/LoadoutStore";
import { itemGlyph } from "../inventory/ItemTooltip";

export class PlayerInspect {
  readonly root: HTMLElement;
  blocking = false;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay inspect-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      if ((e.target as HTMLElement).closest("[data-close]")) this.close();
    });
  }

  show(data: InspectOut, store: LoadoutStore): void {
    this.blocking = true;
    this.root.hidden = false;
    const gear = EQUIP_SLOTS.map((slot) => {
      const id = data.equipment[slot];
      const item = store.item(id);
      return `<div>${slot}: ${item ? `${itemGlyph(item.icon)} ${escapeHtml(item.name)}` : "empty"}</div>`;
    }).join("");
    this.root.innerHTML = `
      <div class="rpg-card inspect-card">
        <header><h3>INSPECT</h3><button type="button" data-close>Close</button></header>
        <div class="char-name">${escapeHtml(data.title ? `[${data.title}] ${data.name}` : data.name)}</div>
        <div>Level ${data.level} · ${escapeHtml(data.class)} · Power ${data.powerRating ?? 0}</div>
        <div>Guild ${escapeHtml(data.guild || data.guildTag || "—")}</div>
        <div>Title ${escapeHtml(data.title || "—")}</div>
        <div>Region ${escapeHtml(data.region || "—")} · ${escapeHtml(data.status || "")}</div>
        <div>Achievement ${data.achievementScore ?? 0}</div>
        <div class="inspect-gear">${gear}</div>
      </div>`;
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }
}

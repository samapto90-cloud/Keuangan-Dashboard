import { escapeHtml } from "../dialogue/DialogueUI";
import type { LoadoutStore } from "./LoadoutStore";
import { EquipmentUI } from "./EquipmentUI";

export class CharacterPanel {
  readonly root: HTMLElement;
  blocking = false;
  onUnequip: ((slot: string) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay char-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = e.target as HTMLElement;
      if (t.closest("[data-close]")) {
        this.close();
        return;
      }
      const eq = t.closest("[data-equip]") as HTMLElement | null;
      if (eq?.dataset.equip) this.onUnequip?.(eq.dataset.equip);
    });
  }

  toggle(store: LoadoutStore, name: string): void {
    if (this.blocking) this.close();
    else this.open(store, name);
  }

  open(store: LoadoutStore, name: string): void {
    this.blocking = true;
    this.root.hidden = false;
    this.render(store, name);
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }

  render(store: LoadoutStore, name: string): void {
    if (!this.blocking) return;
    const s = store.stats;
    const crit = s ? `${(s.criticalChance * 100).toFixed(1)}%` : "—";
    this.root.innerHTML = `
      <div class="rpg-card char-card">
        <header><h3>CHARACTER</h3><button type="button" data-close>Close</button></header>
        <div class="char-grid">
          <div class="char-stats">
            <div class="char-name">${escapeHtml(name)}</div>
            <div>Combat Power ${s?.powerRating ?? 0}</div>
            <div>Level ${s?.level ?? 1}</div>
            <div>Class ${escapeHtml(s?.class ?? "WARRIOR")}</div>
            <div>HP ${s?.hp ?? 0} / ${s?.maxHp ?? 0}</div>
            <div>Energy ${s?.energy ?? 0} / ${s?.maxEnergy ?? 0}</div>
            <div>Attack ${s?.attack ?? 0}</div>
            <div>Defense ${s?.defense ?? 0}</div>
            <div>Strength ${s?.strength ?? 0}</div>
            <div>Agility ${s?.agility ?? 0}</div>
            <div>Critical ${crit}</div>
          </div>
          ${EquipmentUI.html(store, null)}
        </div>
      </div>`;
  }
}

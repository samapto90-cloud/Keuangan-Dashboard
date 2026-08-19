import { escapeHtml } from "../dialogue/DialogueUI";
import type { LoadoutStore } from "../inventory/LoadoutStore";
import { EquipmentUI } from "../inventory/EquipmentUI";
import type { ProgressionStore } from "./ProgressionStore";
import { AttributeUI } from "./AttributeUI";

export class CharacterProgressionUI {
  readonly root: HTMLElement;
  blocking = false;
  readonly attributes: AttributeUI;
  onOpenTree: (() => void) | null = null;
  onOpenForms: (() => void) | null = null;
  onOpenBuilds: (() => void) | null = null;
  onSetStyle: ((styleId: string) => void) | null = null;
  onUnequip: ((slot: string) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay char-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.attributes = new AttributeUI(this.root);
    this.root.addEventListener("click", (e) => {
      const t = e.target as HTMLElement;
      if (t.closest("[data-close]")) this.close();
      if (t.closest("[data-tree]")) this.onOpenTree?.();
      if (t.closest("[data-forms]")) this.onOpenForms?.();
      if (t.closest("[data-builds]")) this.onOpenBuilds?.();
      const st = t.closest("[data-style]") as HTMLElement | null;
      if (st?.dataset.style) this.onSetStyle?.(st.dataset.style);
      const eq = t.closest("[data-equip]") as HTMLElement | null;
      if (eq?.dataset.equip) this.onUnequip?.(eq.dataset.equip);
    });
  }

  toggle(loadout: LoadoutStore, progression: ProgressionStore, name: string): void {
    if (this.blocking) this.close();
    else this.open(loadout, progression, name);
  }

  open(loadout: LoadoutStore, progression: ProgressionStore, name: string): void {
    this.blocking = true;
    this.root.hidden = false;
    this.render(loadout, progression, name);
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }

  render(loadout: LoadoutStore, progression: ProgressionStore, name: string): void {
    if (!this.blocking) return;
    const s = loadout.stats;
    const v = progression.view;
    this.root.innerHTML = `
      <div class="rpg-card char-card">
        <header><h3>CHARACTER</h3><button type="button" data-close>Close</button></header>
        <div class="char-grid">
          <div class="char-stats">
            <div class="char-name">${escapeHtml(name)}</div>
            <div>Combat Rating ${v?.combatRating ?? v?.powerRating ?? s?.powerRating ?? 0}</div>
            <div>Level ${s?.level ?? v?.level ?? 1}</div>
            <div>Style ${escapeHtml(v?.combatStyle ?? "DAWN_FIST")}</div>
            <div>Mastery ${Object.values(v?.styleMastery ?? {}).reduce((a, n) => a + n, 0)}</div>
            <div>Transformation ${escapeHtml(progression.formLabel())}</div>
            <div>HP ${s?.hp ?? 0} / ${s?.maxHp ?? 0}</div>
            <div>Energy ${s?.energy ?? 0} / ${s?.maxEnergy ?? 0}</div>
            <div>POWER ${s?.strength ?? 0} · RESOLVE ${s?.defense ?? 0}</div>
            <div>AGILITY ${s?.agility ?? 0} · FOCUS ${s?.energyPower ?? 0} · VITALITY ${v?.spentVit ?? 0}</div>
            <div>Skills ${(v?.unlockedSkills ?? []).length}</div>
            <div class="style-row">${(v?.styles ?? [{ id: "DAWN_FIST", name: "DAWN FIST" }, { id: "WIND_STEP", name: "WIND STEP" }, { id: "IRON_GUARD", name: "IRON GUARD" }, { id: "CELESTIAL_FLOW", name: "CELESTIAL FLOW" }]).map((st) => `<button type="button" data-style="${st.id}" class="${v?.combatStyle === st.id ? "on" : ""}">${escapeHtml(st.name)}</button>`).join("")}</div>
            <div class="char-actions">
              <button type="button" data-tree>SKILL TREE</button>
              <button type="button" data-forms>FORMS</button>
              <button type="button" data-builds>BUILDS</button>
            </div>
            <div class="attr-host"></div>
          </div>
          ${EquipmentUI.html(loadout, null)}
        </div>
      </div>`;
    const host = this.root.querySelector(".attr-host");
    if (host) {
      host.appendChild(this.attributes.root);
      this.attributes.render(progression);
    }
  }
}

import { escapeHtml } from "../dialogue/DialogueUI";
import type { ProgressionStore } from "./ProgressionStore";

export class BuildUI {
  readonly root: HTMLElement;
  blocking = false;
  onSave: ((slot: number, name: string) => void) | null = null;
  onLoad: ((slot: number) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay build-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = e.target as HTMLElement;
      if (t.closest("[data-close]")) this.close();
      const save = t.closest("[data-save]") as HTMLElement | null;
      if (save) {
        const slot = Number(save.dataset.save);
        const input = this.root.querySelector<HTMLInputElement>(`[data-name="${slot}"]`);
        this.onSave?.(slot, input?.value || `Build ${slot + 1}`);
      }
      const load = t.closest("[data-load]") as HTMLElement | null;
      if (load) this.onLoad?.(Number(load.dataset.load));
    });
  }

  toggle(store: ProgressionStore): void {
    if (this.blocking) this.close();
    else this.open(store);
  }

  open(store: ProgressionStore): void {
    this.blocking = true;
    this.root.hidden = false;
    this.render(store);
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }

  render(store: ProgressionStore): void {
    if (!this.blocking) return;
    const slots = store.view?.builds ?? [
      { id: "build-0", slot: 0, name: "Fist Master", style: "DAWN_FIST", formId: "", score: 0, active: false },
      { id: "build-1", slot: 1, name: "Wind Runner", style: "WIND_STEP", formId: "", score: 0, active: false },
      { id: "build-2", slot: 2, name: "Guardian", style: "IRON_GUARD", formId: "", score: 0, active: false },
    ];
    while (slots.length < 3) {
      slots.push({ id: `build-${slots.length}`, slot: slots.length, name: "", style: "", formId: "", score: 0, active: false });
    }
    this.root.innerHTML = `
      <div class="rpg-card build-card">
        <header><h3>BUILD MANAGER</h3><button type="button" data-close>Close</button></header>
        <div class="build-grid">
          ${slots
            .slice(0, 3)
            .map(
              (b, i) => `<div class="build-slot ${b.active ? "on" : ""}">
                <input data-name="${i}" value="${escapeHtml(b.name || ["Fist Master", "Wind Runner", "Guardian"][i] || `Build ${i + 1}`)}" maxlength="24"/>
                <div>${escapeHtml(b.style || "-")} · Score ${b.score || 0}</div>
                <button type="button" data-save="${i}">SAVE</button>
                <button type="button" data-load="${i}">LOAD BUILD</button>
              </div>`,
            )
            .join("")}
        </div>
      </div>`;
  }
}

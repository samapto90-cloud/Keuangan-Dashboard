import type { ProgressionStore } from "./ProgressionStore";

export class AttributeUI {
  readonly root: HTMLElement;
  onAllocate: ((stat: string) => void) | null = null;
  onReset: (() => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "attr-panel";
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = e.target as HTMLElement;
      if (t.dataset.reset) this.onReset?.();
      if (t.dataset.stat) this.onAllocate?.(t.dataset.stat);
    });
  }

  render(store: ProgressionStore): void {
    const v = store.view;
    const row = (id: string, label: string, spent: number) =>
      `<div class="attr-row"><span>${label}</span><b>${spent}</b><button type="button" data-stat="${id}" ${!v || v.attributePoints < 1 ? "disabled" : ""}>+</button></div>`;
    this.root.innerHTML = `
      <div class="attr-head">ATTRIBUTE <small>${v?.attributePoints ?? 0} poin</small></div>
      ${row("STR", "POWER", v?.spentStr ?? 0)}
      ${row("DEF", "RESOLVE", v?.spentDef ?? 0)}
      ${row("AGI", "AGILITY", v?.spentAgi ?? 0)}
      ${row("ENG", "FOCUS", v?.spentEng ?? 0)}
      ${row("VIT", "VITALITY", v?.spentVit ?? 0)}
      <button type="button" class="attr-reset" data-reset="1">RESET ATTRIBUTES</button>
      <small>Free reset ${v?.attrResetLeft ?? 3}</small>
    `;
  }
}

import { escapeHtml } from "../dialogue/DialogueUI";
import type { ProgressionStore } from "./ProgressionStore";

export class TransformationWheel {
  readonly root: HTMLElement;
  blocking = false;
  onSelect: ((formId: string) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay form-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = e.target as HTMLElement;
      if (t.closest("[data-close]")) {
        this.close();
        return;
      }
      const form = t.closest("[data-form]") as HTMLElement | null;
      if (form?.dataset.form) {
        this.onSelect?.(form.dataset.form);
        this.close();
      }
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
    const forms = store.view?.forms ?? [{ id: "normal", name: "NORMAL FORM", shortName: "NORMAL", unlocked: true, active: true }];
    this.root.innerHTML = `
      <div class="rpg-card form-card">
        <header><h3>TRANSFORMATION</h3><button type="button" data-close>Close</button></header>
        <div class="form-wheel">
          ${forms
            .map(
              (f) =>
                `<button type="button" class="form-node ${f.active ? "active" : ""} ${f.unlocked ? "" : "locked"}" data-form="${f.id}" ${f.unlocked ? "" : "disabled"}>
                  <b>${escapeHtml(f.shortName)}</b>
                  <span>${f.unlocked ? escapeHtml(f.storyName || f.name) : "LOCKED"}</span>
                </button>`,
            )
            .join("")}
        </div>
      </div>`;
  }
}

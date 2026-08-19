import type { ProgressionStore } from "./ProgressionStore";

export class TransformationUI {
  readonly root: HTMLElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "hud-transform";
    host.appendChild(this.root);
  }

  update(store: ProgressionStore): void {
    const v = store.view;
    const energy = v ? Math.round((v.transEnergy / Math.max(1, v.maxTransEnergy)) * 10) : 0;
    const bar = "█".repeat(Math.max(0, energy)) + "░".repeat(Math.max(0, 10 - energy));
    this.root.innerHTML = `
      <div>${store.formLabel()}</div>
      <div class="hud-tbar">${bar}</div>
      <div>${store.readyLabel()}</div>
    `;
  }
}

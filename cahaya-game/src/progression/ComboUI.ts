export class ComboUI {
  readonly root: HTMLElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "hud-combo";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  show(hits: number, finisher: boolean): void {
    if (hits < 2 && !finisher) {
      this.root.hidden = true;
      return;
    }
    this.root.hidden = false;
    this.root.textContent = finisher ? "FINISH!" : `${hits} HIT`;
    window.setTimeout(() => {
      this.root.hidden = true;
    }, finisher ? 700 : 480);
  }
}

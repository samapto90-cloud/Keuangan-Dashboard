export class FishingUI {
  readonly root: HTMLElement;
  blocking = false;
  progress = 8;
  dir = 1;
  targetA = 42;
  targetB = 68;
  prompt = "Tap / SPACE saat di zona.";
  onCatch: ((progress: number) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay fish-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", () => this.catchNow());
  }

  start(targetA: number, targetB: number, prompt?: string): void {
    this.targetA = targetA;
    this.targetB = targetB;
    this.progress = 8;
    this.dir = 1;
    this.prompt = prompt || "Tap / SPACE saat di zona.";
    this.blocking = true;
    this.root.hidden = false;
    this.render();
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }

  tick(dt: number): void {
    if (!this.blocking) return;
    this.progress += this.dir * dt * 70;
    if (this.progress >= 100) {
      this.progress = 100;
      this.dir = -1;
    }
    if (this.progress <= 0) {
      this.progress = 0;
      this.dir = 1;
    }
    this.render();
  }

  catchNow(): void {
    if (!this.blocking) return;
    const value = Math.round(this.progress);
    this.close();
    this.onCatch?.(value);
  }

  private render(): void {
    const a = this.targetA;
    const b = this.targetB;
    this.root.innerHTML = `
      <div class="rpg-card fish-card">
        <h3>MEMANCING</h3>
        <p>${this.prompt}</p>
        <div class="fish-bar">
          <i class="fish-zone" style="left:${a}%;width:${Math.max(4, b - a)}%"></i>
          <b class="fish-cursor" style="left:${this.progress}%"></b>
        </div>
        <button type="button">TAP / SPACE</button>
      </div>`;
  }
}

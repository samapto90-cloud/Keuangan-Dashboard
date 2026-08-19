export class DungeonTimerUI {
  readonly root: HTMLElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "dun-timer";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  set(seconds: number, visible: boolean): void {
    this.root.hidden = !visible;
    if (!visible) return;
    const m = Math.floor(Math.max(0, seconds) / 60);
    const s = Math.max(0, seconds) % 60;
    this.root.textContent = `TIME ${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
    this.root.classList.toggle("low", seconds > 0 && seconds <= 60);
  }
}

import type { DungeonView } from "../network/NetworkMessage";

export class DungeonMap {
  readonly root: HTMLElement;
  private readonly canvas: HTMLCanvasElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "dun-map";
    this.root.hidden = true;
    this.canvas = document.createElement("canvas");
    this.canvas.width = 128;
    this.canvas.height = 128;
    this.root.appendChild(this.canvas);
    host.appendChild(this.root);
  }

  apply(view: DungeonView | null, selfX: number, selfZ: number): void {
    const on = !!view && view.state !== "COMPLETED" && view.state !== "FAILED";
    this.root.hidden = !on;
    if (!on) return;
    const ctx = this.canvas.getContext("2d");
    if (!ctx) return;
    ctx.fillStyle = "rgba(7,17,31,0.72)";
    ctx.fillRect(0, 0, 128, 128);
    ctx.strokeStyle = "rgba(232,184,109,0.45)";
    ctx.strokeRect(4, 4, 120, 120);
    const px = (x: number) => 64 + x * 2.2;
    const pz = (z: number) => 20 + z * 3.6;
    ctx.fillStyle = "#86efac";
    ctx.beginPath();
    ctx.arc(px(selfX), pz(selfZ), 4, 0, Math.PI * 2);
    ctx.fill();
    if (view?.boss?.alive) {
      ctx.fillStyle = "#ef4444";
      ctx.fillRect(px(0) - 5, pz(20) - 5, 10, 10);
    }
    ctx.fillStyle = "#e8b86d";
    ctx.fillRect(px(0) - 3, pz(2) - 3, 6, 6);
    for (const m of view?.members ?? []) {
      ctx.fillStyle = m.playerId === (view?.members ?? []).find(() => true)?.playerId ? "#86efac" : "#60a5fa";
      ctx.beginPath();
      ctx.arc(64, 48, 3, 0, Math.PI * 2);
      ctx.fill();
    }
  }
}

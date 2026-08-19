import type { DungeonView } from "../network/NetworkMessage";

export class DungeonObjectiveUI {
  readonly root: HTMLElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "dun-objective";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  apply(view: DungeonView | null): void {
    if (!view || view.state === "COMPLETED" || view.state === "FAILED") {
      this.root.hidden = true;
      return;
    }
    this.root.hidden = false;
    const prog = view.count > 0 ? ` ${view.progress}/${view.count}` : "";
    const boss = view.objectiveType === "BOSS" && view.state !== "BOSS" ? " · BOSS: LOCKED" : "";
    const mech = view.mechanic ? `<div>MECHANIC ${view.mechanic}${view.guideHp ? ` · GUIDE ${view.guideHp}` : ""}</div>` : "";
    const shield = view.eduShield || view.crystalShield ? "<div>EDU SHIELD ACTIVE</div>" : "";
    const room = view.room ? `<div>${view.room}${view.checkpoint ? ` · ${view.checkpoint}` : ""}</div>` : "";
    this.root.innerHTML = `<span>WAVE ${view.wave}/${view.waveTotal}</span><strong>${view.objective || "Jelajahi dungeon"}${prog}${boss}</strong>${room}${mech}${shield}`;
  }
}

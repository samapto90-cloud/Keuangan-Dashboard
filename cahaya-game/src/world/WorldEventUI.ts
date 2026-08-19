import type { WorldEventView } from "../network/NetworkMessage";

export class WorldEventUI {
  readonly root: HTMLDivElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "world-event-banner";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  apply(ev: WorldEventView | null | undefined): void {
    if (!ev || (ev.state !== "ACTIVE" && ev.state !== "ANNOUNCED" && ev.state !== "SUCCESS" && ev.state !== "WAITING" && ev.state !== "FINAL")) {
      this.root.hidden = true;
      return;
    }
    this.root.hidden = false;
    const phase = ev.phase || ev.state;
    const timer = ev.startsIn && ev.startsIn > 0
      ? `starts in ${ev.startsIn}s`
      : ev.endsIn && ev.endsIn > 0
        ? `${phase === "FINAL" ? "ends in" : "active"} ${ev.endsIn}s`
        : phase;
    const gate = ev.maxGateHp ? ` · GATE ${ev.gateHp}/${ev.maxGateHp}` : "";
    const obj = ev.objective ? ` · ${ev.objective} ${ev.progress ?? 0}/${ev.need ?? 0}` : "";
    this.root.innerHTML = `<strong>WORLD EVENT</strong><span>${ev.announce || ev.name}</span><em>${timer}${gate}${obj}</em>`;
  }

  applyBoss(boss: { name?: string; hp?: number; maxHp?: number; phaseName?: string; announce?: string } | null | undefined): void {
    if (!boss) return;
    this.root.hidden = false;
    this.root.innerHTML = `<strong>WORLD BOSS</strong><span>${boss.announce || boss.name}</span><em>${boss.phaseName || ""} ${boss.hp}/${boss.maxHp}</em>`;
  }
}

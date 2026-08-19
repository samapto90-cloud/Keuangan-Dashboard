import { escapeHtml } from "../dialogue/DialogueUI";
import type { BossView } from "../network/NetworkMessage";

export class BossHealthBar {
  readonly root: HTMLElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "boss-bar";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  apply(boss: BossView | null | undefined): void {
    if (!boss || !boss.alive) {
      this.root.hidden = true;
      return;
    }
    this.root.hidden = false;
    const pct = boss.maxHp > 0 ? Math.max(0, Math.min(100, (boss.hp / boss.maxHp) * 100)) : 0;
    const label = boss.enraged ? "ENRAGED" : `PHASE ${boss.phase}`;
    this.root.innerHTML = `
      <div class="boss-title">${escapeHtml(boss.name)} <span>${escapeHtml(boss.title || "")}</span></div>
      <div class="boss-meta">Lv ${boss.level} · ${escapeHtml(label)}${boss.enraged ? " · ENRAGE" : ""}</div>
      <div class="hud-bar hud-hp"><i style="width:${pct}%"></i></div>
      <div class="boss-hp">${boss.hp} / ${boss.maxHp}</div>`;
  }
}

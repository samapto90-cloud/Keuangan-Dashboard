import { escapeHtml } from "../dialogue/DialogueUI";
import type { ProgressionStore } from "./ProgressionStore";

const BRANCHES = ["COMBAT", "ENERGY", "DEFENSE", "MOBILITY", "TRANSFORMATION"];
const BRANCH_LABEL: Record<string, string> = {
  COMBAT: "OFFENSE",
  ENERGY: "ENERGY",
  DEFENSE: "DEFENSE",
  MOBILITY: "MOBILITY",
  TRANSFORMATION: "TRANSFORMATION",
};

export class SkillTreeUI {
  readonly root: HTMLElement;
  blocking = false;
  onUnlock: ((nodeId: string, skillId: string) => void) | null = null;
  onReset: (() => void) | null = null;
  private preview = "";

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay skill-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = e.target as HTMLElement;
      if (t.closest("[data-close]")) {
        this.close();
        return;
      }
      if (t.closest("[data-reset-skills]")) {
        this.onReset?.();
        return;
      }
      const node = t.closest("[data-node]") as HTMLElement | null;
      if (node?.dataset.node) {
        this.preview = node.dataset.node;
        this.onUnlock?.(node.dataset.node, node.dataset.skill || "");
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
    const v = store.view;
    const cols = BRANCHES.map((branch) => {
      const nodes = (v?.nodes ?? []).filter((n) => n.branch === branch);
      return `<div class="tree-col"><h4>${BRANCH_LABEL[branch] ?? branch}</h4>${nodes
        .map((n) => {
          const state = n.unlocked ? "unlocked" : n.available ? "available" : "locked";
          return `<button type="button" class="tree-node ${n.unlocked ? "on" : ""} ${state}" data-node="${n.id}" data-skill="${n.skillId}">
              <b>${escapeHtml((n.name || n.skillId).replaceAll("_", " "))}</b>
              <span>Lv ${n.requiredLevel} · ${n.cost} SP · ${state}</span>
            </button>`;
        })
        .join("")}</div>`;
    }).join("");
    const sel = (v?.nodes ?? []).find((n) => n.id === this.preview) ?? (v?.nodes ?? [])[0];
    const preview = sel
      ? `<div class="skill-preview"><h4>${escapeHtml(sel.name || sel.skillId)}</h4>
          <p>${escapeHtml(sel.description || "Skill node.")}</p>
          <div>Energy ${sel.energyCost ?? 0} · CD ${sel.cooldown ?? 0}s · Dmg ${sel.damage ?? 0} · Range ${sel.range ?? 0}</div>
          <div>${escapeHtml(sel.effect || "combat")}</div></div>`
      : "";
    this.root.innerHTML = `
      <div class="rpg-card skill-card">
        <header><h3>SKILL TREE</h3><span>${v?.skillPoints ?? 0} Skill Point</span>
          <button type="button" data-reset-skills>RESET</button>
          <button type="button" data-close>Close</button></header>
        <div class="tree-grid">${cols}</div>
        ${preview}
      </div>`;
  }
}

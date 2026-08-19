import { SKILLS, type SkillDef } from "./SkillData";

export class SkillBar {
  readonly root: HTMLElement;
  private readonly fills = new Map<string, HTMLElement>();

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "skill-bar";
    this.root.innerHTML = SKILLS.map(
      (s, i) =>
        `<button type="button" class="skill-btn ${s.id === "celestial_impact" ? "skill-ult" : ""}" data-id="${s.id}"><b>${["1", "2", "3", "4", "R"][i] ?? ""}</b><span>${s.name}</span><i data-cd="${s.id}"></i></button>`,
    ).join("");
    host.appendChild(this.root);
    for (const skill of SKILLS) {
      const el = this.root.querySelector(`[data-cd="${skill.id}"]`);
      if (el instanceof HTMLElement) this.fills.set(skill.id, el);
    }
  }

  onClick(handler: (skill: SkillDef) => void): void {
    this.root.addEventListener("click", (e) => {
      const btn = (e.target as HTMLElement).closest("button[data-id]");
      if (!(btn instanceof HTMLElement)) return;
      const skill = SKILLS.find((s) => s.id === btn.dataset.id);
      if (skill) handler(skill);
    });
  }

  update(readyAt: Record<string, number>, now: number): void {
    for (const skill of SKILLS) {
      const el = this.fills.get(skill.id);
      if (!el) continue;
      const left = (readyAt[skill.id] ?? 0) - now;
      const pct = left <= 0 ? 0 : Math.min(1, left / (skill.cooldown * 1000));
      el.style.setProperty("--cd", String(pct));
      const btn = el.closest(".skill-btn");
      if (btn instanceof HTMLElement) btn.style.setProperty("--cd", String(pct));
    }
  }
}

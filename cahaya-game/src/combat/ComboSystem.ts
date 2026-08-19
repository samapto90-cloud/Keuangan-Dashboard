import { COMBAT } from "../game/GameConfig";
import { ATTACKS, type AttackData } from "./AttackData";
import type { AttackKind } from "./CombatState";

const RECIPES: AttackKind[][] = [
  ["punch", "punch", "kick"],
  ["punch", "kick", "punch"],
  ["kick", "kick", "punch"],
  ["punch", "punch", "kick", "energy"],
];

export class ComboSystem {
  sequence: AttackKind[] = [];
  count = 0;
  private windowTimer = 0;
  private counterTimer = 0;

  resolveData(kind: AttackKind): AttackData {
    if (this.windowTimer <= 0) this.sequence = [];
    const next = [...this.sequence, kind];
    const punchCount = next.filter((k) => k === "punch").length;
    const kickCount = next.filter((k) => k === "kick").length;
    if (kind === "punch") return punchCount >= 2 ? ATTACKS.PUNCH_2 : ATTACKS.PUNCH_1;
    if (kind === "kick") return kickCount >= 2 ? ATTACKS.KICK_2 : ATTACKS.KICK_1;
    return ATTACKS.ENERGY_1;
  }

  register(kind: AttackKind): void {
    if (this.windowTimer <= 0) this.sequence = [];
    this.sequence.push(kind);
    if (this.sequence.length > 4) this.sequence.splice(0, this.sequence.length - 4);
    this.windowTimer = COMBAT.comboWindow;
    const matched = RECIPES.some((recipe) => startsWith(this.sequence, recipe) || equalSeq(this.sequence, recipe));
    if (!matched && this.sequence.length > 1) {
      const last = this.sequence[this.sequence.length - 1];
      this.sequence = last ? [last] : [];
    }
  }

  onHit(): void {
    this.count = Math.min(10, this.count + 1);
    this.counterTimer = COMBAT.comboReset;
  }

  update(dt: number): void {
    if (this.windowTimer > 0) {
      this.windowTimer -= dt;
      if (this.windowTimer <= 0) this.sequence = [];
    }
    if (this.counterTimer > 0) {
      this.counterTimer -= dt;
      if (this.counterTimer <= 0) this.count = 0;
    }
  }

  inWindow(): boolean {
    return this.windowTimer > 0;
  }
}

function startsWith(seq: AttackKind[], recipe: AttackKind[]): boolean {
  if (seq.length > recipe.length) return false;
  return seq.every((k, i) => k === recipe[i]);
}

function equalSeq(a: AttackKind[], b: AttackKind[]): boolean {
  return a.length === b.length && a.every((k, i) => k === b[i]);
}

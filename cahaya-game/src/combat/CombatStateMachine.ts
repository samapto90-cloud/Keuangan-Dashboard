import type { CombatState } from "./CombatState";

export class CombatStateMachine {
  state: CombatState = "IDLE";

  apply(next: CombatState): CombatState {
    if (this.state === "DEAD" && next !== "IDLE" && next !== "RESPAWNING") return this.state;
    this.state = next;
    return this.state;
  }
}

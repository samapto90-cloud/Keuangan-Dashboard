import { attackDuration, type AttackData } from "./AttackData";
import type { AttackKind, AttackPhase } from "./CombatState";
import { ComboSystem } from "./ComboSystem";
import { EnergySystem } from "./EnergySystem";
import type { Player } from "../player/Player";

export interface ActiveAttack {
  data: AttackData;
  elapsed: number;
  phase: AttackPhase;
  hitTargets: Set<string>;
  spawnedProjectile: boolean;
}

export class AttackController {
  current: ActiveAttack | null = null;
  private queued: AttackKind | null = null;
  private readonly cooldown = new Map<string, number>();

  constructor(
    private readonly combo: ComboSystem,
    private readonly energy: EnergySystem,
  ) {}

  get attacking(): boolean {
    return this.current !== null;
  }

  request(kind: AttackKind, player: Player): void {
    if (player.health.isDead()) return;
    if (player.combatState === "DODGING" || player.combatState === "HIT" || player.combatState === "STUNNED") return;
    if (kind === "special" || kind === "ultimate" || kind === "dodge") return;
    if (this.current) {
      this.queued = kind;
      return;
    }
    this.tryStart(kind, player, false);
  }

  update(dt: number, player: Player): void {
    for (const [id, left] of this.cooldown) {
      const next = left - dt;
      if (next <= 0) this.cooldown.delete(id);
      else this.cooldown.set(id, next);
    }
    if (!this.current) {
      if (this.queued) {
        const kind = this.queued;
        this.queued = null;
        this.tryStart(kind, player, true);
      }
      return;
    }
    this.current.elapsed += dt;
    const d = this.current.data;
    if (this.current.elapsed < d.startup) this.current.phase = "startup";
    else if (this.current.elapsed < d.startup + d.active) this.current.phase = "active";
    else if (this.current.elapsed < attackDuration(d)) this.current.phase = "recovery";
    else {
      this.current = null;
      if (player.combatState === "ATTACKING" || player.combatState === "COMBO" || player.combatState === "ENERGY_ATTACK") {
        player.combatState = "IDLE";
      }
      if (this.queued) {
        const kind = this.queued;
        this.queued = null;
        this.tryStart(kind, player, true);
      }
    }
  }

  private tryStart(kind: AttackKind, player: Player, fromCombo: boolean): void {
    const data = this.combo.resolveData(kind);
    if (!fromCombo && !this.combo.inWindow()) {
      const cd = this.cooldown.get(data.id) ?? 0;
      if (cd > 0) return;
    }
    if (data.energyCost > 0 && !this.energy.spend(player.stats, data)) return;
    this.combo.register(kind);
    this.current = {
      data,
      elapsed: 0,
      phase: "startup",
      hitTargets: new Set(),
      spawnedProjectile: false,
    };
    this.cooldown.set(data.id, data.cooldown);
    player.combatState = data.kind === "energy" ? "ENERGY_ATTACK" : this.combo.sequence.length > 1 ? "COMBO" : "ATTACKING";
  }

  cancel(): void {
    this.current = null;
    this.queued = null;
  }
}

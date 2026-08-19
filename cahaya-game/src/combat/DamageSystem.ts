import { COMBAT } from "../game/GameConfig";
import type { DamageType } from "./AttackData";
import type { Combatant } from "./Combatant";

export interface DamageInfo {
  attackerId: string;
  targetId: string;
  damage: number;
  damageType: DamageType;
  knockback: number;
  stunDuration: number;
  critical: boolean;
  source: string;
}

export interface DamageResult {
  applied: boolean;
  dealt: number;
  killed: boolean;
  info: DamageInfo;
}

export class DamageSystem {
  roll(base: DamageInfo, critChance: number = COMBAT.critChance): DamageInfo {
    const critical = Math.random() < critChance;
    return {
      ...base,
      critical,
      damage: critical ? Math.round(base.damage * COMBAT.critMultiplier) : base.damage,
    };
  }

  apply(target: Combatant, info: DamageInfo): DamageResult {
    if (target.invincible || target.health.isDead()) {
      return { applied: false, dealt: 0, killed: false, info };
    }
    const dealt = target.health.takeDamage(info.damage);
    const killed = target.health.isDead();
    if (killed) target.combatState = "DEAD";
    else if (info.stunDuration > 0.18) target.combatState = "STUNNED";
    else target.combatState = "HIT";
    return { applied: dealt > 0, dealt, killed, info };
  }
}

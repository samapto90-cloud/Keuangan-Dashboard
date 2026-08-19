import { COMBAT } from "../game/GameConfig";
import type { AttackData } from "./AttackData";
import type { PlayerStats } from "../player/PlayerStats";

export class EnergySystem {
  canAfford(stats: PlayerStats, attack: AttackData): boolean {
    return stats.energy >= attack.energyCost;
  }

  spend(stats: PlayerStats, attack: AttackData): boolean {
    if (!this.canAfford(stats, attack)) return false;
    stats.energy -= attack.energyCost;
    return true;
  }

  recover(stats: PlayerStats, dt: number, attacking: boolean): void {
    if (attacking) return;
    stats.energy = Math.min(stats.maxEnergy, stats.energy + COMBAT.energyRecoverPerSec * dt);
  }
}

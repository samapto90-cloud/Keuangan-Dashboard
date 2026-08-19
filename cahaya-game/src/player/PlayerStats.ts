import { PLAYER_NAME, STAMINA } from "../game/GameConfig";
import { HealthComponent } from "../combat/HealthComponent";

export class PlayerStats {
  name = PLAYER_NAME;
  readonly health = new HealthComponent(100);
  level = 1;
  energy: number = 100;
  maxEnergy: number = 100;
  stamina: number = STAMINA.max;
  maxStamina: number = STAMINA.max;
  exp = 0;
  expToNext = 120;
  /** Setelah stamina habis, sprint terkunci sampai pulih sebagian. */
  exhausted = false;
  attack = 8;
  defense = 4;
  strength = 8;
  agility = 0;
  energyPower = 0;
  criticalChance = 0.05;

  hpRatio(): number {
    return this.health.getHealthPercent();
  }

  energyRatio(): number {
    return this.energy / this.maxEnergy;
  }

  staminaRatio(): number {
    return this.stamina / this.maxStamina;
  }
}

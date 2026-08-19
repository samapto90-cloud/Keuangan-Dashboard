import type * as THREE from "three";
import type { CombatState } from "./CombatState";
import type { HealthComponent } from "./HealthComponent";

export type CombatTeam = "player" | "enemy";

export interface Hurtbox {
  radius: number;
  height: number;
  yOffset: number;
}

export interface Combatant {
  readonly id: string;
  readonly team: CombatTeam;
  readonly position: THREE.Vector3;
  readonly velocity: THREE.Vector3;
  readonly group: THREE.Object3D;
  facingYaw: number;
  health: HealthComponent;
  hurtbox: Hurtbox;
  invincible: boolean;
  combatState: CombatState;
}

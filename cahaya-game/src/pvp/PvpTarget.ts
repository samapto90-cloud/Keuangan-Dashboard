import * as THREE from "three";
import { HealthComponent } from "../combat/HealthComponent";
import type { Combatant, CombatTeam, Hurtbox } from "../combat/Combatant";
import type { CombatState } from "../combat/CombatState";
import type { RemotePlayer } from "../network/RemotePlayer";

export class PvpTarget implements Combatant {
  readonly team: CombatTeam = "enemy";
  readonly velocity = new THREE.Vector3();
  readonly health: HealthComponent;
  readonly hurtbox: Hurtbox = { radius: 0.45, height: 1.7, yOffset: 0.85 };
  facingYaw = 0;
  invincible = false;
  combatState: CombatState = "IDLE";
  name: string;
  level: number;

  constructor(private readonly remote: RemotePlayer) {
    this.health = new HealthComponent(remote.state.maxHp || 100, remote.state.hp);
    this.name = remote.state.name;
    this.level = remote.state.level;
  }

  get id(): string {
    return this.remote.id;
  }

  get position(): THREE.Vector3 {
    return this.remote.group.position;
  }

  get group(): THREE.Object3D {
    return this.remote.group;
  }

  sync(): void {
    this.health.hp = this.remote.state.hp;
    this.health.maxHp = this.remote.state.maxHp || this.health.maxHp;
    this.name = this.remote.state.name;
    this.level = this.remote.state.level;
    this.combatState = (this.remote.combatPose as CombatState) || "IDLE";
  }
}

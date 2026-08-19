import * as THREE from "three";
import { PLAYER } from "../game/GameConfig";
import { PlayerStats } from "./PlayerStats";
import { buildPlayerModel, disposePlayerModel } from "./PlayerModel";
import { PLAYER_FORMS, type PlayerFormId } from "./PlayerForms";
import type { CombatState } from "../combat/CombatState";
import type { Combatant, CombatTeam, Hurtbox } from "../combat/Combatant";
import type { HealthComponent } from "../combat/HealthComponent";

export type AnimState =
  | "IDLE"
  | "WALK"
  | "RUN"
  | "JUMP"
  | "FALL"
  | "LAND"
  | "PUNCH"
  | "KICK"
  | "COMBO"
  | "DODGE"
  | "ENERGY_ATTACK"
  | "HIT"
  | "STUN"
  | "DEAD";

export class Player implements Combatant {
  readonly id = "player";
  readonly team: CombatTeam = "player";
  readonly group = new THREE.Group();
  readonly mesh = new THREE.Group();
  readonly stats = new PlayerStats();
  readonly velocity = new THREE.Vector3();
  readonly hurtbox: Hurtbox = { radius: PLAYER.radius, height: PLAYER.height, yOffset: 0.85 };
  isGrounded = true;
  facingYaw = 0;
  currentSpeed = 0;
  animState: AnimState = "IDLE";
  combatPose: AnimState | null = null;
  combatState: CombatState = "IDLE";
  invincible = false;
  landTimer = 0;
  formId: PlayerFormId = "asal";
  className = "WARRIOR";

  armL?: THREE.Object3D;
  armR?: THREE.Object3D;
  legL?: THREE.Object3D;
  legR?: THREE.Object3D;
  aura?: THREE.Object3D;
  lightning?: THREE.Object3D;

  constructor() {
    this.group.add(this.mesh);
    this.group.position.set(0, PLAYER.groundY, 6);
    void this.loadPlayerModel();
  }

  formName(): string {
    return PLAYER_FORMS[this.formId].name;
  }

  get health(): HealthComponent {
    return this.stats.health;
  }

  setForm(id: PlayerFormId): void {
    if (this.formId === id && this.mesh.children.length > 0) return;
    this.formId = id;
    this.rebuild();
  }

  /** Ganti placeholder dengan `player.glb` nanti tanpa mengubah PlayerController. */
  async loadPlayerModel(url?: string): Promise<void> {
    if (url) {
      console.info("Player model GLB belum dipasang, url diabaikan:", url);
    }
    this.rebuild();
  }

  get position(): THREE.Vector3 {
    return this.group.position;
  }

  private rebuild(): void {
    const yaw = this.mesh.rotation.y;
    disposePlayerModel(this.mesh);
    this.mesh.clear();
    const rig = buildPlayerModel(this.formId);
    this.mesh.add(rig.root);
    this.armL = rig.armL;
    this.armR = rig.armR;
    this.legL = rig.legL;
    this.legR = rig.legR;
    this.aura = rig.aura;
    this.lightning = rig.lightning;
    this.mesh.rotation.y = yaw;
  }
}

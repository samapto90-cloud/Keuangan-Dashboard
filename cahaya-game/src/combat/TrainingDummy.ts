import * as THREE from "three";
import { PLAYER } from "../game/GameConfig";
import { HealthComponent } from "./HealthComponent";
import type { Combatant, CombatTeam, Hurtbox } from "./Combatant";
import type { CombatState } from "./CombatState";

export class TrainingDummy implements Combatant {
  readonly id: string;
  readonly team: CombatTeam = "enemy";
  readonly group = new THREE.Group();
  readonly mesh = new THREE.Group();
  readonly velocity = new THREE.Vector3();
  readonly health: HealthComponent;
  readonly hurtbox: Hurtbox = { radius: 0.42, height: 1.55, yOffset: 0.85 };
  facingYaw = Math.PI;
  invincible = false;
  combatState: CombatState = "IDLE";
  stunTimer = 0;
  name: string;
  private readonly bodyMat: THREE.MeshStandardMaterial;
  private hitFlash = 0;
  private deadPose = false;

  constructor(id: string, x: number, z: number, name: string) {
    this.id = id;
    this.name = name;
    this.health = new HealthComponent(100);
    this.group.position.set(x, PLAYER.groundY, z);
    this.group.add(this.mesh);
    this.bodyMat = new THREE.MeshStandardMaterial({ color: 0xc4a574, roughness: 0.85 });
    this.build();
  }

  get position(): THREE.Vector3 {
    return this.group.position;
  }

  applyHit(knockbackDir: THREE.Vector3, force: number, stun: number): void {
    if (this.health.isDead()) return;
    this.velocity.x += knockbackDir.x * force;
    this.velocity.z += knockbackDir.z * force;
    this.stunTimer = Math.max(this.stunTimer, stun);
    this.hitFlash = 0.14;
    this.combatState = stun > 0.18 ? "STUNNED" : "HIT";
  }

  update(dt: number): void {
    if (this.stunTimer > 0) this.stunTimer -= dt;
    if (this.hitFlash > 0) {
      this.hitFlash -= dt;
      this.bodyMat.emissive.setHex(0xff6644);
      this.bodyMat.emissiveIntensity = 0.7;
    } else {
      this.bodyMat.emissive.setHex(0x000000);
      this.bodyMat.emissiveIntensity = 0;
    }

    this.velocity.y -= 22 * dt;
    this.velocity.x *= Math.exp(-6 * dt);
    this.velocity.z *= Math.exp(-6 * dt);
    this.position.x += this.velocity.x * dt;
    this.position.y += this.velocity.y * dt;
    this.position.z += this.velocity.z * dt;
    if (this.position.y <= PLAYER.groundY) {
      this.position.y = PLAYER.groundY;
      this.velocity.y = 0;
    }

    if (this.health.isDead()) {
      this.combatState = "DEAD";
      this.invincible = true;
      if (!this.deadPose) {
        this.deadPose = true;
        this.mesh.rotation.x = Math.PI / 2;
        this.mesh.position.y = 0.28;
      }
      return;
    }

    if (this.stunTimer <= 0 && (this.combatState === "HIT" || this.combatState === "STUNNED")) {
      this.combatState = "IDLE";
    }
    if (this.combatState === "HIT" || this.combatState === "STUNNED") {
      this.mesh.rotation.x = -0.18;
    } else {
      this.mesh.rotation.x = 0;
    }
  }

  private build(): void {
    const pole = new THREE.Mesh(new THREE.CylinderGeometry(0.08, 0.1, 0.35, 8), new THREE.MeshStandardMaterial({ color: 0x6b4423 }));
    pole.position.y = 0.18;
    const body = new THREE.Mesh(new THREE.CylinderGeometry(0.28, 0.32, 0.95, 10), this.bodyMat);
    body.position.y = 0.85;
    const head = new THREE.Mesh(new THREE.SphereGeometry(0.22, 12, 12), this.bodyMat);
    head.position.y = 1.48;
    const sash = new THREE.Mesh(
      new THREE.TorusGeometry(0.3, 0.04, 6, 12),
      new THREE.MeshStandardMaterial({ color: 0x1aa6a0, roughness: 0.5 }),
    );
    sash.position.y = 1.22;
    sash.rotation.x = Math.PI / 2;
    this.mesh.add(pole, body, head, sash);
    this.mesh.traverse((o) => {
      if (o instanceof THREE.Mesh) o.castShadow = true;
    });
  }
}

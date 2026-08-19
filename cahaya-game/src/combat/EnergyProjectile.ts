import * as THREE from "three";
import type { AttackData } from "./AttackData";
import type { Combatant } from "./Combatant";

export interface Projectile {
  active: boolean;
  ownerId: string;
  damage: number;
  radius: number;
  lifetime: number;
  age: number;
  knockback: number;
  stunDuration: number;
  source: string;
  mesh: THREE.Mesh;
  velocity: THREE.Vector3;
}

const POOL = 8;

export class ProjectilePool {
  readonly items: Projectile[] = [];
  private readonly scratch = new THREE.Vector3();

  constructor(scene: THREE.Scene) {
    const geo = new THREE.SphereGeometry(0.22, 12, 12);
    for (let i = 0; i < POOL; i++) {
      const mat = new THREE.MeshStandardMaterial({
        color: 0x9ef6ff,
        emissive: 0x3ec4ff,
        emissiveIntensity: 1.6,
        transparent: true,
        opacity: 0.92,
      });
      const mesh = new THREE.Mesh(geo, mat);
      mesh.visible = false;
      mesh.castShadow = false;
      scene.add(mesh);
      this.items.push({
        active: false,
        ownerId: "",
        damage: 0,
        radius: 0.5,
        lifetime: 1.2,
        age: 0,
        knockback: 0,
        stunDuration: 0,
        source: "ENERGY_1",
        mesh,
        velocity: new THREE.Vector3(),
      });
    }
  }

  spawn(
    origin: THREE.Vector3,
    direction: THREE.Vector3,
    ownerId: string,
    attack: AttackData,
  ): Projectile | null {
    const item = this.items.find((p) => !p.active);
    if (!item) return null;
    const dir = direction.clone().normalize();
    item.active = true;
    item.ownerId = ownerId;
    item.damage = attack.damage;
    item.radius = attack.hitboxRadius;
    item.lifetime = 1.2;
    item.age = 0;
    item.knockback = attack.knockback;
    item.stunDuration = attack.stunDuration;
    item.source = attack.id;
    item.velocity.copy(dir).multiplyScalar(14);
    item.mesh.position.copy(origin);
    item.mesh.visible = true;
    item.mesh.scale.set(1, 1, 1);
    return item;
  }

  update(dt: number, combatants: Combatant[], onHit: (proj: Projectile, target: Combatant) => void): void {
    for (const proj of this.items) {
      if (!proj.active) continue;
      proj.age += dt;
      proj.mesh.position.addScaledVector(proj.velocity, dt);
      proj.mesh.rotation.y += dt * 8;
      const pulse = 1 + Math.sin(proj.age * 18) * 0.08;
      proj.mesh.scale.setScalar(pulse);
      if (proj.age >= proj.lifetime) {
        this.release(proj);
        continue;
      }
      for (const target of combatants) {
        if (!proj.active) break;
        if (target.id === proj.ownerId || target.health.isDead() || target.invincible) continue;
        this.scratch.set(
          target.position.x,
          target.position.y + target.hurtbox.yOffset,
          target.position.z,
        );
        if (proj.mesh.position.distanceTo(this.scratch) <= proj.radius + target.hurtbox.radius) {
          onHit(proj, target);
          this.release(proj);
        }
      }
    }
  }

  release(proj: Projectile): void {
    proj.active = false;
    proj.mesh.visible = false;
  }

  dispose(): void {
    for (const proj of this.items) {
      proj.mesh.geometry.dispose();
      const mat = proj.mesh.material;
      if (Array.isArray(mat)) mat.forEach((m) => m.dispose());
      else mat.dispose();
      proj.mesh.parent?.remove(proj.mesh);
    }
  }
}

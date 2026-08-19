import * as THREE from "three";
import type { Combatant } from "./Combatant";

export class TargetSystem {
  current: Combatant | null = null;
  readonly indicator: THREE.Mesh;
  private readonly candidates: Combatant[] = [];

  constructor() {
    this.indicator = new THREE.Mesh(
      new THREE.TorusGeometry(0.55, 0.045, 8, 24),
      new THREE.MeshBasicMaterial({ color: 0xffe08a, transparent: true, opacity: 0.9 }),
    );
    this.indicator.rotation.x = Math.PI / 2;
    this.indicator.visible = false;
  }

  setCandidates(list: Combatant[]): void {
    this.candidates.length = 0;
    this.candidates.push(...list);
  }

  cycle(from: THREE.Vector3): Combatant | null {
    const alive = this.candidates
      .filter((c) => !c.health.isDead())
      .sort((a, b) => a.position.distanceToSquared(from) - b.position.distanceToSquared(from));
    if (alive.length === 0) {
      this.clear();
      return null;
    }
    if (!this.current || this.current.health.isDead()) {
      this.current = alive[0];
      return this.current;
    }
    const idx = alive.findIndex((c) => c.id === this.current?.id);
    this.current = alive[(idx + 1) % alive.length];
    return this.current;
  }

  clear(): void {
    this.current = null;
    this.indicator.visible = false;
  }

  update(from: THREE.Vector3): void {
    if (this.current?.health.isDead()) this.clear();
    if (!this.current) {
      this.indicator.visible = false;
      return;
    }
    const dist = this.current.position.distanceTo(from);
    if (dist > 22) {
      this.clear();
      return;
    }
    this.indicator.visible = true;
    this.indicator.position.set(
      this.current.position.x,
      this.current.position.y + this.current.hurtbox.height + 0.25,
      this.current.position.z,
    );
    this.indicator.rotation.z += 0.03;
  }

  aimDirection(from: THREE.Vector3, fallbackYaw: number, out: THREE.Vector3): THREE.Vector3 {
    if (this.current && !this.current.health.isDead()) {
      return out
        .set(
          this.current.position.x - from.x,
          this.current.position.y + this.current.hurtbox.yOffset - (from.y + 1.15),
          this.current.position.z - from.z,
        )
        .normalize();
    }
    return out.set(Math.sin(fallbackYaw), 0.05, Math.cos(fallbackYaw)).normalize();
  }
}

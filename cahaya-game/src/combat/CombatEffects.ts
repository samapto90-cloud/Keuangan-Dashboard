import * as THREE from "three";
import { ProjectilePool } from "./EnergyProjectile";
import { ATTACKS } from "./AttackData";

export class CombatEffects {
  private readonly puffs: THREE.Mesh[] = [];
  private readonly scratch = new THREE.Vector3();
  private readonly aura: THREE.Mesh;
  private charging = false;

  constructor(
    scene: THREE.Scene,
    readonly bolts: ProjectilePool,
  ) {
    const geo = new THREE.SphereGeometry(0.08, 6, 6);
    for (let i = 0; i < 14; i++) {
      const mat = new THREE.MeshBasicMaterial({ color: i % 3 ? 0xffe08a : 0xf8fafc, transparent: true, opacity: 0 });
      const mesh = new THREE.Mesh(geo, mat);
      mesh.visible = false;
      scene.add(mesh);
      this.puffs.push(mesh);
    }
    this.aura = new THREE.Mesh(
      new THREE.SphereGeometry(0.85, 12, 12),
      new THREE.MeshBasicMaterial({ color: 0xfde68a, transparent: true, opacity: 0.22, depthWrite: false }),
    );
    this.aura.visible = false;
    scene.add(this.aura);
  }

  setCharge(on: boolean, at: THREE.Vector3): void {
    this.charging = on;
    this.aura.visible = on;
    if (on) this.aura.position.set(at.x, at.y + 1.0, at.z);
  }

  impact(at: THREE.Vector3): void {
    const puff = this.puffs.find((p) => !p.visible) ?? this.puffs[0];
    if (!puff) return;
    puff.position.copy(at);
    puff.scale.setScalar(1);
    puff.visible = true;
    const mat = puff.material as THREE.MeshBasicMaterial;
    mat.opacity = 0.85;
  }

  bolt(from: THREE.Vector3, dir: THREE.Vector3, ownerId: string): void {
    this.bolts.spawn(from, dir, ownerId, ATTACKS.ENERGY_1);
  }

  update(dt: number): void {
    for (const puff of this.puffs) {
      if (!puff.visible) continue;
      const mat = puff.material as THREE.MeshBasicMaterial;
      mat.opacity -= dt * 2.4;
      puff.scale.multiplyScalar(1 + dt * 3);
      if (mat.opacity <= 0) puff.visible = false;
    }
    if (this.charging) {
      const pulse = 0.85 + Math.sin(performance.now() * 0.008) * 0.12;
      this.aura.scale.setScalar(pulse);
      const mat = this.aura.material as THREE.MeshBasicMaterial;
      mat.opacity = 0.18 + Math.sin(performance.now() * 0.01) * 0.06;
    }
    this.scratch.set(0, 0, 0);
  }

  dispose(): void {
    this.aura.parent?.remove(this.aura);
    this.bolts.dispose();
  }
}

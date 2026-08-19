import * as THREE from "three";
import type { TransformView } from "./ProgressionStore";

const PALETTE: Record<string, number> = {
  "aura-1": 0xfde68a,
  "aura-2": 0x67e8f9,
  "aura-3": 0x86efac,
  "celestial-4": 0xc4b5fd,
};

export class TransformationFX {
  private readonly group = new THREE.Group();
  private readonly sparks: THREE.Points;
  private readonly ring: THREE.Mesh;
  private readonly wave: THREE.Mesh;
  private timer = 0;
  private color = 0xfde68a;

  constructor(scene: THREE.Scene) {
    const sparkGeo = new THREE.BufferGeometry();
    const count = 72;
    const pos = new Float32Array(count * 3);
    sparkGeo.setAttribute("position", new THREE.BufferAttribute(pos, 3));
    this.sparks = new THREE.Points(
      sparkGeo,
      new THREE.PointsMaterial({ color: this.color, size: 0.08, transparent: true, opacity: 0.85, depthWrite: false }),
    );
    this.ring = new THREE.Mesh(
      new THREE.RingGeometry(0.6, 0.85, 24),
      new THREE.MeshBasicMaterial({ color: this.color, transparent: true, opacity: 0, side: THREE.DoubleSide, depthWrite: false }),
    );
    this.ring.rotation.x = -Math.PI / 2;
    this.wave = new THREE.Mesh(
      new THREE.CircleGeometry(0.4, 20),
      new THREE.MeshBasicMaterial({ color: 0xfff7d6, transparent: true, opacity: 0, depthWrite: false }),
    );
    this.wave.rotation.x = -Math.PI / 2;
    this.group.add(this.sparks, this.ring, this.wave);
    this.group.visible = false;
    scene.add(this.group);
  }

  burst(at: THREE.Vector3, view: TransformView): void {
    this.color = PALETTE[view.formId] ?? 0xfde68a;
    (this.sparks.material as THREE.PointsMaterial).color.setHex(this.color);
    (this.ring.material as THREE.MeshBasicMaterial).color.setHex(this.color);
    this.group.position.copy(at);
    this.group.position.y = 0.05;
    this.group.visible = true;
    this.timer = 1.65;
    (this.ring.material as THREE.MeshBasicMaterial).opacity = 0.85;
    (this.wave.material as THREE.MeshBasicMaterial).opacity = 0.7;
    this.ring.scale.setScalar(1);
    this.wave.scale.setScalar(1);
  }

  follow(at: THREE.Vector3, formId: string, active: boolean, dt: number): void {
    if (!active) {
      if (this.timer <= 0) this.group.visible = false;
      else this.tick(dt);
      return;
    }
    this.group.visible = true;
    this.group.position.copy(at);
    this.group.position.y = 0.04;
    this.color = PALETTE[formId] ?? this.color;
    (this.sparks.material as THREE.PointsMaterial).color.setHex(this.color);
    const pos = this.sparks.geometry.getAttribute("position") as THREE.BufferAttribute;
    const n = pos.count;
    const t = performance.now() * 0.004;
    const spread = formId === "celestial-4" ? 1.4 : formId === "aura-3" ? 1.15 : formId === "aura-2" ? 0.95 : 0.7;
    for (let i = 0; i < n; i++) {
      const a = (i / n) * Math.PI * 2 + t;
      pos.setXYZ(i, Math.cos(a) * spread * (0.4 + (i % 5) * 0.12), 0.3 + (i % 7) * 0.16, Math.sin(a) * spread * (0.4 + (i % 4) * 0.12));
    }
    pos.needsUpdate = true;
    this.tick(dt);
  }

  private tick(dt: number): void {
    if (this.timer <= 0) return;
    this.timer -= dt;
    this.ring.scale.multiplyScalar(1 + dt * 3.2);
    this.wave.scale.multiplyScalar(1 + dt * 4.5);
    (this.ring.material as THREE.MeshBasicMaterial).opacity *= 0.9;
    (this.wave.material as THREE.MeshBasicMaterial).opacity *= 0.86;
  }
}

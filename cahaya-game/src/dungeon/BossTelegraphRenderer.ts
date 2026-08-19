import * as THREE from "three";
import type { BossAOEOut, BossTelegraphOut } from "../network/NetworkMessage";

export class BossTelegraphRenderer {
  private readonly group = new THREE.Group();
  private mesh: THREE.Object3D | null = null;
  private until = 0;
  private pulse = false;
  private born = 0;
  vfxOn = true;

  constructor(scene: THREE.Scene) {
    scene.add(this.group);
  }

  show(ev: BossTelegraphOut): void {
    this.clear();
    if (!this.vfxOn) {
      this.until = ev.until || Date.now() + 1200;
      return;
    }
    const r = Math.max(1.2, ev.radius || 1.8);
    const shape = (ev.shape || "circle").toLowerCase();
    const root = new THREE.Group();
    const fill = new THREE.MeshBasicMaterial({ color: 0xef4444, transparent: true, opacity: 0.32, depthWrite: false });
    const edge = new THREE.LineBasicMaterial({ color: 0xfde68a, transparent: true, opacity: 0.95 });
    if (shape === "line") {
      const geo = new THREE.PlaneGeometry(Math.max(1.2, r), 10);
      const mesh = new THREE.Mesh(geo, fill);
      mesh.rotation.x = -Math.PI / 2;
      mesh.position.set(ev.x, 0.08, ev.z);
      root.add(mesh);
    } else if (shape === "cone") {
      const geo = new THREE.CircleGeometry(r, 24, 0, Math.PI / 2.4);
      const mesh = new THREE.Mesh(geo, fill);
      mesh.rotation.x = -Math.PI / 2;
      mesh.position.set(ev.x, 0.08, ev.z);
      root.add(mesh);
    } else if (shape === "ring") {
      const ring = new THREE.Mesh(new THREE.RingGeometry(Math.max(0.4, r - 1.3), r, 28), fill);
      ring.rotation.x = -Math.PI / 2;
      ring.position.set(ev.x, 0.08, ev.z);
      root.add(ring);
    } else {
      const mesh = new THREE.Mesh(new THREE.CircleGeometry(r, 28), fill);
      mesh.rotation.x = -Math.PI / 2;
      mesh.position.set(ev.x, 0.08, ev.z);
      root.add(mesh);
    }
    const loop = new THREE.LineLoop(new THREE.EdgesGeometry(new THREE.CircleGeometry(r, 20)), edge);
    loop.rotation.x = -Math.PI / 2;
    loop.position.set(ev.x, 0.1, ev.z);
    root.add(loop);
    const mark = new THREE.Mesh(new THREE.SphereGeometry(0.18, 8, 8), new THREE.MeshBasicMaterial({ color: 0xfacc15 }));
    mark.position.set(ev.x, 0.35, ev.z);
    root.add(mark);
    this.mesh = root;
    this.group.add(root);
    this.until = ev.until || Date.now() + 1200;
    this.pulse = ev.pulse !== false;
    this.born = performance.now();
  }

  impact(ev: BossAOEOut): void {
    this.show({ instanceId: ev.instanceId, skill: "AOE", x: ev.x, z: ev.z, radius: ev.radius, until: Date.now() + 280, vfx: "impact", pulse: true, shape: "circle" });
  }

  update(): void {
    if (this.until && Date.now() > this.until) this.clear();
    if (this.pulse && this.mesh) {
      const s = 1 + Math.sin((performance.now() - this.born) / 140) * 0.08;
      this.mesh.scale.set(s, 1, s);
    }
  }

  clear(): void {
    this.mesh?.removeFromParent();
    this.mesh = null;
    this.until = 0;
  }
}

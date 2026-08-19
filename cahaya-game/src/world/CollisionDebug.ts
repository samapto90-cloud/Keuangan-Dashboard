import * as THREE from "three";
import type { CollisionWorld } from "./Collision";
import {
  NPC_AVOIDANCE_RADIUS,
  NPC_PERSONAL_SPACE,
  PLAYER_CAPSULE,
  VILLAGE_AABBS,
  VILLAGE_CIRCLES,
  VILLAGE_ROCKS,
  VILLAGE_TREES,
} from "./WorldColliders";

const COLORS: Record<string, number> = {
  BUILDING: 0xf97316,
  OBSTACLE: 0xef4444,
  WORLD: 0x3b82f6,
  PLAYER: 0x22c55e,
  NPC: 0xa855f7,
};

export class CollisionDebug {
  readonly group = new THREE.Group();
  private readonly colliders = new THREE.Group();
  private readonly crowd = new THREE.Group();
  private showCrowd = false;

  constructor(
    scene: THREE.Scene,
    private readonly collision: CollisionWorld,
  ) {
    scene.add(this.group);
    this.group.add(this.colliders, this.crowd);
    this.colliders.visible = false;
    this.crowd.visible = false;
    this.buildStaticColliders();
  }

  setShowColliders(on: boolean): void {
    this.colliders.visible = on;
  }

  setShowCrowd(on: boolean): void {
    this.showCrowd = on;
    this.crowd.visible = on;
  }

  update(
    playerX: number,
    playerZ: number,
    npcs: Array<{ x: number; z: number; type: string }>,
  ): void {
    if (!this.showCrowd) return;
    while (this.crowd.children.length) {
      this.crowd.remove(this.crowd.children[0]!);
    }
    this.crowd.add(ring(playerX, playerZ, PLAYER_CAPSULE.radius, COLORS.PLAYER ?? 0x22c55e, 0.04));
    for (const n of npcs) {
      const r = n.type === "CHILD" ? 0.26 : 0.32;
      this.crowd.add(ring(n.x, n.z, r, COLORS.NPC ?? 0xa855f7, 0.035));
      this.crowd.add(ring(n.x, n.z, NPC_PERSONAL_SPACE, COLORS.NPC ?? 0xa855f7, 0.015, 0.35));
      this.crowd.add(ring(n.x, n.z, NPC_AVOIDANCE_RADIUS, COLORS.NPC ?? 0xa855f7, 0.012, 0.2));
    }
  }

  private buildStaticColliders(): void {
    const mat = (layer: string, opacity = 0.55) =>
      new THREE.MeshBasicMaterial({
        color: COLORS[layer] ?? 0xffffff,
        wireframe: true,
        transparent: true,
        opacity,
        depthTest: false,
      });
    for (const b of this.collision.aabbs.length ? this.collision.aabbs : VILLAGE_AABBS) {
      const box = new THREE.Mesh(new THREE.BoxGeometry(b.w, 0.08, b.d), mat(b.layer));
      box.position.set(b.x, 0.12, b.z);
      this.colliders.add(box);
    }
    const circles = [...VILLAGE_CIRCLES, ...VILLAGE_TREES, ...VILLAGE_ROCKS];
    for (const c of circles) {
      if (c.walkable) continue;
      this.colliders.add(ring(c.x, c.z, c.r, COLORS[c.layer] ?? 0xef4444, 0.06));
    }
  }
}

function ring(x: number, z: number, radius: number, color: number, y = 0.05, opacity = 0.7): THREE.Mesh {
  const geo = new THREE.RingGeometry(radius * 0.92, radius, 24);
  const mat = new THREE.MeshBasicMaterial({ color, transparent: true, opacity, side: THREE.DoubleSide, depthTest: false });
  const mesh = new THREE.Mesh(geo, mat);
  mesh.rotation.x = -Math.PI / 2;
  mesh.position.set(x, y, z);
  return mesh;
}

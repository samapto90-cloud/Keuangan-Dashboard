import * as THREE from "three";
import type { ObjectSnapshot } from "../network/NetworkMessage";

export class WorldInteraction {
  readonly group = new THREE.Group();
  private readonly nodes = new Map<string, THREE.Object3D>();
  private readonly snaps = new Map<string, ObjectSnapshot>();

  constructor(scene: THREE.Scene) {
    scene.add(this.group);
  }

  sync(list: ObjectSnapshot[], claimed: (id: string) => boolean, forestOpen: boolean): void {
    const seen = new Set<string>();
    for (const snap of list) {
      seen.add(snap.id);
      let node = this.nodes.get(snap.id);
      if (!node) {
        node = buildObject(snap);
        this.nodes.set(snap.id, node);
        this.group.add(node);
      }
      this.snaps.set(snap.id, snap);
      node.visible = !(snap.kind === "crystal" && claimed(snap.id));
      if (snap.kind === "chest") node.rotation.y = claimed(snap.id) ? 0.6 : 0;
      if (snap.kind === "gate") {
        node.traverse((c) => {
          if (c instanceof THREE.Mesh && c.userData.bar) c.visible = !forestOpen;
        });
      }
    }
    for (const [id, node] of this.nodes) {
      if (seen.has(id)) continue;
      node.removeFromParent();
      this.nodes.delete(id);
      this.snaps.delete(id);
    }
  }

  nearest(x: number, z: number, maxDist: number): ObjectSnapshot | null {
    let best: ObjectSnapshot | null = null;
    let bestD = maxDist;
    for (const snap of this.snaps.values()) {
      const d = Math.hypot(snap.x - x, snap.z - z);
      if (d < bestD) {
        bestD = d;
        best = snap;
      }
    }
    return best;
  }

  all(): ObjectSnapshot[] {
    return [...this.snaps.values()];
  }
}

function buildObject(snap: ObjectSnapshot): THREE.Object3D {
  const g = new THREE.Group();
  g.position.set(snap.x, 0, snap.z);
  if (snap.kind === "chest") {
    const box = new THREE.Mesh(
      new THREE.BoxGeometry(0.7, 0.45, 0.5),
      new THREE.MeshStandardMaterial({ color: 0xb45309, roughness: 0.55, metalness: 0.2 }),
    );
    box.position.y = 0.28;
    box.castShadow = true;
    g.add(box);
  } else if (snap.kind === "crystal") {
    const gem = new THREE.Mesh(
      new THREE.OctahedronGeometry(0.28, 0),
      new THREE.MeshStandardMaterial({ color: 0x67e8f9, emissive: 0x22d3ee, emissiveIntensity: 0.9, roughness: 0.2 }),
    );
    gem.position.y = 0.55;
    gem.castShadow = true;
    g.add(gem);
  } else if (snap.kind === "drop") {
    const gem = new THREE.Mesh(
      new THREE.OctahedronGeometry(0.22, 0),
      new THREE.MeshStandardMaterial({ color: 0xfbbf24, emissive: 0xf59e0b, emissiveIntensity: 1.1, roughness: 0.25 }),
    );
    gem.position.y = 0.45;
    const ring = new THREE.Mesh(
      new THREE.TorusGeometry(0.28, 0.03, 6, 16),
      new THREE.MeshStandardMaterial({ color: 0xfde68a, emissive: 0xfbbf24, emissiveIntensity: 0.5 }),
    );
    ring.rotation.x = Math.PI / 2;
    ring.position.y = 0.08;
    g.add(gem, ring);
  } else if (snap.kind === "checkpoint") {
    const glow = new THREE.Mesh(
      new THREE.SphereGeometry(0.42, 14, 14),
      new THREE.MeshStandardMaterial({ color: 0xfde68a, emissive: 0xf59e0b, emissiveIntensity: 1.2, roughness: 0.3 }),
    );
    glow.position.y = 0.9;
    const ring = new THREE.Mesh(
      new THREE.TorusGeometry(0.7, 0.05, 8, 20),
      new THREE.MeshStandardMaterial({ color: 0xfbbf24, emissive: 0xf59e0b, emissiveIntensity: 0.6 }),
    );
    ring.rotation.x = Math.PI / 2;
    ring.position.y = 0.06;
    g.add(glow, ring);
  } else if (snap.kind === "gate") {
    const post = new THREE.Mesh(
      new THREE.BoxGeometry(0.35, 3.2, 0.35),
      new THREE.MeshStandardMaterial({ color: 0x78716c, roughness: 0.7 }),
    );
    const left = post.clone();
    left.position.set(-1.6, 1.6, 0);
    const right = post.clone();
    right.position.set(1.6, 1.6, 0);
    const bar = new THREE.Mesh(
      new THREE.BoxGeometry(3.2, 0.18, 0.18),
      new THREE.MeshStandardMaterial({ color: 0xf87171, emissive: 0x991b1b, emissiveIntensity: 0.4 }),
    );
    bar.position.set(0, 1.5, 0);
    bar.userData.bar = true;
    const arch = new THREE.Mesh(
      new THREE.BoxGeometry(3.6, 0.28, 0.4),
      new THREE.MeshStandardMaterial({ color: 0xa8a29e }),
    );
    arch.position.set(0, 3.2, 0);
    g.add(left, right, bar, arch);
  } else if (snap.kind === "dungeon") {
    const stone = new THREE.Mesh(
      new THREE.CylinderGeometry(0.7, 0.85, 0.35, 8),
      new THREE.MeshStandardMaterial({ color: 0x44403c, roughness: 0.7 }),
    );
    stone.position.y = 0.18;
    const glow = new THREE.Mesh(
      new THREE.TorusGeometry(0.85, 0.06, 8, 20),
      new THREE.MeshStandardMaterial({ color: 0x67e8f9, emissive: 0x22d3ee, emissiveIntensity: 0.9 }),
    );
    glow.rotation.x = Math.PI / 2;
    glow.position.y = 0.4;
    g.add(stone, glow);
    const sign = new THREE.Mesh(
      new THREE.BoxGeometry(0.7, 0.9, 0.08),
      new THREE.MeshStandardMaterial({ color: 0xd6b07a }),
    );
    sign.position.y = 1.1;
    g.add(sign);
  } else if (snap.kind.startsWith("gather-")) {
    const color = snap.kind.includes("ore") || snap.kind.includes("stone") ? 0x78716c : snap.kind.includes("wood") ? 0x3f6212 : snap.kind.includes("fiber") ? 0xb91c1c : 0x4ade80;
    const node = new THREE.Mesh(
      new THREE.DodecahedronGeometry(0.32, 0),
      new THREE.MeshStandardMaterial({ color, roughness: 0.55, emissive: color, emissiveIntensity: 0.18 }),
    );
    node.position.y = 0.42;
    node.castShadow = true;
    g.add(node);
  } else if (snap.kind === "forge" || snap.kind === "workbench" || snap.kind === "cooking-fire" || snap.kind === "alchemy") {
    const col = snap.kind === "forge" ? 0xf97316 : snap.kind === "cooking-fire" ? 0xef4444 : snap.kind === "alchemy" ? 0xa78bfa : 0xc4a574;
    const block = new THREE.Mesh(
      new THREE.BoxGeometry(0.9, 0.55, 0.7),
      new THREE.MeshStandardMaterial({ color: col, roughness: 0.5, metalness: snap.kind === "forge" ? 0.4 : 0.1 }),
    );
    block.position.y = 0.3;
    block.castShadow = true;
    g.add(block);
  } else if (snap.kind === "fishing-spot") {
    const ring = new THREE.Mesh(
      new THREE.TorusGeometry(0.55, 0.05, 8, 20),
      new THREE.MeshStandardMaterial({ color: 0x38bdf8, emissive: 0x0ea5e9, emissiveIntensity: 0.7 }),
    );
    ring.rotation.x = Math.PI / 2;
    ring.position.y = 0.08;
    g.add(ring);
  }
  return g;
}

export interface FastTravelPoint {
  id: string;
  name: string;
  x: number;
  z: number;
  discovered: boolean;
}

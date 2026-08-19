import * as THREE from "three";
import type { LandmarkView } from "../network/NetworkMessage";

export class LandmarkSystem {
  readonly group = new THREE.Group();
  private readonly byId = new Map<string, THREE.Object3D>();

  constructor(parent: THREE.Object3D) {
    parent.add(this.group);
  }

  sync(list: LandmarkView[], playerZ: number): void {
    const seen = new Set<string>();
    for (const lm of list) {
      seen.add(lm.id);
      let mesh = this.byId.get(lm.id);
      if (!mesh) {
        mesh = this.make(lm);
        this.byId.set(lm.id, mesh);
        this.group.add(mesh);
      }
      mesh.position.set(lm.x, 0, lm.z);
      const far = tallLandmark(lm.id) ? 120 : 48;
      mesh.visible = Math.abs(playerZ - lm.z) < far;
      mesh.userData.discovered = lm.discovered;
    }
    for (const [id, mesh] of this.byId) {
      if (!seen.has(id)) {
        this.group.remove(mesh);
        this.byId.delete(id);
      }
    }
  }

  private make(lm: LandmarkView): THREE.Object3D {
    const g = new THREE.Group();
    const id = lm.id.toLowerCase();
    const stone = new THREE.MeshStandardMaterial({ color: 0xc4b49a, roughness: 0.78 });
    if (id.includes("gate") || id.includes("desa")) {
      const a = new THREE.Mesh(new THREE.CylinderGeometry(0.3, 0.36, 3.6, 8), stone);
      a.position.set(-1.2, 1.8, 0);
      const b = a.clone();
      b.position.x = 1.2;
      const arch = new THREE.Mesh(new THREE.TorusGeometry(1.3, 0.2, 8, 14, Math.PI), stone);
      arch.position.y = 3.2;
      arch.rotation.z = Math.PI;
      g.add(a, b, arch);
    } else if (id.includes("hutan") || id.includes("pohon") || id.includes("tree")) {
      const trunk = new THREE.Mesh(new THREE.CylinderGeometry(0.4, 0.55, 4.2, 8), new THREE.MeshStandardMaterial({ color: 0x5c3a24, roughness: 0.9 }));
      trunk.position.y = 2.1;
      const crown = new THREE.Mesh(new THREE.SphereGeometry(1.8, 10, 8), new THREE.MeshStandardMaterial({ color: 0x2f7a3a, roughness: 0.75 }));
      crown.position.y = 4.4;
      g.add(trunk, crown);
    } else if (id.includes("gunung") || id.includes("air") || id.includes("fall")) {
      const rock = new THREE.Mesh(new THREE.ConeGeometry(1.8, 5.2, 6), stone);
      rock.position.y = 2.5;
      const fall = new THREE.Mesh(
        new THREE.PlaneGeometry(0.7, 4.2),
        new THREE.MeshStandardMaterial({ color: 0xa8e4ff, transparent: true, opacity: 0.5, roughness: 0.12, side: THREE.DoubleSide }),
      );
      fall.position.set(-0.9, 2.2, 0.6);
      g.add(rock, fall);
    } else if (id.includes("kota") || id.includes("menara") || id.includes("tower")) {
      const tower = new THREE.Mesh(new THREE.CylinderGeometry(0.55, 0.7, 5.2, 10), stone);
      tower.position.y = 2.6;
      g.add(tower);
    } else if (id.includes("danau") || id.includes("kuil") || id.includes("candi")) {
      const base = new THREE.Mesh(new THREE.CylinderGeometry(1.4, 1.6, 0.4, 10), stone);
      base.position.y = 0.2;
      const spire = new THREE.Mesh(new THREE.ConeGeometry(0.7, 2.2, 8), new THREE.MeshStandardMaterial({ color: 0xf4e4a4, roughness: 0.45, emissive: 0xc9a227, emissiveIntensity: 0.2 }));
      spire.position.y = 1.5;
      g.add(base, spire);
    } else if (id.includes("masjid") || id.includes("sanctum") || id.includes("horizon")) {
      const minaret = new THREE.Mesh(
        new THREE.CylinderGeometry(0.35, 0.42, 7.2, 10),
        new THREE.MeshStandardMaterial({ color: 0xf8f1d4, emissive: 0xffe08a, emissiveIntensity: 0.2, roughness: 0.4 }),
      );
      minaret.position.y = 3.6;
      const dome = new THREE.Mesh(
        new THREE.SphereGeometry(1.6, 12, 10, 0, Math.PI * 2, 0, Math.PI / 2),
        new THREE.MeshStandardMaterial({ color: 0xfff3c4, emissive: 0xffd36a, emissiveIntensity: 0.16, roughness: 0.35 }),
      );
      dome.position.y = 4.4;
      g.add(minaret, dome);
    } else {
      const pillar = new THREE.Mesh(new THREE.CylinderGeometry(0.35, 0.45, 2.4, 8), stone);
      pillar.position.y = 1.2;
      g.add(pillar);
    }
    g.userData.id = lm.id;
    g.traverse((o) => {
      if (o instanceof THREE.Mesh) o.castShadow = true;
    });
    return g;
  }
}

function tallLandmark(id: string): boolean {
  const s = id.toLowerCase();
  return s.includes("masjid") || s.includes("gunung") || s.includes("peak") || s.includes("tower") || s.includes("horizon") || s.includes("celestial") || s.includes("sky");
}

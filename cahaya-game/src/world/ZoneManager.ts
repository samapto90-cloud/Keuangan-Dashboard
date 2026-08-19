import * as THREE from "three";

const REGIONS = [
  { id: "village", name: "Dawn Village", titleId: "DESA AWAL", minZ: -12, maxZ: 22, color: 0x6f9b4e },
  { id: "forest", name: "Mistwood Forest", titleId: "HUTAN LARANGAN", minZ: 22, maxZ: 50, color: 0x2f6b3a },
  { id: "valley", name: "Stone Valley", titleId: "GUNUNG KABUT", minZ: 50, maxZ: 72, color: 0x8a7a5a },
  { id: "plains", name: "River of Light", titleId: "SUNGAI KEHIDUPAN", minZ: 72, maxZ: 94, color: 0x3a6b8a },
  { id: "canyon", name: "Crimson Plains", titleId: "DATARAN MERAH", minZ: 94, maxZ: 114, color: 0xb45a3a },
  { id: "temple", name: "Celestial Mountains", titleId: "KUIL TUA", minZ: 114, maxZ: 130, color: 0x8ec4d4 },
  { id: "ruins", name: "Ancient Desert", titleId: "GURUN PANJANG", minZ: 130, maxZ: 145, color: 0xc4a574 },
  { id: "celestial", name: "Holy Horizon", titleId: "DESA PARA SILUMAN", minZ: 145, maxZ: 158, color: 0xd8c6ff },
  { id: "masjid", name: "Sanctum of Light", titleId: "MASJID CAHAYA", minZ: 158, maxZ: 175, color: 0xf5e6b8 },
  { id: "horizon", name: "Hall of Horizon", titleId: "WILAYAH TERAKHIR", minZ: 175, maxZ: 190, color: 0xc9a227 },
];

export class ZoneManager {
  readonly group = new THREE.Group();
  private readonly strips = new Map<string, THREE.Object3D>();
  private active = "";

  constructor(parent: THREE.Object3D) {
    parent.add(this.group);
    for (const r of REGIONS) this.strips.set(r.id, this.makeStrip(r));
  }

  zoneAt(z: number): (typeof REGIONS)[number] {
    return REGIONS.find((r) => z >= r.minZ && z < r.maxZ) ?? REGIONS[0];
  }

  update(playerZ: number): void {
    const zone = this.zoneAt(playerZ);
    this.active = zone.id;
    for (const r of REGIONS) {
      const mesh = this.strips.get(r.id);
      if (!mesh) continue;
      const dist = Math.min(Math.abs(playerZ - r.minZ), Math.abs(playerZ - r.maxZ), playerZ >= r.minZ && playerZ < r.maxZ ? 0 : 99);
      mesh.visible = dist < 36;
    }
  }

  current(): string {
    return this.active;
  }

  private makeStrip(r: (typeof REGIONS)[number]): THREE.Object3D {
    const g = new THREE.Group();
    const depth = r.maxZ - r.minZ;
    const ground = new THREE.Mesh(
      new THREE.PlaneGeometry(48, depth),
      new THREE.MeshStandardMaterial({ color: r.color, roughness: 0.94, metalness: 0.02 }),
    );
    ground.rotation.x = -Math.PI / 2;
    ground.position.set(0, r.id === "village" ? -0.04 : 0.01, (r.minZ + r.maxZ) / 2);
    ground.receiveShadow = true;
    g.add(ground);
    if (r.id !== "horizon" && r.id !== "masjid") {
      const road = new THREE.Mesh(
        new THREE.PlaneGeometry(4.7, Math.max(8, depth - 2)),
        new THREE.MeshStandardMaterial({ color: 0xd2bc93, roughness: 1 }),
      );
      road.rotation.x = -Math.PI / 2;
      road.position.set(0, 0.032, (r.minZ + r.maxZ) / 2);
      road.receiveShadow = true;
      g.add(road);
    } else if (r.id === "masjid") {
      const approach = new THREE.Mesh(
        new THREE.PlaneGeometry(5.2, depth - 1),
        new THREE.MeshStandardMaterial({ color: 0xe8d7a8, roughness: 0.92 }),
      );
      approach.rotation.x = -Math.PI / 2;
      approach.position.set(0, 0.034, (r.minZ + r.maxZ) / 2);
      approach.receiveShadow = true;
      g.add(approach);
    }
    if (r.id === "horizon") {
      const hall = new THREE.Mesh(
        new THREE.BoxGeometry(18, 6, 8),
        new THREE.MeshStandardMaterial({ color: 0x1a1624, emissive: 0x3b2a08, emissiveIntensity: 0.22, roughness: 0.35, metalness: 0.2, transparent: true, opacity: 0.55 }),
      );
      hall.position.set(0, 3, (r.minZ + r.maxZ) / 2);
      g.add(hall);
    } else if (r.id === "masjid") {
      for (let i = -2; i <= 2; i++) {
        const pillar = new THREE.Mesh(
          new THREE.CylinderGeometry(0.35, 0.4, 4.2, 8),
          new THREE.MeshStandardMaterial({ color: 0xf8f1d4, emissive: 0xffe08a, emissiveIntensity: 0.18, roughness: 0.4 }),
        );
        pillar.position.set(i * 3.2, 2.1, (r.minZ + r.maxZ) / 2);
        g.add(pillar);
      }
      const dome = new THREE.Mesh(
        new THREE.SphereGeometry(3.4, 16, 12, 0, Math.PI * 2, 0, Math.PI / 2),
        new THREE.MeshStandardMaterial({ color: 0xfff3c4, emissive: 0xffd36a, emissiveIntensity: 0.12, roughness: 0.35 }),
      );
      dome.position.set(0, 4.2, (r.minZ + r.maxZ) / 2);
      g.add(dome);
    } else if (r.id !== "village") {
      const fog = new THREE.Mesh(
        new THREE.BoxGeometry(2, 0.4, 2),
        new THREE.MeshStandardMaterial({ color: r.color, transparent: true, opacity: 0.18 }),
      );
      fog.position.set(0, 1.4, (r.minZ + r.maxZ) / 2);
      g.add(fog);
    }
    this.group.add(g);
    g.visible = r.id === "village";
    return g;
  }
}

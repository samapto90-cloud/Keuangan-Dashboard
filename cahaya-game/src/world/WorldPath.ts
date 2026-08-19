import * as THREE from "three";

export const MAIN_PATH_X = 0;
export const MAIN_PATH_HALF = 2.35;

type Sign = { z: number; left: string; right: string };

const SIGNS: Sign[] = [
  { z: 18.4, left: "← Desa", right: "Hutan Larangan →" },
  { z: 34, left: "← Kali", right: "Gunung Kabut →" },
  { z: 49, left: "← Hutan", right: "Lembah Batu →" },
  { z: 70, left: "← Jembatan", right: "Sungai Cahaya →" },
  { z: 92, left: "← Sungai", right: "Dataran Merah →" },
  { z: 112, left: "← Kuil", right: "Pegunungan →" },
  { z: 128, left: "← Menara", right: "Gurun →" },
  { z: 144, left: "← Gurun", right: "Cakrawala →" },
  { z: 157, left: "← Gerbang", right: "Masjid Cahaya →" },
];

const DEAD_ENDS: Array<{ x: number; z: number; kind: "chest" | "lore" | "ruin" }> = [
  { x: -9.2, z: 16.5, kind: "lore" },
  { x: 9.4, z: 28.5, kind: "chest" },
  { x: -10.6, z: 42, kind: "lore" },
  { x: 11.2, z: 40, kind: "chest" },
  { x: -8.8, z: 62, kind: "ruin" },
  { x: 9.6, z: 86, kind: "chest" },
  { x: -9.4, z: 108, kind: "lore" },
  { x: 10.2, z: 138, kind: "ruin" },
];

export class WorldPath {
  readonly group = new THREE.Group();
  readonly waypoints = new THREE.Group();
  private readonly wood = new THREE.MeshStandardMaterial({ color: 0x6b4423, roughness: 0.9 });
  private readonly stone = new THREE.MeshStandardMaterial({ color: 0x8a8f86, roughness: 0.95 });
  private readonly dirtSide = new THREE.MeshStandardMaterial({ color: 0xb09a72, roughness: 1 });
  private readonly dirtSecret = new THREE.MeshStandardMaterial({ color: 0x7d6b52, roughness: 1 });
  private readonly lampGlow = new THREE.MeshStandardMaterial({ color: 0xffe08a, emissive: 0xfbbf24, emissiveIntensity: 0.55 });
  private readonly board = new THREE.MeshStandardMaterial({ color: 0x3f2a14, roughness: 0.8 });

  constructor(parent: THREE.Object3D) {
    parent.add(this.group);
    this.group.add(this.waypoints);
    this.build();
  }

  setFollow(on: boolean, playerZ: number): void {
    this.waypoints.visible = on;
    if (!on) return;
    const kids = this.waypoints.children;
    for (const k of kids) {
      const z = k.position.z;
      k.visible = z > playerZ - 4 && z < playerZ + 48;
    }
  }

  private build(): void {
    this.addSideAndSecret();
    this.addLamps();
    this.addAnchors();
    this.addSigns();
    this.addBridgesAndGates();
    this.addBlockedCues();
    this.addDeadEnds();
    this.addDistantLandmarks();
    this.addFollowWaypoints();
  }

  private addSideAndSecret(): void {
    const bands = [
      { min: 8, max: 22, side: 8.6 },
      { min: 22, max: 50, side: 8.8 },
      { min: 50, max: 72, side: 7.8 },
      { min: 72, max: 94, side: 8.2 },
      { min: 94, max: 114, side: 8.4 },
      { min: 114, max: 130, side: 7.6 },
      { min: 130, max: 145, side: 8.0 },
      { min: 145, max: 158, side: 7.4 },
    ];
    for (const b of bands) {
      const depth = b.max - b.min;
      const mid = (b.min + b.max) / 2;
      const side = new THREE.Mesh(new THREE.PlaneGeometry(1.55, depth - 3), this.dirtSide);
      side.rotation.x = -Math.PI / 2;
      side.position.set(b.side, 0.028, mid);
      side.receiveShadow = true;
      const side2 = side.clone();
      side2.position.x = -b.side;
      this.group.add(side, side2);
      const secret = new THREE.Mesh(new THREE.PlaneGeometry(0.72, Math.max(6, depth * 0.45)), this.dirtSecret);
      secret.rotation.x = -Math.PI / 2;
      secret.position.set(b.side + 3.4, 0.026, mid + 2.4);
      secret.receiveShadow = true;
      this.group.add(secret);
    }
  }

  private addLamps(): void {
    for (let z = 8; z <= 168; z += 14) {
      for (const x of [-2.45, 2.45]) {
        const pole = new THREE.Mesh(new THREE.CylinderGeometry(0.05, 0.07, 1.85, 6), this.wood);
        pole.position.set(x, 0.92, z);
        const glow = new THREE.Mesh(new THREE.SphereGeometry(0.13, 8, 8), this.lampGlow);
        glow.position.set(x, 1.88, z);
        this.group.add(pole, glow);
      }
    }
  }

  private addAnchors(): void {
    for (let z = 12; z <= 164; z += 18) {
      const kind = Math.floor(z / 18) % 4;
      if (kind === 0) {
        const rock = new THREE.Mesh(new THREE.DodecahedronGeometry(0.42, 0), this.stone);
        rock.position.set(-3.1, 0.28, z);
        rock.castShadow = true;
        this.group.add(rock);
      } else if (kind === 1) {
        const trunk = new THREE.Mesh(new THREE.CylinderGeometry(0.16, 0.22, 1.6, 6), this.wood);
        trunk.position.set(3.4, 0.8, z);
        const crown = new THREE.Mesh(new THREE.SphereGeometry(0.7, 8, 8), new THREE.MeshStandardMaterial({ color: 0x2f7a3a, roughness: 0.8 }));
        crown.position.set(3.4, 1.7, z);
        this.group.add(trunk, crown);
      } else if (kind === 2) {
        const post = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.9, 0.12), this.wood);
        post.position.set(-3.3, 0.45, z);
        const rail = new THREE.Mesh(new THREE.BoxGeometry(1.6, 0.06, 0.08), this.wood);
        rail.position.set(-2.6, 0.7, z);
        this.group.add(post, rail);
      } else {
        const step = new THREE.Mesh(new THREE.BoxGeometry(0.7, 0.12, 0.55), this.stone);
        step.position.set(2.9, 0.08, z);
        this.group.add(step);
      }
    }
  }

  private addSigns(): void {
    for (const s of SIGNS) {
      const post = new THREE.Mesh(new THREE.CylinderGeometry(0.06, 0.08, 1.7, 6), this.wood);
      post.position.set(3.15, 0.85, s.z);
      const plate = new THREE.Mesh(new THREE.BoxGeometry(1.35, 0.55, 0.06), this.board);
      plate.position.set(3.15, 1.55, s.z);
      const canvas = document.createElement("canvas");
      canvas.width = 256;
      canvas.height = 128;
      const ctx = canvas.getContext("2d");
      if (ctx) {
        ctx.fillStyle = "#3f2a14";
        ctx.fillRect(0, 0, 256, 128);
        ctx.fillStyle = "#f3e6c8";
        ctx.font = "bold 28px sans-serif";
        ctx.textAlign = "center";
        ctx.fillText(s.left, 128, 48);
        ctx.fillText(s.right, 128, 92);
      }
      const tex = new THREE.CanvasTexture(canvas);
      const face = new THREE.Mesh(
        new THREE.PlaneGeometry(1.28, 0.5),
        new THREE.MeshBasicMaterial({ map: tex, transparent: true }),
      );
      face.position.set(3.15, 1.55, s.z + 0.04);
      this.group.add(post, plate, face);
    }
  }

  private addBridgesAndGates(): void {
    const bridge = new THREE.Mesh(new THREE.BoxGeometry(3.2, 0.18, 5.2), this.wood);
    bridge.position.set(0, 0.16, 36);
    const archL = new THREE.Mesh(new THREE.CylinderGeometry(0.28, 0.34, 3.4, 8), this.stone);
    archL.position.set(-1.5, 1.7, 21.6);
    const archR = archL.clone();
    archR.position.x = 1.5;
    const lintel = new THREE.Mesh(new THREE.BoxGeometry(3.4, 0.28, 0.4), this.stone);
    lintel.position.set(0, 3.3, 21.6);
    const stoneL = new THREE.Mesh(new THREE.BoxGeometry(0.7, 3.4, 0.55), this.stone);
    stoneL.position.set(-1.7, 1.7, 56);
    const stoneR = stoneL.clone();
    stoneR.position.x = 1.7;
    const cap = new THREE.Mesh(new THREE.BoxGeometry(4.2, 0.4, 0.6), this.stone);
    cap.position.set(0, 3.5, 56);
    this.group.add(bridge, archL, archR, lintel, stoneL, stoneR, cap);
  }

  private addBlockedCues(): void {
    const rubble = new THREE.MeshStandardMaterial({ color: 0x6a645c, roughness: 1 });
    const spots = [
      { x: 8.6, z: 31.5 },
      { x: -8.4, z: 78 },
      { x: 8.2, z: 118 },
    ];
    for (const s of spots) {
      for (let i = 0; i < 4; i++) {
        const rock = new THREE.Mesh(new THREE.DodecahedronGeometry(0.28 + i * 0.08, 0), rubble);
        rock.position.set(s.x + (i % 2) * 0.5, 0.22, s.z + i * 0.25);
        rock.castShadow = true;
        this.group.add(rock);
      }
      const fence = new THREE.Mesh(new THREE.BoxGeometry(1.8, 0.7, 0.12), this.wood);
      fence.position.set(s.x, 0.4, s.z - 0.6);
      this.group.add(fence);
    }
  }

  private addDeadEnds(): void {
    const chestMat = new THREE.MeshStandardMaterial({ color: 0x8a5a28, roughness: 0.7, metalness: 0.12 });
    const gold = new THREE.MeshStandardMaterial({ color: 0xc9a227, roughness: 0.45, metalness: 0.35 });
    for (const d of DEAD_ENDS) {
      if (d.kind === "chest") {
        const box = new THREE.Mesh(new THREE.BoxGeometry(0.55, 0.32, 0.38), chestMat);
        box.position.set(d.x, 0.22, d.z);
        const lid = new THREE.Mesh(new THREE.BoxGeometry(0.58, 0.08, 0.4), gold);
        lid.position.set(d.x, 0.4, d.z);
        this.group.add(box, lid);
      } else if (d.kind === "lore") {
        const stele = new THREE.Mesh(new THREE.BoxGeometry(0.35, 1.1, 0.12), this.stone);
        stele.position.set(d.x, 0.55, d.z);
        this.group.add(stele);
      } else {
        const ruin = new THREE.Mesh(new THREE.BoxGeometry(1.1, 0.7, 0.7), this.stone);
        ruin.position.set(d.x, 0.35, d.z);
        this.group.add(ruin);
      }
    }
  }

  private addDistantLandmarks(): void {
    const mountain = new THREE.Mesh(
      new THREE.ConeGeometry(7.5, 16, 6),
      new THREE.MeshStandardMaterial({ color: 0x9aa7b4, roughness: 0.92, fog: false }),
    );
    mountain.position.set(-6, 7.2, 48);
    mountain.castShadow = false;
    const peak = mountain.clone();
    peak.scale.set(0.7, 0.85, 0.7);
    peak.position.set(4, 8.4, 122);
    const minaret = new THREE.Mesh(
      new THREE.CylinderGeometry(0.55, 0.7, 18, 10),
      new THREE.MeshStandardMaterial({ color: 0xf8f1d4, emissive: 0xffe08a, emissiveIntensity: 0.16, roughness: 0.4, fog: false }),
    );
    minaret.position.set(0, 9, 166);
    const dome = new THREE.Mesh(
      new THREE.SphereGeometry(4.4, 16, 12, 0, Math.PI * 2, 0, Math.PI / 2),
      new THREE.MeshStandardMaterial({ color: 0xfff3c4, emissive: 0xffd36a, emissiveIntensity: 0.14, roughness: 0.35, fog: false }),
    );
    dome.position.set(0, 7.2, 166);
    this.group.add(mountain, peak, minaret, dome);
  }

  private addFollowWaypoints(): void {
    const mat = new THREE.MeshStandardMaterial({ color: 0xe8d5a3, emissive: 0xc9a227, emissiveIntensity: 0.18, roughness: 0.55, transparent: true, opacity: 0.55 });
    for (let z = 10; z <= 166; z += 16) {
      const pebble = new THREE.Mesh(new THREE.SphereGeometry(0.09, 6, 6), mat);
      pebble.position.set(0, 0.08, z);
      this.waypoints.add(pebble);
    }
    this.waypoints.visible = false;
  }
}

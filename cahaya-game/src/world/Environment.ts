import * as THREE from "three";
import { WORLD } from "../game/GameConfig";
import { WorldTimeSystem } from "./WorldTimeSystem";

export class Environment {
  readonly lights: THREE.Light[] = [];
  readonly props: THREE.Object3D[] = [];
  readonly time = new WorldTimeSystem();
  readonly sun: THREE.DirectionalLight;
  readonly hemi: THREE.HemisphereLight;
  readonly arena: THREE.Group;
  readonly valley: THREE.Group;
  private instanceMode: "hub" | "dungeon" | "arena" | "bg" = "hub";
  private readonly shared = {
    plaster: new THREE.MeshStandardMaterial({ color: 0xf3e6c8, roughness: 0.86 }),
    roof: new THREE.MeshStandardMaterial({ color: 0xc45c3e, roughness: 0.72 }),
    wood: new THREE.MeshStandardMaterial({ color: 0x7a4e2d, roughness: 0.9 }),
    stone: new THREE.MeshStandardMaterial({ color: 0x8a8f86, roughness: 0.95 }),
    leaf: new THREE.MeshStandardMaterial({ color: 0x2f7a3a, roughness: 0.75 }),
    trunk: new THREE.MeshStandardMaterial({ color: 0x6b4423, roughness: 0.9 }),
    grass: new THREE.MeshStandardMaterial({ color: 0x5f9a45, roughness: 1 }),
  };

  constructor(
    private readonly scene: THREE.Scene,
    private readonly hub: THREE.Group,
    private readonly dungeon: THREE.Group,
  ) {
    this.hemi = new THREE.HemisphereLight(0xe8f4ff, 0x6b8a4a, 0.95);
    this.sun = new THREE.DirectionalLight(0xfff1d0, 1.28);
    this.arena = new THREE.Group();
    this.valley = new THREE.Group();
    this.arena.visible = false;
    this.valley.visible = false;
    scene.add(this.arena, this.valley);
  }

  build(): void {
    this.scene.background = new THREE.Color(WORLD.skyTop);
    this.scene.fog = new THREE.Fog(WORLD.fogColor, 36, 155);
    this.sun.position.set(18, 28, 12);
    this.sun.castShadow = true;
    this.sun.shadow.mapSize.set(1024, 1024);
    this.sun.shadow.camera.near = 2;
    this.sun.shadow.camera.far = 80;
    this.sun.shadow.camera.left = -32;
    this.sun.shadow.camera.right = 32;
    this.sun.shadow.camera.top = 32;
    this.sun.shadow.camera.bottom = -32;
    const rim = new THREE.DirectionalLight(0x9ec9ff, 0.45);
    rim.position.set(-12, 8, -10);
    const fill = new THREE.PointLight(0xffe0b8, 0.35, 28);
    fill.position.set(0, 4.2, 6);
    this.scene.add(this.hemi, this.sun, rim, fill);
    this.lights.push(this.hemi, this.sun, rim, fill);

    const ground = new THREE.Mesh(
      new THREE.CircleGeometry(WORLD.groundSize / 2, 48),
      new THREE.MeshStandardMaterial({ color: 0x6f9b4e, roughness: 0.92, metalness: 0.02 }),
    );
    ground.rotation.x = -Math.PI / 2;
    ground.receiveShadow = true;
    this.hub.add(ground);
    this.props.push(ground);

    this.addPath();
    this.addSquare();
    this.addHouses();
    this.addMarket();
    this.addDawnCity();
    this.addLearningCorner();
    this.addFence();
    this.addWaterBridge();
    this.addLamps();
    this.addBenches();
    this.addGrass();
    this.scatterTrees();
    this.scatterRocks();
    this.addForestEntrance();
    this.buildDungeon();
    this.buildArena();
    this.buildValley();
  }

  setDungeonMode(on: boolean): void {
    this.setInstanceMode(on ? "dungeon" : "hub");
  }

  setInstanceMode(mode: "hub" | "dungeon" | "arena" | "bg"): void {
    this.instanceMode = mode;
    this.hub.visible = mode === "hub";
    this.dungeon.visible = mode === "dungeon";
    this.arena.visible = mode === "arena";
    this.valley.visible = mode === "bg";
    if (mode === "arena") {
      this.scene.background = new THREE.Color(0x1a2240);
      this.scene.fog = new THREE.Fog(0x243056, 18, 70);
    } else if (mode === "bg") {
      this.scene.background = new THREE.Color(0x87b4d6);
      this.scene.fog = new THREE.Fog(0x9ec4a8, 28, 110);
    } else {
      this.scene.background = new THREE.Color(WORLD.skyTop);
      this.scene.fog = new THREE.Fog(WORLD.fogColor, 36, 155);
    }
  }

  isHubMode(): boolean {
    return this.instanceMode === "hub";
  }

  setPhase(phase: string): void {
    if (phase === "DAY" || phase === "EVENING" || phase === "NIGHT") {
      this.time.phase = phase;
      this.time.apply(this.scene, this.sun, this.hemi);
    }
  }

  private addPath(): void {
    const path = new THREE.Mesh(
      new THREE.PlaneGeometry(4.4, 46),
      new THREE.MeshStandardMaterial({ color: 0xcbb58a, roughness: 1 }),
    );
    path.rotation.x = -Math.PI / 2;
    path.position.set(0, 0.02, 8);
    path.receiveShadow = true;
    this.hub.add(path);
    const cross = new THREE.Mesh(
      new THREE.PlaneGeometry(22, 3.4),
      new THREE.MeshStandardMaterial({ color: 0xd2bc93, roughness: 1 }),
    );
    cross.rotation.x = -Math.PI / 2;
    cross.position.set(0, 0.025, 4.2);
    this.hub.add(cross);
  }

  private addSquare(): void {
    const plaza = new THREE.Mesh(
      new THREE.CircleGeometry(5.2, 28),
      new THREE.MeshStandardMaterial({ color: 0xd8c49a, roughness: 0.95 }),
    );
    plaza.rotation.x = -Math.PI / 2;
    plaza.position.set(0, 0.03, 4.4);
    plaza.receiveShadow = true;
    this.hub.add(plaza);
    const well = new THREE.Mesh(
      new THREE.CylinderGeometry(0.7, 0.82, 0.5, 12),
      this.shared.stone,
    );
    well.position.set(-1.1, 0.28, 4.6);
    well.castShadow = true;
    this.hub.add(well);
  }

  private addHouses(): void {
    const spots: Array<[number, number, number]> = [
      [-11.5, 8.5, 0.4],
      [-13.2, 5.2, -0.3],
      [-10.4, 11.4, 0.8],
    ];
    for (const [x, z, rot] of spots) this.hub.add(this.house(x, z, rot));
  }

  private house(x: number, z: number, rot: number): THREE.Group {
    const g = new THREE.Group();
    g.position.set(x, 0, z);
    g.rotation.y = rot;
    const body = new THREE.Mesh(new THREE.BoxGeometry(3.2, 2.2, 2.6), this.shared.plaster);
    body.position.y = 1.1;
    body.castShadow = true;
    const roof = new THREE.Mesh(new THREE.ConeGeometry(2.4, 1.3, 4), this.shared.roof);
    roof.position.y = 2.7;
    roof.rotation.y = Math.PI / 4;
    roof.castShadow = true;
    g.add(body, roof);
    this.props.push(g);
    return g;
  }

  private addMarket(): void {
    for (let i = 0; i < 3; i++) {
      const stall = new THREE.Group();
      stall.position.set(9.2 + (i % 2) * 1.4, 0, 1.8 + i * 1.8);
      const cloth = new THREE.Mesh(
        new THREE.BoxGeometry(1.8, 0.08, 1.4),
        new THREE.MeshStandardMaterial({ color: i === 1 ? 0x38bdf8 : 0xf472b6, roughness: 0.7 }),
      );
      cloth.position.y = 1.55;
      const pole = new THREE.Mesh(new THREE.CylinderGeometry(0.05, 0.05, 1.6, 6), this.shared.wood);
      pole.position.set(-0.7, 0.8, -0.5);
      const pole2 = pole.clone();
      pole2.position.x = 0.7;
      stall.add(cloth, pole, pole2);
      this.hub.add(stall);
    }
  }

  private addDawnCity(): void {
    const hall = this.house(-6.6, 0.6, 0.2);
    hall.scale.set(1.15, 1.2, 1.1);
    const inn = this.house(7.4, 8.2, -0.5);
    const emblem = new THREE.Mesh(new THREE.TorusGeometry(0.45, 0.08, 8, 16), new THREE.MeshStandardMaterial({ color: 0xf4d27a, roughness: 0.4 }));
    emblem.position.set(-6.6, 2.8, 0.6);
    emblem.rotation.x = Math.PI / 2;
    this.hub.add(emblem);
    const stall = new THREE.Mesh(new THREE.BoxGeometry(2.2, 0.7, 1.4), this.shared.wood);
    stall.position.set(11.2, 0.35, 4.8);
    stall.castShadow = true;
    this.hub.add(hall, stall, inn);
    const gate = new THREE.Group();
    const postL = new THREE.Mesh(new THREE.CylinderGeometry(0.28, 0.34, 4.2, 8), this.shared.stone);
    postL.position.set(-1.6, 2.1, -8);
    const postR = postL.clone();
    postR.position.x = 1.6;
    const arch = new THREE.Mesh(new THREE.TorusGeometry(1.7, 0.22, 8, 16, Math.PI), this.shared.stone);
    arch.position.set(0, 3.4, -8);
    arch.rotation.z = Math.PI;
    const banner = new THREE.Mesh(new THREE.PlaneGeometry(1.4, 0.7), new THREE.MeshStandardMaterial({ color: 0x1aa6a0, roughness: 0.6, side: THREE.DoubleSide }));
    banner.position.set(0, 2.6, -8.05);
    gate.add(postL, postR, arch, banner);
    this.hub.add(gate);
    this.props.push(emblem, stall, gate);
  }

  private addLearningCorner(): void {
    const desk = new THREE.Mesh(new THREE.BoxGeometry(1.6, 0.45, 0.8), this.shared.wood);
    desk.position.set(-9.2, 0.35, 1.1);
    desk.castShadow = true;
    const shelf = new THREE.Mesh(new THREE.BoxGeometry(1.4, 1.6, 0.28), this.shared.wood);
    shelf.position.set(-10.6, 0.9, 2.6);
    const board = new THREE.Mesh(
      new THREE.BoxGeometry(1.8, 1.1, 0.08),
      new THREE.MeshStandardMaterial({ color: 0x365314, roughness: 0.8 }),
    );
    board.position.set(-8.1, 1.3, 0.4);
    this.hub.add(desk, shelf, board);
  }

  private addFence(): void {
    for (let i = -8; i <= 8; i++) {
      if (Math.abs(i) < 2) continue;
      const post = new THREE.Mesh(new THREE.BoxGeometry(0.12, 0.9, 0.12), this.shared.wood);
      post.position.set(i * 0.7, 0.45, 20.4);
      this.hub.add(post);
    }
  }

  private addWaterBridge(): void {
    const water = new THREE.Mesh(
      new THREE.PlaneGeometry(14, 3.2),
      new THREE.MeshStandardMaterial({ color: 0x4aa3c7, roughness: 0.2, metalness: 0.15, transparent: true, opacity: 0.85 }),
    );
    water.rotation.x = -Math.PI / 2;
    water.position.set(-12, 0.04, 14.5);
    const bridge = new THREE.Mesh(new THREE.BoxGeometry(2.4, 0.16, 4.2), this.shared.wood);
    bridge.position.set(-12, 0.22, 14.5);
    this.hub.add(water, bridge);
  }

  private addLamps(): void {
    const lampMat = new THREE.MeshStandardMaterial({ color: 0xffe08a, emissive: 0xfbbf24, emissiveIntensity: 0.8 });
    const zs = [2, 8, 14, 19];
    for (const z of zs) {
      for (const x of [-2.6, 2.6]) {
        const pole = new THREE.Mesh(new THREE.CylinderGeometry(0.06, 0.08, 2.1, 6), this.shared.wood);
        pole.position.set(x, 1.05, z);
        const glow = new THREE.Mesh(new THREE.SphereGeometry(0.16, 8, 8), lampMat);
        glow.position.set(x, 2.15, z);
        this.hub.add(pole, glow);
      }
    }
  }

  private addBenches(): void {
    for (const [x, z] of [[-3.4, 6.2], [3.6, 5.4]] as Array<[number, number]>) {
      const seat = new THREE.Mesh(new THREE.BoxGeometry(1.4, 0.12, 0.4), this.shared.wood);
      seat.position.set(x, 0.42, z);
      this.hub.add(seat);
    }
  }

  private addGrass(): void {
    for (let i = 0; i < 40; i++) {
      const tuft = new THREE.Mesh(new THREE.ConeGeometry(0.08, 0.28, 4), this.shared.grass);
      const a = i * 0.7;
      tuft.position.set(Math.cos(a) * (4 + (i % 6)), 0.12, 6 + Math.sin(a) * 5);
      this.hub.add(tuft);
    }
  }

  private scatterTrees(): void {
    for (let i = 0; i < 22; i++) {
      const group = new THREE.Group();
      const ang = (i / 22) * Math.PI * 2;
      const dist = 16 + (i % 5) * 2.2;
      group.position.set(Math.cos(ang) * dist, 0, Math.sin(ang) * dist * 0.55 + 2);
      const trunk = new THREE.Mesh(new THREE.CylinderGeometry(0.16, 0.24, 1.7, 8), this.shared.trunk);
      trunk.position.y = 0.85;
      trunk.castShadow = true;
      const c1 = new THREE.Mesh(new THREE.SphereGeometry(0.85, 8, 8), this.shared.leaf);
      c1.position.y = 2.05;
      const c2 = c1.clone();
      c2.scale.setScalar(0.78);
      c2.position.set(0.35, 2.35, 0.1);
      const c3 = c1.clone();
      c3.scale.setScalar(0.7);
      c3.position.set(-0.3, 2.4, -0.15);
      group.add(trunk, c1, c2, c3);
      this.hub.add(group);
      this.props.push(group);
    }
  }

  private scatterRocks(): void {
    for (let i = 0; i < 10; i++) {
      const rock = new THREE.Mesh(new THREE.DodecahedronGeometry(0.32 + (i % 3) * 0.12, 0), this.shared.stone);
      rock.position.set(-7 + i * 1.6, 0.22, 12.5 + (i % 2));
      rock.castShadow = true;
      this.hub.add(rock);
      this.props.push(rock);
    }
  }

  private addForestEntrance(): void {
    const moss = new THREE.MeshStandardMaterial({ color: 0x3f6b32, roughness: 0.85 });
    for (let i = 0; i < 10; i++) {
      const tree = new THREE.Group();
      tree.position.set((i - 4.5) * 2.1, 0, 27 + (i % 3) * 1.8);
      const trunk = new THREE.Mesh(new THREE.CylinderGeometry(0.22, 0.3, 2.2, 6), this.shared.trunk);
      trunk.position.y = 1.1;
      const crown = new THREE.Mesh(new THREE.ConeGeometry(1.5, 3.2, 7), moss);
      crown.position.y = 3.1;
      tree.add(trunk, crown);
      this.hub.add(tree);
      this.props.push(tree);
    }
    const mist = new THREE.Mesh(
      new THREE.PlaneGeometry(18, 8),
      new THREE.MeshStandardMaterial({ color: 0x9bc4a3, transparent: true, opacity: 0.18, depthWrite: false }),
    );
    mist.rotation.x = -Math.PI / 2;
    mist.position.set(0, 0.06, 31);
    this.hub.add(mist);
  }

  private buildDungeon(): void {
    this.dungeon.visible = false;
    const floor = new THREE.Mesh(
      new THREE.CircleGeometry(30, 40),
      new THREE.MeshStandardMaterial({ color: 0x3a4f38, roughness: 0.95 }),
    );
    floor.rotation.x = -Math.PI / 2;
    floor.receiveShadow = true;
    this.dungeon.add(floor);
    this.props.push(floor);
    const moss = new THREE.MeshStandardMaterial({ color: 0x1f3d28, roughness: 0.82 });
    for (let i = 0; i < 16; i++) {
      const tree = new THREE.Group();
      const ang = (i / 16) * Math.PI * 2;
      tree.position.set(Math.cos(ang) * 18, 0, 10 + Math.sin(ang) * 10);
      const trunk = new THREE.Mesh(new THREE.CylinderGeometry(0.28, 0.38, 2.6, 6), this.shared.trunk);
      trunk.position.y = 1.3;
      const crown = new THREE.Mesh(new THREE.ConeGeometry(1.8, 3.6, 7), moss);
      crown.position.y = 3.4;
      tree.add(trunk, crown);
      this.dungeon.add(tree);
      this.props.push(tree);
    }
    const mist = new THREE.Mesh(
      new THREE.CircleGeometry(8, 24),
      new THREE.MeshStandardMaterial({ color: 0xb7cfc0, transparent: true, opacity: 0.22, depthWrite: false }),
    );
    mist.rotation.x = -Math.PI / 2;
    mist.position.set(0, 0.05, 16);
    const arena = new THREE.Mesh(
      new THREE.CircleGeometry(7, 28),
      new THREE.MeshStandardMaterial({ color: 0x2c3328, roughness: 1 }),
    );
    arena.rotation.x = -Math.PI / 2;
    arena.position.set(0, 0.04, 20);
    this.dungeon.add(mist, arena);
  }

  private buildArena(): void {
    this.arena.visible = false;
    const floor = new THREE.Mesh(
      new THREE.CircleGeometry(22, 48),
      new THREE.MeshStandardMaterial({ color: 0x3d4a6d, roughness: 0.55, metalness: 0.18 }),
    );
    floor.rotation.x = -Math.PI / 2;
    floor.receiveShadow = true;
    this.arena.add(floor);
    for (let i = 0; i < 8; i++) {
      const ang = (i / 8) * Math.PI * 2;
      const plat = new THREE.Mesh(
        new THREE.CylinderGeometry(2.2, 2.4, 0.35, 10),
        new THREE.MeshStandardMaterial({ color: 0x6b7bb5, roughness: 0.4, emissive: 0x1c2a55, emissiveIntensity: 0.35 }),
      );
      plat.position.set(Math.cos(ang) * 10, 1.1, Math.sin(ang) * 10);
      this.arena.add(plat);
      const pillar = new THREE.Mesh(
        new THREE.CylinderGeometry(0.28, 0.34, 4.2, 8),
        new THREE.MeshStandardMaterial({ color: 0xc9e4ff, emissive: 0x4cc9f0, emissiveIntensity: 0.55, roughness: 0.25 }),
      );
      pillar.position.set(Math.cos(ang) * 16, 2.1, Math.sin(ang) * 16);
      this.arena.add(pillar);
    }
    const ring = new THREE.Mesh(
      new THREE.TorusGeometry(20.5, 0.18, 8, 48),
      new THREE.MeshStandardMaterial({ color: 0x9bd4ff, emissive: 0x4ea8de, emissiveIntensity: 0.4 }),
    );
    ring.rotation.x = Math.PI / 2;
    ring.position.y = 0.4;
    this.arena.add(ring);
    this.props.push(floor);
  }

  private buildValley(): void {
    this.valley.visible = false;
    const ground = new THREE.Mesh(
      new THREE.CircleGeometry(40, 48),
      new THREE.MeshStandardMaterial({ color: 0x5d8a46, roughness: 0.95 }),
    );
    ground.rotation.x = -Math.PI / 2;
    ground.receiveShadow = true;
    this.valley.add(ground);
    const ruin = new THREE.MeshStandardMaterial({ color: 0x8b8378, roughness: 0.92 });
    for (let i = 0; i < 6; i++) {
      const block = new THREE.Mesh(new THREE.BoxGeometry(1.6 + (i % 2), 1.2, 1.4), ruin);
      block.position.set(-14 + i * 5.5, 0.6, 14 - (i % 3) * 4);
      this.valley.add(block);
    }
    const bridge = new THREE.Mesh(new THREE.BoxGeometry(18, 0.28, 2.4), new THREE.MeshStandardMaterial({ color: 0x6b5344, roughness: 0.88 }));
    bridge.position.set(0, 0.2, 8);
    this.valley.add(bridge);
    const shrineMat = new THREE.MeshStandardMaterial({ color: 0xddefff, emissive: 0x67e8f9, emissiveIntensity: 0.45 });
    for (const x of [-12, 0, 12]) {
      const shrine = new THREE.Mesh(new THREE.CylinderGeometry(1.6, 1.8, 0.5, 12), shrineMat);
      shrine.position.set(x, 0.28, 0);
      this.valley.add(shrine);
    }
    const cliff = new THREE.Mesh(new THREE.ConeGeometry(6, 14, 6), this.shared.stone);
    cliff.position.set(18, 6.5, -8);
    const fall = new THREE.Mesh(
      new THREE.PlaneGeometry(1.2, 8),
      new THREE.MeshStandardMaterial({ color: 0xa8e4ff, transparent: true, opacity: 0.55, roughness: 0.15, metalness: 0.2, side: THREE.DoubleSide }),
    );
    fall.position.set(16.4, 4, -6);
    this.valley.add(cliff, fall);
    this.props.push(ground);
  }
}

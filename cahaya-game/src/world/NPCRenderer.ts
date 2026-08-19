import * as THREE from "three";
import type { NPCSnapshot } from "../network/NetworkMessage";
import { clothMat, hash32, leatherMat, metalMat, skinMat, std } from "../art/PBR";
import {
  NPC_AVOIDANCE_RADIUS,
  npcBodyRadius,
  resolveWorldXZ,
  softSeparationXZ,
} from "./WorldColliders";

const SKIN = [0xe4b48a, 0xd9a07a, 0xc68662, 0x8d5524, 0xf1c27d];

export class NPCRenderer {
  readonly group = new THREE.Group();
  private readonly nodes = new Map<string, NPCNode>();
  talkingId = "";

  constructor(scene: THREE.Scene) {
    scene.add(this.group);
  }

  sync(list: NPCSnapshot[], markerFor: (id: string) => string): void {
    const seen = new Set<string>();
    for (const snap of list) {
      seen.add(snap.id);
      let node = this.nodes.get(snap.id);
      if (!node) {
        node = new NPCNode(snap);
        this.nodes.set(snap.id, node);
        this.group.add(node.root);
      }
      node.apply(snap, markerFor(snap.id));
    }
    for (const [id, node] of this.nodes) {
      if (seen.has(id)) continue;
      node.root.removeFromParent();
      this.nodes.delete(id);
    }
  }

  nearest(x: number, z: number, maxDist: number): NPCSnapshot | null {
    let best: NPCSnapshot | null = null;
    let bestD = maxDist;
    for (const node of this.nodes.values()) {
      const d = Math.hypot(node.root.position.x - x, node.root.position.z - z);
      if (d < bestD) {
        bestD = d;
        best = { ...node.snap, x: node.root.position.x, z: node.root.position.z };
      }
    }
    return best;
  }

  update(dt: number, camera: THREE.Camera, playerX = 0, playerZ = 0): void {
    const others = [...this.nodes.values()];
    for (const node of others) node.update(dt, camera, playerX, playerZ, this.talkingId === node.snap.id, others);
  }

  positions(): Array<{ x: number; z: number; type: string }> {
    return [...this.nodes.values()].map((n) => ({ x: n.root.position.x, z: n.root.position.z, type: n.snap.type }));
  }

  all(): NPCSnapshot[] {
    return [...this.nodes.values()].map((n) => n.snap);
  }
}

class NPCNode {
  readonly root = new THREE.Group();
  snap: NPCSnapshot;
  private readonly body: THREE.Group;
  private readonly hips: THREE.Group;
  private readonly head: THREE.Group;
  private readonly armL: THREE.Group;
  private readonly armR: THREE.Group;
  private readonly legL: THREE.Group;
  private readonly legR: THREE.Group;
  private readonly plate: THREE.Sprite;
  private readonly plateTex: THREE.CanvasTexture;
  private readonly plateCanvas: HTMLCanvasElement;
  private readonly plateMat: THREE.SpriteMaterial;
  private marker = "";
  private x: number;
  private z: number;
  private yaw: number;
  private speed = 0;
  private tx: number;
  private tz: number;
  private tyaw: number;
  private readonly seed: number;
  private armLX = 0;
  private armRX = 0;
  private legLX = 0;
  private legRX = 0;

  constructor(snap: NPCSnapshot) {
    this.snap = snap;
    this.seed = hash32(snap.id);
    this.x = snap.x;
    this.z = snap.z;
    this.yaw = snap.yaw;
    this.tx = snap.x;
    this.tz = snap.z;
    this.tyaw = snap.yaw;
    const board = snap.type === "QUEST_BOARD";
    const rig = board ? emptyRig(makeBoard()) : makeNpcBody(snap);
    this.body = rig.root;
    this.hips = rig.hips;
    this.head = rig.head;
    this.armL = rig.armL;
    this.armR = rig.armR;
    this.legL = rig.legL;
    this.legR = rig.legR;
    this.root.add(this.body);
    const plate = makeNpcPlate(snap.name, snap.role, "", snap.type);
    this.plate = plate.sprite;
    this.plateTex = plate.tex;
    this.plateCanvas = plate.canvas;
    this.plateMat = plate.mat;
    this.plate.position.y = snap.type === "CHILD" ? 1.42 : 1.92;
    this.root.add(this.plate);
    this.root.position.set(snap.x, 0, snap.z);
    this.body.rotation.y = snap.yaw;
  }

  apply(snap: NPCSnapshot, marker: string): void {
    this.snap = snap;
    const jump = Math.hypot(snap.x - this.x, snap.z - this.z);
    if (jump > 6) {
      this.x = snap.x;
      this.z = snap.z;
      this.yaw = snap.yaw;
    }
    this.tx = snap.x;
    this.tz = snap.z;
    this.tyaw = snap.yaw;
    if (marker !== this.marker) {
      this.marker = marker;
      drawNpcPlate(this.plateCanvas, snap.name, snap.role, marker, snap.type);
      this.plateTex.needsUpdate = true;
    }
  }

  update(dt: number, camera: THREE.Camera, playerX: number, playerZ: number, talking: boolean, othersList: NPCNode[]): void {
    const board = this.snap.type === "QUEST_BOARD";
    if (!talking && !board) {
      const rad = npcBodyRadius(this.snap.type);
      const others = othersList
        .filter((o) => o !== this)
        .map((o) => ({ x: o.x, z: o.z, r: npcBodyRadius(o.snap.type) }));
      const sep = softSeparationXZ(this.x, this.z, rad, others);
      this.x += sep.x * dt * 1.8;
      this.z += sep.z * dt * 1.8;
      const pdx = this.x - playerX;
      const pdz = this.z - playerZ;
      const pd = Math.hypot(pdx, pdz);
      if (pd < NPC_AVOIDANCE_RADIUS && pd > 0.001) {
        const push = ((NPC_AVOIDANCE_RADIUS - pd) / pd) * 0.35 * dt;
        this.x += pdx * push;
        this.z += pdz * push;
      }
    }

    const wantX = talking ? this.x : this.tx;
    const wantZ = talking ? this.z : this.tz;
    const toX = wantX - this.x;
    const toZ = wantZ - this.z;
    const distMove = Math.hypot(toX, toZ);
    const face = talking ? Math.atan2(playerX - this.x, playerZ - this.z) : distMove > 0.05 ? Math.atan2(toX, toZ) : this.tyaw;
    this.yaw = turnToward(this.yaw, face, 8 * dt);

    let wantSpeed = 0;
    if (!talking && !board && distMove > 0.06) {
      const aligned = Math.abs(angleDiff(this.yaw, face)) < 0.55;
      wantSpeed = aligned ? Math.min(1.45, distMove * 2.4) : 0.12;
      if (distMove < 0.85) wantSpeed *= Math.max(0.08, distMove / 0.85);
    }
    const accel = wantSpeed >= this.speed ? 3.6 : 6.2;
    if (this.speed < wantSpeed) this.speed = Math.min(wantSpeed, this.speed + accel * dt);
    else this.speed = Math.max(wantSpeed, this.speed - accel * dt);
    if (this.speed < 0.03) this.speed = 0;
    if (this.speed > 0 && !board) {
      const step = Math.min(this.speed * dt, distMove);
      if (distMove > 0.0001) {
        this.x += (toX / distMove) * step;
        this.z += (toZ / distMove) * step;
      }
    }

    const bodyR = npcBodyRadius(this.snap.type);
    const resolved = resolveWorldXZ(this.x, this.z, bodyR);
    this.x = resolved.x;
    this.z = resolved.z;

    this.root.position.set(this.x, 0, this.z);
    this.body.rotation.y = this.yaw;
    this.body.position.y = 0;

    const dist = camera.position.distanceTo(this.root.position);
    const quest = this.marker === "!" || this.marker === "?";
    const maxD = quest ? 35 : 20;
    const show = dist < maxD;
    this.plate.visible = show;
    if (show) {
      const w = THREE.MathUtils.clamp(dist * 0.042, 0.55, 0.88);
      this.plate.scale.set(w, w * 0.28, 1);
      let op = 1;
      if (dist > 10) op = 1 - (dist - 10) / (maxD - 10);
      if (quest) op = Math.min(1, op + 0.12);
      this.plateMat.opacity = Math.max(0, op);
    }

    if (board || dist > 28) return;
    const k = 1 - Math.exp(-12 * dt);
    const t = performance.now() * 0.001 + (this.seed % 97) * 0.17;
    const moving = this.speed > 0.05;
    const amp = moving ? Math.min(0.55, this.speed * 0.42) : 0;
    const freq = 7.2 + (this.seed % 5) * 0.15;
    const swing = Math.sin(t * freq) * amp;
    let tArmL = swing;
    let tArmR = -swing;
    let tLegL = -swing;
    let tLegR = swing;
    if (!moving) {
      const idle = 0.5 + (this.seed % 3);
      const breath = Math.sin(t * (1.5 + idle * 0.2)) * 0.012;
      this.hips.position.y = (this.snap.type === "CHILD" ? 0.62 : 0.92) + breath;
      this.head.rotation.y = talking ? Math.sin(t * 2.4) * 0.08 : Math.sin(t * 0.45 + idle) * 0.18;
      this.head.rotation.x = talking ? Math.sin(t * 3.1) * 0.04 : Math.sin(t * 0.7) * 0.03;
      tArmL = talking ? -0.35 + Math.sin(t * 3) * 0.12 : 0.08 + Math.sin(t * 1.6) * 0.03;
      tArmR = talking ? -0.15 : -0.08 + Math.sin(t * 1.6 + 1) * 0.03;
      tLegL = 0;
      tLegR = 0;
    } else {
      this.hips.position.y = this.snap.type === "CHILD" ? 0.62 : 0.92;
      this.head.rotation.y += (0 - this.head.rotation.y) * k;
      this.head.rotation.x += (0 - this.head.rotation.x) * k;
    }
    this.armLX += (tArmL - this.armLX) * k;
    this.armRX += (tArmR - this.armRX) * k;
    this.legLX += (tLegL - this.legLX) * k;
    this.legRX += (tLegR - this.legRX) * k;
    this.armL.rotation.x = this.armLX;
    this.armR.rotation.x = this.armRX;
    this.legL.rotation.x = this.legLX;
    this.legR.rotation.x = this.legRX;
  }
}

function turnToward(from: number, to: number, maxStep: number): number {
  const d = THREE.MathUtils.clamp(angleDiff(from, to), -maxStep, maxStep);
  return from + d;
}

function angleDiff(from: number, to: number): number {
  let d = to - from;
  while (d > Math.PI) d -= Math.PI * 2;
  while (d < -Math.PI) d += Math.PI * 2;
  return d;
}

interface NpcRig {
  root: THREE.Group;
  hips: THREE.Group;
  head: THREE.Group;
  armL: THREE.Group;
  armR: THREE.Group;
  legL: THREE.Group;
  legR: THREE.Group;
}

function emptyRig(root: THREE.Group): NpcRig {
  const g = new THREE.Group();
  return { root, hips: g, head: g, armL: g, armR: g, legL: g, legR: g };
}

function makeBoard(): THREE.Group {
  const g = new THREE.Group();
  const board = new THREE.Mesh(new THREE.BoxGeometry(1.15, 1.45, 0.1), leatherMat(0x6b3f1d));
  board.position.y = 1.1;
  const post = new THREE.Mesh(new THREE.CylinderGeometry(0.07, 0.08, 1.1, 8), leatherMat(0x4a2c14));
  post.position.y = 0.55;
  g.add(board, post);
  return g;
}

function makeNpcBody(snap: NPCSnapshot): NpcRig {
  const h = hash32(snap.id + snap.type);
  const skin = skinMat(SKIN[h % SKIN.length]);
  const cloth = clothMat(palette(snap.type), "npc-" + snap.type);
  const hairMat = std(hairColor(h), 0.45, 0.02);
  const child = snap.type === "CHILD";
  const root = new THREE.Group();
  const scale = child ? 0.68 : snap.type === "ELDER" ? 0.9 : snap.type === "GUARD" ? 1.1 : snap.type === "FARMER" ? 1.04 : 1;
  const stocky = snap.type === "BLACKSMITH" || snap.type === "GUARD" ? 1.18 : snap.type === "ELDER" || snap.type === "HEALER" ? 0.92 : 1;
  const hips = new THREE.Group();
  hips.position.y = child ? 0.62 : 0.92;
  const torso = new THREE.Mesh(new THREE.SphereGeometry(0.17, 12, 10), cloth);
  torso.scale.set(1.05 * stocky, child ? 1.05 : 1.28, 0.68);
  torso.position.y = 0.26;
  const head = makeNpcHead(skin, hairMat, snap.type, h, child);
  head.position.y = 0.58;
  const armL = makeLimb(skin, cloth, -1, true);
  const armR = makeLimb(skin, cloth, 1, true);
  hips.add(torso, head, armL, armR);
  const legs = new THREE.Group();
  const legL = makeLimb(skin, cloth, -1, false);
  const legR = makeLimb(skin, cloth, 1, false);
  legs.add(legL, legR);
  root.add(hips, legs);
  addGear(root, snap.type, h);
  root.scale.setScalar(scale);
  root.traverse((o) => {
    if (o instanceof THREE.Mesh) {
      o.castShadow = true;
      o.receiveShadow = true;
    }
  });
  return { root, hips, head, armL, armR, legL, legR };
}

function makeNpcHead(skin: THREE.Material, hair: THREE.Material, type: string, h: number, child: boolean): THREE.Group {
  const g = new THREE.Group();
  const skull = new THREE.Mesh(new THREE.SphereGeometry(child ? 0.155 : 0.138, 14, 12), skin);
  skull.scale.set(0.92, 1.04, 0.95);
  g.add(skull);
  const eye = (side: number) => {
    const e = new THREE.Mesh(new THREE.SphereGeometry(0.018, 8, 6), std(0xf7f4ee, 0.3));
    e.position.set(side * 0.045, 0.02, 0.11);
    const iris = new THREE.Mesh(new THREE.SphereGeometry(0.01, 6, 6), std(0x1f2937, 0.25));
    iris.position.set(side * 0.045, 0.02, 0.122);
    g.add(e, iris);
  };
  eye(-1);
  eye(1);
  const mouth = new THREE.Mesh(new THREE.CapsuleGeometry(0.006, 0.028, 3, 5), std(0x4a3028, 0.6));
  mouth.rotation.z = Math.PI / 2;
  mouth.position.set(0, -0.05, 0.12);
  g.add(mouth);
  g.add(makeNpcHair(hair, type, h, child));
  return g;
}

function makeNpcHair(mat: THREE.Material, type: string, h: number, child: boolean): THREE.Group {
  const g = new THREE.Group();
  const cap = new THREE.Mesh(new THREE.SphereGeometry(child ? 0.16 : 0.145, 10, 8, 0, Math.PI * 2, 0, 1.15), mat);
  cap.position.y = 0.04;
  g.add(cap);
  if (type === "FARMER" || type === "GUARD") return g;
  const style = type === "ELDER" ? 0 : type === "HEALER" ? 2 : h % 5;
  if (style === 0) {
    const bun = new THREE.Mesh(new THREE.SphereGeometry(0.07, 8, 8), mat);
    bun.position.set(0, 0.14, -0.04);
    g.add(bun);
  } else if (style === 2) {
    for (let i = 0; i < 6; i++) {
      const lock = new THREE.Mesh(new THREE.CapsuleGeometry(0.02, 0.28, 3, 5), mat);
      lock.position.set(-0.08 + i * 0.03, -0.08, -0.04);
      lock.rotation.x = 0.5;
      g.add(lock);
    }
  } else if (style === 3) {
    const bang = new THREE.Mesh(new THREE.CapsuleGeometry(0.025, 0.1, 3, 5), mat);
    bang.position.set(0.03, 0.06, 0.1);
    bang.rotation.x = -0.8;
    g.add(bang);
  }
  return g;
}

function makeLimb(skin: THREE.Material, cloth: THREE.Material, side: number, arm: boolean): THREE.Group {
  const g = new THREE.Group();
  const m = new THREE.Mesh(new THREE.CapsuleGeometry(arm ? 0.042 : 0.052, arm ? 0.3 : 0.4, 4, 8), cloth);
  m.position.set(side * (arm ? 0.2 : 0.09), arm ? 0.2 : 0.32, 0);
  m.rotation.z = side * (arm ? 0.14 : 0.04);
  g.add(m);
  if (arm) {
    const hand = new THREE.Mesh(new THREE.SphereGeometry(0.038, 8, 8), skin);
    hand.scale.set(0.85, 1, 0.7);
    hand.position.set(side * 0.24, -0.02, 0);
    g.add(hand);
  }
  return g;
}

function addGear(g: THREE.Group, type: string, h: number): void {
  if (type === "ELDER") {
    const staff = new THREE.Mesh(new THREE.CylinderGeometry(0.025, 0.03, 1.6, 6), leatherMat());
    staff.position.set(0.28, 0.9, 0.1);
    g.add(staff);
  } else if (type === "MERCHANT") {
    const pack = new THREE.Mesh(new THREE.BoxGeometry(0.28, 0.22, 0.18), leatherMat(0x8b5a2b));
    pack.position.set(0, 1.15, -0.2);
    g.add(pack);
  } else if (type === "BLACKSMITH") {
    const apron = new THREE.Mesh(new THREE.PlaneGeometry(0.32, 0.45), leatherMat(0x3f2a14));
    apron.position.set(0, 1.05, 0.16);
    g.add(apron);
  } else if (type === "GUARD") {
    const helm = new THREE.Mesh(new THREE.SphereGeometry(0.17, 10, 8, 0, Math.PI * 2, 0, 1.1), metalMat());
    helm.position.y = 1.62;
    g.add(helm);
  } else if (type === "HEALER") {
    const sash = new THREE.Mesh(new THREE.TorusGeometry(0.16, 0.025, 6, 12), std(0x34d399, 0.5));
    sash.rotation.x = Math.PI / 2;
    sash.position.y = 1.05;
    g.add(sash);
  } else if (type === "TEACHER" || type === "SCHOLAR") {
    const book = new THREE.Mesh(new THREE.BoxGeometry(0.14, 0.04, 0.18), std(0xfde68a, 0.7));
    book.position.set(0.22, 1.05, 0.12);
    g.add(book);
  } else if (type === "FARMER") {
    const hat = new THREE.Mesh(new THREE.CylinderGeometry(0.22, 0.08, 0.08, 10), leatherMat(0x6b5344));
    hat.position.y = 1.72;
    g.add(hat);
  } else if (h % 3 === 0) {
    const scarf = new THREE.Mesh(new THREE.TorusGeometry(0.12, 0.02, 5, 10), clothMat(0x1aa6a0, "scarf"));
    scarf.position.y = 1.42;
    scarf.rotation.x = 1.1;
    g.add(scarf);
  }
}

function palette(type: string): number {
  const map: Record<string, number> = {
    ELDER: 0xc9a227,
    TEACHER: 0x3b6ea8,
    MERCHANT: 0xb45309,
    GUARD: 0x475569,
    HEALER: 0x047857,
    TRAVELER: 0x0e7490,
    GUIDE: 0x1d4ed8,
    TRAINER: 0x9d174d,
    CHILD: 0xdb2777,
    BLACKSMITH: 0x57534e,
    COOK: 0xd97706,
    FISHER: 0x0284c7,
    FARMER: 0x4d7c0f,
    MASTER: 0xa16207,
    GUILD: 0xb45309,
  };
  return map[type] ?? 0x64748b;
}

function hairColor(h: number): number {
  return [0x1a1410, 0x3b2314, 0x6b4423, 0x111118, 0xc9a227][h % 5];
}

function makeNpcPlate(name: string, role: string, marker: string, type: string): { sprite: THREE.Sprite; tex: THREE.CanvasTexture; canvas: HTMLCanvasElement; mat: THREE.SpriteMaterial } {
  const canvas = document.createElement("canvas");
  canvas.width = 256;
  canvas.height = 64;
  drawNpcPlate(canvas, name, role, marker, type);
  const tex = new THREE.CanvasTexture(canvas);
  tex.colorSpace = THREE.SRGBColorSpace;
  const mat = new THREE.SpriteMaterial({ map: tex, transparent: true, depthTest: true, depthWrite: false });
  const sprite = new THREE.Sprite(mat);
  sprite.scale.set(0.72, 0.2, 1);
  sprite.renderOrder = 3;
  return { sprite, tex, canvas, mat };
}

function plateColor(type: string, marker: string): string {
  if (marker === "!" || marker === "?") return "#efe3b0";
  if (type === "MERCHANT") return "#e8d5a3";
  return "#e8eef4";
}

function drawNpcPlate(canvas: HTMLCanvasElement, name: string, _role: string, marker: string, type: string): void {
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = "rgba(8,14,24,0.38)";
  roundRect(ctx, 48, 18, 160, 36, 8);
  ctx.fill();
  ctx.fillStyle = plateColor(type, marker);
  ctx.font = "600 15px Segoe UI, sans-serif";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.shadowColor = "rgba(0,0,0,0.45)";
  ctx.shadowBlur = 3;
  ctx.fillText(name, 128, 32);
  ctx.shadowBlur = 0;
  if (marker) {
    ctx.fillStyle = marker === "!" ? "#e8c547" : "#86c9a0";
    ctx.beginPath();
    ctx.arc(128, 10, 5, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = "#1a1408";
    ctx.font = "600 9px Segoe UI, sans-serif";
    ctx.fillText(marker, 128, 10);
  }
}

function roundRect(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, r: number): void {
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + w, y, x + w, y + h, r);
  ctx.arcTo(x + w, y + h, x, y + h, r);
  ctx.arcTo(x, y + h, x, y, r);
  ctx.arcTo(x, y, x + w, y, r);
  ctx.closePath();
}

import * as THREE from "three";
import { PLAYER } from "../game/GameConfig";
import { HealthComponent } from "./HealthComponent";
import type { Combatant, CombatTeam, Hurtbox } from "./Combatant";
import type { CombatState } from "./CombatState";
import { interpFactor, lerpAngle } from "../network/Interpolation";
import type { EnemySnapshot } from "../network/NetworkMessage";
import { buildEnemyMesh } from "../art/EnemyArt";
import { hash32 } from "../art/PBR";

const KIND_COLOR: Record<string, number> = {
  forest_fang: 0x4a7c3f,
  shadow_imp: 0x4b3a6a,
  stone_beast: 0x8a8174,
  training_dummy: 0xc4a574,
  elite_shadow_beast: 0x3b0764,
  ragha: 0x1e3a2f,
};

export class NetworkEnemy implements Combatant {
  readonly id: string;
  readonly team: CombatTeam = "enemy";
  readonly group = new THREE.Group();
  readonly mesh = new THREE.Group();
  readonly velocity = new THREE.Vector3();
  readonly health: HealthComponent;
  readonly hurtbox: Hurtbox = { radius: 0.48, height: 1.5, yOffset: 0.8 };
  facingYaw = 0;
  invincible = false;
  combatState: CombatState = "IDLE";
  name: string;
  level: number;
  kind: string;
  rank: string;
  private readonly bodyMat: THREE.MeshStandardMaterial;
  private readonly current = new THREE.Vector3();
  private readonly target = new THREE.Vector3();
  private fromYaw = 0;
  private toYaw = 0;
  private elapsed = 1;
  private hitFlash = 0;
  private deadPose = false;
  private readonly label: THREE.Sprite;
  private readonly labelTex: THREE.CanvasTexture;
  private readonly labelCanvas: HTMLCanvasElement;
  private readonly labelMat: THREE.SpriteMaterial;
  namedTarget = false;

  constructor(snap: EnemySnapshot) {
    this.id = snap.id;
    this.kind = snap.kind;
    this.name = snap.name;
    this.level = snap.level;
    this.rank = snap.rank || "normal";
    this.health = new HealthComponent(snap.maxHp, snap.hp);
    this.bodyMat = new THREE.MeshStandardMaterial({
      color: KIND_COLOR[snap.kind] ?? ((hash32(snap.kind) % 0x889988) + 0x223322),
      roughness: 0.72,
      metalness: 0.08,
    });
    this.group.add(this.mesh);
    this.build(snap.kind, snap.rank);
    const plate = makePlate(this.name, this.level, this.health);
    this.label = plate.sprite;
    this.labelTex = plate.tex;
    this.labelCanvas = plate.canvas;
    this.labelMat = plate.mat;
    this.label.position.y = snap.rank === "boss" || snap.rank === "world" || snap.kind === "ragha" ? 2.7 : 1.95;
    this.group.add(this.label);
    this.current.set(snap.x, snap.y, snap.z);
    this.target.copy(this.current);
    this.group.position.copy(this.current);
    this.facingYaw = snap.yaw;
    this.mesh.rotation.y = snap.yaw;
    this.applyHp(snap.hp, snap.maxHp, snap.st === "DEAD");
  }

  get position(): THREE.Vector3 {
    return this.group.position;
  }

  applySnapshot(snap: EnemySnapshot): void {
    this.current.copy(this.group.position);
    this.target.set(snap.x, snap.y, snap.z);
    this.fromYaw = this.mesh.rotation.y;
    this.toYaw = snap.yaw;
    this.elapsed = 0;
    this.name = snap.name;
    this.level = snap.level;
    this.applyHp(snap.hp, snap.maxHp, snap.st === "DEAD");
    if (snap.st === "ATTACK") this.combatState = "ATTACKING";
    else if (snap.st === "HIT") this.combatState = "HIT";
    else if (snap.st !== "DEAD") this.combatState = "IDLE";
  }

  applyHp(hp: number, maxHp: number, dead: boolean): void {
    this.health.maxHp = maxHp;
    this.health.hp = Math.max(0, hp);
    if (dead || hp <= 0) {
      this.combatState = "DEAD";
      this.invincible = true;
      if (!this.deadPose) {
        this.deadPose = true;
        this.mesh.rotation.x = Math.PI / 2;
        this.mesh.position.y = 0.28;
      }
    } else if (this.deadPose) {
      this.deadPose = false;
      this.invincible = false;
      this.mesh.rotation.x = 0;
      this.mesh.position.y = 0;
      this.combatState = "IDLE";
    }
    drawPlate(this.labelCanvas, this.name, this.level, this.health);
    this.labelTex.needsUpdate = true;
  }

  flash(): void {
    this.hitFlash = 0.14;
  }

  update(dt: number, camera: THREE.Camera): void {
    this.elapsed += dt;
    const a = interpFactor(this.elapsed);
    this.group.position.lerpVectors(this.current, this.target, a);
    this.mesh.rotation.y = lerpAngle(this.fromYaw, this.toYaw, a);
    this.facingYaw = this.mesh.rotation.y;
    const dist = camera.position.distanceTo(this.group.position);
    const important = this.rank === "boss" || this.rank === "world" || this.rank === "elite" || this.rank === "mini" || this.kind === "ragha";
    const show = (this.namedTarget || important) && dist < (important ? 32 : 18) && !this.deadPose;
    this.label.visible = show;
    if (show) {
      const w = THREE.MathUtils.clamp(dist * 0.04, 0.55, important ? 1.05 : 0.82);
      this.label.scale.set(w, w * 0.32, 1);
      this.labelMat.opacity = dist > 16 ? Math.max(0.2, 1 - (dist - 16) / 16) : 1;
    }
    const far = dist > 32;
    this.mesh.visible = dist < 70;
    if (this.hitFlash > 0) {
      this.hitFlash -= dt;
      this.bodyMat.emissive.setHex(0xff5533);
      this.bodyMat.emissiveIntensity = far ? 0.3 : 0.8;
      const heavy = this.rank === "boss" || this.rank === "world";
      const lean = heavy ? -0.06 : this.rank === "elite" || this.rank === "mini" ? -0.16 : -0.22;
      if (!this.deadPose) this.mesh.rotation.x = lean * (this.hitFlash / 0.14);
    } else {
      this.bodyMat.emissive.setHex(0x000000);
      this.bodyMat.emissiveIntensity = 0;
      if (!this.deadPose) this.mesh.rotation.x += (0 - this.mesh.rotation.x) * Math.min(1, dt * 10);
    }
  }

  dispose(): void {
    this.labelTex.dispose();
    this.label.material.dispose();
    this.group.traverse((child) => {
      if (!(child instanceof THREE.Mesh)) return;
      child.geometry.dispose();
      const mat = child.material;
      if (Array.isArray(mat)) mat.forEach((m) => m.dispose());
      else mat.dispose();
    });
    this.group.removeFromParent();
  }

  private build(kind: string, rank: string): void {
    const body = buildEnemyMesh(kind, rank, this.bodyMat);
    this.mesh.add(body);
    this.mesh.traverse((o) => {
      if (o instanceof THREE.Mesh) {
        o.castShadow = true;
        o.receiveShadow = true;
      }
    });
    this.group.position.y = PLAYER.groundY;
  }
}

function makePlate(name: string, level: number, health: HealthComponent): { sprite: THREE.Sprite; tex: THREE.CanvasTexture; canvas: HTMLCanvasElement; mat: THREE.SpriteMaterial } {
  const canvas = document.createElement("canvas");
  canvas.width = 256;
  canvas.height = 72;
  drawPlate(canvas, name, level, health);
  const tex = new THREE.CanvasTexture(canvas);
  tex.colorSpace = THREE.SRGBColorSpace;
  const mat = new THREE.SpriteMaterial({ map: tex, transparent: true, depthTest: true, depthWrite: false });
  const sprite = new THREE.Sprite(mat);
  sprite.scale.set(0.78, 0.24, 1);
  sprite.renderOrder = 3;
  return { sprite, tex, canvas, mat };
}

function drawPlate(canvas: HTMLCanvasElement, name: string, level: number, health: HealthComponent): void {
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = "rgba(7,17,31,0.42)";
  ctx.fillRect(36, 10, 184, 52);
  ctx.fillStyle = "#f0c9c9";
  ctx.font = "600 14px Segoe UI, sans-serif";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  ctx.fillText(name, 128, 28);
  ctx.font = "11px Segoe UI, sans-serif";
  ctx.fillStyle = "#d6e4f0";
  ctx.fillText(`LV. ${level}`, 128, 46);
  const ratio = health.getHealthPercent();
  ctx.fillStyle = "rgba(255,255,255,0.16)";
  ctx.fillRect(28, 56, 200, 8);
  ctx.fillStyle = "#ef4444";
  ctx.fillRect(28, 56, 200 * ratio, 8);
}

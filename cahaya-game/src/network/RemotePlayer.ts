import * as THREE from "three";
import { buildPlayerModel, disposePlayerModel } from "../player/PlayerModel";
import type { PlayerFormId } from "../player/PlayerForms";
import { visualForm } from "../progression/ProgressionStore";
import type { AnimState } from "../player/Player";
import { interpFactor, lerpAngle } from "./Interpolation";
import type { MovementState, PlayerSnapshot, PlayerSpawn } from "./NetworkMessage";

export interface RemotePlayerState {
  id: string;
  name: string;
  level: number;
  className: string;
  hp: number;
  maxHp: number;
  movementState: MovementState | string;
  guildTag?: string;
  title?: string;
  mounted?: boolean;
}

export class RemotePlayer {
  readonly id: string;
  readonly group = new THREE.Group();
  readonly mesh = new THREE.Group();
  readonly state: RemotePlayerState;
  private readonly currentPosition = new THREE.Vector3();
  private readonly targetPosition = new THREE.Vector3();
  private interpolationFactor = 1;
  private elapsed = 0;
  private fromYaw = 0;
  private toYaw = 0;
  private readonly label: THREE.Sprite;
  private readonly labelTex: THREE.CanvasTexture;
  private readonly labelCanvas: HTMLCanvasElement;
  private readonly labelMat: THREE.SpriteMaterial;
  private armL: THREE.Object3D;
  private armR: THREE.Object3D;
  private legL: THREE.Object3D;
  private legR: THREE.Object3D;
  private animTime = 0;
  lastOffset = 0;
  combatPose = "";
  private combatPoseTimer = 0;
  private visualForm: PlayerFormId = "asal";

  constructor(spawn: PlayerSpawn) {
    this.id = spawn.playerId;
    this.state = {
      id: spawn.playerId,
      name: spawn.name,
      level: spawn.level,
      className: spawn.class,
      hp: spawn.hp,
      maxHp: spawn.maxHp,
      movementState: spawn.state,
      guildTag: spawn.guildTag,
      title: spawn.title,
      mounted: spawn.mounted,
    };
    const rig = buildPlayerModel("asal");
    this.mesh.add(rig.root);
    this.armL = rig.armL;
    this.armR = rig.armR;
    this.legL = rig.legL;
    this.legR = rig.legR;
    this.group.add(this.mesh);
    const plate = makeNameplate(this.state);
    this.label = plate.sprite;
    this.labelTex = plate.tex;
    this.labelCanvas = plate.canvas;
    this.labelMat = plate.mat;
    this.label.position.y = 2.02;
    this.group.add(this.label);
    this.currentPosition.set(spawn.x, spawn.y, spawn.z);
    this.targetPosition.copy(this.currentPosition);
    this.fromYaw = spawn.yaw;
    this.toYaw = spawn.yaw;
    this.group.position.copy(this.currentPosition);
    this.mesh.rotation.y = spawn.yaw;
    if (spawn.formId) this.applyForm(spawn.formId);
  }

  applyForm(formId: string): void {
    const next = visualForm(formId);
    if (next === this.visualForm && this.mesh.children.length > 0) return;
    this.visualForm = next;
    disposePlayerModel(this.mesh);
    const rig = buildPlayerModel(next);
    this.mesh.add(rig.root);
    this.armL = rig.armL;
    this.armR = rig.armR;
    this.legL = rig.legL;
    this.legR = rig.legR;
  }

  applyMount(mounted: boolean): void {
    let mount = this.group.getObjectByName("wind-runner");
    if (mounted) {
      if (!mount) {
        mount = makeRemoteMount();
        this.group.add(mount);
      }
      mount.visible = true;
      this.mesh.position.y = 0.72;
    } else if (mount) {
      mount.visible = false;
      this.mesh.position.y = 0;
    }
  }

  applySnapshot(snap: PlayerSnapshot): void {
    this.lastOffset = this.group.position.distanceTo(this.targetPosition);
    this.currentPosition.copy(this.group.position);
    this.targetPosition.set(snap.x, snap.y, snap.z);
    this.fromYaw = this.mesh.rotation.y;
    this.toYaw = snap.yaw;
    this.elapsed = 0;
    this.interpolationFactor = 0;
    this.state.movementState = snap.st;
    if (typeof snap.hp === "number") this.state.hp = snap.hp;
    if (typeof snap.maxHp === "number") this.state.maxHp = snap.maxHp;
    if (typeof snap.level === "number") this.state.level = snap.level;
    if (snap.guildTag !== undefined) this.state.guildTag = snap.guildTag;
    if (snap.title !== undefined) this.state.title = snap.title;
    if (snap.formId) this.applyForm(snap.formId);
    if (snap.cs === "DEAD" || snap.cs === "DOWNED") this.combatPose = "DEAD";
    this.applyMount(!!snap.mounted);
    drawNameplate(this.labelCanvas, this.state);
    this.labelTex.needsUpdate = true;
  }

  applyHp(hp: number, maxHp: number): void {
    this.state.hp = hp;
    this.state.maxHp = maxHp;
    drawNameplate(this.labelCanvas, this.state);
    this.labelTex.needsUpdate = true;
  }

  playCombat(attackType: string, skillId?: string): void {
    const kind = skillId ?? attackType;
    this.combatPose = kind === "kick" || kind === "whirlwind_kick" ? "KICK"
      : kind === "dodge" ? "DODGE"
        : kind === "dead" ? "DEAD"
          : kind === "energy_bolt" || kind === "energy" ? "ENERGY"
            : "PUNCH";
    this.combatPoseTimer = 0.35;
  }

  applySpawn(spawn: PlayerSpawn): void {
    this.state.name = spawn.name;
    this.state.level = spawn.level;
    this.state.className = spawn.class;
    this.state.hp = spawn.hp;
    this.state.maxHp = spawn.maxHp;
    this.state.guildTag = spawn.guildTag;
    this.state.title = spawn.title;
    this.state.mounted = spawn.mounted;
    if (spawn.formId) this.applyForm(spawn.formId);
    this.state.movementState = spawn.state;
    drawNameplate(this.labelCanvas, this.state);
    this.labelTex.needsUpdate = true;
    this.applySnapshot({
      id: spawn.playerId,
      x: spawn.x,
      y: spawn.y,
      z: spawn.z,
      yaw: spawn.yaw,
      vx: 0,
      vy: 0,
      vz: 0,
      st: spawn.state,
      seq: 0,
    });
  }

  update(dt: number, camera: THREE.Camera): void {
    this.elapsed += dt;
    this.interpolationFactor = interpFactor(this.elapsed);
    this.group.position.lerpVectors(this.currentPosition, this.targetPosition, this.interpolationFactor);
    this.mesh.rotation.y = lerpAngle(this.fromYaw, this.toYaw, this.interpolationFactor);
    this.animTime += dt;
    if (this.combatPoseTimer > 0) this.combatPoseTimer -= dt;
    const dist = camera.position.distanceTo(this.group.position);
    const show = dist <= 22;
    this.label.visible = show;
    if (show) {
      const w = THREE.MathUtils.clamp(dist * 0.04, 0.58, 0.92);
      this.label.scale.set(w, w * 0.38, 1);
      this.labelMat.opacity = dist > 14 ? Math.max(0.25, 1 - (dist - 14) / 8) : 1;
    }
    if (dist > 38) return;
    if (this.combatPoseTimer > 0 || this.combatPose === "DEAD") this.applyCombatPose();
    else this.applyPose(this.state.movementState);
  }

  dispose(): void {
    this.labelTex.dispose();
    this.label.material.dispose();
    disposePlayerModel(this.mesh);
    this.group.removeFromParent();
  }

  private applyCombatPose(): void {
    if (this.combatPose === "DEAD") {
      this.mesh.rotation.x = Math.PI / 2;
      return;
    }
    this.mesh.rotation.x = 0;
    if (this.combatPose === "KICK") {
      this.legR.rotation.x = -1.2;
      this.armL.rotation.x = -0.4;
      return;
    }
    if (this.combatPose === "DODGE") {
      this.armL.rotation.x = 0.5;
      this.armR.rotation.x = 0.5;
      return;
    }
    if (this.combatPose === "ENERGY") {
      this.armL.rotation.x = -1.2;
      this.armR.rotation.x = -1.2;
      return;
    }
    this.armR.rotation.x = -1.15;
    this.armL.rotation.x = 0.35;
  }

  private applyPose(state: string): void {
    const anim = toAnim(state);
    const run = anim === "RUN";
    const walk = anim === "WALK";
    const air = anim === "JUMP" || anim === "FALL";
    const amp = run ? 0.85 : walk ? 0.5 : 0;
    const freq = run ? 12 : walk ? 7.5 : 2.2;
    const swing = Math.sin(this.animTime * freq) * amp;
    if (air) {
      this.armL.rotation.x = -0.7;
      this.armR.rotation.x = -0.7;
      this.legL.rotation.x = 0.35;
      this.legR.rotation.x = -0.15;
      return;
    }
    this.armL.rotation.x = swing;
    this.armR.rotation.x = -swing;
    this.legL.rotation.x = -swing;
    this.legR.rotation.x = swing;
  }
}

function toAnim(state: string): AnimState {
  if (state === "RUN" || state === "WALK" || state === "JUMP" || state === "FALL" || state === "IDLE") {
    return state;
  }
  return "IDLE";
}

function makeNameplate(state: RemotePlayerState): { sprite: THREE.Sprite; tex: THREE.CanvasTexture; canvas: HTMLCanvasElement; mat: THREE.SpriteMaterial } {
  const canvas = document.createElement("canvas");
  canvas.width = 256;
  canvas.height = 96;
  drawNameplate(canvas, state);
  const tex = new THREE.CanvasTexture(canvas);
  tex.colorSpace = THREE.SRGBColorSpace;
  const mat = new THREE.SpriteMaterial({ map: tex, transparent: true, depthTest: true, depthWrite: false });
  const sprite = new THREE.Sprite(mat);
  sprite.scale.set(0.82, 0.32, 1);
  sprite.renderOrder = 3;
  return { sprite, tex, canvas, mat };
}

function drawNameplate(canvas: HTMLCanvasElement, state: RemotePlayerState): void {
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = "rgba(7,17,31,0.4)";
  ctx.fillRect(36, 12, 184, 72);
  ctx.fillStyle = "#efe3b0";
  ctx.font = "600 11px Segoe UI, sans-serif";
  ctx.textAlign = "center";
  ctx.textBaseline = "middle";
  const prefix = [state.title ? `[${state.title}]` : "", state.guildTag ? `<${state.guildTag}>` : ""].filter(Boolean).join(" ");
  if (prefix) ctx.fillText(prefix, 128, 22);
  ctx.fillStyle = "#f4f8fc";
  ctx.font = "600 16px Segoe UI, sans-serif";
  ctx.fillText(state.name.toUpperCase(), 128, prefix ? 44 : 28);
  ctx.fillStyle = "#d6e4f0";
  ctx.font = "12px Segoe UI, sans-serif";
  ctx.fillText(`LV. ${state.level}${state.mounted ? "  ·  MOUNT" : ""}`, 128, prefix ? 66 : 50);
  const ratio = state.maxHp <= 0 ? 1 : Math.max(0, Math.min(1, state.hp / state.maxHp));
  ctx.fillStyle = "rgba(255,255,255,0.16)";
  ctx.fillRect(48, 78, 160, 6);
  ctx.fillStyle = "#ef4444";
  ctx.fillRect(48, 78, 160 * ratio, 6);
}

function makeRemoteMount(): THREE.Group {
  const g = new THREE.Group();
  g.name = "wind-runner";
  const hide = new THREE.MeshStandardMaterial({ color: 0xc9d6e4, roughness: 0.55 });
  const body = new THREE.Mesh(new THREE.CapsuleGeometry(0.38, 1.1, 4, 8), hide);
  body.rotation.z = Math.PI / 2;
  body.position.set(0, 0.55, 0);
  const head = new THREE.Mesh(new THREE.SphereGeometry(0.28, 8, 8), hide);
  head.position.set(0, 0.85, 0.7);
  g.add(body, head);
  return g;
}

import * as THREE from "three";
import { CAMERA, COMBAT } from "./GameConfig";
import type { Player } from "../player/Player";

export class CameraController {
  yaw = 0;
  pitch = 0.22;
  distance: number = CAMERA.distance;
  private readonly desired = new THREE.Vector3();
  private readonly look = new THREE.Vector3();
  private readonly ray = new THREE.Raycaster();
  private readonly shakeOff = new THREE.Vector3();
  private dragging = false;
  private allowPointerLock = true;
  private lockPoint: THREE.Vector3 | null = null;
  private shake = 0;
  private zoomPulse = 0;
  private fovPulse = 0;
  private cinematic = 0;
  private cinematicMode = "follow";
  private mountedBoost = 0;
  private presence: "explore" | "combat" | "boss" = "explore";
  private presenceDist = CAMERA.distance;
  private readonly camOff = new THREE.Vector3();
  private peekLeft = 0;
  private readonly peekLook = new THREE.Vector3();

  setPresence(mode: "explore" | "combat" | "boss"): void {
    this.presence = mode;
  }

  setMounted(on: boolean): void {
    this.mountedBoost = on ? 1.35 : 0;
  }

  beginCinematic(mode = "follow"): void {
    this.cinematic = 1;
    this.cinematicMode = mode || "follow";
    this.pitch = this.cinematicMode === "zoom" ? 0.02 : 0.1;
  }

  endCinematic(): void {
    this.cinematic = 0;
  }
  private readonly occluders: THREE.Object3D[];

  constructor(
    readonly camera: THREE.PerspectiveCamera,
    private readonly player: Player,
    private readonly canvas: HTMLCanvasElement,
    occluders: THREE.Object3D[],
    private readonly autoOrbit: boolean,
  ) {
    this.occluders = occluders;
    this.canvas.addEventListener("pointerdown", this.onPointerDown);
    window.addEventListener("pointerup", this.onPointerUp);
    window.addEventListener("pointermove", this.onPointerMove);
    this.canvas.addEventListener("wheel", this.onWheel, { passive: false });
    document.addEventListener("pointerlockchange", this.onLockChange);
  }

  dispose(): void {
    this.canvas.removeEventListener("pointerdown", this.onPointerDown);
    window.removeEventListener("pointerup", this.onPointerUp);
    window.removeEventListener("pointermove", this.onPointerMove);
    this.canvas.removeEventListener("wheel", this.onWheel);
    document.removeEventListener("pointerlockchange", this.onLockChange);
    if (document.pointerLockElement === this.canvas) document.exitPointerLock();
  }

  setPointerLockEnabled(on: boolean): void {
    this.allowPointerLock = on;
    if (!on) this.exitLock();
  }

  requestLock(): void {
    if (this.autoOrbit || !this.allowPointerLock) return;
    if (document.pointerLockElement !== this.canvas) void this.canvas.requestPointerLock();
  }

  exitLock(): void {
    if (document.pointerLockElement === this.canvas) document.exitPointerLock();
  }

  getYaw(): number {
    return this.yaw;
  }

  setLockTarget(pos: THREE.Vector3 | null): void {
    this.lockPoint = pos;
  }

  addShake(amount: number): void {
    this.shake = Math.min(COMBAT.shakeMax, this.shake + amount);
  }

  pulseZoom(): void {
    this.zoomPulse = 0.55;
    this.addShake(0.16);
  }

  pulseFov(): void {
    this.fovPulse = 0.55;
  }

  peekAt(x: number, y: number, z: number, seconds = 2.8): void {
    this.peekLeft = seconds;
    this.peekLook.set(x, y, z);
  }

  update(dt: number): void {
    if (this.autoOrbit && !this.dragging && document.pointerLockElement !== this.canvas) {
      this.yaw += dt * (this.cinematic > 0 ? 0.04 : 0.12);
    }
    if (this.cinematic > 0) {
      const mode = this.cinematicMode;
      if (mode === "pan") this.yaw += dt * 0.12;
      else if (mode === "focus") this.yaw += dt * 0.01;
      else if (mode === "zoom") this.yaw += dt * 0.02;
      else this.yaw += dt * 0.03;
      const targetPitch = mode === "zoom" ? 0.02 : mode === "focus" ? 0.16 : 0.08;
      this.pitch += (targetPitch - this.pitch) * Math.min(1, dt * 1.2);
    }

    const origin = this.look.set(
      this.player.position.x,
      this.player.position.y + CAMERA.lookY,
      this.player.position.z,
    );
    if (this.lockPoint) {
      origin.x += (this.lockPoint.x - origin.x) * 0.1;
      origin.z += (this.lockPoint.z - origin.z) * 0.1;
    } else if (this.peekLeft > 0) {
      origin.x += (this.peekLook.x - origin.x) * 0.08;
      origin.y += (this.peekLook.y - origin.y) * 0.05;
      origin.z += (this.peekLook.z - origin.z) * 0.08;
      this.peekLeft = Math.max(0, this.peekLeft - dt);
    }
    const extraH = this.lockPoint ? 0.95 : this.presence === "boss" ? 0.55 : 0;
    const want =
      this.presence === "combat" ? Math.min(this.distance, 6.15) :
      this.presence === "boss" ? Math.max(this.distance, 10.4) :
      Math.max(this.distance, 8.35);
    this.presenceDist += (want - this.presenceDist) * Math.min(1, dt * 3.2);
    const dist = (this.lockPoint ? Math.max(this.presenceDist, COMBAT.lockCameraDistance) : this.presenceDist) + this.mountedBoost;
    const zoomPulse = this.zoomPulse > 0 ? 0.86 : 1;
    const cinZoom = this.cinematic > 0 && this.cinematicMode === "zoom" ? 0.62 : this.cinematic > 0 && this.cinematicMode === "focus" ? 0.82 : 1;
    const zoom = zoomPulse * cinZoom;
    this.zoomPulse = Math.max(0, this.zoomPulse - dt);
    this.desired.set(
      origin.x + Math.sin(this.yaw) * -dist * zoom,
      origin.y + CAMERA.height - CAMERA.lookY + this.pitch * 3.4 + extraH,
      origin.z + Math.cos(this.yaw) * -dist * zoom,
    );
    this.desired.y = Math.max(0.45, this.desired.y);

    this.camOff.copy(this.desired).sub(origin);
    const maxDist = this.camOff.length();
    let blocked = false;
    if (maxDist > 0.001) {
      this.camOff.multiplyScalar(1 / maxDist);
      this.ray.set(origin, this.camOff);
      this.ray.far = maxDist;
      const hits = this.ray.intersectObjects(this.occluders, true);
      const hit = hits.find((h) => h.distance > 0.2 && objectWorldVisible(h.object));
      if (hit) {
        blocked = true;
        this.desired.copy(origin).addScaledVector(this.camOff, Math.max(CAMERA.minDistance * 0.45, hit.distance - CAMERA.collisionPadding));
      }
    }

    const smooth = 1 - Math.exp(-CAMERA.smoothness * dt * (blocked ? 3.2 : 1));
    this.camera.position.lerp(this.desired, blocked ? Math.max(smooth, 0.55) : smooth);
    if (this.shake > 0) {
      this.shakeOff.set((Math.random() - 0.5) * this.shake, (Math.random() - 0.5) * this.shake * 0.6, (Math.random() - 0.5) * this.shake);
      this.camera.position.add(this.shakeOff);
      this.shake *= Math.exp(-12 * dt);
      if (this.shake < 0.003) this.shake = 0;
    }
    this.camera.lookAt(origin);
    if (this.fovPulse > 0) {
      this.camera.fov = 50 + 6 * this.fovPulse;
      this.camera.updateProjectionMatrix();
      this.fovPulse = Math.max(0, this.fovPulse - dt * 1.4);
      if (this.fovPulse <= 0) {
        this.camera.fov = 50;
        this.camera.updateProjectionMatrix();
      }
    }
  }

  private onPointerDown = (e: PointerEvent): void => {
    if (e.button !== 0) return;
    this.dragging = true;
    if (!this.autoOrbit && this.allowPointerLock) this.requestLock();
  };

  private onPointerUp = (): void => {
    this.dragging = false;
  };

  private onPointerMove = (e: PointerEvent): void => {
    const locked = document.pointerLockElement === this.canvas;
    if (!locked && !this.dragging) return;
    this.yaw -= e.movementX * CAMERA.rotateSpeed;
    this.pitch = THREE.MathUtils.clamp(
      this.pitch + e.movementY * CAMERA.rotateSpeed,
      CAMERA.minPitch,
      CAMERA.maxPitch,
    );
  };

  private onWheel = (e: WheelEvent): void => {
    e.preventDefault();
    const dir = Math.sign(e.deltaY);
    this.distance = THREE.MathUtils.clamp(
      this.distance + dir * 0.45,
      CAMERA.minDistance,
      CAMERA.maxDistance,
    );
  };

  private onLockChange = (): void => {
    this.dragging = document.pointerLockElement === this.canvas;
  };
}

function objectWorldVisible(obj: THREE.Object3D): boolean {
  let cur: THREE.Object3D | null = obj;
  while (cur) {
    if (!cur.visible) return false;
    cur = cur.parent;
  }
  return true;
}

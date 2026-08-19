import * as THREE from "three";
import { PLAYER, STAMINA } from "../game/GameConfig";
import type { InputManager } from "../game/InputManager";
import type { CollisionWorld } from "../world/Collision";
import type { Player } from "./Player";

export type MoveMode = "normal" | "halt" | "dash";

export class PlayerController {
  private readonly planar = new THREE.Vector3();
  private readonly desired = new THREE.Vector3();

  constructor(
    private readonly player: Player,
    private readonly input: InputManager,
    private readonly getCameraYaw: () => number,
    private readonly collision: CollisionWorld,
    private readonly mountSpeed: () => number = () => 1,
  ) {}

  update(dt: number, move: MoveMode = "normal"): void {
    const stats = this.player.stats;
    if (stats.stamina <= 0) stats.exhausted = true;
    if (stats.exhausted && stats.stamina >= stats.maxStamina * 0.25) stats.exhausted = false;

    if (move === "normal") {
      const forward = this.input.compositeForward();
      const strafe = this.input.compositeStrafe();
      const inputMag = Math.min(1, Math.hypot(forward, strafe));
      const moving = inputMag > 0.08;
      const wantSprint = this.input.isSprintPressed() && moving && this.player.isGrounded;
      const canSprint = wantSprint && !stats.exhausted && stats.stamina > STAMINA.minToSprint;
      const targetSpeed = (canSprint ? PLAYER.runSpeed : PLAYER.walkSpeed) * inputMag * this.mountSpeed();

      if (canSprint) {
        stats.stamina = Math.max(0, stats.stamina - STAMINA.sprintDrainPerSec * dt);
      } else {
        stats.stamina = Math.min(stats.maxStamina, stats.stamina + STAMINA.recoverPerSec * dt);
      }

      if (moving) {
        const yaw = cameraMoveYaw(this.getCameraYaw(), strafe, forward);
        this.desired.set(Math.sin(yaw), 0, Math.cos(yaw)).multiplyScalar(targetSpeed);
        const accel = PLAYER.acceleration * dt;
        this.planar.set(this.player.velocity.x, 0, this.player.velocity.z);
        this.planar.lerp(this.desired, 1 - Math.exp(-accel));
        this.player.velocity.x = this.planar.x;
        this.player.velocity.z = this.planar.z;
        this.player.facingYaw = yaw;
        const turn = 1 - Math.exp(-PLAYER.turnSpeed * dt);
        this.player.mesh.rotation.y = lerpAngle(this.player.mesh.rotation.y, yaw, turn);
      } else {
        const decel = Math.exp(-PLAYER.deceleration * dt);
        this.player.velocity.x *= decel;
        this.player.velocity.z *= decel;
        if (Math.hypot(this.player.velocity.x, this.player.velocity.z) < 0.04) {
          this.player.velocity.x = 0;
          this.player.velocity.z = 0;
        }
      }

      if (this.input.isJumpPressed() && this.player.isGrounded) {
        this.player.velocity.y = PLAYER.jumpForce;
        this.player.isGrounded = false;
      }
    } else {
      stats.stamina = Math.min(stats.maxStamina, stats.stamina + STAMINA.recoverPerSec * dt);
      if (move === "halt") {
        const decel = Math.exp(-PLAYER.deceleration * dt);
        this.player.velocity.x *= decel;
        this.player.velocity.z *= decel;
      }
      void this.input.isJumpPressed();
    }

    this.player.velocity.y -= PLAYER.gravity * dt;
    this.player.position.x += this.player.velocity.x * dt;
    this.player.position.y += this.player.velocity.y * dt;
    this.player.position.z += this.player.velocity.z * dt;

    const horiz = this.collision.resolveHorizontal(this.player.position.x, this.player.position.z);
    this.player.position.x = horiz.x;
    this.player.position.z = horiz.z;

    const vertical = this.collision.resolveVertical(this.player);
    if (vertical.landed) this.player.landTimer = 0.16;

    if (this.player.landTimer > 0) this.player.landTimer -= dt;

    const limit = PLAYER.worldLimit;
    this.player.position.x = THREE.MathUtils.clamp(this.player.position.x, -limit, limit);
    this.player.position.z = THREE.MathUtils.clamp(this.player.position.z, -limit, limit);
    this.player.currentSpeed = Math.hypot(this.player.velocity.x, this.player.velocity.z);
  }
}

function lerpAngle(from: number, to: number, t: number): number {
  let diff = to - from;
  while (diff > Math.PI) diff -= Math.PI * 2;
  while (diff < -Math.PI) diff += Math.PI * 2;
  return from + diff * t;
}

/** Camera-relative heading. Strafe is negated so D/right matches Three.js camera +X. */
export function cameraMoveYaw(cameraYaw: number, strafe: number, forward: number): number {
  return cameraYaw + Math.atan2(-strafe, forward);
}

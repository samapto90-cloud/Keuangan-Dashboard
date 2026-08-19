import * as THREE from "three";
import { COMBAT } from "../game/GameConfig";
import { ATTACKS } from "./AttackData";
import type { Player } from "../player/Player";
import { cameraMoveYaw } from "../player/PlayerController";
import type { InputManager } from "../game/InputManager";

export class DodgeSystem {
  active = false;
  iFrame = 0;
  cooldown = 0;
  private timer = 0;
  private readonly dir = new THREE.Vector3();

  canStart(player: Player): boolean {
    return (
      !this.active &&
      this.cooldown <= 0 &&
      player.stats.stamina >= ATTACKS.DODGE.staminaCost &&
      !player.health.isDead() &&
      player.combatState !== "HIT" &&
      player.combatState !== "STUNNED"
    );
  }

  tryStart(player: Player, input: InputManager, cameraYaw: number): boolean {
    if (!this.canStart(player)) return false;
    const forward = input.compositeForward();
    const strafe = input.compositeStrafe();
    if (Math.hypot(forward, strafe) > 0.12) {
      const yaw = cameraMoveYaw(cameraYaw, strafe, forward);
      this.dir.set(Math.sin(yaw), 0, Math.cos(yaw)).normalize();
      player.facingYaw = yaw;
      player.mesh.rotation.y = yaw;
    } else {
      this.dir.set(Math.sin(player.facingYaw), 0, Math.cos(player.facingYaw));
    }
    player.stats.stamina -= ATTACKS.DODGE.staminaCost;
    this.active = true;
    this.timer = COMBAT.dodgeDuration;
    this.iFrame = COMBAT.dodgeIFrame;
    this.cooldown = COMBAT.dodgeCooldown;
    player.combatState = "DODGING";
    player.invincible = true;
    player.velocity.x = this.dir.x * COMBAT.dodgeSpeed;
    player.velocity.z = this.dir.z * COMBAT.dodgeSpeed;
    return true;
  }

  update(dt: number, player: Player): void {
    if (this.cooldown > 0) this.cooldown -= dt;
    if (!this.active) return;
    this.timer -= dt;
    this.iFrame -= dt;
    player.invincible = this.iFrame > 0;
    player.velocity.x = this.dir.x * COMBAT.dodgeSpeed;
    player.velocity.z = this.dir.z * COMBAT.dodgeSpeed;
    if (this.timer <= 0) {
      this.active = false;
      player.invincible = false;
      if (player.combatState === "DODGING") player.combatState = "IDLE";
      player.velocity.x *= 0.35;
      player.velocity.z *= 0.35;
    }
  }
}

import type { AnimState, Player } from "./Player";

export class PlayerAnimationController {
  private time = 0;
  private landLock = 0;
  private armLX = 0;
  private armRX = 0;
  private armLZ = 0.08;
  private armRZ = -0.08;
  private legLX = 0;
  private legRX = 0;
  private meshX = 0;

  update(player: Player, dt: number, moving: boolean, sprinting: boolean, glance = 0): void {
    this.time += dt;
    const next = this.resolve(player, moving, sprinting);
    if (next === "LAND" && player.animState !== "LAND") this.landLock = 0.18;
    if (player.animState === "LAND" && !player.combatPose) {
      this.landLock -= dt;
      if (this.landLock > 0) player.animState = "LAND";
      else player.animState = next === "LAND" ? "IDLE" : next;
    } else {
      player.animState = next;
    }
    this.applyPose(player, dt, glance);
  }

  private resolve(player: Player, moving: boolean, sprinting: boolean): AnimState {
    if (player.combatPose) return player.combatPose;
    if (!player.isGrounded) {
      return player.velocity.y > 0.6 ? "JUMP" : "FALL";
    }
    if (player.landTimer > 0) return "LAND";
    if (moving && sprinting) return "RUN";
    if (moving) return "WALK";
    return "IDLE";
  }

  private applyPose(player: Player, dt: number, glance = 0): void {
    const { armL, armR, legL, legR, animState, aura, lightning } = player;
    if (!armL || !armR || !legL || !legR) return;

    let tArmLX = 0;
    let tArmRX = 0;
    let tArmLZ = 0.08;
    let tArmRZ = -0.08;
    let tLegLX = 0;
    let tLegRX = 0;
    let tMeshX = 0;

    if (animState === "PUNCH" || animState === "COMBO") {
      const wind = (Math.sin(this.time * 14) + 1) * 0.5;
      tArmRX = -0.45 - wind * 1.05;
      tArmLX = 0.35;
      tLegLX = 0.12;
      tLegRX = -0.16;
    } else if (animState === "KICK") {
      const wind = (Math.sin(this.time * 11) + 1) * 0.5;
      tArmLX = -0.4;
      tArmRX = -0.2;
      tLegRX = -0.4 - wind * 1.15;
      tLegLX = 0.22;
    } else if (animState === "ENERGY_ATTACK") {
      tArmLX = -1.2;
      tArmRX = -1.2;
      tArmLZ = 0.25;
      tArmRZ = -0.25;
    } else if (animState === "DODGE") {
      tArmLX = 0.5;
      tArmRX = 0.5;
      tMeshX = 0.18;
    } else if (animState === "HIT" || animState === "STUN") {
      tArmLX = 0.4;
      tArmRX = 0.4;
      tMeshX = -0.12;
    } else if (animState === "DEAD") {
      tMeshX = Math.PI / 2;
    } else {
      const walk = animState === "WALK";
      const run = animState === "RUN";
      const air = animState === "JUMP" || animState === "FALL";
      const amp = run ? 0.85 : walk ? 0.5 : 0;
      const freq = run ? 12 : walk ? 7.5 : 2.2;
      const swing = Math.sin(this.time * freq) * amp;
      if (air) {
        tArmLX = -0.7;
        tArmRX = -0.7;
        tLegLX = 0.35;
        tLegRX = -0.15;
      } else if (animState === "LAND") {
        tArmLX = 0.2;
        tArmRX = 0.2;
        tLegLX = 0.15;
        tLegRX = 0.15;
      } else {
        tArmLX = swing;
        tArmRX = -swing;
        tLegLX = -swing;
        tLegRX = swing;
        if (animState === "IDLE") {
          const hang = Math.sin(this.time * 2.1) * 0.04;
          tArmLZ = 0.08 + hang;
          tArmRZ = -0.08 - hang;
        }
      }
    }

    const k = 1 - Math.exp(-14 * dt);
    this.armLX += (tArmLX - this.armLX) * k;
    this.armRX += (tArmRX - this.armRX) * k;
    this.armLZ += (tArmLZ - this.armLZ) * k;
    this.armRZ += (tArmRZ - this.armRZ) * k;
    this.legLX += (tLegLX - this.legLX) * k;
    this.legRX += (tLegRX - this.legRX) * k;
    this.meshX += (tMeshX - this.meshX) * k;
    armL.rotation.x = this.armLX;
    armR.rotation.x = this.armRX;
    armL.rotation.z = this.armLZ;
    armR.rotation.z = this.armRZ;
    legL.rotation.x = this.legLX;
    legR.rotation.x = this.legRX;
    player.mesh.rotation.x = this.meshX;
    if (animState === "IDLE" && glance > 0) {
      player.mesh.traverse((o) => {
        if (o.userData.attach === "head") o.rotation.y = glance * 0.55;
      });
    } else {
      player.mesh.traverse((o) => {
        if (o.userData.attach === "head") o.rotation.y *= 1 - Math.min(1, dt * 6);
      });
    }

    if (aura) {
      const pulse = 1 + Math.sin(this.time * 5) * 0.04;
      aura.scale.set(pulse, 1 + Math.sin(this.time * 3) * 0.03, pulse);
    }
    if (lightning) {
      lightning.visible = Math.sin(this.time * 28) > 0.15;
      lightning.rotation.y = this.time * 2.2;
    }
  }
}

import * as THREE from "three";
import { NET } from "../game/GameConfig";
import type { Player } from "../player/Player";
import { planarDistance } from "./Interpolation";
import type { PlayerSnapshot } from "./NetworkMessage";

export class Reconciliation {
  offset = 0;
  serverPos = new THREE.Vector3();

  apply(player: Player, snap: PlayerSnapshot | null, dt: number): void {
    if (!snap) {
      this.offset = 0;
      return;
    }
    this.serverPos.set(snap.x, snap.y, snap.z);
    this.offset = planarDistance(player.position.x, player.position.z, snap.x, snap.z);
    const dy = Math.abs(player.position.y - snap.y);
    if (this.offset > NET.correctHard || dy > NET.correctHard) {
      player.position.set(snap.x, snap.y, snap.z);
      player.velocity.set(snap.vx, snap.vy, snap.vz);
      player.mesh.rotation.y = snap.yaw;
      player.facingYaw = snap.yaw;
      return;
    }
    if (this.offset < NET.correctSoft && dy < NET.correctSoft) return;
    const t = 1 - Math.exp(-8 * dt);
    player.position.x = THREE.MathUtils.lerp(player.position.x, snap.x, t);
    player.position.y = THREE.MathUtils.lerp(player.position.y, snap.y, t);
    player.position.z = THREE.MathUtils.lerp(player.position.z, snap.z, t);
  }
}

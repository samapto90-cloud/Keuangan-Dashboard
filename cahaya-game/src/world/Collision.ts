import { PLAYER } from "../game/GameConfig";
import type { Player } from "../player/Player";
import {
  npcBodyRadius,
  resolveWorldXZ,
  softSeparationXZ,
  type AabbCollider,
  type CircleCollider,
  type ColliderKind,
} from "./WorldColliders";

/** @deprecated use ColliderKind from WorldColliders */
export type { ColliderKind };

export interface SphereCollider {
  kind: ColliderKind;
  x: number;
  y: number;
  z: number;
  radius: number;
}

export class CollisionWorld {
  readonly obstacles: SphereCollider[] = [];
  readonly aabbs: AabbCollider[] = [];
  readonly circles: CircleCollider[] = [];
  readonly trees: CircleCollider[] = [];
  readonly rocks: CircleCollider[] = [];

  getGroundHeight(_x: number, _z: number): number {
    return PLAYER.groundY;
  }

  resolveVertical(player: Player): { grounded: boolean; landed: boolean } {
    const ground = this.getGroundHeight(player.position.x, player.position.z);
    const wasAir = !player.isGrounded;
    if (!Number.isFinite(player.position.y) || player.position.y < ground - 8 || player.position.y > 40) {
      player.position.y = ground;
      player.velocity.y = 0;
      player.isGrounded = true;
      return { grounded: true, landed: wasAir };
    }
    if (player.position.y <= ground) {
      player.position.y = ground;
      if (player.velocity.y < 0) player.velocity.y = 0;
      player.isGrounded = true;
      return { grounded: true, landed: wasAir };
    }
    player.isGrounded = false;
    return { grounded: false, landed: false };
  }

  resolveHorizontal(x: number, z: number, radius = PLAYER.radius): { x: number; z: number } {
    return resolveWorldXZ(x, z, radius, this.aabbs, this.circles, this.trees, this.rocks);
  }

  resolveNpcSoft(
    x: number,
    z: number,
    radius: number,
    npcs: Array<{ x: number; z: number; type: string }>,
    skipIdx = -1,
  ): { x: number; z: number } {
    const others = npcs
      .map((n, i) => (i === skipIdx ? null : { x: n.x, z: n.z, r: npcBodyRadius(n.type) }))
      .filter((o): o is { x: number; z: number; r: number } => o != null);
    const sep = softSeparationXZ(x, z, radius, others);
    return { x: x + sep.x * 0.35, z: z + sep.z * 0.35 };
  }

  resolvePlayerPosition(player: Player, npcs: Array<{ x: number; z: number; type: string }>): void {
    let { x, z } = this.resolveHorizontal(player.position.x, player.position.z, PLAYER.radius);
    const nudge = this.resolveNpcSoft(x, z, PLAYER.radius, npcs);
    x = nudge.x;
    z = nudge.z;
    ({ x, z } = this.resolveHorizontal(x, z, PLAYER.radius));
    player.position.x = x;
    player.position.z = z;
  }
}

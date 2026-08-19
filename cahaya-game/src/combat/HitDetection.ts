import * as THREE from "three";
import type { AttackData } from "./AttackData";
import type { Combatant } from "./Combatant";

export interface HitQuery {
  attacker: Combatant;
  attack: AttackData;
  alreadyHit: Set<string>;
}

const hitCenter = new THREE.Vector3();
const hurtCenter = new THREE.Vector3();

export class HitDetection {
  collect(query: HitQuery, candidates: Combatant[]): Combatant[] {
    const { attacker, attack, alreadyHit } = query;
    if (attack.hitboxRadius <= 0) return [];
    const yaw = attacker.facingYaw;
    hitCenter.set(
      attacker.position.x + Math.sin(yaw) * attack.range * 0.55,
      attacker.position.y + attack.hitboxHeight,
      attacker.position.z + Math.cos(yaw) * attack.range * 0.55,
    );
    const hits: Combatant[] = [];
    for (const target of candidates) {
      if (target.id === attacker.id) continue;
      if (target.team === attacker.team) continue;
      if (target.health.isDead() || target.invincible) continue;
      if (alreadyHit.has(target.id)) continue;
      hurtCenter.set(
        target.position.x,
        target.position.y + target.hurtbox.yOffset,
        target.position.z,
      );
      const dist = hitCenter.distanceTo(hurtCenter);
      if (dist <= attack.hitboxRadius + target.hurtbox.radius) hits.push(target);
    }
    return hits;
  }
}

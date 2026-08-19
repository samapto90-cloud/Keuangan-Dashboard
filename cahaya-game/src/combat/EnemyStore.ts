import * as THREE from "three";
import type { EnemySnapshot } from "../network/NetworkMessage";
import { NetworkEnemy } from "./NetworkEnemy";
import type { Combatant } from "./Combatant";

export class EnemyStore {
  private readonly byId = new Map<string, NetworkEnemy>();

  constructor(private readonly scene: THREE.Scene) {}

  upsert(snap: EnemySnapshot): void {
    const existing = this.byId.get(snap.id);
    if (existing) {
      existing.applySnapshot(snap);
      return;
    }
    const enemy = new NetworkEnemy(snap);
    this.byId.set(snap.id, enemy);
    this.scene.add(enemy.group);
  }

  sync(list: EnemySnapshot[]): void {
    const seen = new Set<string>();
    for (const snap of list) {
      seen.add(snap.id);
      const existing = this.byId.get(snap.id);
      if (existing) existing.applySnapshot(snap);
      else {
        const enemy = new NetworkEnemy(snap);
        this.byId.set(snap.id, enemy);
        this.scene.add(enemy.group);
      }
    }
    for (const [id, enemy] of this.byId) {
      if (seen.has(id)) continue;
      this.byId.delete(id);
      enemy.dispose();
    }
  }

  get(id: string): NetworkEnemy | undefined {
    return this.byId.get(id);
  }

  list(): Combatant[] {
    return [...this.byId.values()];
  }

  applyHp(id: string, hp: number, maxHp: number, dead: boolean): void {
    this.byId.get(id)?.applyHp(hp, maxHp, dead);
  }

  flash(id: string): void {
    this.byId.get(id)?.flash();
  }

  update(dt: number, camera: THREE.Camera, targetId = ""): void {
    for (const enemy of this.byId.values()) {
      enemy.namedTarget = enemy.id === targetId;
      enemy.update(dt, camera);
    }
  }

  nearest(from: THREE.Vector3, maxDist: number): NetworkEnemy | null {
    let best: NetworkEnemy | null = null;
    let bestD = maxDist;
    for (const enemy of this.byId.values()) {
      if (enemy.health.isDead()) continue;
      const d = enemy.position.distanceTo(from);
      if (d <= bestD) {
        bestD = d;
        best = enemy;
      }
    }
    return best;
  }

  dispose(): void {
    for (const enemy of this.byId.values()) enemy.dispose();
    this.byId.clear();
  }
}

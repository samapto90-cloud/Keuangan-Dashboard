import * as THREE from "three";
import { PlayerFactory } from "./PlayerFactory";
import type { RemotePlayer } from "./RemotePlayer";
import type { PlayerSnapshot, PlayerSpawn } from "./NetworkMessage";

export class RemotePlayerStore {
  private readonly byId = new Map<string, RemotePlayer>();
  private readonly factory: PlayerFactory;

  constructor(scene: THREE.Scene) {
    this.factory = new PlayerFactory(scene);
  }

  spawn(spawn: PlayerSpawn, localId: string): void {
    if (!spawn.playerId || spawn.playerId === localId) return;
    const existing = this.byId.get(spawn.playerId);
    if (existing) {
      existing.applySpawn(spawn);
      return;
    }
    this.byId.set(spawn.playerId, this.factory.createRemotePlayer(spawn));
  }

  applySnapshot(snap: PlayerSnapshot, localId: string): void {
    if (!snap.id || snap.id === localId) return;
    const existing = this.byId.get(snap.id);
    if (existing) {
      existing.applySnapshot(snap);
      return;
    }
    this.spawn({
      playerId: snap.id,
      name: "Player",
      level: 1,
      class: "WARRIOR",
      hp: 100,
      maxHp: 100,
      x: snap.x,
      y: snap.y,
      z: snap.z,
      yaw: snap.yaw,
      state: snap.st,
    }, localId);
  }

  get(id: string): RemotePlayer | undefined {
    return this.byId.get(id);
  }

  playCombat(id: string, attackType: string, skillId?: string): void {
    this.byId.get(id)?.playCombat(attackType, skillId);
  }

  applyHp(id: string, hp: number, maxHp: number): void {
    this.byId.get(id)?.applyHp(hp, maxHp);
  }

  remove(id: string): void {
    const remote = this.byId.get(id);
    if (!remote) return;
    this.byId.delete(id);
    this.factory.destroyRemotePlayer(remote);
  }

  update(dt: number, camera: THREE.Camera): void {
    for (const remote of this.byId.values()) remote.update(dt, camera);
  }

  all(): RemotePlayer[] {
    return [...this.byId.values()];
  }

  count(): number {
    return this.byId.size;
  }

  meanOffset(): number {
    if (this.byId.size === 0) return 0;
    let sum = 0;
    for (const remote of this.byId.values()) sum += remote.lastOffset;
    return sum / this.byId.size;
  }

  dots(): Array<{ x: number; z: number; self: boolean }> {
    const out: Array<{ x: number; z: number; self: boolean }> = [];
    for (const remote of this.byId.values()) {
      out.push({ x: remote.group.position.x, z: remote.group.position.z, self: false });
    }
    return out;
  }

  dispose(): void {
    for (const remote of this.byId.values()) this.factory.destroyRemotePlayer(remote);
    this.byId.clear();
  }
}

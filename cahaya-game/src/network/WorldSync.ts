import type { Player } from "../player/Player";
import type { PlayerSnapshot, PlayerSpawn, WelcomeOut, WorldSnapshot } from "./NetworkMessage";
import { RemotePlayerStore } from "./RemotePlayerStore";
import type { EnemyStore } from "../combat/EnemyStore";

export class WorldSync {
  remotes: RemotePlayerStore;
  online = 0;
  worldId = "";
  channel = "";
  serverSelf: PlayerSnapshot | null = null;
  lastSnapshot: WorldSnapshot | null = null;

  constructor(
    remotes: RemotePlayerStore,
    private readonly enemies: EnemyStore,
  ) {
    this.remotes = remotes;
  }

  applyWelcome(w: WelcomeOut, local: Player): void {
    this.worldId = w.worldId;
    this.channel = w.channel;
    this.online = w.snapshot.online || 1 + w.players.length;
    local.position.set(w.self.x, w.self.y, w.self.z);
    local.mesh.rotation.y = w.self.yaw;
    local.facingYaw = w.self.yaw;
    local.velocity.set(0, 0, 0);
    local.stats.name = w.self.name;
    local.stats.level = w.self.level;
    local.className = w.self.class;
    this.remotes.dispose();
    for (const other of w.players) this.remotes.spawn(other, w.playerId);
    if (typeof w.self.energy === "number") local.stats.energy = w.self.energy;
    if (typeof w.self.exp === "number") local.stats.exp = w.self.exp;
    if (typeof w.self.expToNext === "number") local.stats.expToNext = w.self.expToNext;
    local.stats.health.hp = w.self.hp;
    local.stats.health.maxHp = w.self.maxHp;
    this.applySnapshot(w.snapshot, w.playerId);
  }

  spawn(spawn: PlayerSpawn, localId: string): void {
    this.remotes.spawn(spawn, localId);
  }

  despawn(id: string): void {
    this.remotes.remove(id);
  }

  applySnapshot(snap: WorldSnapshot, localId: string): void {
    this.lastSnapshot = snap;
    this.worldId = snap.worldId;
    this.channel = snap.channel;
    this.online = snap.online;
    this.serverSelf = snap.players.find((p) => p.id === localId) ?? this.serverSelf;
    for (const player of snap.players) this.remotes.applySnapshot(player, localId);
    this.enemies.sync(snap.enemies ?? []);
  }

  applyVitals(local: Player): void {
    const snap = this.serverSelf;
    if (!snap) return;
    if (typeof snap.hp === "number") local.stats.health.hp = snap.hp;
    if (typeof snap.maxHp === "number") local.stats.health.maxHp = snap.maxHp;
    if (typeof snap.energy === "number") local.stats.energy = snap.energy;
    if (typeof snap.maxEnergy === "number") local.stats.maxEnergy = snap.maxEnergy;
    if (typeof snap.stamina === "number") local.stats.stamina = snap.stamina;
    if (typeof snap.level === "number") local.stats.level = snap.level;
    if (typeof snap.exp === "number") local.stats.exp = snap.exp;
    if (typeof snap.expToNext === "number") local.stats.expToNext = snap.expToNext;
    if (snap.cs === "DEAD" || snap.cs === "DOWNED") local.combatState = snap.cs;
    else if (local.combatState === "HIT" || local.combatState === "STUNNED" || local.combatState === "DEAD" || local.combatState === "DOWNED") {
      local.combatState = snap.cs === "STUNNED" ? "STUNNED" : "IDLE";
    }
  }
}

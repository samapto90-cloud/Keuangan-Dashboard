import * as THREE from "three";
import { DEBUG_MODE } from "../game/GameConfig";
import type { Player } from "../player/Player";
import type { NetworkClient } from "../network/NetworkClient";
import type {
  AttackResult,
  DamageResult,
  DeathOut,
  EnemyDeathOut,
  EnemySnapshot,
  LevelUpOut,
  NetMsgType,
  RejectOut,
  RespawnOut,
} from "../network/NetworkMessage";
import type { EnemyStore } from "./EnemyStore";
import type { CombatEffects } from "./CombatEffects";
import type { CombatFeedback } from "./CombatFeedback";
import type { RemotePlayerStore } from "../network/RemotePlayerStore";
import { SKILLS } from "./SkillData";

export class CombatController {
  respawnAt = 0;
  levelUpUntil = 0;
  onPerfectDodge: (() => void) | null = null;
  readonly skillReadyAt: Record<string, number> = {};
  private respawnSent = false;
  private readonly hitPos = new THREE.Vector3();
  private readonly aim = new THREE.Vector3();

  constructor(
    private readonly player: Player,
    private readonly enemies: EnemyStore,
    private readonly remotes: RemotePlayerStore,
    private readonly effects: CombatEffects,
    private readonly feedback: CombatFeedback,
    private readonly canvas: HTMLCanvasElement,
    private readonly camera: THREE.Camera,
  ) {}

  sendAttack(net: NetworkClient, attackType: string, targetId: string, direction: number): void {
    net.sendCombat("PLAYER_ATTACK", {
      attackType,
      targetId,
      timestamp: Date.now(),
      direction,
    });
  }

  sendSkill(net: NetworkClient, skillId: string, targetId: string, direction: number): void {
    const def = SKILLS.find((s) => s.id === skillId);
    if (def) this.skillReadyAt[skillId] = performance.now() + def.cooldown * 1000;
    net.sendCombat("PLAYER_SKILL", {
      skillId,
      targetId,
      timestamp: Date.now(),
      direction,
    });
  }

  sendDodge(net: NetworkClient, ax: number, az: number, yaw: number): void {
    net.sendCombat("PLAYER_DODGE", { timestamp: Date.now(), ax, az, yaw });
  }

  sendCharge(net: NetworkClient, on: boolean): void {
    net.sendCombat("PLAYER_CHARGE", { on, timestamp: Date.now() });
  }

  sendCounter(net: NetworkClient): void {
    net.sendCombat("PLAYER_COUNTER", { timestamp: Date.now() });
  }

  sendRespawn(net: NetworkClient): void {
    if (this.respawnSent) return;
    this.respawnSent = true;
    net.sendCombat("PLAYER_RESPAWN", { timestamp: Date.now() });
  }

  handle(type: NetMsgType, data: unknown): void {
    switch (type) {
      case "ATTACK_RESULT":
        this.onAttack(data as AttackResult);
        break;
      case "DAMAGE_RESULT":
        this.onDamage(data as DamageResult, true);
        break;
      case "ENEMY_HIT":
      case "PLAYER_HIT":
        this.onDamage(data as DamageResult, false);
        break;
      case "ENEMY_DEATH":
        this.onEnemyDeath(data as EnemyDeathOut);
        break;
      case "ENEMY_SPAWN":
        this.enemies.upsert(data as EnemySnapshot);
        break;
      case "PLAYER_DEATH":
        this.onDeath(data as DeathOut);
        break;
      case "PLAYER_RESPAWN":
        this.onRespawn(data as RespawnOut);
        break;
      case "PLAYER_LEVEL_UP":
        this.onLevelUp(data as LevelUpOut);
        break;
      case "ACTION_REJECT":
        this.onReject(data as RejectOut);
        break;
      default:
        break;
    }
  }

  private onAttack(ev: AttackResult): void {
    if (ev.attackType === "perfect_dodge") {
      this.onPerfectDodge?.();
      this.feedback.setCombo(Math.max(2, this.feedback.comboCount), false);
      return;
    }
    if (ev.attackerId === this.player.id) {
      if (ev.comboHits) this.feedback.setCombo(ev.comboHits, ev.finisher === true);
      if (ev.skillId === "celestial_impact" || ev.skillId === "light_burst" || ev.skillId === "dawn_wave") {
        this.feedback.onHit(this.player.position, this.camera, this.canvas, 0, ev.skillId === "celestial_impact");
      }
      return;
    }
    this.remotes.playCombat(ev.attackerId, ev.attackType, ev.skillId);
    if (ev.skillId === "energy_bolt" || ev.skillId === "light_burst" || ev.attackType === "energy" || ev.skillId === "celestial_impact") {
      const remote = this.remotes.get(ev.attackerId);
      const origin = remote?.group.position ?? this.player.position;
      this.aim.set(Math.sin(remote?.mesh.rotation.y ?? 0), 0.05, Math.cos(remote?.mesh.rotation.y ?? 0));
      this.effects.bolt(origin.clone().setY(origin.y + 1.15), this.aim, ev.attackerId);
    }
  }

  private onDamage(ev: DamageResult, fx: boolean): void {
    this.hitPos.set(ev.hitX, ev.hitY, ev.hitZ);
    if (fx) {
      this.feedback.onHit(this.hitPos, this.camera, this.canvas, ev.damage, ev.isCritical, !!ev.blocked);
      this.effects.impact(this.hitPos);
    }
    if (ev.kind === "enemy") {
      this.enemies.applyHp(ev.targetId, ev.targetHp, ev.targetMaxHp, ev.killed);
      this.enemies.flash(ev.targetId);
      if (fx) this.feedback.setCombo(Math.min(10, this.feedback.comboCount + 1));
    } else if (ev.targetId === this.player.id) {
      this.player.stats.health.hp = ev.targetHp;
      this.player.stats.health.maxHp = ev.targetMaxHp;
      if (ev.killed) this.player.combatState = "DEAD";
      else this.player.combatState = "HIT";
    } else {
      this.remotes.applyHp(ev.targetId, ev.targetHp, ev.targetMaxHp);
    }
    if (DEBUG_MODE) {
      console.info(
        `[COMBAT] ${ev.attackerId} hit ${ev.targetId} Damage: ${ev.damage} Critical: ${ev.isCritical} Target HP: ${ev.targetHp}/${ev.targetMaxHp}`,
      );
    }
  }

  private onEnemyDeath(ev: EnemyDeathOut): void {
    const enemy = this.enemies.get(ev.enemyId);
    if (enemy) this.enemies.applyHp(ev.enemyId, 0, enemy.health.maxHp, true);
  }

  private onDeath(ev: DeathOut): void {
    if (ev.playerId === this.player.id) {
      this.player.combatState = "DEAD";
      this.player.stats.health.hp = 0;
      this.respawnAt = ev.respawnAt;
      this.respawnSent = false;
    } else {
      this.remotes.playCombat(ev.playerId, "dead");
    }
  }

  private onRespawn(ev: RespawnOut): void {
    if (ev.playerId === this.player.id) {
      this.player.position.set(ev.x, ev.y, ev.z);
      this.player.stats.health.hp = ev.hp;
      this.player.combatState = "IDLE";
      this.respawnAt = 0;
      this.respawnSent = false;
    }
  }

  private onLevelUp(ev: LevelUpOut): void {
    if (ev.playerId !== this.player.id) return;
    this.player.stats.level = ev.newLevel;
    this.player.stats.health.maxHp = ev.maxHp;
    this.player.stats.health.hp = ev.maxHp;
    this.levelUpUntil = performance.now() + 1800;
    this.feedback.showLevelUp(ev.fromLevel ?? ev.newLevel - 1, ev.newLevel, ev.attributePoints ?? 0, ev.skillPoints ?? 0);
  }

  private onReject(ev: RejectOut): void {
    if (ev.playerId && ev.playerId !== this.player.id) return;
    if (ev.action === "PLAYER_SKILL") {
      for (const skill of SKILLS) this.skillReadyAt[skill.id] = 0;
    }
    if (DEBUG_MODE) console.info(`[COMBAT] REJECT ${ev.action} ${ev.reason}`);
  }
}

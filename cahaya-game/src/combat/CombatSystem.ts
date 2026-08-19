import * as THREE from "three";
import { COMBAT } from "../game/GameConfig";
import type { Player } from "../player/Player";
import type { InputManager } from "../game/InputManager";
import type { CameraController } from "../game/CameraController";
import type { AttackKind } from "./CombatState";
import { ComboSystem } from "./ComboSystem";
import { AttackController } from "./AttackController";
import { DodgeSystem } from "./DodgeSystem";
import { EnergySystem } from "./EnergySystem";
import { ProjectilePool } from "./EnergyProjectile";
import { TargetSystem } from "./TargetSystem";
import { AudioManager } from "./AudioManager";
import { CombatFeedback } from "./CombatFeedback";
import { CombatEffects } from "./CombatEffects";
import { CombatController } from "./CombatController";
import { SkillBar } from "./SkillBar";
import { SKILLS } from "./SkillData";
import type { Combatant } from "./Combatant";
import type { HealthComponent } from "./HealthComponent";
import type { EnemyStore } from "./EnemyStore";
import type { NetworkClient } from "../network/NetworkClient";
import type { RemotePlayerStore } from "../network/RemotePlayerStore";

export interface CombatView {
  comboCount: number;
  target: { name: string; level: number; health: HealthComponent } | null;
  respawnAt: number;
  skillReadyAt: Record<string, number>;
}

export class CombatSystem {
  inputPaused = false;
  readonly combo = new ComboSystem();
  readonly dodge = new DodgeSystem();
  readonly energy = new EnergySystem();
  readonly attacks: AttackController;
  readonly targets = new TargetSystem();
  readonly audio = new AudioManager();
  readonly feedback: CombatFeedback;
  readonly effects: CombatEffects;
  readonly controller: CombatController;
  readonly skills: SkillBar;
  readonly view: CombatView = { comboCount: 0, target: null, respawnAt: 0, skillReadyAt: {} };
  pvpTargets: Combatant[] = [];
  charging = false;
  private hitStopRemain = 0;

  constructor(
    scene: THREE.Scene,
    hudRoot: HTMLElement,
    private readonly player: Player,
    private readonly enemies: EnemyStore,
    remotes: RemotePlayerStore,
    private readonly net: NetworkClient,
    private readonly input: InputManager,
    private readonly camera: CameraController,
    canvas: HTMLCanvasElement,
  ) {
    this.attacks = new AttackController(this.combo, this.energy);
    this.effects = new CombatEffects(scene, new ProjectilePool(scene));
    this.feedback = new CombatFeedback(hudRoot);
    this.skills = new SkillBar(hudRoot);
    this.controller = new CombatController(
      player,
      enemies,
      remotes,
      this.effects,
      this.feedback,
      canvas,
      camera.camera,
    );
    scene.add(this.targets.indicator);
    this.skills.onClick((skill) => this.castSkill(skill.id));
    this.controller.onPerfectDodge = () => this.notePerfectDodge();
  }

  consumeHitStop(): number {
    const v = this.hitStopRemain;
    this.hitStopRemain = 0;
    return v;
  }

  prepare(dt: number): void {
    this.combo.update(dt);
    this.handleInput();
    this.dodge.update(dt, this.player);
    this.attacks.update(dt, this.player);
    this.playStartSound(dt);
    this.energy.recover(this.player.stats, dt, this.attacks.attacking);
  }

  resolve(dt: number): void {
    this.effects.bolts.update(dt, [], () => undefined);
    this.effects.update(dt);
    this.targets.setCandidates(this.pvpTargets.length ? this.pvpTargets : this.enemies.list());
    this.targets.update(this.player.position);
    this.feedback.setCombo(this.combo.count);
    this.feedback.update(dt);
    this.skills.update(this.controller.skillReadyAt, performance.now());
    this.view.comboCount = this.combo.count;
    this.view.respawnAt = this.controller.respawnAt;
    this.view.skillReadyAt = this.controller.skillReadyAt;
    const locked = this.targets.current;
    if (locked && "name" in locked && "level" in locked && "health" in locked) {
      this.view.target = locked as CombatView["target"];
    } else {
      this.view.target = locked ? (this.enemies.get(locked.id) ?? null) : null;
    }
    this.camera.setLockTarget(this.targets.current?.position ?? null);
    this.camera.addShake(this.feedback.consumeShake());
    this.syncPlayerAnim();
    if (this.player.combatState === "DEAD" && this.controller.respawnAt && Date.now() >= this.controller.respawnAt) {
      this.controller.sendRespawn(this.net);
    }
  }

  dispose(): void {
    this.effects.dispose();
  }

  private handleInput(): void {
    if (this.inputPaused) return;
    if (this.input.consumeTarget()) this.targets.cycle(this.player.position);
    if (this.player.health.isDead()) return;
    if (this.input.consumeDodge()) {
      if (this.dodge.tryStart(this.player, this.input, this.camera.getYaw())) {
        this.attacks.cancel();
        this.audio.dodge();
        this.controller.sendDodge(this.net, this.input.compositeStrafe(), this.input.compositeForward(), this.camera.getYaw());
      }
    }
    if (this.dodge.active) return;
    this.syncCharge();
    if (this.input.consumePunch()) {
      if (this.input.blocking()) this.controller.sendCounter(this.net);
      else this.tryAttack("punch");
    }
    if (this.input.consumeKick()) this.tryAttack("kick");
    if (this.input.consumeSkill(1)) this.castSkill("power_strike");
    if (this.input.consumeSkill(2)) this.castSkill("flash_step");
    if (this.input.consumeSkill(3)) this.castSkill("dawn_wave");
    if (this.input.consumeSkill(4)) this.castSkill("light_burst");
    if (this.input.consumeUltimate()) {
      this.castSkill("celestial_impact");
      this.camera.pulseFov();
    }
    this.syncBlock();
  }

  private lastCharge = false;
  private syncCharge(): void {
    const on = this.input.charging();
    this.charging = on;
    this.effects.setCharge(on, this.player.position);
    if (on === this.lastCharge) return;
    this.lastCharge = on;
    if (on) this.audio.charge();
    this.controller.sendCharge(this.net, on);
    this.player.combatState = on ? "CHARGE" : this.player.combatState === "CHARGE" ? "IDLE" : this.player.combatState;
  }

  private lastBlock = false;
  private syncBlock(): void {
    const on = this.input.blocking();
    if (on === this.lastBlock) return;
    this.lastBlock = on;
    this.net.sendCombat("PLAYER_BLOCK", { on, timestamp: Date.now() });
  }

  private tryAttack(kind: AttackKind): void {
    this.ensureTarget(COMBAT.basicAttackRange);
    this.faceLockOrForward();
    this.attacks.request(kind, this.player);
    const targetId = this.targets.current && !this.targets.current.health.isDead() ? this.targets.current.id : "";
    this.controller.sendAttack(this.net, kind === "kick" ? "kick" : "punch", targetId, this.player.facingYaw);
  }

  private castSkill(skillId: string): void {
    const def = SKILLS.find((s) => s.id === skillId);
    if (!def) return;
    if (this.player.health.isDead()) return;
    this.ensureTarget(def.range || COMBAT.targetSearchRange);
    this.faceLockOrForward();
    if (skillId === "energy_bolt") this.attacks.request("energy", this.player);
    else if (skillId === "whirlwind_kick") this.attacks.request("kick", this.player);
    else this.attacks.request("punch", this.player);
    const targetId = this.targets.current && !this.targets.current.health.isDead() ? this.targets.current.id : "";
    this.controller.sendSkill(this.net, skillId, targetId, this.player.facingYaw);
    if (skillId === "energy_bolt" || skillId === "light_burst" || skillId === "celestial_impact" || skillId === "dawn_wave") {
      const dir = new THREE.Vector3();
      this.targets.aimDirection(this.player.position, this.player.facingYaw, dir);
      const origin = this.player.position.clone();
      origin.y += 1.15;
      origin.addScaledVector(dir, 0.7);
      this.effects.bolt(origin, dir, this.player.id);
      this.audio.energy();
    }
    if (skillId === "flash_step") this.audio.dodge();
  }

  private ensureTarget(range: number): void {
    if (!COMBAT.autoTarget) return;
    if (this.targets.current && !this.targets.current.health.isDead()) return;
    const near = this.enemies.nearest(this.player.position, Math.max(range, COMBAT.targetSearchRange));
    if (!near) return;
    const dx = near.position.x - this.player.position.x;
    const dz = near.position.z - this.player.position.z;
    const yaw = this.player.facingYaw;
    const dot = Math.sin(yaw) * dx + Math.cos(yaw) * dz;
    if (dot < 0) return;
    this.targets.current = near;
  }

  private playStartSound(dt: number): void {
    const atk = this.attacks.current;
    if (!atk || atk.elapsed > dt + 0.0001) return;
    if (atk.data.kind === "punch") this.audio.punch();
    else if (atk.data.kind === "kick") this.audio.kick();
    else if (atk.data.kind === "energy") this.audio.energy();
  }

  private faceLockOrForward(): void {
    const t = this.targets.current;
    if (!t || t.health.isDead()) return;
    const yaw = Math.atan2(t.position.x - this.player.position.x, t.position.z - this.player.position.z);
    this.player.facingYaw = yaw;
    this.player.mesh.rotation.y = yaw;
  }

  notePerfectDodge(): void {
    this.hitStopRemain = Math.max(this.hitStopRemain, 0.16);
    this.audio.perfect();
    this.feedback.setCombo(Math.max(2, this.combo.count), false);
  }

  noteHeavyHit(crit: boolean): void {
    this.hitStopRemain = Math.max(this.hitStopRemain, crit ? 0.08 : 0.05);
    if (crit) this.audio.crit();
    else this.audio.heavy();
  }

  private syncPlayerAnim(): void {
    const s = this.player.combatState;
    if (s === "DODGING") this.player.combatPose = "DODGE";
    else if (s === "ENERGY_ATTACK") this.player.combatPose = "ENERGY_ATTACK";
    else if (s === "ATTACKING" || s === "COMBO" || s === "CASTING") {
      const kind = this.attacks.current?.data.kind;
      this.player.combatPose = kind === "kick" ? "KICK" : kind === "energy" ? "ENERGY_ATTACK" : "PUNCH";
      if (s === "COMBO") this.player.combatPose = this.player.combatPose === "KICK" ? "KICK" : "COMBO";
    } else if (s === "HIT") this.player.combatPose = "HIT";
    else if (s === "STUNNED") this.player.combatPose = "STUN";
    else if (s === "DEAD") this.player.combatPose = "DEAD";
    else if (s === "GUARD") this.player.combatPose = null;
    else if (s === "CHARGE" || s === "TRANSFORM") this.player.combatPose = "ENERGY_ATTACK";
    else this.player.combatPose = null;
  }
}

import * as THREE from "three";
import { DamageNumber } from "./DamageNumber";

export class CombatFeedback {
  comboCount = 0;
  private shake = 0;
  private readonly numbers: DamageNumber;
  private readonly comboEl: HTMLElement;
  private readonly critEl: HTMLElement;
  private readonly levelEl: HTMLElement;

  constructor(hudRoot: HTMLElement) {
    const layer = document.createElement("div");
    layer.className = "dmg-layer";
    this.comboEl = document.createElement("div");
    this.comboEl.className = "hud-combo";
    this.comboEl.hidden = true;
    this.critEl = document.createElement("div");
    this.critEl.className = "hud-crit";
    this.critEl.hidden = true;
    this.critEl.textContent = "CRITICAL!";
    this.levelEl = document.createElement("div");
    this.levelEl.className = "hud-levelup";
    this.levelEl.hidden = true;
    this.levelEl.textContent = "LEVEL UP!";
    hudRoot.append(layer, this.comboEl, this.critEl, this.levelEl);
    this.numbers = new DamageNumber(layer);
  }

  numbersOn = true;
  shakeOn = true;

  onHit(world: THREE.Vector3, camera: THREE.Camera, canvas: HTMLCanvasElement, damage: number, crit: boolean, blocked = false): void {
    if (this.shakeOn) this.shake = Math.min(0.22, this.shake + (crit ? 0.12 : blocked ? 0.03 : 0.06));
    if (this.numbersOn) this.numbers.spawn(world, camera, canvas, damage, crit, blocked);
    if (crit) {
      this.critEl.hidden = false;
      window.setTimeout(() => {
        this.critEl.hidden = true;
      }, 420);
    }
  }

  showLevelUp(from = 0, to = 0, attr = 0, skill = 0): void {
    this.levelEl.hidden = false;
    this.levelEl.innerHTML = from && to
      ? `<div>LEVEL UP</div><div>LEVEL ${from} → LEVEL ${to}</div><div>Attribute Point +${attr} · Skill Point +${skill}</div>`
      : "LEVEL UP!";
    window.setTimeout(() => {
      this.levelEl.hidden = true;
    }, 1800);
  }

  setCombo(count: number, finisher = false): void {
    this.comboCount = Math.min(10, count);
    if (finisher) {
      this.comboEl.hidden = false;
      this.comboEl.textContent = "FINISH!";
    } else if (this.comboCount >= 2) {
      this.comboEl.hidden = false;
      this.comboEl.textContent = `COMBO x${this.comboCount}`;
    } else {
      this.comboEl.hidden = true;
    }
  }

  consumeShake(): number {
    const v = this.shake;
    this.shake *= 0.82;
    if (this.shake < 0.004) this.shake = 0;
    return v;
  }

  update(dt: number): void {
    this.numbers.update(dt);
  }
}

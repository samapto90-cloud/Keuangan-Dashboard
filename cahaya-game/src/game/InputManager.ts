export class InputManager {
  axisForward = 0;
  axisStrafe = 0;
  private jumpHeld = false;
  private jumpQueued = false;
  private sprintHeld = false;
  private pauseQueued = false;
  private formIndexQueued: number | null = null;
  private cycleFormQueued = false;
  private punchQueued = false;
  private kickQueued = false;
  private dodgeQueued = false;
  private energyQueued = false;
  private chargeHeld = false;
  private skill1Queued = false;
  private skill2Queued = false;
  private skill3Queued = false;
  private skill4Queued = false;
  private blockHeld = false;
  private ultimateQueued = false;
  private transformQueued = false;
  private potionQueued = false;
  private overlayCloseQueued = false;
  private inventoryQueued = false;
  private characterQueued = false;
  private partyQueued = false;
  private socialQueued = false;
  private guildQueued = false;
  private chatQueued = false;
  private interactQueued = false;
  private targetQueued = false;
  private mountQueued = false;
  private mountsQueued = false;
  private journalQueued = false;
  private mapQueued = false;
  private pvpQueued = false;
  private endgameQueued = false;
  decorateMode = false;
  private rotateQueued = false;
  private removeQueued = false;
  private readonly keys = new Set<string>();
  private readonly padPrev = new Set<number>();
  private padStick = false;

  constructor() {
    window.addEventListener("keydown", this.onKeyDown, { passive: false });
    window.addEventListener("keyup", this.onKeyUp);
    window.addEventListener("blur", this.clear);
    window.addEventListener("pointerdown", this.onPointerDown);
    window.addEventListener("contextmenu", this.onContextMenu);
  }

  dispose(): void {
    window.removeEventListener("keydown", this.onKeyDown);
    window.removeEventListener("keyup", this.onKeyUp);
    window.removeEventListener("blur", this.clear);
    window.removeEventListener("pointerdown", this.onPointerDown);
    window.removeEventListener("contextmenu", this.onContextMenu);
  }

  setVirtualAxes(strafe: number, forward: number): void {
    this.axisStrafe = clamp(strafe, -1, 1);
    this.axisForward = clamp(forward, -1, 1);
  }

  setVirtualSprint(on: boolean): void {
    this.sprintHeld = on;
  }

  queueVirtualJump(): void {
    this.jumpQueued = true;
  }

  queuePause(): void {
    this.pauseQueued = true;
  }

  queueFormIndex(index: number): void {
    this.formIndexQueued = index;
  }

  queueCycleForm(): void {
    this.cycleFormQueued = true;
  }

  consumeFormIndex(): number | null {
    const value = this.formIndexQueued;
    this.formIndexQueued = null;
    return value;
  }

  consumeCycleForm(): boolean {
    if (!this.cycleFormQueued) return false;
    this.cycleFormQueued = false;
    return true;
  }

  queuePunch(): void {
    this.punchQueued = true;
  }

  queueKick(): void {
    this.kickQueued = true;
  }

  queueDodge(): void {
    this.dodgeQueued = true;
  }

  queueEnergy(): void {
    this.energyQueued = true;
  }

  queueTarget(): void {
    this.targetQueued = true;
  }

  consumePunch(): boolean {
    if (!this.punchQueued) return false;
    this.punchQueued = false;
    return true;
  }

  consumeKick(): boolean {
    if (!this.kickQueued) return false;
    this.kickQueued = false;
    return true;
  }

  consumeDodge(): boolean {
    if (!this.dodgeQueued) return false;
    this.dodgeQueued = false;
    return true;
  }

  consumeEnergy(): boolean {
    if (!this.energyQueued) return false;
    this.energyQueued = false;
    return true;
  }

  charging(): boolean {
    return this.chargeHeld;
  }

  queueCharge(on: boolean): void {
    this.chargeHeld = on;
  }

  consumeSkill(slot: 1 | 2 | 3 | 4): boolean {
    if (slot === 1) {
      if (!this.skill1Queued) return false;
      this.skill1Queued = false;
      return true;
    }
    if (slot === 2) {
      if (!this.skill2Queued) return false;
      this.skill2Queued = false;
      return true;
    }
    if (slot === 3) {
      if (!this.skill3Queued) return false;
      this.skill3Queued = false;
      return true;
    }
    if (!this.skill4Queued) return false;
    this.skill4Queued = false;
    return true;
  }

  blocking(): boolean {
    return this.blockHeld;
  }

  queueBlock(on: boolean): void {
    this.blockHeld = on;
  }

  queueUltimate(): void {
    this.ultimateQueued = true;
  }

  consumeUltimate(): boolean {
    if (!this.ultimateQueued) return false;
    this.ultimateQueued = false;
    return true;
  }

  consumeDecorateRotate(): boolean {
    if (!this.rotateQueued) return false;
    this.rotateQueued = false;
    return true;
  }

  consumeDecorateRemove(): boolean {
    if (!this.removeQueued) return false;
    this.removeQueued = false;
    return true;
  }

  queueTransform(): void {
    this.transformQueued = true;
  }

  consumeTransform(): boolean {
    if (!this.transformQueued) return false;
    this.transformQueued = false;
    return true;
  }

  queuePotion(): void {
    this.potionQueued = true;
  }

  consumePotion(): boolean {
    if (!this.potionQueued) return false;
    this.potionQueued = false;
    return true;
  }

  queueInventory(): void {
    this.inventoryQueued = true;
  }

  queueCharacter(): void {
    this.characterQueued = true;
  }

  queueParty(): void {
    this.partyQueued = true;
  }

  queueSocial(): void {
    this.socialQueued = true;
  }

  consumeInventory(): boolean {
    if (!this.inventoryQueued) return false;
    this.inventoryQueued = false;
    return true;
  }

  consumeCharacter(): boolean {
    if (!this.characterQueued) return false;
    this.characterQueued = false;
    return true;
  }

  consumeParty(): boolean {
    if (!this.partyQueued) return false;
    this.partyQueued = false;
    return true;
  }

  consumeSocial(): boolean {
    if (!this.socialQueued) return false;
    this.socialQueued = false;
    return true;
  }

  queueGuild(): void {
    this.guildQueued = true;
  }

  consumeGuild(): boolean {
    if (!this.guildQueued) return false;
    this.guildQueued = false;
    return true;
  }

  queueChat(): void {
    this.chatQueued = true;
  }

  consumeChat(): boolean {
    if (!this.chatQueued) return false;
    this.chatQueued = false;
    return true;
  }

  queueMount(): void {
    this.mountQueued = true;
  }

  consumeMount(): boolean {
    if (!this.mountQueued) return false;
    this.mountQueued = false;
    return true;
  }

  consumeMounts(): boolean {
    if (!this.mountsQueued) return false;
    this.mountsQueued = false;
    return true;
  }

  consumeJournal(): boolean {
    if (!this.journalQueued) return false;
    this.journalQueued = false;
    return true;
  }

  consumeMap(): boolean {
    if (!this.mapQueued) return false;
    this.mapQueued = false;
    return true;
  }

  queuePvp(): void {
    this.pvpQueued = true;
  }

  consumePvp(): boolean {
    if (!this.pvpQueued) return false;
    this.pvpQueued = false;
    return true;
  }

  consumeEndgame(): boolean {
    if (!this.endgameQueued) return false;
    this.endgameQueued = false;
    return true;
  }

  consumeOverlayClose(): boolean {
    if (!this.overlayCloseQueued) return false;
    this.overlayCloseQueued = false;
    return true;
  }

  queueInteract(): void {
    this.interactQueued = true;
  }

  consumeInteract(): boolean {
    if (!this.interactQueued) return false;
    this.interactQueued = false;
    return true;
  }

  queueSkill(slot: 1 | 2 | 3 | 4): void {
    if (slot === 1) this.skill1Queued = true;
    else if (slot === 2) this.skill2Queued = true;
    else if (slot === 3) this.skill3Queued = true;
    else this.skill4Queued = true;
  }

  consumeTarget(): boolean {
    if (!this.targetQueued) return false;
    this.targetQueued = false;
    return true;
  }

  pollGamepad(): void {
    const pads = navigator.getGamepads?.() ?? [];
    const pad = pads.find((p) => p);
    if (!pad) {
      this.padPrev.clear();
      return;
    }
    const edge = (i: number, fn: () => void): void => {
      const down = pad.buttons[i]?.pressed === true;
      if (down && !this.padPrev.has(i)) fn();
      if (down) this.padPrev.add(i);
      else this.padPrev.delete(i);
    };
    edge(0, () => this.queuePunch());
    edge(1, () => this.queueKick());
    edge(2, () => this.queueEnergy());
    edge(4, () => this.queueDodge());
    edge(5, () => this.queueTarget());
    const ax = pad.axes[0] ?? 0;
    const ay = -(pad.axes[1] ?? 0);
    if (Math.hypot(ax, ay) > 0.25) {
      this.padStick = true;
      this.setVirtualAxes(ax, ay);
    } else if (this.padStick) {
      this.padStick = false;
      this.setVirtualAxes(0, 0);
    }
  }

  isMovingForward(): boolean {
    return this.compositeForward() > 0.25;
  }

  isMovingBackward(): boolean {
    return this.compositeForward() < -0.25;
  }

  isMovingLeft(): boolean {
    return this.compositeStrafe() < -0.25;
  }

  isMovingRight(): boolean {
    return this.compositeStrafe() > 0.25;
  }

  peekJumpQueued(): boolean {
    return this.jumpQueued;
  }

  isJumpPressed(): boolean {
    if (!this.jumpQueued) return false;
    this.jumpQueued = false;
    return true;
  }

  isSprintPressed(): boolean {
    return this.sprintHeld || this.keys.has("ShiftLeft") || this.keys.has("ShiftRight");
  }

  consumePause(): boolean {
    if (!this.pauseQueued) return false;
    this.pauseQueued = false;
    return true;
  }

  compositeForward(): number {
    const k = Number(this.keys.has("KeyW") || this.keys.has("ArrowUp")) -
      Number(this.keys.has("KeyS") || this.keys.has("ArrowDown"));
    const v = this.axisForward;
    const merged = Math.abs(v) > Math.abs(k) ? v : k;
    return clamp(merged, -1, 1);
  }

  compositeStrafe(): number {
    const k = Number(this.keys.has("KeyD") || this.keys.has("ArrowRight")) -
      Number(this.keys.has("KeyA") || this.keys.has("ArrowLeft"));
    const v = this.axisStrafe;
    const merged = Math.abs(v) > Math.abs(k) ? v : k;
    return clamp(merged, -1, 1);
  }

  private onKeyDown = (e: KeyboardEvent): void => {
    const target = e.target as HTMLElement | null;
    if (target?.closest("input, textarea, select, [contenteditable]")) {
      if (e.code === "Escape") this.overlayCloseQueued = true;
      return;
    }
    if (["Space", "ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight", "Tab"].includes(e.code)) {
      e.preventDefault();
    }
    if (e.repeat) return;
    this.keys.add(e.code);
    if (e.code === "Space" && !this.jumpHeld) {
      this.dodgeQueued = true;
    }
    if (e.code === "KeyF" && !this.jumpHeld) {
      this.jumpHeld = true;
      this.jumpQueued = true;
    }
    if (e.code === "Escape") {
      this.overlayCloseQueued = true;
      this.pauseQueued = true;
    }
    if (e.code === "KeyI") this.inventoryQueued = true;
    if (e.code === "KeyC") this.characterQueued = true;
    if (e.code === "KeyP") this.partyQueued = true;
    if (e.code === "KeyO") this.socialQueued = true;
    if (e.code === "KeyG") this.guildQueued = true;
    if (e.code === "Enter") this.chatQueued = true;
    if (e.code === "Digit1") this.skill1Queued = true;
    if (e.code === "Digit2") this.skill2Queued = true;
    if (e.code === "Digit3") this.skill3Queued = true;
    if (e.code === "Digit4") this.skill4Queued = true;
    if (e.code === "KeyJ") this.punchQueued = true;
    if (e.code === "KeyK") this.kickQueued = true;
    if (e.code === "KeyL") this.dodgeQueued = true;
    if (e.code === "KeyU") this.chargeHeld = true;
    if (e.code === "KeyQ") this.ultimateQueued = true;
    if (e.code === "KeyR") {
      if (this.decorateMode) this.rotateQueued = true;
      else this.ultimateQueued = true;
    }
    if (e.code === "Delete" || e.code === "Backspace") {
      if (this.decorateMode) this.removeQueued = true;
    }
    if (e.code === "ShiftLeft" || e.code === "ShiftRight") this.blockHeld = true;
    if (e.code === "KeyT") this.transformQueued = true;
    if (e.code === "KeyM") this.mountQueued = true;
    if (e.code === "KeyY") this.mountsQueued = true;
    if (e.code === "KeyN") this.journalQueued = true;
    if (e.code === "KeyB") this.mapQueued = true;
    if (e.code === "KeyV") this.pvpQueued = true;
    if (e.code === "KeyH") this.endgameQueued = true;
    if (e.code === "KeyE") this.interactQueued = true;
    if (e.code === "Tab") this.targetQueued = true;
  };

  private onKeyUp = (e: KeyboardEvent): void => {
    this.keys.delete(e.code);
    if (e.code === "Space" || e.code === "KeyF") this.jumpHeld = false;
    if (e.code === "ShiftLeft" || e.code === "ShiftRight") this.blockHeld = false;
    if (e.code === "KeyU") this.chargeHeld = false;
  };

  private onPointerDown = (e: PointerEvent): void => {
    const t = e.target as HTMLElement | null;
    if (t?.closest("button, input, textarea, .rpg-overlay, .dlg-overlay, .edu-overlay, .mobile-controls, .skill-bar, .hud-menu, .endgame-overlay, .pvp-overlay")) return;
    if (e.button === 0) this.punchQueued = true;
    if (e.button === 2) this.kickQueued = true;
  };

  private onContextMenu = (e: MouseEvent): void => {
    e.preventDefault();
  };

  private clear = (): void => {
    this.keys.clear();
    this.jumpHeld = false;
    this.axisForward = 0;
    this.axisStrafe = 0;
  };
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n));
}

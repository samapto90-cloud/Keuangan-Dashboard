import type { InputManager } from "../game/InputManager";

export class MobileControls {
  readonly root: HTMLElement;
  onAdventure: (() => void) | null = null;
  onMounts: (() => void) | null = null;
  onEndgame: (() => void) | null = null;
  onDungeon: (() => void) | null = null;
  onCraft: (() => void) | null = null;
  onLife: (() => void) | null = null;
  onHome: (() => void) | null = null;
  onPhoto: (() => void) | null = null;
  private readonly knob: HTMLElement;
  private readonly base: HTMLElement;
  private pointerId: number | null = null;
  private sprintOn = false;

  constructor(
    host: HTMLElement,
    private readonly input: InputManager,
  ) {
    this.root = document.createElement("div");
    this.root.className = "mobile-controls";
    this.root.innerHTML = `
      <div class="joy-base" id="joy-base"><div class="joy-knob" id="joy-knob"></div></div>
      <div class="mobile-combat">
        <button type="button" class="act-btn act-punch" id="btn-punch">ATTACK</button>
        <button type="button" class="act-btn act-kick" id="btn-heavy">KICK</button>
        <button type="button" class="act-btn act-skill" id="btn-skill1">SKILL 1</button>
        <button type="button" class="act-btn act-skill" id="btn-skill2">SKILL 2</button>
        <button type="button" class="act-btn act-skill" id="btn-skill3">SKILL 3</button>
        <button type="button" class="act-btn act-energy act-ult" id="btn-ult">ULTIMATE</button>
        <button type="button" class="act-btn act-dodge" id="btn-dodge">DODGE</button>
        <button type="button" class="act-btn act-block" id="btn-block">BLOCK</button>
        <button type="button" class="act-btn act-charge" id="btn-charge">CHARGE</button>
        <button type="button" class="act-btn act-form" id="btn-xform">ASCEND</button>
        <button type="button" class="act-btn" id="btn-potion">POTION</button>
        <button type="button" class="act-btn act-interact" id="btn-interact">INTERACT</button>
        <button type="button" class="act-btn act-target" id="btn-target">TARGET</button>
      </div>
      <div class="mobile-rpg">
        <button type="button" class="act-btn act-rpg" id="btn-inv">INV</button>
        <button type="button" class="act-btn act-rpg" id="btn-craft">CRAFT</button>
        <button type="button" class="act-btn act-rpg" id="btn-life">LIFE</button>
        <button type="button" class="act-btn act-rpg" id="btn-home">HOME</button>
        <button type="button" class="act-btn act-rpg" id="btn-photo">PHOTO</button>
        <button type="button" class="act-btn act-rpg" id="btn-char">CHAR</button>
        <button type="button" class="act-btn act-rpg" id="btn-party">PARTY</button>
        <button type="button" class="act-btn act-rpg" id="btn-social">SOCIAL</button>
        <button type="button" class="act-btn act-rpg" id="btn-guild">GUILD</button>
        <button type="button" class="act-btn act-rpg" id="btn-chat">CHAT</button>
        <button type="button" class="act-btn act-rpg" id="btn-mount">MOUNT</button>
        <button type="button" class="act-btn act-rpg" id="btn-mounts">MOUNTS</button>
        <button type="button" class="act-btn act-rpg" id="btn-pvp">PVP</button>
        <button type="button" class="act-btn act-rpg" id="btn-end">END</button>
        <button type="button" class="act-btn act-rpg" id="btn-dun">DUN</button>
        <button type="button" class="act-btn act-rpg" id="btn-adv">ADV</button>
      </div>
      <div class="mobile-actions">
        <button type="button" class="act-btn act-form" id="btn-form">Wujud</button>
        <button type="button" class="act-btn" id="btn-sprint">Lari</button>
        <button type="button" class="act-btn act-jump" id="btn-jump">Lompat</button>
        <button type="button" class="act-btn act-pause" id="btn-pause">Jeda</button>
      </div>
    `;
    host.appendChild(this.root);
    this.base = this.root.querySelector("#joy-base") as HTMLElement;
    this.knob = this.root.querySelector("#joy-knob") as HTMLElement;
    this.base.addEventListener("pointerdown", this.onDown);
    window.addEventListener("pointerup", this.onUp);
    window.addEventListener("pointercancel", this.onUp);
    window.addEventListener("pointermove", this.onMove);
    this.root.querySelector("#btn-jump")?.addEventListener("pointerdown", (e) => {
      e.preventDefault();
      this.input.queueVirtualJump();
    });
    const sprintBtn = this.root.querySelector("#btn-sprint");
    if (sprintBtn instanceof HTMLElement) {
      const releaseSprint = (): void => {
        this.sprintOn = false;
        this.input.setVirtualSprint(false);
        sprintBtn.classList.remove("on");
      };
      sprintBtn.addEventListener("pointerdown", (e) => {
        e.preventDefault();
        this.sprintOn = true;
        this.input.setVirtualSprint(true);
        sprintBtn.classList.add("on");
        sprintBtn.setPointerCapture(e.pointerId);
      });
      sprintBtn.addEventListener("pointerup", releaseSprint);
      sprintBtn.addEventListener("pointercancel", releaseSprint);
    }
    this.root.querySelector("#btn-pause")?.addEventListener("click", () => this.input.queuePause());
    this.root.querySelector("#btn-form")?.addEventListener("pointerdown", (e) => {
      e.preventDefault();
      this.input.queueTransform();
    });
    const tap = (id: string, fn: () => void): void => {
      this.root.querySelector(id)?.addEventListener("pointerdown", (e) => {
        e.preventDefault();
        fn();
      });
    };
    tap("#btn-punch", () => this.input.queuePunch());
    tap("#btn-heavy", () => this.input.queueKick());
    tap("#btn-skill1", () => this.input.queueSkill(1));
    tap("#btn-skill2", () => this.input.queueSkill(2));
    tap("#btn-skill3", () => this.input.queueSkill(3));
    tap("#btn-ult", () => this.input.queueUltimate());
    tap("#btn-xform", () => this.input.queueTransform());
    tap("#btn-potion", () => this.input.queuePotion());
    tap("#btn-interact", () => this.input.queueInteract());
    tap("#btn-dodge", () => this.input.queueDodge());
    tap("#btn-target", () => this.input.queueTarget());
    tap("#btn-inv", () => this.input.queueInventory());
    tap("#btn-craft", () => this.onCraft?.());
    tap("#btn-life", () => this.onLife?.());
    tap("#btn-home", () => this.onHome?.());
    tap("#btn-photo", () => this.onPhoto?.());
    tap("#btn-char", () => this.input.queueCharacter());
    tap("#btn-party", () => this.input.queueParty());
    tap("#btn-social", () => this.input.queueSocial());
    tap("#btn-guild", () => this.input.queueGuild());
    tap("#btn-chat", () => this.input.queueChat());
    tap("#btn-mount", () => this.input.queueMount());
    tap("#btn-mounts", () => this.onMounts?.());
    tap("#btn-pvp", () => this.input.queuePvp());
    tap("#btn-end", () => this.onEndgame?.());
    tap("#btn-dun", () => this.onDungeon?.());
    tap("#btn-adv", () => this.onAdventure?.());
    const blockBtn = this.root.querySelector("#btn-block");
    if (blockBtn instanceof HTMLElement) {
      blockBtn.addEventListener("pointerdown", (e) => {
        e.preventDefault();
        this.input.queueBlock(true);
        blockBtn.setPointerCapture(e.pointerId);
      });
      blockBtn.addEventListener("pointerup", () => this.input.queueBlock(false));
      blockBtn.addEventListener("pointercancel", () => this.input.queueBlock(false));
    }
    const chargeBtn = this.root.querySelector("#btn-charge");
    if (chargeBtn instanceof HTMLElement) {
      chargeBtn.addEventListener("pointerdown", (e) => {
        e.preventDefault();
        this.input.queueCharge(true);
        chargeBtn.setPointerCapture(e.pointerId);
      });
      chargeBtn.addEventListener("pointerup", () => this.input.queueCharge(false));
      chargeBtn.addEventListener("pointercancel", () => this.input.queueCharge(false));
    }
  }

  setVisible(on: boolean): void {
    this.root.hidden = !on;
    if (!on) {
      this.input.setVirtualAxes(0, 0);
      if (this.sprintOn) this.input.setVirtualSprint(false);
    }
  }

  dispose(): void {
    this.base.removeEventListener("pointerdown", this.onDown);
    window.removeEventListener("pointerup", this.onUp);
    window.removeEventListener("pointercancel", this.onUp);
    window.removeEventListener("pointermove", this.onMove);
    this.root.remove();
  }

  private onDown = (e: PointerEvent): void => {
    this.pointerId = e.pointerId;
    this.base.setPointerCapture(e.pointerId);
    this.apply(e);
  };

  private onUp = (e: PointerEvent): void => {
    if (this.pointerId !== null && e.pointerId !== this.pointerId && e.type !== "pointercancel") return;
    this.pointerId = null;
    this.knob.style.transform = "translate(-50%, -50%)";
    this.input.setVirtualAxes(0, 0);
  };

  private onMove = (e: PointerEvent): void => {
    if (this.pointerId !== e.pointerId) return;
    this.apply(e);
  };

  private apply(e: PointerEvent): void {
    const rect = this.base.getBoundingClientRect();
    const cx = rect.left + rect.width / 2;
    const cy = rect.top + rect.height / 2;
    let dx = e.clientX - cx;
    let dy = e.clientY - cy;
    const max = rect.width * 0.38;
    const mag = Math.hypot(dx, dy);
    if (mag > max) {
      dx = (dx / mag) * max;
      dy = (dy / mag) * max;
    }
    this.knob.style.transform = `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px))`;
    this.input.setVirtualAxes(dx / max, -dy / max);
  }
}

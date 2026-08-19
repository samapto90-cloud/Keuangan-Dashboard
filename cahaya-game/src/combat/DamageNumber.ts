import * as THREE from "three";

interface Floater {
  el: HTMLElement;
  life: number;
  active: boolean;
}

export class DamageNumber {
  private readonly pool: Floater[] = [];
  private readonly scratch = new THREE.Vector3();

  constructor(layer: HTMLElement, size = 12) {
    for (let i = 0; i < size; i++) {
      const el = document.createElement("span");
      el.className = "dmg-num";
      el.hidden = true;
      layer.appendChild(el);
      this.pool.push({ el, life: 0, active: false });
    }
  }

  spawn(world: THREE.Vector3, camera: THREE.Camera, canvas: HTMLCanvasElement, damage: number, crit: boolean, blocked = false): void {
    const f = this.pool.find((p) => !p.active) ?? this.pool[0];
    if (!f) return;
    f.active = true;
    f.life = 0.8;
    f.el.hidden = false;
    f.el.className = blocked ? "dmg-num blocked" : crit ? "dmg-num crit" : "dmg-num";
    f.el.textContent = blocked ? `BLOCK ${damage}` : crit ? `CRITICAL ${damage}` : String(damage);
    this.place(f.el, world, camera, canvas);
  }

  update(dt: number): void {
    for (const f of this.pool) {
      if (!f.active) continue;
      f.life -= dt;
      f.el.style.opacity = String(Math.max(0, f.life / 0.8));
      f.el.style.transform = `translate(-50%, ${-(0.8 - f.life) * 36}px)`;
      if (f.life <= 0) {
        f.active = false;
        f.el.hidden = true;
      }
    }
  }

  private place(el: HTMLElement, world: THREE.Vector3, camera: THREE.Camera, canvas: HTMLCanvasElement): void {
    this.scratch.copy(world).project(camera);
    const x = (this.scratch.x * 0.5 + 0.5) * canvas.clientWidth;
    const y = (-this.scratch.y * 0.5 + 0.5) * canvas.clientHeight;
    el.style.left = `${x}px`;
    el.style.top = `${y}px`;
  }
}

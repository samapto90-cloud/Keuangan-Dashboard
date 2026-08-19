import * as THREE from "three";

export type MountEntry = {
  mountId: string;
  name: string;
  type?: string;
  rarity?: string;
  speed?: number;
  source?: string;
  status?: string;
  owned?: boolean;
  favorite?: boolean;
  appearance?: string;
  terrain?: string;
};

export class MountCollection {
  readonly root: HTMLDivElement;
  blocking = false;
  mounts: MountEntry[] = [];
  active = "";
  private selected = "";
  private rot = 0.4;
  private zoom = 1;
  private renderer: THREE.WebGLRenderer | null = null;
  private scene: THREE.Scene | null = null;
  private camera: THREE.PerspectiveCamera | null = null;
  private mesh: THREE.Object3D | null = null;
  onSummon: ((id: string) => void) | null = null;
  onDismiss: (() => void) | null = null;
  onFavorite: ((id: string) => void) | null = null;
  onEquip: ((id: string) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay mount-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  apply(data: { mounts?: MountEntry[]; active?: string; mounted?: boolean; mountId?: string }): void {
    this.mounts = data.mounts ?? [];
    this.active = data.active || data.mountId || "";
    if (!this.selected) this.selected = this.active || this.mounts.find((m) => m.owned)?.mountId || this.mounts[0]?.mountId || "";
    if (this.blocking) this.render(!!data.mounted);
  }

  toggle(mounted: boolean): void {
    if (this.blocking) this.close();
    else this.open(mounted);
  }

  open(mounted: boolean): void {
    this.blocking = true;
    this.root.hidden = false;
    this.render(mounted);
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
    this.disposePreview();
  }

  private render(mounted: boolean): void {
    const cur = this.mounts.find((m) => m.mountId === this.selected);
    const owned = !!cur?.owned;
    const name = cur?.name || "???";
    const rows = this.mounts.map((m) => `<button type="button" class="mount-row${m.mountId === this.selected ? " on" : ""}" data-mid="${m.mountId}">${m.favorite ? "★ " : ""}${m.name} · ${m.status || ""}</button>`).join("");
    this.root.innerHTML = `
      <div class="rpg-card mount-card">
        <header><h3>MOUNTS</h3><button type="button" class="mount-close">Tutup</button></header>
        <div class="mount-layout">
          <canvas class="mount-preview" width="280" height="220"></canvas>
          <div class="mount-list">${rows || "<p>Belum ada tunggangan.</p>"}</div>
          <div class="mount-info">
            <h4>${name}</h4>
            <p>Type: ${cur?.type || "-"}</p>
            <p>Speed: ${cur?.speed != null ? cur.speed.toFixed(2) : "-"}</p>
            <p>Source: ${cur?.source || "-"}</p>
            <p>Status: ${cur?.status || "LOCKED"}</p>
            <p>Rarity: ${cur?.rarity || "-"}</p>
            <p>Terrain: ${cur?.terrain || "open"}</p>
            <div class="inv-actions">
              <button type="button" class="mount-rot">Rotate</button>
              <button type="button" class="mount-zoom">Zoom</button>
              <button type="button" class="mount-fav" ${owned ? "" : "disabled"}>Favorite</button>
              <button type="button" class="mount-eq" ${owned ? "" : "disabled"}>Equip</button>
              <button type="button" class="mount-call" ${owned ? "" : "disabled"}>${mounted ? "DISMISS" : "CALL MOUNT"}</button>
            </div>
          </div>
        </div>
      </div>`;
    this.root.querySelector(".mount-close")?.addEventListener("click", () => this.close());
    this.root.querySelectorAll("[data-mid]").forEach((el) => {
      el.addEventListener("click", () => {
        this.selected = (el as HTMLElement).dataset.mid || "";
        this.render(mounted);
      });
    });
    this.root.querySelector(".mount-rot")?.addEventListener("click", () => {
      this.rot += 0.6;
      this.drawPreview();
    });
    this.root.querySelector(".mount-zoom")?.addEventListener("click", () => {
      this.zoom = this.zoom > 1.2 ? 0.9 : 1.35;
      this.drawPreview();
    });
    this.root.querySelector(".mount-fav")?.addEventListener("click", () => {
      if (this.selected) this.onFavorite?.(this.selected);
    });
    this.root.querySelector(".mount-eq")?.addEventListener("click", () => {
      if (this.selected) this.onEquip?.(this.selected);
    });
    this.root.querySelector(".mount-call")?.addEventListener("click", () => {
      if (mounted) this.onDismiss?.();
      else if (this.selected) this.onSummon?.(this.selected);
    });
    this.setupPreview();
  }

  private setupPreview(): void {
    const canvas = this.root.querySelector(".mount-preview") as HTMLCanvasElement | null;
    if (!canvas) return;
    this.disposePreview();
    this.renderer = new THREE.WebGLRenderer({ canvas, antialias: true, alpha: true });
    this.renderer.setSize(280, 220, false);
    this.scene = new THREE.Scene();
    this.camera = new THREE.PerspectiveCamera(42, 280 / 220, 0.1, 20);
    this.camera.position.set(0, 1.4, 3.4);
    this.scene.add(new THREE.AmbientLight(0xffffff, 0.8));
    const sun = new THREE.DirectionalLight(0xfff1c1, 0.7);
    sun.position.set(2, 4, 3);
    this.scene.add(sun);
    this.drawPreview();
  }

  private drawPreview(): void {
    if (!this.scene || !this.camera || !this.renderer) return;
    if (this.mesh) this.scene.remove(this.mesh);
    const cur = this.mounts.find((m) => m.mountId === this.selected);
    this.mesh = buildMountMesh(cur?.owned ? cur.mountId : "locked");
    this.mesh.rotation.y = this.rot;
    this.mesh.scale.setScalar(this.zoom);
    this.scene.add(this.mesh);
    this.camera.position.z = 3.4 / this.zoom;
    this.renderer.render(this.scene, this.camera);
  }

  private disposePreview(): void {
    this.renderer?.dispose();
    this.renderer = null;
    this.scene = null;
    this.camera = null;
    this.mesh = null;
  }
}

export function buildMountMesh(id: string): THREE.Group {
  const g = new THREE.Group();
  g.name = id;
  const color = mountColor(id);
  const mat = new THREE.MeshStandardMaterial({ color, roughness: 0.55 });
  const body = new THREE.Mesh(new THREE.CapsuleGeometry(0.38, 1.05, 4, 8), mat);
  body.rotation.z = Math.PI / 2;
  body.position.set(0, 0.55, 0);
  const head = new THREE.Mesh(new THREE.SphereGeometry(id === "celestial-bird" ? 0.22 : 0.28, 8, 8), mat);
  head.position.set(0, 0.85, 0.7);
  if (id === "celestial-bird") {
    const wing = new THREE.Mesh(new THREE.BoxGeometry(1.6, 0.08, 0.4), mat);
    wing.position.set(0, 0.7, 0);
    g.add(wing);
  }
  const seat = new THREE.Object3D();
  seat.name = "seat";
  seat.position.set(0, 0.95, 0);
  g.add(body, head, seat);
  return g;
}

function mountColor(id: string): number {
  switch (id) {
    case "dawn-wolf": return 0xd4c4a8;
    case "sky-fox": return 0xc9a27a;
    case "forest-deer": return 0x8b6b45;
    case "stone-beast": return 0x8a8f99;
    case "celestial-bird": return 0xe8d48a;
    case "wind-runner": return 0xc9d6e4;
    default: return 0x3a4250;
  }
}

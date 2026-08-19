import * as THREE from "three";
import { buildMountMesh } from "./MountCollection";

export class MountSystem {
  mounted = false;
  mountId = "";
  readonly mesh: THREE.Group;
  private body: THREE.Group;
  private vfx: THREE.Mesh | null = null;
  lowFx = false;

  constructor(playerGroup: THREE.Group, private readonly playerMesh: THREE.Group) {
    this.mesh = new THREE.Group();
    this.mesh.name = "mount-root";
    this.body = buildMountMesh("wind-runner");
    this.mesh.add(this.body);
    this.mesh.visible = false;
    playerGroup.add(this.mesh);
  }

  apply(mounted: boolean, mountId: string, graphics: "LOW" | "MEDIUM" | "HIGH" | "ULTRA" = "HIGH"): void {
    this.lowFx = graphics === "LOW";
    const id = mountId || "wind-runner";
    if (id !== this.mountId) {
      this.mesh.remove(this.body);
      this.body = buildMountMesh(id);
      this.mesh.add(this.body);
    }
    this.mounted = mounted;
    this.mountId = mountId;
    this.mesh.visible = mounted && !!mountId;
    this.playerMesh.position.y = mounted ? 0.72 : 0;
    if (mounted && !this.lowFx) this.flash(0x9fd6ff);
    if (!mounted && mountId && !this.lowFx) this.flash(0xdbeafe);
  }

  speed(): number {
    if (!this.mounted) return 1;
    switch (this.mountId) {
      case "dawn-wolf": return 1.42;
      case "sky-fox": return 1.48;
      case "forest-deer": return 1.4;
      case "stone-beast": return 1.38;
      case "celestial-bird": return 1.5;
      default: return 1.45;
    }
  }

  applyRemote(group: THREE.Group, mesh: THREE.Group, mounted: boolean, mountId = "wind-runner"): void {
    const name = mountId || "wind-runner";
    let mount = group.getObjectByName(name) as THREE.Group | null;
    if (mounted) {
      if (!mount) {
        mount = buildMountMesh(name);
        group.add(mount);
      }
      mount.visible = true;
      mesh.position.y = 0.72;
    } else if (mount) {
      mount.visible = false;
      mesh.position.y = 0;
    }
  }

  private flash(color: number): void {
    if (!this.mesh.parent) return;
    if (this.vfx) this.mesh.parent.remove(this.vfx);
    const m = new THREE.Mesh(
      new THREE.RingGeometry(0.4, 0.7, 16),
      new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.45, side: THREE.DoubleSide }),
    );
    m.rotation.x = -Math.PI / 2;
    m.position.y = 0.05;
    this.vfx = m;
    this.mesh.parent.add(m);
    window.setTimeout(() => {
      if (this.vfx?.parent) this.vfx.parent.remove(this.vfx);
      this.vfx = null;
    }, 280);
  }
}

import * as THREE from "three";

const COLORS: Record<string, number> = {
  "dawn-pup": 0xfbbf24,
  "forest-fox": 0xc2410c,
  "little-turtle": 0x4ade80,
  "sky-bird": 0x7dd3fc,
  "stone-cub": 0x94a3b8,
};

export class PetFollower {
  readonly mesh = new THREE.Group();
  petId = "";
  lowFx = false;
  private bob = 0;

  constructor(parent: THREE.Object3D) {
    parent.add(this.mesh);
    this.mesh.visible = false;
  }

  apply(petId: string, graphics: "LOW" | "MEDIUM" | "HIGH" | "ULTRA" = "HIGH"): void {
    this.lowFx = graphics === "LOW";
    if (!petId) {
      this.petId = "";
      this.mesh.visible = false;
      return;
    }
    if (petId !== this.petId) {
      this.rebuild(petId);
      this.petId = petId;
    }
    this.mesh.visible = true;
  }

  update(dt: number): void {
    if (!this.mesh.visible) return;
    this.bob += dt;
    const idle = Math.sin(this.bob * 2.2) * 0.04;
    this.mesh.position.set(-0.55, 0.35 + idle, -0.35);
    this.mesh.rotation.y = Math.sin(this.bob * 1.4) * 0.15;
  }

  private rebuild(id: string): void {
    this.mesh.clear();
    const color = COLORS[id] ?? 0xfde68a;
    const body = new THREE.Mesh(
      new THREE.SphereGeometry(0.18, this.lowFx ? 6 : 10, this.lowFx ? 6 : 10),
      new THREE.MeshStandardMaterial({ color, emissive: this.lowFx ? 0x000000 : color, emissiveIntensity: this.lowFx ? 0 : 0.12 }),
    );
    this.mesh.add(body);
    if (!this.lowFx) {
      const scarf = new THREE.Mesh(
        new THREE.TorusGeometry(0.14, 0.03, 6, 10),
        new THREE.MeshStandardMaterial({ color: 0xf43f5e }),
      );
      scarf.rotation.x = Math.PI / 2;
      scarf.position.y = 0.02;
      this.mesh.add(scarf);
    }
  }
}

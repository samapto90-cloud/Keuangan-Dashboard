import * as THREE from "three";

export class TrainingGround {
  readonly root = new THREE.Group();

  constructor(parent: THREE.Object3D) {
    parent.add(this.root);
    this.buildPad();
  }

  private buildPad(): void {
    const pad = new THREE.Mesh(
      new THREE.CircleGeometry(7.5, 32),
      new THREE.MeshStandardMaterial({ color: 0xb89a6a, roughness: 0.92 }),
    );
    pad.rotation.x = -Math.PI / 2;
    pad.position.set(0, 0.03, 9);
    pad.receiveShadow = true;
    const ring = new THREE.Mesh(
      new THREE.TorusGeometry(7.4, 0.06, 6, 40),
      new THREE.MeshStandardMaterial({ color: 0xe8b86d, roughness: 0.45, metalness: 0.2 }),
    );
    ring.rotation.x = Math.PI / 2;
    ring.position.set(0, 0.06, 9);
    const sign = new THREE.Mesh(
      new THREE.PlaneGeometry(3.4, 0.7),
      new THREE.MeshBasicMaterial({ map: makeSign(), transparent: true }),
    );
    sign.position.set(0, 2.2, 13.2);
    this.root.add(pad, ring, sign);
  }
}

function makeSign(): THREE.CanvasTexture {
  const c = document.createElement("canvas");
  c.width = 512;
  c.height = 128;
  const ctx = c.getContext("2d");
  if (ctx) {
    ctx.fillStyle = "rgba(7,17,31,0.72)";
    ctx.fillRect(0, 0, 512, 128);
    ctx.strokeStyle = "#e8b86d";
    ctx.lineWidth = 8;
    ctx.strokeRect(8, 8, 496, 112);
    ctx.fillStyle = "#f4f8fc";
    ctx.font = "bold 36px Segoe UI, sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText("TRAINING GROUND", 256, 64);
  }
  const tex = new THREE.CanvasTexture(c);
  tex.colorSpace = THREE.SRGBColorSpace;
  return tex;
}

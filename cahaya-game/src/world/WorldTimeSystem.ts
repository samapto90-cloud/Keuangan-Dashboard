import * as THREE from "three";

export type TimePhase = "DAY" | "EVENING" | "NIGHT";

/** Configurable day/night architecture. Full cycle is later. */
export class WorldTimeSystem {
  phase: TimePhase = "DAY";

  apply(scene: THREE.Scene, sun: THREE.DirectionalLight, hemi: THREE.HemisphereLight): void {
    if (this.phase === "NIGHT") {
      scene.background = new THREE.Color(0x1a2740);
      scene.fog = new THREE.Fog(0x1a2740, 22, 110);
      sun.intensity = 0.28;
      sun.color.setHex(0x9bb7ff);
      hemi.intensity = 0.45;
      return;
    }
    if (this.phase === "EVENING") {
      scene.background = new THREE.Color(0xf0b27a);
      scene.fog = new THREE.Fog(0xd9a066, 24, 110);
      sun.intensity = 0.85;
      sun.color.setHex(0xffc58a);
      hemi.intensity = 0.7;
      return;
    }
    scene.background = new THREE.Color(0xb8e0ef);
    scene.fog = new THREE.Fog(0x8fb4c8, 28, 120);
    sun.intensity = 1.35;
    sun.color.setHex(0xfff1d0);
    hemi.intensity = 1.05;
  }
}

import * as THREE from "three";

export class WeatherSystem {
  private rain: THREE.Points | null = null;
  private leaves: THREE.Points | null = null;
  weather = "CLEAR";

  apply(scene: THREE.Scene, weather: string): void {
    this.weather = weather || "CLEAR";
    if (this.rain) {
      scene.remove(this.rain);
      this.rain = null;
    }
    if (!this.leaves) {
      this.leaves = makeLeaves();
      scene.add(this.leaves);
    }
    const fog = scene.fog;
    if (!(fog instanceof THREE.Fog)) return;
    if (this.weather === "FOG") {
      fog.near = 16;
      fog.far = 52;
    } else if (this.weather === "RAIN" || this.weather === "STORM") {
      fog.near = 20;
      fog.far = 70;
      this.rain = makeRain();
      scene.add(this.rain);
    } else if (this.weather === "SNOW") {
      fog.near = 22;
      fog.far = 74;
    } else if (this.weather === "CLOUDY") {
      fog.near = 24;
      fog.far = 82;
    } else {
      fog.near = 28;
      fog.far = 96;
    }
  }

  update(dt: number): void {
    if (this.rain) {
      this.rain.position.y -= dt * 6;
      if (this.rain.position.y < -4) this.rain.position.y = 4;
    }
    if (this.leaves) {
      this.leaves.rotation.y += dt * 0.08;
      this.leaves.position.y = Math.sin(performance.now() * 0.0004) * 0.4;
    }
  }
}

function makeLeaves(): THREE.Points {
  const geo = new THREE.BufferGeometry();
  const n = 48;
  const pos = new Float32Array(n * 3);
  for (let i = 0; i < n; i++) {
    pos[i * 3] = (Math.random() - 0.5) * 36;
    pos[i * 3 + 1] = 1.2 + Math.random() * 4;
    pos[i * 3 + 2] = 8 + Math.random() * 28;
  }
  geo.setAttribute("position", new THREE.BufferAttribute(pos, 3));
  return new THREE.Points(geo, new THREE.PointsMaterial({ color: 0xc4a35a, size: 0.12, transparent: true, opacity: 0.55, depthWrite: false }));
}

function makeRain(): THREE.Points {
  const geo = new THREE.BufferGeometry();
  const n = 80;
  const pos = new Float32Array(n * 3);
  for (let i = 0; i < n; i++) {
    pos[i * 3] = (Math.random() - 0.5) * 40;
    pos[i * 3 + 1] = Math.random() * 10;
    pos[i * 3 + 2] = (Math.random() - 0.5) * 40;
  }
  geo.setAttribute("position", new THREE.BufferAttribute(pos, 3));
  return new THREE.Points(geo, new THREE.PointsMaterial({ color: 0xc9e4ff, size: 0.08, transparent: true, opacity: 0.45 }));
}

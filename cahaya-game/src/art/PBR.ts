import * as THREE from "three";

const cache = new Map<string, THREE.CanvasTexture>();

export function noiseMap(key: string, a: string, b: string, grain = 48): THREE.CanvasTexture {
  const hit = cache.get(key);
  if (hit) return hit;
  const canvas = document.createElement("canvas");
  canvas.width = 128;
  canvas.height = 128;
  const ctx = canvas.getContext("2d");
  if (ctx) {
    ctx.fillStyle = a;
    ctx.fillRect(0, 0, 128, 128);
    for (let i = 0; i < grain; i++) {
      ctx.fillStyle = b;
      ctx.globalAlpha = 0.08 + (i % 5) * 0.02;
      ctx.fillRect((i * 17) % 128, (i * 29) % 128, 6 + (i % 8), 4 + (i % 6));
    }
    ctx.globalAlpha = 1;
  }
  const map = new THREE.CanvasTexture(canvas);
  map.wrapS = map.wrapT = THREE.RepeatWrapping;
  map.colorSpace = THREE.SRGBColorSpace;
  cache.set(key, map);
  return map;
}

export function std(
  color: number,
  roughness: number,
  metalness = 0.04,
  map?: THREE.Texture,
  emissive = 0x000000,
  emissiveIntensity = 0,
): THREE.MeshStandardMaterial {
  return new THREE.MeshStandardMaterial({
    color,
    roughness,
    metalness,
    map,
    emissive,
    emissiveIntensity,
  });
}

export function skinMat(hex = 0xe2b089): THREE.MeshStandardMaterial {
  return std(hex, 0.58, 0.0, noiseMap("skin", "#e2b089", "#c48962", 36));
}

export function clothMat(hex: number, key: string): THREE.MeshStandardMaterial {
  return std(hex, 0.78, 0.02, noiseMap(key, "#8899aa", "#334455", 64));
}

export function leatherMat(hex = 0x5c3a24): THREE.MeshStandardMaterial {
  return std(hex, 0.72, 0.08, noiseMap("leather", "#5c3a24", "#2a1810", 40));
}

export function metalMat(hex = 0xc9c4b8, rough = 0.32): THREE.MeshStandardMaterial {
  return std(hex, rough, 0.82, noiseMap("metal", "#d8d4cc", "#8a8680", 28));
}

export function woodMat(hex = 0x7a4e2d): THREE.MeshStandardMaterial {
  return std(hex, 0.9, 0.02, noiseMap("wood", "#7a4e2d", "#3e2414", 40));
}

export function stoneMat(hex = 0x8a8f86): THREE.MeshStandardMaterial {
  return std(hex, 0.94, 0.04, noiseMap("stone", "#8a8f86", "#5c615c", 50));
}

export function hash32(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) h = Math.imul(h ^ s.charCodeAt(i), 16777619);
  return h >>> 0;
}

import { PLAYER } from "../game/GameConfig";

/** Collision layers — mirror go-app/mmo/world_colliders.go */
export const COLLISION_LAYERS = {
  PLAYER: "PLAYER",
  NPC: "NPC",
  ENEMY: "ENEMY",
  BUILDING: "BUILDING",
  WORLD: "WORLD",
  OBSTACLE: "OBSTACLE",
  PROJECTILE: "PROJECTILE",
  TRIGGER: "TRIGGER",
} as const;

export type ColliderKind =
  | "ground"
  | "rock"
  | "house"
  | "tree"
  | "wall"
  | "fence"
  | "water"
  | "bridge"
  | "npc"
  | "enemy"
  | "player";

export interface AabbCollider {
  kind: ColliderKind;
  layer: string;
  x: number;
  z: number;
  w: number;
  d: number;
}

export interface CircleCollider {
  kind: ColliderKind;
  layer: string;
  x: number;
  z: number;
  r: number;
  walkable?: boolean;
}

/** Dawn Village static colliders — synced with Environment.ts / server world_colliders.go */
export const VILLAGE_AABBS: AabbCollider[] = [
  { kind: "house", layer: COLLISION_LAYERS.BUILDING, x: -11.5, z: 8.5, w: 3.4, d: 2.8 },
  { kind: "house", layer: COLLISION_LAYERS.BUILDING, x: -13.2, z: 5.2, w: 3.4, d: 2.8 },
  { kind: "house", layer: COLLISION_LAYERS.BUILDING, x: -10.4, z: 11.4, w: 3.4, d: 2.8 },
  { kind: "house", layer: COLLISION_LAYERS.BUILDING, x: -6.6, z: 0.6, w: 3.8, d: 3.0 },
  { kind: "house", layer: COLLISION_LAYERS.BUILDING, x: 7.4, z: 8.2, w: 3.4, d: 2.8 },
  { kind: "house", layer: COLLISION_LAYERS.BUILDING, x: 11.2, z: 4.8, w: 2.4, d: 1.6 },
  { kind: "wall", layer: COLLISION_LAYERS.OBSTACLE, x: -9.2, z: 1.1, w: 1.8, d: 0.9 },
  { kind: "wall", layer: COLLISION_LAYERS.OBSTACLE, x: -10.6, z: 2.6, w: 1.5, d: 0.4 },
  { kind: "wall", layer: COLLISION_LAYERS.OBSTACLE, x: -8.1, z: 0.4, w: 1.9, d: 0.2 },
  { kind: "water", layer: COLLISION_LAYERS.WORLD, x: -12.0, z: 14.5, w: 14.0, d: 3.4 },
];

export const VILLAGE_CIRCLES: CircleCollider[] = [
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: -1.1, z: 4.6, r: 0.85 },
  { kind: "bridge", layer: COLLISION_LAYERS.WORLD, x: -12.0, z: 14.5, r: 1.3, walkable: true },
];

export const VILLAGE_TREES: CircleCollider[] = villageTreeCircles();

function villageTreeCircles(): CircleCollider[] {
  const out: CircleCollider[] = [];
  for (let i = 0; i < 22; i++) {
    const ang = (i / 22) * Math.PI * 2;
    const dist = 16 + (i % 5) * 2.2;
    out.push({
      kind: "tree",
      layer: COLLISION_LAYERS.OBSTACLE,
      x: Math.cos(ang) * dist,
      z: Math.sin(ang) * dist * 0.55 + 2,
      r: 0.9,
    });
  }
  for (let i = 0; i < 10; i++) {
    out.push({
      kind: "tree",
      layer: COLLISION_LAYERS.OBSTACLE,
      x: (i - 4.5) * 2.1,
      z: 27 + (i % 3) * 1.8,
      r: 0.85,
    });
  }
  return out;
}

export const VILLAGE_ROCKS: CircleCollider[] = [
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: -7.0, z: 12.5, r: 0.38 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: -5.4, z: 13.5, r: 0.42 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: -3.8, z: 12.5, r: 0.35 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: -2.2, z: 13.5, r: 0.4 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: -0.6, z: 12.5, r: 0.36 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: 1.0, z: 13.5, r: 0.41 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: 2.6, z: 12.5, r: 0.39 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: 4.2, z: 13.5, r: 0.37 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: 5.8, z: 12.5, r: 0.4 },
  { kind: "rock", layer: COLLISION_LAYERS.OBSTACLE, x: 7.4, z: 13.5, r: 0.38 },
];

export const VILLAGE_FENCE_Z = 20.4;

export const NPC_PERSONAL_SPACE = 0.78;
export const NPC_AVOIDANCE_RADIUS = 1.15;

export function onBridge(x: number, z: number): boolean {
  return Math.abs(x + 12) < 1.3 && Math.abs(z - 14.5) < 2.2;
}

export function hitVillageFence(x: number, z: number, radius: number): boolean {
  if (Math.abs(z - VILLAGE_FENCE_Z) > radius + 0.12) return false;
  if (Math.abs(x) < 1.45) return false;
  if (Math.abs(x) > 6.2) return false;
  return true;
}

function pushVillageFence(x: number, z: number, radius: number): [number, number] {
  if (!hitVillageFence(x, z, radius)) return [x, z];
  if (z >= VILLAGE_FENCE_Z) z = VILLAGE_FENCE_Z + radius + 0.12;
  else z = VILLAGE_FENCE_Z - radius - 0.12;
  return [x, z];
}

function insideAabb(x: number, z: number, pad: number, b: AabbCollider): boolean {
  return Math.abs(x - b.x) < b.w / 2 + pad && Math.abs(z - b.z) < b.d / 2 + pad;
}

function pushOutAabb(x: number, z: number, pad: number, b: AabbCollider): [number, number] {
  const hwx = b.w / 2 + pad;
  const hdz = b.d / 2 + pad;
  const dx = x - b.x;
  const dz = z - b.z;
  if (Math.abs(dx) >= hwx || Math.abs(dz) >= hdz) return [x, z];
  const ox = hwx - Math.abs(dx);
  const oz = hdz - Math.abs(dz);
  if (ox < oz) {
    x = dx >= 0 ? b.x + hwx : b.x - hwx;
  } else {
    z = dz >= 0 ? b.z + hdz : b.z - hdz;
  }
  return [x, z];
}

function pushOutCircle(x: number, z: number, pad: number, c: CircleCollider): [number, number] {
  const dx = x - c.x;
  const dz = z - c.z;
  const d = Math.hypot(dx, dz);
  const minD = c.r + pad;
  if (d < 0.001) {
    return [c.x + minD, c.z];
  }
  if (d >= minD) return [x, z];
  const push = (minD - d) / d;
  return [x + dx * push, z + dz * push];
}

export function resolveWorldXZ(
  x: number,
  z: number,
  radius: number,
  aabbs = VILLAGE_AABBS,
  circles = VILLAGE_CIRCLES,
  trees = VILLAGE_TREES,
  rocks = VILLAGE_ROCKS,
): { x: number; z: number } {
  for (let pass = 0; pass < 4; pass++) {
    for (const b of aabbs) {
      if (b.kind === "water" && onBridge(x, z)) continue;
      [x, z] = pushOutAabb(x, z, radius, b);
    }
    for (const c of circles) {
      if (c.walkable) continue;
      [x, z] = pushOutCircle(x, z, radius, c);
    }
    for (const t of trees) {
      [x, z] = pushOutCircle(x, z, radius, t);
    }
    for (const r of rocks) {
      [x, z] = pushOutCircle(x, z, radius, r);
    }
    [x, z] = pushVillageFence(x, z, radius);
  }
  return { x, z };
}

export function isWalkableXZ(x: number, z: number, radius: number): boolean {
  if (onBridge(x, z)) return true;
  for (const b of VILLAGE_AABBS) {
    if (insideAabb(x, z, radius, b)) return false;
  }
  for (const c of VILLAGE_CIRCLES) {
    if (c.walkable) continue;
    if (Math.hypot(x - c.x, z - c.z) < c.r + radius) return false;
  }
  for (const t of VILLAGE_TREES) {
    if (Math.hypot(x - t.x, z - t.z) < t.r + radius) return false;
  }
  for (const r of VILLAGE_ROCKS) {
    if (Math.hypot(x - r.x, z - r.z) < r.r + radius) return false;
  }
  if (hitVillageFence(x, z, radius)) return false;
  return true;
}

export function softSeparationXZ(
  x: number,
  z: number,
  radius: number,
  others: Array<{ x: number; z: number; r: number }>,
): { x: number; z: number } {
  let sx = 0;
  let sz = 0;
  for (const o of others) {
    const dx = x - o.x;
    const dz = z - o.z;
    const d = Math.hypot(dx, dz);
    const minD = radius + o.r;
    if (d < 0.001 || d >= minD) continue;
    const push = (minD - d) / d;
    sx += dx * push;
    sz += dz * push;
  }
  return { x: sx, z: sz };
}

export function npcBodyRadius(type: string): number {
  return type === "CHILD" ? 0.26 : 0.32;
}

export const PLAYER_CAPSULE = { radius: PLAYER.radius, height: PLAYER.height };

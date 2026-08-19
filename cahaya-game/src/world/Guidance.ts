/** World +Z is north, +X is east. Compass, minimap, objective, and NPC hints share this. */

export const CARDINALS = ["N", "NE", "E", "SE", "S", "SW", "W", "NW"] as const;
export type Cardinal = (typeof CARDINALS)[number];

export function cardinalFromDelta(dx: number, dz: number): Cardinal {
  if (dx === 0 && dz === 0) return "N";
  let deg = (Math.atan2(dx, dz) * 180) / Math.PI;
  if (deg < 0) deg += 360;
  const i = Math.round(deg / 45) % 8;
  return CARDINALS[i < 0 ? i + 8 : i];
}

export function cardinalID(c: Cardinal): string {
  switch (c) {
    case "N": return "utara";
    case "NE": return "timur laut";
    case "E": return "timur";
    case "SE": return "tenggara";
    case "S": return "selatan";
    case "SW": return "barat daya";
    case "W": return "barat";
    case "NW": return "barat laut";
  }
}

export function cardinalJV(c: Cardinal): string {
  switch (c) {
    case "N": return "ngalor";
    case "NE": return "ngetan-lor";
    case "E": return "ngetan";
    case "SE": return "ngetan-kidul";
    case "S": return "ngidul";
    case "SW": return "ngulon-kidul";
    case "W": return "ngulon";
    case "NW": return "ngulon-lor";
  }
}

export function heightMark(dy: number): "up" | "down" | "level" {
  if (dy > 1.4) return "up";
  if (dy < -1.4) return "down";
  return "level";
}

export function heightGlyph(mark: "up" | "down" | "level"): string {
  if (mark === "up") return "▲";
  if (mark === "down") return "▼";
  return "●";
}

export function formatMeters(dist: number): string {
  if (dist >= 1000) return `${(dist / 1000).toFixed(1)} km`;
  return `${Math.round(dist)} m`;
}

export function markerScale(dist: number): number {
  if (dist < 12) return 1;
  if (dist > 180) return 0.42;
  return 1 - (dist - 12) * 0.0032;
}

export function softGuidance(landmark: string, cardinal: Cardinal): { jv: string; id: string } {
  const name = landmark || "tujuan";
  return {
    jv: `Le, yen arep menyang ${name}, mlakua ${cardinalJV(cardinal)}.`,
    id: `Jalan menuju ${name} berada di arah ${cardinalID(cardinal)}.`,
  };
}

export function mainPathSamples(fromZ: number, toZ: number, step = 8): Array<{ x: number; z: number }> {
  const out: Array<{ x: number; z: number }> = [];
  const a = Math.min(fromZ, toZ);
  const b = Math.max(fromZ, toZ);
  for (let z = a; z <= b; z += step) out.push({ x: 0, z });
  out.push({ x: 0, z: b });
  return out;
}

export class GuidanceSystem {
  pulse = 0;
  toast = "";
  toastSub = "";
  toastUntil = 0;
  npcUntil = 0;
  followMain = false;
  regionBanner = "";
  regionUntil = 0;
  idle = 0;
  peekUntil = 0;
  peeked = "";
  occluded = false;
  glance = 0;
  private closest = 9999;
  private awayMeters = 0;
  private awayTime = 0;
  private lastNavKey = "";
  private lastZone = "";

  resetObjective(navX: number, navZ: number): void {
    this.lastNavKey = `${navX},${navZ}`;
    this.closest = 9999;
    this.awayMeters = 0;
    this.awayTime = 0;
  }

  update(dt: number, opts: {
    x: number;
    y: number;
    z: number;
    navX: number;
    navZ: number;
    landmark: string;
    zone: string;
    zoneTitle: string;
    moving: boolean;
    combat: boolean;
    npcName?: string;
  }): void {
    const key = `${opts.navX},${opts.navZ}`;
    if (key !== this.lastNavKey) this.resetObjective(opts.navX, opts.navZ);
    const dist = Math.hypot(opts.navX - opts.x, opts.navZ - opts.z);
    if (dist < this.closest) {
      this.closest = dist;
      this.awayMeters = 0;
      this.awayTime = 0;
    } else {
      this.awayMeters = dist - this.closest;
      if (this.awayMeters > 4) this.awayTime += dt;
      else this.awayTime = 0;
    }

    const away = this.awayMeters > 8;
    if (away) this.pulse = Math.min(1, this.pulse + dt * 0.7);
    else this.pulse = Math.max(0, this.pulse - dt * 1.4);

    if (this.awayMeters > 100 && this.awayTime >= 15 && performance.now() > this.npcUntil) {
      const c = cardinalFromDelta(opts.navX - opts.x, opts.navZ - opts.z);
      const msg = softGuidance(opts.landmark, c);
      const who = opts.npcName ? `${opts.npcName}: ` : "";
      this.toast = who + msg.jv;
      this.toastSub = msg.id;
      this.toastUntil = performance.now() + 5200;
      this.npcUntil = performance.now() + 60000;
      this.awayTime = 0;
    }

    if (opts.zone && opts.zone !== this.lastZone) {
      this.lastZone = opts.zone;
      this.regionBanner = opts.zoneTitle;
      this.regionUntil = performance.now() + 3800;
    }

    if (opts.moving || opts.combat) this.idle = 0;
    else this.idle += dt;
    this.glance = this.idle >= 8 && dist > 6 ? 1 : Math.max(0, this.glance - dt);

    if (this.peekUntil > 0) this.peekUntil = Math.max(0, this.peekUntil - dt);
  }

  maybePeek(landmarkId: string, dist: number): boolean {
    if (!landmarkId || dist > 28 || dist < 8) return false;
    if (this.peeked === landmarkId) return false;
    this.peeked = landmarkId;
    this.peekUntil = 2.8;
    return true;
  }

  toastVisible(): boolean {
    return !!this.toast && performance.now() < this.toastUntil;
  }

  bannerVisible(): boolean {
    return !!this.regionBanner && performance.now() < this.regionUntil;
  }
}

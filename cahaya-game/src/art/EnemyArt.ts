import * as THREE from "three";
import { hash32, metalMat, std } from "./PBR";

const SILUMAN = [
  "ragha", "gravon", "velra", "kairoth", "nymra", "vorak", "zeran", "malvek", "torga", "selka",
  "dravon", "morgha", "kelran", "arvok", "nyxara", "torven", "veyra", "gharon", "lurak", "sorya",
  "varok", "zavra", "orvak", "merra", "tharos", "ragna", "velkor", "aranya", "gorven", "sylra",
  "korvan", "elyra", "avaron",
] as const;

export function buildEnemyMesh(kind: string, rank: string, shared: THREE.MeshStandardMaterial): THREE.Group {
  if (kind === "training_dummy") return dummy(shared);
  if (kind === "forest_fang") return fang(shared);
  if (kind === "shadow_imp") return imp(shared);
  if (kind === "stone_beast") return golem(shared);
  if (kind === "elite_shadow_beast") return elite(shared);
  const idx = SILUMAN.indexOf(kind as (typeof SILUMAN)[number]);
  const mesh = idx >= 0 ? siluman(idx, shared) : hashedBeast(kind, shared);
  const scale =
    rank === "world" ? 3.4 :
    rank === "boss" || kind === "avaron" ? 2.55 :
    rank === "elite" || rank === "mini" ? 1.55 :
    idx >= 0 && idx % 5 === 4 ? 1.7 :
    1;
  mesh.scale.setScalar(scale);
  mesh.userData.silumanIndex = idx;
  return mesh;
}

function dummy(mat: THREE.MeshStandardMaterial): THREE.Group {
  const g = new THREE.Group();
  const post = new THREE.Mesh(new THREE.CylinderGeometry(0.12, 0.16, 1.5, 8), std(0x8b5a2b, 0.9, 0.02));
  post.position.y = 0.75;
  const torso = new THREE.Mesh(new THREE.CylinderGeometry(0.28, 0.32, 0.7, 10), mat);
  torso.position.y = 1.15;
  const head = new THREE.Mesh(new THREE.SphereGeometry(0.2, 10, 10), mat);
  head.position.y = 1.65;
  g.add(post, torso, head);
  return g;
}

function fang(mat: THREE.MeshStandardMaterial): THREE.Group {
  const g = new THREE.Group();
  const body = new THREE.Mesh(new THREE.SphereGeometry(0.38, 12, 10), mat);
  body.scale.set(1.4, 0.75, 1.8);
  body.position.set(0, 0.45, 0);
  const head = new THREE.Mesh(new THREE.SphereGeometry(0.26, 12, 10), mat);
  head.scale.set(1.1, 0.85, 1.4);
  head.position.set(0, 0.55, 0.42);
  const ear = new THREE.Mesh(new THREE.ConeGeometry(0.1, 0.32, 6), mat);
  ear.position.set(-0.14, 0.82, 0.32);
  const ear2 = ear.clone();
  ear2.position.x = 0.14;
  const tail = new THREE.Mesh(new THREE.CapsuleGeometry(0.06, 0.7, 4, 6), mat);
  tail.position.set(0, 0.4, -0.7);
  tail.rotation.x = 0.8;
  g.add(body, head, ear, ear2, tail);
  return g;
}

function imp(mat: THREE.MeshStandardMaterial): THREE.Group {
  const g = new THREE.Group();
  const body = new THREE.Mesh(new THREE.SphereGeometry(0.22, 10, 10), mat);
  body.scale.set(0.9, 1.35, 0.8);
  body.position.y = 0.55;
  const head = new THREE.Mesh(new THREE.SphereGeometry(0.2, 10, 10), mat);
  head.position.y = 1.05;
  const horn = new THREE.Mesh(new THREE.ConeGeometry(0.05, 0.28, 6), std(0x1a1028, 0.5, 0.1));
  horn.position.set(-0.08, 1.28, 0);
  const horn2 = horn.clone();
  horn2.position.x = 0.08;
  const wing = new THREE.Mesh(new THREE.CircleGeometry(0.28, 8), new THREE.MeshStandardMaterial({ color: 0x2a1840, side: THREE.DoubleSide, roughness: 0.7 }));
  wing.position.set(-0.32, 0.7, -0.05);
  wing.rotation.y = 0.6;
  const wing2 = wing.clone();
  wing2.position.x = 0.32;
  wing2.rotation.y = -0.6;
  g.add(body, head, horn, horn2, wing, wing2);
  return g;
}

function golem(mat: THREE.MeshStandardMaterial): THREE.Group {
  const g = new THREE.Group();
  const body = new THREE.Mesh(new THREE.DodecahedronGeometry(0.42, 0), mat);
  body.position.y = 0.7;
  const head = new THREE.Mesh(new THREE.DodecahedronGeometry(0.28, 0), mat);
  head.position.y = 1.28;
  const arm = new THREE.Mesh(new THREE.CylinderGeometry(0.12, 0.16, 0.7, 6), mat);
  arm.position.set(-0.45, 0.7, 0);
  arm.rotation.z = 0.4;
  const arm2 = arm.clone();
  arm2.position.x = 0.45;
  arm2.rotation.z = -0.4;
  g.add(body, head, arm, arm2);
  return g;
}

function elite(mat: THREE.MeshStandardMaterial): THREE.Group {
  const g = fang(mat);
  g.scale.setScalar(1.35);
  const crest = new THREE.Mesh(new THREE.ConeGeometry(0.18, 0.5, 6), std(0x7e22ce, 0.4, 0.2, undefined, 0xa855f7, 0.6));
  crest.position.set(0, 1.15, 0.2);
  g.add(crest);
  return g;
}

function hashedBeast(kind: string, mat: THREE.MeshStandardMaterial): THREE.Group {
  return siluman(hash32(kind) % 33, mat);
}

function siluman(i: number, mat: THREE.MeshStandardMaterial): THREE.Group {
  const accent = std(hue(i * 17 + 40), 0.42, 0.18, undefined, hue(i * 17 + 40), 0.45);
  const metal = metalMat(hue(i * 11 + 200), 0.28 + (i % 5) * 0.06);
  const g = bodyFamily(i, mat, accent);
  addHead(g, i, mat, accent, metal);
  addWeapon(g, i, metal, accent);
  addVfx(g, i);
  if (i === 32) {
    const crown = new THREE.Mesh(new THREE.TorusGeometry(0.42, 0.05, 6, 16), metal);
    crown.position.y = 2.15;
    crown.rotation.x = Math.PI / 2;
    const halo = new THREE.Mesh(new THREE.RingGeometry(0.7, 0.95, 20), new THREE.MeshBasicMaterial({ color: 0xffe08a, transparent: true, opacity: 0.35, side: THREE.DoubleSide, depthWrite: false }));
    halo.rotation.x = Math.PI / 2;
    halo.position.y = 2.4;
    g.add(crown, halo);
  }
  return g;
}

function bodyFamily(i: number, mat: THREE.MeshStandardMaterial, accent: THREE.MeshStandardMaterial): THREE.Group {
  const g = new THREE.Group();
  const fam = i % 11;
  if (fam === 0) {
    const shroud = new THREE.Mesh(new THREE.ConeGeometry(0.38 + (i % 3) * 0.04, 1.55, 8), mat);
    shroud.position.y = 0.72;
    g.add(shroud);
  } else if (fam === 1) {
    const body = new THREE.Mesh(new THREE.DodecahedronGeometry(0.4 + (i % 4) * 0.04, 0), mat);
    body.position.y = 0.75;
    g.add(body, limb(mat, -0.5, 0.7, 0.45), limb(mat, 0.5, 0.7, -0.45));
  } else if (fam === 2) {
    const body = new THREE.Mesh(new THREE.SphereGeometry(0.32, 10, 8), mat);
    body.scale.set(0.85, 1.55, 0.8);
    body.position.y = 0.85;
    const thorn = new THREE.Mesh(new THREE.ConeGeometry(0.08, 0.4, 5), accent);
    thorn.position.set(0.2, 1.1, 0.1);
    g.add(body, thorn);
  } else if (fam === 3) {
    const body = new THREE.Mesh(new THREE.SphereGeometry(0.42, 12, 10), mat);
    body.scale.set(1.5, 0.7, 1.7);
    body.position.y = 0.5;
    g.add(body);
    for (let n = 0; n < 4; n++) {
      const a = (n / 4) * Math.PI * 2;
      const leg = new THREE.Mesh(new THREE.CapsuleGeometry(0.05, 0.5, 3, 6), mat);
      leg.position.set(Math.cos(a) * 0.35, 0.22, Math.sin(a) * 0.35);
      g.add(leg);
    }
  } else if (fam === 4) {
    const body = new THREE.Mesh(new THREE.CapsuleGeometry(0.22, 1.15, 5, 8), mat);
    body.position.y = 0.95;
    g.add(body, limb(mat, -0.28, 1.1, 0.7), limb(mat, 0.28, 1.1, -0.7));
  } else if (fam === 5) {
    for (let s = 0; s < 7; s++) {
      const seg = new THREE.Mesh(new THREE.SphereGeometry(0.16 - s * 0.012, 8, 8), s % 2 ? accent : mat);
      seg.position.set(0, 0.22 + s * 0.16, s * 0.1);
      g.add(seg);
    }
  } else if (fam === 6) {
    const body = new THREE.Mesh(new THREE.SphereGeometry(0.3, 10, 8), mat);
    body.position.y = 0.75;
    const wing = new THREE.Mesh(new THREE.ConeGeometry(0.5, 0.1, 4), mat);
    wing.rotation.z = 1.15;
    wing.position.set(-0.5, 0.85, 0);
    const wing2 = wing.clone();
    wing2.position.x = 0.5;
    wing2.rotation.z = -1.15;
    g.add(body, wing, wing2);
  } else if (fam === 7) {
    const body = new THREE.Mesh(new THREE.OctahedronGeometry(0.38), accent);
    body.position.y = 0.95;
    const ring = new THREE.Mesh(new THREE.TorusGeometry(0.48, 0.045, 6, 18), mat);
    ring.position.y = 0.95;
    g.add(body, ring);
  } else if (fam === 8) {
    const stem = new THREE.Mesh(new THREE.CylinderGeometry(0.08, 0.14, 1.0, 7), std(0x3f6b32, 0.85));
    stem.position.y = 0.5;
    const bloom = new THREE.Mesh(new THREE.SphereGeometry(0.34, 10, 8), mat);
    bloom.position.y = 1.15;
    g.add(stem, bloom);
  } else if (fam === 9) {
    const body = new THREE.Mesh(new THREE.CapsuleGeometry(0.26, 0.7, 6, 10), mat);
    body.position.y = 0.9;
    g.add(body, limb(mat, -0.32, 0.95, 0.35), limb(mat, 0.32, 0.95, -0.35));
  } else {
    const body = new THREE.Mesh(new THREE.ConeGeometry(0.42, 1.7, 5), mat);
    body.position.y = 0.85;
    g.add(body);
  }
  return g;
}

function addHead(g: THREE.Group, i: number, mat: THREE.MeshStandardMaterial, accent: THREE.MeshStandardMaterial, metal: THREE.MeshStandardMaterial): void {
  const y = 1.45 + (i % 5) * 0.06;
  const style = i % 8;
  if (style === 0) {
    const helm = new THREE.Mesh(new THREE.SphereGeometry(0.28, 12, 8, 0, Math.PI * 2, 0, 1.15), metal);
    helm.position.y = y;
    g.add(helm);
  } else if (style === 1) {
    const head = new THREE.Mesh(new THREE.DodecahedronGeometry(0.24, 0), mat);
    head.position.y = y;
    g.add(head);
  } else if (style === 2) {
    const head = new THREE.Mesh(new THREE.SphereGeometry(0.22, 10, 8), mat);
    head.position.y = y;
    const mask = new THREE.Mesh(new THREE.CircleGeometry(0.16, 8), accent);
    mask.position.set(0, y, 0.2);
    g.add(head, mask);
  } else if (style === 3) {
    const head = new THREE.Mesh(new THREE.ConeGeometry(0.2, 0.55, 6), mat);
    head.position.y = y + 0.1;
    g.add(head);
  } else if (style === 4) {
    const head = new THREE.Mesh(new THREE.SphereGeometry(0.2, 10, 8), mat);
    head.position.y = y;
    for (const sx of [-1, 1]) {
      const horn = new THREE.Mesh(new THREE.ConeGeometry(0.05, 0.34 + (i % 3) * 0.05, 5), metal);
      horn.position.set(sx * 0.12, y + 0.28, 0);
      horn.rotation.z = sx * -0.35;
      g.add(horn);
    }
    g.add(head);
  } else if (style === 5) {
    const jaw = new THREE.Mesh(new THREE.SphereGeometry(0.26, 10, 8), mat);
    jaw.scale.set(1.2, 0.7, 1.4);
    jaw.position.set(0, y - 0.05, 0.08);
    g.add(jaw);
  } else if (style === 6) {
    const crystal = new THREE.Mesh(new THREE.OctahedronGeometry(0.22), accent);
    crystal.position.y = y;
    g.add(crystal);
  } else {
    const head = new THREE.Mesh(new THREE.SphereGeometry(0.24, 12, 10), mat);
    head.position.y = y;
    const crest = new THREE.Mesh(new THREE.ConeGeometry(0.1, 0.42, 5), accent);
    crest.position.set(0, y + 0.32, 0.04);
    g.add(head, crest);
  }
}

function addWeapon(g: THREE.Group, i: number, metal: THREE.MeshStandardMaterial, accent: THREE.MeshStandardMaterial): void {
  const kind = i % 9;
  if (kind === 0) {
    const blade = new THREE.Mesh(new THREE.BoxGeometry(0.07, 1.25, 0.16), metal);
    blade.position.set(0.55, 1.05, 0.08);
    g.add(blade);
  } else if (kind === 1) {
    const staff = new THREE.Mesh(new THREE.CylinderGeometry(0.03, 0.04, 1.6, 6), metal);
    staff.position.set(0.42, 0.95, 0.1);
    const orb = new THREE.Mesh(new THREE.SphereGeometry(0.12, 10, 8), accent);
    orb.position.set(0.42, 1.75, 0.1);
    g.add(staff, orb);
  } else if (kind === 2) {
    const claw = new THREE.Mesh(new THREE.ConeGeometry(0.06, 0.45, 4), metal);
    claw.position.set(0.48, 0.85, 0.22);
    claw.rotation.x = 0.9;
    g.add(claw);
  } else if (kind === 3) {
    const scythe = new THREE.Mesh(new THREE.TorusGeometry(0.32, 0.04, 5, 12, Math.PI), metal);
    scythe.position.set(0.5, 1.35, 0);
    scythe.rotation.z = 0.4;
    g.add(scythe);
  } else if (kind === 4) {
    const hammer = new THREE.Mesh(new THREE.CylinderGeometry(0.16, 0.16, 0.28, 6), metal);
    hammer.position.set(0.5, 0.95, 0.12);
    g.add(hammer);
  } else if (kind === 5) {
    const spear = new THREE.Mesh(new THREE.CylinderGeometry(0.025, 0.03, 1.8, 6), metal);
    spear.position.set(0.4, 1.1, 0.15);
    spear.rotation.z = -0.25;
    g.add(spear);
  } else if (kind === 6) {
    const fan = new THREE.Mesh(new THREE.CircleGeometry(0.32, 7), accent);
    fan.position.set(0.48, 1.05, 0.1);
    fan.rotation.y = 0.5;
    g.add(fan);
  } else if (kind === 7) {
    const shield = new THREE.Mesh(new THREE.CircleGeometry(0.34, 8), metal);
    shield.position.set(-0.45, 1.0, 0.12);
    g.add(shield);
  } else {
    const orb = new THREE.Mesh(new THREE.OctahedronGeometry(0.16), accent);
    orb.position.set(0.5, 1.2, 0.1);
    g.add(orb);
  }
}

function addVfx(g: THREE.Group, i: number): void {
  const n = 10 + (i % 8);
  const pos = new Float32Array(n * 3);
  for (let p = 0; p < n; p++) {
    const a = (p / n) * Math.PI * 2;
    pos[p * 3] = Math.cos(a) * (0.35 + (i % 4) * 0.06);
    pos[p * 3 + 1] = 0.4 + (p % 6) * 0.18;
    pos[p * 3 + 2] = Math.sin(a) * (0.35 + (i % 3) * 0.05);
  }
  const geo = new THREE.BufferGeometry();
  geo.setAttribute("position", new THREE.BufferAttribute(pos, 3));
  const pts = new THREE.Points(geo, new THREE.PointsMaterial({
    color: hue(i * 17 + 40),
    size: 0.06,
    transparent: true,
    opacity: 0.7,
    depthWrite: false,
    blending: THREE.AdditiveBlending,
  }));
  g.add(pts);
}

function limb(mat: THREE.Material, x: number, y: number, rot: number): THREE.Mesh {
  const m = new THREE.Mesh(new THREE.CapsuleGeometry(0.08, 0.55, 4, 6), mat);
  m.position.set(x, y, 0);
  m.rotation.z = rot;
  return m;
}

function hue(deg: number): number {
  const t = ((deg % 360) + 360) % 360 / 360;
  const c = new THREE.Color().setHSL(t, 0.55, 0.42);
  return c.getHex();
}

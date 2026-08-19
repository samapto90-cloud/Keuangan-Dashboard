import * as THREE from "three";
import { PLAYER_FORMS, type PlayerFormDef, type PlayerFormId } from "./PlayerForms";
import { clothMat, leatherMat, metalMat, noiseMap, skinMat, std } from "../art/PBR";

export interface PlayerRig {
  root: THREE.Group;
  armL: THREE.Group;
  armR: THREE.Group;
  legL: THREE.Group;
  legR: THREE.Group;
  aura?: THREE.Object3D;
  lightning?: THREE.Object3D;
}

export function buildPlayerModel(formId: PlayerFormId): PlayerRig {
  const form = PLAYER_FORMS[formId];
  const root = new THREE.Group();
  const mats = makeMaterials(form);

  const hips = new THREE.Group();
  hips.position.y = 0.98;
  root.add(hips);

  const torso = makeTorso(form, mats);
  torso.userData.attach = "body";
  hips.add(torso);

  const head = makeHead(form, mats);
  head.userData.attach = "head";
  head.position.set(0, 0.62, 0.02);
  torso.add(head);

  const armL = makeArm(form, mats, -1);
  const armR = makeArm(form, mats, 1);
  armL.position.set(-0.24, 0.5, 0);
  armR.position.set(0.24, 0.5, 0);
  armR.userData.attach = "weapon";
  torso.add(armL, armR);

  const legL = makeLeg(form, mats, -1);
  const legR = makeLeg(form, mats, 1);
  legL.position.set(-0.1, 0, 0);
  legR.position.set(0.1, 0, 0);
  legL.userData.attach = "legs";
  hips.add(legL, legR);

  if (form.tail) hips.add(makeTail(mats.fur));

  let aura: THREE.Object3D | undefined;
  aura = makeAura(form);
  root.add(aura);
  let lightning: THREE.Object3D | undefined;
  if (form.lightning !== "none") {
    lightning = makeLightning(form.lightning === "cyan" ? 0x9ef6ff : 0xffe08a);
    root.add(lightning);
  }

  root.traverse((obj) => {
    if (obj instanceof THREE.Mesh) {
      obj.castShadow = true;
      obj.receiveShadow = true;
    }
  });
  return { root, armL, armR, legL, legR, aura, lightning };
}

export function disposePlayerModel(root: THREE.Object3D): void {
  root.traverse((child) => {
    if (!(child instanceof THREE.Mesh) && !(child instanceof THREE.Points)) return;
    child.geometry.dispose();
    const raw = child.material;
    const list = Array.isArray(raw) ? raw : [raw];
    for (const mat of list) mat.dispose();
  });
}

interface Mats {
  skin: THREE.MeshStandardMaterial;
  hair: THREE.MeshStandardMaterial;
  shirt: THREE.MeshStandardMaterial;
  vest: THREE.MeshStandardMaterial;
  gold: THREE.MeshStandardMaterial;
  pants: THREE.MeshStandardMaterial;
  leather: THREE.MeshStandardMaterial;
  shoe: THREE.MeshStandardMaterial;
  fur: THREE.MeshStandardMaterial;
  eyeWhite: THREE.MeshStandardMaterial;
  iris: THREE.MeshStandardMaterial;
  dark: THREE.MeshStandardMaterial;
}

function makeMaterials(form: PlayerFormDef): Mats {
  const goldHair = form.hair.startsWith("gold");
  return {
    skin: skinMat(0xe4b48a),
    hair: goldHair
      ? std(0xffe566, 0.38, 0.08, noiseMap("hair-g", "#ffe566", "#c9a227", 30), 0xffc107, 0.35)
      : std(0x1a1418, 0.48, 0.02, noiseMap("hair-k", "#1a1418", "#3a3034", 24)),
    shirt: clothMat(0xf2efe8, "shirt"),
    vest: vestMaterial(form.vest === "gold-crack"),
    gold: metalMat(0xd4a017, 0.28),
    pants: clothMat(0x1c1d24, "pants"),
    leather: leatherMat(),
    shoe: std(form.shoes === "boots" ? 0x1a1512 : 0xeceae4, 0.55, 0.06),
    fur: std(0xb71c1c, 0.92, 0.0, noiseMap("fur", "#b71c1c", "#5c1010", 40)),
    eyeWhite: std(0xf7f4ee, 0.28, 0.02),
    iris: std(form.eyeColor, 0.22, 0.04, undefined, form.eyeEmissive, form.eyeEmissive ? 0.85 : 0),
    dark: std(0x1a1512, 0.55),
  };
}

function vestMaterial(goldCrack: boolean): THREE.MeshStandardMaterial {
  return std(goldCrack ? 0x2a2418 : 0x1c1d22, 0.52, goldCrack ? 0.38 : 0.12, noiseMap(goldCrack ? "vest-g" : "vest", "#1c1d22", goldCrack ? "#e0b84a" : "#6b6b72", 40), goldCrack ? 0x4a3508 : 0x000000, goldCrack ? 0.18 : 0);
}

function makeTorso(form: PlayerFormDef, mats: Mats): THREE.Group {
  const g = new THREE.Group();
  const chest = new THREE.Mesh(new THREE.SphereGeometry(0.2, 14, 12), form.shirt === "bare" ? mats.skin : mats.shirt);
  chest.scale.set(1.15, 1.35, 0.72);
  chest.position.y = 0.32;
  const waist = new THREE.Mesh(new THREE.SphereGeometry(0.16, 12, 10), form.shirt === "bare" ? mats.skin : mats.shirt);
  waist.scale.set(1.05, 0.9, 0.68);
  waist.position.y = 0.08;
  g.add(chest, waist);

  if (form.vest !== "none") {
    const wrap = new THREE.Mesh(new THREE.SphereGeometry(0.22, 12, 10, 0, Math.PI * 2, 0.2, 1.8), mats.vest);
    wrap.scale.set(1.2, 1.15, 0.78);
    wrap.position.y = 0.28;
    g.add(wrap);
    const trim = new THREE.Mesh(new THREE.TorusGeometry(0.17, 0.012, 6, 18), mats.gold);
    trim.position.set(0, 0.5, 0.08);
    trim.rotation.x = 1.2;
    g.add(trim);
  }

  const sash = new THREE.Mesh(
    new THREE.TorusGeometry(0.2, 0.035, 6, 16),
    std(0x14b8a6, 0.45, 0.08, undefined, 0x0f766e, 0.35),
  );
  sash.rotation.set(1.15, 0, 0.55);
  sash.position.set(0, 0.34, 0.02);
  const padL = new THREE.Mesh(new THREE.SphereGeometry(0.09, 10, 8), mats.gold);
  padL.scale.set(1.2, 0.55, 0.9);
  padL.position.set(-0.22, 0.52, 0.02);
  const padR = padL.clone();
  padR.position.x = 0.22;
  g.add(sash, padL, padR);

  const scarf = new THREE.Mesh(new THREE.CapsuleGeometry(0.05, 0.55, 4, 8), std(0x0f766e, 0.48, 0.05, undefined, 0x14b8a6, 0.25));
  scarf.position.set(0.08, 0.18, -0.16);
  scarf.rotation.set(0.85, 0.2, 0.35);
  g.add(scarf);

  const belt = new THREE.Mesh(new THREE.TorusGeometry(0.17, 0.028, 6, 16), mats.leather);
  belt.rotation.x = Math.PI / 2;
  belt.position.y = 0.02;
  const buckle = new THREE.Mesh(new THREE.SphereGeometry(0.035, 10, 8), mats.gold);
  buckle.scale.set(1.1, 0.7, 0.5);
  buckle.position.set(0, 0.02, 0.18);
  g.add(belt, buckle);

  if (form.fur) {
    const cape = new THREE.Mesh(new THREE.SphereGeometry(0.18, 10, 8, 0, Math.PI * 2, 0.4, 1.4), mats.fur);
    cape.scale.set(1.4, 1.1, 0.9);
    cape.position.set(0, 0.42, -0.08);
    g.add(cape);
  }
  return g;
}

function makeHead(form: PlayerFormDef, mats: Mats): THREE.Group {
  const g = new THREE.Group();
  const skull = new THREE.Mesh(new THREE.SphereGeometry(0.155, 18, 16), mats.skin);
  skull.scale.set(0.92, 1.05, 0.95);
  g.add(skull);
  const jaw = new THREE.Mesh(new THREE.SphereGeometry(0.1, 12, 10), mats.skin);
  jaw.scale.set(0.95, 0.7, 0.85);
  jaw.position.set(0, -0.08, 0.02);
  g.add(jaw);

  const earL = new THREE.Mesh(new THREE.SphereGeometry(0.032, 8, 8), mats.skin);
  earL.scale.set(0.55, 1.1, 0.7);
  earL.position.set(-0.15, 0.01, -0.02);
  const earR = earL.clone();
  earR.position.x = 0.15;
  g.add(earL, earR);

  g.add(makeEye(mats, -1), makeEye(mats, 1));
  if (form.brows) {
    const brow = new THREE.Mesh(new THREE.CapsuleGeometry(0.012, 0.055, 3, 6), mats.hair);
    const browL = brow.clone();
    browL.position.set(-0.055, 0.055, 0.13);
    browL.rotation.z = 0.25;
    const browR = brow.clone();
    browR.position.set(0.055, 0.055, 0.13);
    browR.rotation.z = -0.25;
    g.add(browL, browR);
  }
  const nose = new THREE.Mesh(new THREE.SphereGeometry(0.022, 8, 8), mats.skin);
  nose.scale.set(0.7, 1.1, 1.2);
  nose.position.set(0, -0.01, 0.145);
  const mouth = new THREE.Mesh(new THREE.CapsuleGeometry(0.01, 0.04, 3, 6), mats.dark);
  mouth.rotation.z = Math.PI / 2;
  mouth.position.set(0, -0.07, 0.138);
  g.add(nose, mouth);
  g.add(makeHair(form, mats.hair));
  return g;
}

function makeEye(mats: Mats, side: number): THREE.Group {
  const g = new THREE.Group();
  g.position.set(side * 0.052, 0.02, 0.125);
  const white = new THREE.Mesh(new THREE.SphereGeometry(0.028, 10, 8), mats.eyeWhite);
  white.scale.set(1, 0.78, 0.55);
  const iris = new THREE.Mesh(new THREE.SphereGeometry(0.016, 10, 8), mats.iris);
  iris.position.z = 0.012;
  const pupil = new THREE.Mesh(new THREE.SphereGeometry(0.008, 8, 8), mats.dark);
  pupil.position.z = 0.02;
  g.add(white, iris, pupil);
  return g;
}

function makeHair(form: PlayerFormDef, mat: THREE.Material): THREE.Group {
  const g = new THREE.Group();
  const dome = new THREE.Mesh(new THREE.SphereGeometry(0.17, 12, 10, 0, Math.PI * 2, 0, 1.2), mat);
  dome.position.y = 0.04;
  g.add(dome);
  const long = form.hair.endsWith("long");
  const n = long ? 16 : 12;
  for (let i = 0; i < n; i++) {
    const a = -0.95 + (i / (n - 1)) * 1.9;
    const lock = new THREE.Mesh(new THREE.CapsuleGeometry(0.026, long ? 0.48 : 0.2, 3, 6), mat);
    lock.position.set(Math.sin(a) * 0.11, long ? -0.14 : 0.12, Math.cos(a) * 0.08 - 0.02);
    lock.rotation.set(long ? 0.55 : -0.95, a * 0.4, a * 0.22);
    g.add(lock);
  }
  const crest = new THREE.Mesh(new THREE.CapsuleGeometry(0.034, 0.42, 3, 6), mat);
  crest.position.set(-0.06, 0.28, 0.02);
  crest.rotation.set(-0.35, 0.4, -0.55);
  const side = new THREE.Mesh(new THREE.CapsuleGeometry(0.028, 0.32, 3, 6), mat);
  side.position.set(0.14, 0.08, 0.06);
  side.rotation.set(-0.2, -0.6, 0.9);
  const bang = new THREE.Mesh(new THREE.CapsuleGeometry(0.03, 0.16, 3, 6), mat);
  bang.position.set(0.05, 0.06, 0.14);
  bang.rotation.x = -1.05;
  g.add(crest, side, bang);
  return g;
}

function makeArm(form: PlayerFormDef, mats: Mats, side: number): THREE.Group {
  const g = new THREE.Group();
  const sleeve = form.shirt === "white";
  const upper = new THREE.Mesh(new THREE.CapsuleGeometry(0.055, 0.28, 5, 10), form.fur ? mats.fur : sleeve ? mats.shirt : mats.skin);
  upper.position.y = -0.16;
  const lower = new THREE.Mesh(new THREE.CapsuleGeometry(0.048, 0.26, 5, 10), mats.skin);
  lower.position.y = -0.44;
  g.add(upper, lower, makeHand(mats));
  const glove = new THREE.Mesh(new THREE.TorusGeometry(0.05, 0.012, 6, 12), mats.leather);
  glove.rotation.x = Math.PI / 2;
  glove.position.y = -0.58;
  g.add(glove);
  g.rotation.z = side * 0.12;
  return g;
}

function makeHand(mats: Mats): THREE.Group {
  const g = new THREE.Group();
  g.position.y = -0.62;
  const palm = new THREE.Mesh(new THREE.SphereGeometry(0.042, 8, 8), mats.skin);
  palm.scale.set(0.85, 1, 0.7);
  g.add(palm);
  for (let i = 0; i < 4; i++) {
    const f = new THREE.Mesh(new THREE.CapsuleGeometry(0.01, 0.055, 3, 5), mats.skin);
    f.position.set(0.01, -0.05, -0.03 + i * 0.018);
    f.rotation.x = 0.35;
    g.add(f);
  }
  return g;
}

function makeLeg(form: PlayerFormDef, mats: Mats, side: number): THREE.Group {
  const g = new THREE.Group();
  g.rotation.z = side * 0.04;
  const thigh = new THREE.Mesh(new THREE.CapsuleGeometry(0.075, 0.36, 5, 10), mats.pants);
  thigh.position.y = -0.26;
  const shin = new THREE.Mesh(new THREE.CapsuleGeometry(0.06, 0.34, 5, 10), form.pants === "torn" && form.fur ? mats.fur : mats.pants);
  shin.position.y = -0.62;
  g.add(thigh, shin);
  const shoe = new THREE.Mesh(new THREE.SphereGeometry(0.07, 10, 8), mats.shoe);
  shoe.scale.set(0.85, 0.55, 1.35);
  shoe.position.set(0, -0.88, 0.04);
  g.add(shoe);
  if (form.shoes === "boots") {
    const cuff = new THREE.Mesh(new THREE.CylinderGeometry(0.07, 0.075, 0.12, 10), mats.leather);
    cuff.position.y = -0.78;
    g.add(cuff);
  }
  return g;
}

function makeTail(fur: THREE.Material): THREE.Group {
  const g = new THREE.Group();
  g.position.set(0, -0.08, -0.12);
  for (let i = 0; i < 5; i++) {
    const seg = new THREE.Mesh(new THREE.SphereGeometry(0.055 - i * 0.006, 8, 8), fur);
    seg.position.set(0.04 * i, -0.03 * i, -0.09 * i);
    g.add(seg);
  }
  return g;
}

function makeAura(form: PlayerFormDef): THREE.Group {
  const color = form.aura === "orange" ? 0xff6a2a : form.aura === "none" ? 0x2dd4bf : 0xffe08a;
  const g = new THREE.Group();
  const n = form.aura === "gold-strong" ? 64 : 42;
  const pos = new Float32Array(n * 3);
  for (let i = 0; i < n; i++) {
    const a = (i / n) * Math.PI * 2;
    pos[i * 3] = Math.cos(a) * (0.28 + (i % 5) * 0.05);
    pos[i * 3 + 1] = 0.2 + (i % 9) * 0.16;
    pos[i * 3 + 2] = Math.sin(a) * (0.28 + (i % 4) * 0.05);
  }
  const geo = new THREE.BufferGeometry();
  geo.setAttribute("position", new THREE.BufferAttribute(pos, 3));
  const pts = new THREE.Points(geo, new THREE.PointsMaterial({ color, size: 0.07, transparent: true, opacity: 0.75, depthWrite: false, blending: THREE.AdditiveBlending }));
  const glow = new THREE.Mesh(
    new THREE.CapsuleGeometry(0.42, 1.05, 6, 12),
    new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.08, blending: THREE.AdditiveBlending, depthWrite: false }),
  );
  glow.position.y = 0.95;
  g.add(pts, glow);
  g.position.y = 0.1;
  return g;
}

function makeLightning(color: number): THREE.Group {
  const g = new THREE.Group();
  const mat = new THREE.MeshBasicMaterial({ color, transparent: true, opacity: 0.8, blending: THREE.AdditiveBlending, depthWrite: false });
  for (let i = 0; i < 8; i++) {
    const bolt = new THREE.Mesh(new THREE.CapsuleGeometry(0.012, 0.32, 2, 4), mat);
    const a = (i / 8) * Math.PI * 2;
    bolt.position.set(Math.cos(a) * 0.38, 0.75 + (i % 3) * 0.2, Math.sin(a) * 0.38);
    bolt.rotation.z = a * 0.5;
    g.add(bolt);
  }
  return g;
}

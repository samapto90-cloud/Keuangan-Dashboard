import * as THREE from "three";
import type { Player } from "../player/Player";
import type { EquipmentView } from "../network/NetworkMessage";

const MARK = "eq-attach";

export class EquipmentVisual {
  apply(player: Player, gear: EquipmentView): void {
    clearAttach(player.mesh);
    const head = findAttach(player.mesh, "head");
    const body = findAttach(player.mesh, "body");
    const weapon = findAttach(player.mesh, "weapon") ?? player.armR;
    const legs = findAttach(player.mesh, "legs") ?? player.legL;
    if (gear.HEAD && head) head.add(cap());
    if (gear.BODY && body) body.add(chestPlate());
    if (gear.LEGS && legs) legs.add(bootCuff());
    if (gear.WEAPON && weapon) weapon.add(staff());
    if (gear.ACCESSORY_1 && weapon) weapon.add(gloveBand());
    if (gear.ACCESSORY_2 && head) head.add(pendant());
    if (gear.ACCESSORY_3 && weapon) weapon.add(bracelet());
  }
}

function findAttach(root: THREE.Object3D, id: string): THREE.Object3D | undefined {
  let found: THREE.Object3D | undefined;
  root.traverse((obj) => {
    if (!found && obj.userData.attach === id) found = obj;
  });
  return found;
}

function clearAttach(root: THREE.Object3D): void {
  const drop: THREE.Object3D[] = [];
  root.traverse((obj) => {
    if (obj.userData[MARK]) drop.push(obj);
  });
  for (const obj of drop) obj.removeFromParent();
}

function tagged(mesh: THREE.Object3D): THREE.Object3D {
  mesh.userData[MARK] = true;
  mesh.traverse((c) => {
    c.userData[MARK] = true;
  });
  return mesh;
}

function cap(): THREE.Object3D {
  const m = new THREE.Mesh(
    new THREE.SphereGeometry(0.2, 10, 8, 0, Math.PI * 2, 0, Math.PI / 2),
    new THREE.MeshStandardMaterial({ color: 0xc4a574, roughness: 0.7 }),
  );
  m.position.y = 0.22;
  return tagged(m);
}

function chestPlate(): THREE.Object3D {
  const m = new THREE.Mesh(
    new THREE.SphereGeometry(0.22, 12, 10, 0, Math.PI * 2, 0.35, 1.6),
    new THREE.MeshStandardMaterial({ color: 0x3f6b4e, roughness: 0.42, metalness: 0.38 }),
  );
  m.scale.set(1.15, 1.05, 0.78);
  m.position.set(0, 0.3, 0.02);
  return tagged(m);
}

function bootCuff(): THREE.Object3D {
  const m = new THREE.Mesh(
    new THREE.CylinderGeometry(0.09, 0.1, 0.12, 8),
    new THREE.MeshStandardMaterial({ color: 0x8fb9c9, roughness: 0.5 }),
  );
  m.position.y = -0.55;
  return tagged(m);
}

function staff(): THREE.Object3D {
  const g = new THREE.Group();
  const pole = new THREE.Mesh(
    new THREE.CylinderGeometry(0.03, 0.035, 1.15, 6),
    new THREE.MeshStandardMaterial({ color: 0x6b4423, roughness: 0.7 }),
  );
  pole.position.set(0.08, -0.15, 0.18);
  const gem = new THREE.Mesh(
    new THREE.OctahedronGeometry(0.08, 0),
    new THREE.MeshStandardMaterial({ color: 0x67e8f9, emissive: 0x22d3ee, emissiveIntensity: 0.8 }),
  );
  gem.position.set(0.08, 0.48, 0.18);
  g.add(pole, gem);
  return tagged(g);
}

function gloveBand(): THREE.Object3D {
  const m = new THREE.Mesh(
    new THREE.TorusGeometry(0.07, 0.018, 6, 10),
    new THREE.MeshStandardMaterial({ color: 0xd4a017, metalness: 0.5, roughness: 0.35 }),
  );
  m.position.set(0.02, -0.42, 0);
  m.rotation.x = Math.PI / 2;
  return tagged(m);
}

function pendant(): THREE.Object3D {
  const m = new THREE.Mesh(
    new THREE.OctahedronGeometry(0.05, 0),
    new THREE.MeshStandardMaterial({ color: 0xfbbf24, emissive: 0xb45309, emissiveIntensity: 0.5 }),
  );
  m.position.set(0, -0.18, 0.16);
  return tagged(m);
}

function bracelet(): THREE.Object3D {
  const m = new THREE.Mesh(
    new THREE.TorusGeometry(0.06, 0.014, 6, 10),
    new THREE.MeshStandardMaterial({ color: 0x7dd3fc, metalness: 0.4, roughness: 0.4 }),
  );
  m.position.set(0.02, -0.28, 0);
  m.rotation.x = Math.PI / 2;
  return tagged(m);
}

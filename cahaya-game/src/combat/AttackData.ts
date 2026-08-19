import type { AttackKind } from "./CombatState";

export type DamageType = "PHYSICAL" | "ENERGY" | "SPECIAL" | "BOSS";

export interface AttackData {
  id: string;
  name: string;
  kind: AttackKind;
  damage: number;
  range: number;
  staminaCost: number;
  energyCost: number;
  startup: number;
  active: number;
  recovery: number;
  cooldown: number;
  knockback: number;
  stunDuration: number;
  hitboxRadius: number;
  hitboxHeight: number;
  damageType: DamageType;
  critChance: number;
  hitStop: number;
  projectile: boolean;
}

export const ATTACKS = {
  PUNCH_1: {
    id: "PUNCH_1",
    name: "Pukulan",
    kind: "punch",
    damage: 10,
    range: 1.5,
    staminaCost: 0,
    energyCost: 0,
    startup: 0.1,
    active: 0.1,
    recovery: 0.15,
    cooldown: 0.18,
    knockback: 2.2,
    stunDuration: 0.12,
    hitboxRadius: 0.8,
    hitboxHeight: 0.9,
    damageType: "PHYSICAL",
    critChance: 0.05,
    hitStop: 0.03,
    projectile: false,
  },
  PUNCH_2: {
    id: "PUNCH_2",
    name: "Pukulan Lanjut",
    kind: "punch",
    damage: 12,
    range: 1.55,
    staminaCost: 0,
    energyCost: 0,
    startup: 0.08,
    active: 0.1,
    recovery: 0.14,
    cooldown: 0.16,
    knockback: 2.6,
    stunDuration: 0.14,
    hitboxRadius: 0.85,
    hitboxHeight: 0.9,
    damageType: "PHYSICAL",
    critChance: 0.05,
    hitStop: 0.04,
    projectile: false,
  },
  KICK_1: {
    id: "KICK_1",
    name: "Tendangan",
    kind: "kick",
    damage: 15,
    range: 1.7,
    staminaCost: 0,
    energyCost: 0,
    startup: 0.12,
    active: 0.12,
    recovery: 0.21,
    cooldown: 0.22,
    knockback: 4.2,
    stunDuration: 0.2,
    hitboxRadius: 1.0,
    hitboxHeight: 0.95,
    damageType: "PHYSICAL",
    critChance: 0.05,
    hitStop: 0.05,
    projectile: false,
  },
  KICK_2: {
    id: "KICK_2",
    name: "Tendangan Lanjut",
    kind: "kick",
    damage: 18,
    range: 1.8,
    staminaCost: 0,
    energyCost: 0,
    startup: 0.1,
    active: 0.12,
    recovery: 0.2,
    cooldown: 0.2,
    knockback: 5.0,
    stunDuration: 0.24,
    hitboxRadius: 1.05,
    hitboxHeight: 1.0,
    damageType: "PHYSICAL",
    critChance: 0.06,
    hitStop: 0.06,
    projectile: false,
  },
  ENERGY_1: {
    id: "ENERGY_1",
    name: "Serangan Energi",
    kind: "energy",
    damage: 25,
    range: 18,
    staminaCost: 0,
    energyCost: 20,
    startup: 0.16,
    active: 0.12,
    recovery: 0.28,
    cooldown: 0.45,
    knockback: 3.4,
    stunDuration: 0.18,
    hitboxRadius: 0.5,
    hitboxHeight: 0.5,
    damageType: "ENERGY",
    critChance: 0.05,
    hitStop: 0.045,
    projectile: true,
  },
  DODGE: {
    id: "DODGE",
    name: "Elakan",
    kind: "dodge",
    damage: 0,
    range: 0,
    staminaCost: 20,
    energyCost: 0,
    startup: 0,
    active: 0.4,
    recovery: 0,
    cooldown: 0.55,
    knockback: 0,
    stunDuration: 0,
    hitboxRadius: 0,
    hitboxHeight: 0,
    damageType: "PHYSICAL",
    critChance: 0,
    hitStop: 0,
    projectile: false,
  },
  SPECIAL_1: {
    id: "SPECIAL_1",
    name: "Serangan Khusus",
    kind: "special",
    damage: 40,
    range: 2.2,
    staminaCost: 0,
    energyCost: 40,
    startup: 0.2,
    active: 0.16,
    recovery: 0.4,
    cooldown: 1.2,
    knockback: 6,
    stunDuration: 0.4,
    hitboxRadius: 1.2,
    hitboxHeight: 1.2,
    damageType: "SPECIAL",
    critChance: 0.08,
    hitStop: 0.08,
    projectile: false,
  },
  ULTIMATE: {
    id: "ULTIMATE",
    name: "Ultimate",
    kind: "ultimate",
    damage: 80,
    range: 4,
    staminaCost: 0,
    energyCost: 80,
    startup: 0.4,
    active: 0.3,
    recovery: 0.8,
    cooldown: 8,
    knockback: 8,
    stunDuration: 0.8,
    hitboxRadius: 2,
    hitboxHeight: 2,
    damageType: "SPECIAL",
    critChance: 0.1,
    hitStop: 0.08,
    projectile: false,
  },
} as const satisfies Record<string, AttackData>;

export type AttackId = keyof typeof ATTACKS;

export function attackDuration(data: AttackData): number {
  return data.startup + data.active + data.recovery;
}

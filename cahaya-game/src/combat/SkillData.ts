export interface SkillDef {
  id: string;
  name: string;
  type: "PHYSICAL" | "ENERGY" | string;
  damage: number;
  energyCost: number;
  staminaCost: number;
  cooldown: number;
  range: number;
  radius: number;
  requiredLevel: number;
  animation: string;
  vfx: string;
}

export const SKILLS: readonly SkillDef[] = [
  {
    id: "power_strike",
    name: "LIGHT STRIKE",
    type: "PHYSICAL",
    damage: 22,
    energyCost: 10,
    staminaCost: 0,
    cooldown: 3,
    range: 2.2,
    radius: 0,
    requiredLevel: 1,
    animation: "PUNCH",
    vfx: "strike",
  },
  {
    id: "flash_step",
    name: "FLASH STEP",
    type: "PHYSICAL",
    damage: 0,
    energyCost: 12,
    staminaCost: 8,
    cooldown: 5,
    range: 0,
    radius: 0,
    requiredLevel: 1,
    animation: "DODGE",
    vfx: "dash",
  },
  {
    id: "dawn_wave",
    name: "DAWN WAVE",
    type: "ENERGY",
    damage: 20,
    energyCost: 28,
    staminaCost: 0,
    cooldown: 8,
    range: 1.4,
    radius: 2.8,
    requiredLevel: 1,
    animation: "KICK",
    vfx: "wave",
  },
  {
    id: "light_burst",
    name: "LIGHT BURST",
    type: "ENERGY",
    damage: 26,
    energyCost: 18,
    staminaCost: 0,
    cooldown: 5,
    range: 16,
    radius: 0,
    requiredLevel: 1,
    animation: "ENERGY_ATTACK",
    vfx: "burst",
  },
  {
    id: "celestial_impact",
    name: "FINAL LIGHT",
    type: "ENERGY",
    damage: 80,
    energyCost: 100,
    staminaCost: 0,
    cooldown: 120,
    range: 8,
    radius: 5,
    requiredLevel: 20,
    animation: "ENERGY_ATTACK",
    vfx: "ultimate",
  },
];

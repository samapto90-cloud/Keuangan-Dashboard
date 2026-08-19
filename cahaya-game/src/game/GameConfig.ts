export const GAME_TITLE = "Petualangan Menuju Cahaya";
export const GAME_VERSION = "1.0.0-beta";
export const GAME_PHASE = "30/30";
export const PLAYER_NAME = "Raka";

/** Debug overlay hanya di development atau ?debug=1 — tidak di production. */
export const DEBUG_MODE =
  import.meta.env.DEV || new URLSearchParams(window.location.search).get("debug") === "1";

export const PLAYER = {
  walkSpeed: 3,
  runSpeed: 6,
  jumpForce: 7.2,
  gravity: 22,
  acceleration: 28,
  deceleration: 22,
  turnSpeed: 10,
  height: 1.7,
  radius: 0.38,
  worldLimit: 200,
  groundY: 0,
} as const;

export const STAMINA = {
  max: 100,
  sprintDrainPerSec: 5,
  recoverPerSec: 15,
  minToSprint: 4,
} as const;

export const CAMERA = {
  distance: 7.4,
  height: 3.85,
  lookY: 1.28,
  minPitch: -0.28,
  maxPitch: 0.92,
  rotateSpeed: 0.0024,
  smoothness: 7.2,
  minDistance: 4.2,
  maxDistance: 12,
  collisionPadding: 0.28,
} as const;

export const GRAPHICS = {
  LOW: { shadow: false, particles: 0.25, npcDist: 18, envDetail: 0.4, shadowSize: 512, bloom: false },
  MEDIUM: { shadow: true, particles: 0.6, npcDist: 28, envDetail: 0.7, shadowSize: 1024, bloom: false },
  HIGH: { shadow: true, particles: 1, npcDist: 36, envDetail: 1, shadowSize: 2048, bloom: true },
  ULTRA: { shadow: true, particles: 1.2, npcDist: 48, envDetail: 1.2, shadowSize: 2048, bloom: true },
} as const;

export type GraphicsQuality = keyof typeof GRAPHICS;
export type GraphicsPreset = GraphicsQuality | "AUTO";

export function resolveGraphics(preset: GraphicsPreset): GraphicsQuality {
  if (preset !== "AUTO") return preset;
  const nav = navigator as Navigator & { deviceMemory?: number };
  const mem = nav.deviceMemory ?? 8;
  const cores = navigator.hardwareConcurrency ?? 4;
  const mobile = window.matchMedia("(pointer: coarse)").matches || /Mobi|Android|iPhone|iPad/i.test(navigator.userAgent);
  if (mobile || mem <= 4 || cores <= 4) return "LOW";
  if (mem <= 6 || cores <= 6) return "MEDIUM";
  if (mem >= 12 && cores >= 8) return "ULTRA";
  return "HIGH";
}

export const WORLD = {
  groundSize: 90,
  fogColor: 0x8fb4c8,
  skyTop: 0xb8e0ef,
  skyHorizon: 0xf4d9a8,
} as const;

export const KEYBINDS = {
  attack: "LMB",
  heavy: "RMB",
  skill1: "Digit1",
  skill2: "Digit2",
  skill3: "Digit3",
  skill4: "Digit4",
  ultimate: "KeyR",
  dodge: "Space",
  transform: "KeyT",
  jump: "KeyF",
  mount: "KeyM",
} as const;

export const COMBAT = {
  comboWindow: 0.8,
  comboReset: 2.0,
  critChance: 0.05,
  critMultiplier: 1.5,
  energyRecoverPerSec: 8,
  dodgeDuration: 0.4,
  dodgeIFrame: 0.25,
  dodgeCooldown: 0.55,
  dodgeStamina: 20,
  dodgeSpeed: 14,
  lockCameraDistance: 8.4,
  shakeMax: 0.22,
  hpRegenPctPerSec: 0.01,
  respawnSec: 5,
  basicAttackRange: 1.6,
  punchRange: 1.55,
  kickRange: 1.8,
  skillRange: 18,
  targetSearchRange: 12,
  autoTarget: true,
} as const;

export const NET = {
  tickRate: 20,
  inputHz: 20,
  pingIntervalMs: 3000,
  reconnectMs: 1800,
  correctSoft: 0.45,
  correctHard: 8,
  interpDuration: 0.1,
  nameplateMaxDist: 28,
  walkAnimSpeed: 0.35,
  runAnimSpeed: 4.5,
} as const;

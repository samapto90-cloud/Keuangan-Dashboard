export const BOARD_GRID = 10;
export const BOARD_SIZE = 100;
export const MIN_POSITION = 1;
export const MAX_POSITION = 100;
export const MAX_PLAYERS = 8;
export const MOVE_DURATION = 200;
export const SNAKE_DURATION = 720;
export const LADDER_DURATION = 720;
export const DICE_DURATION = 650;

export type BoardConfig = {
  size: number;
  totalCells: number;
  snakes: Record<number, number>;
  ladders: Record<number, number>;
};

export const SNAKES: Record<number, number> = {
  // Ular (tanpa gambar) — jalur terpisah agar tidak numpuk dengan tangga
  97: 75,
  93: 55,
  87: 52,
  66: 34,
  57: 24,
  43: 18,
  35: 12,
};

export const LADDERS: Record<number, number> = {
  // Tangga dimatikan — hanya ular
};

export const DEFAULT_BOARD: BoardConfig = {
  size: BOARD_GRID,
  totalCells: BOARD_SIZE,
  snakes: { ...SNAKES },
  ladders: { ...LADDERS },
};

export const PLAYER_COLORS = [
  "#2563eb",
  "#0ea5e9",
  "#ef4444",
  "#3b82f6",
  "#dc2626",
  "#059669",
  "#8b5cf6",
  "#f59e0b",
] as const;
export const PLAYER_COLOR_NAMES = [
  "NAVY",
  "RED",
  "GREEN",
  "BLUE",
  "ORANGE",
  "PURPLE",
  "TEAL",
  "GOLD",
] as const;

export function zoneFor(position: number): string {
  if (position <= 20) return "desa";
  if (position <= 40) return "hutan";
  if (position <= 60) return "gunung";
  if (position <= 80) return "candi";
  return "puncak";
}

export const BOARD_GRID = 10;
export const BOARD_SIZE = 100;
export const MIN_POSITION = 1;
export const MAX_POSITION = 100;
export const MAX_PLAYERS = 4;
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
  97: 78,
  95: 75,
  92: 71,
  88: 48,
  87: 36,
  62: 18,
  54: 34,
  17: 7,
};

export const LADDERS: Record<number, number> = {
  4: 25,
  9: 31,
  20: 42,
  28: 55,
  40: 59,
  51: 73,
  63: 81,
  70: 93,
};

export const DEFAULT_BOARD: BoardConfig = {
  size: BOARD_GRID,
  totalCells: BOARD_SIZE,
  snakes: { ...SNAKES },
  ladders: { ...LADDERS },
};

/* Order matches mockup seats: Blue, Green, Yellow, Red */
export const PLAYER_COLORS = ["#3498db", "#2ecc71", "#f1c40f", "#e74c3c"] as const;
export const PLAYER_COLOR_NAMES = ["BLUE", "GREEN", "YELLOW", "RED"] as const;

export function zoneFor(position: number): string {
  if (position <= 20) return "desa";
  if (position <= 40) return "hutan";
  if (position <= 60) return "gunung";
  if (position <= 80) return "candi";
  return "puncak";
}

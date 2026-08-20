import {
  BOARD_GRID,
  DEFAULT_BOARD,
  MAX_POSITION,
  MIN_POSITION,
  type BoardConfig,
} from "./config";

/** Posisi di luar papan sebelum masuk kotak 1. */
export const OFFBOARD_START = 0;

export type CellCoord = { row: number; col: number };

export type MoveResult = {
  dice: number;
  walkPath: number[];
  walkFinal: number;
  snakeFrom?: number;
  snakeTo?: number;
  ladderFrom?: number;
  ladderTo?: number;
  final: number;
  reached100: boolean;
  bounced: boolean;
  log: string;
};

/** Area start kiri (fraksi lebar playfield). */
export const GRID_INSET_LEFT = 0.1;
export const GRID_WIDTH_FRAC = 0.9;

export function getCellCoordinate(position: number): CellCoord | null {
  if (position === OFFBOARD_START) return { row: 0, col: -1 };
  if (position < MIN_POSITION || position > MAX_POSITION) return null;
  const idx = position - 1;
  const row = Math.floor(idx / BOARD_GRID);
  const offset = idx % BOARD_GRID;
  const col = row % 2 === 1 ? BOARD_GRID - 1 - offset : offset;
  return { row, col };
}

export function getPositionFromCoordinate(row: number, col: number): number | null {
  if (row < 0 || row >= BOARD_GRID || col < 0 || col >= BOARD_GRID) return null;
  const offset = row % 2 === 1 ? BOARD_GRID - 1 - col : col;
  return row * BOARD_GRID + offset + 1;
}

/** Persen left/bottom di token-layer (termasuk zona start kiri). */
export function tokenBoardPercent(
  position: number,
  off: { dx: number; dy: number } = { dx: 0, dy: 0 },
): { left: number; bottom: number } | null {
  if (position <= OFFBOARD_START) {
    // Zona start di kiri luar kotak 1
    return {
      left: (0.045 + off.dx * 0.035) * 100,
      bottom: (0.06 + off.dy * 0.04) * 100,
    };
  }
  const coord = getCellCoordinate(position);
  if (!coord || coord.col < 0) return null;
  const left = (GRID_INSET_LEFT + ((coord.col + 0.5 + off.dx) / BOARD_GRID) * GRID_WIDTH_FRAC) * 100;
  const bottom = ((coord.row + 0.5 + off.dy) / BOARD_GRID) * 100;
  return { left, bottom };
}

/** Koordinat SVG viewBox 0–100 (selaras token-layer). */
export function boardSvgPoint(position: number): { x: number; y: number } | null {
  if (position <= OFFBOARD_START) {
    return { x: 4.5, y: 94 };
  }
  const coord = getCellCoordinate(position);
  if (!coord || coord.col < 0) return null;
  const x = (GRID_INSET_LEFT + ((coord.col + 0.5) / BOARD_GRID) * GRID_WIDTH_FRAC) * 100;
  const y = 100 - ((coord.row + 0.5) / BOARD_GRID) * 100;
  return { x, y };
}

export function validateBoardConfig(cfg: BoardConfig): string | null {
  if (cfg.size !== BOARD_GRID || cfg.totalCells !== MAX_POSITION) return "board size";
  const starts = new Map<number, string>();
  const dests = new Map<number, string>();
  for (const [fs, ts] of Object.entries(cfg.snakes)) {
    const from = Number(fs);
    const to = Number(ts);
    if (from <= to) return `snake ${from}`;
    if (starts.has(from) || dests.has(to)) return "duplicate";
    starts.set(from, "snake");
    dests.set(to, "snake");
  }
  for (const [fs, ts] of Object.entries(cfg.ladders)) {
    const from = Number(fs);
    const to = Number(ts);
    if (to <= from) return `ladder ${from}`;
    if (starts.has(from) || dests.has(to)) return "duplicate";
    starts.set(from, "ladder");
    dests.set(to, "ladder");
  }
  for (const [fs, ts] of Object.entries(cfg.snakes)) {
    if (starts.get(Number(ts)) === "snake") return `snake chain ${fs}`;
  }
  for (const [fs, ts] of Object.entries(cfg.ladders)) {
    if (starts.get(Number(ts)) === "ladder") return `ladder chain ${fs}`;
  }
  return null;
}

export function walkPath(from: number, dice: number): { path: number[]; bounced: boolean } {
  const path: number[] = [];
  if (dice < 1 || dice > 6) return { path, bounced: false };
  if (from < OFFBOARD_START || from > MAX_POSITION) return { path, bounced: false };
  const over = from + dice - MAX_POSITION;
  if (over > 0) {
    const up = MAX_POSITION - from;
    let pos = from;
    for (let i = 0; i < up; i++) {
      pos += 1;
      path.push(pos);
    }
    for (let i = 0; i < over; i++) {
      pos -= 1;
      path.push(pos);
    }
    return { path, bounced: true };
  }
  let pos = from;
  for (let i = 0; i < dice; i++) {
    pos += 1;
    path.push(pos);
  }
  return { path, bounced: false };
}

export function resolveMove(cfg: BoardConfig, from: number, dice: number): MoveResult {
  const { path, bounced } = walkPath(from, dice);
  const walkFinal = path.length ? path[path.length - 1] : from;
  let final = walkFinal;
  const out: MoveResult = { dice, walkPath: path, walkFinal, final, bounced, reached100: false, log: "" };
  if (cfg.snakes[walkFinal]) {
    out.snakeFrom = walkFinal;
    out.snakeTo = cfg.snakes[walkFinal];
    final = cfg.snakes[walkFinal]!;
  } else if (cfg.ladders[walkFinal]) {
    out.ladderFrom = walkFinal;
    out.ladderTo = cfg.ladders[walkFinal];
    final = cfg.ladders[walkFinal]!;
  }
  out.final = final;
  out.reached100 = final >= MAX_POSITION;
  out.log = `dice=${dice} ${from}->${walkFinal}` + (final !== walkFinal ? `=>${final}` : "");
  return out;
}

export function isSnake(cfg: BoardConfig, pos: number): boolean {
  return Boolean(cfg.snakes[pos]);
}

export function snakeDest(cfg: BoardConfig, pos: number): number | undefined {
  return cfg.snakes[pos];
}

export const snakeTarget = snakeDest;

export function isLadder(cfg: BoardConfig, pos: number): boolean {
  return Boolean(cfg.ladders[pos]);
}

export function ladderDest(cfg: BoardConfig, pos: number): number | undefined {
  return cfg.ladders[pos];
}

export const ladderTarget = ladderDest;

export function nextTurnIndex(current: number, playerCount: number): number {
  if (playerCount <= 0) return 0;
  return (current + 1) % playerCount;
}

/** Offset antrean pemain di kotak yang sama — berdiri berdampingan, sedikit depth. */
export function tokenOffsets(count: number, index: number): { dx: number; dy: number } {
  if (count <= 1) return { dx: 0, dy: 0.1 };
  if (count === 2) {
    return index === 0 ? { dx: -0.28, dy: 0.1 } : { dx: 0.28, dy: 0.08 };
  }
  if (count === 3) {
    const pts = [
      { dx: -0.32, dy: 0.12 },
      { dx: 0.02, dy: 0.02 },
      { dx: 0.32, dy: 0.1 },
    ];
    return pts[index % 3]!;
  }
  const pts = [
    { dx: -0.3, dy: 0.14 },
    { dx: 0.3, dy: 0.12 },
    { dx: -0.16, dy: -0.06 },
    { dx: 0.16, dy: -0.08 },
  ];
  return pts[index % 4]!;
}

export function tokenCrowdScale(count: number): number {
  if (count <= 1) return 1;
  if (count === 2) return 0.9;
  if (count === 3) return 0.82;
  return 0.74;
}

export function rollDiceLocal(): number {
  const buf = new Uint8Array(1);
  crypto.getRandomValues(buf);
  return (buf[0] % 6) + 1;
}

export const BoardEngine = {
  getCellCoordinate,
  getPositionFromCoordinate,
  tokenBoardPercent,
  boardSvgPoint,
  validateBoardConfig,
  walkPath,
  resolveMove,
  nextTurnIndex,
  tokenOffsets,
  tokenCrowdScale,
  defaultConfig: DEFAULT_BOARD,
  OFFBOARD_START,
};

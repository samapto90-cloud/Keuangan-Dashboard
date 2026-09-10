import {
  BOARD_GRID,
  DEFAULT_BOARD,
  MAX_POSITION,
  MIN_POSITION,
  type BoardConfig,
} from "./config";
import { stoneCenterPercent } from "./stonePath";

/** Posisi start: di luar nomor (0). Visual di batu 1; dadu N mendarat di N (lewat 1…N). */
export const OFFBOARD_START = 0;

/** Pakai layout batu di ilustrasi rumah tangga (bukan grid 10×10 kaku). */
export const USE_STONE_PATH = true;

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

/** Area start kiri (fraksi lebar playfield) — fallback grid. */
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

/** Barisan start di batu 1 (kotak pertama) — offset kerumunan di sekitar angka 1. */
export function startLineupPercent(count: number, index: number): { left: number; bottom: number } {
  const base = stoneCenterPercent(MIN_POSITION) ?? { left: 14.0, bottom: 24.0 };
  const off = tokenOffsets(Math.max(1, Math.min(8, count)), index);
  return {
    left: base.left + off.dx * 1.2,
    bottom: base.bottom + off.dy * 1.2,
  };
}

/** Persen left/bottom di token-layer (start = batu 1). */
export function tokenBoardPercent(
  position: number,
  off: { dx: number; dy: number } = { dx: 0, dy: 0 },
): { left: number; bottom: number } | null {
  if (position <= OFFBOARD_START) {
    const base = stoneCenterPercent(MIN_POSITION);
    if (!base) return startLineupPercent(1, 0);
    return {
      left: base.left + off.dx * 2.2,
      bottom: base.bottom + off.dy * 1.4,
    };
  }
  if (USE_STONE_PATH) {
    const c = stoneCenterPercent(position);
    if (!c) return null;
    return {
      left: c.left + off.dx * 2.2,
      bottom: c.bottom + off.dy * 1.4,
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
    const p = tokenBoardPercent(OFFBOARD_START);
    if (!p) return null;
    return { x: p.left, y: 100 - p.bottom };
  }
  if (USE_STONE_PATH) {
    const c = stoneCenterPercent(position);
    if (!c) return null;
    return { x: c.left, y: 100 - c.bottom };
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

/** Offset antrean pemain di kotak yang sama — tetap di dalam batu. */
export function tokenOffsets(count: number, index: number): { dx: number; dy: number } {
  if (count <= 1) return { dx: 0, dy: 0 };
  if (count === 2) {
    return index === 0 ? { dx: -0.22, dy: 0.02 } : { dx: 0.22, dy: -0.02 };
  }
  if (count === 3) {
    const pts = [
      { dx: -0.24, dy: 0.04 },
      { dx: 0.0, dy: -0.03 },
      { dx: 0.24, dy: 0.04 },
    ];
    return pts[index % 3]!;
  }
  if (count === 4) {
    const pts = [
      { dx: -0.2, dy: 0.06 },
      { dx: 0.2, dy: 0.06 },
      { dx: -0.2, dy: -0.06 },
      { dx: 0.2, dy: -0.06 },
    ];
    return pts[index % 4]!;
  }
  if (count === 5) {
    const pts = [
      { dx: -0.28, dy: 0.08 },
      { dx: 0.28, dy: 0.08 },
      { dx: 0.0, dy: 0.0 },
      { dx: -0.2, dy: -0.08 },
      { dx: 0.2, dy: -0.08 },
    ];
    return pts[index % 5]!;
  }
  if (count === 6) {
    const pts = [
      { dx: -0.28, dy: 0.08 },
      { dx: 0.0, dy: 0.08 },
      { dx: 0.28, dy: 0.08 },
      { dx: -0.28, dy: -0.08 },
      { dx: 0.0, dy: -0.08 },
      { dx: 0.28, dy: -0.08 },
    ];
    return pts[index % 6]!;
  }
  if (count === 7) {
    const pts = [
      { dx: -0.3, dy: 0.1 },
      { dx: -0.1, dy: 0.1 },
      { dx: 0.1, dy: 0.1 },
      { dx: 0.3, dy: 0.1 },
      { dx: -0.22, dy: -0.08 },
      { dx: 0.0, dy: -0.08 },
      { dx: 0.22, dy: -0.08 },
    ];
    return pts[index % 7]!;
  }
  const pts = [
    { dx: -0.3, dy: 0.1 },
    { dx: -0.1, dy: 0.1 },
    { dx: 0.1, dy: 0.1 },
    { dx: 0.3, dy: 0.1 },
    { dx: -0.3, dy: -0.1 },
    { dx: -0.1, dy: -0.1 },
    { dx: 0.1, dy: -0.1 },
    { dx: 0.3, dy: -0.1 },
  ];
  return pts[index % 8]!;
}

export function tokenCrowdScale(count: number): number {
  if (count <= 1) return 1;
  if (count === 2) return 0.92;
  if (count === 3) return 0.84;
  if (count === 4) return 0.76;
  if (count === 5) return 0.7;
  if (count === 6) return 0.64;
  if (count === 7) return 0.58;
  return 0.54;
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
  startLineupPercent,
  defaultConfig: DEFAULT_BOARD,
  OFFBOARD_START,
};

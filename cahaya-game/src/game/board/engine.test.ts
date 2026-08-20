import { DEFAULT_BOARD } from "./config";
import {
  getCellCoordinate,
  getPositionFromCoordinate,
  nextTurnIndex,
  OFFBOARD_START,
  resolveMove,
  tokenBoardPercent,
  tokenOffsets,
  validateBoardConfig,
} from "./engine";

function assert(cond: unknown, msg: string): void {
  if (!cond) throw new Error(msg);
}

export function runBoardUnitTests(): void {
  const c1 = getCellCoordinate(1);
  assert(c1 && c1.row === 0 && c1.col === 0, "coord 1");
  const c10 = getCellCoordinate(10);
  assert(c10 && c10.row === 0 && c10.col === 9, "coord 10");
  const c11 = getCellCoordinate(11);
  assert(c11 && c11.row === 1 && c11.col === 9, "coord 11");
  const c100 = getCellCoordinate(100);
  assert(c100 && c100.row === 9 && c100.col === 0, "coord 100");
  assert(getPositionFromCoordinate(0, 0) === 1, "pos 1");
  assert(getPositionFromCoordinate(9, 0) === 100, "pos 100");
  assert(validateBoardConfig(DEFAULT_BOARD) === null, "config");
  assert(resolveMove(DEFAULT_BOARD, 1, 1).walkFinal === 2, "1+1");
  const enter = resolveMove(DEFAULT_BOARD, OFFBOARD_START, 4);
  assert(enter.walkFinal === 4 && enter.walkPath.length === 4 && enter.walkPath[0] === 1, "0+4 enter");
  assert(resolveMove(DEFAULT_BOARD, OFFBOARD_START, 1).final === 1, "0+1");
  const startPct = tokenBoardPercent(OFFBOARD_START);
  assert(startPct && startPct.left < 10 && startPct.bottom < 20, "offboard percent");
  const cell1 = tokenBoardPercent(1);
  assert(cell1 && cell1.left > 10, "cell1 inset");
  assert(resolveMove(DEFAULT_BOARD, 10, 6).walkFinal === 16, "10+6");
  assert(resolveMove(DEFAULT_BOARD, 50, 6).walkFinal === 56, "50+6");
  assert(resolveMove(DEFAULT_BOARD, 98, 5).walkFinal === 97, "bounce 98+5");
  assert(resolveMove(DEFAULT_BOARD, 99, 2).walkFinal === 99, "bounce 99+2");
  assert(resolveMove(DEFAULT_BOARD, 97, 3).final === 100, "97+3");
  const snake = resolveMove(DEFAULT_BOARD, 87, 1);
  assert(snake.walkFinal === 88 && snake.snakeTo === 48 && snake.final === 48, "snake 88");
  const ladder = resolveMove(DEFAULT_BOARD, 27, 1);
  assert(ladder.walkFinal === 28 && ladder.ladderTo === 55 && ladder.final === 55, "ladder 28");
  assert(nextTurnIndex(0, 4) === 1 && nextTurnIndex(3, 4) === 0, "turn 4");
  assert(nextTurnIndex(0, 2) === 1 && nextTurnIndex(1, 2) === 0, "turn 2");
  const seen = new Set<string>();
  for (let i = 0; i < 4; i++) {
    const o = tokenOffsets(4, i);
    const k = `${o.dx},${o.dy}`;
    assert(!seen.has(k), "token overlap");
    seen.add(k);
  }
  assert(resolveMove(DEFAULT_BOARD, 99, 1).reached100, "finish");
}

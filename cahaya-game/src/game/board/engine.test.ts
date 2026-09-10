import { DEFAULT_BOARD } from "./config";
import {
  getCellCoordinate,
  getPositionFromCoordinate,
  nextTurnIndex,
  OFFBOARD_START,
  resolveMove,
  tokenBoardPercent,
  tokenOffsets,
  startLineupPercent,
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
  assert(resolveMove(DEFAULT_BOARD, 1, 4).walkFinal === 5 && resolveMove(DEFAULT_BOARD, 1, 4).walkPath[0] === 2, "from1 +4 → 2..5");
  const enter = resolveMove(DEFAULT_BOARD, OFFBOARD_START, 4);
  assert(enter.walkFinal === 4 && enter.walkPath.length === 4 && enter.walkPath[0] === 1, "start+4 → 1..4");
  assert(resolveMove(DEFAULT_BOARD, OFFBOARD_START, 1).final === 1, "start+1 → 1");
  const startPct = tokenBoardPercent(OFFBOARD_START);
  const cell1 = tokenBoardPercent(1);
  assert(startPct && cell1 && Math.abs(startPct.left - cell1.left) < 0.01, "start sits on stone 1");
  assert(cell1 && cell1.left > 12 && cell1.left < 16 && cell1.bottom > 16 && cell1.bottom < 26, "cell1 stone");
  const line2a = startLineupPercent(2, 0);
  const line2b = startLineupPercent(2, 1);
  assert(line2a.bottom > 16 && line2b.left !== line2a.left, "start lineup around cell 1");
  assert(tokenBoardPercent(100) && tokenBoardPercent(100)!.bottom > 60, "cell100 top-left");
  assert(tokenBoardPercent(1) && tokenBoardPercent(1)!.bottom > 16 && tokenBoardPercent(1)!.bottom < 26, "cell1 bottom");
  const cell10 = tokenBoardPercent(10);
  assert(cell10 && cell10.left > 85, "cell10 right");
  assert(resolveMove(DEFAULT_BOARD, 10, 6).walkFinal === 16, "10+6");
  assert(resolveMove(DEFAULT_BOARD, 50, 6).walkFinal === 56, "50+6");
  assert(resolveMove(DEFAULT_BOARD, 98, 5).walkFinal === 97, "bounce 98+5");
  assert(resolveMove(DEFAULT_BOARD, 99, 2).walkFinal === 99, "bounce 99+2");
  assert(resolveMove(DEFAULT_BOARD, 97, 3).final === 100, "97+3");
  const snake = resolveMove(DEFAULT_BOARD, 96, 1);
  assert(snake.walkFinal === 97 && snake.snakeTo === 75 && snake.final === 75, "snake 97");
  const noLadder = resolveMove(DEFAULT_BOARD, 2, 1);
  assert(noLadder.walkFinal === 3 && !noLadder.ladderTo && noLadder.final === 3, "no ladder");
  const plain = resolveMove(DEFAULT_BOARD, 77, 1);
  assert(plain.walkFinal === 78 && !plain.ladderTo && plain.final === 78, "no jackpot ladder");
  assert(nextTurnIndex(0, 4) === 1 && nextTurnIndex(3, 4) === 0, "turn 4");
  assert(nextTurnIndex(0, 2) === 1 && nextTurnIndex(1, 2) === 0, "turn 2");
  for (const n of [2, 3, 4, 5, 6, 7, 8]) {
    const seen = new Set<string>();
    for (let i = 0; i < n; i++) {
      const o = tokenOffsets(n, i);
      const k = `${o.dx},${o.dy}`;
      assert(!seen.has(k), `token overlap n=${n}`);
      seen.add(k);
    }
  }
  assert(resolveMove(DEFAULT_BOARD, 99, 1).reached100, "finish");
}

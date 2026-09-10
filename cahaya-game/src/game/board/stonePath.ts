/**
 * Koordinat batu 1–100 — board-gunung.png (1024×1024).
 * Pusat diukur dari angka putih; ukuran hampir penuh permukaan batu.
 * Seni & hotspot harus dalam .board-playfield yang sama.
 */

export type StoneRect = {
  left: number;
  top: number;
  w: number;
  h: number;
};

/** Geser seluruh hotspot ke bawah (% tinggi papan) agar pas dengan batu di ilustrasi. */
const SHIFT_Y = 1.45;

/** Pusat X kiri→kanan visual per baris (0=bawah … 9=atas). */
const ROW_X: ReadonlyArray<ReadonlyArray<number>> = [
  [14.0, 22.7, 31.5, 40.0, 48.8, 57.3, 65.9, 74.4, 83.0, 91.8],
  [15.1, 23.7, 32.1, 40.4, 48.9, 57.2, 65.4, 73.7, 81.9, 90.1],
  [16.2, 24.5, 32.8, 41.1, 49.6, 57.3, 65.2, 73.2, 81.2, 89.1],
  [17.5, 25.5, 33.4, 41.3, 49.2, 57.0, 64.7, 72.5, 80.2, 87.8],
  [17.8, 26.2, 33.9, 41.5, 49.1, 56.7, 64.2, 71.9, 79.5, 86.8],
  [19.4, 27.0, 34.5, 41.7, 49.1, 56.5, 63.9, 71.2, 78.5, 85.5],
  [19.7, 27.6, 34.9, 42.0, 49.1, 56.3, 63.4, 70.6, 77.8, 84.9],
  [21.3, 28.6, 35.6, 42.4, 49.4, 56.4, 63.3, 70.3, 77.1, 83.8],
  [22.3, 29.0, 35.8, 42.5, 49.2, 56.0, 62.7, 69.6, 76.3, 82.9],
  [23.2, 29.8, 36.4, 42.8, 49.4, 56.0, 62.6, 69.2, 75.7, 82.1],
];

/** Pusat Y angka (dari atas gambar). Baris 0 = kotak 1–10. */
const ROW_Y: ReadonlyArray<number> = [
  76.0, 70.2, 65.2, 59.5, 54.1, 48.9, 43.9, 39.0, 34.1, 29.3,
];

/** Hampir memenuhi batu; celah rumput ~2–4%. */
const FILL_W = 0.98;
const FILL_H = 0.92;

function localPitchX(row: number, col: number): number {
  const xs = ROW_X[row]!;
  if (col <= 0) return xs[1]! - xs[0]!;
  if (col >= 9) return xs[9]! - xs[8]!;
  return 0.5 * (xs[col]! - xs[col - 1]! + (xs[col + 1]! - xs[col]!));
}

function localPitchY(row: number): number {
  if (row <= 0) return ROW_Y[0]! - ROW_Y[1]!;
  if (row >= 9) return ROW_Y[8]! - ROW_Y[9]!;
  return 0.5 * (ROW_Y[row - 1]! - ROW_Y[row]! + (ROW_Y[row]! - ROW_Y[row + 1]!));
}

function buildStoneLayout(): StoneRect[] {
  const stones: StoneRect[] = [];
  for (let pos = 1; pos <= 100; pos++) {
    const idx = pos - 1;
    const row = Math.floor(idx / 10);
    const offset = idx % 10;
    const col = row % 2 === 1 ? 9 - offset : offset;
    const cx = ROW_X[row]![col]!;
    const cy = ROW_Y[row]! + SHIFT_Y;
    const w = localPitchX(row, col) * FILL_W;
    const h = localPitchY(row) * FILL_H;
    stones.push({
      left: cx - w / 2,
      top: cy - h / 2,
      w,
      h,
    });
  }
  return stones;
}

export const STONE_LAYOUT: StoneRect[] = buildStoneLayout();

export function stoneForPosition(position: number): StoneRect | null {
  if (position < 1 || position > 100) return null;
  return STONE_LAYOUT[position - 1] ?? null;
}

/** Titik angka putih di batu (setelah SHIFT_Y). */
export function stoneNumberPercent(position: number): { left: number; bottom: number } | null {
  if (position < 1 || position > 100) return null;
  const idx = position - 1;
  const row = Math.floor(idx / 10);
  const offset = idx % 10;
  const col = row % 2 === 1 ? 9 - offset : offset;
  const cx = ROW_X[row]![col]!;
  const cy = ROW_Y[row]! + SHIFT_Y;
  return {
    left: cx,
    bottom: 100 - cy,
  };
}

/**
 * Pusat pion = tengah hotspot batu (ikut geser ke bawah bersama kotak).
 */
export function stoneCenterPercent(position: number): { left: number; bottom: number } | null {
  const stone = stoneForPosition(position);
  if (!stone) return null;
  return {
    left: stone.left + stone.w / 2,
    bottom: 100 - (stone.top + stone.h * 0.5),
  };
}

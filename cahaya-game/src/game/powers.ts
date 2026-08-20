export type PowerKind = "bomb" | "thunder" | "superman";

export type PowerBag = {
  bomb: number;
  thunder: number;
  superman: number;
};

/** Stok di papan per match. */
export const POWER_BOARD_LIMITS: Record<PowerKind, number> = {
  bomb: 1,
  thunder: 5,
  superman: 3,
};

export const POWER_META: Record<
  PowerKind,
  { label: string; icon: string; hint: string; needsTarget: boolean; allowSelf: boolean }
> = {
  bomb: {
    label: "Bom",
    icon: "💣",
    hint: "Lempar ke lawan — keluar papan & mulai dari START. (1 di papan)",
    needsTarget: true,
    allowSelf: false,
  },
  thunder: {
    label: "Petir",
    icon: "⚡",
    hint: "Turunkan lawan 3 langkah. (5 di papan)",
    needsTarget: true,
    allowSelf: false,
  },
	superman: {
    label: "Pesawat",
    icon: "✈️",
    hint: "Terbang naik 3 kotak — diri sendiri atau bantu teman. (3 di papan)",
    needsTarget: true,
    allowSelf: true,
  },
};

export function emptyBag(): PowerBag {
  return { bomb: 0, thunder: 0, superman: 0 };
}

export function bagTotal(bag: PowerBag): number {
  return bag.bomb + bag.thunder + bag.superman;
}

export function addPower(bag: PowerBag, kind: PowerKind): PowerBag {
  return { ...bag, [kind]: bag[kind] + 1 };
}

export function takePower(bag: PowerBag, kind: PowerKind): PowerBag | null {
  if (bag[kind] <= 0) return null;
  return { ...bag, [kind]: bag[kind] - 1 };
}

export function pointsFromPosition(position: number): number {
  return Math.max(0, position);
}

export function applyBomb(_from: number): number {
  return 0;
}

export function applyThunder(from: number): number {
  return Math.max(0, from - 3);
}

export function applySuperman(from: number): number {
  return Math.min(100, from + 3);
}

export function powerLabel(kind: PowerKind): string {
  return `${POWER_META[kind].icon} ${POWER_META[kind].label}`;
}

function shuffleInPlace<T>(arr: T[]): T[] {
  for (let i = arr.length - 1; i > 0; i--) {
    const buf = new Uint32Array(1);
    crypto.getRandomValues(buf);
    const j = buf[0]! % (i + 1);
    const tmp = arr[i]!;
    arr[i] = arr[j]!;
    arr[j] = tmp;
  }
  return arr;
}

/** Sebar item di kotak acak (hindari 1, 100, ular/tangga start+dest). */
export function spawnPowerCells(
  snakes: Record<number, number>,
  ladders: Record<number, number>,
): Record<number, PowerKind> {
  const blocked = new Set<number>([1, 100]);
  for (const [a, b] of Object.entries(snakes)) {
    blocked.add(Number(a));
    blocked.add(Number(b));
  }
  for (const [a, b] of Object.entries(ladders)) {
    blocked.add(Number(a));
    blocked.add(Number(b));
  }
  const pool: number[] = [];
  for (let p = 2; p <= 99; p++) {
    if (!blocked.has(p)) pool.push(p);
  }
  shuffleInPlace(pool);
  const out: Record<number, PowerKind> = {};
  let i = 0;
  const place = (kind: PowerKind, n: number): void => {
    for (let k = 0; k < n && i < pool.length; k++, i++) {
      out[pool[i]!] = kind;
    }
  };
  place("bomb", POWER_BOARD_LIMITS.bomb);
  place("thunder", POWER_BOARD_LIMITS.thunder);
  place("superman", POWER_BOARD_LIMITS.superman);
  return out;
}

export function powerIcon(kind: PowerKind): string {
  return POWER_META[kind].icon;
}

import { boardSvgPoint } from "../game/board/engine";

type Pt = { x: number; y: number };

const CELL = 9;

function dist(a: Pt, b: Pt): number {
  return Math.hypot(b.x - a.x, b.y - a.y) || 1;
}

function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t;
}

/** Kurva S alami (2 gelombang) — tubuh ular yang melilit. */
function naturalSpine(a: Pt, b: Pt, twist: number): Pt[] {
  const dx = b.x - a.x;
  const dy = b.y - a.y;
  const len = Math.hypot(dx, dy) || 1;
  const nx = -dy / len;
  const ny = dx / len;
  const sign = twist % 2 === 0 ? 1 : -1;
  const amp = Math.min(CELL * 0.55, Math.max(CELL * 0.22, len * 0.16)) * sign;
  const n = Math.max(18, Math.round(len / 1.8));
  const pts: Pt[] = [];
  for (let i = 0; i <= n; i++) {
    const t = i / n;
    const ease = t * t * (3 - 2 * t);
    // dua lengkung S + sedikit noise deterministik
    const wiggle =
      Math.sin(Math.PI * t * 2) * amp * (0.55 + 0.45 * Math.sin(Math.PI * t)) +
      Math.sin(Math.PI * t * 3.2 + twist) * amp * 0.18;
    pts.push({
      x: a.x + dx * ease + nx * wiggle,
      y: a.y + dy * ease + ny * wiggle,
    });
  }
  return pts;
}

function catmull(pts: Pt[]): string {
  if (pts.length < 2) return "";
  if (pts.length === 2) return `M ${pts[0]!.x} ${pts[0]!.y} L ${pts[1]!.x} ${pts[1]!.y}`;
  let d = `M ${pts[0]!.x.toFixed(2)} ${pts[0]!.y.toFixed(2)}`;
  for (let i = 0; i < pts.length - 1; i++) {
    const p0 = pts[Math.max(0, i - 1)]!;
    const p1 = pts[i]!;
    const p2 = pts[i + 1]!;
    const p3 = pts[Math.min(pts.length - 1, i + 2)]!;
    const c1x = p1.x + (p2.x - p0.x) / 6;
    const c1y = p1.y + (p2.y - p0.y) / 6;
    const c2x = p2.x - (p3.x - p1.x) / 6;
    const c2y = p2.y - (p3.y - p1.y) / 6;
    d += ` C ${c1x.toFixed(2)} ${c1y.toFixed(2)} ${c2x.toFixed(2)} ${c2y.toFixed(2)} ${p2.x.toFixed(2)} ${p2.y.toFixed(2)}`;
  }
  return d;
}

function ribbon(spine: Pt[], headW: number, tailW: number): string {
  if (spine.length < 2) return "";
  const left: Pt[] = [];
  const right: Pt[] = [];
  for (let i = 0; i < spine.length; i++) {
    const p = spine[i]!;
    const prev = spine[Math.max(0, i - 1)]!;
    const next = spine[Math.min(spine.length - 1, i + 1)]!;
    const tx = next.x - prev.x;
    const ty = next.y - prev.y;
    const tl = Math.hypot(tx, ty) || 1;
    const nx = -ty / tl;
    const ny = tx / tl;
    const t = i / (spine.length - 1);
    // kepala lebih tebal, mengecil ke ekor secara natural
    const taper = Math.pow(t, 1.15);
    const w = lerp(headW, tailW, taper);
    left.push({ x: p.x + nx * w, y: p.y + ny * w });
    right.push({ x: p.x - nx * w, y: p.y - ny * w });
  }
  const f = (p: Pt) => `${p.x.toFixed(2)} ${p.y.toFixed(2)}`;
  let d = `M ${f(left[0]!)}`;
  for (let i = 1; i < left.length; i++) d += ` L ${f(left[i]!)}`;
  for (let i = right.length - 1; i >= 0; i--) d += ` L ${f(right[i]!)}`;
  return `${d} Z`;
}

function angle(from: Pt, toward: Pt): number {
  return (Math.atan2(toward.y - from.y, toward.x - from.x) * 180) / Math.PI;
}

export const SNAKE_SPRITES = [
  "/cahaya/raka/snakes/ular-1.png",
  "/cahaya/raka/snakes/ular-2.png",
  "/cahaya/raka/snakes/ular-3.png",
  "/cahaya/raka/snakes/ular-4.png",
  "/cahaya/raka/snakes/ular-5.png",
  "/cahaya/raka/snakes/ular-6.png",
  "/cahaya/raka/snakes/ular-7.png",
  "/cahaya/raka/snakes/ular-8.png",
] as const;

/** Palet ular alami (hutan / tanah / sawah) — bukan neon. */
const SNAKE_NATURAL = [
  { belly: "#c4a574", scale: "#2d5a27", mid: "#3d7a35", edge: "#1a3d16", spot: "#1e4d18" },
  { belly: "#d2b48c", scale: "#6b4423", mid: "#8b5a2b", edge: "#3e2723", spot: "#4e342e" },
  { belly: "#e8d5a3", scale: "#556b2f", mid: "#6b8e23", edge: "#334015", spot: "#3d4f1c" },
  { belly: "#c9b896", scale: "#8b4513", mid: "#a0522d", edge: "#5d2e0c", spot: "#6d3b12" },
  { belly: "#b8c9a0", scale: "#2e4a3e", mid: "#3d6b55", edge: "#1b3329", spot: "#244438" },
  { belly: "#e0c090", scale: "#a67c00", mid: "#c9a227", edge: "#6b5200", spot: "#7a5f08" },
  { belly: "#d7ccc8", scale: "#5d4037", mid: "#795548", edge: "#3e2723", spot: "#4e342e" },
  { belly: "#cfd8c0", scale: "#33691e", mid: "#558b2f", edge: "#1b5e20", spot: "#2e7d32" },
];

const LADDER_SPRITE = "/cahaya/raka/ladder-bamboo.png";

export const SHOW_ROUTE_OVERLAY = true;
/** Gambar ular dimatikan — mekanik ular tetap jalan tanpa sprite tubuh/kepala */
export const SHOW_SNAKE_OVERLAY = false;
export const SHOW_LADDER_SPRITES = false;

export function renderBoardRoutes(
  svg: SVGSVGElement,
  snakes: Record<number, number>,
  ladders: Record<number, number>,
): void {
  if (!SHOW_ROUTE_OVERLAY) {
    svg.innerHTML = "";
    return;
  }

  const parts: string[] = [
    `<defs>
      <filter id="softDrop" x="-35%" y="-35%" width="170%" height="170%">
        <feDropShadow dx="0.1" dy="0.22" stdDeviation="0.28" flood-color="#000" flood-opacity="0.4"/>
      </filter>
      <filter id="ladderDrop" x="-30%" y="-30%" width="160%" height="160%">
        <feDropShadow dx="0.15" dy="0.28" stdDeviation="0.35" flood-color="#000" flood-opacity="0.48"/>
      </filter>
      <filter id="snakeShade" x="-40%" y="-40%" width="180%" height="180%">
        <feDropShadow dx="0.12" dy="0.2" stdDeviation="0.32" flood-color="#000" flood-opacity="0.45"/>
      </filter>
    </defs>`,
  ];

  for (const [fromS, toS] of Object.entries(ladders)) {
    const from = boardSvgPoint(Number(fromS));
    const to = boardSvgPoint(Number(toS));
    if (!from || !to) continue;
    const len = dist(from, to);
    const mx = (from.x + to.x) / 2;
    const my = (from.y + to.y) / 2;
    const rot = angle(from, to) + 90;
    // Tipis & pas di antara batu — kurangi tabrakan visual
    const ladderH = Math.max(7.5, len * 0.94);
    const ladderW = Math.min(3.4, Math.max(2.2, len * 0.075));

    if (SHOW_LADDER_SPRITES) {
      parts.push(
        `<image class="ladder-sprite" href="${LADDER_SPRITE}" xlink:href="${LADDER_SPRITE}" x="${(-ladderW / 2).toFixed(2)}" y="${(-ladderH / 2).toFixed(2)}" width="${ladderW.toFixed(2)}" height="${ladderH.toFixed(2)}" preserveAspectRatio="xMidYMid meet" transform="translate(${mx.toFixed(2)} ${my.toFixed(2)}) rotate(${rot.toFixed(2)})" filter="url(#ladderDrop)" opacity="0.92"/>`,
      );
    }
  }

  if (!SHOW_SNAKE_OVERLAY) {
    svg.innerHTML = parts.join("");
    return;
  }

  const entries = Object.entries(snakes)
    .map(([fs, ts]) => [Number(fs), Number(ts)] as const)
    .sort((a, b) => b[0] - a[0]);

  let si = 0;
  for (const [fromPos, toPos] of entries) {
    const from = boardSvgPoint(fromPos);
    const to = boardSvgPoint(toPos);
    if (!from || !to) continue;
    const pal = SNAKE_NATURAL[si % SNAKE_NATURAL.length]!;
    const sprite = SNAKE_SPRITES[si % SNAKE_SPRITES.length]!;
    const spine = naturalSpine(from, to, si);
    const len = dist(from, to);
    const headW = Math.min(CELL * 0.28, Math.max(CELL * 0.16, len * 0.028));
    const tailW = headW * 0.22;
    const body = ribbon(spine, headW, tailW);
    const center = catmull(spine);
    const gid = `snNat${si}`;
    const bellyId = `snBel${si}`;
    const headAng = angle(spine[0]!, spine[Math.min(4, spine.length - 1)]!);
    const headSize = Math.min(CELL * 0.55, 4.8);

    parts.push(`<linearGradient id="${gid}" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="${pal.mid}"/>
      <stop offset="45%" stop-color="${pal.scale}"/>
      <stop offset="100%" stop-color="${pal.edge}"/>
    </linearGradient>`);
    parts.push(`<linearGradient id="${bellyId}" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="${pal.belly}"/>
      <stop offset="100%" stop-color="${pal.mid}" stop-opacity="0.35"/>
    </linearGradient>`);

    // bayangan tubuh
    parts.push(
      `<path d="${body}" fill="#000" opacity="0.22" transform="translate(0.25 0.35)" filter="url(#snakeShade)"/>`,
    );
    // tubuh utama
    parts.push(
      `<path class="snake-body" d="${body}" fill="url(#${gid})" stroke="${pal.edge}" stroke-width="0.22" filter="url(#snakeShade)"/>`,
    );
    // garis perut
    parts.push(
      `<path d="${center}" fill="none" stroke="url(#${bellyId})" stroke-width="${headW * 0.85}" stroke-linecap="round" opacity="0.75"/>`,
    );
    // sisik / spot alami
    const spotN = Math.max(5, Math.min(12, Math.round(spine.length / 2.2)));
    for (let i = 1; i < spotN; i++) {
      const t = i / spotN;
      const idx = Math.min(spine.length - 2, Math.floor(t * (spine.length - 1)));
      const p = spine[idx]!;
      const q = spine[idx + 1]!;
      const ang = angle(p, q);
      const s = lerp(headW * 0.7, headW * 0.25, t);
      parts.push(
        `<ellipse cx="0" cy="0" rx="${(s * 0.85).toFixed(2)}" ry="${(s * 0.38).toFixed(2)}" fill="${pal.spot}" opacity="${0.45 - t * 0.2}" transform="translate(${p.x.toFixed(2)} ${p.y.toFixed(2)}) rotate(${ang.toFixed(1)})"/>`,
      );
    }
    // ekor runcing
    const tip = spine[spine.length - 1]!;
    const pre = spine[spine.length - 2]!;
    const tipAng = angle(pre, tip);
    parts.push(
      `<path d="M 0 0 L ${(tailW * 3.2).toFixed(2)} ${(tailW * 1.1).toFixed(2)} L ${(tailW * 3.2).toFixed(2)} ${(-tailW * 1.1).toFixed(2)} Z" fill="${pal.edge}" opacity="0.95" transform="translate(${tip.x.toFixed(2)} ${tip.y.toFixed(2)}) rotate(${tipAng.toFixed(1)})"/>`,
    );
    // kepala sprite lebih besar & jelas
    const rot = headAng + 90;
    parts.push(
      `<g class="snake-head" filter="url(#snakeShade)" transform="translate(${from.x.toFixed(2)} ${from.y.toFixed(2)}) rotate(${rot.toFixed(1)})">
        <circle r="${(headSize * 0.38).toFixed(2)}" fill="${pal.mid}" opacity="0.35"/>
        <image href="${sprite}" xlink:href="${sprite}" x="${(-headSize / 2).toFixed(2)}" y="${(-headSize / 2).toFixed(2)}" width="${headSize.toFixed(2)}" height="${headSize.toFixed(2)}" preserveAspectRatio="xMidYMid meet"/>
      </g>`,
    );
    si += 1;
  }

  svg.innerHTML = parts.join("");
}

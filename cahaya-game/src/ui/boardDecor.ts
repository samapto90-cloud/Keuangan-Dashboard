import { boardSvgPoint } from "../game/board/engine";

type Pt = { x: number; y: number };

const CELL = 9;

function dist(a: Pt, b: Pt): number {
  return Math.hypot(b.x - a.x, b.y - a.y) || 1;
}

function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t;
}

/** Kurva lembut satu gelombang — modern, tidak berlebih. */
function softArc(a: Pt, b: Pt, twist: number): Pt[] {
  const dx = b.x - a.x;
  const dy = b.y - a.y;
  const len = Math.hypot(dx, dy) || 1;
  const nx = -dy / len;
  const ny = dx / len;
  const sign = twist % 2 === 0 ? 1 : -1;
  const amp = Math.min(CELL * 0.38, Math.max(CELL * 0.14, len * 0.1)) * sign;
  const n = Math.max(10, Math.round(len / 2.6));
  const pts: Pt[] = [];
  for (let i = 0; i <= n; i++) {
    const t = i / n;
    const ease = t * t * (3 - 2 * t);
    const wiggle = Math.sin(Math.PI * t) * amp;
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
    const w = lerp(headW, tailW, t * t);
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

/** Palet neon modern — tiap ular beda aksen, tetap satu bahasa visual. */
const SNAKE_NEON = [
  { core: "#ff6b6b", mid: "#e53935", glow: "#ff8a80", dark: "#7f1d1d" },
  { core: "#ce93d8", mid: "#8e24aa", glow: "#e1bee7", dark: "#4a148c" },
  { core: "#4fc3f7", mid: "#0288d1", glow: "#81d4fa", dark: "#01579b" },
  { core: "#ffd54f", mid: "#f9a825", glow: "#ffe082", dark: "#e65100" },
  { core: "#26a69a", mid: "#00897b", glow: "#80cbc4", dark: "#004d40" },
  { core: "#ffb74d", mid: "#ef6c00", glow: "#ffcc80", dark: "#bf360c" },
  { core: "#81c784", mid: "#43a047", glow: "#a5d6a7", dark: "#1b5e20" },
  { core: "#b39ddb", mid: "#5e35b1", glow: "#d1c4e9", dark: "#311b92" },
];

const LADDER_NEON = [
  { rail: "#ffe082", railDark: "#f9a825", rung: "#fff8e1", glow: "#ffca28" },
  { rail: "#80deea", railDark: "#00acc1", rung: "#e0f7fa", glow: "#26c6da" },
  { rail: "#ffcc80", railDark: "#fb8c00", rung: "#fff3e0", glow: "#ffa726" },
];

export function renderBoardRoutes(
  svg: SVGSVGElement,
  snakes: Record<number, number>,
  ladders: Record<number, number>,
): void {
  const parts: string[] = [
    `<defs>
      <filter id="neonSoft" x="-40%" y="-40%" width="180%" height="180%">
        <feGaussianBlur stdDeviation="0.45" result="b"/>
        <feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge>
      </filter>
      <filter id="softDrop" x="-30%" y="-30%" width="160%" height="160%">
        <feDropShadow dx="0.08" dy="0.18" stdDeviation="0.22" flood-color="#000" flood-opacity="0.35"/>
      </filter>
    </defs>`,
  ];

  let li = 0;
  for (const [fromS, toS] of Object.entries(ladders)) {
    const from = boardSvgPoint(Number(fromS));
    const to = boardSvgPoint(Number(toS));
    if (!from || !to) continue;
    const style = LADDER_NEON[li % LADDER_NEON.length]!;
    const dx = to.x - from.x;
    const dy = to.y - from.y;
    const len = dist(from, to);
    const ux = dx / len;
    const uy = dy / len;
    const half = Math.min(CELL * 0.26, Math.max(CELL * 0.16, len * 0.022));
    const px = -uy * half;
    const py = ux * half;
    const railW = Math.min(0.62, half * 0.42);
    const rungs = Math.max(4, Math.min(8, Math.round(len / 4.2)));
    const gid = `ladG${li}`;

    parts.push(`<linearGradient id="${gid}" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="${style.rail}"/>
      <stop offset="100%" stop-color="${style.railDark}"/>
    </linearGradient>`);

    // glow backdrop
    for (const side of [1, -1] as const) {
      const ox = px * side;
      const oy = py * side;
      parts.push(
        `<line x1="${from.x + ox}" y1="${from.y + oy}" x2="${to.x + ox}" y2="${to.y + oy}" stroke="${style.glow}" stroke-width="${railW + 0.55}" stroke-linecap="round" opacity="0.28" filter="url(#neonSoft)"/>`,
      );
      parts.push(
        `<line class="ladder-rail" x1="${from.x + ox}" y1="${from.y + oy}" x2="${to.x + ox}" y2="${to.y + oy}" stroke="url(#${gid})" stroke-width="${railW}" stroke-linecap="round" filter="url(#softDrop)"/>`,
      );
      parts.push(
        `<line x1="${from.x + ox}" y1="${from.y + oy}" x2="${to.x + ox}" y2="${to.y + oy}" stroke="#fff" stroke-width="${railW * 0.28}" stroke-linecap="round" opacity="0.55"/>`,
      );
    }

    for (let i = 1; i <= rungs; i++) {
      const t = i / (rungs + 1);
      const x1 = from.x + px + dx * t;
      const y1 = from.y + py + dy * t;
      const x2 = from.x - px + dx * t;
      const y2 = from.y - py + dy * t;
      parts.push(
        `<line class="ladder-rung" x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" stroke="${style.rung}" stroke-width="${railW * 0.7}" stroke-linecap="round" opacity="0.95"/>`,
      );
      parts.push(
        `<line x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" stroke="${style.glow}" stroke-width="${railW * 0.28}" stroke-linecap="round" opacity="0.45"/>`,
      );
    }

    const cap = Math.min(0.42, half * 0.28);
    for (const side of [1, -1] as const) {
      parts.push(`<circle cx="${from.x + px * side}" cy="${from.y + py * side}" r="${cap}" fill="${style.rail}" stroke="${style.railDark}" stroke-width="0.1"/>`);
      parts.push(`<circle cx="${to.x + px * side}" cy="${to.y + py * side}" r="${cap * 1.05}" fill="${style.glow}" stroke="${style.rail}" stroke-width="0.1"/>`);
    }
    li += 1;
  }

  const entries = Object.entries(snakes)
    .map(([fs, ts]) => [Number(fs), Number(ts)] as const)
    .sort((a, b) => b[0] - a[0]);

  let si = 0;
  for (const [fromPos, toPos] of entries) {
    const from = boardSvgPoint(fromPos);
    const to = boardSvgPoint(toPos);
    if (!from || !to) continue;
    const pal = SNAKE_NEON[si % SNAKE_NEON.length]!;
    const sprite = SNAKE_SPRITES[si % SNAKE_SPRITES.length]!;
    const spine = softArc(from, to, si);
    const len = dist(from, to);
    const headW = Math.min(CELL * 0.16, Math.max(CELL * 0.1, len * 0.014));
    const tailW = headW * 0.3;
    const body = ribbon(spine, headW, tailW);
    const center = catmull(spine);
    const gid = `snG${si}`;
    const headAng = angle(spine[0]!, spine[Math.min(3, spine.length - 1)]!);
    const headSize = Math.min(CELL * 0.42, 3.6);

    parts.push(`<linearGradient id="${gid}" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="${pal.glow}"/>
      <stop offset="45%" stop-color="${pal.core}"/>
      <stop offset="100%" stop-color="${pal.dark}"/>
    </linearGradient>`);

    // soft neon trail under body
    parts.push(
      `<path d="${center}" fill="none" stroke="${pal.glow}" stroke-width="${headW * 2.4}" stroke-linecap="round" opacity="0.22" filter="url(#neonSoft)"/>`,
    );
    parts.push(
      `<path class="snake-body" d="${body}" fill="url(#${gid})" stroke="${pal.mid}" stroke-width="0.14" filter="url(#softDrop)"/>`,
    );
    parts.push(
      `<path d="${center}" fill="none" stroke="#fff" stroke-width="${headW * 0.35}" stroke-linecap="round" opacity="0.35"/>`,
    );

    // subtle segment dashes along spine
    const dashN = Math.max(3, Math.min(7, Math.round(spine.length / 3)));
    for (let i = 1; i < dashN; i++) {
      const t = i / dashN;
      const idx = Math.min(spine.length - 2, Math.floor(t * (spine.length - 1)));
      const p = spine[idx]!;
      const q = spine[idx + 1]!;
      const ang = angle(p, q);
      const s = lerp(headW * 0.55, headW * 0.2, t);
      parts.push(
        `<ellipse cx="0" cy="0" rx="${s * 0.9}" ry="${s * 0.35}" fill="${pal.glow}" opacity="${0.35 - t * 0.15}" transform="translate(${p.x} ${p.y}) rotate(${ang})"/>`,
      );
    }

    const tip = spine[spine.length - 1]!;
    const pre = spine[spine.length - 2]!;
    const tipAng = angle(pre, tip);
    parts.push(
      `<path d="M 0 0 L ${tailW * 2.6} ${tailW} L ${tailW * 2.6} ${-tailW} Z" fill="${pal.dark}" opacity="0.9" transform="translate(${tip.x} ${tip.y}) rotate(${tipAng})"/>`,
    );

    const rot = headAng + 90;
    parts.push(
      `<g class="snake-head" filter="url(#softDrop)" transform="translate(${from.x} ${from.y}) rotate(${rot})">
        <circle r="${headSize * 0.32}" fill="${pal.glow}" opacity="0.25" filter="url(#neonSoft)"/>
        <image href="${sprite}" xlink:href="${sprite}" x="${-headSize / 2}" y="${-headSize / 2}" width="${headSize}" height="${headSize}" preserveAspectRatio="xMidYMid meet"/>
      </g>`,
    );
    si += 1;
  }

  svg.innerHTML = parts.join("");
}

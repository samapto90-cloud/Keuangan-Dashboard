export const ASSETS = {
  board: { overlay: "css:board-overlay" },
  token: { pawn: "css:pawn" },
  snake: { path: "svg:snake" },
  ladder: { path: "svg:ladder" },
  avatar: { fallback: "css:avatar" },
  icons: { set: "svg:nt-icon" },
  audio: { sfx: "webaudio:sfx", music: "webaudio:music" },
} as const;

export const TOKEN_MARKS = ["circle", "diamond", "square", "triangle", "star", "hex", "heart", "plus"] as const;

/** Tokoh bidak — media dari ular-tangga-3d, role seat 0–7 */
export const PLAYER_CHARACTERS = [
  {
    id: "prabowo",
    name: "Prabowo",
    src: "/cahaya/raka/players/prabowo.png",
    head: "/cahaya/raka/players/heads/prabowo.png",
    color: "#2563eb",
  },
  {
    id: "gibran",
    name: "Gibran",
    src: "/cahaya/raka/players/gibran.png",
    head: "/cahaya/raka/players/heads/gibran.png",
    color: "#0ea5e9",
  },
  {
    id: "jokowi",
    name: "Jokowi",
    src: "/cahaya/raka/players/jokowi.png",
    head: "/cahaya/raka/players/heads/jokowi.png",
    color: "#ef4444",
  },
  {
    id: "anies",
    name: "Anies",
    src: "/cahaya/raka/players/anies.png",
    head: "/cahaya/raka/players/heads/anies.png",
    color: "#3b82f6",
  },
  {
    id: "ganjar",
    name: "Ganjar",
    src: "/cahaya/raka/players/ganjar.png",
    head: "/cahaya/raka/players/heads/ganjar.png",
    color: "#dc2626",
  },
  {
    id: "sri-mulyani",
    name: "Sri Mulyani",
    src: "/cahaya/raka/players/sri-mulyani.png",
    head: "/cahaya/raka/players/heads/sri-mulyani.png",
    color: "#059669",
  },
  {
    id: "khofifah",
    name: "Khofifah",
    src: "/cahaya/raka/players/khofifah.png",
    head: "/cahaya/raka/players/heads/khofifah.png",
    color: "#8b5cf6",
  },
  {
    id: "bahlil",
    name: "Bahlil",
    src: "/cahaya/raka/players/bahlil.png",
    head: "/cahaya/raka/players/heads/bahlil.png",
    color: "#f59e0b",
  },
] as const;

function escAttr(v: string): string {
  return v.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c] || c);
}

export function playerCharacter(index: number): (typeof PLAYER_CHARACTERS)[number] {
  return PLAYER_CHARACTERS[Math.abs(index) % PLAYER_CHARACTERS.length];
}

export function playerSpriteSrc(index: number): string {
  return playerCharacter(index).src;
}

export function tokenMark(index: number): string {
  return TOKEN_MARKS[Math.abs(index) % TOKEN_MARKS.length];
}

export function truncateName(name: string, max = 14): string {
  const t = name.trim();
  if (t.length <= max) return t;
  return `${t.slice(0, max - 1)}…`;
}

/** Token papan: tubuh utuh (tanpa lingkaran di kaki) */
export function pawnSpriteHtml(index: number, alt: string): string {
  const ch = playerCharacter(index);
  return `<img class="pawn-sprite pawn-full" src="${ch.src}" alt="${escAttr(alt)}" draggable="false" />`;
}

/** Kartu/lobby: tubuh penuh */
export function avatarSpriteHtml(index: number, alt: string): string {
  const ch = playerCharacter(index);
  return `<img class="pcard-sprite" src="${ch.src}" alt="${escAttr(alt)}" draggable="false" />`;
}

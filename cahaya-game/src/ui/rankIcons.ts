/** Original rank badges. Geometric Nusantara-inspired marks, not copied from other games. */

const TIERS = ["BRONZE", "SILVER", "GOLD", "PLATINUM", "DIAMOND", "MASTER", "GRANDMASTER"] as const;

export function rankIcon(tier?: string, size = 36): string {
  const t = String(tier || "BRONZE").toUpperCase();
  const fill = {
    BRONZE: "#b56a3a",
    SILVER: "#c5d0d8",
    GOLD: "#d4b45a",
    PLATINUM: "#8fd0c4",
    DIAMOND: "#6ec8e6",
    MASTER: "#c48be8",
    GRANDMASTER: "#f4efe3",
  }[t] || "#b56a3a";
  const mark =
    t === "SILVER"
      ? `<polygon points="12,3 21,12 12,21 3,12"/>`
      : t === "GOLD"
        ? `<path d="M12 3l2.2 6.2H21l-5.3 3.8 2.1 6.5L12 16.2 6.2 19.5 8.3 13 3 9.2h6.8z"/>`
        : t === "PLATINUM"
          ? `<polygon points="12,3 20,8 20,16 12,21 4,16 4,8"/>`
          : t === "DIAMOND"
            ? `<polygon points="12,2 22,10 12,22 2,10"/>`
            : t === "MASTER"
              ? `<path d="M4 16V8l4 3 4-6 4 6 4-3v8H4z"/>`
              : t === "GRANDMASTER"
                ? `<path d="M12 2l2 5 5 .4-3.8 3.2 1.2 5L12 13.2 7.6 15.6 8.8 10.6 5 7.4 10 7z"/><circle cx="12" cy="18" r="2"/>`
                : `<circle cx="12" cy="12" r="7"/>`;
  return `<svg class="rank-icon" width="${size}" height="${size}" viewBox="0 0 24 24" fill="${fill}" stroke="#0a1c14" stroke-width="1.2" aria-hidden="true">${mark}</svg>`;
}

export function nextRankLabel(tier?: string, division?: string): string {
  const t = String(tier || "BRONZE").toUpperCase();
  const d = String(division || "").toUpperCase();
  if (t === "GRANDMASTER") return "Puncak musim";
  if (t === "MASTER") return "GRANDMASTER";
  const di = ["III", "II", "I"].indexOf(d);
  if (di >= 0 && di < 2) return `${t} ${["II", "I"][di]}`;
  const ti = TIERS.indexOf(t as (typeof TIERS)[number]);
  const nxt = TIERS[Math.min(TIERS.length - 1, Math.max(0, ti) + 1)];
  if (nxt === "MASTER" || nxt === "GRANDMASTER") return nxt;
  return `${nxt} III`;
}

export function rrBarPct(rr?: number, tier?: string): number {
  const n = Math.max(0, Number(rr || 0));
  const cap = String(tier || "").toUpperCase() === "MASTER" || String(tier || "").toUpperCase() === "GRANDMASTER" ? 200 : 100;
  return Math.max(4, Math.min(100, Math.round((n % cap) / cap * 100)));
}

export function rrToNext(rr?: number, tier?: string): number {
  const n = Math.max(0, Number(rr || 0));
  const cap = String(tier || "").toUpperCase() === "MASTER" || String(tier || "").toUpperCase() === "GRANDMASTER" ? 200 : 100;
  return Math.max(0, cap - (n % cap));
}

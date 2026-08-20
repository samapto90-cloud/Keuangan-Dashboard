/** Zona lanskap papan — desa, sawah, hutan, danau, laut, bukit, gunung. */
export type BoardZone =
  | "desa"
  | "sawah"
  | "hutan"
  | "danau"
  | "laut"
  | "bukit"
  | "gunung";

const ZONE_BY_BAND: BoardZone[] = [
  "desa",
  "sawah",
  "hutan",
  "danau",
  "laut",
  "bukit",
  "gunung",
];

export function zoneForPosition(pos: number): BoardZone {
  const p = Math.max(1, Math.min(100, pos));
  if (p <= 14) return "desa";
  if (p <= 28) return "sawah";
  if (p <= 42) return "hutan";
  if (p <= 56) return "danau";
  if (p <= 70) return "laut";
  if (p <= 85) return "bukit";
  return "gunung";
}

export function zoneClass(pos: number): string {
  return `zone-${zoneForPosition(pos)}`;
}

export function zoneDecor(_pos: number): string {
  return "";
}

export { ZONE_BY_BAND };

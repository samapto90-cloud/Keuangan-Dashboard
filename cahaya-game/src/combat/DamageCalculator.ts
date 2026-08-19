/** Mirror rumus server — jangan dipakai untuk mengubah HP. */
export function calcDisplayDamage(
  base: number,
  strength: number,
  skillPower: number,
  equipment: number,
  defense: number,
  crit: boolean,
  critMul = 1.5,
): number {
  let raw = base * (1 + strength / 10) * (skillPower || 1) * (equipment || 1);
  let mit = 1 - defense / (defense + 50);
  if (mit < 0.2) mit = 0.2;
  raw *= mit;
  if (crit) raw *= critMul;
  return Math.max(1, Math.round(raw));
}

export type CombatState =
  | "IDLE"
  | "ATTACKING"
  | "COMBO"
  | "DODGING"
  | "CASTING"
  | "HIT"
  | "STUNNED"
  | "DEAD"
  | "DOWNED"
  | "RESPAWNING"
  | "ENERGY_ATTACK"
  | "SPECIAL_ATTACK"
  | "ULTIMATE"
  | "GUARD"
  | "COUNTER"
  | "CHARGE"
  | "TRANSFORM";

export type AttackKind = "punch" | "kick" | "energy" | "dodge" | "special" | "ultimate";

export type AttackPhase = "idle" | "startup" | "active" | "recovery";

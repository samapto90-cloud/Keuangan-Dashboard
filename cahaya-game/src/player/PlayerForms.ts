export type PlayerFormId = "asal" | "cahaya" | "kilat" | "agung" | "fajar";

export interface PlayerFormDef {
  id: PlayerFormId;
  name: string;
  shortName: string;
  hair: "black-short" | "gold-short" | "gold-long" | "black-long";
  eyeColor: number;
  eyeEmissive: number;
  brows: boolean;
  vest: "filigree" | "gold-crack" | "none";
  shirt: "white" | "sleeveless" | "bare";
  pants: "joggers" | "torn";
  shoes: "sneakers" | "boots";
  fur: boolean;
  tail: boolean;
  aura: "none" | "gold" | "gold-strong" | "orange";
  lightning: "none" | "gold" | "cyan";
}

export const PLAYER_FORM_ORDER: PlayerFormId[] = ["asal", "cahaya", "kilat", "agung", "fajar"];

export const PLAYER_FORMS: Record<PlayerFormId, PlayerFormDef> = {
  asal: {
    id: "asal",
    name: "Wujud Asal",
    shortName: "Asal",
    hair: "black-short",
    eyeColor: 0x1c1410,
    eyeEmissive: 0x000000,
    brows: true,
    vest: "filigree",
    shirt: "white",
    pants: "joggers",
    shoes: "sneakers",
    fur: false,
    tail: false,
    aura: "none",
    lightning: "none",
  },
  cahaya: {
    id: "cahaya",
    name: "Wujud Cahaya",
    shortName: "Cahaya",
    hair: "gold-short",
    eyeColor: 0x2ee6d6,
    eyeEmissive: 0x1aa6a0,
    brows: true,
    vest: "gold-crack",
    shirt: "white",
    pants: "joggers",
    shoes: "sneakers",
    fur: false,
    tail: false,
    aura: "gold",
    lightning: "none",
  },
  kilat: {
    id: "kilat",
    name: "Wujud Kilat",
    shortName: "Kilat",
    hair: "gold-short",
    eyeColor: 0x7cffef,
    eyeEmissive: 0x2ee6d6,
    brows: true,
    vest: "gold-crack",
    shirt: "white",
    pants: "joggers",
    shoes: "sneakers",
    fur: false,
    tail: false,
    aura: "gold-strong",
    lightning: "cyan",
  },
  agung: {
    id: "agung",
    name: "Wujud Agung",
    shortName: "Agung",
    hair: "gold-long",
    eyeColor: 0x9dff6a,
    eyeEmissive: 0x7cff4a,
    brows: false,
    vest: "gold-crack",
    shirt: "sleeveless",
    pants: "torn",
    shoes: "boots",
    fur: false,
    tail: false,
    aura: "gold-strong",
    lightning: "gold",
  },
  fajar: {
    id: "fajar",
    name: "Wujud Fajar",
    shortName: "Fajar",
    hair: "black-long",
    eyeColor: 0xff6a2a,
    eyeEmissive: 0xff3b00,
    brows: true,
    vest: "none",
    shirt: "bare",
    pants: "torn",
    shoes: "boots",
    fur: true,
    tail: true,
    aura: "orange",
    lightning: "gold",
  },
};

export function formByIndex(index: number): PlayerFormId {
  return PLAYER_FORM_ORDER[(index + PLAYER_FORM_ORDER.length) % PLAYER_FORM_ORDER.length];
}

export function nextFormId(current: PlayerFormId): PlayerFormId {
  const i = PLAYER_FORM_ORDER.indexOf(current);
  return formByIndex(i + 1);
}

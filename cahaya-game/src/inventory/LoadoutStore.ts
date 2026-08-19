import catalogJson from "../data/items.json";
import type {
  EquipmentView,
  InspectOut,
  InventoryUpdated,
  InvSlotView,
  ItemDefView,
  ItemEffects,
  PartyView,
  SocialState,
  StatsView,
} from "../network/NetworkMessage";

export const INV_CAPACITY = 50;
export const EQUIP_SLOTS = ["HEAD", "BODY", "LEGS", "WEAPON", "ACCESSORY_1", "ACCESSORY_2", "ACCESSORY_3"] as const;
export type EquipSlotId = (typeof EQUIP_SLOTS)[number];

export const RARITY_COLOR: Record<string, string> = {
  COMMON: "#9ca3af",
  UNCOMMON: "#4ade80",
  RARE: "#60a5fa",
  EPIC: "#c084fc",
  LEGENDARY: "#fbbf24",
  MYTHIC: "#f472b6",
};

const fallbackCatalog = catalogJson as ItemDefView[];

export class LoadoutStore {
  catalog = new Map<string, ItemDefView>();
  slots: InvSlotView[] = emptySlots();
  equipment: EquipmentView = {};
  stats: StatsView | null = null;
  coin = 0;
  crystal = 0;
  eduToken = 0;
  battleToken = 0;
  guardianToken = 0;
  raidToken = 0;
  version = 0;
  party: PartyView | null = null;
  social: SocialState = emptySocial();
  inspect: InspectOut | null = null;
  notes: string[] = [];
  tempLoot: InventoryUpdated["tempLoot"] = [];
  setPieces = 0;
  bagCapacity = INV_CAPACITY;
  itemHistory: InventoryUpdated["itemHistory"] = [];

  constructor() {
    this.setCatalog(fallbackCatalog);
  }

  setCatalog(list: ItemDefView[] | undefined): void {
    if (!list?.length) return;
    this.catalog.clear();
    for (const item of list) this.catalog.set(item.id, item);
  }

  applyLoadout(upd: InventoryUpdated): void {
    this.version = upd.inventoryVersion;
    this.equipment = upd.equipment ?? {};
    this.stats = upd.stats;
    this.coin = upd.coin;
    this.crystal = upd.crystal;
    this.eduToken = upd.eduToken;
    this.battleToken = upd.battleToken ?? this.battleToken;
    this.guardianToken = upd.guardianToken ?? this.guardianToken;
    this.raidToken = upd.raidToken ?? this.raidToken;
    this.tempLoot = upd.tempLoot ?? [];
    this.setPieces = upd.setPieces ?? 0;
    this.bagCapacity = upd.bagCapacity || this.slots.length || INV_CAPACITY;
    this.itemHistory = upd.itemHistory ?? this.itemHistory;
    if (upd.slots?.length) {
      const n = Math.max(INV_CAPACITY, this.bagCapacity, ...upd.slots.map((s) => s.index + 1));
      this.slots = emptySlots(n);
      for (const slot of upd.slots) this.slots[slot.index] = slot;
    } else if (upd.changedSlots?.length) {
      for (const slot of upd.changedSlots) {
        if (slot.index >= this.slots.length) {
          const grown = emptySlots(slot.index + 1);
          for (let i = 0; i < this.slots.length; i++) grown[i] = this.slots[i];
          this.slots = grown;
        }
        this.slots[slot.index] = slot;
      }
    }
    if (upd.toast) this.pushNote(upd.toast);
  }

  applyStats(stats: StatsView): void {
    this.stats = stats;
  }

  applyParty(party: PartyView | null | undefined): void {
    this.party = party?.partyId ? party : null;
    this.social.party = this.party;
  }

  applySocial(state: SocialState): void {
    this.social = {
      ...emptySocial(),
      ...state,
      friends: state.friends ?? [],
      pending: state.pending ?? [],
      outgoing: state.outgoing ?? [],
      blocked: state.blocked ?? [],
      nearby: state.nearby ?? [],
      notifies: state.notifies ?? [],
      guild: state.guild,
      wallet: state.wallet,
    };
    if (state.wallet) {
      this.coin = state.wallet.coins ?? this.coin;
      this.battleToken = state.wallet.battleTokens ?? this.battleToken;
      this.guardianToken = state.wallet.guardianTokens ?? this.guardianToken;
      this.raidToken = state.wallet.raidTokens ?? this.raidToken;
      this.eduToken = state.wallet.educationTokens ?? this.eduToken;
    }
    if (state.party !== undefined) this.applyParty(state.party);
  }

  item(id: string | undefined): ItemDefView | null {
    if (!id) return null;
    return this.catalog.get(id) ?? null;
  }

  slot(index: number): InvSlotView {
    return this.slots[index] ?? { index, qty: 0 };
  }

  equipped(slot: EquipSlotId): ItemDefView | null {
    return this.item(this.equipment[slot]);
  }

  preview(item: ItemDefView): { label: string; current: number; next: number; delta: number }[] {
    if ((item.type !== "EQUIPMENT" && item.type !== "WEAPON" && item.type !== "ARMOR" && item.type !== "HELM" && item.type !== "ACCESSORY") || !item.slot) return [];
    const slot = item.slot as EquipSlotId;
    const worn = this.equipped(slot);
    const rows: { label: string; current: number; next: number; delta: number }[] = [];
    for (const key of STAT_KEYS) {
      const delta = num(item.effects[key]) - num(worn?.effects[key]);
      if (!delta && !num(item.effects[key]) && !num(worn?.effects[key])) continue;
      const current = baseStat(this.stats, key);
      rows.push({ label: STAT_LABEL[key], current, next: current + delta, delta });
    }
    return rows;
  }

  pushNote(text: string): void {
    this.notes.push(text);
    if (this.notes.length > 12) this.notes.shift();
  }

  takeNotes(): string[] {
    const out = this.notes.splice(0, this.notes.length);
    return out;
  }
}

const STAT_KEYS: (keyof ItemEffects)[] = [
  "attack",
  "defense",
  "maxHp",
  "maxEnergy",
  "strength",
  "agility",
  "energyPower",
  "criticalChance",
  "movementSpeed",
  "dodge",
  "energyRegen",
];

const STAT_LABEL: Record<string, string> = {
  attack: "ATTACK",
  defense: "DEFENSE",
  maxHp: "MAX HP",
  maxEnergy: "MAX ENERGY",
  strength: "STRENGTH",
  agility: "AGILITY",
  energyPower: "ENERGY POWER",
  criticalChance: "CRITICAL",
  movementSpeed: "MOVE SPEED",
  dodge: "DODGE",
  energyRegen: "ENERGY REGEN",
};

function baseStat(stats: StatsView | null, key: keyof ItemEffects): number {
  if (!stats) return 0;
  if (key === "attack") return stats.attack;
  if (key === "defense") return stats.defense;
  if (key === "maxHp") return stats.maxHp;
  if (key === "maxEnergy") return stats.maxEnergy;
  if (key === "strength") return stats.strength;
  if (key === "agility") return stats.agility;
  if (key === "energyPower") return stats.energyPower;
  if (key === "criticalChance") return stats.criticalChance;
  if (key === "movementSpeed") return stats.moveSpeed;
  return 0;
}

function num(n: number | undefined): number {
  return typeof n === "number" ? n : 0;
}

function emptySlots(n = INV_CAPACITY): InvSlotView[] {
  return Array.from({ length: n }, (_, index) => ({ index, qty: 0 }));
}

function emptySocial(): SocialState {
  return { friends: [], pending: [], outgoing: [], blocked: [], nearby: [] };
}

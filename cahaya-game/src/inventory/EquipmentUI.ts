import { escapeHtml } from "../dialogue/DialogueUI";
import { EQUIP_SLOTS, type EquipSlotId, type LoadoutStore, RARITY_COLOR } from "./LoadoutStore";
import { itemGlyph } from "./ItemTooltip";

const SLOT_LABEL: Record<EquipSlotId, string> = {
  HEAD: "HELM",
  BODY: "ARMOR",
  LEGS: "LEGS",
  WEAPON: "WEAPON",
  ACCESSORY_1: "ACCESSORY 1",
  ACCESSORY_2: "ACCESSORY 2",
  ACCESSORY_3: "ACCESSORY 3",
};

export class EquipmentUI {
  static html(store: LoadoutStore, selectedSlot: EquipSlotId | null): string {
    return `<div class="eq-slots">${EQUIP_SLOTS.map((slot) => {
      const item = store.equipped(slot);
      const color = item ? RARITY_COLOR[item.rarity] ?? "#9ca3af" : "rgba(255,255,255,0.12)";
      const on = selectedSlot === slot ? " selected" : "";
      return `<button type="button" class="eq-slot${on}" data-equip="${slot}" style="--rarity:${color}">
        <span class="eq-kicker">${SLOT_LABEL[slot]}</span>
        <span class="eq-glyph">${item ? itemGlyph(item.icon) : "—"}</span>
        <span class="eq-name">${item ? escapeHtml(item.name) : "empty"}</span>
      </button>`;
    }).join("")}</div>`;
  }
}

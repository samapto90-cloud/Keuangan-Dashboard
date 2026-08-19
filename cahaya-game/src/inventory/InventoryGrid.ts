import type { InvSlotView } from "../network/NetworkMessage";
import { RARITY_COLOR } from "./LoadoutStore";
import { itemGlyph } from "./ItemTooltip";

export class InventoryGrid {
  static html(slots: InvSlotView[], selected: number): string {
    const n = Math.max(slots.length, 50);
    const cells = Array.from({ length: n }, (_, i) => {
      const slot = slots[i] ?? { index: i, qty: 0 };
      const item = slot.item;
      const empty = !item || slot.qty < 1;
      const color = item ? RARITY_COLOR[item.rarity] ?? "#9ca3af" : "transparent";
      const qty = item?.stackable && slot.qty > 1 ? `<span class="inv-qty">${slot.qty}</span>` : "";
      const sel = selected === slot.index ? " selected" : "";
      const lock = slot.locked ? " locked" : "";
      const fav = slot.favorite ? " fav" : "";
      const mark = slot.locked ? "L" : slot.favorite ? "★" : rarityMark(item?.rarity);
      const up = slot.upgrade ? `<span class="inv-up">+${slot.upgrade}</span>` : "";
      return `<button type="button" class="inv-slot${empty ? " empty" : ""}${sel}${lock}${fav}" data-slot="${slot.index}" draggable="${empty ? "false" : "true"}" style="--rarity:${color}">
        ${empty ? `<span class="inv-empty">empty</span>` : `<span class="inv-glyph">${itemGlyph(item?.icon)}</span>${qty}${up}<span class="inv-rarity" title="${item?.rarity ?? ""}">${mark}</span>`}
      </button>`;
    });
    return `<div class="inv-grid">${cells.join("")}</div>`;
  }
}

function rarityMark(rarity?: string): string {
  switch (rarity) {
    case "MYTHIC":
      return "M";
    case "LEGENDARY":
      return "L";
    case "EPIC":
      return "E";
    case "RARE":
      return "R";
    case "UNCOMMON":
      return "U";
    default:
      return "";
  }
}

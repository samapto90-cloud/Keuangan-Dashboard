import { escapeHtml } from "../dialogue/DialogueUI";
import type { ItemDefView } from "../network/NetworkMessage";
import { RARITY_COLOR } from "./LoadoutStore";

export class ItemTooltip {
  static html(item: ItemDefView | null | undefined, qty = 0, extra?: { upgrade?: number; instanceId?: string; setPieces?: number }): string {
    if (!item) return `<div class="item-tip empty">Pilih item untuk melihat detail.</div>`;
    const color = RARITY_COLOR[item.rarity] ?? "#9ca3af";
    const stats = effectLines(item);
    const stack = item.stackable ? `<div>Stack: ${qty} / ${item.maxStack}</div>` : "";
    const lore = item.lore ? `<p class="item-lore">${escapeHtml(item.lore)}</p>` : "";
    const set = item.setId ? `<div class="item-set">DAWN SET ${extra?.setPieces ?? 0}/4</div>` : "";
    const up = extra?.upgrade ? `<div>Upgrade +${extra.upgrade}</div>` : "";
    return `
      <div class="item-tip" style="border-color:${color}">
        <div class="item-tip-name" style="color:${color}">${escapeHtml(item.name)}</div>
        <div class="item-tip-meta">${escapeHtml(item.rarity)} · ${escapeHtml(item.type)}${item.slot ? ` · ${escapeHtml(item.slot)}` : ""} · Lv ${item.itemLevel ?? 1}</div>
        <p>${escapeHtml(item.description)}</p>
        ${stats}
        ${up}
        <div>Required Level ${item.levelRequirement}</div>
        <div>${escapeHtml(item.bind ?? (item.tradable === false ? "SOULBOUND" : "TRADABLE"))}</div>
        <div>Value ${item.value}</div>
        ${stack}
        ${set}
        ${lore}
      </div>`;
  }
}

function effectLines(item: ItemDefView): string {
  const e = item.effects ?? {};
  const rows: string[] = [];
  if (e.healPct) rows.push(`Restore ${Math.round(e.healPct * 100)}% HP`);
  if (e.energyPct) rows.push(`Restore ${Math.round(e.energyPct * 100)}% Energy`);
  if (e.staminaPct) rows.push(`Restore ${Math.round(e.staminaPct * 100)}% Stamina`);
  if (e.attack) rows.push(`Attack +${e.attack}`);
  if (e.defense) rows.push(`Defense +${e.defense}`);
  if (e.maxHp) rows.push(`Max HP +${e.maxHp}`);
  if (e.maxEnergy) rows.push(`Max Energy +${e.maxEnergy}`);
  if (e.strength) rows.push(`Strength +${e.strength}`);
  if (e.agility) rows.push(`Agility +${e.agility}`);
  if (e.energyPower) rows.push(`Energy Power +${e.energyPower}`);
  if (e.criticalChance) rows.push(`Critical +${Math.round(e.criticalChance * 100)}%`);
  if (e.movementSpeed) rows.push(`Move Speed +${Math.round(e.movementSpeed * 100)}%`);
  if (e.range) rows.push(`Range +${e.range}`);
  if (e.dodge) rows.push(`Dodge +${e.dodge}`);
  if (e.energyRegen) rows.push(`Energy Regen +${e.energyRegen}`);
  if (!rows.length) return "";
  return `<ul class="item-stats">${rows.map((r) => `<li>${r}</li>`).join("")}</ul>`;
}

export function itemGlyph(icon: string | undefined): string {
  const map: Record<string, string> = {
    "potion-hp": "♥",
    "potion-en": "✦",
    "potion-st": "⚡",
    crystal: "◇",
    helm: "⌂",
    armor: "▣",
    boots: "▽",
    staff: "†",
    sword: "⚔",
    spear: "↑",
    gloves: "☰",
    pendant: "◆",
    token: "★",
  };
  return map[icon ?? ""] ?? "•";
}

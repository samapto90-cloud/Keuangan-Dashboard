import { escapeHtml } from "../dialogue/DialogueUI";
import type { LoadoutStore } from "./LoadoutStore";
import { EquipmentUI } from "./EquipmentUI";
import { InventoryGrid } from "./InventoryGrid";
import { ItemTooltip } from "./ItemTooltip";

const FILTERS = ["ALL", "WEAPON", "ARMOR", "ACCESSORY", "MATERIAL", "CONSUMABLE", "QUEST"] as const;
const SORTS = ["RARITY", "LEVEL", "TYPE", "POWER", "RECENT"] as const;

export class InventoryUI {
  readonly root: HTMLElement;
  blocking = false;
  selected = -1;
  filter: (typeof FILTERS)[number] = "ALL";
  sort: (typeof SORTS)[number] = "RECENT";
  query = "";
  detail = false;
  rot = 0;
  zoom = 1;
  onUse: ((slot: number) => void) | null = null;
  onEquip: ((slot: number) => void) | null = null;
  onUnequip: ((equipSlot: string) => void) | null = null;
  onDiscard: ((slot: number) => void) | null = null;
  onLock: ((slot: number, on: boolean) => void) | null = null;
  onFavorite: ((slot: number, on: boolean) => void) | null = null;
  onSplit: ((slot: number, qty: number) => void) | null = null;
  onUpgrade: ((slot: number) => void) | null = null;
  onExpand: (() => void) | null = null;
  onLoadoutSave: ((slot: number) => void) | null = null;
  onLoadoutLoad: ((slot: number) => void) | null = null;
  onClaimTemp: (() => void) | null = null;
  onCosmetic: (() => void) | null = null;

  constructor(
    host: HTMLElement,
    private readonly store: LoadoutStore,
    private readonly playerName: () => string,
  ) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay inv-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => this.onClick(e));
    this.root.addEventListener("input", (e) => {
      const t = e.target as HTMLInputElement;
      if (t.dataset.search != null) {
        this.query = t.value;
        this.render();
      }
    });
    this.root.addEventListener("dragstart", (e) => {
      const slot = (e.target as HTMLElement).closest("[data-slot]") as HTMLElement | null;
      if (!slot || this.store.slot(Number(slot.dataset.slot)).qty < 1) return;
      e.dataTransfer?.setData("text/slot", slot.dataset.slot ?? "");
    });
    this.root.addEventListener("dragover", (e) => {
      if ((e.target as HTMLElement).closest("[data-equip]")) e.preventDefault();
    });
    this.root.addEventListener("drop", (e) => {
      const dest = (e.target as HTMLElement).closest("[data-equip]") as HTMLElement | null;
      const from = Number(e.dataTransfer?.getData("text/slot"));
      if (!dest || Number.isNaN(from)) return;
      e.preventDefault();
      this.onEquip?.(from);
    });
  }

  toggle(): void {
    if (this.blocking) this.close();
    else this.open();
  }

  open(): void {
    this.blocking = true;
    this.root.hidden = false;
    this.render();
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }

  render(): void {
    if (!this.blocking) return;
    const slot = this.selected >= 0 ? this.store.slot(this.selected) : null;
    const item = slot?.item && (slot.qty ?? 0) > 0 ? slot.item : null;
    const preview = item ? this.store.preview(item) : [];
    const compare = preview.length
      ? `<div class="eq-compare">${preview.map((r) => {
          const sign = r.delta > 0 ? "+" : "";
          const cls = r.delta > 0 ? "up" : r.delta < 0 ? "down" : "";
          return `<div class="${cls}">${sign}${fmt(r.delta)} ${r.label}</div>`;
        }).join("")}</div>`
      : "";
    const canUse = item?.type === "CONSUMABLE";
    const canEquip = item && (item.type === "EQUIPMENT" || item.type === "WEAPON" || item.type === "ARMOR" || item.type === "HELM" || item.type === "ACCESSORY");
    const rare = rarityAtLeast(item?.rarity, "RARE");
    const shown = this.visibleSlots();
    const temp = this.store.tempLoot ?? [];
    const hist = (this.store.itemHistory ?? []).slice(-6).reverse();
    this.root.innerHTML = `
      <div class="rpg-card inv-card">
        <header>
          <h3>INVENTORY</h3>
          <span>POWER ${this.store.stats?.powerRating ?? 0} · ${this.store.slots.length} slot</span>
        </header>
        <div class="inv-toolbar">
          ${FILTERS.map((f) => `<button type="button" data-filter="${f}" class="${this.filter === f ? "on" : ""}">${f}</button>`).join("")}
          <select data-sort>
            ${SORTS.map((s) => `<option ${this.sort === s ? "selected" : ""}>${s}</option>`).join("")}
          </select>
          <input data-search type="search" placeholder="Cari item" value="${escapeHtml(this.query)}" />
        </div>
        <div class="inv-layout">
          <aside class="inv-preview">
            <div class="char-silhouette" style="transform:rotateY(${this.rot}deg) scale(${this.zoom})">
              <div class="sil-head"></div>
              <div class="sil-body"></div>
              <div class="sil-legs"></div>
            </div>
            <div class="preview-acts">
              <button type="button" data-act="rot-l">◀</button>
              <button type="button" data-act="zoom-in">+</button>
              <button type="button" data-act="zoom-out">−</button>
              <button type="button" data-act="rot-r">▶</button>
            </div>
            <div class="inv-name">${escapeHtml(this.playerName())}</div>
            <div>Lv ${this.store.stats?.level ?? 1} · Combat Power ${this.store.stats?.powerRating ?? 0}</div>
            <div class="item-set">DAWN SET ${this.store.setPieces}/4</div>
            ${EquipmentUI.html(this.store, null)}
            <div class="loadout-row">
              <button type="button" data-act="save-0">Save 1</button>
              <button type="button" data-act="load-0">Load 1</button>
              <button type="button" data-act="save-1">Save 2</button>
              <button type="button" data-act="load-1">Load 2</button>
              <button type="button" data-act="save-2">Save 3</button>
              <button type="button" data-act="load-2">Load 3</button>
            </div>
          </aside>
          <section class="inv-center">
            ${InventoryGrid.html(shown, this.selected)}
            ${temp.length ? `<div class="temp-loot">Temporary loot ${temp.map((r) => escapeHtml(r.name) + " x" + r.qty).join(", ")} <button type="button" data-act="claim-temp">Ambil</button></div>` : ""}
          </section>
          <aside class="inv-info">
            ${ItemTooltip.html(item, slot?.qty ?? 0, { upgrade: slot?.upgrade, instanceId: slot?.itemInstanceId, setPieces: this.store.setPieces })}
            ${compare}
            ${this.detail && item ? `<div class="item-detail">Instance ${escapeHtml(slot?.itemInstanceId ?? "-")}</div>` : ""}
            ${item && (item.type === "EQUIPMENT" || item.type === "WEAPON") ? `<div class="up-preview">Upgrade +${(slot?.upgrade ?? 0) + 1} · Enhancement Stone x${(slot?.upgrade ?? 0) + 1}</div>` : ""}
            ${hist.length ? `<div class="item-hist">${hist.map((h) => `<div>${escapeHtml(h.action)} ${escapeHtml(h.itemId)}</div>`).join("")}</div>` : ""}
          </aside>
        </div>
        <footer class="inv-actions">
          <button type="button" data-act="use" ${canUse ? "" : "disabled"}>Use</button>
          <button type="button" data-act="equip" ${canEquip ? "" : "disabled"}>EQUIP</button>
          <button type="button" data-act="compare" ${canEquip ? "" : "disabled"}>COMPARE</button>
          <button type="button" data-act="detail" ${item ? "" : "disabled"}>DETAIL</button>
          <button type="button" data-act="upgrade" ${canEquip ? "" : "disabled"}>Upgrade</button>
          <button type="button" data-act="split" ${item?.stackable && (slot?.qty ?? 0) > 1 ? "" : "disabled"}>Split</button>
          <button type="button" data-act="lock" ${item ? "" : "disabled"}>${slot?.locked ? "Unlock" : "Lock"}</button>
          <button type="button" data-act="fav" ${item ? "" : "disabled"}>${slot?.favorite ? "Unfav" : "Favorite"}</button>
          <button type="button" data-act="discard" ${item ? "" : "disabled"}>${rare ? "Discard…" : "Discard"}</button>
          <button type="button" data-act="expand">Expand Bag</button>
          <button type="button" data-act="cosmetic">${this.store.stats?.showCosmetic ? "Hide Cosmetic" : "Show Cosmetic"}</button>
          <button type="button" data-act="close">Close</button>
        </footer>
      </div>`;
    const sel = this.root.querySelector("[data-sort]") as HTMLSelectElement | null;
    if (sel) sel.onchange = () => {
      this.sort = sel.value as (typeof SORTS)[number];
      this.render();
    };
  }

  private visibleSlots() {
    const rank: Record<string, number> = { COMMON: 1, UNCOMMON: 2, RARE: 3, EPIC: 4, LEGENDARY: 5, MYTHIC: 6 };
    const q = this.query.trim().toLowerCase();
    const mapped = this.store.slots.map((s, i) => ({ ...s, index: s.index ?? i }));
    const filtered = mapped.filter((s) => {
      const item = s.item;
      if (!item || s.qty < 1) return this.filter === "ALL" && !q;
      if (q && !item.name.toLowerCase().includes(q)) return false;
      if (this.filter === "ALL") return true;
      if (this.filter === "WEAPON") return item.type === "WEAPON" || item.slot === "WEAPON";
      if (this.filter === "ARMOR") return item.slot === "BODY" || item.slot === "HEAD" || item.slot === "LEGS" || item.type === "ARMOR" || item.type === "HELM";
      if (this.filter === "ACCESSORY") return (item.slot ?? "").startsWith("ACCESSORY") || item.type === "ACCESSORY";
      if (this.filter === "MATERIAL") return item.type === "MATERIAL";
      if (this.filter === "CONSUMABLE") return item.type === "CONSUMABLE";
      if (this.filter === "QUEST") return item.type === "QUEST" || item.type === "QUEST_ITEM";
      return true;
    });
    const filled = filtered.filter((s) => s.item && s.qty > 0);
    const empty = this.filter === "ALL" && !q ? mapped.filter((s) => !s.item || s.qty < 1) : [];
    filled.sort((a, b) => {
      const ia = a.item!;
      const ib = b.item!;
      if (this.sort === "RARITY") return (rank[ib.rarity] ?? 0) - (rank[ia.rarity] ?? 0);
      if (this.sort === "LEVEL") return (ib.itemLevel ?? 0) - (ia.itemLevel ?? 0);
      if (this.sort === "TYPE") return ia.type.localeCompare(ib.type);
      if (this.sort === "POWER") return (ib.effects?.attack ?? 0) + (ib.effects?.defense ?? 0) - ((ia.effects?.attack ?? 0) + (ia.effects?.defense ?? 0));
      return b.index - a.index;
    });
    return [...filled, ...empty];
  }

  private onClick(e: MouseEvent): void {
    const t = e.target as HTMLElement;
    const filt = t.closest("[data-filter]") as HTMLElement | null;
    if (filt?.dataset.filter) {
      this.filter = filt.dataset.filter as (typeof FILTERS)[number];
      this.render();
      return;
    }
    const cell = t.closest("[data-slot]") as HTMLElement | null;
    if (cell) {
      this.selected = Number(cell.dataset.slot);
      this.render();
      return;
    }
    const eq = t.closest("[data-equip]") as HTMLElement | null;
    if (eq?.dataset.equip) {
      this.onUnequip?.(eq.dataset.equip);
      return;
    }
    const act = t.closest("[data-act]") as HTMLElement | null;
    if (!act) return;
    const a = act.dataset.act ?? "";
    if (a === "close") this.close();
    if (a === "use" && this.selected >= 0) this.onUse?.(this.selected);
    if (a === "equip" && this.selected >= 0) this.onEquip?.(this.selected);
    if (a === "compare") this.render();
    if (a === "detail") {
      this.detail = !this.detail;
      this.render();
    }
    if (a === "upgrade" && this.selected >= 0) this.onUpgrade?.(this.selected);
    if (a === "split" && this.selected >= 0) {
      const qty = Math.floor((this.store.slot(this.selected).qty ?? 0) / 2);
      if (qty > 0) this.onSplit?.(this.selected, qty);
    }
    if (a === "lock" && this.selected >= 0) this.onLock?.(this.selected, !this.store.slot(this.selected).locked);
    if (a === "fav" && this.selected >= 0) this.onFavorite?.(this.selected, !this.store.slot(this.selected).favorite);
    if (a === "discard" && this.selected >= 0) {
      const it = this.store.slot(this.selected).item;
      if (it && rarityAtLeast(it.rarity, "RARE") && !window.confirm(`Buang ${it.name}?`)) return;
      this.onDiscard?.(this.selected);
    }
    if (a === "expand") this.onExpand?.();
    if (a === "cosmetic") this.onCosmetic?.();
    if (a === "claim-temp") this.onClaimTemp?.();
    if (a === "rot-l") {
      this.rot -= 20;
      this.render();
    }
    if (a === "rot-r") {
      this.rot += 20;
      this.render();
    }
    if (a === "zoom-in") {
      this.zoom = Math.min(1.6, this.zoom + 0.1);
      this.render();
    }
    if (a === "zoom-out") {
      this.zoom = Math.max(0.7, this.zoom - 0.1);
      this.render();
    }
    if (a.startsWith("save-")) this.onLoadoutSave?.(Number(a.slice(5)));
    if (a.startsWith("load-")) this.onLoadoutLoad?.(Number(a.slice(5)));
  }
}

function fmt(n: number): string {
  if (Math.abs(n) >= 1 || n === 0) return String(Math.round(n * 100) / 100);
  return n.toFixed(2);
}

function rarityAtLeast(rarity: string | undefined, min: string): boolean {
  const rank: Record<string, number> = { COMMON: 1, UNCOMMON: 2, RARE: 3, EPIC: 4, LEGENDARY: 5, MYTHIC: 6 };
  return (rank[rarity ?? ""] ?? 0) >= (rank[min] ?? 99);
}

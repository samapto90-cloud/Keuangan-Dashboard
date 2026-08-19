import { escapeHtml } from "../dialogue/DialogueUI";
import type { LoadoutStore } from "../inventory/LoadoutStore";
import type { ShopCatalogOut } from "../network/NetworkMessage";

export class ShopUI {
  readonly root: HTMLElement;
  blocking = false;
  catalog: ShopCatalogOut | null = null;
  onBuy: ((shopItemId: string, itemId: string, tx: string) => void) | null = null;
  onSell: ((slot: number, itemId: string, qty: number) => void) | null = null;

  constructor(
    host: HTMLElement,
    private readonly store: LoadoutStore,
  ) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay shop-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = (e.target as HTMLElement).closest("[data-shop]") as HTMLElement | null;
      if (!t) return;
      if (t.dataset.shop === "close") this.close();
      if (t.dataset.shop === "buy" && t.dataset.id) {
        this.onBuy?.(t.dataset.id, t.dataset.item || "", `buy-${t.dataset.id}-${Date.now()}`);
      }
      if (t.dataset.shop === "sell" && t.dataset.slot) {
        this.onSell?.(Number(t.dataset.slot), t.dataset.item || "", 1);
      }
    });
  }

  apply(cat: ShopCatalogOut): void {
    this.catalog = cat;
    this.open();
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
    const wallet = this.catalog?.wallet;
    const items = (this.catalog?.items ?? []).map((it) => `
      <div class="social-row">
        <div><b>${escapeHtml(it.name)}</b><small>${it.price} ${escapeHtml(it.currency)} · ${it.bought}/${it.purchaseLimit || "∞"} today</small></div>
        <button type="button" data-shop="buy" data-id="${escapeHtml(it.shopItemId)}" data-item="${escapeHtml(it.itemId)}">BUY</button>
      </div>`).join("");
    const bag = this.store.slots.filter((s) => s.item && s.qty > 0 && s.item.tradable !== false).map((s) => `
      <div class="social-row">
        <span>${escapeHtml(s.item!.name)} x${s.qty}</span>
        <button type="button" data-shop="sell" data-slot="${s.index}" data-item="${escapeHtml(s.item!.id)}">SELL</button>
      </div>`).join("");
    this.root.innerHTML = `
      <div class="rpg-card shop-card">
        <header><h3>${escapeHtml(this.catalog?.name || "Dawn Merchant")}</h3><button type="button" data-shop="close">Close</button></header>
        <p>Coin ${wallet?.coins ?? this.store.coin} · Crystal ${wallet?.crystals ?? this.store.crystal}</p>
        <div class="social-body">${items || "<p>Tidak ada barang.</p>"}</div>
        <h4>SELL</h4>
        <div class="social-body">${bag || "<p>Tidak ada item yang bisa dijual.</p>"}</div>
      </div>`;
  }
}

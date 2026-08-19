import { escapeHtml } from "../dialogue/DialogueUI";
import type { MarketState } from "../network/NetworkMessage";
import type { LoadoutStore } from "../inventory/LoadoutStore";

export function marketHtml(state: MarketState | null, store: LoadoutStore): string {
  const listings = (state?.listings ?? []).map((l) => {
    const item = store.item(l.itemId);
    const mine = l.sellerId === "";
    return `<article class="market-card rarity-${escapeHtml((l.rarity || "COMMON").toLowerCase())}">
      <h4>${escapeHtml(item?.name || l.itemId)}</h4>
      <p>${escapeHtml(l.rarity)} · x${l.qty} · Lv${l.level}</p>
      <p class="market-price">${l.price} Coin</p>
      <p>Seller ${escapeHtml(l.seller)}</p>
      <div class="social-acts">
        <button type="button" data-soc="market-buy" data-id="${escapeHtml(l.id)}">BUY</button>
        ${mine || true ? `<button type="button" data-soc="market-cancel" data-id="${escapeHtml(l.id)}">CANCEL</button>` : ""}
      </div>
    </article>`;
  }).join("") || "<p>Tidak ada listing.</p>";
  const hist = (state?.history ?? []).slice(-8).reverse().map((h) =>
    `<div class="social-row"><span>${escapeHtml(h.kind)} ${escapeHtml(h.itemId)} x${h.qty} · ${h.price}c</span></div>`).join("");
  const bag = store.slots.filter((s) => s.item && (s.item.tradable !== false)).slice(0, 8).map((s) =>
    `<button type="button" data-soc="market-list" data-slot="${s.index}">List ${escapeHtml(s.item?.name || "")} x${s.qty}</button>`).join("");
  return `
    <h4>MARKETPLACE</h4>
    <p>Fee 5% · 20 per page</p>
    <div class="market-filters">
      <select name="sort">
        <option value="price">Price low-high</option>
        <option value="price_desc">Price high-low</option>
        <option value="newest">Newest</option>
      </select>
      <button type="button" data-soc="market-refresh">REFRESH</button>
    </div>
    <div class="market-grid">${listings}</div>
    <h4>YOUR BAG</h4>
    <div class="inv-actions">${bag || "<p>Tidak ada item tradeable.</p>"}</div>
    <h4>HISTORY</h4>
    ${hist || "<p>Kosong.</p>"}`;
}

import { escapeHtml } from "../dialogue/DialogueUI";
import type { CraftingRecipeView, CraftingState, MaterialView, ProfessionView } from "../network/NetworkMessage";

export class CraftingUI {
  readonly root: HTMLElement;
  blocking = false;
  tab: "prof" | "book" | "craft" | "bag" | "shop" | "stall" | "workshop" | "fest" = "prof";
  query = "";
  filter = "ALL";
  view: CraftingState | null = null;
  selected = "";
  shopId = "mbah-karya-shop";
  onCraft: ((recipeId: string, stationId: string) => void) | null = null;
  onProfession: ((id: string) => void) | null = null;
  onReset: (() => void) | null = null;
  onShopBuy: ((shopId: string, itemId: string) => void) | null = null;
  onShopSell: ((shopId: string, itemId: string) => void) | null = null;
  onStallList: ((itemId: string, price: number) => void) | null = null;
  onStallBuy: ((sellerId: string, itemId: string) => void) | null = null;
  onWorkshop: (() => void) | null = null;
  onContribute: (() => void) | null = null;
  stationId = "station-forge";

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay craft-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => this.onClick(e));
    this.root.addEventListener("input", (e) => {
      const el = e.target as HTMLInputElement;
      if (el.dataset.cr === "search") {
        this.query = el.value;
        this.render();
      }
    });
  }

  apply(view: CraftingState): void {
    this.view = view;
    if (this.blocking) this.render();
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

  private onClick(e: Event): void {
    const t = (e.target as HTMLElement).closest("[data-cr]") as HTMLElement | null;
    if (!t) return;
    const act = t.dataset.cr || "";
    if (act === "close") this.close();
    if (act === "tab" && t.dataset.id) {
      this.tab = t.dataset.id as typeof this.tab;
      this.render();
    }
    if (act === "prof" && t.dataset.id) this.onProfession?.(t.dataset.id);
    if (act === "reset") this.onReset?.();
    if (act === "pick" && t.dataset.id) {
      this.selected = t.dataset.id;
      this.tab = "craft";
      this.render();
    }
    if (act === "station" && t.dataset.id) {
      this.stationId = t.dataset.id;
      this.render();
    }
    if (act === "craft" && this.selected) this.onCraft?.(this.selected, this.stationId);
    if (act === "buy" && t.dataset.id) {
      if (t.dataset.rare === "1" && !window.confirm("Beli item langka?")) return;
      this.onShopBuy?.(this.shopId, t.dataset.id);
    }
    if (act === "sell" && t.dataset.id) {
      if ((Number(t.dataset.qty) || 0) > 8 && !window.confirm("Jual bahan berharga?")) return;
      this.onShopSell?.(this.shopId, t.dataset.id);
    }
    if (act === "filter" && t.dataset.id) {
      this.filter = t.dataset.id;
      this.render();
    }
    if (act === "shop" && t.dataset.id) {
      this.shopId = t.dataset.id;
      this.render();
    }
    if (act === "stall-list" && t.dataset.id) this.onStallList?.(t.dataset.id, 5);
    if (act === "stall-buy" && t.dataset.seller && t.dataset.id) this.onStallBuy?.(t.dataset.seller, t.dataset.id);
    if (act === "workshop") this.onWorkshop?.();
    if (act === "contrib") this.onContribute?.();
  }

  render(): void {
    const v = this.view;
    const tabs = [
      ["prof", "Profesi"],
      ["book", "Resep"],
      ["craft", "Craft"],
      ["bag", "Bahan"],
      ["shop", "Toko"],
      ["stall", "Stall"],
      ["workshop", "Workshop"],
      ["fest", "Festival"],
    ];
    this.root.innerHTML = `
      <div class="rpg-card craft-card">
        <header><h3>KARYA DESA</h3><button type="button" data-cr="close">Tutup</button></header>
        <p class="craft-wallet">GOLD ${v?.goldBalance ?? v?.gold ?? 0} · Knowledge Token ${v?.knowledgeToken ?? 0}</p>
        <nav class="craft-tabs">${tabs.map(([id, label]) => `<button type="button" data-cr="tab" data-id="${id}" class="${this.tab === id ? "on" : ""}">${label}</button>`).join("")}</nav>
        <div class="craft-body">${this.body(v)}</div>
      </div>`;
  }

  private body(v: CraftingState | null): string {
    if (!v) return "<p>Memuat...</p>";
    if (this.tab === "prof") return this.profHtml(v.professions || []);
    if (this.tab === "book") return this.bookHtml(v.recipes || []);
    if (this.tab === "craft") return this.craftHtml(v);
    if (this.tab === "bag") return this.bagHtml(v);
    if (this.tab === "shop") return this.shopHtml(v);
    if (this.tab === "stall") return this.stallHtml(v);
    if (this.tab === "workshop") {
      return `<p>Guild Workshop: ${v.workshop ? "AKTIF" : "belum"} — bonus waktu craft, bukan kekuatan tempur.</p>
        <button type="button" data-cr="workshop">Buka Workshop</button>
        <button type="button" data-cr="contrib">Sumbang Iron Ore</button>
        <h4>Contribution Log</h4>
        <div class="craft-list">${(v.guildContrib || []).slice(-8).map((c) => `<div>${escapeHtml(c.playerId)} · ${escapeHtml(c.itemId)} x${c.qty}</div>`).join("") || "<p>—</p>"}</div>`;
    }
    return `<div class="fest-banner"><h4>${escapeHtml(v.festivalName || "FESTIVAL KARYA")}</h4>
      <p>${v.festival ? "Festival sedang berlangsung. Craft, masak, mancing, lan pamer karya." : "Festival Karya akan datang ke alun-alun."}</p>
      <p>Hadiah: cosmetic, title, Knowledge Token.</p></div>`;
  }

  private profHtml(list: ProfessionView[]): string {
    return `<p>Aktif: 3 gathering + 2 crafting. Reset terbatas. Max Lv 50.</p>
      <div class="craft-grid">${list.map((p) => `
        <button type="button" data-cr="prof" data-id="${escapeHtml(p.id)}" class="${p.active ? "on" : ""}">
          <strong>${escapeHtml(p.name)}</strong>
          <span>Lv ${p.level} · XP ${p.xp} · Next ${p.next ?? 0}</span>
          <em>${p.active ? "AKTIF" : p.kind}</em>
        </button>`).join("")}</div>
      <p>Achievement: First Harvest · First Mine · First Fish · First Craft · Master Miner · Master Fisher · Master Crafter</p>
      <button type="button" data-cr="reset">Reset profesi</button>`;
  }

  private bookHtml(list: CraftingRecipeView[]): string {
    const cats = ["ALL", "WEAPON", "ARMOR", "ACCESSORY", "POTION", "FOOD", "MATERIAL"];
    const q = this.query.trim().toLowerCase();
    const shown = list.filter((r) => {
      if (this.filter !== "ALL" && r.category !== this.filter) return false;
      if (q && !r.name.toLowerCase().includes(q)) return false;
      return true;
    });
    return `<div class="inv-actions">${cats.map((c) => `<button type="button" data-cr="filter" data-id="${c}" class="${this.filter === c ? "on" : ""}">${c}</button>`).join("")}</div>
      <input data-cr="search" type="search" placeholder="Cari resep" value="${escapeHtml(this.query)}" />
      <div class="craft-list">${shown.map((r) => `
      <button type="button" data-cr="pick" data-id="${escapeHtml(r.id)}">
        <strong>${escapeHtml(r.name)}</strong>
        <span>${escapeHtml(r.category)} · ${escapeHtml(r.status)} · Lv ${r.requiredLevel}</span>
      </button>`).join("")}</div>`;
  }

  private craftHtml(v: CraftingState): string {
    const rec = (v.recipes || []).find((r) => r.id === this.selected);
    const stations = [
      ["station-forge", "BLACKSMITH"],
      ["station-bench", "WORKSHOP"],
      ["station-alchemy", "ALCHEMY"],
      ["station-cook", "COOKING FIRE"],
    ];
    if (!rec) return `<p>Pilih resep di buku.</p><div>${stations.map(([id, n]) => `<button type="button" data-cr="station" data-id="${id}" class="${this.stationId === id ? "on" : ""}">${n}</button>`).join("")}</div>`;
    const mats = (rec.materials || []).map((m) => {
      const id = m.itemId || m.ItemID || "";
      const need = m.qty ?? m.Qty ?? 1;
      const have = rec.owned?.[id] ?? 0;
      return `<li>${escapeHtml(id)} ${have}/${need}</li>`;
    }).join("");
    return `<div class="craft-detail">
      <h4>${escapeHtml(rec.name)}</h4>
      <p>Hasil: ${escapeHtml(rec.result)} · ${escapeHtml(rec.status)} · ${rec.craftTime ? rec.craftTime + " dtk" : "instan"}</p>
      <ul>${mats}</ul>
      <div>${stations.map(([id, n]) => `<button type="button" data-cr="station" data-id="${id}" class="${this.stationId === id ? "on" : ""}">${n}</button>`).join("")}</div>
      <p>Fee: Gold · Hasil ke inventory. Tas penuh = batal, bahan kembali.</p>
      <p class="craft-fx">✦ menempa...</p>
      <button type="button" data-cr="craft">CRAFT</button>
    </div>`;
  }

  private bagHtml(v: CraftingState): string {
    const mats = v.materials || [];
    const codex = v.codex || [];
    return `<h4>Material Bag (stack ${99} / siap 999)</h4>
      <div class="craft-list">${mats.map((m: MaterialView) => `<div><strong>${escapeHtml(m.name)}</strong> x${m.qty}<p>${escapeHtml(m.desc || "")}</p></div>`).join("") || "<p>Kosong.</p>"}</div>
      <h4>Kodeks Bahan</h4>
      <div class="craft-list">${codex.map((c) => `<div><strong>${escapeHtml(c.name)}</strong><p>${escapeHtml(c.desc || "")}</p></div>`).join("") || "<p>Jelajahi wilayah untuk menemukan bahan.</p>"}</div>`;
  }

  private shopHtml(v: CraftingState): string {
    const mats = v.materials || [];
    const shop = (v.merchants || []).find((m) => m.id === this.shopId) || (v.merchants || [])[0];
    const items = shop?.items || [];
    const hist = (v.goldHistory || []).slice(-6).map((h) => `<div>${escapeHtml(h.source)} ${h.amount} → ${escapeHtml(h.destination)}</div>`).join("");
    return `<div class="inv-actions shop-highlight">
      <button type="button" data-cr="shop" data-id="mbah-karya-shop" class="${this.shopId === "mbah-karya-shop" ? "on" : ""}">Mbah Karya</button>
      <button type="button" data-cr="shop" data-id="mbok-rasa-shop" class="${this.shopId === "mbok-rasa-shop" ? "on" : ""}">Mbok Rasa</button>
      <button type="button" data-cr="shop" data-id="pedagang-shop" class="${this.shopId === "pedagang-shop" ? "on" : ""}">Pedagang</button>
    </div>
    <p>GOLD ${v.goldBalance ?? v.gold ?? 0}</p>
    <div class="craft-grid">${items.map((it) => `
      <button type="button" data-cr="buy" data-id="${escapeHtml(it.itemId)}" data-rare="${it.rarity === "RARE" || it.rarity === "EPIC" ? "1" : "0"}">
        Beli ${escapeHtml(it.name || it.itemId)} · ${it.buyPrice ?? it.price ?? 0}g
      </button>`).join("") || `
      <button type="button" data-cr="buy" data-id="mat-valley-stone">Beli Valley Stone</button>
      <button type="button" data-cr="buy" data-id="mat-mistwood">Beli Mistwood</button>
      <button type="button" data-cr="buy" data-id="mat-dawn-berry">Beli Dawn Berry</button>
      <button type="button" data-cr="buy" data-id="mat-iron-ore">Beli Iron Ore</button>
      <button type="button" data-cr="buy" data-id="sacred_wood" data-rare="1">Beli Sacred Wood</button>
      <button type="button" data-cr="buy" data-id="dawn_herb">Beli Dawn Herb</button>`}
    </div>
    <h4>Jual</h4>
    <div class="craft-list">${mats.map((m) => `<button type="button" data-cr="sell" data-id="${escapeHtml(m.id)}" data-qty="${m.qty}">Jual ${escapeHtml(m.name)} x${m.qty}${items.find((it) => it.itemId === m.id)?.sellPrice ? ` · ${items.find((it) => it.itemId === m.id)?.sellPrice}g` : ""}</button>`).join("") || "<p>Tidak ada bahan.</p>"}</div>
    <h4>Economy Log</h4>
    <div class="craft-list">${hist || "<p>—</p>"}</div>`;
  }

  private stallHtml(v: CraftingState): string {
    const mine = (v.stall || []).map((s) => `<div>${escapeHtml(s.itemId)} x${s.qty} · ${s.price} gold</div>`).join("");
    const others = (v.stalls || []).map((s) => (s.items || []).map((it) =>
      `<button type="button" data-cr="stall-buy" data-seller="${escapeHtml(s.playerId)}" data-id="${escapeHtml(it.itemId)}">${escapeHtml(it.itemId)} · ${it.price}g</button>`).join("")).join("");
    const bag = (v.materials || []).map((m) => `<button type="button" data-cr="stall-list" data-id="${escapeHtml(m.id)}">Pasang ${escapeHtml(m.name)}</button>`).join("");
    return `<p>Player Stall. Bukan auction house. Item quest-bound ditolak server.</p>
      <h4>Stall kamu ${v.stallOpen ? "(buka)" : ""}</h4>${mine || "<p>—</p>"}
      <h4>Pasang dari tas bahan</h4>${bag || "<p>Kosong.</p>"}
      <h4>Stall pemain lain</h4>${others || "<p>—</p>"}`;
  }
}

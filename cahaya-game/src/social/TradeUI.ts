import { escapeHtml } from "../dialogue/DialogueUI";
import type { LoadoutStore } from "../inventory/LoadoutStore";
import type { TradeView } from "../network/NetworkMessage";

export class TradeUI {
  readonly root: HTMLElement;
  blocking = false;
  view: TradeView | null = null;
  onReady: (() => void) | null = null;
  onConfirm: ((tx: string) => void) | null = null;
  onCancel: (() => void) | null = null;
  onOffer: ((slots: Array<{ slot: number; itemId: string; qty: number }>, coin: number) => void) | null = null;

  constructor(
    host: HTMLElement,
    private readonly store: LoadoutStore,
    private readonly selfId: () => string,
  ) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay trade-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = (e.target as HTMLElement).closest("[data-tr]") as HTMLElement | null;
      if (!t) return;
      if (t.dataset.tr === "ready") this.onReady?.();
      if (t.dataset.tr === "confirm" && this.view) this.onConfirm?.(this.view.transactionId);
      if (t.dataset.tr === "cancel") {
        this.onCancel?.();
        this.close();
      }
      if (t.dataset.tr === "offer") this.sendOffer();
    });
  }

  apply(view: TradeView): void {
    this.view = view;
    if (view.state === "CANCEL" || view.state === "DONE") {
      this.close();
      return;
    }
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

  private sendOffer(): void {
    const coin = Number((this.root.querySelector("[name=coin]") as HTMLInputElement | null)?.value || 0);
    const slots: Array<{ slot: number; itemId: string; qty: number }> = [];
    this.root.querySelectorAll<HTMLSelectElement>("[data-slot-pick]").forEach((sel, i) => {
      if (!sel.value) return;
      const [slot, itemId] = sel.value.split(":");
      slots.push({ slot: Number(slot), itemId, qty: 1 });
      void i;
    });
    this.onOffer?.(slots.slice(0, 5), Math.max(0, Math.floor(coin)));
  }

  render(): void {
    const v = this.view;
    if (!v) {
      this.root.innerHTML = "";
      return;
    }
    const self = this.selfId();
    const mine = v.a.playerId === self ? v.a : v.b;
    const theirs = v.a.playerId === self ? v.b : v.a;
    const bag = this.store.slots
      .filter((s) => s.item && s.qty > 0 && s.item.tradable !== false)
      .map((s) => `<option value="${s.index}:${escapeHtml(s.item!.id)}">${escapeHtml(s.item!.name)} x${s.qty}</option>`)
      .join("");
    const locked = Boolean(mine.confirm || theirs.confirm);
    this.root.innerHTML = `
      <div class="rpg-card trade-card">
        <header><h3>${locked ? "FINAL TRADE CONFIRMATION" : "TRADE"}</h3><button type="button" data-tr="cancel">Cancel</button></header>
        <div class="trade-cols">
          <div>
            <h4>You ${mine.ready ? "· READY" : ""} ${mine.confirm ? "· CONFIRM" : ""}</h4>
            ${[0, 1, 2, 3, 4].map((i) => `<select data-slot-pick ${locked ? "disabled" : ""}><option value="">Item ${i + 1}</option>${bag}</select>`).join("")}
            <label>Coin <input name="coin" type="number" min="0" value="${mine.coin}" ${locked ? "disabled" : ""} /></label>
            <button type="button" data-tr="offer" ${locked ? "disabled" : ""}>Update offer</button>
          </div>
          <div>
            <h4>Partner ${theirs.ready ? "· READY" : ""} ${theirs.confirm ? "· CONFIRM" : ""}</h4>
            ${(theirs.slots || []).map((s) => `<div>${escapeHtml(s.itemId)} x${s.qty}</div>`).join("") || "<p>—</p>"}
            <div>Coin ${theirs.coin}</div>
          </div>
        </div>
        <footer class="inv-actions">
          <button type="button" data-tr="ready" ${locked ? "disabled" : ""}>READY / CONFIRM</button>
          <button type="button" data-tr="confirm" ${mine.ready && theirs.ready ? "" : "disabled"}>${locked ? "FINAL CONFIRM" : "CONFIRM"}</button>
        </footer>
      </div>`;
  }
}

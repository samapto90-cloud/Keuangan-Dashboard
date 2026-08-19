import { escapeHtml } from "../dialogue/DialogueUI";
import type { LootItemView, LootResult } from "../network/NetworkMessage";

export class DungeonRewardUI {
  readonly root: HTMLElement;
  onClaim: ((claimId: string) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "dun-overlay dun-reward";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const btn = (e.target as HTMLElement).closest("[data-claim]") as HTMLElement | null;
      if (btn?.dataset.claim) this.onClaim?.(btn.dataset.claim);
    });
  }

  show(loot: LootResult): void {
    this.root.hidden = false;
    const items = (loot.items ?? []).map((it) => row(it)).join("") || "<div>Tidak ada item.</div>";
    this.root.innerHTML = `
      <div class="dun-card">
        <div class="dun-kicker">PERSONAL LOOT</div>
        <h3>Hadiah Dungeon</h3>
        <p>+${loot.exp || 0} EXP · +${loot.coin || 0} KOIN${loot.crystal ? ` · +${loot.crystal} CRYSTAL` : ""}</p>
        <div class="dun-loot">${items}</div>
        <button type="button" class="dun-btn" data-claim="${escapeHtml(loot.claimId)}">AMBIL HADIAH</button>
      </div>`;
  }

  hide(): void {
    this.root.hidden = true;
  }
}

function row(it: LootItemView): string {
  return `<div class="dun-loot-row ${escapeHtml(it.rarity || "").toLowerCase()}"><span>${escapeHtml(it.name)}</span><b>${escapeHtml(it.rarity || "COMMON")} · x${it.qty}</b></div>`;
}

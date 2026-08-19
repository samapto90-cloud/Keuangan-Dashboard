import { escapeHtml } from "../dialogue/DialogueUI";

export class PartyInviteUI {
  readonly root: HTMLElement;
  blocking = false;
  onAccept: (() => void) | null = null;
  onDecline: (() => void) | null = null;
  private kind: "party" | "friend" | "guild" | "trade" = "party";
  private from = "";
  fromId = "";

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay invite-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const act = (e.target as HTMLElement).closest("[data-inv]") as HTMLElement | null;
      if (!act) return;
      if (act.dataset.inv === "accept") this.onAccept?.();
      else this.onDecline?.();
      this.close();
    });
  }

  showParty(from: string, fromId: string): void {
    this.kind = "party";
    this.from = from;
    this.fromId = fromId;
    this.open();
  }

  showFriend(from: string, fromId: string): void {
    this.kind = "friend";
    this.from = from;
    this.fromId = fromId;
    this.open();
  }

  showGuild(from: string, fromId: string): void {
    this.kind = "guild";
    this.from = from;
    this.fromId = fromId;
    this.open();
  }

  showTrade(from: string, fromId: string): void {
    this.kind = "trade";
    this.from = from;
    this.fromId = fromId;
    this.open();
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }

  private open(): void {
    this.blocking = true;
    this.root.hidden = false;
    const title = this.kind === "party" ? "PARTY INVITATION"
      : this.kind === "guild" ? "GUILD INVITATION"
        : this.kind === "trade" ? "TRADE REQUEST"
          : "FRIEND REQUEST";
    const text = this.kind === "party"
      ? `${this.from} invited you to a party.`
      : this.kind === "guild"
        ? `${this.from} invited you to a guild.`
        : this.kind === "trade"
          ? `${this.from} wants to trade.`
          : `${this.from} sent you a friend request.`;
    this.root.innerHTML = `
      <div class="rpg-card invite-card">
        <h3>${title}</h3>
        <p>${escapeHtml(text)}</p>
        <footer class="inv-actions">
          <button type="button" data-inv="accept">Accept</button>
          <button type="button" data-inv="decline">Decline</button>
        </footer>
      </div>`;
  }
}

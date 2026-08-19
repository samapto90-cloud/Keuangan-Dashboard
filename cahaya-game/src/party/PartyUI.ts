import { escapeHtml } from "../dialogue/DialogueUI";
import type { LoadoutStore } from "../inventory/LoadoutStore";
import type { FinderListing, PartyMemberView, PartyView } from "../network/NetworkMessage";

export class PartyUI {
  readonly hud: HTMLElement;
  readonly panel: HTMLElement;
  blocking = false;
  listings: FinderListing[] = [];
  onLeave: (() => void) | null = null;
  onKick: ((id: string) => void) | null = null;
  onDisband: (() => void) | null = null;
  onTransfer: ((id: string) => void) | null = null;
  onRole: ((role: string) => void) | null = null;
  onReady: (() => void) | null = null;
  onFollow: (() => void) | null = null;
  onWaypoint: ((id: string) => void) | null = null;
  onFinderCreate: ((activity: string, role: string, minLevel: number) => void) | null = null;
  onFinderJoin: ((id: string) => void) | null = null;
  onFinderList: (() => void) | null = null;
  onCreate: (() => void) | null = null;

  constructor(
    host: HTMLElement,
    private readonly store: LoadoutStore,
    private readonly selfId: () => string,
  ) {
    this.hud = document.createElement("div");
    this.hud.className = "party-hud";
    this.hud.hidden = true;
    this.panel = document.createElement("div");
    this.panel.className = "rpg-overlay party-overlay";
    this.panel.hidden = true;
    host.appendChild(this.hud);
    host.appendChild(this.panel);
    this.panel.addEventListener("click", (e) => {
      const t = (e.target as HTMLElement).closest("[data-party]") as HTMLElement | null;
      if (!t) return;
      if (t.dataset.party === "close") this.close();
      if (t.dataset.party === "leave") this.onLeave?.();
      if (t.dataset.party === "disband") this.onDisband?.();
      if (t.dataset.party === "kick" && t.dataset.id) this.onKick?.(t.dataset.id);
      if (t.dataset.party === "lead" && t.dataset.id) this.onTransfer?.(t.dataset.id);
      if (t.dataset.party === "role" && t.dataset.role) this.onRole?.(t.dataset.role);
      if (t.dataset.party === "ready") this.onReady?.();
      if (t.dataset.party === "follow") this.onFollow?.();
      if (t.dataset.party === "waypoint") this.onWaypoint?.("old-windmill");
      if (t.dataset.party === "list") this.onFinderList?.();
      if (t.dataset.party === "make") this.onCreate?.();
      if (t.dataset.party === "create") {
        const activity = (this.panel.querySelector("[name=activity]") as HTMLSelectElement)?.value || "Dungeon";
        const role = (this.panel.querySelector("[name=role]") as HTMLSelectElement)?.value || "FLEX";
        const min = Number((this.panel.querySelector("[name=minlv]") as HTMLInputElement)?.value || 1);
        this.onFinderCreate?.(activity, role, min);
      }
      if (t.dataset.party === "join" && t.dataset.id) this.onFinderJoin?.(t.dataset.id);
    });
  }

  toggle(): void {
    if (this.blocking) this.close();
    else this.open();
  }

  open(): void {
    this.blocking = true;
    this.panel.hidden = false;
    this.onFinderList?.();
    this.render();
  }

  close(): void {
    this.blocking = false;
    this.panel.hidden = true;
  }

  tick(): void {
    this.renderHud(this.store.party);
  }

  render(): void {
    this.tick();
    if (!this.blocking) return;
    const party = this.store.party;
    const self = this.selfId();
    const leader = party?.leaderId === self;
    const rows = (party?.members ?? []).map((m) => memberRow(m, leader, self)).join("");
    const target = party?.targetName
      ? `<div class="party-target">Target: ${escapeHtml(party.targetName)} Lv${party.targetLevel ?? 0}
          <div class="hud-bar hud-hp"><i style="width:${pct(party.targetHp, party.targetMaxHp)}%"></i></div></div>`
      : "";
    const list = this.listings.map((l) => `
      <div class="social-row">
        <div><b>${escapeHtml(l.leader)}</b> ${escapeHtml(l.activity)} Lv${l.level} ${escapeHtml(l.requiredRole || "FLEX")}
          <small>${l.players}/${l.cap}</small></div>
        <button type="button" data-party="join" data-id="${escapeHtml(l.id)}">JOIN</button>
      </div>`).join("");
    this.panel.innerHTML = `
      <div class="rpg-card party-card">
        <header><h3>PARTY</h3><button type="button" data-party="close">Close</button></header>
        ${party ? `<div class="party-list">${rows}</div>${target}` : `<p>Belum ada party.</p><button type="button" data-party="make">CREATE PARTY</button>`}
        <div class="social-acts">
          <button type="button" data-party="role" data-role="DPS">DPS</button>
          <button type="button" data-party="role" data-role="TANK">TANK</button>
          <button type="button" data-party="role" data-role="SUPPORT">SUPPORT</button>
          <button type="button" data-party="role" data-role="FLEX">FLEX</button>
          <button type="button" data-party="ready">READY</button>
          <button type="button" data-party="follow">FOLLOW PARTY</button>
          <button type="button" data-party="waypoint" ${leader ? "" : "disabled"}>PARTY WAYPOINT</button>
        </div>
        <footer class="inv-actions">
          <button type="button" data-party="leave" ${party ? "" : "disabled"}>Leave</button>
          <button type="button" data-party="disband" ${leader ? "" : "disabled"}>Disband</button>
        </footer>
        <h4>PARTY FINDER</h4>
        <select name="activity">
          <option>Dungeon</option><option>Chapter</option><option>World Event</option><option>Guardian Challenge</option>
        </select>
        <select name="role"><option>FLEX</option><option>DPS</option><option>TANK</option><option>SUPPORT</option></select>
        <input name="minlv" type="number" min="1" value="1" />
        <div class="inv-actions">
          <button type="button" data-party="create">LIST PARTY</button>
          <button type="button" data-party="list">REFRESH</button>
        </div>
        <div class="social-body">${list || "<p>Tidak ada listing.</p>"}</div>
      </div>`;
  }

  private renderHud(party: PartyView | null): void {
    if (!party?.members?.length) {
      this.hud.hidden = true;
      this.hud.innerHTML = "";
      return;
    }
    this.hud.hidden = false;
    this.hud.innerHTML = party.members
      .map((m) => {
        const crown = m.leader ? `<span class="crown">♛</span>` : "";
        const off = m.online ? "" : " offline";
        const down = (m.hp ?? 1) <= 0 || (m.status || "").toUpperCase().includes("DOWN") ? " DOWN" : "";
        return `<div class="party-hud-row${off}">
          <div>${crown}<b>${escapeHtml(m.name)}</b> Lv${m.level} ${escapeHtml(m.role || "")}${m.online ? "" : " OFFLINE"}${down}</div>
          <div class="hud-bar hud-hp"><i style="width:${pct(m.hp, m.maxHp)}%"></i></div>
          <div class="hud-bar hud-energy"><i style="width:${pct(m.energy, m.maxEnergy)}%"></i></div>
          <small>${Math.round(m.distance || 0)}m ${m.ready ? "READY" : "WAITING"}</small>
        </div>`;
      })
      .join("");
  }
}

function memberRow(m: PartyMemberView, leader: boolean, self: string): string {
  const crown = m.leader ? "♛ " : "";
  const kick = leader && m.playerId !== self ? `<button type="button" data-party="kick" data-id="${escapeHtml(m.playerId)}">Kick</button>` : "";
  const lead = leader && m.playerId !== self ? `<button type="button" data-party="lead" data-id="${escapeHtml(m.playerId)}">Leader</button>` : "";
  return `<div class="party-member ${m.online ? "" : "offline"}">
    <div class="party-av">${escapeHtml(m.name.slice(0, 1))}</div>
    <div>
      <b>${crown}${escapeHtml(m.name)}</b> Lv${m.level} ${escapeHtml(m.role || m.class)}
      <div class="hud-bar hud-hp"><i style="width:${pct(m.hp, m.maxHp)}%"></i></div>
      <small>${m.online ? `${Math.round(m.distance)}m ${m.status || ""}` : "OFFLINE"} ${m.ready ? "READY" : "WAITING"}</small>
    </div>
    ${kick}${lead}
  </div>`;
}

function pct(n?: number, max?: number): number {
  if (!max) return 0;
  return Math.max(0, Math.min(100, ((n ?? 0) / max) * 100));
}

import { escapeHtml } from "../dialogue/DialogueUI";
import type { NearbyView } from "../network/NetworkMessage";

export class NearbyPlayers {
  static html(list: NearbyView[]): string {
    if (!list.length) return `<h4>NEARBY</h4><p>Tidak ada pemain dalam 50m.</p>`;
    const rows = list.map((p) => `
      <div class="social-row">
        <div>
          <b>${escapeHtml(p.name)}</b> Lv${p.level} ${escapeHtml(p.class)}
          <small>${Math.round(p.distance)}m · ${p.online ? "ONLINE" : "OFFLINE"}</small>
        </div>
        <div class="social-acts">
          <button type="button" data-soc="inspect" data-id="${escapeHtml(p.playerId)}">Inspect</button>
          <button type="button" data-soc="invite" data-id="${escapeHtml(p.playerId)}">Invite</button>
          <button type="button" data-soc="trade" data-id="${escapeHtml(p.playerId)}">Trade</button>
          <button type="button" data-soc="friend" data-id="${escapeHtml(p.playerId)}">Add Friend</button>
          <button type="button" data-soc="block" data-id="${escapeHtml(p.playerId)}">Block</button>
        </div>
      </div>`).join("");
    return `<h4>NEARBY PLAYERS</h4>${rows}`;
  }
}

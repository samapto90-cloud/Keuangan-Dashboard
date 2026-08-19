import { escapeHtml } from "../dialogue/DialogueUI";
import type { FriendView } from "../network/NetworkMessage";

export class FriendList {
  static html(friends: FriendView[], pending: FriendView[], outgoing: FriendView[], blocked: string[], tab = "ONLINE"): string {
    const online = friends.filter((f) => f.online && (f.status || "ONLINE") !== "AWAY" && (f.status || "") !== "BUSY");
    const away = friends.filter((f) => f.online && ((f.status || "") === "AWAY" || (f.status || "") === "BUSY"));
    const offline = friends.filter((f) => !f.online);
    const shown = tab === "OFFLINE" ? offline : tab === "AWAY" ? away : tab === "REQUESTS" || tab === "BLOCKED" ? [] : online;
    const row = (f: FriendView, actions: string): string => `
      <div class="social-row">
        <div>
          <b>${escapeHtml(f.name)}</b> Lv${f.level} ${escapeHtml(f.guild || "")}
          <small>${escapeHtml(f.status || (f.online ? "ONLINE" : "OFFLINE"))}${f.title ? ` · ${escapeHtml(f.title)}` : ""}</small>
        </div>
        <div class="social-acts">${actions}</div>
      </div>`;
    const acts = (f: FriendView): string => `
      <button type="button" data-soc="invite" data-id="${escapeHtml(f.playerId)}">INVITE PARTY</button>
      <button type="button" data-soc="message" data-id="${escapeHtml(f.playerId)}" data-name="${escapeHtml(f.name)}">MESSAGE</button>
      <button type="button" data-soc="inspect" data-id="${escapeHtml(f.playerId)}">INSPECT</button>
      <button type="button" data-soc="trade" data-id="${escapeHtml(f.playerId)}">TRADE</button>
      <button type="button" data-soc="remove-friend" data-id="${escapeHtml(f.playerId)}">REMOVE</button>
      <button type="button" data-soc="block" data-id="${escapeHtml(f.playerId)}">BLOCK</button>`;
    const friendRows = shown.length
      ? shown.map((f) => row(f, f.online ? acts(f) : `<button type="button" data-soc="remove-friend" data-id="${escapeHtml(f.playerId)}">REMOVE</button>`)).join("")
      : tab === "ONLINE" ? "<p>Tidak ada teman online.</p>" : tab === "OFFLINE" ? "<p>Tidak ada teman offline.</p>" : tab === "AWAY" ? "<p>Tidak ada teman away.</p>" : "";
    const pend = pending.map((f) => row(f,
      `<button type="button" data-soc="accept-friend" data-id="${escapeHtml(f.playerId)}">ACCEPT</button>
       <button type="button" data-soc="decline-friend" data-id="${escapeHtml(f.playerId)}">DECLINE</button>`)).join("");
    const out = outgoing.map((f) => row(f, `<span>Pending</span>`)).join("");
    const block = blocked.length
      ? `<h4>BLOCKED</h4>${blocked.map((id) => `<div class="social-row"><span>${escapeHtml(id)}</span><button type="button" data-soc="unblock" data-id="${escapeHtml(id)}">Unblock</button></div>`).join("")}`
      : "";
    const tabs = `<div class="social-tabs">
      <button type="button" data-ftab="ONLINE" class="${tab === "ONLINE" ? "on" : ""}">ONLINE</button>
      <button type="button" data-ftab="AWAY" class="${tab === "AWAY" ? "on" : ""}">AWAY</button>
      <button type="button" data-ftab="OFFLINE" class="${tab === "OFFLINE" ? "on" : ""}">OFFLINE</button>
      <button type="button" data-ftab="REQUESTS" class="${tab === "REQUESTS" ? "on" : ""}">REQUESTS</button>
    </div>`;
    if (tab === "BLOCKED") {
      return `<h4>BLOCKED</h4>${block || "<p>Tidak ada pemain diblokir.</p>"}`;
    }
    if (tab === "REQUESTS") {
      return `<h4>FRIENDS</h4>${tabs}${pend || "<p>Tidak ada permintaan.</p>"}${out ? `<h4>OUTGOING</h4>${out}` : ""}`;
    }
    return `<h4>FRIENDS</h4>${tabs}${friendRows}`;
  }
}

function lastSeen(ms: number): string {
  if (!ms) return "—";
  const d = Date.now() - ms;
  if (d < 60_000) return "just now";
  if (d < 3_600_000) return `${Math.floor(d / 60_000)}m`;
  if (d < 86_400_000) return `${Math.floor(d / 3_600_000)}h`;
  return `${Math.floor(d / 86_400_000)}d`;
}

void lastSeen;

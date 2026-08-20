import { closeModal, showModal, toast } from "./chrome";
import {
  fetchFriends,
  fetchNotes,
  fetchPublicPlayer,
  friendBlock,
  friendRemove,
  friendRequest,
  friendRespond,
  markNotesRead,
  reportPlayer,
  searchPlayers,
  statusDot,
  type FriendCard,
  type FriendRequest,
  type UlarNote,
} from "../social/api";
import { storedSessionToken } from "../auth/session";
import { GameClient } from "../game/multiplayer/socket";
import { WS_EVENTS } from "../game/multiplayer/events";
import { AVATAR_FACE } from "../progress/types";
import { rankIcon } from "./rankIcons";

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export async function openFriendsModal(opts?: { client?: GameClient | null }): Promise<void> {
  const token = storedSessionToken();
  if (!token) {
    toast("Masuk untuk melihat teman.", "warning");
    return;
  }
  const first = await fetchFriends(token);
  if (!first.ok) {
    toast(first.error, "error");
    return;
  }
  let friends = first.data.friends || [];
  let incoming = first.data.incoming || [];
  let outgoing = first.data.outgoing || [];
  let q = "";
  let found: FriendCard[] = [];

  const paint = (): void => {
    const layer = showModal("friends", render());
    bind(layer);
  };

  const render = (): string => `
    <h2>Teman</h2>
    <form class="nt-row" data-search><input name="q" placeholder="Cari username" value="${esc(q)}" maxlength="16" /><button class="nt-btn" type="submit">Cari</button></form>
    ${found.length ? `<div class="pf-hist">${found.map((f) => row(f, "search")).join("")}</div>` : ""}
    ${incoming.length ? `<p class="nt-kicker">🔔 Permintaan</p>${incoming.map((r) => reqRow(r, "in")).join("")}` : ""}
    ${outgoing.length ? `<p class="nt-kicker">Terkirim</p>${outgoing.map((r) => reqRow(r, "out")).join("")}` : ""}
    <div class="pf-hist">${friends.map((f) => row(f, "list")).join("") || "<p class='nt-hint'>Belum ada teman.</p>"}</div>
    <button type="button" class="nt-btn nt-btn-ghost" data-close>Tutup</button>`;

  const reqRow = (r: FriendRequest, kind: "in" | "out"): string => `
    <article class="nt-card">
      <p>${AVATAR_FACE[r.avatar || ""] || "👤"} <strong>${esc(r.username || r.senderId)}</strong></p>
      ${kind === "in"
        ? `<button class="nt-btn nt-btn-primary" data-acc="${esc(r.id)}">ACCEPT</button> <button class="nt-btn" data-rej="${esc(r.id)}">REJECT</button>`
        : `<button class="nt-btn" data-can="${esc(r.id)}">CANCEL</button>`}
    </article>`;

  const row = (f: FriendCard, kind: string): string => `
    <article class="nt-card friend-row">
      <span>${statusDot(f.status)} ${AVATAR_FACE[f.avatar || ""] || "👤"} ${rankIcon((f.rankLabel || "BRONZE").split(" ")[0], 18)} <strong>${esc(f.username || f.userId)}</strong></span>
      <p class="nt-hint">Lv ${f.level || 1} · ${esc(f.rankLabel || "BRONZE III")}</p>
      <div class="nt-row">
        ${kind === "search" && !f.friends ? `<button class="nt-btn nt-btn-primary" data-add="${esc(f.userId)}">ADD FRIEND</button>` : ""}
        ${kind === "list" ? `<button class="nt-btn" data-view="${esc(f.userId)}">Profil</button>` : ""}
        ${kind === "list" && f.status === "ONLINE" ? `<button class="nt-btn nt-btn-primary" data-inv="${esc(f.userId)}">INVITE TO PLAY</button>` : ""}
        ${kind === "list" && f.status === "IN_GAME" ? `<span class="nt-hint">Sedang bermain.</span>` : ""}
        ${kind === "list" ? `<button class="nt-btn" data-blk="${esc(f.userId)}">Block</button> <button class="nt-btn nt-btn-ghost" data-rm="${esc(f.userId)}">Hapus</button>` : ""}
      </div>
    </article>`;

  const bind = (layer: HTMLElement): void => {
    layer.querySelector("[data-close]")?.addEventListener("click", () => closeModal("friends"));
    layer.querySelector<HTMLFormElement>("[data-search]")?.addEventListener("submit", (e) => {
      e.preventDefault();
      q = String(new FormData(e.target as HTMLFormElement).get("q") || "");
      void searchPlayers(token, q).then((out) => {
        if (!out.ok) toast(out.error, "error");
        else found = out.data.items || [];
        paint();
      });
    });
    layer.querySelectorAll<HTMLButtonElement>("[data-add]").forEach((b) => b.addEventListener("click", () => void onAdd(b.dataset.add || "")));
    layer.querySelectorAll<HTMLButtonElement>("[data-acc]").forEach((b) => b.addEventListener("click", () => void onResp(b.dataset.acc || "", "accept")));
    layer.querySelectorAll<HTMLButtonElement>("[data-rej]").forEach((b) => b.addEventListener("click", () => void onResp(b.dataset.rej || "", "reject")));
    layer.querySelectorAll<HTMLButtonElement>("[data-can]").forEach((b) => b.addEventListener("click", () => void onResp(b.dataset.can || "", "cancel")));
    layer.querySelectorAll<HTMLButtonElement>("[data-inv]").forEach((b) =>
      b.addEventListener("click", () => {
        if (!opts?.client) {
          toast("Buka PLAY ONLINE untuk mengundang.", "warning");
          return;
        }
        opts.client.send(WS_EVENTS.GAME_INVITE, { userId: b.dataset.inv });
        toast("Undangan dikirim", "success");
      }),
    );
    layer.querySelectorAll<HTMLButtonElement>("[data-view]").forEach((b) => b.addEventListener("click", () => void onView(b.dataset.view || "")));
    layer.querySelectorAll<HTMLButtonElement>("[data-blk]").forEach((b) =>
      b.addEventListener("click", () => void friendBlock(token, b.dataset.blk || "").then(() => reload())),
    );
    layer.querySelectorAll<HTMLButtonElement>("[data-rm]").forEach((b) =>
      b.addEventListener("click", () => void friendRemove(token, b.dataset.rm || "").then(() => reload())),
    );
  };

  const reload = async (): Promise<void> => {
    const out = await fetchFriends(token);
    if (out.ok) {
      friends = out.data.friends || [];
      incoming = out.data.incoming || [];
      outgoing = out.data.outgoing || [];
    }
    paint();
  };

  const onAdd = async (id: string): Promise<void> => {
    const out = await friendRequest(token, id);
    if (!out.ok) toast(out.error, "error");
    else toast("Permintaan terkirim", "success");
    await reload();
  };

  const onResp = async (id: string, action: "accept" | "reject" | "cancel"): Promise<void> => {
    const out = await friendRespond(token, id, action);
    if (!out.ok) toast(out.error, "error");
    await reload();
  };

  const onView = async (id: string): Promise<void> => {
    const out = await fetchPublicPlayer(token, id);
    if (!out.ok) {
      toast(out.error, "error");
      return;
    }
    const p = out.data;
    const layer = showModal(
      "friend-profile",
      `<h2>${esc(p.username)}</h2>
      <p>${AVATAR_FACE[p.avatar] || "👤"} Level ${p.level} · ${esc(p.rankLabel || "")}</p>
      <p>Win Rate ${Number(p.winRate || 0).toFixed(0)}% · Akurasi ${Number(p.accuracy || 0).toFixed(1)}%</p>
      <p>${(p.achievements || []).filter((a) => a.unlocked).map((a) => a.name).join(" · ") || "Belum ada achievement"}</p>
      <label>Laporkan
        <select data-reason>
          <option value="Spam">Spam</option>
          <option value="Harassment">Harassment</option>
          <option value="Inappropriate Name">Inappropriate Name</option>
          <option value="Cheating">Cheating</option>
          <option value="Other">Other</option>
        </select>
      </label>
      <button class="nt-btn" data-report>REPORT PLAYER</button>
      <button class="nt-btn nt-btn-ghost" data-close>Tutup</button>`,
    );
    layer.querySelector("[data-close]")?.addEventListener("click", () => closeModal("friend-profile"));
    layer.querySelector("[data-report]")?.addEventListener("click", () => {
      const reason = (layer.querySelector("[data-reason]") as HTMLSelectElement | null)?.value || "Other";
      void reportPlayer(token, id, reason).then((r) => {
        if (!r.ok) toast(r.error, "error");
        else toast("Laporan terkirim", "success");
      });
    });
  };

  paint();
}

export async function openNotificationsModal(): Promise<number> {
  const token = storedSessionToken();
  if (!token) return 0;
  const out = await fetchNotes(token);
  if (!out.ok) {
    toast(out.error, "error");
    return 0;
  }
  const items: UlarNote[] = out.data.items || [];
  const unread = out.data.unread || 0;
  const layer = showModal(
    "notes",
    `<h2>🔔 Notifikasi ${unread ? `<span class="note-badge">${unread}</span>` : ""}</h2>
    <div class="pf-hist">${items
      .map(
        (n) =>
          `<article class="nt-card ${n.read ? "" : "is-unread"}"><p class="nt-kicker">${esc(n.type)}</p><strong>${esc(n.title)}</strong><p>${esc(n.body)}</p></article>`,
      )
      .join("") || "<p class='nt-hint'>Tidak ada notifikasi.</p>"}</div>
    <button class="nt-btn nt-btn-ghost" data-close>Tutup</button>`,
  );
  layer.querySelector("[data-close]")?.addEventListener("click", () => closeModal("notes"));
  await markNotesRead(token);
  return 0;
}

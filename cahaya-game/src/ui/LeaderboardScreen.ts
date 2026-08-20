import { closeModal, showModal, toast } from "./chrome";
import { fetchLeaderboard, fetchSeason } from "../social/api";
import { storedSessionToken } from "../auth/session";
import { AVATAR_FACE } from "../progress/types";
import { rankIcon } from "./rankIcons";

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export async function openLeaderboardModal(): Promise<void> {
  const token = storedSessionToken();
  if (!token) {
    toast("Masuk untuk melihat papan peringkat.", "warning");
    return;
  }
  let kind = "rank";
  let scope = "global";
  let page = 0;
  let my = 0;
  let total = 0;
  let season = "Season 1";
  const s = await fetchSeason(token);
  if (s.ok) season = s.data.name || season;

  const load = async (): Promise<void> => {
    const out = await fetchLeaderboard(token, kind, page, scope);
    if (!out.ok) {
      toast(out.error, "error");
      return;
    }
    my = out.data.myRank || 0;
    total = out.data.total || 0;
    const rows = out.data.items || [];
    const last = (page + 1) * 20 >= total;
    const layer = showModal(
      "board",
      `<h2>Leaderboard</h2>
      <p class="nt-kicker">${esc(season)}</p>
      <p class="my-rank">My Rank ${my ? "#" + my : "—"}</p>
      <div class="nt-row wrap">
        ${["rank", "wins", "xp", "accuracy", "coins"].map((k) => `<button class="nt-btn ${kind === k ? "nt-btn-primary" : ""}" data-k="${k}">${k.toUpperCase()}</button>`).join("")}
      </div>
      <div class="nt-row">
        <button class="nt-btn ${scope === "global" ? "nt-btn-primary" : ""}" data-s="global">GLOBAL</button>
        <button class="nt-btn ${scope === "friends" ? "nt-btn-primary" : ""}" data-s="friends">FRIENDS</button>
      </div>
      <div class="pf-hist lb-list">${rows
        .map((r, i) => {
          const n = page * 20 + i + 1;
          const medal = n === 1 ? "🥇" : n === 2 ? "🥈" : n === 3 ? "🥉" : `#${n}`;
          return `<article class="nt-card lb-row lb-${n}"><strong>${medal} ${rankIcon(r.tier, 22)} ${AVATAR_FACE[r.avatar] || ""} ${esc(r.username)}</strong>
            <p>${esc(r.rankLabel)} · ${r.rr} RR · ${r.wins} W · Lv ${r.level}</p></article>`;
        })
        .join("") || "<p class='nt-hint'>Belum ada data.</p>"}</div>
      <div class="nt-row"><button class="nt-btn" data-prev ${page === 0 ? "disabled" : ""}>Prev</button><button class="nt-btn" data-next ${last ? "disabled" : ""}>Next</button></div>
      <button class="nt-btn nt-btn-ghost" data-close>Tutup</button>`,
    );
    layer.querySelector("[data-close]")?.addEventListener("click", () => closeModal("board"));
    layer.querySelectorAll<HTMLButtonElement>("[data-k]").forEach((b) =>
      b.addEventListener("click", () => {
        kind = b.dataset.k || "rank";
        page = 0;
        void load();
      }),
    );
    layer.querySelectorAll<HTMLButtonElement>("[data-s]").forEach((b) =>
      b.addEventListener("click", () => {
        scope = b.dataset.s || "global";
        page = 0;
        void load();
      }),
    );
    layer.querySelector("[data-prev]")?.addEventListener("click", () => {
      page = Math.max(0, page - 1);
      void load();
    });
    layer.querySelector("[data-next]")?.addEventListener("click", () => {
      if (!last) page += 1;
      void load();
    });
  };
  void load();
}

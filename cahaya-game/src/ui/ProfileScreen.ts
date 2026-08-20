import { closeModal, showModal, toast } from "./chrome";
import { playSfx } from "../audio/manager";
import { claimDaily, fetchDaily, fetchHistory, fetchProfile, updateProfile } from "../progress/api";
import { CurrencyService } from "../progress/currency";
import { applyRewardFx } from "../progress/fx";
import { AVATAR_FACE, fmtPct, xpBarWidth, type DailyStatus, type MatchHistoryItem, type PlayerProfileView } from "../progress/types";
import { fetchRankHistory, type RankHistoryItem } from "../social/api";
import { nextRankLabel, rankIcon, rrBarPct, rrToNext } from "./rankIcons";

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

function titleName(p: PlayerProfileView): string {
  return p.titles?.find((t) => t.id === p.title)?.name || "Pemula";
}

function bar(p: PlayerProfileView): string {
  const span = (p.xpIntoLevel || 0) + (p.xpToNext || 0);
  const cap = span || p.xpNextAt || 0;
  return `<div class="xp-bar" aria-label="XP"><i style="width:${xpBarWidth(p)}%"></i></div>
    <p class="nt-hint">${(p.xpIntoLevel || 0).toLocaleString("id-ID")} / ${cap.toLocaleString("id-ID")} XP · berikutnya level ${(p.level || 1) + (p.xpToNext ? 1 : 0)}</p>`;
}

export async function openProfileModal(token: string): Promise<void> {
  const first = await fetchProfile(token);
  if (!first.ok) {
    toast(first.error, "error");
    return;
  }
  let profile = first.data;
  let daily: DailyStatus | null = null;
  let history: MatchHistoryItem[] = [];
  let ranks: RankHistoryItem[] = [];
  let page = 0;
  let tab: "overview" | "stats" | "achievements" | "history" = "overview";
  const d = await fetchDaily(token);
  if (d.ok) daily = d.data;

  const paint = (): void => {
    const layer = showModal("profile", render());
    bind(layer);
  };

  const render = (): string => `
    <div class="pf-shell">
      <aside class="pf-side">
        <p class="pf-avatar">${AVATAR_FACE[profile.avatar] || "👤"}</p>
        <h2>${esc(profile.username || "Pemain")}</h2>
        <p class="nt-kicker">${esc(titleName(profile))}</p>
        <p class="pf-lvl">LEVEL ${profile.level || 1}</p>
        ${bar(profile)}
        <p class="pf-coins">${CurrencyService.display(profile.coins)}</p>
        <p class="pf-trophies">🏆 ${Number(profile.trophies || profile.wins || 0)} Piala</p>
        <div class="rank-card">
          ${rankIcon(profile.rankTier, 40)}
          <p class="nt-kicker">🏆 ${esc(profile.rankLabel || "BRONZE III")}</p>
          <p>${profile.rankRr || 0} RR</p>
          <div class="xp-bar rr-bar" aria-label="RR"><i style="width:${rrBarPct(profile.rankRr, profile.rankTier)}%"></i></div>
          <p class="nt-hint">Next: ${esc(nextRankLabel(profile.rankTier, profile.rankDivision))} · ${rrToNext(profile.rankRr, profile.rankTier)} RR</p>
          <p class="nt-hint">${esc(profile.seasonName || "Season 1")}${profile.myRank ? " · #" + profile.myRank : ""}</p>
        </div>
        ${profile.currentWinStreak ? `<p class="chip chip-ok">🔥 ${profile.currentWinStreak} WIN STREAK</p>` : ""}
        <p class="nt-hint">Best: 🔥 ${profile.bestWinStreak || 0}</p>
        <nav class="pf-tabs">
          <button type="button" class="${tab === "overview" ? "nt-btn-primary" : ""} nt-btn" data-tab="overview">OVERVIEW</button>
          <button type="button" class="${tab === "stats" ? "nt-btn-primary" : ""} nt-btn" data-tab="stats">STATISTICS</button>
          <button type="button" class="${tab === "achievements" ? "nt-btn-primary" : ""} nt-btn" data-tab="achievements">ACHIEVEMENTS</button>
          <button type="button" class="${tab === "history" ? "nt-btn-primary" : ""} nt-btn" data-tab="history">MATCH HISTORY</button>
        </nav>
      </aside>
      <section class="pf-main">${main()}</section>
    </div>
    <button type="button" class="nt-btn nt-btn-ghost" data-close>Tutup</button>`;

  const main = (): string => {
    if (tab === "stats") return stats();
    if (tab === "achievements") return achs();
    if (tab === "history") return hist();
    return overview();
  };

  const overview = (): string => {
    const day = daily?.day || 1;
    const coins = daily?.coins || 50;
    const claimed = Boolean(daily?.claimed);
    return `<div class="pf-grid">
      <article class="nt-card daily-chest ${claimed ? "is-claimed" : ""}">
        <p class="nt-kicker">🎁 DAILY REWARD</p>
        <p>DAY ${day}</p>
        <p class="pf-coins">+${coins} COINS</p>
        ${claimed ? `<p class="chip chip-ok">CLAIMED ✓</p>` : `<button type="button" class="nt-btn nt-btn-primary" data-claim>CLAIM</button>`}
      </article>
      <article class="nt-card">
        <p class="nt-kicker">Identitas</p>
        <label>Username <input data-f="username" maxlength="16" minlength="3" value="${esc(profile.username || "")}" /></label>
        <label>Avatar
          <select data-f="avatar">${(profile.avatars || []).map((id) => `<option value="${id}" ${id === profile.avatar ? "selected" : ""}>${AVATAR_FACE[id] || ""} ${id}</option>`).join("")}</select>
        </label>
        <label>Gelar
          <select data-f="title">${(profile.titles || []).map((t) => `<option value="${t.id}" ${t.id === profile.title ? "selected" : ""} ${t.unlocked ? "" : "disabled"}>${esc(t.name)}${t.unlocked ? "" : " (terkunci)"}</option>`).join("")}</select>
        </label>
        <button type="button" class="nt-btn nt-btn-primary" data-save>Simpan</button>
      </article>
    </div>`;
  };

  const stats = (): string => {
    const sub = profile.subjectAccuracy || {};
    return `<div class="pf-stats">
      <p>Matches: <strong>${profile.totalMatches || 0}</strong></p>
      <p>Wins: <strong>${profile.wins || 0}</strong></p>
      <p>Piala: <strong>🏆 ${profile.trophies || profile.wins || 0}</strong></p>
      <p>Losses: <strong>${profile.losses || 0}</strong></p>
      <p>Win Rate: <strong>${fmtPct(profile.winRate || 0)}</strong></p>
      <p>Questions: <strong>${profile.totalQuestions || 0}</strong></p>
      <p>Correct: <strong>${profile.correctAnswers || 0}</strong></p>
      <p>Accuracy: <strong>${fmtPct(profile.accuracy || 0)}</strong></p>
      <p>Rank: <strong>${esc(profile.rankLabel || "BRONZE III")}</strong></p>
      <p>RR: <strong>${profile.rankRr || 0}</strong></p>
      <p>Ranked Wins: <strong>${profile.rankedWins || 0}</strong></p>
      <hr/>
      <p class="nt-kicker">Riwayat rank</p>
      ${ranks.length ? ranks.map((h) => `<p>${esc(h.tierBefore)} ${esc(h.divBefore || "")} ${h.rrChange >= 0 ? "+" : ""}${h.rrChange} RR → ${esc(h.tierAfter)} ${esc(h.divAfter || "")}</p>`).join("") : `<button type="button" class="nt-btn" data-ranks>Muat riwayat rank</button>`}
      <hr/>
      <p>PAI: ${fmtPct(sub.PAI?.accuracy || 0)}</p>
      <p>MATEMATIKA: ${fmtPct(sub.MATEMATIKA?.accuracy || 0)}</p>
      <p>BAHASA INGGRIS: ${fmtPct(sub.BAHASA_INGGRIS?.accuracy || 0)}</p>
      <p>BAHASA JAWA: ${fmtPct(sub.BAHASA_JAWA?.accuracy || 0)}</p>
    </div>`;
  };

  const achs = (): string => `<div class="pf-achs">${(profile.achievements || [])
    .map((a) => `<article class="ach-card ${a.unlocked ? "is-on" : "is-off"}"><span>${a.unlocked ? "🏆" : "⬛"}</span><h3>${esc(a.name)}</h3><p>${esc(a.description)}</p></article>`)
    .join("")}</div>`;

  const hist = (): string => {
    if (!history.length) return `<p class="nt-hint">Belum ada riwayat.</p><button type="button" class="nt-btn" data-more>Muat riwayat</button>`;
    return `<div class="pf-hist">${history
      .map((h) => {
        const dte = new Date(h.createdAt).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
        return `<article class="nt-card"><p>${dte} · ${esc(h.mode || "CASUAL")}</p><strong>${h.won ? "WIN" : "LOSS"}</strong> Rank #${h.rank}<p>+${h.xpEarned} XP · +${h.coinsEarned} Coins${h.rrChange ? " · RR " + (h.rrChange > 0 ? "+" : "") + h.rrChange : ""}</p></article>`;
      })
      .join("")}</div><button type="button" class="nt-btn" data-more>Muat lagi</button>`;
  };

  const bind = (layer: HTMLElement): void => {
    layer.querySelector("[data-close]")?.addEventListener("click", () => closeModal("profile"));
    layer.querySelectorAll<HTMLButtonElement>("[data-tab]").forEach((btn) => {
      btn.addEventListener("click", () => {
        tab = btn.dataset.tab as typeof tab;
        if (tab === "history" && history.length === 0) void loadHistory();
        else if (tab === "stats" && ranks.length === 0) void loadRanks();
        else paint();
      });
    });
    layer.querySelector("[data-claim]")?.addEventListener("click", () => void onClaim(layer));
    layer.querySelector("[data-save]")?.addEventListener("click", () => void onSave(layer));
    layer.querySelector("[data-more]")?.addEventListener("click", () => void loadHistory());
    layer.querySelector("[data-ranks]")?.addEventListener("click", () => void loadRanks());
  };

  const loadRanks = async (): Promise<void> => {
    const out = await fetchRankHistory(token, 0);
    if (!out.ok) {
      toast(out.error, "error");
      return;
    }
    ranks = out.data.items || [];
    tab = "stats";
    paint();
  };

  const loadHistory = async (): Promise<void> => {
    const out = await fetchHistory(token, page);
    if (!out.ok) {
      toast(out.error, "error");
      return;
    }
    history = history.concat(out.data.items || []);
    page += 1;
    tab = "history";
    paint();
  };

  const onClaim = async (layer: HTMLElement): Promise<void> => {
    playSfx("coin");
    layer.querySelector(".daily-chest")?.classList.add("is-open");
    const out = await claimDaily(token);
    if (!out.ok) {
      toast(out.error, "error");
      return;
    }
    profile = out.data.profile || profile;
    daily = { claimed: true, day: daily?.day || 1, coins: out.data.reward?.coins || 0, streak: profile.dailyStreak, date: daily?.date || "" };
    applyRewardFx(layer, out.data.reward, profile.achievements);
    toast("Hadiah harian diambil", "success");
    paint();
  };

  const onSave = async (layer: HTMLElement): Promise<void> => {
    const username = (layer.querySelector('[data-f=username]') as HTMLInputElement | null)?.value || "";
    const avatar = (layer.querySelector('[data-f=avatar]') as HTMLSelectElement | null)?.value || "";
    const title = (layer.querySelector('[data-f=title]') as HTMLSelectElement | null)?.value || "";
    const out = await updateProfile(token, { username, avatar, title });
    if (!out.ok) {
      toast(out.error, "error");
      return;
    }
    profile = out.data;
    toast("Profil disimpan", "success");
    paint();
  };

  paint();
}

import { particleScale } from "./prefs";
import { playSfx } from "../audio/manager";

type ToastKind = "success" | "error" | "info" | "warning";

let host: HTMLElement | null = null;

function ensure(): HTMLElement {
  if (host && document.body.contains(host)) return host;
  host = document.createElement("div");
  host.className = "toast-host";
  host.setAttribute("aria-live", "polite");
  document.body.appendChild(host);
  return host;
}

export function toast(message: string, kind: ToastKind = "info"): void {
  const el = document.createElement("div");
  el.className = `toast toast-${kind}`;
  el.textContent = message;
  ensure().appendChild(el);
  window.setTimeout(() => {
    el.classList.add("is-out");
    window.setTimeout(() => el.remove(), 280);
  }, 2400);
}

export function setConnectionBanner(status: "offline" | "connecting" | "online" | "hidden"): void {
  let bar = document.querySelector<HTMLElement>(".conn-banner");
  if (!bar) {
    bar = document.createElement("div");
    bar.className = "conn-banner";
    bar.setAttribute("role", "status");
    document.body.appendChild(bar);
  }
  if (status === "hidden" || status === "online") {
    if (status === "online" && bar.dataset.wasOffline === "1") {
      bar.dataset.wasOffline = "0";
      bar.className = "conn-banner is-ok";
      bar.textContent = "🟢 Terhubung kembali.";
      window.setTimeout(() => {
        bar!.classList.add("is-hide");
      }, 1800);
      return;
    }
    bar.classList.add("is-hide");
    return;
  }
  bar.dataset.wasOffline = "1";
  bar.classList.remove("is-hide");
  bar.className = "conn-banner is-bad";
  bar.textContent =
    status === "connecting"
      ? "🟡 RECONNECTING…"
      : "🔴 Koneksi terputus. Sedang mencoba menghubungkan kembali…";
}

export function showModal(id: string, html: string, opts?: { noBackdropClose?: boolean }): HTMLElement {
  closeModal(id);
  const wrap = document.createElement("div");
  wrap.className = "ui-modal-layer";
  wrap.dataset.modal = id;
  wrap.innerHTML = `<div class="ui-modal nt-card" role="dialog" aria-modal="true">${html}</div>`;
  if (!opts?.noBackdropClose) {
    wrap.addEventListener("click", (e) => {
      if (e.target === wrap) closeModal(id);
    });
  }
  document.body.appendChild(wrap);
  const focus = wrap.querySelector<HTMLElement>("button, [href], input");
  focus?.focus();
  return wrap;
}

export function closeModal(id?: string): void {
  document.querySelectorAll(".ui-modal-layer").forEach((el) => {
    if (!id || (el as HTMLElement).dataset.modal === id) el.remove();
  });
}

export function spawnFloat(parent: HTMLElement, text: string, kind: "good" | "bad"): void {
  const el = document.createElement("div");
  el.className = `float-fx float-${kind}`;
  el.textContent = text;
  parent.appendChild(el);
  window.setTimeout(() => el.remove(), 900);
}

export function sparkle(parent: HTMLElement): void {
  const scale = particleScale();
  if (scale <= 0) return;
  const n = Math.max(2, Math.round(6 * scale));
  for (let i = 0; i < n; i++) {
    const d = document.createElement("span");
    d.className = "spark";
    d.style.setProperty("--dx", `${(i - 2.5) * 10}px`);
    d.style.setProperty("--delay", `${i * 40}ms`);
    parent.appendChild(d);
    window.setTimeout(() => d.remove(), 700);
  }
}

export function confetti(parent: HTMLElement, lavish = false): void {
  const scale = particleScale();
  if (scale <= 0) return;
  const count = Math.max(4, Math.round((lavish ? 36 : 14) * scale));
  const colors = ["#ffd54f", "#3ecf8e", "#6ec8e6", "#ff8a80", "#ce93d8", "#fff59d", "#80deea"];
  for (let i = 0; i < count; i++) {
    const d = document.createElement("span");
    d.className = lavish ? "confetti confetti-lavish" : "confetti";
    d.style.left = `${4 + (i * 92) / count}%`;
    d.style.setProperty("--c", colors[i % colors.length]!);
    d.style.setProperty("--delay", `${i * (lavish ? 18 : 30)}ms`);
    d.style.setProperty("--spin", `${80 + (i % 7) * 40}deg`);
    d.style.setProperty("--fall", `${60 + (i % 5) * 18}px`);
    parent.appendChild(d);
    window.setTimeout(() => d.remove(), lavish ? 2200 : 1200);
  }
}

export async function shareText(title: string, text: string): Promise<void> {
  const payload = { title, text };
  try {
    if (navigator.share) {
      await navigator.share(payload);
      return;
    }
  } catch {
    /* fall through */
  }
  try {
    await navigator.clipboard.writeText(text);
    toast("Hasil disalin ke clipboard.", "success");
  } catch {
    toast("Bagikan tidak tersedia di perangkat ini.", "info");
  }
}

export type ChampionOpts = {
  username: string;
  isSelf?: boolean;
  trophiesAwarded?: number;
  trophiesTotal?: number;
  xp?: number;
  coins?: number;
  subtitle?: string;
  onClose?: () => void;
};

/** Ucapan juara meriah + piala. */
export function showChampionCelebration(opts: ChampionOpts): void {
  playSfx("win");
  playSfx("finish");
  const awarded = opts.trophiesAwarded ?? 1;
  const total = opts.trophiesTotal ?? awarded;
  const title = opts.isSelf ? "SELAMAT, KAMU JUARA!" : "SELAMAT JUARA!";
  const layer = showModal(
    "champion",
    `<div class="champion-blast" data-champ>
      <div class="champion-burst" aria-hidden="true"></div>
      <p class="champion-kicker">🏆 JUARA 1</p>
      <div class="champion-cup" aria-hidden="true">🏆</div>
      <h2 class="champion-title">${escapeHtml(title)}</h2>
      <p class="champion-name">${escapeHtml(opts.username)}</p>
      <p class="champion-msg">${opts.isSelf ? "Hebat! Kamu menuntaskan perjalanan ke kotak 100." : "Perayaan untuk sang juara yang mencapai finish!"}</p>
      <div class="champion-rewards">
        <span class="champion-pill">🏆 +${awarded} Piala</span>
        ${opts.xp ? `<span class="champion-pill">✨ +${opts.xp} XP</span>` : ""}
        ${opts.coins ? `<span class="champion-pill">🪙 +${opts.coins}</span>` : ""}
      </div>
      <p class="champion-total">Total piala: <strong>${total}</strong></p>
      ${opts.subtitle ? `<p class="nt-hint">${escapeHtml(opts.subtitle)}</p>` : ""}
      <button type="button" class="nt-btn nt-btn-primary" data-close>MENU UTAMA</button>
    </div>`,
  );
  const host = layer.querySelector("[data-champ]") as HTMLElement | null;
  if (host) {
    confetti(host, true);
  }
  layer.querySelector("[data-close]")?.addEventListener("click", () => {
    closeModal("champion");
    opts.onClose?.();
  });
}

function escapeHtml(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export type MatchShare = {
  username: string;
  rank: number;
  xp: number;
  coins: number;
  trophies?: number;
  trophyTotal?: number;
  rrDelta?: number;
  mode: string;
};

export function showMatchResultModal(data: MatchShare): void {
  if (data.rank === 1) {
    showChampionCelebration({
      username: data.username,
      isSelf: true,
      trophiesAwarded: data.trophies ?? 1,
      trophiesTotal: data.trophyTotal ?? data.trophies ?? 1,
      xp: data.xp,
      coins: data.coins,
      subtitle: data.rrDelta !== undefined ? `RR ${data.rrDelta >= 0 ? "+" : ""}${data.rrDelta} · ${data.mode}` : data.mode,
    });
    return;
  }
  const rr = data.rrDelta !== undefined ? `<p>RR: ${data.rrDelta >= 0 ? "+" : ""}${data.rrDelta}</p>` : "";
  const layer = showModal(
    "match-result",
    `<div class="match-result">
      <p class="nt-kicker">HASIL PERTANDINGAN</p>
      <h2>Peringkat #${data.rank}</h2>
      <p>${escapeHtml(data.username)}</p>
      <p>+${data.xp} XP · +${data.coins} Coins</p>
      ${rr}
      <div class="share-row">
        <button type="button" class="nt-btn nt-btn-primary" data-share>SHARE RESULT</button>
        <button type="button" class="nt-btn" data-close>Tutup</button>
      </div>
    </div>`,
  );
  const text = `Aku finis #${data.rank} di Ular Tangga Nusantara (${data.mode})! +${data.xp} XP, +${data.coins} coins.`;
  layer.querySelector("[data-share]")?.addEventListener("click", () => {
    void shareText("Ular Tangga Nusantara", text);
  });
  layer.querySelector("[data-close]")?.addEventListener("click", () => closeModal("match-result"));
}

import "./admin.css";
import { apiLogin, clearSession, saveSession, storedSessionToken } from "../auth/session";
import { toast } from "../ui/chrome";

const API = "/cahaya/api/admin";

type Perms = Record<string, boolean>;
type Me = { role: string; perms: Perms; adminId: string; username: string };

type Page =
  | "dashboard" | "players" | "questions" | "matches" | "reports"
  | "achievements" | "rewards" | "ranks" | "seasons" | "config" | "audit-logs";

const MENUS: { id: Page; label: string; perm: string }[] = [
  { id: "dashboard", label: "Dashboard", perm: "PLAYER_VIEW" },
  { id: "players", label: "Players", perm: "PLAYER_VIEW" },
  { id: "questions", label: "Questions", perm: "QUESTION_VIEW" },
  { id: "matches", label: "Matches", perm: "MATCH_VIEW" },
  { id: "reports", label: "Reports", perm: "REPORT_VIEW" },
  { id: "achievements", label: "Achievements", perm: "ACHIEVEMENT_EDIT" },
  { id: "rewards", label: "Rewards", perm: "REWARD_EDIT" },
  { id: "ranks", label: "Ranks", perm: "RANK_EDIT" },
  { id: "seasons", label: "Seasons", perm: "SEASON_MANAGE" },
  { id: "config", label: "Configuration", perm: "CONFIG_VIEW" },
  { id: "audit-logs", label: "Audit Logs", perm: "AUDIT_VIEW" },
];

function prefix(): string {
  const p = location.pathname;
  if (p.startsWith("/cahaya/admin")) return "/cahaya/admin";
  if (p.startsWith("/admin")) return "/admin";
  return "/admin";
}

function currentPage(): { page: Page; id: string } {
  const hash = location.hash.replace(/^#\/?/, "");
  const raw = hash || location.pathname.replace(prefix(), "").replace(/^\//, "") || "dashboard";
  const [page, id = ""] = raw.split("/");
  const known = MENUS.some((m) => m.id === page);
  return { page: (known ? page : "dashboard") as Page, id };
}

function go(page: string): void {
  const pref = location.pathname.startsWith("/cahaya/") && !location.pathname.startsWith("/cahaya/admin")
    ? ""
    : prefix();
  if (!pref || location.search.includes("panel=admin")) {
    location.hash = page;
    return;
  }
  history.pushState({}, "", `${pref}/${page}`.replace(/\/$/, ""));
  window.dispatchEvent(new Event("popstate"));
}

async function api<T>(path: string, init: RequestInit = {}): Promise<{ ok: boolean; status: number; data: T; error: string }> {
  const headers: Record<string, string> = { ...(init.headers as Record<string, string> || {}) };
  const tok = storedSessionToken();
  if (tok) headers.Authorization = `Bearer ${tok}`;
  if (init.body && !(init.body instanceof FormData) && !headers["Content-Type"]) headers["Content-Type"] = "application/json";
  try {
    const res = await fetch(API + path, { ...init, headers });
    const text = await res.text();
    let data: T & { error?: string } = {} as T & { error?: string };
    try { data = JSON.parse(text) as T & { error?: string }; } catch { /* csv */ }
    if (res.status === 401) {
      clearSession();
      return { ok: false, status: 401, data, error: data.error || "sesi berakhir" };
    }
    if (!res.ok) return { ok: false, status: res.status, data, error: data.error || "gagal" };
    return { ok: true, status: res.status, data, error: "" };
  } catch {
    return { ok: false, status: 0, data: {} as T, error: "tidak dapat terhubung" };
  }
}

function esc(s: unknown): string {
  return String(s ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] || c));
}

function when(ms: unknown): string {
  const n = Number(ms);
  if (!n) return "—";
  return new Date(n).toLocaleString("id-ID");
}

function pill(st: string): string {
  const u = (st || "").toUpperCase();
  const k = u === "HEALTHY" || u === "ONLINE" || u === "ACTIVE" || u === "RESOLVED" ? "ok"
    : u === "DEGRADED" || u === "UNDER_REVIEW" || u === "INACTIVE" ? "warn" : u === "DOWN" || u === "BANNED" || u === "OPEN" ? "bad" : "";
  return `<span class="pill ${k}">${esc(st)}</span>`;
}

function spark(daily: Record<string, Record<string, number>>, key: string): string {
  const days = Object.keys(daily).sort();
  const vals = days.map((d) => daily[d]?.[key] || 0);
  if (!vals.length) return `<p class="adm-muted">Tidak ada data.</p>`;
  const max = Math.max(...vals, 1);
  const w = 320, h = 72;
  const pts = vals.map((v, i) => `${(i / Math.max(vals.length - 1, 1)) * w},${h - (v / max) * (h - 8) - 4}`).join(" ");
  return `<svg viewBox="0 0 ${w} ${h}" class="adm-chart" aria-hidden="true"><polyline fill="none" stroke="currentColor" stroke-width="2" points="${pts}"/></svg>`;
}

function confirmModal(title: string, body: string, phrase = ""): Promise<{ ok: boolean; reason: string; typed: string }> {
  return new Promise((resolve) => {
    const layer = document.createElement("div");
    layer.className = "adm-modal";
    layer.innerHTML = `<div class="box">
      <h3>${esc(title)}</h3>
      <p class="adm-muted">${esc(body)}</p>
      <label>Alasan</label>
      <textarea id="adm-reason"></textarea>
      ${phrase ? `<label>Ketik <b>${esc(phrase)}</b></label><input id="adm-phrase" autocomplete="off" />` : ""}
      <div class="adm-row" style="margin-top:12px">
        <button class="adm-btn danger" id="adm-yes">Konfirmasi</button>
        <button class="adm-btn ghost" id="adm-no">Batal</button>
      </div>
    </div>`;
    document.body.appendChild(layer);
    const done = (ok: boolean) => {
      const reason = (layer.querySelector("#adm-reason") as HTMLTextAreaElement).value.trim();
      const typed = phrase ? (layer.querySelector("#adm-phrase") as HTMLInputElement).value.trim() : "";
      layer.remove();
      resolve({ ok, reason, typed });
    };
    layer.querySelector("#adm-yes")?.addEventListener("click", () => {
      if (phrase && (layer.querySelector("#adm-phrase") as HTMLInputElement).value.trim().toUpperCase() !== phrase.toUpperCase()) {
        toast("Konfirmasi tidak cocok", "error");
        return;
      }
      const reason = (layer.querySelector("#adm-reason") as HTMLTextAreaElement).value.trim();
      if (!reason) { toast("Alasan wajib", "error"); return; }
      done(true);
    });
    layer.querySelector("#adm-no")?.addEventListener("click", () => done(false));
  });
}

export function mountAdmin(root: HTMLElement): void {
  let me: Me | null = null;
  let range = "7";
  let page = 0;
  let q = "";
  let status = "";
  let qSubject = "";
  let qDiff = "";
  let qActive = "";
  let timer = 0;

  const render = async (): Promise<void> => {
    const tok = storedSessionToken();
    if (!tok || !me) {
      root.innerHTML = loginHTML();
      bindLogin();
      return;
    }
    const cur = currentPage();
    root.innerHTML = shell(cur.page);
    bindNav();
    await draw(cur);
  };

  const loginHTML = (): string => `
    <div class="adm-login">
      <h1>Admin Panel</h1>
      <p>Ular Tangga Nusantara — live operations</p>
      <label>Username</label><input id="u" autocomplete="username" />
      <label>Password</label><input id="p" type="password" autocomplete="current-password" />
      <p class="adm-err" id="e"></p>
      <button class="adm-btn" id="go" style="margin-top:12px;width:100%">Masuk</button>
    </div>`;

  const bindLogin = (): void => {
    const run = async () => {
      const u = (root.querySelector("#u") as HTMLInputElement).value.trim();
      const p = (root.querySelector("#p") as HTMLInputElement).value;
      const e = root.querySelector("#e") as HTMLElement;
      const res = await apiLogin(u, p);
      if (!res.ok) { e.textContent = res.error; return; }
      saveSession(res.data);
      const meRes = await api<Me>("/me");
      if (!meRes.ok) {
        clearSession();
        e.textContent = meRes.status === 403 ? "Akun ini bukan admin." : meRes.error;
        return;
      }
      me = meRes.data;
      void render();
    };
    root.querySelector("#go")?.addEventListener("click", () => void run());
    root.querySelector("#p")?.addEventListener("keydown", (ev) => { if ((ev as KeyboardEvent).key === "Enter") void run(); });
  };

  const shell = (pageId: Page): string => `
    <div class="adm-shell">
      <aside class="adm-side">
        <div class="adm-brand">ULAR ADMIN</div>
        <nav class="adm-nav">${MENUS.filter((m) => me?.perms[m.perm]).map((m) =>
          `<a href="${prefix()}/${m.id}" data-go="${m.id}" class="${m.id === pageId ? "is-on" : ""}">${m.label}</a>`).join("")}
        </nav>
      </aside>
      <header class="adm-top">
        <b>${MENUS.find((m) => m.id === pageId)?.label || "Admin"}</b>
        <span class="adm-muted">${esc(me?.username)} · ${esc(me?.role)} <button class="adm-btn ghost" id="out">Keluar</button></span>
      </header>
      <main class="adm-main" id="main"><p class="adm-muted">Memuat…</p></main>
    </div>`;

  const bindNav = (): void => {
    root.querySelectorAll<HTMLAnchorElement>("[data-go]").forEach((a) => {
      a.addEventListener("click", (ev) => {
        ev.preventDefault();
        go(a.dataset.go || "dashboard");
      });
    });
    root.querySelector("#out")?.addEventListener("click", () => { clearSession(); me = null; void render(); });
  };

  const main = (): HTMLElement => root.querySelector("#main") as HTMLElement;

  const draw = async (cur: { page: Page; id: string }): Promise<void> => {
    const el = main();
    switch (cur.page) {
      case "dashboard": await dash(el); break;
      case "players": await players(el, cur.id); break;
      case "questions": await questions(el); break;
      case "matches": await matches(el, cur.id); break;
      case "reports": await reports(el); break;
      case "achievements": await achievements(el); break;
      case "rewards": await rewards(el); break;
      case "ranks": await ranks(el); break;
      case "seasons": await seasons(el); break;
      case "config": await config(el); break;
      case "audit-logs": await audit(el); break;
      default: el.innerHTML = "Tidak ditemukan";
    }
  };

  const dash = async (el: HTMLElement): Promise<void> => {
    const qs = `?range=${encodeURIComponent(range)}`;
    const [d, st] = await Promise.all([api<Record<string, unknown>>("/dashboard" + qs), api<Record<string, string>>("/status")]);
    if (!d.ok) { el.innerHTML = `<p class="adm-err">${esc(d.error)}</p>`; return; }
    const x = d.data;
    const s = st.data || {};
    const charts = (x.charts || {}) as { daily?: Record<string, Record<string, number>>; accuracy?: number; subjects?: Record<string, number> };
    el.innerHTML = `
      <div class="adm-toolbar">
        <select id="rng">${["today", "7", "30"].map((n) => `<option value="${n}" ${n === range ? "selected" : ""}>${n === "today" ? "Today" : n + " Days"}</option>`).join("")}</select>
      </div>
      <div class="adm-status">
        ${["api", "database", "websocket", "matchmaking", "questionService"].map((k) => `${k} ${pill(s[k] || "DOWN")}`).join(" · ")}
      </div>
      <div class="adm-cards">
        ${card("Players", x.totalPlayers)}
        ${card("Online Now", x.onlinePlayers)}
        ${card("Active Matches", x.activeMatches)}
        ${card("Matches Today", x.matchesToday)}
        ${card("Questions", x.questions)}
        ${card("Reports Pending", x.reportsPending)}
        ${card("New Users Today", x.newUsersToday)}
      </div>
      <div class="adm-charts">
        <div class="adm-card"><div class="k">Daily Active Players</div>${spark(charts.daily || {}, "dap")}</div>
        <div class="adm-card"><div class="k">Match Count</div>${spark(charts.daily || {}, "matches")}</div>
        <div class="adm-card"><div class="k">New Users</div>${spark(charts.daily || {}, "newUsers")}</div>
        <div class="adm-card"><div class="k">Question Accuracy</div><div class="v">${Number(charts.accuracy || 0).toFixed(1)}%</div></div>
        <div class="adm-card"><div class="k">Subject Distribution</div><p class="adm-muted">${Object.entries(charts.subjects || {}).map(([k, v]) => `${k}: ${v}`).join(" · ") || "—"}</p></div>
      </div>`;
    el.querySelector("#rng")?.addEventListener("change", (ev) => { range = (ev.target as HTMLSelectElement).value; void render(); });
  };

  const card = (k: string, v: unknown): string => `<div class="adm-card"><div class="k">${k}</div><div class="v">${esc(v ?? 0)}</div></div>`;

  const players = async (el: HTMLElement, id: string): Promise<void> => {
    if (id) {
      const r = await api<Record<string, unknown>>(`/players/${id}`);
      if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
      const acc = (r.data.account || {}) as Record<string, unknown>;
      const pf = (r.data.profile || {}) as Record<string, unknown>;
      el.innerHTML = `<button class="adm-btn ghost" data-back>← Players</button>
        <h2>${esc(acc.username)}</h2>
        <p class="adm-muted">${esc(acc.playerId)} · ${esc(pf.rankLabel || pf.rankTier)} · L${esc(pf.level)} · ${esc(pf.coins)} koin</p>
        <div class="adm-row">
          <button class="adm-btn warn" data-act="WARNING">Warn</button>
          <button class="adm-btn warn" data-act="CHAT_MUTE">Mute</button>
          <button class="adm-btn danger" data-act="TEMP_BAN">Temp Ban</button>
          <button class="adm-btn danger" data-act="PERMANENT_BAN">Permanent Ban</button>
          <button class="adm-btn ghost" data-unban>Unban</button>
        </div>
        <h3>Match history</h3><div class="adm-pre">${esc(JSON.stringify(r.data.history || [], null, 2))}</div>
        <h3>Rank / coins / XP / achievements / reports</h3>
        <div class="adm-pre">${esc(JSON.stringify({ rankHistory: r.data.rankHistory, coins: r.data.coins, xp: r.data.xp, achievements: r.data.achievements, reports: r.data.reports, sanctions: r.data.sanctions }, null, 2))}</div>`;
      el.querySelector("[data-back]")?.addEventListener("click", () => go("players"));
      el.querySelectorAll<HTMLButtonElement>("[data-act]").forEach((b) => b.addEventListener("click", async () => {
        const typ = b.dataset.act || "";
        const phrase = typ === "PERMANENT_BAN" ? "PERMANENT BAN" : "";
        const c = await confirmModal("Sanksi " + typ, "Tindakan ini tercatat di audit log.", phrase);
        if (!c.ok) return;
        const body: Record<string, unknown> = { type: typ, reason: c.reason, confirm: c.typed };
        if (typ === "TEMP_BAN") body.endAt = Date.now() + 24 * 3600 * 1000;
        const res = await api(`/players/${id}/sanction`, { method: "POST", body: JSON.stringify(body) });
        toast(res.ok ? "Sanksi diterapkan" : res.error, res.ok ? "success" : "error");
        void render();
      }));
      el.querySelector("[data-unban]")?.addEventListener("click", async () => {
        const res = await api(`/players/${id}/unban`, { method: "POST", body: "{}" });
        toast(res.ok ? "Unban" : res.error, res.ok ? "success" : "error");
        void render();
      });
      return;
    }
    const r = await api<{ items: Record<string, unknown>[]; total: number }>(`/players?q=${encodeURIComponent(q)}&status=${encodeURIComponent(status)}&page=${page}`);
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    el.innerHTML = `
      <div class="adm-toolbar">
        <input id="sq" placeholder="Cari username / user ID" value="${esc(q)}" />
        <select id="st"><option value="">Semua status</option><option value="ONLINE" ${status === "ONLINE" ? "selected" : ""}>ONLINE</option><option value="OFFLINE" ${status === "OFFLINE" ? "selected" : ""}>OFFLINE</option><option value="BANNED" ${status === "BANNED" ? "selected" : ""}>BANNED</option></select>
      </div>
      <div class="adm-table-wrap"><table class="adm"><thead><tr>
        <th>Avatar</th><th>Username</th><th>ID</th><th>Level</th><th>XP</th><th>Coins</th><th>Rank</th><th>RR</th><th>Matches</th><th>Wins</th><th>Status</th><th>Registered</th>
      </tr></thead><tbody>${(r.data.items || []).map((p) => `<tr data-id="${esc(p.userId)}">
        <td>${esc(p.avatar)}</td><td>${esc(p.username)}</td><td>${esc(p.userId)}</td><td>${esc(p.level)}</td><td>${esc(p.xp)}</td>
        <td>${esc(p.coins)}</td><td>${esc(p.rank)}</td><td>${esc(p.rr)}</td><td>${esc(p.matches)}</td><td>${esc(p.wins)}</td>
        <td>${pill(String(p.status))}</td><td>${when(p.createdAt)}</td></tr>`).join("")}</tbody></table></div>
      <div class="adm-pager"><button class="adm-btn ghost" id="prev">Prev</button><span>${page + 1} / ${Math.max(1, Math.ceil((r.data.total || 0) / 20))}</span><button class="adm-btn ghost" id="next">Next</button></div>`;
    const search = el.querySelector("#sq") as HTMLInputElement;
    search.addEventListener("input", () => {
      window.clearTimeout(timer);
      timer = window.setTimeout(() => { q = search.value; page = 0; void render(); }, 280);
    });
    el.querySelector("#st")?.addEventListener("change", (ev) => { status = (ev.target as HTMLSelectElement).value; page = 0; void render(); });
    el.querySelectorAll("tbody tr").forEach((tr) => tr.addEventListener("click", () => go("players/" + (tr as HTMLElement).dataset.id)));
    el.querySelector("#prev")?.addEventListener("click", () => { if (page > 0) { page--; void render(); } });
    el.querySelector("#next")?.addEventListener("click", () => { page++; void render(); });
  };

  const questions = async (el: HTMLElement): Promise<void> => {
    const r = await api<{ items: Record<string, unknown>[]; total: number; summary: Record<string, number> }>(`/questions?q=${encodeURIComponent(q)}&subject=${encodeURIComponent(qSubject)}&difficulty=${encodeURIComponent(qDiff)}&active=${encodeURIComponent(qActive)}&page=${page}`);
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    const s = r.data.summary || {};
    el.innerHTML = `
      <div class="adm-cards">${card("Total", s.total)}${card("PAI", s.pai)}${card("Matematika", s.matematika || s.math)}${card("Inggris", s.bahasaInggris || s.english)}${card("Jawa", s.bahasaJawa || s.jawa)}</div>
      <div class="adm-toolbar">
        <input id="sq" placeholder="Cari teks soal" value="${esc(q)}" />
        <select id="sub"><option value="">Subject</option>${["PAI","MATEMATIKA","BAHASA_INGGRIS","BAHASA_JAWA"].map((s) => `<option ${qSubject===s?"selected":""}>${s}</option>`).join("")}</select>
        <select id="dif"><option value="">Difficulty</option>${["EASY","MEDIUM","HARD"].map((s) => `<option ${qDiff===s?"selected":""}>${s}</option>`).join("")}</select>
        <select id="actv"><option value="">Semua</option><option value="1" ${qActive==="1"?"selected":""}>Active</option><option value="0" ${qActive==="0"?"selected":""}>Inactive</option></select>
        <button class="adm-btn" id="newq">Buat soal</button>
        <button class="adm-btn ghost" id="exp">Export CSV</button>
        <input type="file" id="imp" accept=".csv" hidden />
        <button class="adm-btn ghost" id="impi">Import CSV</button>
      </div>
      <div class="adm-table-wrap"><table class="adm"><thead><tr>
        <th>ID</th><th>Subject</th><th>Difficulty</th><th>Question</th><th>Answer</th><th>Status</th><th>Asked</th><th>Accuracy</th><th>Created</th><th></th>
      </tr></thead><tbody>${(r.data.items || []).map((it) => `<tr>
        <td>${esc(it.id)}</td><td>${esc(it.subject)}</td><td>${esc(it.difficulty)}</td><td>${esc(String(it.question).slice(0, 80))}</td>
        <td>${esc(it.correctAnswer)}</td><td>${pill(String(it.status))}${it.needsReview ? " NEEDS REVIEW" : ""}</td>
        <td>${esc(it.timesAsked)}</td><td>${Number(it.accuracy || 0).toFixed(0)}% ${esc(it.warning)}</td><td>${when(it.createdAt)}</td>
        <td><button class="adm-btn ghost" data-ed="${esc(it.id)}">Edit</button> <button class="adm-btn ghost" data-off="${esc(it.id)}">${it.active ? "Nonaktif" : "Aktifkan"}</button> <button class="adm-btn danger" data-del="${esc(it.id)}">Hapus</button></td>
      </tr>`).join("")}</tbody></table></div>
      <div class="adm-pager"><button class="adm-btn ghost" id="prev">Prev</button><span>hlm ${page + 1}</span><button class="adm-btn ghost" id="next">Next</button></div>
      <div id="qform"></div>`;
    const search = el.querySelector("#sq") as HTMLInputElement;
    search.addEventListener("input", () => {
      window.clearTimeout(timer);
      timer = window.setTimeout(() => { q = search.value; page = 0; void render(); }, 280);
    });
    el.querySelector("#sub")?.addEventListener("change", (ev) => { qSubject = (ev.target as HTMLSelectElement).value; page = 0; void render(); });
    el.querySelector("#dif")?.addEventListener("change", (ev) => { qDiff = (ev.target as HTMLSelectElement).value; page = 0; void render(); });
    el.querySelector("#actv")?.addEventListener("change", (ev) => { qActive = (ev.target as HTMLSelectElement).value; page = 0; void render(); });
    el.querySelector("#prev")?.addEventListener("click", () => { if (page > 0) { page--; void render(); } });
    el.querySelector("#next")?.addEventListener("click", () => { page++; void render(); });
    el.querySelector("#newq")?.addEventListener("click", () => qForm(el.querySelector("#qform") as HTMLElement, null));
    el.querySelector("#exp")?.addEventListener("click", async () => {
      const res = await fetch(API + "/questions/export", { headers: { Authorization: `Bearer ${storedSessionToken()}` } });
      const blob = await res.blob();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = "questions.csv";
      a.click();
    });
    el.querySelector("#impi")?.addEventListener("click", () => (el.querySelector("#imp") as HTMLInputElement).click());
    el.querySelector("#imp")?.addEventListener("change", async (ev) => {
      const file = (ev.target as HTMLInputElement).files?.[0];
      if (!file) return;
      const text = await file.text();
      const preview = await api<{ rows: { row: number; status: string; error?: string; warning?: string; question?: string }[] }>("/questions/import", { method: "POST", headers: { "Content-Type": "text/csv" }, body: text });
      if (!preview.ok) { toast(preview.error, "error"); return; }
      const rows = preview.data.rows || [];
      const bad = rows.filter((x) => x.status === "INVALID").length;
      el.querySelector("#qform")!.innerHTML = `<h3>Preview import</h3><div class="adm-pre">${esc(JSON.stringify(rows, null, 2))}</div>
        <button class="adm-btn" id="commit" ${bad ? "disabled" : ""}>Commit ${bad ? "(perbaiki invalid dulu)" : ""}</button>`;
      el.querySelector("#commit")?.addEventListener("click", async () => {
        const c = await api("/questions/import?commit=1", { method: "POST", headers: { "Content-Type": "text/csv" }, body: text });
        toast(c.ok ? "Import selesai" : c.error, c.ok ? "success" : "error");
        void render();
      });
    });
    el.querySelectorAll<HTMLButtonElement>("[data-ed]").forEach((b) => b.addEventListener("click", () => {
      const row = (r.data.items || []).find((x) => x.id === b.dataset.ed);
      qForm(el.querySelector("#qform") as HTMLElement, row || null);
    }));
    el.querySelectorAll<HTMLButtonElement>("[data-off]").forEach((b) => b.addEventListener("click", async () => {
      const row = (r.data.items || []).find((x) => x.id === b.dataset.off) as Record<string, unknown> | undefined;
      if (!row) return;
      const res = await api(`/questions/${row.id}`, { method: "PUT", body: JSON.stringify({ ...row, active: !row.active }) });
      toast(res.ok ? "Status diubah" : res.error, res.ok ? "success" : "error");
      void render();
    }));
    el.querySelectorAll<HTMLButtonElement>("[data-del]").forEach((b) => b.addEventListener("click", async () => {
      const c = await confirmModal("Hapus soal", "Soft delete — tidak hilang permanen.");
      if (!c.ok) return;
      const res = await api(`/questions/${b.dataset.del}`, { method: "DELETE" });
      toast(res.ok ? "Dihapus" : res.error, res.ok ? "success" : "error");
      void render();
    }));
  };

  const qForm = (host: HTMLElement, row: Record<string, unknown> | null): void => {
    const v = (k: string) => esc(row?.[k] ?? "");
    host.innerHTML = `<form class="adm-form" id="qf">
      <h3>${row ? "Edit soal" : "Soal baru"}</h3>
      <input name="question" required placeholder="Pertanyaan" value="${v("question")}" />
      <input name="optionA" required placeholder="Opsi A" value="${v("optionA")}" />
      <input name="optionB" required placeholder="Opsi B" value="${v("optionB")}" />
      <input name="optionC" required placeholder="Opsi C" value="${v("optionC")}" />
      <input name="optionD" required placeholder="Opsi D" value="${v("optionD")}" />
      <select name="correctAnswer"><option>A</option><option>B</option><option>C</option><option>D</option></select>
      <textarea name="explanation" required placeholder="Penjelasan">${v("explanation")}</textarea>
      <select name="subject"><option>PAI</option><option>MATEMATIKA</option><option>BAHASA_INGGRIS</option><option>BAHASA_JAWA</option></select>
      <select name="grade"><option>SMA</option><option>SMP</option><option>SD</option></select>
      <select name="difficulty"><option>EASY</option><option>MEDIUM</option><option>HARD</option></select>
      <button class="adm-btn">Simpan</button>
    </form>`;
    const form = host.querySelector("#qf") as HTMLFormElement;
    if (row) {
      (form.elements.namedItem("correctAnswer") as HTMLSelectElement).value = String(row.correctAnswer || "A");
      (form.elements.namedItem("subject") as HTMLSelectElement).value = String(row.subject || "PAI");
      (form.elements.namedItem("difficulty") as HTMLSelectElement).value = String(row.difficulty || "EASY");
      (form.elements.namedItem("grade") as HTMLSelectElement).value = String(row.grade || "SMA");
    }
    form.addEventListener("submit", async (ev) => {
      ev.preventDefault();
      const fd = new FormData(form);
      const body: Record<string, unknown> = Object.fromEntries(fd.entries());
      if (row?.id) body.id = row.id;
      body.active = row ? row.active !== false : true;
      const res = row?.id
        ? await api(`/questions/${row.id}`, { method: "PUT", body: JSON.stringify(body) })
        : await api("/questions", { method: "POST", body: JSON.stringify(body) });
      toast(res.ok ? "Tersimpan" : res.error, res.ok ? "success" : "error");
      if (res.ok) void render();
    });
  };

  const matches = async (el: HTMLElement, id: string): Promise<void> => {
    if (id) {
      const r = await api<Record<string, unknown>>(`/matches/${id}`);
      if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
      el.innerHTML = `<button class="adm-btn ghost" data-back>← Matches</button>
        <p class="adm-muted">View only — tidak ada kontrol dadu / jawaban / posisi.</p>
        <div class="adm-pre">${esc(JSON.stringify(r.data, null, 2))}</div>
        ${r.data.live ? `<button class="adm-btn danger" id="term">Terminate</button>` : ""}`;
      el.querySelector("[data-back]")?.addEventListener("click", () => go("matches"));
      el.querySelector("#term")?.addEventListener("click", async () => {
        const c = await confirmModal("Terminate match", "Reward tidak diberikan. Ketik TERMINATE.", "TERMINATE");
        if (!c.ok) return;
        const res = await api(`/matches/${id}/terminate`, { method: "POST", body: JSON.stringify({ reason: c.reason, confirm: c.typed }) });
        toast(res.ok ? "Match dihentikan" : res.error, res.ok ? "success" : "error");
        void render();
      });
      return;
    }
    const r = await api<{ live: Record<string, unknown>[]; history: Record<string, unknown>[]; analytics: Record<string, unknown> }>(`/matches?page=${page}`);
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    const a = r.data.analytics || {};
    el.innerHTML = `<div class="adm-cards">${card("Avg duration (s)", Number(a.averageMatchDuration || 0).toFixed(0))}${card("Avg questions", Number(a.averageQuestions || 0).toFixed(1))}${card("Correct rate", Number(a.correctRate || 0).toFixed(1) + "%")}${card("Disconnect", Number(a.disconnectRate || 0).toFixed(1) + "%")}${card("Abandon", Number(a.abandonRate || 0).toFixed(1) + "%")}</div>
      <h3>Live</h3>
      <div class="adm-table-wrap"><table class="adm"><thead><tr><th>Match</th><th>Mode</th><th>Players</th><th>Status</th><th>Duration</th><th>Turn</th></tr></thead>
      <tbody>${(r.data.live || []).map((m) => `<tr data-id="${esc(m.matchId)}"><td>${esc(m.matchId)}</td><td>${esc(m.mode)}</td><td>${esc((m.players as string[] || []).join(", "))}</td><td>${esc(m.status)}</td><td>${esc(m.duration)}s</td><td>${esc(m.currentTurn)}</td></tr>`).join("") || `<tr><td colspan="6">Tidak ada match live</td></tr>`}</tbody></table></div>
      <h3>History</h3>
      <div class="adm-table-wrap"><table class="adm"><thead><tr><th>Match</th><th>Mode</th><th>Players</th><th>Status</th></tr></thead>
      <tbody>${(r.data.history || []).map((m) => `<tr data-id="${esc(m.matchId)}"><td>${esc(m.matchId)}</td><td>${esc(m.mode)}</td><td>${esc((m.players as string[] || []).join(", "))}</td><td>${esc(m.status)}</td></tr>`).join("")}</tbody></table></div>`;
    el.querySelectorAll("tbody tr[data-id]").forEach((tr) => tr.addEventListener("click", () => go("matches/" + (tr as HTMLElement).dataset.id)));
  };

  const reports = async (el: HTMLElement): Promise<void> => {
    const r = await api<{ items: Record<string, unknown>[] }>(`/reports?page=${page}`);
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    el.innerHTML = `<div class="adm-table-wrap"><table class="adm"><thead><tr><th>Reporter</th><th>Reported</th><th>Reason</th><th>Description</th><th>Match</th><th>Status</th><th>Time</th><th></th></tr></thead><tbody>
      ${(r.data.items || []).map((it) => `<tr>
        <td>${esc(it.reporterId)}</td><td>${esc(it.reportedId)}</td><td>${esc(it.reason)}</td><td>${esc(it.description)}</td>
        <td>${esc(it.matchId)}</td><td>${pill(String(it.status))}</td><td>${when(it.createdAt)}</td>
        <td><button class="adm-btn ghost" data-rs="${esc(it.id)}" data-user="${esc(it.reportedId)}">Resolve</button></td>
      </tr>`).join("")}</tbody></table></div>`;
    el.querySelectorAll<HTMLButtonElement>("[data-rs]").forEach((b) => b.addEventListener("click", async () => {
      const c = await confirmModal("Selesaikan laporan", "Resolution note wajib.");
      if (!c.ok) return;
      const res = await api(`/reports/${b.dataset.rs}/resolve`, { method: "POST", body: JSON.stringify({ status: "RESOLVED", note: c.reason, userId: b.dataset.user }) });
      toast(res.ok ? "Selesai" : res.error, res.ok ? "success" : "error");
      void render();
    }));
  };

  const achievements = async (el: HTMLElement): Promise<void> => {
    const r = await api<{ items: Record<string, unknown>[] }>("/achievements");
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    el.innerHTML = `<p class="adm-muted">Jangan hapus achievement yang sudah dimiliki — nonaktifkan.</p>
      ${(r.data.items || []).map((a, i) => `<div class="adm-card"><b>${esc(a.name)}</b> (${esc(a.code || a.id)})
        <p class="adm-muted">${esc(a.description)}</p>
        <label><input type="checkbox" data-i="${i}" ${a.active ? "checked" : ""}/> Aktif</label>
        XP <input data-xp="${i}" type="number" value="${esc(a.rewardXp)}" style="width:80px"/>
        Koin <input data-coin="${i}" type="number" value="${esc(a.rewardCoins)}" style="width:80px"/>
      </div>`).join("")}
      <button class="adm-btn" id="savea">Simpan</button>`;
    el.querySelector("#savea")?.addEventListener("click", async () => {
      const items = (r.data.items || []).map((a, i) => ({
        ...a,
        active: (el.querySelector(`[data-i="${i}"]`) as HTMLInputElement).checked,
        rewardXp: Number((el.querySelector(`[data-xp="${i}"]`) as HTMLInputElement).value),
        rewardCoins: Number((el.querySelector(`[data-coin="${i}"]`) as HTMLInputElement).value),
      }));
      const res = await api("/achievements", { method: "PUT", body: JSON.stringify({ items }) });
      toast(res.ok ? "Tersimpan" : res.error, res.ok ? "success" : "error");
    });
  };

  const rewards = async (el: HTMLElement): Promise<void> => {
    const r = await api<Record<string, unknown>>("/rewards");
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    const d = (r.data.dailyCoins as number[]) || [50, 75, 100, 125, 150, 200, 500];
    el.innerHTML = `<form class="adm-form" id="rf">
      <h3>Daily reward (7 hari)</h3>
      ${d.map((n, i) => `<label>Hari ${i + 1} koin <input name="d${i}" type="number" min="0" value="${n}"/></label>`).join("")}
      <h3>XP</h3>
      <label>Correct <input name="xpCorrect" type="number" min="0" value="${esc(r.data.xpCorrect)}"/></label>
      <label>Wrong <input name="xpWrong" type="number" min="0" value="${esc(r.data.xpWrong)}"/></label>
      <label>Timeout <input name="xpTimeout" type="number" min="0" value="${esc(r.data.xpTimeout)}"/></label>
      <label>Match complete <input name="xpMatch" type="number" min="0" value="${esc(r.data.xpMatch)}"/></label>
      <label>Win <input name="xpWin" type="number" min="0" value="${esc(r.data.xpWin)}"/></label>
      <h3>Coins</h3>
      <label>Match <input name="coinMatch" type="number" min="0" value="${esc(r.data.coinMatch)}"/></label>
      <label>Win <input name="coinWin" type="number" min="0" value="${esc(r.data.coinWin)}"/></label>
      <button class="adm-btn">Simpan</button>
    </form>`;
    el.querySelector("#rf")?.addEventListener("submit", async (ev) => {
      ev.preventDefault();
      const f = ev.target as HTMLFormElement;
      const num = (n: string) => Number((f.elements.namedItem(n) as HTMLInputElement).value);
      const dailyCoins = [0, 1, 2, 3, 4, 5, 6].map((i) => num("d" + i));
      const res = await api("/rewards", { method: "PUT", body: JSON.stringify({
        dailyCoins, xpCorrect: num("xpCorrect"), xpWrong: num("xpWrong"), xpTimeout: num("xpTimeout"),
        xpMatch: num("xpMatch"), xpWin: num("xpWin"), coinMatch: num("coinMatch"), coinWin: num("coinWin"),
      }) });
      toast(res.ok ? "Config v" + (res.data as { version?: number }).version : res.error, res.ok ? "success" : "error");
    });
  };

  const ranks = async (el: HTMLElement): Promise<void> => {
    const r = await api<Record<string, unknown>>("/ranks");
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    el.innerHTML = `<p class="adm-muted">Mengubah RR tidak memigrasi rank player secara massal.</p>
      <p>Tiers: ${esc((r.data.tiers as string[] || []).join(" → "))}</p>
      <form class="adm-form" id="rk"><label>Win RR <input name="winRr" type="number" min="0" value="${esc(r.data.winRr)}"/></label>
      <label>Loss RR <input name="lossRr" type="number" min="0" value="${esc(r.data.lossRr)}"/></label>
      <button class="adm-btn">Simpan</button></form>`;
    el.querySelector("#rk")?.addEventListener("submit", async (ev) => {
      ev.preventDefault();
      const f = ev.target as HTMLFormElement;
      const res = await api("/ranks", { method: "PUT", body: JSON.stringify({
        winRr: Number((f.elements.namedItem("winRr") as HTMLInputElement).value),
        lossRr: Number((f.elements.namedItem("lossRr") as HTMLInputElement).value),
      }) });
      toast(res.ok ? "Tersimpan" : res.error, res.ok ? "success" : "error");
    });
  };

  const seasons = async (el: HTMLElement): Promise<void> => {
    const r = await api<{ items: Record<string, unknown>[]; active: Record<string, unknown> }>("/seasons");
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    el.innerHTML = `<div class="adm-card"><b>Aktif:</b> ${esc(r.data.active?.name)} (${esc(r.data.active?.id)})</div>
      ${(r.data.items || []).map((s) => `<div class="adm-card">${esc(s.name)} ${pill(s.active ? "ACTIVE" : "ENDED")}
        ${s.active ? `<button class="adm-btn danger" data-end="${esc(s.id)}">End season</button>` : ""}</div>`).join("")}
      <form class="adm-form" id="ns"><h3>Season baru</h3>
        <input name="name" required placeholder="Nama season" />
        <label><input type="checkbox" name="active" /> Jadikan aktif</label>
        <button class="adm-btn">Buat</button></form>`;
    el.querySelector("#ns")?.addEventListener("submit", async (ev) => {
      ev.preventDefault();
      const f = ev.target as HTMLFormElement;
      const res = await api("/seasons", { method: "POST", body: JSON.stringify({
        name: (f.elements.namedItem("name") as HTMLInputElement).value, active: (f.elements.namedItem("active") as HTMLInputElement).checked,
      }) });
      toast(res.ok ? "Season tersimpan" : res.error, res.ok ? "success" : "error");
      if (res.ok) void render();
    });
    el.querySelectorAll<HTMLButtonElement>("[data-end]").forEach((b) => b.addEventListener("click", async () => {
      const prev = await api<Record<string, unknown>>(`/seasons/${b.dataset.end}/preview`);
      const c = await confirmModal("End season", `Preview: ${JSON.stringify(prev.data)} — ketik END SEASON.`, "END SEASON");
      if (!c.ok) return;
      const res = await api(`/seasons/${b.dataset.end}/end`, { method: "POST", body: JSON.stringify({ confirm: c.typed }) });
      toast(res.ok ? "Season ditutup" : res.error, res.ok ? "success" : "error");
      void render();
    }));
  };

  const config = async (el: HTMLElement): Promise<void> => {
    const [c, f] = await Promise.all([api<Record<string, unknown>>("/config"), api<Record<string, boolean>>("/flags")]);
    if (!c.ok) { el.innerHTML = `<p class="adm-err">${esc(c.error)}</p>`; return; }
    const a = (c.data.active || {}) as Record<string, unknown>;
    const vers = (c.data.versions as { version: number }[]) || [];
    el.innerHTML = `<p>Active version: <b>v${esc(c.data.activeVersion)}</b></p>
      <form class="adm-form" id="cf">
        ${["questionTimeLimit", "wrongAnswerPenalty", "maxPlayers", "minPlayers", "reconnectGrace", "matchmakingTimeout"].map((k) =>
          `<label>${k} <input name="${k}" type="number" value="${esc(a[k])}"/></label>`).join("")}
        <button class="adm-btn">Simpan config baru</button>
      </form>
      <h3>Feature flags</h3>
      ${["enableRanked", "enableDailyReward", "enableChat", "enableNewBoard"].map((k) =>
        `<label><input type="checkbox" data-fl="${k}" ${f.data?.[k] ? "checked" : ""}/> ${k}</label>`).join("<br/>")}
      <p><button class="adm-btn ghost" id="svfl">Simpan flags</button></p>
      <h3>Rollback</h3>
      ${vers.map((v) => `<button class="adm-btn ghost" data-rb="${v.version}">v${v.version}</button>`).join(" ")}`;
    el.querySelector("#cf")?.addEventListener("submit", async (ev) => {
      ev.preventDefault();
      const form = ev.target as HTMLFormElement;
      const body = { ...a };
      for (const k of ["questionTimeLimit", "wrongAnswerPenalty", "maxPlayers", "minPlayers", "reconnectGrace", "matchmakingTimeout"]) {
        body[k] = Number((form.elements.namedItem(k) as HTMLInputElement).value);
      }
      const conf = await confirmModal("Ubah config kritis", "Berlaku untuk event berikutnya, bukan match aktif.");
      if (!conf.ok) return;
      const res = await api("/config", { method: "PUT", body: JSON.stringify(body) });
      toast(res.ok ? "Config disimpan" : res.error, res.ok ? "success" : "error");
      void render();
    });
    el.querySelector("#svfl")?.addEventListener("click", async () => {
      const flags: Record<string, boolean> = {};
      el.querySelectorAll<HTMLInputElement>("[data-fl]").forEach((i) => { flags[i.dataset.fl || ""] = i.checked; });
      const res = await api("/flags", { method: "PUT", body: JSON.stringify(flags) });
      toast(res.ok ? "Flags disimpan" : res.error, res.ok ? "success" : "error");
    });
    el.querySelectorAll<HTMLButtonElement>("[data-rb]").forEach((b) => b.addEventListener("click", async () => {
      const conf = await confirmModal("Rollback config", "Kembali ke versi " + b.dataset.rb);
      if (!conf.ok) return;
      const res = await api("/config/rollback", { method: "POST", body: JSON.stringify({ version: Number(b.dataset.rb) }) });
      toast(res.ok ? "Rollback" : res.error, res.ok ? "success" : "error");
      void render();
    }));
  };

  const audit = async (el: HTMLElement): Promise<void> => {
    const r = await api<{ items: Record<string, unknown>[] }>(`/audit-logs?page=${page}`);
    if (!r.ok) { el.innerHTML = `<p class="adm-err">${esc(r.error)}</p>`; return; }
    el.innerHTML = `<div class="adm-table-wrap"><table class="adm"><thead><tr><th>Admin</th><th>Action</th><th>Target</th><th>Time</th></tr></thead><tbody>
      ${(r.data.items || []).map((it) => `<tr data-d='${esc(JSON.stringify(it))}'><td>${esc(it.adminName || it.adminId)}</td><td>${esc(it.action)}</td><td>${esc(it.targetType)} ${esc(it.targetId)}</td><td>${when(it.createdAt)}</td></tr>`).join("")}
    </tbody></table></div>`;
    el.querySelectorAll("tbody tr").forEach((tr) => tr.addEventListener("click", () => {
      alert((tr as HTMLElement).dataset.d || "");
    }));
  };

  window.addEventListener("popstate", () => void render());
  window.addEventListener("hashchange", () => void render());

  void (async () => {
    const tok = storedSessionToken();
    if (tok) {
      const r = await api<Me>("/me");
      if (r.ok) me = r.data;
      else if (r.status === 401) clearSession();
    }
    await render();
  })();
}

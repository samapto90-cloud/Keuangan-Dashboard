import { GAME_PHASE, GAME_TITLE, GAME_VERSION } from "../game/GameConfig";

export type GameSession = {
  token: string;
  playerId: string;
  username: string;
};

export type ContinueInfo = {
  playerId: string;
  username: string;
  email?: string;
  level: number;
  chapter: string;
  chapterTitle: string;
  chapterIndex: number;
  checkpoint: string;
  checkpointName: string;
  region: string;
  newJourney: boolean;
};

const TOKEN_KEY = "cahaya-session";
const NAME_KEY = "cahaya-username";

export function storedSessionToken(): string {
  return window.localStorage.getItem(TOKEN_KEY) || "";
}

export class TitleScreen {
  readonly root: HTMLElement;
  onEnter: ((session: GameSession, profile: ContinueInfo) => void) | null = null;
  onSettings: (() => void) | null = null;
  private view: "login" | "register" | "forgot" | "select" = "login";
  private error = "";
  private busy = false;
  private session: GameSession | null = null;
  private profile: ContinueInfo | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.id = "title-screen";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const btn = (e.target as HTMLElement).closest("[data-title]") as HTMLElement | null;
      if (!btn || this.busy) return;
      const act = btn.dataset.title;
      if (act === "show-register") this.setView("register");
      else if (act === "show-login") this.setView("login");
      else if (act === "show-forgot") this.setView("forgot");
      else if (act === "login") void this.submitLogin();
      else if (act === "register") void this.submitRegister();
      else if (act === "forgot") void this.submitForgot();
      else if (act === "continue") this.enter();
      else if (act === "new") this.enter();
      else if (act === "logout") void this.logout();
      else if (act === "settings") this.onSettings?.();
      else if (act === "credits") this.showCredits();
      else if (act === "back") this.setView(this.session ? "select" : "login");
    });
  }

  open(): void {
    this.root.hidden = false;
    const token = storedSessionToken();
    if (token) void this.resume(token);
    else this.setView("login");
  }

  close(): void {
    this.root.hidden = true;
  }

  private setView(view: TitleScreen["view"]): void {
    this.view = view;
    this.render();
  }

  private enter(): void {
    if (!this.session || !this.profile) return;
    this.onEnter?.(this.session, this.profile);
  }

  private async resume(token: string): Promise<void> {
    this.busy = true;
    this.render();
    const profile = await fetchProfile(token);
    this.busy = false;
    if (!profile) {
      window.localStorage.removeItem(TOKEN_KEY);
      this.session = null;
      this.setView("login");
      return;
    }
    this.session = { token, playerId: profile.playerId, username: profile.username };
    this.profile = profile;
    this.setView("select");
  }

  private async submitLogin(): Promise<void> {
    const username = this.field("user");
    const password = this.field("pass");
    this.busy = true;
    this.error = "";
    this.render();
    const out = await postJSON("/cahaya/api/login", { username, password });
    this.busy = false;
    if (!out.ok) {
      this.error = out.error;
      this.render();
      return;
    }
    this.storeSession(out.session);
    await this.resume(out.session.token);
  }

  private async submitRegister(): Promise<void> {
    const username = this.field("user");
    const email = this.field("email");
    const password = this.field("pass");
    const confirmPassword = this.field("confirm");
    this.busy = true;
    this.error = "";
    this.render();
    const out = await postJSON("/cahaya/api/register", { username, email, password, confirmPassword });
    this.busy = false;
    if (!out.ok) {
      this.error = out.error;
      this.render();
      return;
    }
    this.storeSession(out.session);
    await this.resume(out.session.token);
  }

  private async submitForgot(): Promise<void> {
    const username = this.field("user");
    const email = this.field("email");
    const password = this.field("pass");
    const confirmPassword = this.field("confirm");
    this.busy = true;
    this.error = "";
    this.render();
    const out = await postJSON("/cahaya/api/reset-password", { username, email, password, confirmPassword });
    this.busy = false;
    this.error = out.ok ? "Password diganti. Silakan masuk." : out.error;
    this.view = out.ok ? "login" : "forgot";
    this.render();
  }

  private async logout(): Promise<void> {
    const token = this.session?.token || storedSessionToken();
    if (token) {
      await fetch("/cahaya/api/logout", { method: "POST", headers: { Authorization: `Bearer ${token}` } });
    }
    window.localStorage.removeItem(TOKEN_KEY);
    window.localStorage.removeItem(NAME_KEY);
    this.session = null;
    this.profile = null;
    this.setView("login");
  }

  private storeSession(s: GameSession): void {
    this.session = s;
    window.localStorage.setItem(TOKEN_KEY, s.token);
    window.localStorage.setItem(NAME_KEY, s.username);
  }

  private field(name: string): string {
    const el = this.root.querySelector<HTMLInputElement>(`[name="${name}"]`);
    return el?.value?.trim() ?? "";
  }

  private render(): void {
    if (this.view === "select" && this.profile) {
      this.root.innerHTML = this.selectHTML();
      return;
    }
    if (this.view === "register") {
      this.root.innerHTML = this.formHTML("Daftar akun", "register", true, true);
      return;
    }
    if (this.view === "forgot") {
      this.root.innerHTML = this.formHTML("Ganti password", "forgot", true, true);
      return;
    }
    this.root.innerHTML = this.formHTML("Masuk", "login", false, false);
  }

  private selectHTML(): string {
    const p = this.profile!;
    const journey = p.newJourney
      ? `<button type="button" data-title="new">NEW JOURNEY</button>`
      : `<button type="button" data-title="continue">CONTINUE JOURNEY</button>
         <p class="title-ver">Chapter ${p.chapterIndex} · ${escapeHtml(p.chapterTitle)}<br/>Checkpoint: ${escapeHtml(p.checkpointName || "Desa Awal")}</p>`;
    return `
      <div class="title-card">
        <p class="title-kicker">ORIGINAL WORLD · ${GAME_PHASE}</p>
        <h1>${GAME_TITLE}</h1>
        <p class="title-ver">${GAME_VERSION} · ${escapeHtml(p.username)} · Lv ${p.level}</p>
        <div class="title-actions">
          ${journey}
          <button type="button" data-title="settings">PENGATURAN</button>
          <button type="button" data-title="credits">KREDIT</button>
          <button type="button" data-title="logout">KELUAR</button>
        </div>
      </div>`;
  }

  private formHTML(title: string, action: string, email: boolean, confirm: boolean): string {
    const err = this.error ? `<p class="title-error">${escapeHtml(this.error)}</p>` : "";
    const extra = email
      ? `<input name="email" type="email" autocomplete="email" placeholder="Email" required />`
      : "";
    const conf = confirm
      ? `<input name="confirm" type="password" autocomplete="new-password" placeholder="Konfirmasi password" required />`
      : "";
    const links =
      action === "login"
        ? `<button type="button" data-title="show-register">REGISTER</button>
           <button type="button" data-title="show-forgot">FORGOT PASSWORD</button>`
        : `<button type="button" data-title="show-login">KEMBALI KE LOGIN</button>`;
    return `
      <div class="title-card">
        <p class="title-kicker">ORIGINAL WORLD · ${GAME_PHASE}</p>
        <h1>${GAME_TITLE}</h1>
        <p class="title-ver">${title}</p>
        <form class="title-form" onsubmit="return false">
          <input name="user" type="text" autocomplete="username" placeholder="Username / Email" maxlength="32" required />
          ${extra}
          <input name="pass" type="password" autocomplete="${confirm ? "new-password" : "current-password"}" placeholder="Password" required />
          ${conf}
        </form>
        ${err}
        <div class="title-actions">
          <button type="button" data-title="${action}" ${this.busy ? "disabled" : ""}>${this.busy ? "..." : title.toUpperCase()}</button>
          ${links}
          <button type="button" data-title="credits">KREDIT</button>
        </div>
      </div>`;
  }

  private showCredits(): void {
    this.root.innerHTML = `
      <div class="title-card">
        <p class="title-kicker">KREDIT</p>
        <h1>${GAME_TITLE}</h1>
        <ul class="title-credits">
          <li>Original Game</li>
          <li>Original World</li>
          <li>Original Characters</li>
          <li>Original Music</li>
          <li>Original Assets</li>
        </ul>
        <p class="title-ver">Tidak memakai aset berhak cipta dari game lain.</p>
        <div class="title-actions">
          <button type="button" data-title="back">KEMBALI</button>
        </div>
      </div>`;
  }
}

function escapeHtml(s: string): string {
  return s.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] || c));
}

async function fetchProfile(token: string): Promise<ContinueInfo | null> {
  const res = await fetch("/cahaya/api/profile", { headers: { Authorization: `Bearer ${token}` } });
  if (!res.ok) return null;
  return (await res.json()) as ContinueInfo;
}

async function postJSON(
  url: string,
  body: Record<string, string>,
): Promise<{ ok: true; session: GameSession } | { ok: false; error: string }> {
  try {
    const res = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = (await res.json()) as { error?: string; token?: string; playerId?: string; username?: string };
    if (!res.ok) return { ok: false, error: data.error || "gagal" };
    if (data.token) {
      return { ok: true, session: { token: data.token, playerId: data.playerId || "", username: data.username || "" } };
    }
    return { ok: true, session: { token: "", playerId: "", username: "" } };
  } catch {
    return { ok: false, error: "tidak bisa menghubungi server" };
  }
}

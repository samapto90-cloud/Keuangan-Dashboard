import { GAME_PHASE, GAME_TITLE, GAME_VERSION } from "../game/board/constants";
import { BOARD_SIZE, MAX_PLAYERS } from "../game/board/config";
import { SNAKES } from "../game/snakes";
import { LADDERS } from "../game/ladders";
import { DICE_FACES } from "../game/dice";
import { QUESTION_BANK } from "../game/questions/model";
import { FOUNDATION_STATES } from "../foundation";
import {
  apiLogin,
  apiLogout,
  apiProfile,
  apiRegister,
  clearSession,
  saveSession,
  storedSessionToken,
  type Profile,
  type Session,
} from "../auth/session";
import { connectLobby } from "../game/multiplayer/socket";

type Screen = "home" | "login" | "register" | "lobby";

export function mountApp(root: HTMLElement): void {
  const embed = new URLSearchParams(window.location.search).get("embed") === "1";
  if (embed) document.body.classList.add("embed");

  let screen: Screen = "home";
  let error = "";
  let busy = false;
  let session: Session | null = null;
  let profile: Profile | null = null;
  let wsNote = "WebSocket belum terhubung";

  const render = (): void => {
    root.innerHTML = layout();
    bind();
  };

  const layout = (): string => {
    if (screen === "login") return shell(loginForm());
    if (screen === "register") return shell(registerForm());
    if (screen === "lobby") return shell(lobbyView());
    return shell(homeView());
  };

  const shell = (inner: string): string => `
    <main class="nt-shell">
      <header class="nt-top">
        <p class="nt-kicker">${GAME_VERSION} · ${GAME_PHASE}</p>
        <h1>${GAME_TITLE}</h1>
      </header>
      ${inner}
      ${error ? `<p class="nt-error" role="alert">${escapeHtml(error)}</p>` : ""}
    </main>`;

  const homeView = (): string => `
    <section class="nt-card">
      <p class="nt-lead">Main bersama di papan 1–${BOARD_SIZE}. Maksimal ${MAX_PLAYERS} pemain per ruangan. ${Object.keys(SNAKES).length} ular, ${Object.keys(LADDERS).length} tangga, dadu ${DICE_FACES} sisi. Soal tersimpan: ${QUESTION_BANK.length}. Status mesin: ${FOUNDATION_STATES[0]}.</p>
      <div class="nt-actions">
        <button type="button" class="nt-btn nt-btn-primary" data-go="play">PLAY ONLINE</button>
        <button type="button" class="nt-btn" data-go="create" disabled title="Fase 2">CREATE ROOM</button>
        <button type="button" class="nt-btn" data-go="join" disabled title="Fase 2">JOIN ROOM</button>
      </div>
      <p class="nt-hint">CREATE ROOM dan JOIN ROOM aktif setelah fondasi papan (fase 2).</p>
    </section>`;

  const loginForm = (): string => `
    <form class="nt-card nt-form" data-form="login">
      <label>Username / Email<input name="user" autocomplete="username" required /></label>
      <label>Password<input name="pass" type="password" autocomplete="current-password" required /></label>
      <button class="nt-btn nt-btn-primary" type="submit" ${busy ? "disabled" : ""}>MASUK</button>
      <p class="nt-links"><button type="button" data-go="register">Daftar</button> · <button type="button" data-go="home">Beranda</button></p>
    </form>`;

  const registerForm = (): string => `
    <form class="nt-card nt-form" data-form="register">
      <label>Username<input name="user" required minlength="3" /></label>
      <label>Email<input name="email" type="email" required /></label>
      <label>Password<input name="pass" type="password" required minlength="8" /></label>
      <label>Konfirmasi<input name="confirm" type="password" required minlength="8" /></label>
      <button class="nt-btn nt-btn-primary" type="submit" ${busy ? "disabled" : ""}>DAFTAR</button>
      <p class="nt-links"><button type="button" data-go="login">Sudah punya akun</button></p>
    </form>`;

  const lobbyView = (): string => `
    <section class="nt-card">
      <p>Halo, <strong>${escapeHtml(profile?.username || session?.username || "")}</strong></p>
      <p class="nt-lead">Lobby fase 1 — siap menunggu papan dan dadu.</p>
      <p class="nt-hint">${escapeHtml(wsNote)}</p>
      <div class="nt-actions">
        <button type="button" class="nt-btn" data-go="home">Beranda</button>
        <button type="button" class="nt-btn nt-btn-ghost" data-go="logout">Keluar</button>
      </div>
    </section>`;

  const bind = (): void => {
    root.querySelectorAll<HTMLButtonElement>("[data-go]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const go = btn.dataset.go;
        if (go === "play") void goPlay();
        else if (go === "login") setScreen("login");
        else if (go === "register") setScreen("register");
        else if (go === "home") setScreen("home");
        else if (go === "logout") void doLogout();
      });
    });
    const login = root.querySelector<HTMLFormElement>("[data-form=login]");
    login?.addEventListener("submit", (e) => {
      e.preventDefault();
      const fd = new FormData(login);
      void doLogin(String(fd.get("user") || ""), String(fd.get("pass") || ""));
    });
    const reg = root.querySelector<HTMLFormElement>("[data-form=register]");
    reg?.addEventListener("submit", (e) => {
      e.preventDefault();
      const fd = new FormData(reg);
      void doRegister(
        String(fd.get("user") || ""),
        String(fd.get("email") || ""),
        String(fd.get("pass") || ""),
        String(fd.get("confirm") || ""),
      );
    });
  };

  const setScreen = (next: Screen): void => {
    screen = next;
    error = "";
    render();
  };

  const goPlay = async (): Promise<void> => {
    const token = storedSessionToken();
    if (!token) {
      setScreen("login");
      return;
    }
    busy = true;
    render();
    const p = await apiProfile(token);
    busy = false;
    if (!p) {
      clearSession();
      setScreen("login");
      return;
    }
    session = { token, playerId: p.playerId, username: p.username };
    profile = p;
    connectLobby(token, () => {
      wsNote = "WebSocket terhubung · lobby aktif";
      render();
    });
    setScreen("lobby");
  };

  const doLogin = async (username: string, password: string): Promise<void> => {
    busy = true;
    error = "";
    render();
    const out = await apiLogin(username, password);
    busy = false;
    if (!out.ok) {
      error = out.error;
      render();
      return;
    }
    saveSession(out.data);
    session = out.data;
    await goPlay();
  };

  const doRegister = async (username: string, email: string, password: string, confirm: string): Promise<void> => {
    busy = true;
    error = "";
    render();
    const out = await apiRegister(username, email, password, confirm);
    busy = false;
    if (!out.ok) {
      error = out.error;
      render();
      return;
    }
    saveSession(out.data);
    session = out.data;
    await goPlay();
  };

  const doLogout = async (): Promise<void> => {
    const token = storedSessionToken();
    if (token) await apiLogout(token);
    clearSession();
    session = null;
    profile = null;
    setScreen("home");
  };

  void (async () => {
    const token = storedSessionToken();
    if (!token) {
      render();
      return;
    }
    const p = await apiProfile(token);
    if (p) {
      session = { token, playerId: p.playerId, username: p.username };
      profile = p;
      screen = embed ? "login" : "home";
    }
    render();
  })();
}

function escapeHtml(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

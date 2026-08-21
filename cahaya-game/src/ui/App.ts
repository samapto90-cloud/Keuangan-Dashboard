import { GAME_TITLE, GAME_VERSION, RELEASE_NAME } from "../game/board/constants";
import { BOARD_SIZE, MAX_PLAYERS } from "../game/board/config";
import { SNAKES } from "../game/snakes";
import { LADDERS } from "../game/ladders";
import { DICE_FACES } from "../game/dice";
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
import { GameClient } from "../game/multiplayer/socket";
import { WS_EVENTS } from "../game/multiplayer/events";
import { mountBoard } from "./BoardScreen";
import { mountOnline } from "./OnlinePlay";
import { openProfileModal } from "./ProfileScreen";
import { openFriendsModal, openNotificationsModal } from "./SocialScreen";
import { openLeaderboardModal } from "./LeaderboardScreen";
import { loadPrefs, savePrefs, type FontSize, type Quality, type ThemeMode } from "./prefs";
import { applyAudioPrefs, playSfx, startMusic } from "../audio/manager";
import { icon } from "./icons";
import { closeModal, showModal, toast } from "./chrome";
import { fetchNotes, fetchPrivacy, savePrivacy } from "../social/api";
import { openFeedbackModal } from "./FeedbackForm";
import { bindInstallButton, renderInstallButton } from "./pwa";
import { openHowToPlayModal, resetTutorial, mountOnboarding } from "./Onboarding";

type Screen = "home" | "login" | "register" | "lobby" | "board" | "online";

export function mountApp(root: HTMLElement): void {
  const embed = new URLSearchParams(window.location.search).get("embed") === "1";
  if (embed) document.body.classList.add("embed");

  let screen: Screen = "home";
  let error = "";
  let busy = false;
  let session: Session | null = null;
  let profile: Profile | null = null;
  let wsNote = "WebSocket belum terhubung";
  let playerCount = 2;
  let withNpc = false;
  let eduGrade: "SD" | "SMA" = ((): "SD" | "SMA" => {
    try {
      const g = localStorage.getItem("ular-edu-grade");
      return g === "SD" ? "SD" : "SMA";
    } catch {
      return "SMA";
    }
  })();
  let gradeChosen = true; // sudah ada default / pilihan beranda — jangan tanya ulang
  const setEduGrade = (g: "SD" | "SMA"): void => {
    eduGrade = g;
    gradeChosen = true;
    try {
      localStorage.setItem("ular-edu-grade", g);
    } catch {
      /* ignore */
    }
  };
  let net: GameClient | null = null;
  let unread = 0;
  let startQueue: "" | "CASUAL" | "RANKED" = "";
  let unsubGlobal: (() => void) | null = null;

  const ensurePresence = (token: string): GameClient => {
    if (!net) net = new GameClient();
    if (net.status === "offline") net.connect(token);
    if (!unsubGlobal) {
      unsubGlobal = net.addListener((type, data) => {
        if (type === WS_EVENTS.GAME_INVITE && screen !== "online") {
          const inviteId = String(data.inviteId || "");
          const username = String(data.username || "Pemain");
          if (!inviteId) return;
          showModal(
            "game-invite",
            `<h2>Undangan bermain</h2>
            <p class="nt-lead"><strong>${escapeHtml(username)}</strong> mengajakmu main Ular Tangga.</p>
            <div class="nt-actions">
              <button type="button" class="nt-btn nt-btn-primary" data-inv-yes>Terima</button>
              <button type="button" class="nt-btn" data-inv-no>Tolak</button>
            </div>`,
          );
          const layer = document.querySelector("[data-modal=game-invite]");
          layer?.querySelector("[data-inv-yes]")?.addEventListener("click", () => {
            net?.send(WS_EVENTS.INVITE_RESPOND, { inviteId, accept: true });
            closeModal("game-invite");
            startQueue = "";
            setScreen("online");
          });
          layer?.querySelector("[data-inv-no]")?.addEventListener("click", () => {
            net?.send(WS_EVENTS.INVITE_RESPOND, { inviteId, accept: false });
            closeModal("game-invite");
          });
        }
        if (type === WS_EVENTS.SOCIAL_NOTIFY && screen !== "online") {
          toast(String(data.title || data.type || "Notifikasi"), "info");
          void refreshUnread().then(() => {
            if (screen === "home") render();
          });
        }
      });
    }
    return net;
  };

  const render = (): void => {
    if (screen === "online") {
      if (!net) {
        screen = "login";
      } else if (!root.querySelector(".room-shell") && !root.querySelector(".board-root")) {
        root.innerHTML = "";
        mountOnline(root, { client: net, onExit: () => setScreen("home"), startQueue, grade: eduGrade });
        startQueue = "";
      }
      return;
    }
    if (screen === "board") {
      root.innerHTML = "";
      const human = profile?.username || session?.username || "Pemain";
      const names = withNpc
        ? [human, "Ganjar", "Anies", "Sri Mulyani"].slice(0, playerCount)
        : [human || "Prabowo", "Ganjar", "Anies", "Sri Mulyani"].slice(0, playerCount);
      const vsNpc = withNpc;
      withNpc = false;
      mountBoard(root, { names, withNpc: vsNpc, grade: eduGrade, onExit: () => setScreen("home") });
      return;
    }
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
    <main class="nt-shell page-in">
      <header class="nt-top">
        <p class="nt-kicker">${GAME_TITLE}</p>
        <h1>ULAR TANGGA<br/>NUSANTARA</h1>
        <p class="nt-hint">${GAME_VERSION} · ${RELEASE_NAME}</p>
        ${session ? `<button type="button" class="bell-btn" data-go="notes" aria-label="Notifikasi">${icon("bell")} ${unread ? `<span class="note-badge">${unread}</span>` : ""}</button>` : ""}
      </header>
      ${inner}
      ${error ? `<p class="nt-error" role="alert">${escapeHtml(error)}</p>` : ""}
    </main>`;

  const homeView = (): string => `
    <section class="nt-card">
      <p class="nt-lead">Petualangan papan 1–${BOARD_SIZE}. Maksimal ${MAX_PLAYERS} pemain. ${Object.keys(SNAKES).length} ular, ${Object.keys(LADDERS).length} tangga, dadu ${DICE_FACES} sisi.</p>
      <div class="nt-actions">
        <button type="button" class="nt-btn nt-btn-primary" data-go="play">${icon("users")} PLAY ONLINE</button>
        <button type="button" class="nt-btn nt-btn-primary" data-go="ranked">${icon("trophy")} RANKED</button>
        <button type="button" class="nt-btn" data-go="friends">${icon("friends")} TEMAN</button>
        <button type="button" class="nt-btn" data-go="board-lb">${icon("trophy")} LEADERBOARD</button>
        <button type="button" class="nt-btn" data-go="how">${icon("book")} CARA BERMAIN</button>
        <button type="button" class="nt-btn" data-go="settings">${icon("settings")} PENGATURAN</button>
        <button type="button" class="nt-btn" data-go="profile">${icon("user")} PROFIL</button>
        <button type="button" class="nt-btn" data-go="feedback">${icon("book")} FEEDBACK</button>
        ${renderInstallButton()}
        <button type="button" class="nt-btn nt-btn-ghost" data-go="board">MAIN PAPAN (lokal)</button>
        <button type="button" class="nt-btn nt-btn-primary" data-go="board-npc">🤖 MAIN VS NPC</button>
      </div>
      <p class="nt-hint">Online: lihat siapa yang online, ajak main, atau cari lawan otomatis (2–4 pemain). Lokal: lawan NPC.</p>
      <div class="nt-actions nt-row">
        <button type="button" class="nt-btn ${eduGrade === "SD" ? "nt-btn-primary" : ""}" data-go="grade-sd">Soal SD</button>
        <button type="button" class="nt-btn ${eduGrade === "SMA" ? "nt-btn-primary" : ""}" data-go="grade-sma">Soal SMA</button>
      </div>
      <p class="nt-hint">Bank soal aktif: <strong>${eduGrade}</strong> (pilih sebelum main).</p>
      <div class="nt-actions nt-row">
        <button type="button" class="nt-btn ${playerCount === 2 ? "nt-btn-primary" : ""}" data-go="p2">2 pemain</button>
        <button type="button" class="nt-btn ${playerCount === 3 ? "nt-btn-primary" : ""}" data-go="p3">3 pemain</button>
        <button type="button" class="nt-btn ${playerCount === 4 ? "nt-btn-primary" : ""}" data-go="p4">4 pemain</button>
      </div>
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
      <label>Username<input name="user" required minlength="3" maxlength="16" /></label>
      <label>Email<input name="email" type="email" required /></label>
      <label>Password<input name="pass" type="password" required minlength="8" /></label>
      <label>Konfirmasi<input name="confirm" type="password" required minlength="8" /></label>
      <button class="nt-btn nt-btn-primary" type="submit" ${busy ? "disabled" : ""}>DAFTAR</button>
      <p class="nt-links"><button type="button" data-go="login">Sudah punya akun</button></p>
    </form>`;

  const lobbyView = (): string => `
    <section class="nt-card">
      <p>Halo, <strong>${escapeHtml(profile?.username || session?.username || "")}</strong></p>
      <p class="nt-lead">Lobby siap. Buka papan untuk Board Engine fase 2.</p>
      <p class="nt-hint">${escapeHtml(wsNote)}</p>
      <div class="nt-actions">
        <button type="button" class="nt-btn nt-btn-primary" data-go="play">ROOM ONLINE</button>
        <button type="button" class="nt-btn nt-btn-primary" data-go="board">MAIN PAPAN</button>
        <button type="button" class="nt-btn nt-btn-primary" data-go="board-npc">🤖 MAIN VS NPC</button>
        <button type="button" class="nt-btn" data-go="home">Beranda</button>
        <button type="button" class="nt-btn nt-btn-ghost" data-go="logout">Keluar</button>
      </div>
    </section>`;

  const bind = (): void => {
    root.querySelectorAll<HTMLButtonElement>("[data-go]").forEach((btn) => {
      btn.addEventListener("click", () => {
        playSfx("button");
        const go = btn.dataset.go;
        if (go === "play" || go === "online") startWithGrade(() => void goPlay(""));
        else if (go === "ranked") startWithGrade(() => void goPlay("RANKED"));
        else if (go === "friends") {
          const token = storedSessionToken();
          if (token) ensurePresence(token);
          void openFriendsModal({ client: net });
        }
        else if (go === "board-lb") void openLeaderboardModal();
        else if (go === "notes") void openNotes();
        else if (go === "board") {
          withNpc = false;
          startWithGrade(() => setScreen("board"));
        } else if (go === "board-npc") {
          withNpc = true;
          startWithGrade(() => setScreen("board"));
        }
        else if (go === "grade-sd") {
          setEduGrade("SD");
          render();
        } else if (go === "grade-sma") {
          setEduGrade("SMA");
          render();
        }
        else if (go === "how") openHowToPlayModal();
        else if (go === "feedback") openFeedbackModal({ page: "home" });
        else if (go === "settings") openSettings();
        else if (go === "profile") openProfile();
        else if (go === "p2") {
          playerCount = 2;
          render();
        } else if (go === "p3") {
          playerCount = 3;
          render();
        } else if (go === "p4") {
          playerCount = 4;
          render();
        }
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

  /** Pakai pilihan beranda; modal hanya sekali jika belum pernah dipilih. */
  const startWithGrade = (next: () => void): void => {
    if (gradeChosen) {
      next();
      return;
    }
    showModal(
      "grade-pick",
      `<h2>Pilih tingkat soal</h2>
      <p class="nt-lead">Bank soal dari kumpulan SD & SMA. Pilih sebelum permainan dimulai.</p>
      <div class="nt-actions">
        <button type="button" class="nt-btn ${eduGrade === "SD" ? "nt-btn-primary" : ""}" data-grade="SD">📚 Soal SD</button>
        <button type="button" class="nt-btn ${eduGrade === "SMA" ? "nt-btn-primary" : ""}" data-grade="SMA">🎓 Soal SMA</button>
      </div>
      <button type="button" class="nt-btn nt-btn-primary" data-grade-go>Lanjut</button>`,
    );
    const layer = document.querySelector("[data-modal=grade-pick]");
    layer?.querySelectorAll<HTMLButtonElement>("[data-grade]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const g = btn.dataset.grade;
        if (g === "SD" || g === "SMA") {
          setEduGrade(g);
          layer.querySelectorAll("[data-grade]").forEach((b) => b.classList.toggle("nt-btn-primary", (b as HTMLElement).dataset.grade === g));
        }
      });
    });
    layer?.querySelector("[data-grade-go]")?.addEventListener("click", () => {
      setEduGrade(eduGrade);
      closeModal();
      next();
    });
  };

  const openSettings = (): void => {
    const p = loadPrefs();
    showModal(
      "settings",
      `<h2>Pengaturan</h2>
      <div class="settings-grid">
        <div class="setting-section"><h3>Audio</h3>
          <label class="setting-row">Master <input type="range" min="0" max="100" value="${Math.round(p.master * 100)}" data-k="master"/></label>
          <label class="setting-row">Musik <input type="range" min="0" max="100" value="${Math.round(p.music * 100)}" data-k="music"/></label>
          <label class="setting-row">Efek suara <input type="range" min="0" max="100" value="${Math.round(p.sfx * 100)}" data-k="sfx"/></label>
          <label class="setting-row">Bisukan semua <input type="checkbox" data-k="mute" ${p.mute ? "checked" : ""}/></label>
        </div>
        <div class="setting-section"><h3>Grafis & Aksesibilitas</h3>
          <label class="setting-row">Kualitas
            <select data-k="quality">
              <option value="low" ${p.quality === "low" ? "selected" : ""}>Rendah</option>
              <option value="medium" ${p.quality === "medium" ? "selected" : ""}>Sedang</option>
              <option value="high" ${p.quality === "high" ? "selected" : ""}>Tinggi</option>
            </select>
          </label>
          <label class="setting-row">Gerakan berkurang <input type="checkbox" data-k="reduced" ${p.reduced ? "checked" : ""}/></label>
          <label class="setting-row">Ukuran teks
            <select data-k="fontSize">
              <option value="small" ${p.fontSize === "small" ? "selected" : ""}>Kecil</option>
              <option value="normal" ${p.fontSize === "normal" ? "selected" : ""}>Normal</option>
              <option value="large" ${p.fontSize === "large" ? "selected" : ""}>Besar</option>
            </select>
          </label>
          <label class="setting-row">Tema
            <select data-k="theme">
              <option value="system" ${p.theme === "system" ? "selected" : ""}>Sistem</option>
              <option value="light" ${p.theme === "light" ? "selected" : ""}>Terang</option>
              <option value="dark" ${p.theme === "dark" ? "selected" : ""}>Gelap</option>
            </select>
          </label>
        </div>
        <div class="setting-section"><h3>Lainnya</h3>
          <label class="setting-row">Notifikasi in-game <input type="checkbox" data-k="notifications" ${p.notifications ? "checked" : ""}/></label>
          <label class="setting-row">Bahasa <select disabled><option>Bahasa Indonesia</option></select></label>
          <button type="button" class="nt-btn nt-btn-ghost" data-reset-tutorial>Reset Tutorial</button>
        </div>
        <p class="nt-kicker">Privasi</p>
        <label class="setting-row">Izinkan permintaan teman <input type="checkbox" data-p="allowFriendRequests" checked/></label>
        <label class="setting-row">Izinkan undangan game <input type="checkbox" data-p="allowGameInvites" checked/></label>
        <label class="setting-row">Tampilkan status online <input type="checkbox" data-p="showOnlineStatus" checked/></label>
      </div>
      <button type="button" class="nt-btn nt-btn-primary" data-close>Simpan</button>`,
    );
    const layer = document.querySelector("[data-modal=settings]");
    const token = storedSessionToken();
    if (token) {
      void fetchPrivacy(token).then((out) => {
        if (!out.ok) return;
        const set = (k: "allowFriendRequests" | "allowGameInvites" | "showOnlineStatus") => {
          const el = layer?.querySelector<HTMLInputElement>(`[data-p=${k}]`);
          if (el) el.checked = Boolean(out.data[k]);
        };
        set("allowFriendRequests");
        set("allowGameInvites");
        set("showOnlineStatus");
      });
    }
    layer?.querySelectorAll("input,select").forEach((el) => {
      el.addEventListener("change", () => {
        const k = (el as HTMLElement).dataset.k;
        if (k === "master") savePrefs({ master: Number((el as HTMLInputElement).value) / 100 });
        if (k === "music") savePrefs({ music: Number((el as HTMLInputElement).value) / 100 });
        if (k === "sfx") savePrefs({ sfx: Number((el as HTMLInputElement).value) / 100 });
        if (k === "mute") savePrefs({ mute: (el as HTMLInputElement).checked });
        if (k === "reduced") savePrefs({ reduced: (el as HTMLInputElement).checked });
        if (k === "quality") savePrefs({ quality: (el as HTMLSelectElement).value as Quality });
        if (k === "fontSize") savePrefs({ fontSize: (el as HTMLSelectElement).value as FontSize });
        if (k === "theme") savePrefs({ theme: (el as HTMLSelectElement).value as ThemeMode });
        if (k === "notifications") savePrefs({ notifications: (el as HTMLInputElement).checked });
        applyAudioPrefs();
      });
    });
    layer?.querySelector("[data-reset-tutorial]")?.addEventListener("click", () => {
      resetTutorial();
      toast("Tutorial akan ditampilkan saat permainan berikutnya.", "info");
    });
    bindInstallButton(layer || document);
    layer?.querySelectorAll<HTMLInputElement>("[data-p]").forEach((el) => {
      el.addEventListener("change", () => {
        if (!token) return;
        const box = (k: string) => Boolean(layer?.querySelector<HTMLInputElement>(`[data-p=${k}]`)?.checked);
        void savePrivacy(token, {
          allowFriendRequests: box("allowFriendRequests"),
          allowGameInvites: box("allowGameInvites"),
          showOnlineStatus: box("showOnlineStatus"),
        });
      });
    });
    layer?.querySelector("[data-close]")?.addEventListener("click", () => closeModal("settings"));
  };

  const openProfile = (): void => {
    const token = storedSessionToken();
    if (!token) {
      toast("Masuk untuk melihat profil.", "warning");
      setScreen("login");
      return;
    }
    void openProfileModal(token);
  };

  const setScreen = (next: Screen): void => {
    screen = next;
    error = "";
    render();
  };

  const openNotes = async (): Promise<void> => {
    unread = await openNotificationsModal();
    render();
  };

  const refreshUnread = async (): Promise<void> => {
    const token = storedSessionToken();
    if (!token) return;
    const out = await fetchNotes(token);
    if (out.ok) unread = out.data.unread || 0;
  };

  const goPlay = async (queue: "" | "CASUAL" | "RANKED" = ""): Promise<void> => {
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
    startQueue = queue;
    const client = ensurePresence(token);
    setScreen("online");
    if (client.status === "offline") client.connect(token);
  };

  const doLogin = async (username: string, password: string): Promise<void> => {
    busy = true;
    error = "";
    render();
    const out = await apiLogin(username, password);
    busy = false;
    if (!out.ok) {
      error = out.error;
      toast(out.error, "error");
      render();
      return;
    }
    saveSession(out.data);
    session = out.data;
    ensurePresence(out.data.token);
    if (!loadPrefs().tutorialCompleted) {
      root.innerHTML = "";
      mountOnboarding(root, { defaultUsername: out.data.username, onDone: () => void goPlay() });
      return;
    }
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
      toast(out.error, "error");
      render();
      return;
    }
    saveSession(out.data);
    session = out.data;
    ensurePresence(out.data.token);
    if (!loadPrefs().tutorialCompleted) {
      root.innerHTML = "";
      mountOnboarding(root, { defaultUsername: out.data.username, onDone: () => void goPlay() });
      return;
    }
    await goPlay();
  };

  const doLogout = async (): Promise<void> => {
    const token = storedSessionToken();
    if (token) await apiLogout(token);
    clearSession();
    session = null;
    profile = null;
    unsubGlobal?.();
    unsubGlobal = null;
    net?.ws?.close();
    net = null;
    setScreen("home");
  };

  void (async () => {
    loadPrefs();
    applyAudioPrefs();
    document.body.addEventListener("pointerdown", () => startMusic(), { once: true });
    document.addEventListener("ular:open-settings", () => openSettings());
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
      ensurePresence(token);
      await refreshUnread();
    }
    render();
  })();
}

function escapeHtml(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

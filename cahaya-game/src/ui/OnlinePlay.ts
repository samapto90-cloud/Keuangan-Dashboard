import { BOARD_GRID, DEFAULT_BOARD, type BoardConfig } from "../game/board/config";
import { BoardEngine, tokenBoardPercent } from "../game/board/engine";
import { BoardAnimationManager } from "../game/board/animation";
import { GameClient } from "../game/multiplayer/socket";
import { WS_EVENTS } from "../game/multiplayer/events";
import { playSfx } from "../audio/manager";
import { mountQuestionOverlay, unmountQuestionOverlay, type QuestionPublic, type QuestionResultView } from "./QuestionModal";
import { renderBoardRoutes } from "./boardDecor";
import { zoneClass, zoneDecor } from "./boardZones";
import { avatarSpriteHtml, pawnSpriteHtml, truncateName } from "../assets/registry";
import { connChip, icon } from "./icons";
import { confetti, setConnectionBanner, showMatchResultModal, showChampionCelebration, sparkle, spawnFloat, toast } from "./chrome";
import {
  applyPremiumRollState,
  activateBoardTheme,
  bindPremiumDock,
  paintPremiumStrip,
  paintPremiumTurnHud,
  premiumBoardShell,
  type PremiumSeat,
} from "./boardPremium";
import { storedSessionToken } from "../auth/session";
import { fetchProfile } from "../progress/api";
import { applyRewardFx } from "../progress/fx";
import type { RewardEvent } from "../progress/types";
import { pingTone } from "../social/api";
import { friendRequest } from "../social/api";
import { openFriendsModal } from "./SocialScreen";

import { openPowerInventory, powerGrantBanner } from "./PowerModal";
import { powerIcon, powerLabel, type PowerBag, type PowerKind } from "../game/powers";

type Seat = {
  userId: string;
  username: string;
  color: string;
  position: number;
  items?: Record<string, number>;
  isReady: boolean;
  isConnected: boolean;
  connState?: string;
  playState?: string;
};

type OnlineCard = { userId: string; username: string; status: string; inGame?: boolean; friends?: boolean };

function seatBag(s: Seat): PowerBag {
  const it = s.items || {};
  return {
    bomb: Number(it.bomb || 0),
    thunder: Number(it.thunder || 0),
    superman: Number(it.superman || 0),
  };
}

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export function mountOnline(root: HTMLElement, opts: { client: GameClient; onExit: () => void; startQueue?: "" | "CASUAL" | "RANKED"; grade?: "SD" | "SMA" }): void {
  let themeOff: (() => void) | null = null;
  const eduGrade = opts.grade === "SD" ? "SD" : "SMA";
  const onExit = (): void => {
    unsub?.();
    unsub = null;
    themeOff?.();
    themeOff = null;
    opts.onExit();
  };
  const cfg: BoardConfig = { ...DEFAULT_BOARD, snakes: { ...DEFAULT_BOARD.snakes }, ladders: { ...DEFAULT_BOARD.ladders } };
  const anim = new BoardAnimationManager();
  let snap: Record<string, unknown> = {};
  let view: "room" | "table" = "room";
  let error = "";
  let lastDice = 0;
  let animating = false;
  let conn: string = opts.client.status;
  const dev = import.meta.env.DEV;
  let activeQ: QuestionPublic | null = null;
  let qResult: QuestionResultView | null = null;
  let qTick = 0;
  let answered = false;
  let searching = false;
  let searchMode = "CASUAL";
  let searchAt = 0;
  let searchTick = 0;
  let pingMs = 0;
  let maxPlayers = 2;
  let matchFound: { ends: number; mode: string } | null = null;
  let readyCheck = false;
  let invite: { inviteId: string; username: string; expiresAt: number } | null = null;
  let bootQueue = opts.startQueue || "";
  let onlinePlayers: OnlineCard[] = [];
  let unsub: (() => void) | null = null;

  const players = (): Seat[] => (Array.isArray(snap.players) && snap.players[0] && typeof snap.players[0] === "object" ? (snap.players as Seat[]) : []);
  const code = (): string => String(snap.roomCode || "");
  const status = (): string => String(snap.status || "WAITING");
  const hostId = (): string => String(snap.hostId || "");
  const currentId = (): string => String(snap.currentPlayerId || "");
  const myId = (): string => opts.client.myId;
  const profileName = (): string => players().find((s) => s.userId === myId())?.username || "Pemain";
  const isHost = (): boolean => hostId() === myId();
  const cap = (): number => Number(snap.maxPlayers || maxPlayers || 4);
  const roomMode = (): string => String(snap.mode || "");

  const clock = (ms: number): string => {
    const s = Math.max(0, Math.floor(ms / 1000));
    return `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;
  };

  const paint = (): void => {
    if (view === "table") return paintTable();
    window.clearInterval(searchTick);
    const seats = players();
    const slots = Array.from({ length: cap() }, (_, i) => seats[i]);
    const meReady = Boolean(seats.find((s) => s.userId === myId())?.isReady);
    const canStart = seats.length >= 2 && seats.every((s) => s.isReady);
    const searchClock = searching ? clock(Date.now() - searchAt) : "00:00";
    const foundLeft = matchFound ? Math.max(0, Math.ceil((matchFound.ends - Date.now()) / 1000)) : 0;
    const chat = Array.isArray(snap.chat) ? (snap.chat as { username?: string; text?: string; emote?: string }[]) : [];
    root.innerHTML = `
      <main class="nt-shell room-shell page-in">
        <header class="nt-top">
          <p class="nt-kicker">ULAR TANGGA NUSANTARA</p>
          <h1>${code() ? "Menunggu lawan" : "Main Online"}</h1>
          <p>${code() ? `${seats.length}/${cap()} · ` : ""}${connChip(conn)} · ${pingTone(pingMs)}${roomMode() ? " · " + esc(roomMode()) : ""} · Soal ${eduGrade}</p>
        </header>
        ${searching && !matchFound ? `<section class="nt-card mm-card"><p class="nt-kicker">MENCARI LAWAN... ${esc(searchMode)} · ${maxPlayers} pemain</p><p class="mm-clock" id="search-clock">${searchClock}</p><button class="nt-btn" data-act="cancel-q">BATAL</button></section>` : ""}
        ${matchFound && !readyCheck ? `<section class="nt-card mm-card mm-found"><p>MATCH DITEMUKAN!</p><p class="mm-clock">${foundLeft}</p></section>` : ""}
        ${readyCheck ? `<section class="nt-card mm-card"><p class="nt-kicker">SIAP BERMAIN?</p><button class="nt-btn nt-btn-primary" data-act="match-ready">READY</button></section>` : ""}
        ${invite ? `<section class="nt-card invite-pop"><p>🎮 ${esc(invite.username).toUpperCase()} MENGUNDANGMU</p><p>Main Ular Tangga?</p><button class="nt-btn nt-btn-primary" data-act="inv-yes">TERIMA</button> <button class="nt-btn" data-act="inv-no">TOLAK</button></section>` : ""}
        ${!code() && !searching ? `<section class="nt-card room-hint">
          <p class="room-hint-title">Pemain online</p>
          <p class="room-hint-text">Pilih ukuran partai, lalu <strong>ajak bermain</strong> atau <strong>cari lawan otomatis</strong>. Tidak perlu kode room.</p>
          <div class="nt-row" style="margin:8px 0">${[2, 3, 4].map((n) => `<button class="nt-btn ${maxPlayers === n ? "nt-btn-primary" : ""}" data-max="${n}">${n} pemain</button>`).join("")}</div>
          <div class="pf-hist online-list">${onlinePlayers.length
            ? onlinePlayers.map((p) => `
              <article class="nt-card friend-row">
                <span><strong>${esc(p.username)}</strong> · ${esc(p.status === "IN_GAME" ? "Sedang main" : "Online")}</span>
                <div class="nt-row">
                  ${p.status !== "IN_GAME" ? `<button class="nt-btn nt-btn-primary" data-challenge="${esc(p.userId)}">Ajak main</button>` : ""}
                  ${!p.friends ? `<button class="nt-btn" data-addfriend="${esc(p.userId)}">Tambah teman</button>` : `<span class="nt-hint">Teman</span>`}
                </div>
              </article>`).join("")
            : `<p class="nt-hint">Belum ada pemain lain online. Bagikan link game atau tekan Cari lawan.</p>`}
          </div>
        </section>` : ""}
        ${code() ? `<section class="nt-card seats">
          <p class="seats-label">Pemain (${seats.length}/${cap()})</p>
          ${slots
            .map((s, i) =>
              s
                ? `<article class="seat-card ${s.userId === currentId() ? "glow" : ""} ${s.isReady ? "is-ready" : ""}">
                    <span class="seat-pawn is-sprite" style="--pawn:${s.color}">${avatarSpriteHtml(i, s.username)}</span>
                    <div><strong title="${esc(s.username)}">${esc(truncateName(s.username))}</strong>
                    <p>${s.isReady ? "READY" : "BELUM SIAP"} · ${s.isConnected ? "Online" : "Terputus"}</p></div>
                    ${isHost() && s.userId !== myId() ? `<button class="nt-btn nt-btn-ghost" data-kick="${esc(s.userId)}">KICK</button>` : ""}
                  </article>`
                : `<article class="seat-card empty"><span class="seat-empty-ico">＋</span> Slot kosong ${i + 1}</article>`,
            )
            .join("")}
        </section>` : ""}
        ${error ? `<p class="nt-error" role="alert">${esc(error)}</p>` : ""}
        <div class="nt-actions">
          ${!code() && !searching ? `<button class="nt-btn nt-btn-primary" data-act="queue-c">CARI LAWAN</button>` : ""}
          ${!code() && !searching ? `<button class="nt-btn nt-btn-primary" data-act="queue-r">RANKED</button>` : ""}
          ${!code() && !searching ? `<button class="nt-btn" data-act="friends">TEMAN</button>` : ""}
          ${!code() && !searching ? `<button class="nt-btn" data-act="refresh-online">Segarkan online</button>` : ""}
          ${code() && !readyCheck && !matchFound ? `<button class="nt-btn nt-btn-primary" data-act="ready">${meReady ? "BATAL READY" : "READY"}</button>` : ""}
          ${isHost() && code() && !matchFound ? `<button class="nt-btn nt-btn-primary" data-act="start" ${canStart ? "" : "disabled"}>MULAI PERMAINAN</button>` : ""}
          ${isHost() && code() && !matchFound && status() === "WAITING" ? `<div class="nt-row">${[2, 3, 4].map((n) => `<button class="nt-btn ${cap() === n ? "nt-btn-primary" : ""}" data-max="${n}">${n} pemain</button>`).join("")}</div>` : ""}
          ${code() ? `<button class="nt-btn" data-act="friends">TEMAN</button>` : ""}
          <button class="nt-btn" data-act="leave">${code() || searching ? "KELUAR" : "MENU"}</button>
          ${conn === "offline" ? `<button class="nt-btn nt-btn-primary" data-act="reconnect">SAMBUNG ULANG</button>` : ""}
        </div>
        ${code() ? `<div class="chat-box lobby-chat">
          <div id="chat-log">${chat.map((c) => `<p><strong>${esc(c.username || "")}:</strong> ${esc(c.text || c.emote || "")}</p>`).join("")}</div>
          <form data-act="chat" class="nt-row"><input name="msg" maxlength="200" placeholder="pesan" /><button class="nt-btn" type="submit">Kirim</button></form>
          <div class="emotes">${["👏", "😂", "🔥", "😮", "GG"].map((e) => `<button type="button" data-emote="${e}">${e}</button>`).join("")}</div>
        </div>` : ""}
        ${dev ? `<pre class="debug-panel">${esc(JSON.stringify({ matchId: snap.matchId, seq: snap.seq, status: status(), current: currentId(), conn, ping: pingMs, online: onlinePlayers.length }, null, 2))}</pre>` : ""}
      </main>`;
    bindRoom();
    if (searching || matchFound) {
      searchTick = window.setInterval(() => {
        const el = root.querySelector("#search-clock");
        if (el && searching && !matchFound) el.textContent = clock(Date.now() - searchAt);
        if (matchFound && Date.now() >= matchFound.ends && !readyCheck) {
          readyCheck = true;
          paint();
        } else if (matchFound && !readyCheck) {
          const clockEl = root.querySelector(".mm-found .mm-clock");
          if (clockEl) clockEl.textContent = String(Math.max(0, Math.ceil((matchFound.ends - Date.now()) / 1000)));
        }
      }, 250);
    }
  };

  const joinQueue = (mode: string): void => {
    searchMode = mode;
    searching = true;
    searchAt = Date.now();
    opts.client.send(WS_EVENTS.QUEUE_JOIN, { mode, region: "ID-JKT", grade: eduGrade, preferredSize: maxPlayers, maxPlayers });
    paint();
  };

  const bindRoom = (): void => {
    root.querySelector("[data-act=ready]")?.addEventListener("click", () => {
      const me = players().find((s) => s.userId === myId());
      opts.client.send(WS_EVENTS.PLAYER_READY, { ready: !me?.isReady });
    });
    root.querySelector("[data-act=start]")?.addEventListener("click", () => opts.client.send(WS_EVENTS.ROOM_START));
    root.querySelector("[data-act=leave]")?.addEventListener("click", () => {
      if (searching) opts.client.send(WS_EVENTS.QUEUE_LEAVE);
      if (code()) opts.client.send(WS_EVENTS.ROOM_LEAVE);
      onExit();
    });
    root.querySelector("[data-act=reconnect]")?.addEventListener("click", () => opts.client.reconnect());
    root.querySelector("[data-act=queue-c]")?.addEventListener("click", () => joinQueue("CASUAL"));
    root.querySelector("[data-act=queue-r]")?.addEventListener("click", () => joinQueue("RANKED"));
    root.querySelector("[data-act=cancel-q]")?.addEventListener("click", () => {
      opts.client.send(WS_EVENTS.QUEUE_LEAVE);
      searching = false;
      paint();
    });
    root.querySelector("[data-act=match-ready]")?.addEventListener("click", () => opts.client.send(WS_EVENTS.MATCH_READY, { ready: true }));
    root.querySelector("[data-act=friends]")?.addEventListener("click", () => void openFriendsModal({ client: opts.client }));
    root.querySelector("[data-act=refresh-online]")?.addEventListener("click", () => opts.client.send(WS_EVENTS.ONLINE_LIST_REQ));
    root.querySelector("[data-act=inv-yes]")?.addEventListener("click", () => {
      if (invite) opts.client.send(WS_EVENTS.INVITE_RESPOND, { inviteId: invite.inviteId, accept: true });
      invite = null;
      paint();
    });
    root.querySelector("[data-act=inv-no]")?.addEventListener("click", () => {
      if (invite) opts.client.send(WS_EVENTS.INVITE_RESPOND, { inviteId: invite.inviteId, accept: false });
      invite = null;
      paint();
    });
    root.querySelectorAll<HTMLButtonElement>("[data-max]").forEach((b) =>
      b.addEventListener("click", () => {
        maxPlayers = Number(b.dataset.max) || 2;
        if (code() && isHost()) opts.client.send(WS_EVENTS.ROOM_CREATE, { maxPlayers, grade: eduGrade });
        else paint();
      }),
    );
    root.querySelectorAll<HTMLButtonElement>("[data-challenge]").forEach((b) =>
      b.addEventListener("click", () => {
        opts.client.send(WS_EVENTS.GAME_INVITE, { userId: b.dataset.challenge, maxPlayers });
        toast("Undangan dikirim — menunggu lawan menerima.", "success");
      }),
    );
    root.querySelectorAll<HTMLButtonElement>("[data-addfriend]").forEach((b) =>
      b.addEventListener("click", () => {
        const token = storedSessionToken();
        if (!token) {
          toast("Masuk dulu untuk menambah teman.", "warning");
          return;
        }
        void friendRequest(token, b.dataset.addfriend || "").then((out) => {
          if (!out.ok) toast(out.error, "error");
          else {
            toast("Permintaan teman terkirim", "success");
            opts.client.send(WS_EVENTS.ONLINE_LIST_REQ);
          }
        });
      }),
    );
    root.querySelectorAll<HTMLButtonElement>("[data-kick]").forEach((b) =>
      b.addEventListener("click", () => opts.client.send(WS_EVENTS.ROOM_KICK, { userId: b.dataset.kick })),
    );
    const cf = root.querySelector<HTMLFormElement>("[data-act=chat]");
    cf?.addEventListener("submit", (e) => {
      e.preventDefault();
      const text = String(new FormData(cf).get("msg") || "");
      opts.client.send(WS_EVENTS.ROOM_CHAT, { text });
      cf.reset();
    });
    root.querySelectorAll<HTMLButtonElement>("[data-emote]").forEach((b) => {
      b.addEventListener("click", () => opts.client.send(WS_EVENTS.ROOM_EMOTE, { emote: b.dataset.emote || "" }));
    });
  };

  let grid: HTMLElement;
  let layer: HTMLElement;
  let tokens = new Map<string, HTMLElement>();
  let visual = new Map<string, number>();
  let diceEl: HTMLElement;
  let rollBtn: HTMLElement;
  let playerStrip: HTMLElement;
  let turnHud: HTMLElement;
  let tableBuilt = false;

  const ensureTable = (): void => {
    if (tableBuilt) return;
    tableBuilt = true;
    if (!themeOff) themeOff = activateBoardTheme();
    root.innerHTML = "";
    root.classList.add("board-root");
    root.innerHTML = `<div class="board-app">${premiumBoardShell(`${dev ? `<pre class="debug-panel premium-debug" id="dbg"></pre>` : ""}<div id="q-host"></div><div id="result-host"></div>`)}</div>`;
    grid = root.querySelector("#board-grid")!;
    layer = root.querySelector("#token-layer")!;
    playerStrip = root.querySelector("#player-strip")!;
    turnHud = root.querySelector("#turn-hud")!;
    diceEl = root.querySelector("#dice")!;
    rollBtn = root.querySelector("#roll-btn")!;
    for (let visualRow = 0; visualRow < BOARD_GRID; visualRow++) {
      const boardRow = BOARD_GRID - 1 - visualRow;
      for (let col = 0; col < BOARD_GRID; col++) {
        const pos = BoardEngine.getPositionFromCoordinate(boardRow, col)!;
        const cell = document.createElement("button");
        cell.type = "button";
        const tile = pos % 4;
        cell.className = `cell tile-${tile} color-${tile === 0 ? 1 : tile + 1} ${zoneClass(pos)}`;
        cell.dataset.pos = String(pos);
        if (pos === 1) cell.classList.add("cell-enter");
        if (pos === 100) cell.classList.add("cell-finish");
        if (cfg.snakes[pos]) cell.classList.add("cell-snake");
        if (cfg.ladders[pos]) cell.classList.add("cell-ladder");
        cell.innerHTML = `<span class="cell-num">${pos}</span>${zoneDecor(pos)}${pos === 1 ? `<span class="cell-tag">MASUK</span>` : ""}${pos === 100 ? `<span class="cell-tag">FINISH</span>` : ""}`;
        grid.appendChild(cell);
      }
    }
    const routes = root.querySelector<SVGSVGElement>("#route-layer");
    if (routes) renderBoardRoutes(routes, cfg.snakes, cfg.ladders);
    bindPremiumDock(root, {
      onRoll: () => {
        if (animating || currentId() !== myId()) return;
        applyPremiumRollState(rollBtn, "Lempar dadu", false);
        opts.client.send(WS_EVENTS.GAME_ROLL_REQUEST);
      },
      onExit: onExit,
      onChat: () => {
        const panel = root.querySelector<HTMLElement>("#board-chat");
        if (!panel) return;
        panel.hidden = !panel.hidden;
        paintBoardChat();
      },
      onHistory: () => toast("Item di papan: 💣×1 · ⚡×5 · ✈️×3. Ambil dengan mendarat di kotaknya.", "info"),
      onInventory: () => {
        if (currentId() !== myId() || status() !== "PLAYING" || animating || activeQ) {
          toast("Inventory hanya saat giliranmu.", "info");
          return;
        }
        const me = players().find((p) => p.userId === myId());
        if (!me) return;
        const host = root.querySelector<HTMLElement>("#power-host") || root;
        openPowerInventory(host, {
          bag: seatBag(me),
          selfId: myId(),
          targets: players().map((p) => ({ id: p.userId, username: p.username, isSelf: p.userId === myId() })),
          onUse: (kind, targetId) => {
            opts.client.send(WS_EVENTS.POWER_USE, { item: kind, targetId });
          },
          onClose: () => undefined,
        });
      },
    });

    const chatForm = root.querySelector<HTMLFormElement>("#board-chat-form");
    chatForm?.addEventListener("submit", (e) => {
      e.preventDefault();
      const input = chatForm.elements.namedItem("msg") as HTMLInputElement | null;
      const text = (input?.value || "").trim();
      if (!text) return;
      opts.client.send(WS_EVENTS.ROOM_CHAT, { text });
      if (input) input.value = "";
    });
  };

  const paintPowerMarks = (): void => {
    const cells = (snap.powerCells || {}) as Record<string, string>;
    root.querySelectorAll<HTMLElement>(".cell").forEach((cell) => {
      const pos = String(cell.dataset.pos || "");
      const k = cells[pos] as PowerKind | undefined;
      let mark = cell.querySelector(".cell-power");
      if (k !== "bomb" && k !== "thunder" && k !== "superman") {
        mark?.remove();
        return;
      }
      if (!mark) {
        mark = document.createElement("span");
        mark.className = "cell-power";
        cell.appendChild(mark);
      }
      mark.textContent = powerIcon(k);
      mark.setAttribute("title", powerLabel(k));
    });
  };

  const paintBoardChat = (): void => {
    const log = root.querySelector("#board-chat-log");
    if (!log) return;
    const chat = Array.isArray(snap.chat) ? (snap.chat as { username?: string; text?: string; emote?: string }[]) : [];
    log.innerHTML = chat.length
      ? chat
          .slice(-40)
          .map((c) => `<p><strong>${esc(c.username || "")}:</strong> ${esc(c.text || c.emote || "")}</p>`)
          .join("")
      : `<p class="power-empty">Belum ada pesan. Sapa lawanmu!</p>`;
    log.scrollTop = log.scrollHeight;
  };

  const place = (userId: string, position: number): void => {
    visual.set(userId, position);
    const tok = tokens.get(userId);
    if (!tok) return;
    const here = [...visual.entries()].filter(([, p]) => p === position).map(([id]) => id);
    const slot = Math.max(0, here.indexOf(userId));
    const off = BoardEngine.tokenOffsets(here.length, slot);
    const pct = tokenBoardPercent(position, off);
    if (!pct) return;
    tok.style.left = `${pct.left}%`;
    tok.style.bottom = `${pct.bottom}%`;
    tok.style.setProperty("--pawn-scale", String(BoardEngine.tokenCrowdScale(here.length)));
    tok.style.zIndex = String(20 + Math.round((0.3 - off.dy) * 40) + slot);
  };

  const syncTokens = (): void => {
    const seats = players();
    for (const s of seats) {
      if (!tokens.has(s.userId)) {
        const el = document.createElement("div");
        el.className = "pawn is-sprite";
        el.style.setProperty("--pawn", s.color);
        const idx = players().findIndex((p) => p.userId === s.userId);
        el.innerHTML = pawnSpriteHtml(idx >= 0 ? idx : 0, s.username);
        el.setAttribute("aria-label", s.username);
        layer.appendChild(el);
        tokens.set(s.userId, el);
      }
      if (!animating) place(s.userId, s.position);
    }
  };

  const paintTable = (): void => {
    ensureTable();
    const cur = players().find((p) => p.userId === currentId());
    const mine = currentId() === myId() && status() === "PLAYING" && !animating && !activeQ;
    const label = activeQ ? "SOAL AKTIF" : mine ? "ROLL DADU" : "MENUNGGU...";
    applyPremiumRollState(rollBtn, label, mine && !activeQ);
    const seats: PremiumSeat[] = players().map((p) => ({
      id: p.userId,
      username: p.username,
      color: p.color,
      position: p.position,
      bag: seatBag(p),
      isConnected: p.isConnected,
    }));
    paintPremiumStrip(playerStrip, seats, currentId());
    const qTimer =
      activeQ?.endsAt && activeQ.playerId === myId()
        ? Math.max(0, Math.ceil((activeQ.endsAt - Date.now()) / 1000))
        : undefined;
    paintPremiumTurnHud(turnHud, cur?.username || "—", {
      position: cur?.position,
      isMine: cur?.userId === myId(),
      timerSec: qTimer,
    });
    const connLine = root.querySelector("#conn-line");
    if (connLine) {
      (connLine as HTMLElement).hidden = false;
      connLine.textContent = `${connChip(conn)} · ${pingTone(pingMs)}`;
    }
    const err = root.querySelector("#board-error");
    if (err) {
      (err as HTMLElement).hidden = !error;
      err.textContent = error;
    }
    const dbg = root.querySelector("#dbg");
    if (dbg) dbg.textContent = JSON.stringify({ seq: snap.seq, matchId: snap.matchId, phase: snap.phase, current: currentId(), q: activeQ?.id }, null, 2);
    syncTokens();
    paintPowerMarks();
    paintBoardChat();
    const frame = root.querySelector("#board-frame");
    frame?.classList.toggle("is-final", Boolean(activeQ?.final) || players().some((p) => p.position >= 100));
    root.querySelectorAll<HTMLElement>(".cell").forEach((c) => {
      const pos = Number(c.dataset.pos);
      c.classList.toggle("is-current", pos === (cur?.position || 0));
    });
    const qHost = root.querySelector("#q-host") as HTMLElement | null;
    if (qHost && activeQ) {
      const overlay = mountQuestionOverlay(qHost, {
        question: activeQ,
        selfId: myId(),
        answering: activeQ.playerId === myId() && !answered && !qResult,
        result: qResult,
        onAnswer: (letter) => {
          if (answered) return;
          answered = true;
          opts.client.send(WS_EVENTS.QUESTION_ANSWER, { answer: letter });
        },
      });
      window.clearInterval(qTick);
      qTick = window.setInterval(() => overlay.tick(), 250);
    } else if (qHost) {
      window.clearInterval(qTick);
      unmountQuestionOverlay(qHost);
    }
    paintResult();
  };

  let champCelebrated = false;

  const paintResult = (): void => {
    const host = root.querySelector("#result-host");
    if (!host) return;
    if (status() !== "FINISHED") {
      host.innerHTML = "";
      champCelebrated = false;
      return;
    }
    // Jika modal juara sudah tampil, panel hasil hanya ringkas (tanpa ulang “SELAMAT”).
    const ranked = [...players()].sort((a, b) => b.position - a.position);
    const winner = ranked.find((p) => p.userId === String(snap.winnerId || "")) || ranked[0];
    const stats = (snap.playerStats || {}) as Record<string, { correctAnswers?: number; wrongAnswers?: number; accuracy?: number }>;
    const rewards = (snap.rewards || []) as RewardEvent[];
    const mine = rewards.find((r) => r.userId === myId());
    const compact = champCelebrated;
    host.innerHTML = `<div class="result-layer"><div class="result-card nt-card ${compact ? "" : "champion-card"}">
      ${compact
        ? `<p class="nt-kicker">${icon("trophy")} HASIL AKHIR</p>
           <h2>${esc(truncateName(winner?.username || "—"))} juara</h2>
           ${mine ? `<p class="pf-coins">✨ +${mine.xp || 0} XP · 🪙 +${mine.coins || 0}${mine.trophies ? ` · 🏆 +${mine.trophies}` : ""}</p>` : ""}`
        : `<p class="nt-kicker">${icon("trophy")} SELAMAT JUARA!</p>
           <div class="champion-cup">🏆</div>
           <span class="seat-pawn" style="--pawn:${winner?.color || "#d4b45a"}"></span>
           <h2>${esc(truncateName(winner?.username || "—"))}</h2>
           <p class="champion-msg">Perayaan meriah untuk sang juara yang mencapai kotak 100!</p>
           ${mine?.won || winner?.userId === myId() ? `<p class="champion-pill">🏆 +${mine?.trophies ?? 1} Piala${mine?.trophyTotal ? ` · Total ${mine.trophyTotal}` : ""}</p>` : ""}
           ${mine ? `<p class="pf-coins">✨ +${mine.xp || 0} XP · 🪙 +${mine.coins || 0}${mine.rrDelta ? " · RR " + (mine.rrDelta > 0 ? "+" : "") + mine.rrDelta : ""}</p>
           ${mine.rankLabel ? `<p>🏅 ${esc(mine.rankLabel)}</p>` : ""}` : ""}`}
      <div class="result-list">${ranked
        .map((p, i) => {
          const st = stats[p.userId] || {};
          const medal = i === 0 ? "🥇" : i === 1 ? "🥈" : i === 2 ? "🥉" : `${i + 1}.`;
          return `<div class="seat"><span>${medal}</span><strong>${esc(truncateName(p.username))}</strong> · Kotak ${p.position} · ${st.correctAnswers || 0} benar · ${st.wrongAnswers || 0} salah · ${st.accuracy || 0}%</div>`;
        })
        .join("")}</div>
      <div class="nt-actions">
        <button type="button" class="nt-btn nt-btn-primary" data-r="again">MAIN LAGI</button>
        <button type="button" class="nt-btn" data-r="room">KEMBALI KE ROOM</button>
        <button type="button" class="nt-btn nt-btn-ghost" data-r="menu">MENU UTAMA</button>
      </div>
    </div></div>`;
    host.querySelector("[data-r=again]")?.addEventListener("click", () => {
      opts.client.send(WS_EVENTS.ROOM_LEAVE);
      tableBuilt = false;
      view = "room";
      snap = {};
      champCelebrated = false;
      paint();
    });
    host.querySelector("[data-r=room]")?.addEventListener("click", () => {
      tableBuilt = false;
      view = "room";
      champCelebrated = false;
      paint();
    });
    host.querySelector("[data-r=menu]")?.addEventListener("click", onExit);
  };

  opts.client.onStatus = (s) => {
    conn = s;
    if (s === "offline") setConnectionBanner("offline");
    else if (s === "connecting") setConnectionBanner("connecting");
    else setConnectionBanner("online");
    if (view === "room") paint();
    else paintTable();
  };
  unsub = opts.client.addListener((type, data) => {
    const skipMerge =
      type === WS_EVENTS.PONG ||
      type === WS_EVENTS.QUEUE_UPDATE ||
      type === WS_EVENTS.MATCH_FOUND ||
      type === WS_EVENTS.GAME_INVITE ||
      type === WS_EVENTS.SOCIAL_NOTIFY ||
      type === WS_EVENTS.ONLINE_LIST;
    const seatsLike = Array.isArray(data.players) && data.players[0] && typeof data.players[0] === "object";
    if (!skipMerge && (data.roomCode || seatsLike || data.status || data.hostId || data.chat)) {
      const next = { ...snap, ...data };
      if (Array.isArray(data.players) && !seatsLike) delete next.players;
      snap = next;
    }
    if (type === WS_EVENTS.ONLINE_LIST) {
      const list = Array.isArray(data.players) ? (data.players as OnlineCard[]) : [];
      onlinePlayers = list.filter((p) => p && p.userId && p.userId !== myId());
      if (view === "room") paint();
      return;
    }
    if (type === WS_EVENTS.PONG) {
      const t = Number(data.t || 0);
      if (t) pingMs = Math.max(1, Date.now() - t);
      return;
    }
    if (type === WS_EVENTS.QUEUE_UPDATE) {
      searching = Boolean(data.searching);
      if (data.mode) searchMode = String(data.mode);
      if (searching && !searchAt) searchAt = Date.now();
      if (!searching) {
        matchFound = null;
        readyCheck = false;
      }
    }
    if (type === WS_EVENTS.MATCH_FOUND) {
      searching = false;
      readyCheck = false;
      matchFound = { ends: Date.now() + Number(data.countdown || 5) * 1000, mode: String(data.mode || searchMode) };
      toast("MATCH FOUND!", "success");
    }
    if (type === WS_EVENTS.GAME_INVITE) {
      invite = { inviteId: String(data.inviteId || ""), username: String(data.username || "Pemain"), expiresAt: Number(data.expiresAt || 0) };
      window.setTimeout(() => {
        if (invite && invite.inviteId === String(data.inviteId || "")) {
          invite = null;
          if (view === "room") paint();
        }
      }, 30000);
    }
    if (type === WS_EVENTS.SOCIAL_NOTIFY) toast(String(data.title || data.type || "Notifikasi"), "info");
    if (type === WS_EVENTS.ULAR_ERROR) {
      error = String(data.message || "Terjadi kesalahan. Silakan coba lagi.");
      toast(error, "error");
      if (String(data.code || "") === "KICKED") {
        snap = {};
        searching = false;
        matchFound = null;
        readyCheck = false;
      }
      if (view === "room") paint();
      else paintTable();
      return;
    }
    if (type === WS_EVENTS.ROOM_START || status() === "STARTING" || status() === "PLAYING" || status() === "FINISHED") {
      view = "table";
      searching = false;
      matchFound = null;
      readyCheck = false;
    }
    if (type === WS_EVENTS.DICE_RESULT) {
      lastDice = Number(data.dice || 0);
      const res = root.querySelector("#dice-result");
      if (res) res.textContent = `🎲 ${lastDice} · ${lastDice} langkah`;
      if (diceEl) void anim.animateDice(diceEl, lastDice);
    }
    if (type === WS_EVENTS.QUESTION_START) {
      playSfx("question_open");
      const q = (data.question || data) as QuestionPublic;
      activeQ = q.id ? q : ((data as { question?: QuestionPublic }).question as QuestionPublic);
      if (!activeQ?.id && snap.question) activeQ = snap.question as QuestionPublic;
      qResult = null;
      answered = false;
    }
    if (type === WS_EVENTS.QUESTION_RESULT || type === WS_EVENTS.QUESTION_TIMEOUT || type === WS_EVENTS.QUESTION_PENALTY) {
      qResult = data as QuestionResultView;
      const frame = root.querySelector("#board-frame") as HTMLElement | null;
      if (qResult.result === "CORRECT") {
        playSfx("correct");
        if (frame) {
          sparkle(frame);
          spawnFloat(frame, "BENAR!", "good");
        }
      } else if (qResult.timeout || type === WS_EVENTS.QUESTION_TIMEOUT) {
        playSfx("timeout");
        if (frame) spawnFloat(frame, "Waktu Habis! -10", "bad");
      } else {
        playSfx("wrong");
        if (frame) spawnFloat(frame, "Belum tepat. -10", "bad");
      }
      const reward = qResult.reward;
      if (type !== WS_EVENTS.QUESTION_PENALTY && String(data.userId || "") === myId()) applyRewardFx(frame, reward);
      const uid = String(data.userId || "");
      const path = Array.isArray(data.path) ? (data.path as number[]) : [];
      const tok = tokens.get(uid);
      if (tok && path.length > 1 && !qResult.correct) {
        animating = true;
        void (async () => {
          await anim.animatePenalty(tok, path, (pos) => place(uid, pos));
          animating = false;
          paintTable();
        })();
      }
    }
    if (type === WS_EVENTS.POWER_PICKUP) {
      const item = String(data.item || "") as PowerKind;
      const frame = root.querySelector("#board-frame") as HTMLElement | null;
      if (item === "bomb" || item === "thunder" || item === "superman") {
        toast(String(data.message || powerGrantBanner(item)), "info");
        if (frame) spawnFloat(frame, powerGrantBanner(item), "good");
      }
      if (data.powerCells) snap = { ...snap, powerCells: data.powerCells };
      paintPowerMarks();
    }
    if (type === WS_EVENTS.ROOM_CHAT || type === WS_EVENTS.ROOM_EMOTE) {
      const line = {
        userId: String(data.userId || ""),
        username: String(data.username || ""),
        text: String(data.text || ""),
        emote: String(data.emote || ""),
        at: Number(data.at || Date.now()),
      };
      const prev = Array.isArray(snap.chat) ? [...(snap.chat as object[])] : [];
      prev.push(line);
      snap = { ...snap, chat: prev.slice(-40) };
      paintBoardChat();
      const panel = root.querySelector<HTMLElement>("#board-chat");
      if (panel && panel.hidden && line.userId !== myId() && (line.text || line.emote)) {
        toast(`${line.username}: ${line.text || line.emote}`.slice(0, 80), "info");
      }
    }
    if (type === WS_EVENTS.POWER_USED) {
      const frame = root.querySelector("#board-frame") as HTMLElement | null;
      const tid = String(data.targetId || "");
      const toPos = Number(data.toPos ?? 0);
      const fromPos = Number(data.fromPos ?? toPos);
      const msg = String(data.message || "Item dipakai");
      toast(msg, "info");
      if (frame) spawnFloat(frame, msg.slice(0, 42), String(data.item) === "superman" ? "good" : "bad");
      if (data.powerCells) snap = { ...snap, powerCells: data.powerCells };
      const tok = tokens.get(tid);
      if (tok) {
        animating = true;
        void (async () => {
          await anim.animateTokenMove(tok, fromPos === toPos ? [] : [toPos], (pos) => place(tid, pos));
          place(tid, toPos);
          animating = false;
          paintTable();
        })();
      }
    }
    if (type === WS_EVENTS.QUESTION_COMPLETE || type === WS_EVENTS.TURN_CHANGED) {
      activeQ = null;
      qResult = null;
      answered = false;
    }
    if ((type === WS_EVENTS.GAME_STATE || type === WS_EVENTS.PLAYER_RECONNECT) && snap.question && !qResult) {
      activeQ = snap.question as QuestionPublic;
    }
    if (type === WS_EVENTS.PLAYER_MOVING) {
      const uid = String(data.userId || "");
      const move = (data.move || {}) as { walkPath?: number[]; walkFinal?: number; snakeTo?: number; ladderTo?: number; final?: number };
      const tok = tokens.get(uid);
      const path = move.walkPath || [];
      if (tok && path.length) {
        animating = true;
        void (async () => {
          await anim.animateTokenMove(tok, path, (pos) => place(uid, pos));
          if (move.snakeTo) await anim.animateSnake(root.querySelector("#board-frame") as HTMLElement, tok, move.walkFinal || path[path.length - 1], move.snakeTo, (pos) => place(uid, pos));
          else if (move.ladderTo) {
            await anim.animateLadder(tok, move.ladderTo, (pos) => place(uid, pos));
            const frame = root.querySelector("#board-frame") as HTMLElement | null;
            if (frame) sparkle(frame);
          }
          place(uid, move.final || move.walkFinal || path[path.length - 1]);
          animating = false;
          if (uid === myId()) opts.client.send(WS_EVENTS.MOVEMENT_ANIMATION_COMPLETE);
          paintTable();
        })();
      }
    }
    if (type === WS_EVENTS.GAME_FINISHED) {
      const frame = root.querySelector("#board-frame") as HTMLElement | null;
      if (frame) confetti(frame, true);
      const rewards = ((data.rewards || snap.rewards || []) as RewardEvent[]);
      const ranked = (Array.isArray(data.players) ? data.players : snap.players || []) as { userId?: string; username?: string; rank?: number }[];
      const mine = rewards.find((r) => r.userId === myId());
      applyRewardFx(frame, mine);
      const winnerId = String(data.winnerId || snap.winnerId || "");
      const myRank = Number(ranked.find((p) => p.userId === myId())?.rank || (myId() === winnerId ? 1 : 0));
      if (!champCelebrated && (myRank === 1 || (myRank <= 0 && winnerId))) {
        champCelebrated = true;
        if (myRank === 1) {
          showMatchResultModal({
            username: profileName(),
            rank: 1,
            xp: mine?.xp || 0,
            coins: mine?.coins || 0,
            trophies: mine?.trophies ?? 1,
            trophyTotal: mine?.trophyTotal,
            rrDelta: mine?.rrDelta,
            mode: roomMode() || "CASUAL",
          });
        } else {
          const wName = ranked.find((p) => p.userId === winnerId)?.username || "Juara";
          showChampionCelebration({
            username: wName,
            isSelf: false,
            trophiesAwarded: 1,
            subtitle: "Pertandingan selesai",
          });
        }
      } else if (!champCelebrated && mine && myRank > 1) {
        champCelebrated = true;
        showMatchResultModal({
          username: profileName(),
          rank: myRank,
          xp: mine.xp || 0,
          coins: mine.coins || 0,
          rrDelta: mine.rrDelta,
          mode: roomMode() || "CASUAL",
        });
      }
    }
    if (type === WS_EVENTS.PLAYER_RECONNECT || type === WS_EVENTS.AUTH_OK) {
      const tok = storedSessionToken();
      if (tok) void fetchProfile(tok);
      if (type === WS_EVENTS.AUTH_OK && bootQueue) {
        const q = bootQueue;
        bootQueue = "";
        joinQueue(q);
      }
    }
    if (view === "table") paintTable();
    else paint();
  });

  opts.client.send(WS_EVENTS.ONLINE_LIST_REQ);
  if (opts.client.status === "online" && bootQueue) {
    const q = bootQueue;
    bootQueue = "";
    joinQueue(q);
  }
  paint();
}

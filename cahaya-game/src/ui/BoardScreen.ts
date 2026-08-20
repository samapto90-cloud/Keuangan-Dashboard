import {
  DEFAULT_BOARD,
  PLAYER_COLORS,
  BOARD_GRID,
  MOVE_DURATION,
  type BoardConfig,
} from "../game/board/config";
import {
  BoardEngine,
  OFFBOARD_START,
  resolveMove,
  rollDiceLocal,
  tokenBoardPercent,
  type MoveResult,
} from "../game/board/engine";
import { BoardAnimationManager } from "../game/board/animation";
import { playSfx } from "../audio/manager";
import { mountQuestionOverlay, unmountQuestionOverlay, type QuestionPublic, type QuestionResultView } from "./QuestionModal";
import { renderBoardRoutes } from "./boardDecor";
import { zoneClass, zoneDecor } from "./boardZones";
import { pawnSpriteHtml, playerCharacter } from "../assets/registry";
import { confetti, sparkle, spawnFloat, toast, showChampionCelebration } from "./chrome";
import {
  applyPremiumRollState,
  activateBoardTheme,
  bindPremiumDock,
  paintPremiumStrip,
  paintPremiumTurnHud,
  premiumBoardShell,
  type PremiumSeat,
} from "./boardPremium";
import {
  addPower,
  applyBomb,
  applySuperman,
  applyThunder,
  emptyBag,
  powerIcon,
  powerLabel,
  spawnPowerCells,
  takePower,
  type PowerBag,
  type PowerKind,
} from "../game/powers";
import { openPowerInventory, powerGrantBanner } from "./PowerModal";

export type BoardPlayer = {
  id: string;
  userId: string;
  username: string;
  avatar: string;
  color: string;
  position: number;
  bag: PowerBag;
  isReady: boolean;
  isConnected: boolean;
  isNpc?: boolean;
};

type Flow = "WAITING" | "READY" | "ROLLING" | "MOVING" | "SNAKE" | "LADDER" | "NEXT_TURN" | "FINISHED";

const NPC_NAMES = ["Ganjar", "Anies", "Sri Mulyani"];

function escapeHtml(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

async function serverResolve(position: number, player: string): Promise<MoveResult | null> {
  try {
    const res = await fetch("/cahaya/api/ular/resolve", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ position, player }),
    });
    if (!res.ok) return null;
    const data = (await res.json()) as MoveResult;
    if (!data.walkPath || !data.dice) return null;
    return data;
  } catch {
    return null;
  }
}

export function mountBoard(
  root: HTMLElement,
  opts: { names: string[]; withNpc?: boolean; grade?: "SD" | "SMA"; onExit: () => void },
): void {
  const cfg: BoardConfig = { ...DEFAULT_BOARD, snakes: { ...DEFAULT_BOARD.snakes }, ladders: { ...DEFAULT_BOARD.ladders } };
  const anim = new BoardAnimationManager();
  const withNpc = Boolean(opts.withNpc);
  const eduGrade = opts.grade === "SD" ? "SD" : "SMA";
  const players: BoardPlayer[] = opts.names.slice(0, 4).map((name, i) => {
    const ch = playerCharacter(i);
    return {
      id: `p${i + 1}`,
      userId: `p${i + 1}`,
      username: withNpc && i > 0 ? NPC_NAMES[i - 1] || ch.name : name || ch.name,
      avatar: ch.id,
      color: ch.color || PLAYER_COLORS[i],
      position: OFFBOARD_START,
      bag: emptyBag(),
      isReady: true,
      isConnected: true,
      isNpc: withNpc && i > 0,
    };
  });
  let turn = 0;
  let flow: Flow = "READY";
  let lastDice = 0;
  let lastLog = "";
  let error = "";
  let hover = 0;
  let movingIndex = -1;
  let npcTimer = 0;
  const visualPos = players.map((p) => p.position);
  const logs: string[] = [];
  const dev = import.meta.env.DEV;

  root.innerHTML = "";
  root.classList.add("board-root");
  const deactivateBoardTheme = activateBoardTheme();
  const prevExit = opts.onExit;
  const onExit = (): void => {
    window.clearTimeout(npcTimer);
    deactivateBoardTheme();
    prevExit();
  };

  const app = document.createElement("div");
  app.className = "board-app";
  app.innerHTML = premiumBoardShell(`
    ${dev ? `<details class="debug-panel premium-debug" open><summary>Debug</summary><pre id="debug-pre"></pre><div id="dev-log"></div></details>` : ""}
    <div id="q-host"></div>
    <div id="result-host"></div>
  `);
  root.appendChild(app);

  const grid = app.querySelector<HTMLElement>("#board-grid")!;
  const layer = app.querySelector<HTMLElement>("#token-layer")!;
  const frame = app.querySelector<HTMLElement>("#board-frame")!;
  const playerStrip = app.querySelector<HTMLElement>("#player-strip")!;
  const turnHud = app.querySelector<HTMLElement>("#turn-hud")!;
  const diceEl = app.querySelector<HTMLElement>("#dice")!;
  const rollBtn = app.querySelector<HTMLElement>("#roll-btn")!;
  const diceResult = app.querySelector<HTMLElement>("#dice-result")!;
  const errEl = app.querySelector<HTMLElement>("#board-error")!;
  const debugPre = app.querySelector<HTMLElement>("#debug-pre");
  const devLog = app.querySelector<HTMLElement>("#dev-log");
  const qHost = app.querySelector<HTMLElement>("#q-host")!;
  let usedIds: string[] = [];
  let powerCells: Record<number, PowerKind> = spawnPowerCells(cfg.snakes, cfg.ladders);

  const paintPowerMarks = (): void => {
    for (const cell of cells) {
      const pos = Number(cell.dataset.pos);
      const kind = powerCells[pos];
      let mark = cell.querySelector(".cell-power");
      if (!kind) {
        mark?.remove();
        continue;
      }
      if (!mark) {
        mark = document.createElement("span");
        mark.className = "cell-power";
        cell.appendChild(mark);
      }
      mark.textContent = powerIcon(kind);
      mark.setAttribute("title", powerLabel(kind));
    }
  };

  const tryPickup = (idx: number, pos: number): void => {
    const kind = powerCells[pos];
    if (!kind) return;
    delete powerCells[pos];
    players[idx]!.bag = addPower(players[idx]!.bag, kind);
    spawnFloat(frame, `${powerGrantBanner(kind)} (kotak ${pos})`, "good");
    playSfx("correct");
    paintPowerMarks();
    renderChrome();
  };

  const cells: HTMLElement[] = [];
  for (let visualRow = 0; visualRow < BOARD_GRID; visualRow++) {
    const boardRow = BOARD_GRID - 1 - visualRow;
    for (let col = 0; col < BOARD_GRID; col++) {
      const pos = BoardEngine.getPositionFromCoordinate(boardRow, col)!;
      const cell = document.createElement("button");
      cell.type = "button";
      const tile = pos % 4;
      cell.className = `cell tile-${tile} color-${tile === 0 ? 1 : tile + 1} ${zoneClass(pos)}`;
      cell.dataset.pos = String(pos);
      cell.setAttribute("aria-label", `Kotak ${pos}`);
      if (pos === 1) cell.classList.add("cell-enter");
      if (pos === 100) cell.classList.add("cell-finish");
      if (cfg.snakes[pos]) cell.classList.add("cell-snake");
      if (cfg.ladders[pos]) cell.classList.add("cell-ladder");
      cell.innerHTML = `<span class="cell-num">${pos}</span>${zoneDecor(pos)}${pos === 1 ? `<span class="cell-tag">MASUK</span>` : ""}${pos === 100 ? `<span class="cell-tag">FINISH</span>` : ""}`;
      cell.addEventListener("mouseenter", () => {
        hover = pos;
        cell.classList.add("is-hover");
      });
      cell.addEventListener("mouseleave", () => {
        hover = 0;
        cell.classList.remove("is-hover");
      });
      grid.appendChild(cell);
      cells.push(cell);
    }
  }
  const routes = app.querySelector<SVGSVGElement>("#route-layer");
  if (routes) renderBoardRoutes(routes, cfg.snakes, cfg.ladders);

  const tokens = players.map((p, i) => {
    const el = document.createElement("div");
    el.className = "pawn is-sprite";
    el.dataset.id = p.id;
    el.style.setProperty("--pawn", p.color);
    el.innerHTML = pawnSpriteHtml(i, p.username);
    el.setAttribute("aria-label", p.username);
    layer.appendChild(el);
    void i;
    return el;
  });

  const occupants = (pos: number): number[] =>
    visualPos.map((p, i) => (p === pos ? i : -1)).filter((i) => i >= 0);

  const placeToken = (index: number, position: number): void => {
    visualPos[index] = position;
    const here = occupants(position);
    const count = Math.max(here.length, 1);
    const slot = Math.max(0, here.indexOf(index));
    const off = BoardEngine.tokenOffsets(count, slot);
    const pct = tokenBoardPercent(position, off);
    if (!pct) return;
    const tok = tokens[index];
    tok.style.left = `${pct.left}%`;
    tok.style.bottom = `${pct.bottom}%`;
    tok.style.setProperty("--pawn-scale", String(BoardEngine.tokenCrowdScale(count)));
    tok.style.zIndex = String(20 + Math.round((0.3 - off.dy) * 40) + slot);
  };

  const paintHighlights = (currentPos: number, dest?: number): void => {
    for (const cell of cells) {
      const pos = Number(cell.dataset.pos);
      cell.classList.toggle("is-current", pos === currentPos);
      cell.classList.toggle("is-dest", dest === pos);
    }
  };

  const applyPowerLocal = async (actorIdx: number, kind: PowerKind, targetId: string): Promise<void> => {
    const actor = players[actorIdx];
    const nextBag = takePower(actor.bag, kind);
    if (!nextBag) {
      toast("Item tidak tersedia.", "warning");
      return;
    }
    actor.bag = nextBag;
    const tIdx = players.findIndex((p) => p.id === targetId || p.userId === targetId);
    if (tIdx < 0) return;
    const target = players[tIdx]!;
    const from = target.position;
    let to = from;
    if (kind === "bomb") to = applyBomb(from);
    else if (kind === "thunder") to = applyThunder(from);
    else to = applySuperman(from);
    target.position = to;
    visualPos[tIdx] = to;
    const label =
      kind === "bomb"
        ? `💣 ${actor.username} → ${target.username} START!`
        : kind === "thunder"
          ? `⚡ ${target.username} -3`
          : `✈️ ${target.username} +3`;
    spawnFloat(frame, label, kind === "superman" ? "good" : "bad");
    playSfx(kind === "superman" ? "correct" : "wrong");
    await anim.animateTokenMove(tokens[tIdx]!, from === to ? [] : [to], (pos) => {
      placeToken(tIdx, pos);
      relayoutTokens();
    });
    placeToken(tIdx, to);
    relayoutTokens();
    renderChrome();
  };

  const maybeNpcUsePower = async (idx: number): Promise<void> => {
    const npc = players[idx];
    if (!npc?.isNpc) return;
    const human = players.find((p) => !p.isNpc);
    if (!human) return;
    let kind: PowerKind | null = null;
    if (npc.bag.bomb > 0) kind = "bomb";
    else if (npc.bag.thunder > 0) kind = "thunder";
    else if (npc.bag.superman > 0 && Math.random() < 0.5) kind = "superman";
    if (!kind) return;
    const targetId = kind === "superman" && Math.random() < 0.35 ? npc.id : human.id;
    if (kind !== "superman" && targetId === npc.id) return;
    await applyPowerLocal(idx, kind, targetId);
  };

  const openInventoryFor = (idx: number): void => {
    if (flow !== "READY" || players[turn]?.isNpc || idx !== turn) {
      toast("Inventory hanya saat giliranmu.", "info");
      return;
    }
    const host = app.querySelector<HTMLElement>("#power-host") || app;
    openPowerInventory(host, {
      bag: players[idx]!.bag,
      selfId: players[idx]!.id,
      targets: players.map((p) => ({ id: p.id, username: p.isNpc ? `🤖 ${p.username}` : p.username, isSelf: p.id === players[idx]!.id })),
      onUse: (kind, targetId) => {
        void applyPowerLocal(idx, kind, targetId);
      },
      onClose: () => undefined,
    });
  };

  let runTurn: () => Promise<void> = async () => undefined;

  const scheduleNpcIfNeeded = (): void => {
    window.clearTimeout(npcTimer);
    if (flow !== "READY") return;
    const cur = players[turn];
    if (!cur?.isNpc) return;
    npcTimer = window.setTimeout(() => {
      if (flow !== "READY" || !players[turn]?.isNpc) return;
      void (async () => {
        await maybeNpcUsePower(turn);
        if (flow === "READY" && players[turn]?.isNpc) void runTurn();
      })();
    }, 850 + Math.floor(Math.random() * 500));
  };

  const renderChrome = (): void => {
    const cur = players[turn];
    const humanTurn = !cur.isNpc;
    const locked = flow !== "READY" || cur.position >= 100 || !humanTurn;
    const label = flow === "FINISHED" ? "SELESAI" : cur.isNpc ? "GILIRAN NPC..." : locked ? "MENUNGGU..." : "ROLL DADU";
    applyPremiumRollState(rollBtn, label, !locked);
    const seats: PremiumSeat[] = players.map((p) => ({
      id: p.id,
      username: p.isNpc ? `🤖 ${p.username}` : p.username,
      color: p.color,
      position: p.position,
      bag: p.bag,
    }));
    paintPremiumStrip(playerStrip, seats, cur.id);
    paintPremiumTurnHud(turnHud, cur.isNpc ? `🤖 ${cur.username}` : cur.username, {
      position: cur.position,
      isMine: humanTurn,
    });
    errEl.hidden = !error;
    errEl.textContent = error;
    if (debugPre) {
      debugPre.textContent = JSON.stringify(
        {
          currentPlayer: cur.username,
          isNpc: Boolean(cur.isNpc),
          position: cur.position,
          dice: lastDice,
          gameState: flow,
          hover,
          movingIndex,
          moveMs: MOVE_DURATION,
          snake: cfg.snakes[cur.position] || null,
          ladder: cfg.ladders[cur.position] || null,
        },
        null,
        2,
      );
    }
    if (devLog) devLog.textContent = logs.slice(-8).join("\n");
    paintHighlights(cur.position);
    scheduleNpcIfNeeded();
  };

  const relayoutTokens = (): void => {
    visualPos.forEach((pos, i) => placeToken(i, pos));
  };

  const runQuestion = async (idx: number, final: boolean): Promise<"win" | "continue"> => {
    const player = players[idx];
    playSfx("question_open");
    let q: QuestionPublic | null = null;
    let practiceId = "";
    try {
      const res = await fetch("/cahaya/api/ular/practice/question", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ final, used: usedIds, grade: eduGrade }),
      });
      if (res.ok) {
        const data = (await res.json()) as { practiceId: string; question: QuestionPublic; used?: string[] };
        practiceId = data.practiceId;
        q = data.question;
        if (data.used) usedIds = data.used;
      }
    } catch {
      q = null;
    }
    if (!q) return final ? "win" : "continue";
    const box: { current: QuestionResultView | null } = { current: null };

    const submitAnswer = async (letter: string): Promise<void> => {
      const ans = await fetch("/cahaya/api/ular/practice/answer", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ practiceId, answer: letter, position: player.position }),
      });
      box.current = ans.ok ? ((await ans.json()) as QuestionResultView) : { result: "WRONG", correct: false, timeout: false };
    };

    if (player.isNpc) {
      const overlay = mountQuestionOverlay(qHost, {
        question: q,
        selfId: player.id,
        answering: false,
      });
      diceResult.textContent = `${player.username} menjawab soal...`;
      await new Promise((r) => window.setTimeout(r, 900 + Math.floor(Math.random() * 700)));
      const letters = ["A", "B", "C", "D"];
      await submitAnswer(letters[Math.floor(Math.random() * letters.length)]!);
      mountQuestionOverlay(qHost, { question: q, selfId: player.id, answering: false, result: box.current });
      void overlay;
    } else {
      await new Promise<void>((resolve) => {
        const overlay = mountQuestionOverlay(qHost, {
          question: q!,
          selfId: player.id,
          answering: true,
          onAnswer: (letter) => {
            void (async () => {
              await submitAnswer(letter);
              mountQuestionOverlay(qHost, { question: q!, selfId: player.id, answering: false, result: box.current });
              resolve();
            })();
          },
        });
        const iv = window.setInterval(() => overlay.tick(), 250);
        window.setTimeout(() => {
          if (box.current) return;
          void (async () => {
            await submitAnswer("");
            if (!box.current) box.current = { result: "TIMEOUT", correct: false, timeout: true };
            mountQuestionOverlay(qHost, { question: q!, selfId: player.id, answering: false, result: box.current });
            window.clearInterval(iv);
            resolve();
          })();
        }, (q.timeLimit || 15) * 1000 + 50);
      });
    }
    const settled = box.current;
    if (settled?.correct) {
      playSfx("correct");
      sparkle(frame);
      spawnFloat(frame, "BENAR!", "good");
    } else if (settled?.timeout) {
      playSfx("timeout");
      spawnFloat(frame, "Waktu Habis! -10", "bad");
    } else {
      playSfx("wrong");
      spawnFloat(frame, "Belum tepat. -10", "bad");
    }
    if (settled && !settled.correct) {
      const path = settled.path || [];
      await anim.animatePenalty(tokens[idx], path, (pos) => {
        placeToken(idx, pos);
        relayoutTokens();
      });
      player.position = settled.positionAfterPenalty || 1;
      visualPos[idx] = player.position;
    }
    await new Promise((r) => window.setTimeout(r, 700));
    unmountQuestionOverlay(qHost);
    if (settled?.won) return "win";
    return "continue";
  };

  runTurn = async (): Promise<void> => {
    if (flow !== "READY") return;
    const idx = turn;
    const player = players[idx];
    error = "";
    flow = "ROLLING";
    applyPremiumRollState(rollBtn, "Lempar dadu", false);
    renderChrome();
    let result = await serverResolve(player.position, player.username);
    if (!result) {
      const dice = rollDiceLocal();
      result = resolveMove(cfg, player.position, dice);
      result.log = `${player.username} ${result.log} (local-fallback)`;
    }
    if (!result.dice || result.dice < 1) {
      error = "Terjadi kesalahan. Silakan coba lagi.";
      flow = "READY";
      renderChrome();
      return;
    }
    lastDice = result.dice;
    diceResult.textContent = `${player.isNpc ? "🤖 " : "🎲 "}${player.username}: ${result.dice} · ${result.dice} langkah`;
    await anim.animateDice(diceEl, result.dice);
    flow = "MOVING";
    movingIndex = idx;
    renderChrome();
    const moving = tokens[idx];
    await anim.animateTokenMove(moving, result.walkPath, (pos) => {
      placeToken(idx, pos);
      relayoutTokens();
    });
    placeToken(idx, result.walkFinal);
    relayoutTokens();
    if (result.snakeTo) {
      flow = "SNAKE";
      renderChrome();
      await anim.animateSnake(frame, moving, result.walkFinal, result.snakeTo, (pos) => {
        placeToken(idx, pos);
        relayoutTokens();
      });
    } else if (result.ladderTo) {
      flow = "LADDER";
      renderChrome();
      await anim.animateLadder(moving, result.ladderTo, (pos) => {
        placeToken(idx, pos);
        relayoutTokens();
      });
    }
    player.position = result.final;
    visualPos[idx] = result.final;
    movingIndex = -1;
    relayoutTokens();
    tryPickup(idx, result.final);
    lastLog = `${player.username} ${result.log}`;
    logs.push(lastLog);
    const asked = await runQuestion(idx, Boolean(result.reached100));
    if (asked === "win") {
      flow = "FINISHED";
      const isHuman = !player.isNpc;
      const trophyKey = "ular-local-trophies";
      let trophyTotal = Number(localStorage.getItem(trophyKey) || "0") || 0;
      if (isHuman) {
        trophyTotal += 1;
        localStorage.setItem(trophyKey, String(trophyTotal));
      }
      confetti(frame, true);
      showChampionCelebration({
        username: player.username,
        isSelf: isHuman,
        trophiesAwarded: isHuman ? 1 : 0,
        trophiesTotal: isHuman ? trophyTotal : undefined,
        subtitle: player.isNpc ? "NPC menang — coba lagi!" : "Main lokal · +1 piala",
        onClose: onExit,
      });
      renderChrome();
      return;
    }
    flow = "NEXT_TURN";
    await anim.animateTurnTransition(turnHud);
    turn = BoardEngine.nextTurnIndex(idx, players.length);
    flow = "READY";
    renderChrome();
  };

  bindPremiumDock(app, {
    onRoll: () => {
      if (players[turn]?.isNpc) return;
      void runTurn();
    },
    onExit,
    onHistory: () => toast("Item di papan: 💣×1 · ⚡×5 · ✈️×3 — ambil dengan mendarat di kotaknya.", "info"),
    onInventory: () => openInventoryFor(turn),
    onChat: () => {
      const panel = app.querySelector<HTMLElement>("#board-chat");
      if (!panel) return;
      panel.hidden = !panel.hidden;
      const log = app.querySelector("#board-chat-log");
      if (log && !log.childElementCount) {
        log.innerHTML = `<p class="power-empty">Chat lokal: tulis pesan untuk catatan giliran (online: chat antar pemain).</p>`;
      }
    },
  });

  app.querySelector("#board-chat-form")?.addEventListener("submit", (e) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const input = form.elements.namedItem("msg") as HTMLInputElement | null;
    const text = (input?.value || "").trim();
    if (!text) return;
    const log = app.querySelector("#board-chat-log");
    const me = players.find((p) => !p.isNpc) || players[0];
    if (log) {
      const p = document.createElement("p");
      p.innerHTML = `<strong>${escapeHtml(me?.username || "Saya")}:</strong> ${escapeHtml(text)}`;
      log.appendChild(p);
      log.scrollTop = log.scrollHeight;
    }
    if (input) input.value = "";
  });

  paintPowerMarks();
  void fetch("/cahaya/api/ular/board")
    .then((r) => (r.ok ? r.json() : null))
    .then((remote: BoardConfig | null) => {
      if (remote?.snakes && remote.ladders) {
        cfg.snakes = remote.snakes;
        cfg.ladders = remote.ladders;
        powerCells = spawnPowerCells(cfg.snakes, cfg.ladders);
        for (const cell of cells) {
          const pos = Number(cell.dataset.pos);
          cell.classList.toggle("cell-snake", Boolean(cfg.snakes[pos]));
          cell.classList.toggle("cell-ladder", Boolean(cfg.ladders[pos]));
        }
        const rt = app.querySelector<SVGSVGElement>("#route-layer");
        if (rt) renderBoardRoutes(rt, cfg.snakes, cfg.ladders);
        paintPowerMarks();
      }
    })
    .catch(() => undefined);

  relayoutTokens();
  renderChrome();
  if (import.meta.env.DEV) {
    void import("../game/board/engine.test").then((m) => m.runBoardUnitTests());
  }
}

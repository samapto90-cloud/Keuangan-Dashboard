import "./style.css";
import { Game } from "./game/Game";
import { TitleScreen, storedSessionToken } from "./ui/TitleScreen";

const canvas = document.querySelector<HTMLCanvasElement>("#game-canvas");
const loading = document.querySelector<HTMLElement>("#loading-screen");
const fill = document.querySelector<HTMLElement>("#loading-fill");
const tip = document.querySelector<HTMLElement>("#loading-text");
const hud = document.querySelector<HTMLElement>("#hud");
const root = document.querySelector<HTMLElement>("#game-root");

if (!canvas || !hud || !root) {
  throw new Error("Elemen game tidak lengkap");
}

const gameCanvas = canvas;
const gameHud = hud;
const gameRoot = root;

const tips = [
  "Masuk dengan akunmu. Progress tersimpan di server, bukan hanya di perangkat.",
  "NPC desa berbicara dalam Bahasa Jawa. Terjemahan bisa dinyalakan di Pengaturan.",
  "Gold, item, dan damage selalu dihitung server.",
];

function setProgress(n: number, text?: string): void {
  if (fill) fill.style.width = `${Math.max(0, Math.min(100, n))}%`;
  if (text && tip) tip.textContent = text;
}

function yieldFrame(): Promise<void> {
  return new Promise((resolve) => {
    requestAnimationFrame(() => resolve());
  });
}

let game: Game | null = null;
const title = new TitleScreen(gameRoot);

const begin = (): void => {
  if (!game) return;
  title.close();
  gameHud.removeAttribute("hidden");
  game.start();
};

title.onEnter = (session) => {
  if (!game) return;
  game.useSession(session.token, session.username);
  begin();
};
title.onSettings = () => {
  if (!game) return;
  const token = storedSessionToken();
  if (!token) return;
  game.useSession(token, window.localStorage.getItem("cahaya-username") || "");
  begin();
  const panel = gameHud.querySelector<HTMLElement>("#hud-settings-panel");
  if (panel) panel.hidden = false;
};

window.addEventListener("error", (ev) => {
  const msg = ev.message || "Kesalahan klien";
  if (game) game.hud.toast(`ERROR\n${msg}`);
  else setProgress(100, `Gagal memuat: ${msg}`);
});

async function boot(): Promise<void> {
  setProgress(12, "Memuat dunia petualangan...");
  await yieldFrame();
  setProgress(28, "Menyiapkan grafis...");
  await yieldFrame();
  try {
    game = new Game(gameCanvas, gameHud);
  } catch (err) {
    const msg = err instanceof Error ? err.message : "Kesalahan klien";
    setProgress(100, `Gagal memuat: ${msg}`);
    return;
  }
  if (game.state.embedMode) document.body.classList.add("embed");
  setProgress(72, tips[0]);
  await yieldFrame();
  setProgress(100, tips[1]);
  loading?.setAttribute("hidden", "");
  game.state.phase = "title";
  title.open();
}

void boot();

window.addEventListener("beforeunload", () => game?.dispose());

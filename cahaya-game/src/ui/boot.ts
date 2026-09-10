import { storedSessionToken } from "../auth/session";
import { PLAYER_CHARACTERS } from "../assets/registry";
import { initPWA } from "./pwa";
import { loadPrefs } from "./prefs";
import { showLoading, hideLoading, showOfflineScreen } from "./LoadingScreen";
import { mountLanding } from "./LandingPage";
import { mountOnboarding, openHowToPlayModal } from "./Onboarding";

const WARM_URLS = [
  "/cahaya/raka/board-gunung.png",
  "/cahaya/raka/ladder-bamboo.png",
  ...PLAYER_CHARACTERS.slice(0, 4).map((c) => c.src),
];

function warmAssets(): Promise<void> {
  const jobs = WARM_URLS.map(
    (url) =>
      new Promise<void>((resolve) => {
        const img = new Image();
        img.decoding = "async";
        img.onload = () => resolve();
        img.onerror = () => resolve();
        img.src = url;
      }),
  );
  return Promise.race([
    Promise.all(jobs).then(() => undefined),
    new Promise<void>((r) => window.setTimeout(r, 900)),
  ]);
}

function nextFrame(): Promise<void> {
  return new Promise((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
}

export async function bootGameApp(mountApp: (root: HTMLElement) => void): Promise<void> {
  const app = document.querySelector<HTMLElement>("#app");
  if (!app) throw new Error("Elemen #app tidak ada");

  showLoading();
  loadPrefs();
  initPWA();

  await Promise.all([warmAssets(), new Promise((r) => window.setTimeout(r, 220))]);
  await nextFrame();
  hideLoading();

  const params = new URLSearchParams(location.search);
  const forcePlay = params.get("play") === "1";
  const isOfflinePage = params.get("offline") === "1";

  if (!navigator.onLine && !isOfflinePage) {
    showOfflineScreen(() => location.reload());
    return;
  }

  window.addEventListener("offline", () => showOfflineScreen(() => location.reload()));

  const enterApp = (): void => {
    const prefs = loadPrefs();
    if (!prefs.tutorialCompleted) {
      mountOnboarding(app, {
        defaultUsername: storedSessionToken() ? undefined : prefs.nickname,
        onDone: () => mountApp(app),
      });
      return;
    }
    mountApp(app);
  };

  if (forcePlay || storedSessionToken()) {
    enterApp();
    return;
  }

  mountLanding(app, {
    onPlay: enterApp,
    onHow: openHowToPlayModal,
  });
}

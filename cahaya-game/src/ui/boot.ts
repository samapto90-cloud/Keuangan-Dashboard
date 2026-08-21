import { storedSessionToken } from "../auth/session";
import { initPWA } from "./pwa";
import { loadPrefs } from "./prefs";
import { showLoading, hideLoading, showOfflineScreen } from "./LoadingScreen";
import { mountLanding } from "./LandingPage";
import { mountOnboarding, openHowToPlayModal } from "./Onboarding";

export async function bootGameApp(mountApp: (root: HTMLElement) => void): Promise<void> {
  const app = document.querySelector<HTMLElement>("#app");
  if (!app) throw new Error("Elemen #app tidak ada");

  showLoading();
  loadPrefs();
  initPWA();

  await new Promise((r) => window.setTimeout(r, 350));
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

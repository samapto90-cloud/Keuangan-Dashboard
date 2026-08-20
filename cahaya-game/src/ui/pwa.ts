import { toast } from "./chrome";

interface BeforeInstallPromptEvent extends Event {
  prompt(): Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

let deferred: BeforeInstallPromptEvent | null = null;
let installed = false;

export function initPWA(): void {
  if (window.matchMedia("(display-mode: standalone)").matches) {
    installed = true;
    document.documentElement.dataset.pwa = "installed";
  }

  window.addEventListener("beforeinstallprompt", (e) => {
    e.preventDefault();
    deferred = e as BeforeInstallPromptEvent;
    document.documentElement.dataset.pwa = "installable";
  });

  window.addEventListener("appinstalled", () => {
    installed = true;
    deferred = null;
    document.documentElement.dataset.pwa = "installed";
    toast("Game terpasang di perangkat!", "success");
  });

  if ("serviceWorker" in navigator && import.meta.env.PROD) {
    void navigator.serviceWorker.register("/cahaya/sw.js", { scope: "/cahaya/" }).catch(() => undefined);
  }
}

export function canInstallPWA(): boolean {
  return !installed && deferred !== null;
}

export async function promptInstallPWA(): Promise<boolean> {
  if (!deferred) {
    toast("Instalasi belum tersedia di browser ini.", "info");
    return false;
  }
  await deferred.prompt();
  const choice = await deferred.userChoice;
  if (choice.outcome === "accepted") {
    deferred = null;
    return true;
  }
  return false;
}

export function renderInstallButton(): string {
  if (installed || !deferred) return "";
  return `<button type="button" class="nt-btn install-btn" data-pwa-install>📲 INSTALL GAME</button>`;
}

export function bindInstallButton(root: ParentNode): void {
  root.querySelector("[data-pwa-install]")?.addEventListener("click", () => {
    void promptInstallPWA();
  });
}

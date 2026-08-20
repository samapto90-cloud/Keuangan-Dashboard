let layer: HTMLElement | null = null;

export function showLoading(): void {
  if (layer) return;
  layer = document.createElement("div");
  layer.className = "boot-loading";
  layer.setAttribute("role", "status");
  layer.innerHTML = `
    <div class="boot-loading-inner">
      <p class="boot-kicker">ULAR TANGGA NUSANTARA</p>
      <div class="boot-bar" aria-hidden="true"><span class="boot-bar-fill"></span></div>
      <p class="boot-hint">Loading…</p>
    </div>`;
  document.body.appendChild(layer);
}

export function hideLoading(): void {
  if (!layer) return;
  layer.classList.add("is-out");
  const el = layer;
  window.setTimeout(() => el.remove(), 320);
  layer = null;
}

export function showOfflineScreen(onRetry: () => void): void {
  let host = document.querySelector<HTMLElement>(".offline-screen");
  if (host) return;
  host = document.createElement("div");
  host.className = "offline-screen";
  host.innerHTML = `
    <div class="offline-card nt-card">
      <p class="offline-icon" aria-hidden="true">📡</p>
      <h2>Koneksi Terputus</h2>
      <p>Periksa koneksi internetmu dan coba kembali.</p>
      <button type="button" class="nt-btn nt-btn-primary" data-retry>COBA LAGI</button>
    </div>`;
  host.querySelector("[data-retry]")?.addEventListener("click", () => {
    if (navigator.onLine) {
      host?.remove();
      onRetry();
    } else {
      onRetry();
    }
  });
  document.body.appendChild(host);
}

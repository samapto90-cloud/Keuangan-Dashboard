/* Ular Tangga Nusantara — network-first, bust stale theme CSS. */
const CACHE = "ular-static-v1.1.3-invite-join";
const OFFLINE_URL = "/cahaya/?offline=1";

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE).then((cache) =>
      cache.addAll([OFFLINE_URL, "/cahaya/favicon.svg", "/cahaya/manifest.webmanifest"]).catch(() => undefined),
    ),
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))),
    ).then(() => self.clients.claim()),
  );
});

function isApiOrWs(url) {
  return url.pathname.includes("/api/") || url.pathname.includes("/ws");
}

self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);
  if (event.request.method !== "GET" || isApiOrWs(url)) return;
  if (!url.pathname.startsWith("/cahaya/")) return;

  // Always network-first for HTML/CSS/JS so theme fixes are not stuck in cache
  const isShell =
    url.pathname.endsWith(".js") ||
    url.pathname.endsWith(".css") ||
    url.pathname.endsWith(".html") ||
    url.pathname === "/cahaya/" ||
    url.pathname.endsWith("/cahaya");

  event.respondWith(
    fetch(event.request)
      .then((res) => {
        if (res.ok && isShell) {
          const copy = res.clone();
          caches.open(CACHE).then((c) => c.put(event.request, copy));
        }
        return res;
      })
      .catch(async () => {
        const cached = await caches.match(event.request);
        if (cached) return cached;
        if (event.request.mode === "navigate") {
          const offline = await caches.match(OFFLINE_URL);
          if (offline) return offline;
        }
        return new Response("Offline", { status: 503, headers: { "Content-Type": "text/plain" } });
      }),
  );
});

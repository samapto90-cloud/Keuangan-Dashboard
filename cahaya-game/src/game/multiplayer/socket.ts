import { WS_EVENTS } from "./events";

export type ConnStatus = "offline" | "connecting" | "online";

export type NetMsg = { type: string; data?: Record<string, unknown> };

export function wsURL(): string {
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
  // Same-host: Hostinger tidak membuka :8888 ke publik.
  return `${proto}//${window.location.host}/cahaya/ws`;
}

export class GameClient {
  ws: WebSocket | null = null;
  status: ConnStatus = "offline";
  lastSeq = 0;
  seen = new Set<string>();
  myId = "";
  onStatus: ((s: ConnStatus) => void) | null = null;
  /** @deprecated prefer addListener */
  onEvent: ((type: string, data: Record<string, unknown>) => void) | null = null;
  private listeners = new Set<(type: string, data: Record<string, unknown>) => void>();
  private token = "";
  private pingTimer = 0;
  private mode: "ws" | "http" = "ws";
  private pollAbort: AbortController | null = null;
  private httpActive = false;
  private intentionalClose = false;
  private reconnectTimer = 0;
  private reconnectAttempt = 0;
  private connGen = 0;
  private statusDebounce = 0;
  private lastOnlineAt = 0;

  addListener(fn: (type: string, data: Record<string, unknown>) => void): () => void {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }

  connect(token: string): void {
    this.token = token;
    this.intentionalClose = false;
    window.clearTimeout(this.reconnectTimer);
    this.stopHttp();
    this.mode = "ws";
    this.setStatus("connecting", false);
    const gen = ++this.connGen;
    const ws = new WebSocket(wsURL());
    this.ws = ws;
    let opened = false;
    const failTimer = window.setTimeout(() => {
      if (!opened && gen === this.connGen) {
        try {
          ws.close();
        } catch {
          /* ignore */
        }
        this.fallbackOrRetry("open-timeout");
      }
    }, 4500);
    ws.addEventListener("open", () => {
      if (gen !== this.connGen) return;
      opened = true;
      window.clearTimeout(failTimer);
      this.mode = "ws";
      try {
        ws.send(JSON.stringify({ type: WS_EVENTS.AUTH, data: { token } }));
      } catch {
        this.fallbackOrRetry("auth-send");
      }
    });
    ws.addEventListener("message", (ev) => {
      if (gen !== this.connGen) return;
      this.handleRaw(String(ev.data));
    });
    ws.addEventListener("close", () => {
      window.clearTimeout(failTimer);
      if (gen !== this.connGen) return;
      if (this.intentionalClose) {
        this.setStatus("offline", true);
        window.clearInterval(this.pingTimer);
        return;
      }
      if (!opened) {
        this.fallbackOrRetry("close-before-open");
        return;
      }
      window.clearInterval(this.pingTimer);
      this.ws = null;
      this.scheduleReconnect();
    });
    ws.addEventListener("error", () => {
      if (gen !== this.connGen || opened) return;
      window.clearTimeout(failTimer);
      try {
        ws.close();
      } catch {
        /* ignore */
      }
    });
  }

  /** Tutup bersih (keluar menu) — tanpa auto-reconnect. */
  disconnect(): void {
    this.intentionalClose = true;
    window.clearTimeout(this.reconnectTimer);
    window.clearInterval(this.pingTimer);
    this.stopHttp();
    try {
      this.ws?.close();
    } catch {
      /* ignore */
    }
    this.ws = null;
    this.setStatus("offline", true);
  }

  reconnect(): void {
    this.reconnectAttempt = 0;
    window.clearTimeout(this.reconnectTimer);
    const old = this.ws;
    this.intentionalClose = false;
    this.stopHttp();
    this.connect(this.token);
    if (old && old !== this.ws) {
      window.setTimeout(() => {
        try {
          old.close();
        } catch {
          /* ignore */
        }
      }, 600);
    }
  }

  /** @returns true jika pesan terkirim (WS open atau HTTP bridge). */
  send(type: string, data: Record<string, unknown> = {}): boolean {
    if (this.mode === "http") {
      void fetch("/cahaya/api/realtime/send", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.token}`,
        },
        body: JSON.stringify({ type, data }),
      }).catch(() => undefined);
      return true;
    }
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return false;
    try {
      this.ws.send(JSON.stringify({ type, data }));
      return true;
    } catch {
      return false;
    }
  }

  private setStatus(s: ConnStatus, immediate: boolean): void {
    window.clearTimeout(this.statusDebounce);
    const apply = (): void => {
      if (this.status === s) return;
      this.status = s;
      this.onStatus?.(s);
    };
    // Jangan kedip banner/paint saat blip singkat — penyebab layar berkedip di multiplayer.
    if (!immediate && s === "offline" && Date.now() - this.lastOnlineAt < 4000) {
      this.statusDebounce = window.setTimeout(apply, 1800);
      return;
    }
    if (!immediate && s === "connecting" && this.status === "online" && Date.now() - this.lastOnlineAt < 10000) {
      this.statusDebounce = window.setTimeout(apply, 900);
      return;
    }
    if (s === "online") this.lastOnlineAt = Date.now();
    apply();
  }

  private scheduleReconnect(): void {
    if (this.intentionalClose || !this.token) {
      this.setStatus("offline", true);
      return;
    }
    this.setStatus("connecting", false);
    const attempt = this.reconnectAttempt++;
    // Setelah beberapa gagal WS, pakai HTTP bridge (proxy kadang putus Upgrade).
    if (attempt >= 3) {
      this.startHttpBridge();
      return;
    }
    const delay = Math.min(8000, 600 + attempt * 900);
    window.clearTimeout(this.reconnectTimer);
    this.reconnectTimer = window.setTimeout(() => {
      if (this.intentionalClose) return;
      this.connect(this.token);
    }, delay);
  }

  private fallbackOrRetry(_reason: string): void {
    if (this.intentionalClose) return;
    if (this.reconnectAttempt >= 2) {
      this.startHttpBridge();
      return;
    }
    this.scheduleReconnect();
  }

  private startHttpBridge(): void {
    if (this.httpActive || !this.token) return;
    this.httpActive = true;
    this.mode = "http";
    this.ws = null;
    this.setStatus("connecting", false);
    this.pollLoop();
  }

  private stopHttp(): void {
    this.httpActive = false;
    this.pollAbort?.abort();
    this.pollAbort = null;
  }

  private async pollLoop(): Promise<void> {
    while (this.httpActive && this.mode === "http") {
      this.pollAbort = new AbortController();
      try {
        const res = await fetch("/cahaya/api/realtime/poll", {
          method: "GET",
          headers: { Authorization: `Bearer ${this.token}` },
          signal: this.pollAbort.signal,
          cache: "no-store",
        });
        if (res.status === 401) {
          this.setStatus("offline", true);
          this.httpActive = false;
          return;
        }
        const text = await res.text();
        if (text) this.handleRaw(text);
      } catch {
        if (!this.httpActive) return;
        await new Promise((r) => setTimeout(r, 600));
        // Coba balik ke WS sesekali dari HTTP.
        if (this.reconnectAttempt > 0 && this.reconnectAttempt % 5 === 0) {
          this.httpActive = false;
          this.reconnectAttempt = 0;
          this.connect(this.token);
          return;
        }
      }
    }
  }

  private startPing(): void {
    window.clearInterval(this.pingTimer);
    const tick = (): void => {
      if (document.visibilityState === "hidden") return;
      this.send(WS_EVENTS.PING, { t: Date.now() });
    };
    tick();
    this.pingTimer = window.setInterval(tick, 8000);
  }

  private emit(type: string, data: Record<string, unknown>): void {
    this.onEvent?.(type, data);
    this.listeners.forEach((fn) => {
      try {
        fn(type, data);
      } catch {
        /* ignore listener errors */
      }
    });
  }

  private handleRaw(raw: string): void {
    let msg: NetMsg;
    try {
      msg = JSON.parse(raw) as NetMsg;
    } catch {
      return;
    }
    const data = (msg.data || {}) as Record<string, unknown>;
    if (typeof data.youAre === "string" && data.youAre) {
      this.myId = String(data.youAre);
    }
    if (msg.type === WS_EVENTS.AUTH_OK) {
      this.myId = String(data.playerId || data.youAre || this.myId || "");
      this.reconnectAttempt = 0;
      this.setStatus("online", true);
      this.send(WS_EVENTS.JOIN_LOBBY);
      this.startPing();
      this.emit(msg.type, data);
      return;
    }
    if (msg.type === WS_EVENTS.PONG) {
      this.emit(msg.type, data);
      return;
    }
    const seq = Number(data.seq || 0);
    const eventId = String(data.eventId || "");
    if (eventId && this.seen.has(eventId)) return;
    if (eventId) {
      this.seen.add(eventId);
      if (this.seen.size > 400) this.seen.clear();
    }
    // Urutan jaringan bisa acak; izinkan seq sama / sedikit mundur agar TURN/STATE tidak hilang.
    if (seq && this.lastSeq && seq + 5 < this.lastSeq) return;
    if (seq && seq > this.lastSeq) this.lastSeq = seq;
    this.emit(msg.type, data);
  }
}

export function connectLobby(token: string, onHello: (raw: unknown) => void): WebSocket | null {
  const c = new GameClient();
  c.onEvent = (type, data) => {
    if (type === WS_EVENTS.LOBBY_HELLO) onHello({ type, data });
  };
  c.connect(token);
  return c.ws;
}

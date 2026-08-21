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

  addListener(fn: (type: string, data: Record<string, unknown>) => void): () => void {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }

  connect(token: string): void {
    this.token = token;
    this.status = "connecting";
    this.onStatus?.(this.status);
    this.mode = "ws";
    this.stopHttp();
    const ws = new WebSocket(wsURL());
    this.ws = ws;
    let opened = false;
    const failTimer = window.setTimeout(() => {
      if (!opened) {
        try {
          ws.close();
        } catch {
          /* ignore */
        }
        this.startHttpBridge();
      }
    }, 3500);
    ws.addEventListener("open", () => {
      opened = true;
      window.clearTimeout(failTimer);
      this.mode = "ws";
      ws.send(JSON.stringify({ type: WS_EVENTS.AUTH, data: { token } }));
    });
    ws.addEventListener("message", (ev) => this.handleRaw(String(ev.data)));
    ws.addEventListener("close", () => {
      window.clearTimeout(failTimer);
      if (this.mode === "ws" && !opened) {
        this.startHttpBridge();
        return;
      }
      if (this.mode === "ws") {
        this.status = "offline";
        this.onStatus?.(this.status);
        window.clearInterval(this.pingTimer);
      }
    });
    ws.addEventListener("error", () => {
      if (!opened) {
        window.clearTimeout(failTimer);
        try {
          ws.close();
        } catch {
          /* ignore */
        }
        this.startHttpBridge();
      }
    });
  }

  reconnect(): void {
    this.ws?.close();
    this.stopHttp();
    this.connect(this.token);
  }

  send(type: string, data: Record<string, unknown> = {}): void {
    if (this.mode === "http") {
      void fetch("/cahaya/api/realtime/send", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.token}`,
        },
        body: JSON.stringify({ type, data }),
      }).catch(() => undefined);
      return;
    }
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    this.ws.send(JSON.stringify({ type, data }));
  }

  private startHttpBridge(): void {
    if (this.httpActive || !this.token) return;
    this.httpActive = true;
    this.mode = "http";
    this.ws = null;
    this.status = "online";
    this.onStatus?.(this.status);
    this.emit(WS_EVENTS.AUTH_OK, { playerId: this.myId || "", via: "http" });
    this.send(WS_EVENTS.JOIN_LOBBY);
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
          this.status = "offline";
          this.onStatus?.(this.status);
          this.httpActive = false;
          return;
        }
        const text = await res.text();
        if (text) this.handleRaw(text);
      } catch {
        await new Promise((r) => setTimeout(r, 800));
      }
    }
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
    if (msg.type === WS_EVENTS.AUTH_OK) {
      this.myId = String(data.playerId || this.myId || "");
      this.status = "online";
      this.onStatus?.(this.status);
      if (this.mode === "ws") {
        this.send(WS_EVENTS.JOIN_LOBBY);
        window.clearInterval(this.pingTimer);
        this.pingTimer = window.setInterval(() => this.send(WS_EVENTS.PING, { t: Date.now() }), 12000);
      }
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
    if (seq && seq < this.lastSeq) return;
    if (seq) this.lastSeq = seq;
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

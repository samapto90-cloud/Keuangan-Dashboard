import { WS_EVENTS } from "./events";

export type ConnStatus = "offline" | "connecting" | "online";

export type NetMsg = { type: string; data?: Record<string, unknown> };

export function wsURL(): string {
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.hostname.toLowerCase();
  // Proxy PHP di :443 tidak bisa Upgrade WebSocket — hubungkan langsung ke proses Go.
  if (host === "sakubijak.com" || host === "www.sakubijak.com") {
    return `${proto}//${host}:8888/cahaya/ws`;
  }
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

  addListener(fn: (type: string, data: Record<string, unknown>) => void): () => void {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }

  connect(token: string): void {
    this.token = token;
    this.status = "connecting";
    this.onStatus?.(this.status);
    const ws = new WebSocket(wsURL());
    this.ws = ws;
    ws.addEventListener("open", () => {
      ws.send(JSON.stringify({ type: WS_EVENTS.AUTH, data: { token } }));
    });
    ws.addEventListener("message", (ev) => this.handleRaw(String(ev.data)));
    ws.addEventListener("close", () => {
      this.status = "offline";
      this.onStatus?.(this.status);
      window.clearInterval(this.pingTimer);
    });
    ws.addEventListener("error", () => {
      this.status = "offline";
      this.onStatus?.(this.status);
    });
  }

  reconnect(): void {
    this.ws?.close();
    this.connect(this.token);
  }

  send(type: string, data: Record<string, unknown> = {}): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    this.ws.send(JSON.stringify({ type, data }));
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
      this.myId = String(data.playerId || "");
      this.status = "online";
      this.onStatus?.(this.status);
      this.send(WS_EVENTS.JOIN_LOBBY);
      window.clearInterval(this.pingTimer);
      this.pingTimer = window.setInterval(() => this.send(WS_EVENTS.PING, { t: Date.now() }), 12000);
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

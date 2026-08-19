import { WS_EVENTS } from "./events";

export function wsURL(): string {
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${window.location.host}/cahaya/ws`;
}

export function connectLobby(token: string, onHello: (raw: unknown) => void): WebSocket | null {
  try {
    const ws = new WebSocket(wsURL());
    ws.addEventListener("open", () => {
      ws.send(JSON.stringify({ type: WS_EVENTS.AUTH, data: { token } }));
    });
    ws.addEventListener("message", (ev) => {
      try {
        const msg = JSON.parse(String(ev.data)) as { type?: string };
        if (msg.type === WS_EVENTS.AUTH_OK) {
          ws.send(JSON.stringify({ type: WS_EVENTS.JOIN_LOBBY, data: {} }));
        }
        if (msg.type === WS_EVENTS.LOBBY_HELLO) onHello(msg);
      } catch {
        /* ignore malformed */
      }
    });
    return ws;
  } catch {
    return null;
  }
}

import { NET } from "../game/GameConfig";
import { netLog } from "./NetworkLog";
import {
  displayName,
  wsUrl,
  type Envelope,
  type MoveInput,
  type NetMsgType,
  type PlayerSpawn,
  type WelcomeOut,
  type WorldSnapshot,
} from "./NetworkMessage";

export type NetStatus = "offline" | "connecting" | "online" | "reconnecting";

export class NetworkClient {
  status: NetStatus = "offline";
  playerId = "";
  sessionId = "";
  name = displayName();
  channel = "";
  lastError = "";
  pingMs = 0;
  messagesPerSec = 0;
  private socket: WebSocket | null = null;
  private token = "";
  private closed = false;
  private pingTimer = 0;
  private msgCount = 0;
  private msgWindow = 0;
  private reconnectTimer = 0;
  private lastSnapshotLog = 0;

  onWelcome: ((w: WelcomeOut) => void) | null = null;
  onSpawn: ((p: PlayerSpawn) => void) | null = null;
  onDespawn: ((id: string) => void) | null = null;
  onSnapshot: ((s: WorldSnapshot) => void) | null = null;
  onStatus: ((s: NetStatus) => void) | null = null;
  onCombat: ((type: NetMsgType, data: unknown) => void) | null = null;
  onWorld: ((type: NetMsgType, data: unknown) => void) | null = null;

  setSession(token: string, name: string): void {
    this.token = token;
    this.name = name || displayName();
  }

  connect(): void {
    this.closed = false;
    this.clearReconnect();
    if (this.socket) {
      const prev = this.socket;
      this.socket = null;
      try {
        prev.close();
      } catch {
        /* ignore */
      }
    }
    this.setStatus(this.playerId ? "reconnecting" : "connecting");
    const sock = new WebSocket(wsUrl());
    this.socket = sock;
    sock.addEventListener("open", () => {
      if (this.socket !== sock) return;
      netLog("CONNECTED");
      this.send("AUTH", { name: this.name, token: this.token });
    });
    sock.addEventListener("message", (ev) => {
      if (this.socket !== sock) return;
      this.onMessage(String(ev.data));
    });
    sock.addEventListener("close", () => {
      if (this.socket !== sock) return;
      this.socket = null;
      this.stopPing();
      if (this.closed) {
        this.setStatus("offline");
        netLog("DISCONNECTED");
        return;
      }
      this.setStatus("reconnecting");
      netLog("RECONNECTING");
      this.reconnectTimer = window.setTimeout(() => this.connect(), NET.reconnectMs);
    });
    sock.addEventListener("error", () => {
      this.lastError = "koneksi gagal";
    });
  }

  disconnect(): void {
    this.closed = true;
    this.clearReconnect();
    this.stopPing();
    this.socket?.close();
    this.socket = null;
    this.setStatus("offline");
  }

  sendMoveInput(input: MoveInput): void {
    if (this.status !== "online") return;
    this.send("MOVE_INPUT", input);
  }

  sendCombat(type: NetMsgType, payload: object): void {
    if (this.status !== "online") return;
    this.send(type, payload);
  }

  sendWorld(type: NetMsgType, payload: object): void {
    if (this.status !== "online") return;
    this.send(type, payload);
  }

  private onMessage(raw: string): void {
    this.noteMessage();
    let msg: Envelope;
    try {
      msg = JSON.parse(raw) as Envelope;
    } catch {
      return;
    }
    const data = msg.data;
    switch (msg.type) {
      case "AUTH_OK": {
        const auth = data as { token?: string; playerId?: string; sessionId?: string };
        this.token = String(auth.token ?? this.token);
        this.playerId = String(auth.playerId ?? "");
        this.sessionId = String(auth.sessionId ?? this.token);
        netLog("AUTH_OK", this.playerId);
        this.send("JOIN_WORLD", { worldId: "world-01" });
        break;
      }
      case "AUTH_FAIL":
        this.lastError = String((data as { message?: string } | undefined)?.message ?? "auth gagal");
        this.closed = true;
        this.setStatus("offline");
        netLog("AUTH_FAIL");
        break;
      case "WELCOME": {
        const w = data as WelcomeOut;
        this.playerId = w.playerId;
        this.sessionId = w.sessionId;
        this.name = w.self.name;
        this.channel = w.channel;
        this.setStatus("online");
        this.startPing();
        netLog("PLAYER_JOIN", `${w.self.name} ${w.channel}`);
        this.onWelcome?.(w);
        break;
      }
      case "PLAYER_SPAWN":
      case "PLAYER_JOIN": {
        const spawn = data as PlayerSpawn;
        netLog("PLAYER_JOIN", spawn.name);
        this.onSpawn?.(spawn);
        break;
      }
      case "PLAYER_DESPAWN":
      case "PLAYER_LEAVE": {
        const id = String((data as { playerId?: string } | undefined)?.playerId ?? "");
        netLog("PLAYER_LEAVE", id);
        this.onDespawn?.(id);
        break;
      }
      case "WORLD_SNAPSHOT": {
        const snap = data as WorldSnapshot;
        const now = performance.now();
        if (now - this.lastSnapshotLog > 1000) {
          this.lastSnapshotLog = now;
          netLog("WORLD_SNAPSHOT", `online=${snap.online}`);
        }
        this.onSnapshot?.(snap);
        break;
      }
      case "PONG": {
        const pong = data as { t?: number };
        if (typeof pong.t === "number") this.pingMs = Math.max(0, Date.now() - pong.t);
        break;
      }
      case "ERROR":
        this.lastError = String((data as { message?: string } | undefined)?.message ?? "error");
        this.onWorld?.(msg.type, data);
        break;
      case "ATTACK_RESULT":
      case "DAMAGE_RESULT":
      case "PLAYER_HIT":
      case "ENEMY_HIT":
      case "ENEMY_DEATH":
      case "ENEMY_SPAWN":
      case "PLAYER_DEATH":
      case "PLAYER_RESPAWN":
      case "PLAYER_LEVEL_UP":
      case "ACTION_REJECT":
      case "SKILL_USED":
      case "TRANSFORMATION_STARTED":
      case "TRANSFORMATION_UPDATED":
      case "TRANSFORMATION_ENDED":
        this.onCombat?.(msg.type, data);
        this.onWorld?.(msg.type, data);
        break;
      default:
        this.onWorld?.(msg.type, data);
        break;
    }
  }

  private send(type: string, payload: object): void {
    if (this.socket?.readyState !== WebSocket.OPEN) return;
    this.socket.send(JSON.stringify({ type, data: payload }));
  }

  private startPing(): void {
    this.stopPing();
    this.pingTimer = window.setInterval(() => {
      this.send("PING", { t: Date.now() });
    }, NET.pingIntervalMs);
    this.send("PING", { t: Date.now() });
  }

  private stopPing(): void {
    if (this.pingTimer) window.clearInterval(this.pingTimer);
    this.pingTimer = 0;
  }

  private clearReconnect(): void {
    if (this.reconnectTimer) window.clearTimeout(this.reconnectTimer);
    this.reconnectTimer = 0;
  }

  private noteMessage(): void {
    const now = performance.now();
    if (this.msgWindow === 0) this.msgWindow = now;
    this.msgCount += 1;
    const elapsed = now - this.msgWindow;
    if (elapsed >= 1000) {
      this.messagesPerSec = Math.round((this.msgCount * 1000) / elapsed);
      this.msgCount = 0;
      this.msgWindow = now;
    }
  }

  private setStatus(s: NetStatus): void {
    this.status = s;
    this.onStatus?.(s);
  }
}

export { NetworkClient as GameClient };

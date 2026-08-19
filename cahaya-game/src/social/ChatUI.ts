import { escapeHtml } from "../dialogue/DialogueUI";
import type { ChatOut } from "../network/NetworkMessage";

const CHANNELS: { id: string; label: string }[] = [
  { id: "LOCAL", label: "AREA" },
  { id: "WORLD", label: "WORLD" },
  { id: "PARTY", label: "PARTY" },
  { id: "GUILD", label: "GUILD" },
  { id: "WHISPER", label: "PRIVATE" },
  { id: "RAID", label: "RAID" },
];

export class ChatUI {
  readonly root: HTMLElement;
  readonly log: HTMLElement;
  readonly input: HTMLInputElement;
  channel = "LOCAL";
  whisperTo = "";
  focused = false;
  muted = new Set<string>();
  onSend: ((channel: string, text: string, target?: string) => void) | null = null;
  onReport: ((fromId: string, category: string) => void) | null = null;
  onMute: ((fromId: string) => void) | null = null;
  private readonly lines: ChatOut[] = [];

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "chat-panel";
    this.root.innerHTML = `
      <div class="chat-head">
        ${CHANNELS.map((c) => `<button type="button" data-ch="${c.id}">${c.label}</button>`).join("")}
        <button type="button" class="chat-min" data-ch-min>Chat</button>
      </div>
      <div class="chat-log"></div>
      <form class="chat-form">
        <input type="text" maxlength="200" placeholder="Say something..." autocomplete="off" />
      </form>`;
    host.appendChild(this.root);
    this.log = this.root.querySelector(".chat-log") as HTMLElement;
    this.input = this.root.querySelector("input") as HTMLInputElement;
    this.root.addEventListener("click", (e) => {
      const btn = (e.target as HTMLElement).closest("[data-ch]") as HTMLElement | null;
      if (btn?.dataset.ch) {
        this.channel = btn.dataset.ch;
        this.markChannel();
      }
      if ((e.target as HTMLElement).closest("[data-ch-min]")) this.toggleCompact();
      const mute = (e.target as HTMLElement).closest("[data-mute]") as HTMLElement | null;
      if (mute?.dataset.mute) {
        this.muted.add(mute.dataset.mute);
        this.onMute?.(mute.dataset.mute);
      }
      const report = (e.target as HTMLElement).closest("[data-report]") as HTMLElement | null;
      if (report?.dataset.report) this.onReport?.(report.dataset.report, "SPAM");
    });
    this.root.querySelector("form")?.addEventListener("submit", (e) => {
      e.preventDefault();
      this.submit();
    });
    this.input.addEventListener("focus", () => {
      this.focused = true;
      this.root.classList.add("open");
    });
    this.input.addEventListener("blur", () => {
      this.focused = false;
    });
    this.markChannel();
  }

  toggleCompact(): void {
    this.root.classList.toggle("open");
    if (this.root.classList.contains("open")) this.focus();
  }

  focus(): void {
    this.root.classList.add("open");
    this.input.focus();
    this.focused = true;
  }

  blur(): void {
    this.input.blur();
    this.focused = false;
  }

  push(msg: ChatOut): void {
    if (msg.fromId && this.muted.has(msg.fromId) && !msg.system) return;
    this.lines.push(msg);
    if (this.lines.length > 80) this.lines.shift();
    const row = document.createElement("div");
    row.className = `chat-line ch-${(msg.channel || "LOCAL").toLowerCase()}`;
    const tag = msg.system ? "SYSTEM" : msg.channel === "LOCAL" ? "AREA" : msg.channel === "WHISPER" ? "PRIVATE" : msg.channel || "LOCAL";
    const ts = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    row.innerHTML = `<small>${ts}</small> <span class="chat-tag">[${escapeHtml(tag)}]</span> <b>${escapeHtml(msg.from)}</b>: ${escapeHtml(msg.text)}
      ${msg.fromId && !msg.system ? `<button type="button" data-mute="${escapeHtml(msg.fromId)}">MUTE</button><button type="button" data-report="${escapeHtml(msg.fromId)}">REPORT</button>` : ""}`;
    this.log.appendChild(row);
    this.log.scrollTop = this.log.scrollHeight;
  }

  private submit(): void {
    let text = this.input.value.trim();
    if (!text) return;
    let channel = this.channel;
    let target = this.whisperTo;
    if (text.startsWith("/w ") || text.startsWith("/whisper ")) {
      const parts = text.split(" ");
      target = parts[1] || "";
      text = parts.slice(2).join(" ");
      channel = "WHISPER";
    }
    this.onSend?.(channel, text, target);
    this.input.value = "";
  }

  private markChannel(): void {
    this.root.querySelectorAll("[data-ch]").forEach((el) => {
      el.classList.toggle("on", (el as HTMLElement).dataset.ch === this.channel);
    });
    this.input.placeholder = this.channel === "WHISPER" ? "/w Name message" : `${this.channel} chat`;
  }
}

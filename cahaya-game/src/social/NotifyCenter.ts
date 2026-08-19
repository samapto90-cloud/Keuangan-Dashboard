import { escapeHtml } from "../dialogue/DialogueUI";
import type { NotifyView } from "../network/NetworkMessage";

export class NotifyCenter {
  readonly root: HTMLElement;
  blocking = false;
  items: NotifyView[] = [];

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "notify-center";
    host.appendChild(this.root);
    this.root.addEventListener("click", () => this.toggle());
  }

  apply(list: NotifyView[] | undefined): void {
    if (list) this.items = list;
    this.render();
  }

  push(type: string, message: string): void {
    this.items.unshift({ id: `local-${Date.now()}`, type, message, timestamp: Date.now() });
    if (this.items.length > 40) this.items.pop();
    this.render();
  }

  toggle(): void {
    this.blocking = !this.blocking;
    this.root.classList.toggle("open", this.blocking);
    this.render();
  }

  close(): void {
    this.blocking = false;
    this.root.classList.remove("open");
    this.render();
  }

  render(): void {
    const unread = this.items.filter((n) => !n.read).length;
    const list = this.blocking
      ? `<div class="notify-list">${this.items.map((n) => `<div class="prio-${escapeHtml((n.priority || "NORMAL").toLowerCase())}"><small>${escapeHtml(n.type)} · ${escapeHtml(n.priority || "NORMAL")}</small><p>${escapeHtml(n.message)}</p></div>`).join("") || "<p>Tidak ada notifikasi.</p>"}</div>`
      : "";
    this.root.innerHTML = `<button type="button" class="notify-bell">NOTICE ${unread || ""}</button>${list}`;
  }
}

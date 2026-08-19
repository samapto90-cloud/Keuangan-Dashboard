import { escapeHtml } from "../dialogue/DialogueUI";
import type { ChapterView } from "../network/NetworkMessage";

export class ChapterUI {
  readonly root: HTMLElement;
  blocking = false;
  onEnter: ((dungeonId: string) => void) | null = null;
  onClose: (() => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "dun-overlay chapter-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const t = (e.target as HTMLElement).closest("[data-ch]") as HTMLElement | null;
      if (!t) return;
      if (t.dataset.ch === "close") {
        this.close();
        return;
      }
      if (t.dataset.enter) this.onEnter?.(t.dataset.enter);
    });
  }

  open(chapters: ChapterView[]): void {
    this.blocking = true;
    this.root.hidden = false;
    const rows = chapters
      .map((c) => {
        const locked = c.status === "LOCKED";
        const done = c.status === "COMPLETED";
        const tag = done ? "COMPLETED" : locked ? "LOCKED" : "AVAILABLE";
        const btn = locked
          ? ""
          : `<button type="button" data-ch="enter" data-enter="${escapeHtml(c.dungeonId)}">ENTER</button>`;
        return `<article class="ch-row ${c.status.toLowerCase()}">
          <div><strong>${escapeHtml(c.title)}</strong><span>${escapeHtml(c.region)} · ${escapeHtml(c.bossName || c.bossId)}</span></div>
          <em>${escapeHtml(tag)}</em>${btn}
        </article>`;
      })
      .join("");
    this.root.innerHTML = `
      <div class="dun-card dun-wide">
        <div class="dun-kicker">33 PETUALANGAN</div>
        <h3>ADVENTURE</h3>
        <div class="ch-list">${rows}</div>
        <button type="button" data-ch="close">Tutup</button>
      </div>`;
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
    this.onClose?.();
  }

  toggle(chapters: ChapterView[]): void {
    if (this.blocking) this.close();
    else this.open(chapters);
  }
}

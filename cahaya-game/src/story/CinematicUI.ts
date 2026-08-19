export class CinematicUI {
  readonly root: HTMLElement;
  blocking = false;
  onSkip: ((id: string) => void) | null = null;
  onDone: ((id: string) => void) | null = null;
  private elapsed = 0;
  private duration = 8;
  private cinematicId = "";
  private line = 0;
  private lines: string[] = [];
  private subs: string[] = [];

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "cin-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const btn = (e.target as HTMLElement).closest("[data-cin]");
      if (btn instanceof HTMLElement) {
        if (btn.dataset.cin === "skip") this.skip();
        else this.advance();
        return;
      }
      this.skip();
    });
  }

  open(data: { id?: string; title?: string; durationSec?: number; lines?: string[]; subtitle?: string[]; camera?: string; music?: string }): void {
    this.cinematicId = data.id || "";
    this.lines = data.lines?.length ? data.lines : [data.title || ""];
    this.subs = data.subtitle ?? [];
    this.line = 0;
    this.elapsed = 0;
    this.duration = Math.min(30, Math.max(5, data.durationSec || 8));
    this.blocking = true;
    this.root.hidden = false;
    this.render();
  }

  update(dt: number): void {
    if (!this.blocking) return;
    this.elapsed += dt;
    if (this.elapsed >= this.duration) this.finish();
  }

  advance(): void {
    if (!this.blocking) return;
    if (this.line < this.lines.length - 1) {
      this.line += 1;
      this.render();
      return;
    }
    this.finish();
  }

  finish(): void {
    if (!this.blocking) return;
    const id = this.cinematicId;
    this.close();
    this.onDone?.(id);
  }

  skip(): void {
    if (!this.blocking) return;
    const id = this.cinematicId;
    this.close();
    this.onSkip?.(id);
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
    this.cinematicId = "";
  }

  private render(): void {
    const text = this.lines[this.line] || "";
    const sub = this.subs[this.line] || "";
    this.root.innerHTML = `
      <div class="cin-card">
        <p class="cin-line">${escapeCin(text)}</p>
        ${sub ? `<p class="cin-sub">${escapeCin(sub)}</p>` : ""}
        <div class="cin-actions">
          <button type="button" data-cin="next">Lanjut</button>
          <button type="button" data-cin="skip">LEWATI</button>
        </div>
        <p class="cin-hint">SPACE / ENTER lanjut · ESC lewati</p>
      </div>`;
  }
}

function escapeCin(s: string): string {
  return s.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] ?? c)).replace(/\n/g, "<br/>");
}

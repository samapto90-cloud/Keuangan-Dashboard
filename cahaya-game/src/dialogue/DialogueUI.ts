export class DialogueUI {
  readonly root: HTMLElement;
  private onOption: ((id: string) => void) | null = null;
  blocking = false;
  speed: "slow" | "normal" | "fast" = "normal";
  private fullText = "";
  private shown = 0;
  private timer = 0;
  private skipped = false;
  private history: string[] = [];
  private showHistory = false;
  private showTrans = true;
  private subtitle = "";
  private speaker = "";
  private role = "";
  private title = "";
  private options: Array<{ id: string; label: string }> = [];
  private emotion = "";
  private gesture = "";

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "dlg-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    const saved = window.localStorage.getItem("cahaya-dlg-speed");
    if (saved === "slow" || saved === "normal" || saved === "fast") this.speed = saved;
    this.root.addEventListener("click", (e) => {
      const btn = (e.target as HTMLElement).closest("button[data-opt], button[data-dlg]");
      if (btn instanceof HTMLElement) {
        if (btn.dataset.dlg === "hist") {
          this.showHistory = !this.showHistory;
          this.render(false);
          return;
        }
        if (btn.dataset.dlg === "trans") {
          this.showTrans = !this.showTrans;
          this.render(false);
          return;
        }
        if (btn.dataset.opt) this.onOption?.(btn.dataset.opt ?? "close");
        return;
      }
      if (!this.skipped && this.shown < this.fullText.length) {
        this.skipType();
        return;
      }
    });
  }

  setSpeed(speed: "slow" | "normal" | "fast"): void {
    this.speed = speed;
    window.localStorage.setItem("cahaya-dlg-speed", speed);
  }

  setSubtitle(on: boolean): void {
    this.showTrans = on;
    if (this.blocking) this.render(false);
  }

  open(
    title: string,
    speaker: string,
    role: string,
    text: string,
    options: Array<{ id: string; label: string }>,
    handler: (id: string) => void,
    subtitle = "",
    extra?: { emotion?: string; gesture?: string; history?: string[] },
  ): void {
    this.onOption = handler;
    this.blocking = true;
    this.root.hidden = false;
    this.title = title;
    this.speaker = speaker;
    this.role = role;
    this.fullText = text || "";
    this.subtitle = subtitle;
    this.options = options.length ? options : [{ id: "close", label: "Tutup" }];
    this.emotion = extra?.emotion || "";
    this.gesture = extra?.gesture || "";
    if (extra?.history?.length) this.history = extra.history.slice(-12);
    if (text) this.history = [...this.history, `${speaker}: ${text}`].slice(-16);
    this.shown = 0;
    this.skipped = false;
    window.clearInterval(this.timer);
    const ms = this.speed === "slow" ? 55 : this.speed === "fast" ? 12 : 28;
    this.render(true);
    this.timer = window.setInterval(() => {
      if (this.shown >= this.fullText.length) {
        this.skipType();
        return;
      }
      this.shown += 1;
      this.render(true);
    }, ms);
  }

  private skipType(): void {
    this.skipped = true;
    this.shown = this.fullText.length;
    window.clearInterval(this.timer);
    this.render(false);
  }

  private render(typing: boolean): void {
    const visible = this.fullText.slice(0, this.shown);
    const opts = typing ? [] : this.options;
    const hist = this.showHistory
      ? `<div class="dlg-history">${this.history.map((h) => `<p>${escapeHtml(h)}</p>`).join("")}</div>`
      : "";
    this.root.innerHTML = `
      <div class="dlg-card dlg-emo-${escapeHtml(this.emotion)} dlg-ges-${escapeHtml(this.gesture)}">
        <div class="dlg-head">
          <div class="dlg-portrait" data-emo="${escapeHtml(this.emotion)}" data-ges="${escapeHtml(this.gesture)}" aria-hidden="true">${escapeHtml((this.speaker || "?").slice(0, 1))}</div>
          <div>
            <div class="dlg-speaker">${escapeHtml(this.speaker)}${this.role ? `<span>${escapeHtml(this.role)}</span>` : ""}</div>
            <h3>${escapeHtml(this.title)}</h3>
          </div>
        </div>
        <p class="dlg-text">${escapeHtml(visible)}${typing && this.shown < this.fullText.length ? "<i class='dlg-caret'>▌</i>" : ""}</p>
        ${this.showTrans && this.subtitle ? `<p class="dlg-sub">${escapeHtml(this.subtitle)}</p>` : ""}
        <div class="dlg-tools">
          <button type="button" data-dlg="trans">${this.showTrans ? "Sembunyikan terjemahan" : "Terjemahan"}</button>
          <button type="button" data-dlg="hist">Riwayat</button>
        </div>
        ${hist}
        <div class="dlg-opts">${opts.map((o) => `<button type="button" data-opt="${escapeHtml(o.id)}">${escapeHtml(o.label)}</button>`).join("")}</div>
        <p class="dlg-hint">${typing ? "Ketuk untuk skip" : "SPACE / ENTER lanjut · ESC tutup"} · ${this.speed.toUpperCase()}</p>
      </div>`;
  }

  advance(): void {
    if (!this.blocking) return;
    if (!this.skipped && this.shown < this.fullText.length) {
      this.skipType();
      return;
    }
    const opts = this.root.querySelectorAll("button[data-opt]");
    if (opts.length === 1) this.onOption?.((opts[0] as HTMLElement).dataset.opt ?? "close");
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
    this.onOption = null;
    window.clearInterval(this.timer);
  }
}

export function escapeHtml(s: string): string {
  return s.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] ?? c));
}

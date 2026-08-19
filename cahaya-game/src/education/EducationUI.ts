import { escapeHtml } from "../dialogue/DialogueUI";
import type { QuestionOut, EducationFeedback } from "../network/NetworkMessage";

export class EducationUI {
  readonly root: HTMLElement;
  blocking = false;
  private onChoice: ((index: number) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "edu-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => {
      const skip = (e.target as HTMLElement).closest("[data-skip]");
      if (skip instanceof HTMLElement) {
        this.close();
        return;
      }
      const btn = (e.target as HTMLElement).closest("button[data-choice]");
      if (!(btn instanceof HTMLElement)) return;
      this.onChoice?.(Number(btn.dataset.choice));
    });
  }

  showQuestion(q: QuestionOut, handler: (index: number) => void): void {
    this.blocking = true;
    this.onChoice = handler;
    this.root.hidden = false;
    const skipBtn = q.id === "q-dungeon-1" ? `<button type="button" class="edu-choice" data-skip="1">LEWATI</button>` : "";
    this.root.innerHTML = `
      <div class="edu-card">
        <div class="edu-kicker">UJIAN PERJALANAN · ${escapeHtml(q.category)}</div>
        <div class="edu-progress">QUESTION ${q.index} / ${q.total}</div>
        <h2 class="edu-prompt">${escapeHtml(q.prompt)}</h2>
        <div class="edu-choices">
          ${(q.choices ?? []).map((c, i) => `<button type="button" class="edu-choice" data-choice="${i}">${String.fromCharCode(65 + i)}. ${escapeHtml(c)}</button>`).join("")}
          ${skipBtn}
        </div>
      </div>`;
  }

  showFeedback(fb: EducationFeedback, retry?: QuestionOut): void {
    void retry;
    const banner = fb.correct ? (fb.toast === "PERFECT!" ? "PERFECT!" : "Benar!") : "Coba lagi.";
    const extra = fb.explain ? `<p class="edu-explain">${escapeHtml(fb.explain)}</p>` : "";
    const existing = this.root.querySelector(".edu-card");
    if (existing) {
      const note = document.createElement("div");
      note.className = fb.correct ? "edu-ok" : "edu-retry";
      note.textContent = banner;
      existing.prepend(note);
      if (extra) {
        const p = document.createElement("div");
        p.innerHTML = extra;
        existing.appendChild(p.firstElementChild ?? p);
      }
    }
    if (retry) {
      /* Game shows the same question again after a short toast. */
    }
    if (fb.correct && !fb.question && !retry) {
      window.setTimeout(() => this.close(), 1200);
    }
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
    this.onChoice = null;
  }
}

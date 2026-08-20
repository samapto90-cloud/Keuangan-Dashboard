export type QuestionPublic = {
  id: string;
  subject: string;
  category?: string;
  grade?: string;
  difficulty: string;
  question: string;
  optionA: string;
  optionB: string;
  optionC: string;
  optionD: string;
  number?: number;
  questionNo?: number;
  timeLimit: number;
  endsAt?: number;
  playerId?: string;
  username?: string;
  final?: boolean;
};

export type QuestionResultView = {
  result: string;
  correct: boolean;
  timeout: boolean;
  correctAnswer?: string;
  explanation?: string;
  feedback?: string;
  positionBeforePenalty?: number;
  penalty?: number;
  positionAfterPenalty?: number;
  path?: number[];
  won?: boolean;
  final?: boolean;
  userId?: string;
  username?: string;
  powerGrant?: string;
  reward?: { xp?: number; coins?: number; levelUp?: boolean; levelBefore?: number; levelAfter?: number; achievements?: string[] };
};

const SUBJECT_LABEL: Record<string, string> = {
  PAI: "PAI",
  MATEMATIKA: "MATEMATIKA",
  BAHASA_INGGRIS: "BAHASA INGGRIS",
  BAHASA_JAWA: "BAHASA JAWA",
};

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export function subjectLabel(s: string): string {
  return SUBJECT_LABEL[s] || s;
}

export function mountQuestionOverlay(
  host: HTMLElement,
  opts: {
    question: QuestionPublic;
    selfId: string;
    answering: boolean;
    result?: QuestionResultView | null;
    onAnswer?: (letter: string) => void;
    now?: number;
  },
): { tick: () => void; root: HTMLElement } {
  let existing = host.querySelector<HTMLElement>(".q-overlay");
  if (!existing) {
    existing = document.createElement("div");
    existing.className = "q-overlay";
    host.appendChild(existing);
  }
  const q = opts.question;
  const num = q.number || q.questionNo || 1;
  const mine = opts.answering;
  const result = opts.result;
  const remain = Math.max(0, Math.ceil(((q.endsAt || 0) - (opts.now || Date.now())) / 1000));
  const warn = remain <= 10;
  const crit = remain <= 5;
  const deg = Math.max(0, Math.min(360, (remain / (q.timeLimit || 15)) * 360));
  existing.classList.toggle("is-mobile-full", window.matchMedia("(max-width: 768px)").matches);
  existing.classList.toggle("is-result", Boolean(result));
  if (result) {
    const ok = result.result === "CORRECT";
    const title = ok ? "✓ Jawaban Benar!" : result.timeout ? "Waktu Habis!" : "✕ Belum tepat.";
    existing.innerHTML = `<div class="q-modal ${ok ? "q-ok" : "q-bad"}">
      <p class="q-kicker">${title}</p>
      ${!ok && result.correctAnswer ? `<p class="q-answer">Jawaban benar: <strong>${esc(result.correctAnswer)}</strong></p>` : ""}
      <p class="q-explain">${esc(result.explanation || "")}</p>
      <p class="q-feed">${esc(result.feedback || "")}</p>
      ${!ok ? `<p class="q-pen">Mundur ${result.penalty || 10} kotak.</p>` : ""}
      ${ok && result.powerGrant ? `<p class="q-xp">🎁 Hadiah: ${esc(result.powerGrant === "bomb" ? "💣 Bom" : result.powerGrant === "thunder" ? "⚡ Petir" : "✈️ Pesawat")}</p>` : ""}
      ${result.reward?.xp ? `<p class="q-xp">✨ +${result.reward.xp} XP</p>` : ""}
    </div>`;
  } else if (!mine) {
    existing.innerHTML = `<div class="q-modal q-wait">
      <p class="q-kicker">🎓 TANTANGAN</p>
      <p class="q-watch">🔴 ${esc(q.username || "Pemain")} sedang menjawab.</p>
      <p class="q-sub">Menunggu hasil…</p>
    </div>`;
  } else {
    existing.innerHTML = `<div class="q-modal">
      <header class="q-head">
        <p class="q-kicker">${q.final ? "🏆 FINAL CHALLENGE" : "🎓 TANTANGAN"}</p>
        <span class="q-badge">${esc(subjectLabel(q.subject))}</span>
        <span class="q-diff">${esc(q.grade || "SMA")}</span>
        <span class="q-diff">${esc(q.difficulty)}</span>
        <span class="q-no">No. ${num}</span>
        <div class="q-timer ${crit ? "is-crit" : warn ? "is-warn" : ""}" style="--deg:${deg}deg" aria-label="sisa waktu"><span>⏱ ${remain}</span></div>
      </header>
      <p class="q-text">${esc(q.question)}</p>
      <div class="q-opts">
        ${(["A", "B", "C", "D"] as const)
          .map((L) => {
            const text = L === "A" ? q.optionA : L === "B" ? q.optionB : L === "C" ? q.optionC : q.optionD;
            return `<button type="button" class="q-opt" data-ans="${L}"><b>${L}</b> ${esc(text)}</button>`;
          })
          .join("")}
      </div>
    </div>`;
    existing.querySelectorAll<HTMLButtonElement>("[data-ans]").forEach((btn) => {
      btn.addEventListener("click", () => {
        existing.querySelectorAll<HTMLButtonElement>("[data-ans]").forEach((b) => {
          b.disabled = true;
        });
        opts.onAnswer?.(btn.dataset.ans || "");
        btn.classList.add("is-pick");
      });
    });
  }
  return {
    root: existing,
    tick: () => {
      if (result || !mine) return;
      const left = Math.max(0, Math.ceil(((q.endsAt || 0) - Date.now()) / 1000));
      const timer = existing!.querySelector(".q-timer");
      const span = timer?.querySelector("span");
      if (span) span.textContent = `⏱ ${left}`;
      timer?.classList.toggle("is-warn", left <= 10);
      timer?.classList.toggle("is-crit", left <= 5);
      if (timer) (timer as HTMLElement).style.setProperty("--deg", `${Math.max(0, Math.min(360, (left / (q.timeLimit || 15)) * 360))}deg`);
    },
  };
}

export function unmountQuestionOverlay(host: HTMLElement): void {
  host.querySelector(".q-overlay")?.remove();
}

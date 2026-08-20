import { playSfx } from "../audio/manager";
import { closeModal, showModal } from "./chrome";
import { loadPrefs, savePrefs } from "./prefs";

const AVATARS = ["🦅", "🐯", "🐬", "🌋", "🌺", "⚡"];

type Step =
  | "welcome"
  | "avatar"
  | "username"
  | "board"
  | "dice"
  | "question"
  | "snake"
  | "ladder"
  | "done";

const STEPS: Step[] = ["welcome", "avatar", "username", "board", "dice", "question", "snake", "ladder", "done"];

export type OnboardingOpts = {
  defaultUsername?: string;
  onDone: () => void;
};

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export function mountOnboarding(root: HTMLElement, opts: OnboardingOpts): void {
  document.body.classList.add("onboard-active");
  const prefs = loadPrefs();
  let step: Step = "welcome";
  let avatar = prefs.avatar || AVATARS[0];
  let username = opts.defaultUsername || prefs.nickname || "";

  const render = (): void => {
    const idx = STEPS.indexOf(step);
    root.innerHTML = `
      <style id="onboard-contrast-fix">
        body.onboard-active{background:linear-gradient(180deg,#ce1126 0%,#9e0d1e 24%,#0d1410 24%)!important}
        body.onboard-active:before,body.onboard-active:after{display:none!important}
        .onboard-card.nt-card{background:#fff!important;color:#1a1a1a!important;border:2px solid rgba(206,17,38,.2)!important}
        .onboard-card h1,.onboard-card h2{color:#ce1126!important}
        .onboard-card p,.onboard-card label{color:#333!important}
        .onboard-lead{color:#9e0d1e!important}
        .demo-opts button{background:#f5f5f5!important;color:#1a1a1a!important}
      </style>
      <main class="onboard-shell page-in">
        <div class="onboard-progress" aria-hidden="true">${STEPS.map((_, i) => `<span class="${i <= idx ? "on" : ""}"></span>`).join("")}</div>
        ${bodyFor(step)}
        <div class="onboard-actions">
          ${step !== "welcome" ? `<button type="button" class="nt-btn nt-btn-ghost" data-back>Kembali</button>` : ""}
          ${step !== "done" ? `<button type="button" class="nt-btn nt-btn-ghost" data-skip>Lewati Tutorial</button>` : ""}
          <button type="button" class="nt-btn nt-btn-primary" data-next>${step === "done" ? "MULAI BERMAIN" : "Lanjut"}</button>
        </div>
      </main>`;
    bind();
  };

  const bodyFor = (s: Step): string => {
    switch (s) {
      case "welcome":
        return `<section class="onboard-card nt-card"><h1>ULAR TANGGA NUSANTARA</h1><p class="onboard-lead">Belajar, Bermain, dan Bersaing!</p><p>Selamat datang! Ikuti petunjuk singkat untuk mengenal permainan.</p></section>`;
      case "avatar":
        return `<section class="onboard-card nt-card"><h2>Pilih Avatar</h2><p>Pilih simbol yang mewakili dirimu.</p><div class="avatar-grid">${AVATARS.map((a) => `<button type="button" class="avatar-pick ${a === avatar ? "on" : ""}" data-av="${a}">${a}</button>`).join("")}</div></section>`;
      case "username":
        return `<section class="onboard-card nt-card"><h2>Nama Pemain</h2><p>Username akan ditampilkan di papan dan leaderboard.</p><label>Nama<input name="nick" value="${esc(username)}" maxlength="16" autocomplete="nickname"/></label></section>`;
      case "board":
        return `<section class="onboard-card nt-card tutorial-step"><span class="tut-icon">🗺️</span><h2>Kenali Papan</h2><p>Papan 100 kotak. Mulai dari 1, capai 100 untuk menang.</p><div class="mini-board" aria-hidden="true">${Array.from({ length: 10 }, (_, i) => `<span>${i + 1}</span>`).join("")}</div></section>`;
      case "dice":
        return `<section class="onboard-card nt-card tutorial-step"><span class="tut-icon">🎲</span><h2>Dadu</h2><p>Tekan tombol dadu saat giliranmu.</p><button type="button" class="demo-dice" data-demo-dice>🎲 Lempar Demo</button><p class="demo-dice-out" id="demo-dice-out"></p></section>`;
      case "question":
        return `<section class="onboard-card nt-card tutorial-step"><span class="tut-icon">📚</span><h2>Soal Edukasi</h2><p>Setiap tantangan menguji pengetahuanmu.</p><div class="demo-q"><p><strong>2 + 2 = ?</strong></p><div class="demo-opts"><button type="button" data-q="A">A. 3</button><button type="button" data-q="B">B. 4</button><button type="button" data-q="C">C. 5</button><button type="button" data-q="D">D. 6</button></div><p class="demo-q-fb" id="demo-q-fb"></p></div></section>`;
      case "snake":
        return `<section class="onboard-card nt-card tutorial-step warn"><span class="tut-icon">🐍</span><h2>Ular</h2><p>Hati-hati, ular dapat membawamu turun!</p></section>`;
      case "ladder":
        return `<section class="onboard-card nt-card tutorial-step good"><span class="tut-icon">🪜</span><h2>Tangga</h2><p>Naik tangga untuk mencapai tujuan lebih cepat.</p></section>`;
      case "done":
        return `<section class="onboard-card nt-card"><h2>Siap Berpetualang!</h2><p>Avatar: ${avatar} · Nama: ${esc(username || "Pemain")}</p><p>Kamu siap bermain online atau latihan lokal.</p></section>`;
    }
  };

  const bind = (): void => {
    root.querySelector("[data-skip]")?.addEventListener("click", finish);
    root.querySelector("[data-back]")?.addEventListener("click", () => {
      playSfx("button");
      const i = STEPS.indexOf(step);
      if (i > 0) step = STEPS[i - 1];
      render();
    });
    root.querySelector("[data-next]")?.addEventListener("click", () => {
      playSfx("button");
      if (step === "username") {
        const inp = root.querySelector<HTMLInputElement>('input[name="nick"]');
        username = (inp?.value || username).trim().slice(0, 16) || "Pemain";
      }
      if (step === "done") {
        finish();
        return;
      }
      const i = STEPS.indexOf(step);
      step = STEPS[Math.min(i + 1, STEPS.length - 1)];
      render();
    });
    root.querySelectorAll("[data-av]").forEach((btn) => {
      btn.addEventListener("click", () => {
        avatar = (btn as HTMLElement).dataset.av || avatar;
        render();
      });
    });
    root.querySelector("[data-demo-dice]")?.addEventListener("click", () => {
      playSfx("dice_roll");
      const n = 1 + Math.floor(Math.random() * 6);
      const out = root.querySelector("#demo-dice-out");
      if (out) out.textContent = `Kamu mendapat ${n}!`;
    });
    root.querySelectorAll("[data-q]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const pick = (btn as HTMLElement).dataset.q;
        const fb = root.querySelector("#demo-q-fb");
        if (!fb) return;
        if (pick === "B") {
          playSfx("correct");
          fb.textContent = "✓ BENAR! Lanjut bermain.";
          fb.className = "demo-q-fb ok";
        } else {
          playSfx("wrong");
          fb.textContent = "✗ KURANG TEPAT. Mundur sesuai aturan game.";
          fb.className = "demo-q-fb bad";
        }
      });
    });
  };

  const finish = (): void => {
    document.body.classList.remove("onboard-active");
    savePrefs({ tutorialCompleted: true, avatar, nickname: username });
    opts.onDone();
  };

  render();
}

export function resetTutorial(): void {
  savePrefs({ tutorialCompleted: false });
}

export function openHowToPlayModal(): void {
  showModal(
    "how-play",
    `<h2>Cara Bermain</h2>
    <ol class="how-list">
      <li>Lempar dadu pada giliranmu.</li>
      <li>Bidak berjalan selangkah demi selangkah.</li>
      <li>Jawab soal setelah berhenti.</li>
      <li>Salah atau waktu habis: mundur 10 kotak.</li>
      <li>Hindari ular, manfaatkan tangga.</li>
      <li>Capai kotak 100 dan jawab Final Challenge.</li>
    </ol>
    <button type="button" class="nt-btn nt-btn-primary" data-close>Mengerti</button>`,
  );
  document.querySelector("[data-modal=how-play] [data-close]")?.addEventListener("click", () => closeModal("how-play"));
}

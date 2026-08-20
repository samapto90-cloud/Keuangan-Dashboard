import { GAME_TITLE, GAME_VERSION } from "../game/board/constants";
import { playSfx } from "../audio/manager";
import { bindInstallButton, renderInstallButton } from "./pwa";

export type LandingOpts = {
  onPlay: () => void;
  onHow: () => void;
};

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

export function mountLanding(root: HTMLElement, opts: LandingOpts): void {
  document.body.classList.add("landing-active");
  root.innerHTML = `
    <style id="landing-contrast-fix">
      body.landing-active{background:linear-gradient(180deg,#ce1126 0%,#ce1126 36%,#fff 36%,#f3f3f3 100%)!important;color:#1a1a1a!important}
      body.landing-active:before,body.landing-active:after{display:none!important}
      .landing-hero,.landing-hero h1,.landing-hero h1 span,.landing-kicker,.landing-tag,.landing-ver{color:#fff!important}
      .landing .nt-card{background:#fff!important;color:#1a1a1a!important;border:2px solid rgba(206,17,38,.22)!important}
      .landing-section h2,.landing-footer h2{color:#ce1126!important}
      .feature-grid li{background:#f7f7f7!important;color:#1a1a1a!important;border:1px solid #e5e5e5!important;padding:12px;border-radius:12px}
      .feature-grid li strong{color:#111827!important;display:block;font-weight:800}
      .feature-grid li p{color:#374151!important;margin:4px 0 0}
      .how-steps li{background:#f7f7f7!important;color:#111827!important;border:1px solid #e5e5e5!important;padding:10px 12px;border-radius:10px;font-weight:800}
      .how-steps span{background:#ce1126!important;color:#fff!important}
      .landing .nt-btn-primary{background:#ce1126!important;color:#fff!important}
      .landing .nt-btn:not(.nt-btn-primary){background:#fff!important;color:#ce1126!important;border:2px solid #ce1126!important}
    </style>
    <div class="landing page-in">
      <header class="landing-hero">
        <p class="landing-kicker">${esc(GAME_TITLE)} ONLINE</p>
        <h1>ULAR TANGGA<br/><span>NUSANTARA</span></h1>
        <p class="landing-tag">BELAJAR • BERMAIN • MENANG</p>
        <div class="landing-cta">
          <button type="button" class="nt-btn nt-btn-primary landing-play" data-play>MAIN SEKARANG</button>
          <button type="button" class="nt-btn" data-how>LIHAT CARA BERMAIN</button>
          ${renderInstallButton()}
        </div>
        <p class="landing-ver">${esc(GAME_VERSION)}</p>
      </header>

      <section class="landing-section nt-card">
        <h2>Fitur Utama</h2>
        <ul class="feature-grid">
          <li><span class="feat-icon">🎲</span><strong>Multiplayer Online</strong><p>Main bersama hingga 4 pemain.</p></li>
          <li><span class="feat-icon">📚</span><strong>Belajar Sambil Bermain</strong><p>PAI, Matematika, Bahasa Inggris, Bahasa Jawa.</p></li>
          <li><span class="feat-icon">🏆</span><strong>Ranked Mode</strong><p>Naikkan rank dan bersaing.</p></li>
          <li><span class="feat-icon">🎁</span><strong>Reward</strong><p>Dapatkan XP, coins, achievement.</p></li>
        </ul>
      </section>

      <section class="landing-section nt-card">
        <h2>Cara Bermain</h2>
        <ol class="how-steps">
          <li><span>1</span> LEMPAR DADU</li>
          <li><span>2</span> JAWAB SOAL</li>
          <li><span>3</span> HINDARI ULAR</li>
          <li><span>4</span> NAIK TANGGA</li>
          <li><span>5</span> CAPAI KOTAK 100</li>
        </ol>
      </section>

      <footer class="landing-footer nt-card">
        <h2>SIAP MENJADI JUARA?</h2>
        <button type="button" class="nt-btn nt-btn-primary" data-play>MAIN SEKARANG</button>
      </footer>
    </div>`;

  root.querySelectorAll("[data-play]").forEach((btn) => {
    btn.addEventListener("click", () => {
      playSfx("button");
      document.body.classList.remove("landing-active");
      opts.onPlay();
    });
  });
  root.querySelector("[data-how]")?.addEventListener("click", () => {
    playSfx("button");
    opts.onHow();
  });
  bindInstallButton(root);
}

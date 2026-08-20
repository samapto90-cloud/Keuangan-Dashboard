import { playSfx } from "../audio/manager";
import { closeModal, showModal, sparkle, spawnFloat } from "../ui/chrome";
import type { AchievementView, RewardEvent } from "./types";

export function showXpFloat(parent: HTMLElement, xp: number): void {
  if (xp > 0) spawnFloat(parent, `✨ +${xp} XP`, "good");
}

export function showCoinFloat(parent: HTMLElement, coins: number): void {
  if (coins > 0) spawnFloat(parent, `🪙 +${coins}`, "good");
}

export function showTrophyFloat(parent: HTMLElement, n: number): void {
  if (n > 0) spawnFloat(parent, `🏆 +${n} Piala`, "good");
}

export function showLevelUp(from: number, to: number): void {
  playSfx("levelup");
  const layer = showModal(
    "levelup",
    `<div class="lvl-up">
      <p class="nt-kicker">🎉 LEVEL UP!</p>
      <div class="lvl-glow">
        <span class="lvl-from">LEVEL ${from}</span>
        <span>→</span>
        <span class="lvl-to">LEVEL ${to}</span>
      </div>
      <button type="button" class="nt-btn nt-btn-primary" data-close>LANJUTKAN</button>
    </div>`,
  );
  layer.querySelectorAll(".spark-host").forEach((el) => sparkle(el as HTMLElement));
  sparkle(layer.querySelector(".lvl-glow") as HTMLElement);
  layer.querySelector("[data-close]")?.addEventListener("click", () => closeModal("levelup"));
}

export function showAchievementToast(ach: Pick<AchievementView, "name" | "rewardXp" | "rewardCoins">): void {
  playSfx("achievement");
  let host = document.querySelector<HTMLElement>(".ach-toast-host");
  if (!host) {
    host = document.createElement("div");
    host.className = "ach-toast-host";
    document.body.appendChild(host);
  }
  const el = document.createElement("div");
  el.className = "ach-toast";
  el.innerHTML = `<p>🏆 ACHIEVEMENT UNLOCKED</p><strong>${esc(ach.name)}</strong><p>+${ach.rewardXp} XP · +${ach.rewardCoins} Coins</p><button type="button" data-x>Tutup</button>`;
  host.appendChild(el);
  const close = (): void => el.remove();
  el.querySelector("[data-x]")?.addEventListener("click", close);
  window.setTimeout(close, 3200);
}

export function applyRewardFx(parent: HTMLElement | null, ev: RewardEvent | undefined, catalog?: { id: string; name: string; rewardXp: number; rewardCoins: number }[]): void {
  if (!ev) return;
  if (parent) {
    showXpFloat(parent, ev.xp || 0);
    showCoinFloat(parent, ev.coins || 0);
    showTrophyFloat(parent, ev.trophies || 0);
  }
  if (ev.levelUp && ev.levelAfter && ev.levelBefore) showLevelUp(ev.levelBefore, ev.levelAfter);
  for (const id of ev.achievements || []) {
    const found = catalog?.find((a) => a.id === id);
    showAchievementToast({ name: found?.name || id, rewardXp: found?.rewardXp || 0, rewardCoins: found?.rewardCoins || 0 });
  }
}

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

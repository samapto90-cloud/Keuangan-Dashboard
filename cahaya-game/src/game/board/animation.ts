import { resolveMove, walkPath, type MoveResult } from "./engine";
import { DICE_DURATION, LADDER_DURATION, MOVE_DURATION, SNAKE_DURATION } from "./config";
import { playSfx } from "../../audio/manager";
import { reducedMotion } from "../../ui/prefs";

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function faceFromDelta(token: HTMLElement, prevLeft: number, nextLeft: number): void {
  if (nextLeft < prevLeft - 0.12) {
    token.classList.add("face-left");
    token.classList.remove("face-right");
  } else if (nextLeft > prevLeft + 0.12) {
    token.classList.add("face-right");
    token.classList.remove("face-left");
  }
}

/** Tunggu frame berikutnya agar layout apply sebelum sleep transition. */
function nextFrame(): Promise<void> {
  return new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
}

export class BoardAnimationManager {
  async animateDice(el: HTMLElement, value: number): Promise<void> {
    playSfx("dice_roll");
    if (!reducedMotion()) {
      el.classList.add("is-rolling");
      el.textContent = "🌀";
      el.style.setProperty("--dice-spin-x", `${360 + value * 90}deg`);
      el.style.setProperty("--dice-spin-y", `${180 + value * 60}deg`);
    }
    await sleep(DICE_DURATION);
    el.classList.remove("is-rolling");
    el.dataset.face = String(value);
    el.textContent = String(value);
    if (!reducedMotion()) {
      el.classList.add("is-land-dice");
      await sleep(120);
      el.classList.remove("is-land-dice");
    }
  }

  async animateTokenMove(
    token: HTMLElement,
    path: number[],
    place: (pos: number) => void,
  ): Promise<void> {
    const steps = path.length ? path : [];
    if (!steps.length) return;

    if (reducedMotion()) {
      for (const pos of steps) {
        playSfx("token_step");
        place(pos);
        await sleep(30);
      }
      return;
    }

    token.classList.add("is-walking");
    token.classList.add("is-animating");
    token.style.transition = `left ${MOVE_DURATION}ms cubic-bezier(0.22, 0.8, 0.28, 1), bottom ${MOVE_DURATION}ms cubic-bezier(0.22, 0.8, 0.28, 1)`;
    let prevLeft = parseFloat(token.style.left) || 0;

    for (const pos of steps) {
      playSfx("token_step");
      place(pos);
      await nextFrame();
      const nextLeft = parseFloat(token.style.left) || 0;
      faceFromDelta(token, prevLeft, nextLeft);
      prevLeft = nextLeft;
      await sleep(MOVE_DURATION);
    }

    token.classList.remove("is-walking");
    token.classList.add("is-land");
    token.style.transition = `left 120ms ease-out, bottom 120ms ease-out`;
    await sleep(120);
    token.classList.remove("is-land", "is-animating");
    token.style.transition = "none";
  }

  async animateSnake(board: HTMLElement, token: HTMLElement, _from: number, to: number, place: (pos: number) => void): Promise<void> {
    playSfx("snake_slide");
    if (!reducedMotion()) board.classList.add("is-shake");
    token.classList.add("on-snake", "is-sliding", "is-animating");
    const prevLeft = parseFloat(token.style.left) || 0;
    token.style.transition = `left ${SNAKE_DURATION}ms cubic-bezier(0.4, 0.05, 0.55, 1), bottom ${SNAKE_DURATION}ms cubic-bezier(0.4, 0.05, 0.55, 1)`;
    place(to);
    await nextFrame();
    faceFromDelta(token, prevLeft, parseFloat(token.style.left) || 0);
    await sleep(SNAKE_DURATION);
    token.style.transition = "none";
    token.classList.remove("on-snake", "is-sliding", "is-animating");
    board.classList.remove("is-shake");
  }

  async animateLadder(token: HTMLElement, to: number, place: (pos: number) => void): Promise<void> {
    playSfx("ladder_climb");
    token.classList.add("on-ladder", "is-climbing", "is-animating");
    token.style.transition = `left ${LADDER_DURATION}ms cubic-bezier(0.25, 0.7, 0.25, 1), bottom ${LADDER_DURATION}ms cubic-bezier(0.25, 0.7, 0.25, 1)`;
    place(to);
    await nextFrame();
    await sleep(LADDER_DURATION);
    token.style.transition = "none";
    token.classList.remove("on-ladder", "is-climbing", "is-animating");
  }

  async animatePenalty(token: HTMLElement, path: number[], place: (pos: number) => void): Promise<void> {
    playSfx("penalty");
    token.classList.add("is-penalty");
    await this.animateTokenMove(token, path.slice(1), place);
    token.classList.remove("is-penalty");
  }

  async animateTurnTransition(bar: HTMLElement): Promise<void> {
    playSfx("turn_change");
    bar.classList.add("turn-swap");
    await sleep(180);
    bar.classList.remove("turn-swap");
  }
}

export { resolveMove, walkPath };
export type { MoveResult };

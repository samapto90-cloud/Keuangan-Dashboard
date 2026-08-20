import { loadPrefs } from "../ui/prefs";

export type SfxName =
  | "dice_roll"
  | "token_step"
  | "snake_slide"
  | "ladder_climb"
  | "turn_change"
  | "finish"
  | "question_open"
  | "timer_tick"
  | "correct"
  | "wrong"
  | "timeout"
  | "penalty"
  | "win"
  | "button"
  | "levelup"
  | "coin"
  | "achievement"
  | "notification";

const freqs: Record<SfxName, number> = {
  dice_roll: 220,
  token_step: 520,
  snake_slide: 140,
  ladder_climb: 660,
  turn_change: 400,
  finish: 880,
  question_open: 490,
  timer_tick: 610,
  correct: 740,
  wrong: 180,
  timeout: 160,
  penalty: 200,
  win: 980,
  button: 380,
  levelup: 1040,
  coin: 720,
  achievement: 860,
  notification: 440,
};

let ctx: AudioContext | null = null;
let musicNodes: { osc: OscillatorNode; gain: GainNode }[] = [];
let musicOn = false;

function context(): AudioContext | null {
  try {
    const AC = window.AudioContext || (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
    if (!AC) return null;
    if (!ctx) ctx = new AC();
    return ctx;
  } catch {
    return null;
  }
}

function masterMul(): number {
  const p = loadPrefs();
  if (p.mute) return 0;
  return Math.max(0, Math.min(1, p.master));
}

function sfxGain(): number {
  const p = loadPrefs();
  if (p.mute) return 0;
  return 0.05 * p.sfx * masterMul();
}

function musicGain(): number {
  const p = loadPrefs();
  if (p.mute) return 0;
  return 0.028 * p.music * masterMul();
}

export function playSfx(name: SfxName): void {
  const audio = context();
  if (!audio) return;
  void audio.resume().catch(() => undefined);
  const vol = sfxGain();
  if (vol <= 0.0001) return;
  try {
    const osc = audio.createOscillator();
    const gain = audio.createGain();
    osc.frequency.value = freqs[name];
    osc.type = name === "snake_slide" || name === "wrong" ? "sawtooth" : name === "win" ? "square" : "triangle";
    gain.gain.value = vol;
    osc.connect(gain);
    gain.connect(audio.destination);
    osc.start();
    const hold = name === "win" ? 0.28 : 0.12;
    gain.gain.exponentialRampToValueAtTime(0.001, audio.currentTime + hold);
    osc.stop(audio.currentTime + hold + 0.04);
  } catch {
    /* no crash if audio blocked */
  }
}

export function startMusic(): void {
  const audio = context();
  if (!audio || musicOn) return;
  void audio.resume().catch(() => undefined);
  const vol = musicGain();
  if (vol <= 0.0001) return;
  try {
    const notes = [196, 247, 294, 330];
    musicNodes = notes.map((f, i) => {
      const osc = audio.createOscillator();
      const gain = audio.createGain();
      osc.type = "sine";
      osc.frequency.value = f;
      gain.gain.value = vol * (i === 0 ? 1 : 0.45);
      osc.connect(gain);
      gain.connect(audio.destination);
      osc.start();
      return { osc, gain };
    });
    musicOn = true;
  } catch {
    musicOn = false;
  }
}

export function stopMusic(): void {
  for (const n of musicNodes) {
    try {
      n.osc.stop();
    } catch {
      /* ignore */
    }
  }
  musicNodes = [];
  musicOn = false;
}

export function syncMusic(): void {
  const vol = musicGain();
  if (vol <= 0.0001) {
    stopMusic();
    return;
  }
  if (!musicOn) startMusic();
  else musicNodes.forEach((n, i) => {
    n.gain.gain.value = vol * (i === 0 ? 1 : 0.45);
  });
}

export function applyAudioPrefs(): void {
  syncMusic();
}

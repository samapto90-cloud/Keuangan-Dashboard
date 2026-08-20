export const Motion = {
  FAST: 120,
  NORMAL: 220,
  SLOW: 280,
  MODAL: 280,
} as const;

const KEY = "ular-prefs-v2";

export type Quality = "low" | "medium" | "high";
export type ThemeMode = "system" | "light" | "dark";
export type FontSize = "small" | "normal" | "large";

export type UiPrefs = {
  master: number;
  music: number;
  sfx: number;
  mute: boolean;
  reduced: boolean;
  quality: Quality;
  theme: ThemeMode;
  fontSize: FontSize;
  language: "id";
  tutorialCompleted: boolean;
  avatar: string;
  nickname: string;
  notifications: boolean;
};

const defaults: UiPrefs = {
  master: 1,
  music: 0.4,
  sfx: 0.7,
  mute: false,
  reduced: false,
  quality: detectQuality(),
  theme: "dark",
  fontSize: "normal",
  language: "id",
  tutorialCompleted: false,
  avatar: "🦅",
  nickname: "",
  notifications: true,
};

let cache: UiPrefs | null = null;
let qualityUserSet = false;

function detectQuality(): Quality {
  try {
    const mem = (navigator as Navigator & { deviceMemory?: number }).deviceMemory;
    const cores = navigator.hardwareConcurrency || 4;
    if (mem !== undefined && mem <= 2) return "low";
    if (cores <= 4) return "medium";
  } catch {
    /* ignore */
  }
  return "high";
}

export function loadPrefs(): UiPrefs {
  if (cache) return cache;
  try {
    const raw = localStorage.getItem(KEY);
    if (raw) {
      const parsed = JSON.parse(raw) as Partial<UiPrefs> & { landingSeen?: boolean };
      cache = { ...defaults, ...parsed };
      if (parsed.quality) qualityUserSet = true;
    }
  } catch {
    cache = { ...defaults };
  }
  if (!cache) cache = { ...defaults };
  if (!qualityUserSet) cache.quality = detectQuality();
  applyDoc(cache);
  return cache;
}

export function savePrefs(patch: Partial<UiPrefs>): UiPrefs {
  const next = { ...loadPrefs(), ...patch };
  if (patch.quality) qualityUserSet = true;
  cache = next;
  try {
    localStorage.setItem(KEY, JSON.stringify(next));
  } catch {
    /* ignore */
  }
  applyDoc(next);
  return next;
}

function applyDoc(p: UiPrefs): void {
  document.documentElement.dataset.quality = p.quality;
  document.documentElement.dataset.reduced = reducedMotion() ? "1" : "0";
  document.documentElement.dataset.font = p.fontSize;
  document.documentElement.dataset.theme = resolveTheme(p.theme);
}

function resolveTheme(mode: ThemeMode): "light" | "dark" {
  if (mode === "light") return "light";
  if (mode === "dark") return "dark";
  try {
    return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
  } catch {
    return "dark";
  }
}

export function reducedMotion(): boolean {
  const p = cache || loadPrefs();
  try {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return true;
  } catch {
    /* ignore */
  }
  return p.reduced;
}

export function duration(ms: number): number {
  return reducedMotion() ? Math.min(ms, 80) : ms;
}

export function particleScale(): number {
  const q = (cache || loadPrefs()).quality;
  if (q === "low") return 0.35;
  if (q === "medium") return 0.7;
  return 1;
}

// Listen for system theme changes
try {
  window.matchMedia("(prefers-color-scheme: light)").addEventListener("change", () => {
    if (cache?.theme === "system") applyDoc(cache);
  });
} catch {
  /* ignore */
}

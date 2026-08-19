import type { PlayerFormId } from "../player/PlayerForms";

export interface FormView {
  id: string;
  name: string;
  shortName: string;
  visual?: string;
  unlocked: boolean;
  active: boolean;
  energyCost?: number;
  level?: number;
  storyName?: string;
  passive?: string;
  mastery?: number;
}

export interface SkillNodeView {
  id: string;
  skillId: string;
  branch: string;
  branchName?: string;
  cost: number;
  requiredLevel: number;
  prerequisite?: string;
  unlocked: boolean;
  available?: boolean;
  name?: string;
  description?: string;
  energyCost?: number;
  cooldown?: number;
  effect?: string;
  damage?: number;
  range?: number;
}

export interface CombatBuildView {
  id: string;
  slot: number;
  name: string;
  style: string;
  formId: string;
  score: number;
  active: boolean;
}

export interface ProgressionView {
  playerId: string;
  level: number;
  exp: number;
  expToNext: number;
  attributePoints: number;
  skillPoints: number;
  spentStr: number;
  spentDef: number;
  spentAgi: number;
  spentEng: number;
  spentVit: number;
  unlockedSkills: string[];
  unlockedForms: string[];
  formId: string;
  transformState: string;
  transEnergy: number;
  maxTransEnergy: number;
  powerRating: number;
  forms: FormView[];
  nodes: SkillNodeView[];
  transformReady: boolean;
  combatStyle?: string;
  styleMastery?: Record<string, number>;
  formMastery?: Record<string, number>;
  ultCharge?: number;
  maxUltCharge?: number;
  loadout?: string[];
  builds?: CombatBuildView[];
  activeBuild?: number;
  combatRating?: number;
  attrResetLeft?: number;
  training?: { dps: number; hits: number; damage: number; combo: number };
  styles?: Array<{ id: string; name: string; focus: string }>;
}

export interface TransformView {
  playerId: string;
  formId: string;
  name: string;
  visual: string;
  state: string;
  energy: number;
  maxEnergy: number;
  until: number;
  auraColor?: string;
  particles?: string;
  reason?: string;
}

const VISUAL: Record<string, PlayerFormId> = {
  normal: "asal",
  asal: "asal",
  "aura-1": "cahaya",
  cahaya: "cahaya",
  "aura-2": "kilat",
  kilat: "kilat",
  "aura-3": "agung",
  agung: "agung",
  "celestial-4": "fajar",
  fajar: "fajar",
};

export function visualForm(formId?: string, visual?: string): PlayerFormId {
  return VISUAL[visual || ""] ?? VISUAL[formId || ""] ?? "asal";
}

export class ProgressionStore {
  view: ProgressionView | null = null;
  transform: TransformView | null = null;
  selectedForm = "aura-1";

  apply(view: ProgressionView): void {
    this.view = view;
    if (view.formId && view.formId !== "normal") this.selectedForm = view.formId;
  }

  applyTransform(view: TransformView): void {
    this.transform = view;
    if (this.view) {
      this.view.formId = view.formId;
      this.view.transformState = view.state;
      this.view.transEnergy = view.energy;
      this.view.maxTransEnergy = view.maxEnergy;
    }
  }

  formLabel(): string {
    const id = this.view?.formId || "normal";
    if (id === "aura-1") return "ASCENSION I";
    if (id === "aura-2") return "ASCENSION II";
    if (id === "aura-3") return "ASCENSION III";
    if (id === "celestial-4") return "ASCENSION IV";
    return this.view?.forms.find((f) => f.id === id)?.shortName || "NORMAL";
  }

  readyLabel(): string {
    const state = this.view?.transformState || "NORMAL";
    if (state === "TRANSFORMING" || state === "TRANSFORMED") return "ACTIVE";
    if (this.view?.transformReady) return "READY";
    const selected = this.view?.forms.find((f) => f.id === this.selectedForm);
    if (selected && !selected.unlocked) return "LOCKED";
    return "LOCKED";
  }
}

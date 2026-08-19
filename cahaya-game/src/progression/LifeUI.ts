import { escapeHtml } from "../dialogue/DialogueUI";
import type { LifeState, LifeSkillView, PetView } from "../network/NetworkMessage";

export class LifeUI {
  readonly root: HTMLElement;
  blocking = false;
  tab: "skills" | "pets" | "daily" | "collect" | "quiz" = "skills";
  view: LifeState | null = null;
  selectedPet = "";
  onClaimPet: ((id: string) => void) | null = null;
  onSummon: ((id: string) => void) | null = null;
  onCare: ((id: string, action: string) => void) | null = null;
  onDaily: ((id: string) => void) | null = null;
  onCollect: ((id: string) => void) | null = null;
  onQuiz: ((choice: number) => void) | null = null;
  onRefresh: (() => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay life-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("click", (e) => this.onClick(e));
  }

  apply(view: LifeState): void {
    this.view = view;
    if (!this.selectedPet) this.selectedPet = view.activePet || view.pets?.find((p) => p.owned)?.petId || "dawn-pup";
    if (this.blocking) this.render();
  }

  toggle(): void {
    if (this.blocking) this.close();
    else this.open();
  }

  open(): void {
    this.blocking = true;
    this.root.hidden = false;
    this.onRefresh?.();
    this.render();
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }

  render(): void {
    if (!this.blocking) return;
    const v = this.view;
    const skills = (v?.skills ?? []).map((s) => skillRow(s)).join("");
    const pets = (v?.pets ?? []).map((p) =>
      `<button type="button" class="${p.petId === this.selectedPet ? "on" : ""}" data-life="pet" data-id="${escapeHtml(p.petId)}">${escapeHtml(p.name)} · ${p.owned ? p.mood || "CALM" : "LOCKED"}</button>`).join("");
    const cur = (v?.pets ?? []).find((p) => p.petId === this.selectedPet);
    const dailies = (v?.dailies ?? []).map((d) =>
      `<div class="social-row"><span>${escapeHtml(d.title)} ${d.ready ? "· siap" : ""}</span><button type="button" data-life="daily" data-id="${escapeHtml(d.id)}" ${d.claimed || !d.ready ? "disabled" : ""}>${d.claimed ? "CLAIMED" : "CLAIM"}</button></div>`).join("");
    const cols = ["fish", "plants", "recipes", "furniture", "pets"].map((id) =>
      `<button type="button" data-life="col" data-id="${id}">${id.toUpperCase()}${v?.collectionClaimed?.[id] ? " ✓" : ""}</button>`).join("");
    this.root.innerHTML = `
      <div class="rpg-card craft-card">
        <header><h3>LIFE · LV ${v?.lifeLevel ?? 1}</h3><button type="button" data-life="close">Close</button></header>
        <p>Farming ${v?.farmingLevel ?? 1}/30 · plots ${v?.plots ?? 4} · UTC ${escapeHtml(v?.utcDay || "")}</p>
        <div class="craft-tabs">
          <button type="button" data-life="tab" data-id="skills" class="${this.tab === "skills" ? "on" : ""}">LIFE SKILL</button>
          <button type="button" data-life="tab" data-id="pets" class="${this.tab === "pets" ? "on" : ""}">PET COLLECTION</button>
          <button type="button" data-life="tab" data-id="daily" class="${this.tab === "daily" ? "on" : ""}">DAILY LIFE</button>
          <button type="button" data-life="tab" data-id="collect" class="${this.tab === "collect" ? "on" : ""}">COLLECTION</button>
          <button type="button" data-life="tab" data-id="quiz" class="${this.tab === "quiz" ? "on" : ""}">EDU</button>
        </div>
        ${this.tab === "skills" ? `<div class="craft-list">${skills || "<p>Life skills siap.</p>"}</div>` : ""}
        ${this.tab === "pets" ? `
          <div class="craft-grid">${pets}</div>
          <p>${escapeHtml(cur?.name || "")} · happy ${cur?.happiness ?? 0} · ${escapeHtml(cur?.mood || "")}</p>
          <div class="inv-actions">
            ${cur?.owned ? `
              <button type="button" data-life="summon" data-id="${escapeHtml(cur.petId)}">SUMMON</button>
              <button type="button" data-life="care" data-id="feed">FEED</button>
              <button type="button" data-life="care" data-id="play">PLAY</button>
              <button type="button" data-life="care" data-id="rest">REST</button>
              <button type="button" data-life="care" data-id="pet">PET</button>
            ` : `<button type="button" data-life="claim" data-id="${escapeHtml(cur?.petId || "dawn-pup")}">CLAIM</button>`}
          </div>
          <p>Companion only. Bukan combat pet. Tanpa gacha.</p>
        ` : ""}
        ${this.tab === "daily" ? `<div class="craft-list">${dailies}<p>Reset UTC. Tampil sesuai zona waktu perangkat.</p></div>` : ""}
        ${this.tab === "collect" ? `<div class="inv-actions">${cols}</div><p>Hadiah: title, cosmetic, Knowledge Token.</p>` : ""}
        ${this.tab === "quiz" ? `
          <p>Le, yen ana 4 kursi banjur ditambah 2, dadi pira?</p>
          <div class="inv-actions">
            <button type="button" data-life="quiz" data-id="0">A. 5</button>
            <button type="button" data-life="quiz" data-id="1">B. 6</button>
            <button type="button" data-life="quiz" data-id="2">C. 7</button>
          </div>
        ` : ""}
      </div>`;
  }

  private onClick(e: MouseEvent): void {
    const t = (e.target as HTMLElement).closest("[data-life]") as HTMLElement | null;
    if (!t) return;
    const id = t.dataset.id || "";
    if (t.dataset.life === "close") this.close();
    if (t.dataset.life === "tab") {
      this.tab = id as typeof this.tab;
      this.render();
    }
    if (t.dataset.life === "pet") {
      this.selectedPet = id;
      this.render();
    }
    if (t.dataset.life === "claim") this.onClaimPet?.(id);
    if (t.dataset.life === "summon") this.onSummon?.(id);
    if (t.dataset.life === "care") this.onCare?.(this.selectedPet, id);
    if (t.dataset.life === "daily") this.onDaily?.(id);
    if (t.dataset.life === "col") this.onCollect?.(id);
    if (t.dataset.life === "quiz") this.onQuiz?.(Number(id));
  }
}

function skillRow(s: LifeSkillView): string {
  return `<div class="social-row"><span>${escapeHtml(s.name)}</span><span>Lv ${s.level}/${s.max} · XP ${s.xp}</span></div>`;
}

export function petLabel(p: PetView): string {
  return `${p.name} (${p.mood || "CALM"})`;
}

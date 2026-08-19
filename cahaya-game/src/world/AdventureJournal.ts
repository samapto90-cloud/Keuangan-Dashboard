import type { WorldJournal, PlayerProgress, LoreView } from "../network/NetworkMessage";
import { GuardianJournal } from "./GuardianJournal";
import { LoreJournal } from "./LoreJournal";

type JournalTab = "STORY" | "QUEST" | "LORE" | "NPC" | "CHOICES" | "MAIN STORY" | "SIDE STORY" | "SILUMAN CODEX" | "CHARACTERS" | "DISCOVERIES" | "STORY ARCHIVE" | "GUARDIANS" | "REGIONS" | "COLLECTIONS";

export class AdventureJournal {
  readonly root: HTMLDivElement;
  blocking = false;
  private tab: JournalTab = "STORY";
  private readonly guardians = new GuardianJournal();
  private readonly lore = new LoreJournal();
  journal: WorldJournal | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay journal-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  toggle(progress: PlayerProgress | null): void {
    if (this.blocking) this.close();
    else this.open(progress);
  }

  open(progress: PlayerProgress | null): void {
    this.blocking = true;
    this.root.hidden = false;
    this.render(progress);
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }

  apply(j: WorldJournal): void {
    this.journal = j;
    if (this.blocking) this.render(null);
  }

  private render(progress: PlayerProgress | null): void {
    const j = this.journal;
    const tabs: JournalTab[] = ["STORY", "QUEST", "LORE", "NPC", "CHOICES", "SILUMAN CODEX"];
    let body = "";
    if (this.tab === "STORY" || this.tab === "MAIN STORY") {
      const overlay = (j?.overlayChapters ?? []).map((c) => `<div class="journal-row"><b>CH${c.index} ${c.title}</b><span>${c.locked ? "LOCKED" : c.state}</span></div>`).join("");
      const quests = progress?.quests?.filter((q) => q.kind === "main") ?? [];
      body = `<p>${j?.objective ?? "Temukan jalan menuju Masjid Cahaya."}</p><p class="journal-mood">${j?.worldMood ?? ""}</p>${overlay}` + quests.map((q) => `<div class="journal-row"><b>${q.title}</b><span>${q.state}</span><p>${q.description}</p></div>`).join("");
    } else if (this.tab === "QUEST" || this.tab === "SIDE STORY") {
      const quests = progress?.quests ?? [];
      body = quests.map((q) => `<div class="journal-row"><b>${q.title}</b><span>${q.state}</span><p>${q.description}</p></div>`).join("") || "<p>Bantu desa, belajar, dan jelajahi.</p>";
    } else if (this.tab === "LORE") {
      const cards = (j?.loreCards ?? j?.lore ?? []) as LoreView[];
      body = this.lore.render(cards);
    } else if (this.tab === "NPC") {
      body = (j?.npcBook ?? []).map((n) => `<div class="journal-row"><b>${n.name}</b><span>${n.relationship}</span><p>${n.role || ""} · ${n.nextReward || ""} ${n.memory ? "· ingat bantuanmu" : ""}</p></div>`).join("") || "<p>Kenali orang desa.</p>";
    } else if (this.tab === "CHOICES") {
      const rows = (j?.choiceHistory ?? []).map((c) => `<div class="journal-row"><b>${c.id}</b><span>${c.choice}</span><p>${c.impact || "Pilihanmu memengaruhi hubungan dan cerita."}</p></div>`).join("");
      body = `<p>Pilihanmu memengaruhi hubungan dan cerita.</p>${rows || "<p>Belum ada pilihan besar.</p>"}`;
    } else if (this.tab === "SILUMAN CODEX" || this.tab === "GUARDIANS") {
      const folk = (j?.enemyLore ?? []).map((e) => `<div class="journal-row"><b>${e.discovered ? e.name : "????"}</b><span>${e.defeated ? "DEFEATED" : e.encountered ? "MET" : "HIDDEN"}</span><p>${e.discovered ? `${e.region} · ${e.mechanic}` : "Belum ketemu."}</p></div>`).join("");
      body = this.guardians.render(j?.guardians ?? [], j?.tokens ?? 0) + folk;
    } else if (this.tab === "CHARACTERS") {
      body = this.lore.render(j?.lore ?? []);
    } else if (this.tab === "REGIONS") {
      body = (j?.regions ?? []).map((r) => `<div class="journal-row"><b>${r.discovered ? r.name : "????"}</b><span>${r.completion}%</span></div>`).join("") || "<p>Jelajahi dunia.</p>";
    } else if (this.tab === "DISCOVERIES") {
      body = (j?.landmarks ?? []).map((l) => `<div class="journal-row"><b>${l.name}</b><span>${l.discovered ? "FOUND" : "HIDDEN"}</span></div>`).join("") || "<p>Belum ada discovery.</p>";
    } else {
      body = `<p>Guardian ${j?.guardiansDefeated ?? 0}/33</p>`;
    }
    this.root.innerHTML = `<div class="rpg-card journal-card"><h3>PETUALANGAN LORE</h3><div class="journal-tabs">${tabs.map((t) => `<button type="button" data-tab="${t}" class="${t === this.tab ? "on" : ""}">${t}</button>`).join("")}</div><div class="journal-body">${body}</div><button type="button" class="journal-close">Tutup</button></div>`;
    this.root.querySelectorAll("[data-tab]").forEach((el) => {
      el.addEventListener("click", () => {
        this.tab = (el as HTMLElement).dataset.tab as JournalTab;
        this.render(progress);
      });
    });
    this.root.querySelector(".journal-close")?.addEventListener("click", () => this.close());
  }
}

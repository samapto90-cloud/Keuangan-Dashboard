import type { LandmarkView } from "../network/NetworkMessage";

const TIPS = [
  { jv: "Le, aja lali njelajah panggonan anyar.", id: "Jangan lupa menjelajahi tempat baru." },
  { jv: "Arep lelungan menyang kene?", id: "Apakah ingin melakukan perjalanan ke sini?" },
];

export class FastTravel {
  readonly root: HTMLDivElement;
  blocking = false;
  onTravel: ((id: string) => void) | null = null;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay travel-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
  }

  open(landmarks: LandmarkView[]): void {
    const found = landmarks.filter((l) => l.discovered);
    this.blocking = true;
    this.root.hidden = false;
    this.root.innerHTML = `<div class="rpg-card travel-card"><h3>FAST TRAVEL</h3><p>Pilih waypoint. Mount akan disimpan.</p><div class="travel-list"></div><button type="button" class="travel-cancel">CANCEL</button></div>`;
    const list = this.root.querySelector(".travel-list") as HTMLElement;
    for (const lm of found) {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.textContent = `TRAVEL HERE · ${lm.name}`;
      btn.addEventListener("click", () => this.confirm(lm.id, lm.name));
      list.appendChild(btn);
    }
    if (!found.length) list.innerHTML = "<p>Belum ada landmark ditemukan.</p>";
    this.root.querySelector(".travel-cancel")?.addEventListener("click", () => this.close());
  }

  confirm(id: string, name = ""): void {
    this.blocking = true;
    this.root.hidden = false;
    const tip = TIPS[0];
    this.root.innerHTML = `<div class="rpg-card travel-card">
      <h3>FAST TRAVEL</h3>
      <p class="travel-jv">Arep lelungan menyang kene?</p>
      <p class="dlg-sub">Apakah ingin melakukan perjalanan ke sini?</p>
      <p>${name}</p>
      <p class="travel-tip">${tip.jv}<br/><small>${tip.id}</small></p>
      <div class="inv-actions">
        <button type="button" class="travel-yes">TRAVEL HERE</button>
        <button type="button" class="travel-cancel">CANCEL</button>
      </div>
    </div>`;
    this.root.querySelector(".travel-yes")?.addEventListener("click", () => {
      this.onTravel?.(id);
      this.close();
    });
    this.root.querySelector(".travel-cancel")?.addEventListener("click", () => this.close());
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }
}

import type { LoreView } from "../network/NetworkMessage";

export class LoreJournal {
  private overlay: HTMLDivElement | null = null;
  private host: HTMLElement | null = null;
  private onKey: ((e: KeyboardEvent) => void) | null = null;

  attach(host: HTMLElement): void {
    this.host = host;
  }

  render(list: LoreView[]): string {
    if (!list.length) return "<p>Belum ada lore ditemukan.</p>";
    return list.map((l) => {
      const name = l.discovered === false ? "????" : (l.title || "Lore");
      const desc = l.discovered === false ? "Belum ditemukan." : (l.text || "");
      return `<div class="journal-row lore-card"><b>${name}</b><span>${l.discovered === false ? "HIDDEN" : "FOUND"}</span><p>${l.region || "dunia"} · ${desc}</p></div>`;
    }).join("");
  }

  popup(host: HTMLElement, title: string, text: string): void {
    this.attach(host);
    this.close();
    const el = document.createElement("div");
    el.className = "lore-popup";
    el.setAttribute("role", "dialog");
    el.innerHTML = `<div class="lore-popup-card">
      <strong>LORE DISCOVERED</strong>
      <h4>${escapeLore(title)}</h4>
      <p>${escapeLore(text)}</p>
      <button type="button" data-lore="close">Tutup</button>
    </div>`;
    const root = this.host ?? host;
    root.appendChild(el);
    this.overlay = el;
    const dismiss = (ev?: Event): void => {
      ev?.preventDefault();
      ev?.stopPropagation();
      this.close();
    };
    el.addEventListener("click", (e) => {
      const t = e.target as HTMLElement;
      if (t === el || t.closest("[data-lore=close]")) dismiss(e);
    });
    this.onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape" || e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        this.close();
      }
    };
    window.addEventListener("keydown", this.onKey);
  }

  close(): void {
    if (this.onKey) {
      window.removeEventListener("keydown", this.onKey);
      this.onKey = null;
    }
    this.overlay?.remove();
    this.overlay = null;
  }
}

function escapeLore(s: string): string {
  return s.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] || c));
}

export class DiscoverySystem {
  private until = 0;
  readonly root: HTMLDivElement;

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "discover-toast";
    this.root.hidden = true;
    this.root.setAttribute("role", "status");
    host.appendChild(this.root);
    this.root.addEventListener("click", () => this.hide());
  }

  show(name: string, exp: number, title = "REGION DISCOVERED"): void {
    this.root.hidden = false;
    this.root.innerHTML = `<strong>${esc(title)}</strong><span>${esc(name)}</span><em>+${exp} EXP</em><button type="button" class="discover-close">Tutup</button>`;
    this.until = performance.now() + 4200;
  }

  lore(title: string): void {
    this.root.hidden = false;
    this.root.innerHTML = `<strong>LORE DISCOVERED</strong><span>${esc(title)}</span><button type="button" class="discover-close">Tutup</button>`;
    this.until = performance.now() + 3600;
  }

  landmark(name: string): void {
    this.root.hidden = false;
    this.root.innerHTML = `<strong>NEW DISCOVERY</strong><span>${esc(name)}</span><button type="button" class="discover-close">Tutup</button>`;
    this.until = performance.now() + 3600;
  }

  hide(): void {
    this.root.hidden = true;
    this.until = 0;
  }

  update(): void {
    if (!this.root.hidden && this.until > 0 && performance.now() > this.until) this.hide();
  }
}

function esc(s: string): string {
  return s.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] || c));
}

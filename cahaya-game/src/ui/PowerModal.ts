import {
  POWER_META,
  bagTotal,
  type PowerBag,
  type PowerKind,
} from "../game/powers";

export type PowerTarget = { id: string; username: string; isSelf?: boolean };

function esc(v: string): string {
  return v.replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch] || ch);
}

/** Panel inventory + pilih target item. */
export function openPowerInventory(
  host: HTMLElement,
  opts: {
    bag: PowerBag;
    targets: PowerTarget[];
    selfId: string;
    onUse: (kind: PowerKind, targetId: string) => void;
    onClose: () => void;
  },
): () => void {
  const prev = host.querySelector(".power-overlay");
  prev?.remove();

  const overlay = document.createElement("div");
  overlay.className = "power-overlay";
  const total = bagTotal(opts.bag);

  const kinds = (Object.keys(POWER_META) as PowerKind[]).filter((k) => opts.bag[k] > 0);

  overlay.innerHTML = `
    <div class="power-modal" role="dialog" aria-label="Inventory">
      <header class="power-head">
        <strong>🎒 Inventory</strong>
        <button type="button" class="power-close" data-act="close" aria-label="Tutup">✕</button>
      </header>
      ${total === 0
        ? `<p class="power-empty">Belum ada item. Ambil 💣 / ⚡ / ✈️ di kotak acak di papan (stok: 1 / 5 / 3).</p>`
        : `<p class="power-lead">Pilih item, lalu pilih target.</p>
           <div class="power-list" id="power-list"></div>
           <div class="power-targets" id="power-targets" hidden>
             <p class="power-lead">Target:</p>
             <div class="power-target-list"></div>
           </div>`}
    </div>`;
  host.appendChild(overlay);

  let selected: PowerKind | null = null;
  const list = overlay.querySelector("#power-list");
  const targetsWrap = overlay.querySelector<HTMLElement>("#power-targets");
  const targetList = overlay.querySelector<HTMLElement>(".power-target-list");

  const paintTargets = (kind: PowerKind): void => {
    if (!targetsWrap || !targetList) return;
    targetsWrap.hidden = false;
    const meta = POWER_META[kind];
    const rows = opts.targets.filter((t) => {
      if (t.id === opts.selfId) return meta.allowSelf;
      return meta.needsTarget;
    });
    targetList.innerHTML = rows
      .map(
        (t) =>
          `<button type="button" class="power-target-btn" data-tid="${esc(t.id)}">${t.id === opts.selfId ? "🙋 Saya" : esc(t.username)}</button>`,
      )
      .join("");
    targetList.querySelectorAll<HTMLButtonElement>("[data-tid]").forEach((btn) => {
      btn.addEventListener("click", () => {
        if (!selected) return;
        opts.onUse(selected, btn.dataset.tid || "");
        cleanup();
      });
    });
  };

  if (list) {
    list.innerHTML = kinds
      .map((k) => {
        const m = POWER_META[k];
        return `<button type="button" class="power-item" data-kind="${k}">
          <span class="power-ico">${m.icon}</span>
          <span class="power-name">${m.label} ×${opts.bag[k]}</span>
          <span class="power-hint">${m.hint}</span>
        </button>`;
      })
      .join("");
    list.querySelectorAll<HTMLButtonElement>("[data-kind]").forEach((btn) => {
      btn.addEventListener("click", () => {
        selected = btn.dataset.kind as PowerKind;
        list.querySelectorAll(".power-item").forEach((el) => el.classList.remove("is-on"));
        btn.classList.add("is-on");
        paintTargets(selected);
      });
    });
  }

  const cleanup = (): void => {
    overlay.remove();
    opts.onClose();
  };
  overlay.querySelector('[data-act="close"]')?.addEventListener("click", cleanup);
  overlay.addEventListener("click", (e) => {
    if (e.target === overlay) cleanup();
  });
  return cleanup;
}

export function powerGrantBanner(kind: PowerKind): string {
  const m = POWER_META[kind];
  return `${m.icon} Dapat ${m.label}! ${m.hint}`;
}

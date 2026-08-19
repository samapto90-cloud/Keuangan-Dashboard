import type { PartyView } from "../network/NetworkMessage";
import type { WorldJournal } from "../network/NetworkMessage";
import type { WorldEventView, WorldBossView } from "../network/NetworkMessage";

export class WorldMap {
  readonly root: HTMLDivElement;
  blocking = false;
  onTravel: ((id: string) => void) | null = null;
  private panX = 0;
  private panZ = 0;
  private zoom = 1;
  private drag = false;
  private lastX = 0;
  private lastY = 0;
  resFilter = "ALL";
  private lastJournal: WorldJournal | null = null;
  private lastPlayer = { x: 0, z: 0 };
  private lastParty: PartyView | null = null;
  private lastEvent: WorldEventView | null = null;
  private lastBoss: WorldBossView | null = null;
  private lastObjects: Array<{ x: number; z: number; kind: string }> = [];

  constructor(host: HTMLElement) {
    this.root = document.createElement("div");
    this.root.className = "rpg-overlay map-overlay";
    this.root.hidden = true;
    host.appendChild(this.root);
    this.root.addEventListener("pointerdown", (e) => {
      this.drag = true;
      this.lastX = e.clientX;
      this.lastY = e.clientY;
    });
    window.addEventListener("pointerup", () => {
      this.drag = false;
    });
    this.root.addEventListener("pointermove", (e) => {
      if (!this.drag) return;
      this.panX += (e.clientX - this.lastX) / 8;
      this.panZ += (e.clientY - this.lastY) / 8;
      this.lastX = e.clientX;
      this.lastY = e.clientY;
    });
    this.root.addEventListener("wheel", (e) => {
      e.preventDefault();
      this.zoom = Math.min(2.4, Math.max(0.6, this.zoom + (e.deltaY > 0 ? -0.1 : 0.1)));
    }, { passive: false });
  }

  toggle(journal: WorldJournal | null, player: { x: number; z: number }, party: PartyView | null, event: WorldEventView | null, boss?: WorldBossView | null, objects?: Array<{ x: number; z: number; kind: string }>): void {
    if (this.blocking) {
      this.close();
      return;
    }
    this.open(journal, player, party, event, boss, objects);
  }

  open(journal: WorldJournal | null, player: { x: number; z: number }, party: PartyView | null, event: WorldEventView | null, boss?: WorldBossView | null, objects?: Array<{ x: number; z: number; kind: string }>): void {
    this.lastJournal = journal;
    this.lastPlayer = player;
    this.lastParty = party;
    this.lastEvent = event;
    this.lastBoss = boss ?? null;
    this.lastObjects = objects ?? [];
    this.blocking = true;
    this.root.hidden = false;
    const regions = journal?.regions ?? [];
    const lms = clusterLandmarks(journal?.landmarks ?? []);
    const siluman = (journal?.guardians ?? []).filter((g) => g.status === "AVAILABLE" || g.status === "DEFEATED").slice(0, 8);
    const px = 180 + (player.x + this.panX) * (1.1 * this.zoom);
    const pz = 40 + (player.z + this.panZ) * (0.9 * this.zoom);
    this.root.innerHTML = `
      <div class="rpg-card map-card">
        <h3>WORLD OF DAWN</h3>
        <p>${journal?.objective ?? "Temukan jalan menuju Sanctum of Light."} · Guardian ${journal?.guardiansDefeated ?? 0}/${journal?.guardiansTotal ?? 33} · Token ${journal?.tokens ?? 0}/33 · Channel ${journal?.channel ?? "1"}</p>
        <div class="map-canvas" style="position:relative;height:300px;overflow:hidden;background:#132033;border-radius:12px;">
          ${regions.map((r, i) => `<div class="map-region" style="position:absolute;left:16px;top:${8 + i * 28}px;opacity:${r.discovered ? 1 : 0.28}">${r.discovered ? escapeMap(r.title || r.name) : "????"} ${r.unlocked ? "" : "🔒"} ${r.completion}%${r.recommendedLevel ? ` LV${r.recommendedLevel}` : ""}</div>`).join("")}
          ${lms.map((lm) => `<button type="button" data-lm="${lm.id}" class="map-pin" style="position:absolute;left:${180 + lm.x * 1.1}px;top:${40 + lm.z * 0.55}px">${lm.discovered ? waypointIcon(lm.id) : "◇"} ${lm.label ?? ""}</button>`).join("")}
          ${(journal?.markers ?? []).slice(0, 12).map((m) => `<div class="map-marker" style="position:absolute;left:${180 + m.x * 1.1}px;top:${40 + m.z * 0.55}px;font-size:10px">${markerIcon(m.kind)} ${escapeMap(m.name)}</div>`).join("")}
          ${(objects ?? []).filter((o) => resourceMatch(o.kind, this.resFilter)).slice(0, 16).map((o) => `<div class="map-marker" style="position:absolute;left:${180 + o.x * 1.1}px;top:${40 + o.z * 0.55}px;font-size:10px">${markerIcon(o.kind)}</div>`).join("")}
          ${siluman.map((g, i) => `<div class="map-siluman" style="position:absolute;right:10px;top:${8 + i * 16}px;font-size:11px">${String(g.index ?? i + 1).padStart(2, "0")} ${g.status === "DEFEATED" ? "✓" : "?"} ${g.name}</div>`).join("")}
          ${event ? `<div class="map-event" style="position:absolute;left:12px;bottom:28px">WORLD EVENT ${escapeMap(event.name)}${event.objective ? ` · ${escapeMap(event.objective)} ${event.progress ?? 0}/${event.need ?? 0}` : ""}</div>` : ""}
          ${boss ? `<div class="map-boss" style="position:absolute;left:${180 + (boss.x ?? 0) * 1.1}px;top:${40 + (boss.z ?? 58) * 0.55}px">☠</div>` : ""}
          ${boss ? `<div class="map-boss-label" style="position:absolute;left:12px;bottom:8px">WORLD BOSS ${boss.name} ${boss.hp}/${boss.maxHp}</div>` : ""}
          ${journal?.nextWorldBoss && !boss ? `<div class="map-boss-label" style="position:absolute;left:12px;bottom:8px">NEXT WORLD BOSS ${escapeMap(journal.nextWorldBoss.name)}</div>` : ""}
          ${party?.waypointId ? `<div class="map-party-wp" style="position:absolute;left:${180 + (party.waypointX ?? 0) * 1.1}px;top:${40 + (party.waypointZ ?? 0) * 0.55}px">⚑ PARTY</div>` : ""}
          ${party?.members?.map((m, i) => `<div class="map-party" style="position:absolute;right:8px;bottom:${8 + i * 18}px;color:${["#7dd3fc", "#fde68a", "#86efac", "#f9a8d4"][i % 4]}">${m.name || m.playerId}</div>`).join("") ?? ""}
          <div class="map-player" style="position:absolute;left:${px}px;top:${pz}px">●</div>
        </div>
        <div class="map-actions">
          <button type="button" data-rf="ALL">ALL</button>
          <button type="button" data-rf="Mining">Mining</button>
          <button type="button" data-rf="Wood">Wood</button>
          <button type="button" data-rf="Herb">Herb</button>
          <button type="button" data-rf="Fishing">Fishing</button>
          <button type="button" class="map-close">Tutup</button>
        </div>
      </div>`;
    this.root.querySelector(".map-close")?.addEventListener("click", () => this.close());
    this.root.querySelectorAll("[data-lm]").forEach((el) => {
      el.addEventListener("click", () => {
        const id = (el as HTMLElement).dataset.lm;
        if (id) this.onTravel?.(id);
      });
    });
    this.root.querySelectorAll("[data-rf]").forEach((el) => {
      el.addEventListener("click", () => {
        this.resFilter = (el as HTMLElement).dataset.rf || "ALL";
        this.open(this.lastJournal, this.lastPlayer, this.lastParty, this.lastEvent, this.lastBoss, this.lastObjects);
      });
    });
  }

  close(): void {
    this.blocking = false;
    this.root.hidden = true;
  }
}

function clusterLandmarks(list: { id: string; name: string; discovered: boolean; x: number; z: number }[]): { id: string; x: number; z: number; discovered: boolean; label?: string }[] {
  const out: { id: string; x: number; z: number; discovered: boolean; label?: string; n?: number }[] = [];
  for (const lm of list) {
    const hit = out.find((o) => Math.hypot(o.x - lm.x, o.z - lm.z) < 8);
    if (hit) {
      hit.n = (hit.n ?? 1) + 1;
      hit.label = String(hit.n);
      continue;
    }
    out.push({ id: lm.id, x: lm.x, z: lm.z, discovered: lm.discovered, n: 1 });
  }
  return out;
}

function waypointIcon(id: string): string {
  if (id.includes("waypoint") || id.includes("bridge") || id.includes("checkpoint")) return "◈";
  return "◆";
}

function markerIcon(kind: string): string {
  const k = (kind || "").toLowerCase();
  if (k.includes("event")) return "⚠";
  if (k.includes("boss") || k.includes("raid")) return "☠";
  if (k.includes("dungeon")) return "▣";
  if (k.includes("npc") || k.includes("quest")) return "◇";
  if (k.includes("guild")) return "♜";
  if (k.includes("gather-ore") || k.includes("mining") || k.includes("ore") || k.includes("stone")) return "⛏";
  if (k.includes("gather-wood") || k.includes("wood")) return "♣";
  if (k.includes("gather-herb") || k.includes("herb")) return "✿";
  if (k.includes("fishing") || k.includes("fish")) return "≋";
  if (k.includes("resource")) return "✱";
  if (k.includes("town")) return "⌂";
  return "•";
}

function resourceMatch(kind: string, filter: string): boolean {
  const k = kind.toLowerCase();
  if (filter === "ALL") return k.startsWith("gather-") || k.includes("fish") || k.includes("resource");
  if (filter === "Mining") return k.includes("ore") || k.includes("stone") || k.includes("mining");
  if (filter === "Wood") return k.includes("wood");
  if (filter === "Herb") return k.includes("herb") || k.includes("fiber");
  if (filter === "Fishing") return k.includes("fish");
  return false;
}

function escapeMap(s: string): string {
  return s.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c] || c));
}

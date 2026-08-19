import { escapeHtml } from "../dialogue/DialogueUI";
import type { HouseState } from "../network/NetworkMessage";
import type { LoadoutStore } from "../inventory/LoadoutStore";

const CATS = ["FURNITURE", "LIGHTING", "DECORATION", "PLANTS", "DISPLAY"];

export function housingHtml(state: HouseState | null, store: LoadoutStore): string {
  const inside = !!state?.instanceId && !state.left;
  const items = (state?.items ?? []).map((it) => {
    const def = store.item(it.itemId);
    return `<div class="social-row">
      <span>${escapeHtml(def?.name || it.itemId)} @ ${it.x.toFixed(2)}, ${it.z.toFixed(2)}</span>
      <button type="button" data-soc="house-remove" data-id="${escapeHtml(it.id)}">REMOVE</button>
    </div>`;
  }).join("") || "<p>Belum ada dekorasi.</p>";
  const decor = store.slots.filter((s) => s.item?.type === "DECOR").map((s) =>
    `<button type="button" data-soc="house-place" data-slot="${s.index}">Place ${escapeHtml(s.item?.name || "")}</button>`).join("");
  const visitors = (state?.visitors ?? []).map((id) => `<span>${escapeHtml(id)}</span>`).join(" · ") || "—";
  const guests = (state?.guestLog ?? []).map((g) => `${escapeHtml(g.playerId)}`).join(" · ") || "—";
  const plots = (state?.plots ?? []).map((p) =>
    `<div class="social-row"><span>${escapeHtml(p.id)} · ${escapeHtml(p.plant || "empty")} · ${escapeHtml(p.state || "EMPTY")}</span>
      <button type="button" data-soc="hcmd" data-cmd="plant" data-id="${escapeHtml(p.id)}">PLANT</button>
      <button type="button" data-soc="hcmd" data-cmd="water" data-id="${escapeHtml(p.id)}">WATER</button>
      <button type="button" data-soc="hcmd" data-cmd="harvest" data-id="${escapeHtml(p.id)}">HARVEST</button>
    </div>`).join("");
  const loc = (state?.locations ?? [{ id: "village", name: "DAWN VILLAGE" }, { id: "forest", name: "MISTWOOD" }, { id: "plains", name: "RIVER OF LIGHT" }])
    .map((l) => escapeHtml(l.name)).join(" · ");
  const rooms = (state?.rooms ?? []).map((r) => escapeHtml(r.name)).join(" · ");
  return `
    <h4>${escapeHtml(state?.sign || state?.name || "DAWN MEADOW HOME")}</h4>
    <p>${escapeHtml(state?.type || "SMALL")} · ${escapeHtml(state?.access || "PRIVATE")} · ${state?.locked ? "LOCKED" : "OPEN"} · grid ${state?.grid ?? 0.25}m</p>
    <p>Lokasi: ${loc}</p>
    <p>Interior: ${rooms || "Living Room · Rest Area · Display Area · Storage Area · Garden Area"}</p>
    <p>Visitors: ${visitors}</p>
    <p>Guest log: ${guests}</p>
    <div class="inv-actions">
      <button type="button" data-soc="house-enter">ENTER HOUSE</button>
      <button type="button" data-soc="house-leave">EXIT HOUSE</button>
      <button type="button" data-soc="hcmd" data-cmd="hall">GUILD HALL</button>
      <button type="button" data-soc="house-access" data-id="FRIENDS">FRIENDS</button>
      <button type="button" data-soc="house-access" data-id="GUILD">GUILD</button>
      <button type="button" data-soc="house-access" data-id="PUBLIC">PUBLIC</button>
      <button type="button" data-soc="house-access" data-id="PRIVATE">PRIVATE</button>
      <button type="button" data-soc="hcmd" data-cmd="lock">${state?.locked ? "UNLOCK" : "LOCK"}</button>
    </div>
    <h4>CUSTOMIZATION</h4>
    <div class="inv-actions">
      <button type="button" data-soc="hcmd" data-cmd="style" data-id="wall:dawn">WALL DAWN</button>
      <button type="button" data-soc="hcmd" data-cmd="style" data-id="floor:wood">FLOOR WOOD</button>
      <button type="button" data-soc="hcmd" data-cmd="style" data-id="roof:tile">ROOF TILE</button>
      <button type="button" data-soc="hcmd" data-cmd="style" data-id="light:warm">LIGHT WARM</button>
      <button type="button" data-soc="hcmd" data-cmd="style" data-id="color:cream">COLOR CREAM</button>
    </div>
    <h4>FURNITURE CATALOG</h4>
    <p>${CATS.join(" · ")}</p>
    <div class="inv-actions">
      <button type="button" data-soc="hcmd" data-cmd="decorate">${inside ? "DECORATE" : "DECORATE (enter first)"}</button>
      <button type="button" data-soc="emote" data-id="sit">SIT</button>
      <button type="button" data-soc="emote" data-id="wave">WAVE</button>
    </div>
    ${items}
    <div class="inv-actions">${decor || "<p>Bawa kursi/tanaman ke rumah.</p>"}</div>
    <h4>GARDEN / FARMING</h4>
    ${plots || "<p>4 plot awal. Max 12.</p>"}
    <h4>OPEN HOUSE VOTE</h4>
    <div class="inv-actions">
      <button type="button" data-soc="hcmd" data-cmd="vote" data-id="COZY">COZY 5</button>
      <button type="button" data-soc="hcmd" data-cmd="vote" data-id="CREATIVE">CREATIVE 5</button>
      <button type="button" data-soc="hcmd" data-cmd="vote" data-id="NATURE">NATURE 5</button>
    </div>
    <h4>TRAINING DUMMY</h4>
    <p>Letakkan Training Dummy. Combat mati di dalam rumah kecuali dummy.</p>
    ${state?.guildHall ? "<p>Guild Room · Training Room · Craft Room · Meeting Room · Garden. Workshop memakai CraftingService.</p>" : ""}`;
}

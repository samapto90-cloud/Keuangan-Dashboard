import { STONE_LAYOUT } from "../game/board/stonePath";
import { zoneClass, zoneDecor } from "./boardZones";

type CfgSnakesLadders = { snakes: Record<number, number>; ladders: Record<number, number> };

/** Render 100 hot-spot batu mengikuti layout ilustrasi (bukan CSS grid). */
export function mountStoneCells(
  grid: HTMLElement,
  cfg: CfgSnakesLadders,
  opts?: { interactive?: boolean },
): HTMLElement[] {
  grid.classList.add("stone-path-grid");
  grid.innerHTML = "";
  const cells: HTMLElement[] = [];
  for (let pos = 1; pos <= 100; pos++) {
    const stone = STONE_LAYOUT[pos - 1]!;
    const cell = document.createElement(opts?.interactive === false ? "div" : "button");
    if (cell instanceof HTMLButtonElement) cell.type = "button";
    const tile = pos % 4;
    cell.className = `cell stone-cell tile-${tile} color-${tile === 0 ? 1 : tile + 1} ${zoneClass(pos)}`;
    cell.dataset.pos = String(pos);
    cell.setAttribute("aria-label", `Kotak ${pos}`);
    cell.style.left = `${stone.left}%`;
    cell.style.top = `${stone.top}%`;
    cell.style.width = `${stone.w}%`;
    cell.style.height = `${stone.h}%`;
    if (pos === 1) cell.classList.add("cell-enter");
    if (pos === 100) cell.classList.add("cell-finish");
    if (cfg.snakes[pos]) cell.classList.add("cell-snake");
    if (cfg.ladders[pos]) cell.classList.add("cell-ladder");
    // Angka di batu sudah ada di ilustrasi — label tipis hanya untuk aksesibilitas / hover.
    cell.innerHTML = `<span class="cell-num" aria-hidden="true">${pos}</span>${zoneDecor(pos)}`;
    grid.appendChild(cell);
    cells.push(cell);
  }
  return cells;
}

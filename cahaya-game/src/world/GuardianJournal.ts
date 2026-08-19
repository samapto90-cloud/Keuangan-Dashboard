import type { GuardianView } from "../network/NetworkMessage";

export class GuardianJournal {
  render(list: GuardianView[], tokens = 0): string {
    if (!list.length) return "<p>Belum ada data Siluman.</p>";
    const rows = list
      .map((g) => {
        const n = String(g.index ?? 0).padStart(2, "0");
        const mark = g.codexStatus === "ALLY" || g.status === "ALLY" ? "✦" : g.status === "DEFEATED" || g.codexStatus === "DEFEATED" ? "✓" : g.codexStatus === "UNDERSTOOD" ? "◎" : g.status === "AVAILABLE" ? "○" : "?";
        const shown = g.storyName ? `${g.storyName}` : g.name;
        return `<div class="journal-row"><b>${n} ${mark} ${shown}</b><span>${g.codexStatus || g.status}</span><p>${g.title}${g.personality ? " · " + g.personality : ""}${g.weakness ? " · lemah: " + g.weakness : ""}${g.miniBoss ? " · MINI BOSS" : ""}</p>${g.story ? `<p>${g.story}</p>` : ""}</div>`;
      })
      .join("");
    return `<p>33 SILUMAN · Token ${tokens}/33</p>${rows}`;
  }
}

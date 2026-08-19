import type {
  ObjectiveView,
  PlayerProgress,
  QuestView,
} from "../network/NetworkMessage";

export class QuestStore {
  progress: PlayerProgress | null = null;

  apply(next: PlayerProgress): void {
    this.progress = next;
  }

  quest(id: string): QuestView | undefined {
    return this.progress?.quests.find((q) => q.id === id);
  }

  tracked(): QuestView | null {
    const list = this.progress?.quests ?? [];
    return list.find((q) => q.state === "ACTIVE" || q.state === "COMPLETED") ?? null;
  }

  markerFor(npcId: string): "" | "!" | "?" {
    for (const q of this.progress?.quests ?? []) {
      if (q.state === "COMPLETED" && (q.npc === npcId || claimedAt(q) === npcId)) return "?";
      if (q.state === "AVAILABLE" && q.npc === npcId) return "!";
    }
    return "";
  }

  claimed(id: string): boolean {
    return this.progress?.claimed.includes(id) === true;
  }

  forestUnlocked(): boolean {
    return this.progress?.forestUnlocked === true;
  }

  patchWallet(coin: number, crystal: number, eduToken: number): void {
    if (!this.progress) return;
    this.progress.coin = coin;
    this.progress.crystal = crystal;
    this.progress.eduToken = eduToken;
  }
}

function claimedAt(q: QuestView): string {
  if (q.id === "mq004") return "raven";
  if (q.id === "eq001") return "mira";
  if (q.id === "sq001") return "lio";
  if (q.id === "sq002") return "nara";
  return q.npc;
}

export function objectiveLine(q: QuestView): string {
  const obj: ObjectiveView | undefined = q.objectives[0];
  if (!obj) return q.description;
  return `${obj.text}  ${obj.progress} / ${obj.count}`;
}

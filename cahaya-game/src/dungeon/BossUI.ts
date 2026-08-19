import type { Scene } from "three";
import type { BossAOEOut, BossTelegraphOut, DungeonView } from "../network/NetworkMessage";
import { BossHealthBar } from "./BossHealthBar";
import { BossTelegraphRenderer } from "./BossTelegraphRenderer";

export class BossUI {
  readonly bar: BossHealthBar;
  readonly telegraph: BossTelegraphRenderer;

  constructor(host: HTMLElement, scene: Scene) {
    this.bar = new BossHealthBar(host);
    this.telegraph = new BossTelegraphRenderer(scene);
  }

  apply(view: DungeonView | null): void {
    this.bar.apply(view?.boss);
  }

  onTelegraph(ev: BossTelegraphOut): void {
    this.telegraph.show(ev);
  }

  onAoe(ev: BossAOEOut): void {
    this.telegraph.impact(ev);
  }

  update(): void {
    this.telegraph.update();
  }

  clear(): void {
    this.bar.apply(null);
    this.telegraph.clear();
  }
}

import { NET } from "../game/GameConfig";
import type { NetworkClient } from "./NetworkClient";
import { Prediction } from "./Prediction";

export class PlayerSync {
  readonly prediction = new Prediction();
  private lastSend = 0;

  sendLocalInput(
    net: NetworkClient,
    ax: number,
    az: number,
    yaw: number,
    sprint: boolean,
    jump: boolean,
    now: number,
  ): void {
    if (now - this.lastSend < 1000 / NET.inputHz) return;
    this.lastSend = now;
    net.sendMoveInput({
      seq: this.prediction.nextSeq(),
      ax,
      az,
      yaw,
      sprint,
      jump,
    });
  }
}

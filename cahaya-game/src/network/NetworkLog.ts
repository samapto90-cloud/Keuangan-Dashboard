import { DEBUG_MODE } from "../game/GameConfig";

const noisy = new Set(["WORLD_SNAPSHOT", "MOVE_INPUT", "PING", "PONG"]);
let lastSnapshotLog = 0;

export function netLog(event: string, extra = ""): void {
  if (!DEBUG_MODE) return;
  if (event === "WORLD_SNAPSHOT") {
    const now = performance.now();
    if (now - lastSnapshotLog < 1000) return;
    lastSnapshotLog = now;
  }
  if (noisy.has(event) && event !== "WORLD_SNAPSHOT") return;
  console.info(`[NET] ${event}${extra ? ` ${extra}` : ""}`);
}

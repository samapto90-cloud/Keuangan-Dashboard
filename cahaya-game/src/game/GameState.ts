export type GamePhase = "loading" | "title" | "playing";

export class GameState {
  phase: GamePhase = "loading";
  embedMode = false;
  paused = false;
  pointerLocked = false;
  debug = false;

  constructor() {
    const params = new URLSearchParams(window.location.search);
    this.embedMode = params.get("embed") === "1";
    this.debug = params.get("debug") === "1";
  }
}

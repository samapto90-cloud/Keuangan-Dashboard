export {
  displayName,
  spawnToSnapshot,
  wsUrl,
  type AuthIn,
  type AuthOkOut,
  type DespawnOut,
  type Envelope,
  type ErrorOut,
  type JoinWorldIn,
  type MoveInput,
  type MovementState,
  type NetMsgType,
  type PingIn,
  type PlayerSnapshot,
  type PlayerSpawn,
  type PongOut,
  type WelcomeOut,
  type WorldSnapshot,
} from "./NetworkMessage";

/** Alias Phase 1 — spawn publik tanpa inventory/combat. */
export type PlayerPublic = import("./NetworkMessage").PlayerSpawn;

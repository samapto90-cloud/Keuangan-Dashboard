import type { GameFlowState } from "./game/state";
import type { BoardPlayer } from "./profile/player";
import type { Room } from "./rooms/model";
import type { Match } from "./lobby/match";
import type { LeaderboardRow } from "./leaderboard/model";
import { TOKEN_COLORS } from "./game/tokens";
import { MIN_POSITION } from "./game/board/config";

export const FOUNDATION_STATES: GameFlowState[] = [
  "WAITING",
  "READY",
  "STARTING",
  "PLAYER_TURN",
  "ROLLING",
  "MOVING",
  "SNAKE",
  "LADDER",
  "QUESTION",
  "ANSWERING",
  "PENALTY",
  "NEXT_TURN",
  "FINISHED",
];

export const SAMPLE_PLAYER: BoardPlayer = {
  id: "preview",
  userId: "preview",
  username: "Pemain",
  avatar: "default",
  color: TOKEN_COLORS[0],
  position: MIN_POSITION,
  isReady: false,
  isConnected: true,
};

export const ROOM_FIELDS: (keyof Room)[] = ["id", "roomCode", "hostId", "status", "maxPlayers", "createdAt"];
export const MATCH_FIELDS: (keyof Match)[] = ["id", "roomId", "status", "currentPlayerId", "winnerId", "createdAt", "finishedAt"];
export const LB_FIELDS: (keyof LeaderboardRow)[] = ["userId", "username", "wins"];

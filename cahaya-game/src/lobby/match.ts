export type Match = {
  id: string;
  roomId: string;
  status: string;
  currentPlayerId: string;
  winnerId?: string;
  createdAt: string;
  finishedAt?: string;
};

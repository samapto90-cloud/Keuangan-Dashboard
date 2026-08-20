import type { AchievementView } from "../progress/types";

export type FriendCard = {
  userId: string;
  username?: string;
  avatar?: string;
  level?: number;
  rankLabel?: string;
  rankRr?: number;
  status?: string;
  friends?: boolean;
};

export type FriendRequest = {
  id: string;
  senderId: string;
  receiverId: string;
  status: string;
  username?: string;
  avatar?: string;
};

export type SocialPrivacy = {
  allowFriendRequests: boolean;
  allowGameInvites: boolean;
  showOnlineStatus: boolean;
};

export type RankHistoryItem = {
  result: string;
  tierBefore: string;
  divBefore?: string;
  rrBefore: number;
  rrChange: number;
  rrAfter: number;
  tierAfter: string;
  divAfter?: string;
  createdAt: number;
};

export type LeaderboardRow = {
  userId: string;
  username: string;
  avatar: string;
  level: number;
  tier: string;
  division?: string;
  rr: number;
  wins: number;
  xp: number;
  accuracy: number;
  coins: number;
  rankLabel: string;
};

export type UlarNote = {
  id: string;
  type: string;
  title: string;
  body: string;
  read: boolean;
  createdAt: number;
};

async function req<T>(path: string, token: string, init?: RequestInit): Promise<{ ok: true; data: T } | { ok: false; error: string }> {
  try {
    const headers: Record<string, string> = { ...(init?.headers as Record<string, string> | undefined), Authorization: `Bearer ${token}` };
    if (init?.body) headers["Content-Type"] = "application/json";
    const res = await fetch(path, { ...init, headers });
    const data = (await res.json()) as T & { error?: string };
    if (!res.ok) return { ok: false, error: data.error || "permintaan gagal" };
    return { ok: true, data };
  } catch {
    return { ok: false, error: "tidak dapat terhubung ke server" };
  }
}

export async function fetchFriends(token: string) {
  return req<{ friends: FriendCard[]; incoming: FriendRequest[]; outgoing: FriendRequest[]; unread: number }>("/cahaya/api/friends", token);
}

export async function searchPlayers(token: string, q: string) {
  return req<{ items: FriendCard[] }>("/cahaya/api/friends/search?q=" + encodeURIComponent(q), token);
}

export async function friendRequest(token: string, userId: string) {
  return req<FriendRequest>("/cahaya/api/friends/request", token, { method: "POST", body: JSON.stringify({ userId }) });
}

export async function friendRespond(token: string, requestId: string, action: "accept" | "reject" | "cancel") {
  return req<FriendRequest>("/cahaya/api/friends/respond", token, { method: "POST", body: JSON.stringify({ requestId, action }) });
}

export async function friendRemove(token: string, userId: string) {
  return req<{ ok: boolean }>("/cahaya/api/friends/remove", token, { method: "POST", body: JSON.stringify({ userId }) });
}

export async function friendBlock(token: string, userId: string) {
  return req<{ ok: boolean }>("/cahaya/api/friends/block", token, { method: "POST", body: JSON.stringify({ userId }) });
}

export async function fetchPublicPlayer(token: string, userId: string) {
  return req<{ username: string; level: number; rankLabel: string; winRate: number; accuracy: number; achievements: AchievementView[]; avatar: string }>("/cahaya/api/players/public?userId=" + encodeURIComponent(userId), token);
}

export async function fetchLeaderboard(token: string, kind: string, page: number, scope: string) {
  return req<{ items: LeaderboardRow[]; myRank: number; page: number; total: number }>("/cahaya/api/leaderboard?kind=" + kind + "&page=" + page + "&scope=" + scope, token);
}

export async function fetchNotes(token: string) {
  return req<{ items: UlarNote[]; unread: number }>("/cahaya/api/notifications", token);
}

export async function markNotesRead(token: string) {
  return req<{ ok: boolean }>("/cahaya/api/notifications", token, { method: "POST", body: "{}" });
}

export async function fetchSeason(token: string) {
  return req<{ id: string; name: string; active: boolean }>("/cahaya/api/season", token);
}

export async function reportPlayer(token: string, userId: string, reason: string, description = "", matchId = "") {
  return req<{ id: string }>("/cahaya/api/report", token, { method: "POST", body: JSON.stringify({ userId, reason, description, matchId }) });
}

export async function fetchPrivacy(token: string) {
  return req<SocialPrivacy>("/cahaya/api/privacy", token);
}

export async function savePrivacy(token: string, p: SocialPrivacy) {
  return req<SocialPrivacy>("/cahaya/api/privacy", token, { method: "POST", body: JSON.stringify(p) });
}

export async function fetchRankHistory(token: string, page = 0) {
  return req<{ items: RankHistoryItem[]; page: number }>("/cahaya/api/rank/history?page=" + page, token);
}

export function statusDot(status?: string): string {
  if (status === "ONLINE") return "🟢";
  if (status === "IN_GAME") return "🎮";
  if (status === "AWAY") return "🟡";
  return "⚫";
}

export function pingTone(ms: number): string {
  if (ms <= 0) return "—";
  if (ms < 60) return `🟢 ${ms}ms`;
  if (ms < 120) return `🟡 ${ms}ms`;
  return `🔴 ${ms}ms`;
}

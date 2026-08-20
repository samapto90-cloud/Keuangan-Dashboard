import type { DailyStatus, MatchHistoryItem, PlayerProfileView, RewardEvent } from "./types";

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

export async function fetchProfile(token: string) {
  return req<PlayerProfileView>("/cahaya/api/profile", token);
}

export async function fetchHistory(token: string, page = 0) {
  return req<{ items: MatchHistoryItem[]; page: number }>("/cahaya/api/profile/history?page=" + page, token);
}

export async function fetchDaily(token: string) {
  return req<DailyStatus>("/cahaya/api/daily-reward", token);
}

export async function claimDaily(token: string) {
  return req<{ reward: RewardEvent; profile: PlayerProfileView }>("/cahaya/api/daily-reward/claim", token, { method: "POST", body: "{}" });
}

export async function updateProfile(token: string, patch: { username?: string; avatar?: string; title?: string }) {
  return req<PlayerProfileView>("/cahaya/api/profile/update", token, { method: "POST", body: JSON.stringify(patch) });
}

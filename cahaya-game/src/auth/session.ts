export type Session = {
  token: string;
  playerId: string;
  username: string;
};

export type Profile = {
  playerId: string;
  username: string;
  email?: string;
};

const TOKEN_KEY = "cahaya-session";
const NAME_KEY = "cahaya-username";

export function storedSessionToken(): string {
  return window.localStorage.getItem(TOKEN_KEY) || "";
}

export function saveSession(s: Session): void {
  window.localStorage.setItem(TOKEN_KEY, s.token);
  window.localStorage.setItem(NAME_KEY, s.username);
}

export function clearSession(): void {
  window.localStorage.removeItem(TOKEN_KEY);
  window.localStorage.removeItem(NAME_KEY);
}

type APIResult<T> = { ok: true; data: T } | { ok: false; error: string };

async function postJSON<T>(path: string, body: unknown, token = ""): Promise<APIResult<T>> {
  try {
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    if (token) headers.Authorization = `Bearer ${token}`;
    const res = await fetch(path, { method: "POST", headers, body: JSON.stringify(body) });
    const data = (await res.json()) as T & { error?: string };
    if (!res.ok) return { ok: false, error: data.error || "permintaan gagal" };
    return { ok: true, data };
  } catch {
    return { ok: false, error: "tidak dapat terhubung ke server" };
  }
}

export async function apiRegister(username: string, email: string, password: string, confirmPassword: string) {
  return postJSON<Session>("/cahaya/api/register", { username, email, password, confirmPassword });
}

export async function apiLogin(username: string, password: string) {
  return postJSON<Session>("/cahaya/api/login", { username, password });
}

export async function apiLogout(token: string) {
  return postJSON<{ ok: boolean }>("/cahaya/api/logout", {}, token);
}

export async function apiProfile(token: string): Promise<Profile | null> {
  try {
    const res = await fetch("/cahaya/api/profile", { headers: { Authorization: `Bearer ${token}` } });
    if (!res.ok) return null;
    return (await res.json()) as Profile;
  } catch {
    return null;
  }
}

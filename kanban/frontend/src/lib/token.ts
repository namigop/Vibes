/**
 * Bearer-token storage.
 *
 * The token lives in localStorage because the API expects an `Authorization`
 * header on every request, which means JavaScript must be able to read it —
 * and only a cookie can be kept from JavaScript. The trade is explicit: any
 * script on the page can read this token. The app renders no third-party
 * scripts and no user-supplied HTML, and a CSP is set in next.config.ts to
 * keep it that way.
 *
 * Consequences worth remembering:
 *   - logout cannot revoke the token. It stays valid until it expires (24h by
 *     default), which is why the server keeps that TTL short.
 *   - `Authorization` is NOT a cookie, so cross-site requests cannot cause the
 *     browser to attach it. CSRF does not apply to this API.
 */

const TOKEN_KEY = "kanban.auth.token";
const USER_KEY = "kanban.auth.user";

export interface StoredUser {
  id: string;
  username: string;
}

/** Notifies subscribers when the token changes, including on logout. */
const listeners = new Set<() => void>();

function notify(): void {
  for (const listener of listeners) listener();
}

export function subscribeToToken(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  try {
    return window.localStorage.getItem(TOKEN_KEY);
  } catch {
    // Private browsing or a blocked-storage policy. Treated as "signed out"
    // rather than crashing the app.
    return null;
  }
}

export function getStoredUser(): StoredUser | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.localStorage.getItem(USER_KEY);
    if (!raw) return null;
    const parsed: unknown = JSON.parse(raw);
    if (
      typeof parsed === "object" &&
      parsed !== null &&
      typeof (parsed as StoredUser).id === "string" &&
      typeof (parsed as StoredUser).username === "string"
    ) {
      return parsed as StoredUser;
    }
    return null;
  } catch {
    return null;
  }
}

export function setSession(token: string, user: StoredUser): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(TOKEN_KEY, token);
    window.localStorage.setItem(USER_KEY, JSON.stringify(user));
  } catch {
    // Storage unavailable — the session simply will not persist.
  }
  notify();
}

export function clearSession(): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.removeItem(TOKEN_KEY);
    window.localStorage.removeItem(USER_KEY);
  } catch {
    // Nothing useful to do; the in-memory state is cleared regardless.
  }
  notify();
}
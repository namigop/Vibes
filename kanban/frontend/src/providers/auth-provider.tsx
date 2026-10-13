"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import * as api from "@/lib/api";
import { clearSession, getToken, setSession, subscribeToToken } from "@/lib/token";
import type { User } from "@/lib/types";

export type AuthStatus = "checking" | "authenticated" | "anonymous";

interface AuthContextValue {
  status: AuthStatus;
  user: User | null;
  signIn: (username: string, password: string) => Promise<void>;
  signOut: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * There is deliberately no route guard.
 *
 * The token lives in localStorage, which Next's Edge middleware cannot read, so
 * a server-side guard is not possible without moving to cookies. Instead the
 * app discovers the session by asking: on mount we either find no token and
 * render the sign-in view directly, or we call `GET /auth/me` and let a 401
 * mean "signed out".
 *
 * The sign-in view replaces the whole app shell rather than being a separate
 * route, so there is no moment where protected chrome renders first.
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>("checking");
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    // Subscribe first: a token cleared by a 401 anywhere in the app must drop
    // us back to the sign-in view immediately.
    return subscribeToToken(() => {
      if (getToken() === null) {
        setUser(null);
        setStatus("anonymous");
      }
    });
  }, []);

  useEffect(() => {
    let cancelled = false;

    // Both paths settle asynchronously. A signed-out visitor therefore still
    // renders one tick of the loader rather than calling setState
    // synchronously inside the effect body, which React flags as a cascading
    // render (and which really does cost an extra render pass).
    const outcome: Promise<{ status: AuthStatus; user: User | null }> = (() => {
      const token = getToken();
      if (!token) {
        return Promise.resolve({ status: "anonymous" as AuthStatus, user: null });
      }

      // The 401 here is the *expected* result of a signed-out visit, so it
      // must not trigger the global "your session ended" path.
      return api
        .withoutAuthNoise(() => api.currentUser())
        .then(
          (me) => ({ status: "authenticated" as AuthStatus, user: me }),
          // The stored token did not verify: discard it rather than retrying.
          () => ({ status: "anonymous" as AuthStatus, user: null }),
        );
    })();

    outcome.then((result) => {
      if (cancelled) return;
      if (result.status === "anonymous") clearSession();
      setUser(result.user);
      setStatus(result.status);
    });

    return () => {
      cancelled = true;
    };
  }, []);

  const signIn = useCallback(async (username: string, password: string) => {
    const result = await api.login(username, password);
    setSession(result.token, result.user);
    setUser(result.user);
    setStatus("authenticated");
  }, []);

  const signOut = useCallback(() => {
    // The token is stateless, so this cannot invalidate it server-side; it stays
    // usable until it expires. Clearing it locally is all a stateless design
    // can honestly do.
    clearSession();
    setUser(null);
    setStatus("anonymous");
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({ status, user, signIn, signOut }),
    [status, user, signIn, signOut],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside <AuthProvider>");
  return context;
}
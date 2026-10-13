"use client";

import { Kanban, LogIn, TriangleAlert } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";

import { Button } from "@/components/ui/button";
import * as api from "@/lib/api";
import { useAuth } from "@/providers/auth-provider";

/**
 * The signed-out view. It replaces the app shell entirely rather than living
 * at /login, so protected chrome never renders for an anonymous visitor.
 *
 * There is no sign-up: accounts are provisioned by the backend seed
 * (`backend/internal/store/seed.go`). The persona names are listed here purely
 * as a development convenience and are not a supported discovery mechanism —
 * do not grow this into a real account picker.
 */
const DEMO_PERSONAS = ["alice", "bob", "carol", "demo"] as const;

export function SignInView() {
  const { signIn } = useAuth();
  const router = useRouter();

  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSubmitting(true);

    try {
      await signIn(username, password);
      router.push("/");
      router.refresh();
    } catch (cause) {
      setError(api.describeError(cause));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="flex h-full w-full items-center justify-center bg-background px-4 py-10">
      <div className="w-full max-w-sm">
        <div className="mb-6 flex flex-col items-center text-center">
          <span className="flex size-11 items-center justify-center rounded-xl bg-accent text-accent-foreground">
            <Kanban size={22} />
          </span>
          <h1 className="mt-4 text-lg font-semibold tracking-tight">Sign in to Kanban</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Boards are private to their owner.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          <div>
            <label
              htmlFor="username"
              className="mb-1.5 block text-xs font-medium text-muted-foreground"
            >
              Username
            </label>
            <input
              id="username"
              name="username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              autoFocus
              required
              className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm outline-none transition-colors placeholder:text-muted-foreground focus-visible:border-accent focus-visible:ring-2 focus-visible:ring-accent/30"
              placeholder="alice"
            />
          </div>

          <div>
            <label
              htmlFor="password"
              className="mb-1.5 block text-xs font-medium text-muted-foreground"
            >
              Password
            </label>
            <input
              id="password"
              name="password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              required
              className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm outline-none transition-colors placeholder:text-muted-foreground focus-visible:border-accent focus-visible:ring-2 focus-visible:ring-accent/30"
              placeholder="••••••••"
            />
          </div>

          {error ? (
            <div
              role="alert"
              className="flex items-start gap-2 rounded-lg border border-danger/30 bg-danger/10 px-3 py-2 text-xs text-danger"
            >
              <TriangleAlert size={14} className="mt-px shrink-0" />
              <span>{error}</span>
            </div>
          ) : null}

          <Button type="submit" variant="primary" className="w-full" disabled={submitting}>
            <LogIn size={15} />
            {submitting ? "Signing in…" : "Sign in"}
          </Button>
        </form>

        <div className="mt-6 rounded-lg border border-dashed border-border bg-surface-muted/50 p-3">
          <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
            Demo accounts
          </p>
          <p className="mt-1.5 text-xs leading-relaxed text-muted-foreground">
            This build has no sign-up. Accounts are seeded by the backend, so try{" "}
            <span className="font-medium text-foreground">
              {DEMO_PERSONAS.join(", ")}
            </span>{" "}
            with the shared demo password.
          </p>
        </div>
      </div>
    </div>
  );
}
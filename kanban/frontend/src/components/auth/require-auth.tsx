"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { SignInView } from "@/components/auth/sign-in-view";
import { Skeleton } from "@/components/ui/styles";
import { useAuth } from "@/providers/auth-provider";

/**
 * Session gate.
 *
 * There is no route-level guard (the token is in localStorage, which Edge
 * middleware cannot read). Instead the whole app is gated on the session being
 * known:
 *
 *   checking      -> a full-bleed loader, so nothing protected ever paints
 *   anonymous     -> the sign-in view, replacing the shell entirely
 *   authenticated -> the real app
 *
 * The loader covers the brief window where the token has been read from storage
 * but `GET /auth/me` has not come back yet.
 */
export function RequireAuth({ children }: { children: ReactNode }) {
  const { status } = useAuth();
  const router = useRouter();

  useEffect(() => {
    // Signing in from the sign-in view should leave the user on the page they
    // originally asked for.
    if (status === "authenticated") router.refresh();
  }, [status, router]);

  if (status === "checking") {
    return (
      <div
        className="flex h-full w-full flex-col gap-3 bg-background p-6"
        aria-busy="true"
        aria-label="Checking your session"
      >
        <div className="flex gap-4">
          <Skeleton className="h-10 w-64 shrink-0" />
          <div className="flex flex-1 gap-4">
            {[0, 1, 2].map((index) => (
              <Skeleton key={index} className="h-48 w-72 shrink-0" />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (status === "anonymous") {
    return <SignInView />;
  }

  return <>{children}</>;
}
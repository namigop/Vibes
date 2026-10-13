import type { ReactNode } from "react";

import { RequireAuth } from "@/components/auth/require-auth";
import { AppShell } from "@/components/layout/app-shell";
import { BoardsProvider } from "@/providers/boards-provider";
import { AuthProvider } from "@/providers/auth-provider";

/**
 * AuthProvider sits above everything so a 401 raised anywhere can drop the app
 * back to the sign-in view. BoardsProvider is inside RequireAuth so it never
 * issues a board request before a session exists.
 */
export default function AppLayout({ children }: { children: ReactNode }) {
  return (
    <AuthProvider>
      <RequireAuth>
        <BoardsProvider>
          <AppShell>{children}</AppShell>
        </BoardsProvider>
      </RequireAuth>
    </AuthProvider>
  );
}
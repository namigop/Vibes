import type { ReactNode } from "react";

import { AppShell } from "@/components/layout/app-shell";
import { BoardsProvider } from "@/providers/boards-provider";

export default function AppLayout({ children }: { children: ReactNode }) {
  return (
    <BoardsProvider>
      <AppShell>{children}</AppShell>
    </BoardsProvider>
  );
}

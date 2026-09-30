"use client";

import { useState, type ReactNode } from "react";

import { SideNav } from "@/components/layout/side-nav";
import { TopBar } from "@/components/layout/top-bar";

/**
 * App shell: a fixed side nav plus a sticky top bar, with the routed page
 * filling the remaining space. Height is locked to the viewport so the board
 * scrolls horizontally inside the page rather than the whole document.
 */
export function AppShell({ children }: { children: ReactNode }) {
  const [navOpen, setNavOpen] = useState(false);

  return (
    <div className="flex h-full w-full overflow-hidden bg-background text-foreground">
      <SideNav open={navOpen} onClose={() => setNavOpen(false)} />
      <div className="flex min-w-0 flex-1 flex-col">
        <TopBar onOpenNav={() => setNavOpen(true)} />
        <main className="min-h-0 flex-1 overflow-hidden">{children}</main>
      </div>
    </div>
  );
}

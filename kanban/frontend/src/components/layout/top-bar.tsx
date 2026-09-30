"use client";

import { Kanban, Menu } from "lucide-react";

import { ThemeToggle } from "@/components/layout/theme-toggle";
import { IconButton } from "@/components/ui/button";

export function TopBar({ onOpenNav }: { onOpenNav: () => void }) {
  return (
    <header className="flex h-14 shrink-0 items-center gap-2 border-b border-border bg-background/85 px-3 backdrop-blur sm:px-4">
      <IconButton label="Open navigation" className="md:hidden" onClick={onOpenNav}>
        <Menu size={18} />
      </IconButton>

      <div className="flex items-center gap-2 md:hidden">
        <span className="flex size-7 items-center justify-center rounded-lg bg-accent text-accent-foreground">
          <Kanban size={14} />
        </span>
        <span className="text-sm font-semibold tracking-tight">Kanban</span>
      </div>

      <div className="ml-auto flex items-center gap-2">
        <ThemeToggle />
      </div>
    </header>
  );
}

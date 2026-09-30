"use client";

import { Moon, Sun } from "lucide-react";

import { useIsClient } from "@/hooks/use-is-client";
import { useTheme } from "@/providers/theme-provider";
import { cn } from "@/lib/utils";
import { FOCUS_RING } from "@/components/ui/styles";

export function ThemeToggle({ className }: { className?: string }) {
  const { theme, toggleTheme } = useTheme();
  const mounted = useIsClient();

  return (
    <button
      type="button"
      onClick={toggleTheme}
      // Rendered only after hydration: the server has no idea which theme the
      // pre-paint script picked, and guessing would mismatch.
      aria-label={mounted ? `Switch to ${theme === "dark" ? "light" : "dark"} theme` : "Switch theme"}
      title={mounted ? `Switch to ${theme === "dark" ? "light" : "dark"} theme` : "Switch theme"}
      aria-pressed={mounted ? theme === "dark" : undefined}
      className={cn(
        "inline-flex size-9 items-center justify-center rounded-lg border border-border bg-surface text-muted-foreground transition-colors hover:border-border-strong hover:text-foreground",
        FOCUS_RING,
        className,
      )}
    >
      {mounted && theme === "dark" ? (
        <Sun size={16} className="text-warning" />
      ) : (
        <Moon size={16} />
      )}
      <span className="sr-only">
        {mounted ? `Current theme: ${theme}` : "Theme"}
      </span>
    </button>
  );
}

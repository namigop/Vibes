import { cn } from "@/lib/utils";

/** Shared focus treatment so every interactive element rings identically. */
export const FOCUS_RING =
  "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background";

export const INPUT_CLASSES =
  "w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground/70 transition-colors hover:border-border-strong focus-visible:border-accent";

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn("animate-pulse rounded-md bg-surface-strong", className)}
      aria-hidden="true"
    />
  );
}

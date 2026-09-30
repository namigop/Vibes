"use client";

import { CalendarClock, TriangleAlert } from "lucide-react";

import { useIsClient } from "@/hooks/use-is-client";
import { dueStateOf, formatLongDate, formatShortDate, type DueState } from "@/lib/dates";
import { cn } from "@/lib/utils";

const TONES: Record<Exclude<DueState, "none">, string> = {
  overdue: "text-danger",
  today: "text-warning",
  soon: "text-muted-foreground",
  later: "text-muted-foreground",
};

/**
 * `due_date` is a calendar date, so it is parsed in local time (see
 * `lib/dates.ts`) — formatting it as an instant would shift it by a day for
 * anyone west of UTC. The value is only rendered after hydration for the same
 * reason: the server's timezone may differ from the reader's and React would
 * report a mismatch.
 */
export function DueDateChip({
  value,
  className,
}: {
  value: string | null;
  className?: string;
}) {
  const mounted = useIsClient();

  if (!value) return null;

  const state = dueStateOf(value);
  const tone = state === "none" ? TONES.later : TONES[state];
  const label = mounted ? formatShortDate(value) : "";

  // Overdue is flagged with a warning icon and the word "Overdue" as well as
  // colour, so the state survives greyscale and colour-blindness.
  const prefix = state === "overdue" ? "Overdue · " : state === "today" ? "Today · " : "";

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-md bg-surface-muted px-1.5 py-0.5 text-[11px] font-medium",
        tone,
        className,
      )}
      title={mounted ? formatLongDate(value) : undefined}
    >
      {state === "overdue" ? (
        <TriangleAlert size={11} aria-hidden="true" />
      ) : (
        <CalendarClock size={11} aria-hidden="true" />
      )}
      {/* Fixed width until formatted so hydration causes no layout shift. */}
      <span className={cn(!mounted && "invisible")}>{`${prefix}${label}`}</span>
      <span className="sr-only">
        {state === "overdue" ? "Overdue" : "Due"} {mounted ? formatLongDate(value) : ""}
      </span>
    </span>
  );
}

/**
 * `due_date` is a CALENDAR date ("2026-10-15"), not an instant.
 *
 * The trap: `new Date("2026-10-15")` parses as UTC midnight, which is the
 * previous day for any negative UTC offset (e.g. 2026-10-14 in New York). We
 * therefore always parse the three numeric parts by hand into a *local* Date,
 * and we never pass the value through `toISOString()`.
 */

const DATE_PART = /^(\d{4})-(\d{2})-(\d{2})$/;

/** Parse `YYYY-MM-DD` into a local Date, or null when absent/malformed. */
export function parseCalendarDate(value: string | null | undefined): Date | null {
  if (!value) return null;
  const match = DATE_PART.exec(value);
  if (!match) return null;
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const date = new Date(year, month - 1, day);
  // Guards against overflow like "2026-02-31" silently becoming March 3rd.
  if (
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day
  ) {
    return null;
  }
  return date;
}

const shortFormatter = new Intl.DateTimeFormat(undefined, {
  month: "short",
  day: "numeric",
});

const longFormatter = new Intl.DateTimeFormat(undefined, {
  year: "numeric",
  month: "short",
  day: "numeric",
});

/** "Oct 15" — short form for the card face. */
export function formatShortDate(value: string | null | undefined): string {
  const date = parseCalendarDate(value);
  return date ? shortFormatter.format(date) : "";
}

/** "Oct 15, 2026" — long form for the detail drawer. */
export function formatLongDate(value: string | null | undefined): string {
  const date = parseCalendarDate(value);
  return date ? longFormatter.format(date) : "";
}

/** Midnight, local time, today. */
function todayLocal(): Date {
  const now = new Date();
  return new Date(now.getFullYear(), now.getMonth(), now.getDate());
}

export type DueState = "none" | "overdue" | "today" | "soon" | "later";

/**
 * Classify a due date against the local calendar day.
 * Comparison is done on local midnights so a date is never shifted by the
 * viewer's timezone the way `Date.parse` + `toISOString` would.
 */
export function dueStateOf(value: string | null | undefined): DueState {
  const date = parseCalendarDate(value);
  if (!date) return "none";
  const diffDays = Math.round(
    (date.getTime() - todayLocal().getTime()) / 86_400_000,
  );
  if (diffDays < 0) return "overdue";
  if (diffDays === 0) return "today";
  if (diffDays <= 3) return "soon";
  return "later";
}

/** ISO-8601 timestamp (RFC3339 from the server) -> short local date + time. */
export function formatTimestamp(value: string | null | undefined): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

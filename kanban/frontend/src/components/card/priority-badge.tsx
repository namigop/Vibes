import { ArrowDown, ArrowUp, Minus, TriangleAlert, type LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";
import type { Priority } from "@/lib/types";

/**
 * Priority is encoded three ways so it never depends on colour alone: a
 * distinct icon, a written label, and the colour ramp.
 */
const STYLES: Record<Priority, { label: string; icon: LucideIcon; className: string }> = {
  low: {
    label: "Low",
    icon: ArrowDown,
    className: "bg-success-soft text-success",
  },
  medium: {
    label: "Medium",
    icon: Minus,
    className: "bg-surface-muted text-muted-foreground",
  },
  high: {
    label: "High",
    icon: ArrowUp,
    className: "bg-warning-soft text-warning",
  },
  urgent: {
    label: "Urgent",
    icon: TriangleAlert,
    className: "bg-danger-soft text-danger",
  },
};

export function PriorityBadge({
  priority,
  size = "sm",
  className,
}: {
  priority: Priority;
  size?: "sm" | "md";
  className?: string;
}) {
  const { label, icon: Icon, className: tone } = STYLES[priority] ?? STYLES.medium;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-md font-medium",
        tone,
        size === "sm" ? "px-1.5 py-0.5 text-[11px]" : "px-2 py-1 text-xs",
        className,
      )}
    >
      <Icon size={size === "sm" ? 11 : 13} aria-hidden="true" />
      {label}
    </span>
  );
}

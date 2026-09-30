import { CircleUser } from "lucide-react";
import type { CSSProperties } from "react";

import { hueForName } from "@/lib/avatar";
import { cn, initialsOf } from "@/lib/utils";

/**
 * Assignees are free text (there is no user table), so the avatar is derived
 * from the name: initials plus a stable hue, meaning the same person always
 * gets the same colour on every board.
 */
export function AssigneeAvatar({
  name,
  size = 24,
  className,
}: {
  name: string;
  size?: number;
  className?: string;
}) {
  const initials = initialsOf(name);
  const hue = hueForName(name);

  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center justify-center rounded-full font-semibold uppercase text-white ring-1 ring-black/10",
        className,
      )}
      style={
        {
          width: size,
          height: size,
          fontSize: Math.max(9, Math.round(size * 0.38)),
          "--avatar-hue": String(hue),
          backgroundColor: "hsl(var(--avatar-hue), 52%, 42%)",
        } as CSSProperties
      }
      title={name}
    >
      <span aria-hidden="true">{initials}</span>
      <span className="sr-only">Assigned to {name}</span>
    </span>
  );
}

export function UnassignedAvatar({ size = 24 }: { size?: number }) {
  return (
    <span
      className="inline-flex shrink-0 items-center justify-center rounded-full border border-dashed border-border-strong text-muted-foreground"
      style={{ width: size, height: size }}
      title="Unassigned"
    >
      <CircleUser size={Math.round(size * 0.6)} aria-hidden="true" />
      <span className="sr-only">Unassigned</span>
    </span>
  );
}

export function Avatar({
  name,
  size = 24,
}: {
  name: string;
  size?: number;
}) {
  return name.trim() ? <AssigneeAvatar name={name} size={size} /> : <UnassignedAvatar size={size} />;
}

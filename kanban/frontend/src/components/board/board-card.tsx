"use client";

import { useSortable } from "@dnd-kit/sortable";
import { GripVertical } from "lucide-react";

import { Avatar } from "@/components/card/assignee-avatar";
import { DueDateChip } from "@/components/card/due-date-chip";
import { PriorityBadge } from "@/components/card/priority-badge";
import { cardId } from "@/lib/constants";
import { cn } from "@/lib/utils";
import type { Card } from "@/lib/types";
import { FOCUS_RING } from "@/components/ui/styles";

/** Title, description preview, priority, due date, assignee. */
function CardFace({ card }: { card: Card }) {
  return (
    <>
      <p className="line-clamp-2 pr-4 text-sm font-medium leading-snug text-foreground">
        {card.title}
      </p>
      {card.description ? (
        <p className="mt-1 line-clamp-1 text-xs text-muted-foreground">{card.description}</p>
      ) : null}
      <div className="mt-2.5 flex flex-wrap items-center gap-1.5">
        <PriorityBadge priority={card.priority} />
        {card.due_date ? <DueDateChip value={card.due_date} /> : null}
        <span className="ml-auto">
          <Avatar name={card.assignee} size={22} />
        </span>
      </div>
    </>
  );
}

export function BoardCard({
  card,
  onOpen,
}: {
  card: Card;
  onOpen: (card: Card) => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transition,
    isDragging,
  } = useSortable({
    id: cardId(card.id),
    data: { type: "card", cardId: card.id, columnId: card.column_id },
  });

  return (
    <li
      ref={setNodeRef}
      // The dragged card deliberately does NOT receive `transform`: the
      // DragOverlay follows the cursor while this node stays put as a dimmed
      // placeholder, so the gap it leaves behind is visible.
      style={{ transition }}
      className={cn(
        "group relative rounded-xl transition-shadow",
        isDragging && "opacity-40 ring-2 ring-accent/40 ring-dashed",
      )}
      aria-current={isDragging ? "true" : undefined}
    >
      <button
        type="button"
        onClick={() => onOpen(card)}
        className={cn(
          "w-full rounded-xl border border-border bg-surface p-3 text-left shadow-sm transition-colors hover:border-border-strong hover:bg-surface-strong",
          FOCUS_RING,
        )}
      >
        <CardFace card={card} />
      </button>

      {/* Drag handle. Kept as its own control so the card body can stay a plain
          button: no nested interactive elements, and the keyboard drag
          instructions dnd-kit attaches land on a dedicated target. */}
      <span
        ref={setActivatorNodeRef}
        {...attributes}
        {...listeners}
        aria-label={`Drag ${card.title}`}
        className={cn(
          "absolute right-1.5 top-1.5 flex size-6 items-center justify-center rounded-md text-muted-foreground/60 transition-colors hover:bg-surface-muted hover:text-foreground",
          "cursor-grab touch-none active:cursor-grabbing focus-visible:opacity-100",
          FOCUS_RING,
        )}
      >
        <GripVertical size={14} aria-hidden="true" />
      </span>
    </li>
  );
}

/** Static clone rendered inside <DragOverlay>. */
export function CardOverlay({ card }: { card: Card }) {
  return (
    <div className="w-72 rotate-1 rounded-xl border border-accent/40 bg-surface p-3 shadow-2xl shadow-black/25">
      <CardFace card={card} />
    </div>
  );
}

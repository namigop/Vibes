"use client";

import { useDroppable } from "@dnd-kit/core";
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import {
  ChevronLeft,
  ChevronRight,
  Ellipsis,
  GripVertical,
  Pencil,
  Plus,
  Trash,
  X,
} from "lucide-react";
import { useEffect, useRef, useState, type FormEvent, type KeyboardEvent } from "react";

import { BoardCard } from "@/components/board/board-card";
import { Button, IconButton } from "@/components/ui/button";
import {
  Menu,
  MenuContent,
  MenuItem,
  MenuSeparator,
  MenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cardId, columnId, columnZoneId } from "@/lib/constants";
import { cn } from "@/lib/utils";
import type { Card, Column } from "@/lib/types";
import { FOCUS_RING, INPUT_CLASSES } from "@/components/ui/styles";

export interface BoardColumnProps {
  column: Column;
  cards: Card[];
  /** Position in the board, for the move left/right menu items. */
  index: number;
  total: number;
  /** True while any card is being dragged, so empty columns can react. */
  cardInFlight: boolean;
  onOpenCard: (card: Card) => void;
  onRename: (name: string) => void;
  onDelete: () => void;
  onMove: (toIndex: number) => void;
  onCreateCard: (title: string) => Promise<boolean>;
}

export function BoardColumn({
  column,
  cards,
  index,
  total,
  cardInFlight,
  onOpenCard,
  onRename,
  onDelete,
  onMove,
  onCreateCard,
}: BoardColumnProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: columnId(column.id),
    data: { type: "column", columnId: column.id },
  });

  // Registered for every column, not just empty ones: it is the forgiving
  // "drop anywhere in this lane" target, and for an empty column it is the
  // only droppable at all.
  const { setNodeRef: setZoneRef, isOver } = useDroppable({
    id: columnZoneId(column.id),
    data: { type: "columnZone", columnId: column.id },
  });

  const [editingName, setEditingName] = useState(false);
  const [draftName, setDraftName] = useState(column.name);
  const [composing, setComposing] = useState(false);
  const [draftTitle, setDraftTitle] = useState("");
  const [saving, setSaving] = useState(false);
  const composerRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (composing) composerRef.current?.focus();
  }, [composing]);

  const commitName = () => {
    const trimmed = draftName.trim();
    setEditingName(false);
    if (!trimmed || trimmed === column.name) {
      setDraftName(column.name);
      return;
    }
    onRename(trimmed);
  };

  const submitCard = async (event: FormEvent) => {
    event.preventDefault();
    const trimmed = draftTitle.trim();
    if (!trimmed || saving) return;
    setSaving(true);
    const ok = await onCreateCard(trimmed);
    setSaving(false);
    if (ok) {
      setDraftTitle("");
      composerRef.current?.focus();
    }
  };

  return (
    <section
      ref={setNodeRef}
      style={{ transform: CSS.Translate.toString(transform), transition }}
      aria-label={`${column.name} column, ${cards.length} ${
        cards.length === 1 ? "card" : "cards"
      }`}
      className={cn(
        "flex h-full max-h-full w-72 shrink-0 flex-col rounded-2xl border border-border bg-surface-muted",
        isDragging && "opacity-60",
      )}
    >
      <header className="flex shrink-0 items-center gap-1 px-2.5 py-2.5">
        <span
          ref={setActivatorNodeRef}
          {...attributes}
          {...listeners}
          aria-label={`Reorder ${column.name} column`}
          className={cn(
            "flex size-6 shrink-0 cursor-grab touch-none items-center justify-center rounded-md text-muted-foreground/70 transition-colors hover:bg-surface hover:text-foreground active:cursor-grabbing",
            FOCUS_RING,
          )}
        >
          <GripVertical size={14} aria-hidden="true" />
        </span>

        {editingName ? (
          <input
            autoFocus
            value={draftName}
            onChange={(event) => setDraftName(event.target.value)}
            onBlur={commitName}
            onKeyDown={(event: KeyboardEvent<HTMLInputElement>) => {
              if (event.key === "Enter") {
                event.preventDefault();
                commitName();
              } else if (event.key === "Escape") {
                event.preventDefault();
                setDraftName(column.name);
                setEditingName(false);
              }
            }}
            aria-label="Column name"
            className={cn(INPUT_CLASSES, FOCUS_RING, "h-7 py-1 text-sm")}
          />
        ) : (
          <button
            type="button"
            onClick={() => {
              setDraftName(column.name);
              setEditingName(true);
            }}
            className={cn(
              "min-w-0 flex-1 truncate rounded-md px-1 py-0.5 text-left text-sm font-semibold text-foreground transition-colors hover:bg-surface",
              FOCUS_RING,
            )}
          >
            {column.name}
          </button>
        )}

        <span
          className="shrink-0 rounded-md bg-surface-strong px-1.5 py-0.5 text-[11px] font-medium tabular-nums text-muted-foreground"
          title={`${cards.length} ${cards.length === 1 ? "card" : "cards"}`}
        >
          {cards.length}
        </span>

        <Menu>
          <MenuTrigger label={`Actions for ${column.name}`} size="icon-sm">
            <Ellipsis size={14} />
          </MenuTrigger>
          <MenuContent>
            <MenuItem
              icon={<Pencil size={14} />}
              onSelect={() => {
                setDraftName(column.name);
                setEditingName(true);
              }}
            >
              Rename
            </MenuItem>
            <MenuItem
              icon={<ChevronLeft size={14} />}
              disabled={index === 0}
              onSelect={() => onMove(index - 1)}
            >
              Move left
            </MenuItem>
            <MenuItem
              icon={<ChevronRight size={14} />}
              disabled={index === total - 1}
              onSelect={() => onMove(index + 1)}
            >
              Move right
            </MenuItem>
            <MenuSeparator />
            <MenuItem
              destructive
              icon={<Trash size={14} />}
              onSelect={onDelete}
            >
              Delete column
            </MenuItem>
          </MenuContent>
        </Menu>
      </header>

      <div
        ref={setZoneRef}
        className={cn(
          "min-h-0 flex-1 overflow-y-auto rounded-xl px-2 transition-colors",
          isOver && cardInFlight && cards.length === 0 && "bg-accent-soft",
        )}
      >
        {cards.length === 0 ? (
          <div
            className={cn(
              "flex h-24 items-center justify-center rounded-xl border border-dashed text-xs text-muted-foreground transition-colors",
              cardInFlight
                ? "border-accent/60 bg-accent-soft/50 text-accent"
                : "border-border-strong",
            )}
          >
            {cardInFlight ? "Drop cards here" : "No cards yet"}
          </div>
        ) : (
          <SortableContext
            items={cards.map((card) => cardId(card.id))}
            strategy={verticalListSortingStrategy}
          >
            <ul className="space-y-2 py-0.5">
              {cards.map((card) => (
                <BoardCard key={card.id} card={card} onOpen={onOpenCard} />
              ))}
            </ul>
          </SortableContext>
        )}
      </div>

      <div className="shrink-0 p-2">
        {composing ? (
          <form onSubmit={submitCard} className="space-y-2">
            <textarea
              ref={composerRef}
              value={draftTitle}
              onChange={(event) => setDraftTitle(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter" && !event.shiftKey) {
                  event.preventDefault();
                  void submitCard(event);
                } else if (event.key === "Escape") {
                  event.preventDefault();
                  setDraftTitle("");
                  setComposing(false);
                }
              }}
              rows={2}
              maxLength={200}
              placeholder="What needs doing?"
              aria-label="Card title"
              className={cn(INPUT_CLASSES, FOCUS_RING, "resize-none text-sm")}
            />
            <div className="flex items-center gap-1.5">
              <Button
                type="submit"
                variant="primary"
                size="sm"
                disabled={saving || !draftTitle.trim()}
              >
                {saving ? "Adding…" : "Add card"}
              </Button>
              <IconButton
                label="Cancel adding a card"
                size="icon-sm"
                onClick={() => {
                  setDraftTitle("");
                  setComposing(false);
                }}
              >
                <X size={14} />
              </IconButton>
            </div>
          </form>
        ) : (
          <button
            type="button"
            onClick={() => setComposing(true)}
            className={cn(
              "flex w-full items-center gap-1.5 rounded-lg px-2 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-surface hover:text-foreground",
              FOCUS_RING,
            )}
          >
            <Plus size={15} aria-hidden="true" />
            Add card
          </button>
        )}
      </div>
    </section>
  );
}

"use client";

import { Check, Pencil, X } from "lucide-react";
import { useState, type KeyboardEvent } from "react";

import { Button, IconButton } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { Board, Card, Column } from "@/lib/types";
import { FOCUS_RING, INPUT_CLASSES } from "@/components/ui/styles";

export function BoardHeader({
  board,
  columns,
  cards,
  onRename,
}: {
  board: Board;
  columns: Column[];
  cards: Card[];
  onRename: (name: string) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(board.name);

  const commit = () => {
    const trimmed = draft.trim();
    setEditing(false);
    if (!trimmed || trimmed === board.name) {
      setDraft(board.name);
      return;
    }
    onRename(trimmed);
  };

  return (
    <header className="flex shrink-0 flex-wrap items-center gap-x-3 gap-y-1 border-b border-border bg-background px-4 py-3 sm:px-6">
      <h1 className="min-w-0 text-lg font-semibold tracking-tight">
        {editing ? (
          <input
            autoFocus
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            onBlur={commit}
            onKeyDown={(event: KeyboardEvent<HTMLInputElement>) => {
              if (event.key === "Enter") {
                event.preventDefault();
                commit();
              } else if (event.key === "Escape") {
                event.preventDefault();
                setDraft(board.name);
                setEditing(false);
              }
            }}
            aria-label="Board name"
            maxLength={80}
            className={cn(
              INPUT_CLASSES,
              FOCUS_RING,
              "h-9 w-full min-w-0 py-1 text-lg font-semibold tracking-tight sm:w-80",
            )}
          />
        ) : (
          <button
            type="button"
            onClick={() => {
              setDraft(board.name);
              setEditing(true);
            }}
            className={cn(
              "group flex min-w-0 items-center gap-2 rounded-lg px-2 py-1 -mx-2 text-left transition-colors hover:bg-surface-muted",
              FOCUS_RING,
            )}
          >
            <span className="truncate">{board.name}</span>
            <Pencil
              size={14}
              aria-hidden="true"
              className="shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
            />
            <span className="sr-only">Rename board</span>
          </button>
        )}
      </h1>

      {editing ? (
        <div className="flex items-center gap-1">
          <IconButton label="Save board name" size="icon-sm" onClick={commit}>
            <Check size={14} />
          </IconButton>
          <IconButton
            label="Cancel renaming"
            size="icon-sm"
            onClick={() => {
              setDraft(board.name);
              setEditing(false);
            }}
          >
            <X size={14} />
          </IconButton>
        </div>
      ) : null}

      <p className="text-sm text-muted-foreground">
        {columns.length} {columns.length === 1 ? "column" : "columns"} · {cards.length}{" "}
        {cards.length === 1 ? "card" : "cards"}
      </p>

      <div className="ml-auto hidden sm:block">
        <Button
          variant="subtle"
          size="sm"
          onClick={() => {
            setDraft(board.name);
            setEditing(true);
          }}
        >
          <Pencil size={14} aria-hidden="true" />
          Rename
        </Button>
      </div>
    </header>
  );
}

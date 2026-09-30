"use client";

import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  closestCorners,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragOverEvent,
  type DragStartEvent,
  type Over,
} from "@dnd-kit/core";
import {
  SortableContext,
  horizontalListSortingStrategy,
  sortableKeyboardCoordinates,
} from "@dnd-kit/sortable";
import { Columns3 } from "lucide-react";
import { useRef, useState } from "react";

import { AddColumnTile } from "@/components/board/add-column-tile";
import { BoardHeader } from "@/components/board/board-header";
import { BoardColumn } from "@/components/board/board-column";
import { CardOverlay } from "@/components/board/board-card";
import { CardDrawer } from "@/components/card/card-drawer";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { columnId } from "@/lib/constants";
import { columnIdOf, findCard, type BoardState } from "@/lib/board-state";
import { cn } from "@/lib/utils";
import type { Card, Column, CreateCardRequest } from "@/lib/types";

interface DropData {
  type?: string;
  cardId?: string;
  columnId?: string;
}

export interface BoardViewProps {
  state: BoardState;
  getState: () => BoardState | null;
  renameBoard: (name: string) => void;
  createColumn: (name: string) => Promise<boolean>;
  renameColumn: (columnId: string, name: string) => void;
  deleteColumn: (columnId: string) => void;
  moveColumn: (columnId: string, toIndex: number, snapshot?: BoardState) => void;
  createCard: (columnId: string, input: CreateCardRequest) => Promise<Card | null>;
  updateCard: (cardId: string, input: CreateCardRequest) => Promise<Card | null>;
  deleteCard: (cardId: string) => Promise<boolean>;
  moveCard: (
    cardId: string,
    toColumnId: string,
    toIndex: number,
    snapshot?: BoardState,
  ) => void;
  /** Local-only reposition used to preview a card crossing columns mid-drag. */
  previewCardMove: (cardId: string, toColumnId: string, toIndex: number) => void;
  restoreSnapshot: (snapshot: BoardState) => void;
}

export function BoardView({
  state,
  getState,
  renameBoard,
  createColumn,
  renameColumn,
  deleteColumn,
  moveColumn,
  createCard,
  updateCard,
  deleteCard,
  moveCard,
  previewCardMove,
  restoreSnapshot,
}: BoardViewProps) {
  const [activeCardId, setActiveCardId] = useState<string | null>(null);
  const [activeColumnId, setActiveColumnId] = useState<string | null>(null);
  const [openCard, setOpenCard] = useState<Card | null>(null);
  const [columnPendingDelete, setColumnPendingDelete] = useState<Column | null>(null);

  // State captured when the gesture starts: a failed drop rewinds the whole
  // drag, not just the last preview step the user happened to see.
  const snapshotRef = useRef<BoardState | null>(null);
  // Set once a cross-column preview has moved the card, which is what tells us
  // the local layout already reflects where it will land.
  const previewedRef = useRef(false);

  const sensors = useSensors(
    // A small activation distance keeps clicks on a card from becoming drags.
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const allCards = Object.values(state.cardsByColumn).flat();
  const activeCard = activeCardId ? findCard(state, activeCardId) : null;
  const activeColumn = activeColumnId
    ? state.columns.find((column) => column.id === activeColumnId) ?? null
    : null;

  const openCardColumn = openCard
    ? state.columns.find((column) => column.id === openCard.column_id)
    : undefined;

  const handleDragStart = (event: DragStartEvent) => {
    snapshotRef.current = getState() ?? state;
    previewedRef.current = false;
    const data = event.active.data.current as DropData | undefined;
    if (data?.type === "card" && data.cardId) {
      setActiveCardId(data.cardId);
      setActiveColumnId(null);
    } else if (data?.type === "column" && data.columnId) {
      setActiveColumnId(data.columnId);
      setActiveCardId(null);
    } else {
      setActiveCardId(null);
      setActiveColumnId(null);
    }
  };

  const handleDragOver = (event: DragOverEvent) => {
    const data = event.active.data.current as DropData | undefined;
    if (data?.type !== "card" || !data.cardId || !event.over) return;
    const live = getState() ?? state;
    const overData = event.over.data.current as DropData | undefined;
    const overColumnId = overData?.columnId;
    if (!overColumnId) return;
    // Only cross-column moves need a preview; reordering inside one column is
    // already animated by the sortable transform of the siblings.
    if (overColumnId === columnIdOf(live, data.cardId)) return;
    previewCardMove(data.cardId, overColumnId, indexForOver(live, event.over));
    previewedRef.current = true;
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const data = event.active.data.current as DropData | undefined;
    const over = event.over;
    const snapshot = snapshotRef.current ?? undefined;
    // Resolve the destination against the live state: cross-column previews may
    // have moved the card since the last render.
    const live = getState() ?? state;

    if (!over || !data) {
      // Dropped on nothing: rewind the previews, nothing was persisted.
      if (snapshot) restoreSnapshot(snapshot);
      setActiveCardId(null);
      setActiveColumnId(null);
      return;
    }

    if (data.type === "card" && data.cardId) {
      const currentColumnId = columnIdOf(live, data.cardId);
      const overData = over.data.current as DropData | undefined;
      if (currentColumnId) {
        let toColumnId: string;
        let toIndex: number;
        if (previewedRef.current) {
          // The preview already placed it; the current slot IS the destination.
          toColumnId = currentColumnId;
          toIndex = live.cardsByColumn[currentColumnId]!.findIndex(
            (card) => card.id === data.cardId,
          );
        } else if (overData?.columnId) {
          toColumnId = overData.columnId;
          toIndex = indexForOver(live, over);
        } else {
          if (snapshot) restoreSnapshot(snapshot);
          setActiveCardId(null);
          setActiveColumnId(null);
          return;
        }
        void moveCard(data.cardId, toColumnId, toIndex, snapshot);
      }
    } else if (data.type === "column" && data.columnId) {
      const overData = over.data.current as DropData | undefined;
      const toIndex = overData?.columnId
        ? live.columns.findIndex((column) => column.id === overData.columnId)
        : -1;
      const fromIndex = live.columns.findIndex((column) => column.id === data.columnId);
      if (toIndex >= 0 && toIndex !== fromIndex) {
        void moveColumn(data.columnId, toIndex, snapshot);
      }
    }

    setActiveCardId(null);
    setActiveColumnId(null);
  };

  const handleDragCancel = () => {
    if (snapshotRef.current) restoreSnapshot(snapshotRef.current);
    setActiveCardId(null);
    setActiveColumnId(null);
  };

  return (
    <div className="flex h-full min-h-0 flex-col">
      <BoardHeader
        board={state.board}
        columns={state.columns}
        cards={allCards}
        onRename={renameBoard}
      />

      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
        onDragCancel={handleDragCancel}
      >
        <div className="min-h-0 flex-1 overflow-x-auto overflow-y-hidden">
          {state.columns.length === 0 ? (
            <div className="px-4 pt-6 sm:px-6">
              <div className="max-w-md rounded-2xl border border-dashed border-border-strong bg-surface-muted/60 p-6 text-center">
                <Columns3 size={20} className="mx-auto text-muted-foreground" />
                <p className="mt-2 text-sm font-medium">This board has no columns yet</p>
                <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
                  Add a column to start shaping the workflow — the stages you move a
                  deal through.
                </p>
              </div>
            </div>
          ) : null}

          <SortableContext
            items={state.columns.map((column) => columnId(column.id))}
            strategy={horizontalListSortingStrategy}
          >
            <div
              className={cn(
                "flex h-full min-h-0 w-max items-stretch gap-3 p-4 sm:gap-4 sm:px-6",
                state.columns.length === 0 ? "h-auto items-start" : "",
              )}
            >
              {state.columns.map((column, index) => (
                <BoardColumn
                  key={column.id}
                  column={column}
                  cards={state.cardsByColumn[column.id] ?? []}
                  index={index}
                  total={state.columns.length}
                  cardInFlight={activeCard !== null}
                  onOpenCard={setOpenCard}
                  onRename={(name) => renameColumn(column.id, name)}
                  onDelete={() => setColumnPendingDelete(column)}
                  onMove={(toIndex) => moveColumn(column.id, toIndex)}
                  onCreateCard={async (title) => {
                    const card = await createCard(column.id, {
                      title,
                      priority: "medium",
                    });
                    return card !== null;
                  }}
                />
              ))}
              <AddColumnTile onCreate={createColumn} />
            </div>
          </SortableContext>
        </div>

        <DragOverlay
          dropAnimation={{ duration: 200, easing: "cubic-bezier(0.2, 0, 0, 1)" }}
        >
          {activeCard ? (
            <CardOverlay card={activeCard} />
          ) : activeColumn ? (
            <div className="w-72 rotate-1 rounded-2xl border border-accent/40 bg-surface-muted p-3 shadow-2xl shadow-black/25">
              <p className="truncate text-sm font-semibold text-foreground">
                {activeColumn.name}
              </p>
              <p className="mt-0.5 text-xs text-muted-foreground">
                {(state.cardsByColumn[activeColumn.id] ?? []).length} cards
              </p>
            </div>
          ) : null}
        </DragOverlay>
      </DndContext>

      {openCard ? (
        <CardDrawer
          key={openCard.id}
          card={openCard}
          columnName={openCardColumn?.name ?? "Unassigned"}
          knownAssignees={Array.from(
            new Set(allCards.map((card) => card.assignee.trim()).filter(Boolean)),
          )}
          onClose={() => setOpenCard(null)}
          onSave={async (input) => (await updateCard(openCard.id, input)) !== null}
          onDelete={() => deleteCard(openCard.id)}
        />
      ) : null}

      <ConfirmDialog
        open={columnPendingDelete !== null}
        onOpenChange={(next) => {
          if (!next) setColumnPendingDelete(null);
        }}
        destructive
        title={`Delete “${columnPendingDelete?.name ?? ""}”?`}
        confirmLabel="Delete column"
        description={
          <>
            <p>
              This removes the column and the{" "}
              {(columnPendingDelete
                ? (state.cardsByColumn[columnPendingDelete.id] ?? []).length
                : 0) ?? 0}{" "}
              card(s) inside it.
            </p>
            <p className="font-medium text-danger">This cannot be undone.</p>
          </>
        }
        onConfirm={async () => {
          if (!columnPendingDelete) return;
          await deleteColumn(columnPendingDelete.id);
          setColumnPendingDelete(null);
        }}
      />
    </div>
  );
}

/**
 * Where a card would land for this drop target: the hovered card's own slot,
 * or the end of the lane when the pointer is over the column itself. Always
 * computed against the state as it stands right now.
 */
function indexForOver(state: BoardState, over: Over): number {
  const data = over.data.current as DropData | undefined;
  if (!data?.columnId) return 0;
  const cards = state.cardsByColumn[data.columnId] ?? [];
  if (data.type === "card" && data.cardId) {
    const index = cards.findIndex((card) => card.id === data.cardId);
    if (index >= 0) return index;
  }
  return cards.length;
}

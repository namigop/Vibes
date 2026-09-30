"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import * as api from "@/lib/api";
import {
  appendCard,
  appendColumn,
  columnIdOf,
  createBoardState,
  findCard,
  moveCardInState,
  moveColumnInState,
  reconcileCard,
  removeCard,
  removeColumn,
  renameBoardInState,
  replaceBoard,
  replaceCard,
  replaceColumn,
  replaceColumns,
  type BoardState,
} from "@/lib/board-state";
import type { Card, CreateCardRequest, Priority } from "@/lib/types";
import { useBoards } from "@/providers/boards-provider";
import { useToast } from "@/providers/toast-provider";

type Status = "loading" | "ready" | "error";

/** One request, no React: the caller decides what to do with the result. */
async function fetchBoardDetail(boardId: string): Promise<BoardState> {
  const detail = await api.getBoard(boardId);
  return createBoardState(detail.board, detail.columns, detail.cards);
}

export interface UseBoardResult {
  state: BoardState | null;
  status: Status;
  error: string | null;
  reload: () => void;
  /**
   * Reads the live state, including moves applied by drag previews that have
   * not re-rendered yet. Drag handlers must resolve their drop target against
   * this rather than the render closure.
   */
  getState: () => BoardState | null;

  renameBoard: (name: string) => Promise<void>;
  createColumn: (name: string) => Promise<boolean>;
  renameColumn: (columnId: string, name: string) => Promise<void>;
  deleteColumn: (columnId: string) => Promise<boolean>;
  moveColumn: (columnId: string, toIndex: number, snapshot?: BoardState) => Promise<void>;

  createCard: (columnId: string, input: CreateCardRequest) => Promise<Card | null>;
  updateCard: (cardId: string, input: CreateCardRequest) => Promise<Card | null>;
  deleteCard: (cardId: string) => Promise<boolean>;
  /** Local-only reposition: previews a cross-column drag, never hits the API. */
  previewCardMove: (cardId: string, toColumnId: string, toIndex: number) => void;
  /** Local-only rollback of a drag that ended outside any drop target. */
  restoreSnapshot: (snapshot: BoardState) => void;
  /**
   * `toIndex` is the destination index *after* the move, exactly as
   * `POST /cards/{id}/move` expects. `snapshot` lets a drag pass the state it
   * captured at drag start, so a failed drop rewinds the whole gesture rather
   * than the single last step the user saw.
   */
  moveCard: (
    cardId: string,
    toColumnId: string,
    toIndex: number,
    snapshot?: BoardState,
  ) => Promise<void>;
}

/**
 * Owns one board's data and every write against it.
 *
 * Two rules hold for all mutations:
 *  1. Local state is updated first and the request goes out after — the UI
 *     never waits on the network.
 *  2. If the request fails, the pre-mutation snapshot is restored and the
 *     server's `{ message }` is surfaced in a toast.
 *
 * Writes are sequenced per entity: every mutation bumps a counter for its key
 * and only applies the server's response if no newer write to the same key was
 * started in the meantime. A slow response can therefore never clobber a newer
 * local state, and cannot roll back over it either.
 */
export function useBoard(boardId: string): UseBoardResult {
  const [state, setState] = useState<BoardState | null>(null);
  const [status, setStatus] = useState<Status>("loading");
  const [error, setError] = useState<string | null>(null);

  const stateRef = useRef<BoardState | null>(null);
  const writeSeq = useRef(new Map<string, number>());

  const toast = useToast();
  const { applyBoard } = useBoards();

  // Ref + state are written together so async continuations always read the
  // latest committed value, never a stale render closure.
  const commit = useCallback((next: BoardState | null) => {
    stateRef.current = next;
    setState(next);
  }, []);

  const beginWrite = useCallback((key: string) => {
    const next = (writeSeq.current.get(key) ?? 0) + 1;
    writeSeq.current.set(key, next);
    return next;
  }, []);

  const isLatestWrite = useCallback(
    (key: string, seq: number) => writeSeq.current.get(key) === seq,
    [],
  );

  const fail = useCallback(
    (cause: unknown, title: string) => {
      toast.error(title, api.describeError(cause));
    },
    [toast],
  );

  // The effect only *starts* the request; every setState happens in the
  // promise callbacks (the pattern the React compiler lint rules want, and it
  // avoids a cascading re-render on mount). `status` starts as "loading" and
  // `BoardScreen` is keyed by boardId, so nothing stale is ever rendered.
  useEffect(() => {
    let ignore = false;
    fetchBoardDetail(boardId).then(
      (next) => {
        if (ignore) return;
        commit(next);
        setStatus("ready");
      },
      (cause: unknown) => {
        if (ignore) return;
        setError(api.describeError(cause));
        setStatus("error");
      },
    );
    return () => {
      ignore = true;
    };
  }, [boardId, commit]);

  const reload = useCallback(() => {
    setError(null);
    setStatus("loading");
    fetchBoardDetail(boardId).then(
      (next) => {
        commit(next);
        setStatus("ready");
      },
      (cause: unknown) => {
        setError(api.describeError(cause));
        setStatus("error");
      },
    );
  }, [boardId, commit]);

  /* ------------------------------- board ------------------------------ */

  const renameBoard = useCallback(
    async (name: string) => {
      const before = stateRef.current;
      const trimmed = name.trim();
      if (!before || !trimmed) return;

      const key = `board:${before.board.id}`;
      const seq = beginWrite(key);
      commit(renameBoardInState(before, trimmed));
      try {
        const saved = await api.updateBoard(before.board.id, trimmed);
        if (!isLatestWrite(key, seq)) return;
        commit(replaceBoard(stateRef.current ?? before, saved));
        applyBoard(saved);
        toast.success("Board renamed");
      } catch (cause) {
        if (!isLatestWrite(key, seq)) return;
        commit(before);
        fail(cause, "Could not rename the board");
      }
    },
    [applyBoard, beginWrite, commit, fail, isLatestWrite, toast],
  );

  /* ------------------------------ columns ----------------------------- */

  const createColumn = useCallback(
    async (name: string) => {
      const before = stateRef.current;
      const trimmed = name.trim();
      if (!before || !trimmed) return false;
      try {
        const column = await api.createColumn(before.board.id, trimmed);
        if (stateRef.current) commit(appendColumn(stateRef.current, column));
        toast.success("Column added", trimmed);
        return true;
      } catch (cause) {
        fail(cause, "Could not add the column");
        return false;
      }
    },
    [commit, fail, toast],
  );

  const renameColumn = useCallback(
    async (columnId: string, name: string) => {
      const before = stateRef.current;
      const trimmed = name.trim();
      if (!before || !trimmed) return;

      const key = `column:${columnId}`;
      const seq = beginWrite(key);
      const existing = before.columns.find((column) => column.id === columnId);
      if (!existing) return;
      commit(replaceColumn(before, { ...existing, name: trimmed }));
      try {
        const saved = await api.updateColumn(columnId, trimmed);
        if (!isLatestWrite(key, seq)) return;
        const current = stateRef.current;
        if (current) commit(replaceColumn(current, saved));
        toast.success("Column renamed");
      } catch (cause) {
        if (!isLatestWrite(key, seq)) return;
        commit(before);
        fail(cause, "Could not rename the column");
      }
    },
    [beginWrite, commit, fail, isLatestWrite, toast],
  );

  const deleteColumn = useCallback(
    async (columnId: string) => {
      const before = stateRef.current;
      if (!before) return false;
      const key = `column:${columnId}`;
      const seq = beginWrite(key);
      const column = before.columns.find((item) => item.id === columnId);
      if (!column) return false;
      commit(removeColumn(before, columnId));
      try {
        await api.deleteColumn(columnId);
        if (!isLatestWrite(key, seq)) return false;
        toast.success("Column deleted", column.name);
        return true;
      } catch (cause) {
        if (!isLatestWrite(key, seq)) return false;
        commit(before);
        fail(cause, "Could not delete the column");
        return false;
      }
    },
    [beginWrite, commit, fail, isLatestWrite, toast],
  );

  const moveColumn = useCallback(
    async (columnId: string, toIndex: number, snapshot?: BoardState) => {
      const before = stateRef.current;
      if (!before) return;
      const key = `column:${columnId}`;
      const seq = beginWrite(key);
      const rollback = snapshot ?? before;
      commit(moveColumnInState(before, columnId, toIndex));
      try {
        // The server reindexes every column and returns them in order, so the
        // local list is replaced wholesale rather than patched.
        const columns = await api.moveColumn(columnId, toIndex);
        if (!isLatestWrite(key, seq)) return;
        const current = stateRef.current;
        if (current) commit(replaceColumns(current, columns));
      } catch (cause) {
        if (!isLatestWrite(key, seq)) return;
        commit(rollback);
        fail(cause, "Could not move the column");
      }
    },
    [beginWrite, commit, fail, isLatestWrite],
  );

  /* ------------------------------- cards ------------------------------ */

  const createCard = useCallback(
    async (columnId: string, input: CreateCardRequest) => {
      try {
        // Creates are not optimistic: the server owns the UUID and a temporary
        // id would be a lie the moment anything else touches the card. The
        // composers show a pending state instead.
        const card = await api.createCard(columnId, input);
        if (stateRef.current) commit(appendCard(stateRef.current, card));
        return card;
      } catch (cause) {
        fail(cause, "Could not add the card");
        return null;
      }
    },
    [commit, fail],
  );

  const updateCard = useCallback(
    async (cardId: string, input: CreateCardRequest) => {
      const before = stateRef.current;
      const existing = before ? findCard(before, cardId) : null;
      if (!before || !existing) return null;

      const key = `card:${cardId}`;
      const seq = beginWrite(key);
      const draft: Card = {
        ...existing,
        title: input.title.trim(),
        description: input.description ?? "",
        assignee: input.assignee?.trim() ?? "",
        priority: (input.priority ?? existing.priority) as Priority,
        due_date: input.due_date ?? null,
      };
      commit(replaceCard(before, draft));
      try {
        const saved = await api.updateCard(cardId, {
          title: draft.title,
          description: draft.description,
          assignee: draft.assignee,
          priority: draft.priority,
          due_date: draft.due_date,
        });
        if (!isLatestWrite(key, seq)) return null;
        const current = stateRef.current;
        if (current) commit(reconcileCard(current, saved));
        return saved;
      } catch (cause) {
        if (!isLatestWrite(key, seq)) return null;
        commit(before);
        fail(cause, "Could not save the card");
        return null;
      }
    },
    [beginWrite, commit, fail, isLatestWrite],
  );

  const deleteCard = useCallback(
    async (cardId: string) => {
      const before = stateRef.current;
      const existing = before ? findCard(before, cardId) : null;
      if (!before || !existing) return false;
      const key = `card:${cardId}`;
      const seq = beginWrite(key);
      commit(removeCard(before, cardId));
      try {
        await api.deleteCard(cardId);
        if (!isLatestWrite(key, seq)) return false;
        toast.success("Card deleted", existing.title);
        return true;
      } catch (cause) {
        if (!isLatestWrite(key, seq)) return false;
        commit(before);
        fail(cause, "Could not delete the card");
        return false;
      }
    },
    [beginWrite, commit, fail, isLatestWrite, toast],
  );

  const moveCard = useCallback(
    async (cardId: string, toColumnId: string, toIndex: number, snapshot?: BoardState) => {
      const before = stateRef.current;
      if (!before) return;
      const key = `card:${cardId}`;
      const seq = beginWrite(key);
      const rollback = snapshot ?? before;

      // During a drag the preview has usually already put the card in place;
      // this call is a no-op then, and the real work is persisting it.
      commit(moveCardInState(before, cardId, toColumnId, toIndex));

      try {
        const saved = await api.moveCard(cardId, {
          column_id: toColumnId,
          position: toIndex,
        });
        if (!isLatestWrite(key, seq)) return;
        const current = stateRef.current;
        if (current) commit(reconcileCard(current, saved));
      } catch (cause) {
        if (!isLatestWrite(key, seq)) return;
        commit(rollback);
        fail(cause, "Could not move the card");
      }
    },
    [beginWrite, commit, fail, isLatestWrite],
  );

  /** Local-only: moves the card in the DOM while a drag is still in flight. */
  const previewCardMove = useCallback(
    (cardId: string, toColumnId: string, toIndex: number) => {
      const before = stateRef.current;
      if (!before) return;
      commit(moveCardInState(before, cardId, toColumnId, toIndex));
    },
    [commit],
  );

  /** Local-only: rewind to a pre-drag snapshot (dropped on nothing / Esc). */
  const restoreSnapshot = useCallback(
    (snapshot: BoardState) => {
      commit(snapshot);
    },
    [commit],
  );

  return {
    state,
    status,
    error,
    reload,
    getState: () => stateRef.current,
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
  };
}

/** Position of a card inside its column, or -1. */
export function indexOfCard(state: BoardState, cardId: string): number {
  const columnId = columnIdOf(state, cardId);
  if (!columnId) return -1;
  return state.cardsByColumn[columnId]!.findIndex((card) => card.id === cardId);
}

import type { Board, Card, Column } from "@/lib/types";

/**
 * The single source of truth for a rendered board. `GET /boards/{id}` returns a
 * flat card list; grouping happens here, once, so no component has to.
 */
export interface BoardState {
  board: Board;
  /** Ordered by position, always gap-free 0..n-1. */
  columns: Column[];
  /** columnId -> cards ordered by position. */
  cardsByColumn: Record<string, Card[]>;
}

export function sortByPosition<T extends { position: number }>(items: T[]): T[] {
  return [...items].sort((a, b) => a.position - b.position);
}

export function groupCardsByColumn(columns: Column[], cards: Card[]): Record<string, Card[]> {
  const grouped: Record<string, Card[]> = {};
  for (const column of columns) grouped[column.id] = [];
  for (const card of cards) {
    const bucket = grouped[card.column_id];
    // A card whose column vanished (deleted mid-session) is dropped rather than
    // crashing the render.
    if (bucket) bucket.push(card);
  }
  for (const key of Object.keys(grouped)) grouped[key] = sortByPosition(grouped[key]);
  return grouped;
}

export function createBoardState(
  board: Board,
  columns: Column[],
  cards: Card[],
): BoardState {
  const orderedColumns = sortByPosition(columns);
  return {
    board,
    columns: orderedColumns,
    cardsByColumn: groupCardsByColumn(orderedColumns, cards),
  };
}

/** Re-stamp a card list with gap-free 0..n-1 positions. */
function reindex(cards: Card[]): Card[] {
  return cards.map((card, index) =>
    card.position === index ? card : { ...card, position: index },
  );
}

export function columnIdOf(state: BoardState, cardId: string): string | null {
  for (const [columnId, cards] of Object.entries(state.cardsByColumn)) {
    if (cards.some((card) => card.id === cardId)) return columnId;
  }
  return null;
}

export function findCard(state: BoardState, cardId: string): Card | null {
  for (const cards of Object.values(state.cardsByColumn)) {
    const found = cards.find((card) => card.id === cardId);
    if (found) return found;
  }
  return null;
}

/**
 * Pure card move, used for the optimistic update. `toIndex` is the desired
 * index inside the destination column *after* the move; out-of-range values are
 * clamped, mirroring the server.
 */
export function moveCardInState(
  state: BoardState,
  cardId: string,
  toColumnId: string,
  toIndex: number,
): BoardState {
  const fromColumnId = columnIdOf(state, cardId);
  if (!fromColumnId || !state.cardsByColumn[toColumnId]) return state;

  const card = findCard(state, cardId);
  if (!card) return state;

  const source = state.cardsByColumn[fromColumnId]!;
  const without = source.filter((item) => item.id !== cardId);

  if (fromColumnId === toColumnId) {
    const index = Math.max(0, Math.min(toIndex, without.length));
    const reordered = [...without.slice(0, index), card, ...without.slice(index)];
    return {
      ...state,
      cardsByColumn: { ...state.cardsByColumn, [toColumnId]: reindex(reordered) },
    };
  }

  const target = state.cardsByColumn[toColumnId]!;
  const index = Math.max(0, Math.min(toIndex, target.length));
  const moved = { ...card, column_id: toColumnId };
  const nextTarget = [...target.slice(0, index), moved, ...target.slice(index)];

  return {
    ...state,
    cardsByColumn: {
      ...state.cardsByColumn,
      [fromColumnId]: reindex(without),
      [toColumnId]: reindex(nextTarget),
    },
  };
}

export function moveColumnInState(
  state: BoardState,
  columnId: string,
  toIndex: number,
): BoardState {
  const from = state.columns.findIndex((column) => column.id === columnId);
  if (from === -1) return state;
  const without = state.columns.filter((column) => column.id !== columnId);
  const index = Math.max(0, Math.min(toIndex, without.length));
  const moved = state.columns[from]!;
  const columns = [...without.slice(0, index), moved, ...without.slice(index)];
  return {
    ...state,
    columns: columns.map((column, i) => ({ ...column, position: i })),
  };
}

/** Replace the whole column list (the server returns the full reindexed set). */
export function replaceColumns(state: BoardState, columns: Column[]): BoardState {
  const ordered = sortByPosition(columns).map((column, index) => ({
    ...column,
    position: index,
  }));
  const known = new Set(ordered.map((column) => column.id));
  const cardsByColumn: Record<string, Card[]> = {};
  for (const column of ordered) cardsByColumn[column.id] = [];
  for (const [columnId, cards] of Object.entries(state.cardsByColumn)) {
    if (known.has(columnId)) cardsByColumn[columnId] = cards;
  }
  return { ...state, columns: ordered, cardsByColumn };
}

export function appendCard(state: BoardState, card: Card): BoardState {
  const bucket = state.cardsByColumn[card.column_id];
  if (!bucket) return state;
  return {
    ...state,
    cardsByColumn: {
      ...state.cardsByColumn,
      [card.column_id]: reindex([...bucket, card]),
    },
  };
}

export function replaceCard(state: BoardState, card: Card): BoardState {
  const bucket = state.cardsByColumn[card.column_id];
  if (!bucket) return state;
  return {
    ...state,
    cardsByColumn: {
      ...state.cardsByColumn,
      [card.column_id]: reindex(
        bucket.map((item) => (item.id === card.id ? card : item)),
      ),
    },
  };
}

export function removeCard(state: BoardState, cardId: string): BoardState {
  const columnId = columnIdOf(state, cardId);
  if (!columnId) return state;
  return {
    ...state,
    cardsByColumn: {
      ...state.cardsByColumn,
      [columnId]: reindex(
        state.cardsByColumn[columnId]!.filter((card) => card.id !== cardId),
      ),
    },
  };
}

export function appendColumn(state: BoardState, column: Column): BoardState {
  return {
    ...state,
    columns: [...state.columns, { ...column, position: state.columns.length }],
    cardsByColumn: { ...state.cardsByColumn, [column.id]: [] },
  };
}

export function removeColumn(state: BoardState, columnId: string): BoardState {
  const cardsByColumn = { ...state.cardsByColumn };
  delete cardsByColumn[columnId];
  return {
    ...state,
    columns: state.columns
      .filter((column) => column.id !== columnId)
      .map((column, index) => ({ ...column, position: index })),
    cardsByColumn,
  };
}

/** Rename/replace a column in place. */
export function replaceColumn(state: BoardState, column: Column): BoardState {
  if (!state.cardsByColumn[column.id]) return state;
  return {
    ...state,
    columns: state.columns.map((item) => (item.id === column.id ? column : item)),
  };
}

export function renameBoardInState(state: BoardState, name: string): BoardState {
  return { ...state, board: { ...state.board, name } };
}

/** Adopt the server's board record wholesale (after a PATCH). */
export function replaceBoard(state: BoardState, board: Board): BoardState {
  return { ...state, board };
}

/**
 * Fold a server response back into local state: the card lands exactly where
 * the server settled it (it may have clamped the requested position) and its
 * server-side field values win.
 */
export function reconcileCard(state: BoardState, card: Card): BoardState {
  const moved = moveCardInState(state, card.id, card.column_id, card.position);
  return replaceCard(moved, card);
}

export function totalCardCount(state: BoardState): number {
  return Object.values(state.cardsByColumn).reduce((sum, cards) => sum + cards.length, 0);
}

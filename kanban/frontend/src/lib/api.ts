import { API_BASE_URL } from "@/lib/constants";
import type {
  ApiErrorEnvelope,
  Board,
  BoardDetail,
  Card,
  Column,
  CreateCardRequest,
  HealthResponse,
  MoveCardRequest,
  MoveColumnRequest,
  Priority,
} from "@/lib/types";

/**
 * Every non-2xx response is turned into an ApiError carrying the contract's
 * `{ error, message }` envelope so callers can show `message` to the user and
 * branch on `error`. Network failures become status 0 / "network_error".
 */
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly path: string;

  constructor(status: number, code: string, message: string, path: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.path = path;
  }
}

const FALLBACK_MESSAGES: Record<number, string> = {
  400: "The request was rejected as invalid.",
  404: "That item no longer exists — it may have been deleted.",
  409: "That name is already taken.",
  500: "The server hit an unexpected error.",
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${path}`;

  let response: Response;
  try {
    response = await fetch(url, {
      ...init,
      headers: {
        Accept: "application/json",
        ...(init?.body ? { "Content-Type": "application/json" } : null),
        ...init?.headers,
      },
      // Board data is user-owned and changes constantly: never serve it stale.
      cache: "no-store",
    });
  } catch {
    throw new ApiError(
      0,
      "network_error",
      `Could not reach the API at ${API_BASE_URL}. Is the backend running?`,
      path,
    );
  }

  if (response.status === 204) return undefined as T;

  const raw = await response.text();
  const body = raw ? (safeParse(raw) as unknown) : null;

  if (!response.ok) {
    const envelope = isErrorEnvelope(body) ? body : null;
    throw new ApiError(
      response.status,
      envelope?.error ?? "unknown_error",
      envelope?.message ??
        FALLBACK_MESSAGES[response.status] ??
        `Request failed with status ${response.status}.`,
      path,
    );
  }

  if (body === null) {
    throw new ApiError(
      response.status,
      "invalid_response",
      "The server returned a response that was not valid JSON.",
      path,
    );
  }
  return body as T;
}

function safeParse(raw: string): unknown {
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function isErrorEnvelope(value: unknown): value is ApiErrorEnvelope {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as ApiErrorEnvelope).error === "string" &&
    typeof (value as ApiErrorEnvelope).message === "string"
  );
}

function jsonBody(body: unknown): string {
  return JSON.stringify(body);
}

/* ------------------------------------------------------------------ *
 * Normalisation
 *
 * The contract notes two traps the client must absorb:
 *  1. a `null` slice (columns/cards) must become `[]` so `.map` never throws;
 *  2. `cards` arrive flat and have to be grouped by column client-side.
 * ------------------------------------------------------------------ */

function toPriority(value: unknown): Priority {
  return value === "low" || value === "medium" || value === "high" || value === "urgent"
    ? value
    : "medium";
}

export function normaliseCard(raw: Partial<Card> & { id: string; column_id: string; title: string }): Card {
  return {
    id: raw.id,
    column_id: raw.column_id,
    title: raw.title ?? "",
    description: raw.description ?? "",
    assignee: raw.assignee ?? "",
    priority: toPriority(raw.priority),
    due_date: raw.due_date ?? null,
    position: typeof raw.position === "number" ? raw.position : 0,
    created_at: raw.created_at ?? "",
    updated_at: raw.updated_at ?? "",
  };
}

export function normaliseBoardDetail(raw: BoardDetail): BoardDetail {
  return {
    board: raw.board,
    columns: raw.columns ?? [],
    cards: (raw.cards ?? []).map(normaliseCard),
  };
}

/* ------------------------------ Boards ----------------------------- */

export function listBoards(): Promise<Board[]> {
  return request<Board[]>("/boards");
}

export function createBoard(name: string): Promise<Board> {
  return request<Board>("/boards", { method: "POST", body: jsonBody({ name }) });
}

export function getBoard(id: string): Promise<BoardDetail> {
  return request<BoardDetail>(`/boards/${id}`).then(normaliseBoardDetail);
}

export function updateBoard(id: string, name: string): Promise<Board> {
  return request<Board>(`/boards/${id}`, {
    method: "PATCH",
    body: jsonBody({ name }),
  });
}

export function deleteBoard(id: string): Promise<void> {
  return request<void>(`/boards/${id}`, { method: "DELETE" });
}

/* ------------------------------ Columns ---------------------------- */

export function createColumn(boardId: string, name: string): Promise<Column> {
  return request<Column>(`/boards/${boardId}/columns`, {
    method: "POST",
    body: jsonBody({ name }),
  });
}

export function updateColumn(id: string, name: string): Promise<Column> {
  return request<Column>(`/columns/${id}`, {
    method: "PATCH",
    body: jsonBody({ name }),
  });
}

/** Reindexes every column of the board and returns them in the new order. */
export function moveColumn(id: string, position: number): Promise<Column[]> {
  return request<Column[]>(`/columns/${id}/move`, {
    method: "POST",
    body: jsonBody({ position } satisfies MoveColumnRequest),
  });
}

export function deleteColumn(id: string): Promise<void> {
  return request<void>(`/columns/${id}`, { method: "DELETE" });
}

/* ------------------------------- Cards ----------------------------- */

export function createCard(columnId: string, payload: CreateCardRequest): Promise<Card> {
  return request<Card>(`/columns/${columnId}/cards`, {
    method: "POST",
    body: jsonBody(payload),
  }).then(normaliseCard);
}

export function updateCard(id: string, payload: CreateCardRequest): Promise<Card> {
  return request<Card>(`/cards/${id}`, {
    method: "PATCH",
    body: jsonBody(payload),
  }).then(normaliseCard);
}

export function moveCard(id: string, payload: MoveCardRequest): Promise<Card> {
  return request<Card>(`/cards/${id}/move`, {
    method: "POST",
    body: jsonBody(payload),
  }).then(normaliseCard);
}

export function deleteCard(id: string): Promise<void> {
  return request<void>(`/cards/${id}`, { method: "DELETE" });
}

/* ------------------------------- Health ---------------------------- */

export function health(): Promise<HealthResponse> {
  return request<HealthResponse>("/health");
}

/** Human-facing text for anything thrown by this module. */
export function describeError(error: unknown): string {
  if (error instanceof ApiError) return error.message;
  if (error instanceof Error && error.message) return error.message;
  return "Something went wrong. Please try again.";
}

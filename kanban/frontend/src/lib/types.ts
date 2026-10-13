import type { PRIORITIES } from "@/lib/constants";

export type Priority = (typeof PRIORITIES)[number];

/** The signed-in account. The API never sends password_hash. */
export interface User {
  id: string;
  username: string;
}

/** Response body of POST /auth/login. */
export interface LoginResponse {
  token: string;
  token_type: string;
  expires_at: string;
  user: User;
}

export interface Board {
  id: string;
  name: string;
  position: number;
  /** The single user who can see this board. */
  owner_id: string;
  created_at: string;
  updated_at: string;
}

export interface Column {
  id: string;
  board_id: string;
  name: string;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface Card {
  id: string;
  column_id: string;
  title: string;
  /** "" when unset — the server normalises null/"" to "". */
  description: string;
  /** "" when unassigned.
   *
   *  Free text naming a person involved in the work — deliberately NOT an
   *  account. Boards are owner-only, so a user reference here would name
   *  someone who by construction cannot open this card.
   */
  assignee: string;
  priority: Priority;
  /** Calendar date `YYYY-MM-DD`, or null. Never a timestamp. */
  due_date: string | null;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface BoardDetail {
  board: Board;
  /** Ordered by position. */
  columns: Column[];
  /** Ordered by (column.position, position). */
  cards: Card[];
}

export interface ApiErrorEnvelope {
  /** Stable machine-readable slug, e.g. "validation_failed", "unauthorized". */
  error: string;
  /** Human-readable, safe to show in a toast. */
  message: string;
}

/** Sent verbatim for both `POST /columns/{id}/cards` and `PATCH /cards/{id}`. */
export interface CreateCardRequest {
  title: string;
  description?: string;
  assignee?: string;
  priority?: Priority;
  due_date?: string | null;
}

export interface MoveCardRequest {
  column_id: string;
  position: number;
}

export interface MoveColumnRequest {
  position: number;
}

export interface HealthResponse {
  status: string;
  db: string;
}

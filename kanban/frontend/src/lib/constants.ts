/**
 * App-wide constants.
 *
 * The API base URL is a NEXT_PUBLIC_ variable so it is inlined at build time by
 * Next.js; it is NOT read at runtime from the server. Change it in `.env.local`
 * (dev) or in the deployment environment (prod) and rebuild.
 */
export const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";

export const API_BASE_URL = stripTrailingSlash(
  process.env.NEXT_PUBLIC_API_BASE_URL || DEFAULT_API_BASE_URL,
);

/** localStorage key holding the user's explicit theme choice. */
export const THEME_STORAGE_KEY = "kanban.theme";

/** Sortable id namespaces. Cards and columns share one DndContext, so their ids
 *  must live in disjoint spaces to avoid a collision on drag. */
export const cardId = (id: string) => `card::${id}`;
export const columnId = (id: string) => `column::${id}`;
export const columnZoneId = (id: string) => `columnzone::${id}`;

export const PRIORITIES = ["low", "medium", "high", "urgent"] as const;

function stripTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

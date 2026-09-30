import type { ApiError, Todo } from "./types";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

class ApiClientError extends Error {
  readonly status: number;
  readonly code: string;
  readonly details?: ApiError["error"]["details"];

  constructor(status: number, code: string, message: string, details?: ApiError["error"]["details"]) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
    this.code = code;
    this.details = details;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...(init?.headers ?? {}),
    },
  });

  if (res.status === 204) {
    // No-content success path (DELETE).
    return undefined as T;
  }

  const text = await res.text();
  const body = text ? safeJson(text) : undefined;

  if (!res.ok) {
    const errBody = (body as ApiError | undefined)?.error;
    throw new ApiClientError(
      res.status,
      errBody?.code ?? "unknown_error",
      errBody?.message ?? `request failed with status ${res.status}`,
      errBody?.details,
    );
  }

  return body as T;
}

function safeJson(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}

// ---- Typed wrappers ---------------------------------------------------------

export const api = {
  listTodos(): Promise<{ items: Todo[] }> {
    return request<{ items: Todo[] }>("/todos", { method: "GET" });
  },

  createTodo(input: { title: string; description?: string }): Promise<Todo> {
    return request<Todo>("/todos", {
      method: "POST",
      body: JSON.stringify({
        title: input.title,
        description: input.description ?? "",
      }),
    });
  },

  updateTodo(
    id: string,
    patch: { title?: string; description?: string; done?: boolean },
  ): Promise<Todo> {
    return request<Todo>(`/todos/${encodeURIComponent(id)}`, {
      method: "PATCH",
      body: JSON.stringify(patch),
    });
  },

  deleteTodo(id: string): Promise<void> {
    return request<void>(`/todos/${encodeURIComponent(id)}`, { method: "DELETE" });
  },
};

export { ApiClientError };

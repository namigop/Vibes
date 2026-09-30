// Todo mirrors the Go backend DTO exactly. Field names are snake_case to match
// the JSON returned by the Go handlers.
export interface Todo {
  id: string;
  title: string;
  description: string;
  done: boolean;
  created_at: string; // ISO-8601 string
  updated_at: string;
}

// Error envelope returned by the Go API on non-2xx responses.
export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Array<{ field: string; message: string }>;
  };
}

"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { TodoList } from "@/components/TodoList";
import { NewTodoForm } from "@/components/NewTodoForm";
import { ErrorBanner } from "@/components/ErrorBanner";

const TODOS_KEY = ["todos"] as const;

export default function HomePage() {
  const qc = useQueryClient();

  const todosQuery = useQuery({
    queryKey: TODOS_KEY,
    queryFn: () => api.listTodos(),
  });

  // Single invalidate helper used by every mutation in child components.
  const onMutated = () => qc.invalidateQueries({ queryKey: TODOS_KEY });

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-2xl flex-col gap-6 px-4 py-8">
      <header>
        <h1 className="text-2xl font-semibold tracking-tight">Todo MVP</h1>
        <p className="text-sm text-slate-500 dark:text-slate-400">
          A minimal CRUD list backed by Go + Postgres. ERIK
        </p>
      </header>

      <NewTodoForm onCreated={onMutated} />

      {todosQuery.isError ? (
        <ErrorBanner error={todosQuery.error} />
      ) : null}

      <section aria-label="Todos">
        {todosQuery.isPending ? (
          <p className="text-sm text-slate-500">Loading…</p>
        ) : todosQuery.data ? (
          <TodoList items={todosQuery.data.items} onMutated={onMutated} />
        ) : null}
      </section>

      <footer className="mt-auto pt-4 text-center text-xs text-slate-400">
        {todosQuery.data ? `${todosQuery.data.items.length} item(s)` : ""}
      </footer>
    </main>
  );
}

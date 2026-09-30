"use client";

import type { Todo } from "@/lib/types";
import { TodoItem } from "./TodoItem";

interface Props {
  items: Todo[];
  onMutated: () => void;
}

export function TodoList({ items, onMutated }: Props) {
  if (items.length === 0) {
    return (
      <p className="rounded-md border border-dashed border-slate-300 p-6 text-center text-sm text-slate-500 dark:border-slate-700 dark:text-slate-400">
        Nothing here yet — add your first todo above.
      </p>
    );
  }

  return (
    <ul className="flex flex-col gap-2" data-testid="todo-list">
      {items.map((t) => (
        <TodoItem key={t.id} todo={t} onMutated={onMutated} />
      ))}
    </ul>
  );
}

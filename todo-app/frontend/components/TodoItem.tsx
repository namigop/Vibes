"use client";

import { useState, useRef, useEffect } from "react";
import clsx from "clsx";
import type { Todo } from "@/lib/types";
import { api } from "@/lib/api";

interface Props {
  todo: Todo;
  onMutated: () => void;
}

export function TodoItem({ todo, onMutated }: Props) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(todo.title);
  const [busy, setBusy] = useState(false);
  const inputRef = useRef<HTMLInputElement | null>(null);

  useEffect(() => {
    if (editing) {
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [editing]);

  async function toggleDone() {
    setBusy(true);
    try {
      await api.updateTodo(todo.id, { done: !todo.done });
      onMutated();
    } finally {
      setBusy(false);
    }
  }

  async function saveEdit() {
    const next = draft.trim();
    if (next === "" || next === todo.title) {
      setDraft(todo.title);
      setEditing(false);
      return;
    }
    setBusy(true);
    try {
      await api.updateTodo(todo.id, { title: next });
      onMutated();
      setEditing(false);
    } finally {
      setBusy(false);
    }
  }

  async function remove() {
    if (!confirm(`Delete "${todo.title}"?`)) return;
    setBusy(true);
    try {
      await api.deleteTodo(todo.id);
      onMutated();
    } finally {
      setBusy(false);
    }
  }

  return (
    <li
      className={clsx(
        "flex items-start gap-3 rounded-md border border-slate-200 bg-white p-3 shadow-sm transition",
        "dark:border-slate-700 dark:bg-slate-800",
        busy && "opacity-60",
        todo.done && "bg-slate-50 dark:bg-slate-800/60",
      )}
    >
      <input
        type="checkbox"
        checked={todo.done}
        onChange={toggleDone}
        disabled={busy}
        aria-label={todo.done ? "Mark as not done" : "Mark as done"}
        className="mt-1 h-4 w-4 cursor-pointer accent-blue-600"
      />

      <div className="flex-1">
        {editing ? (
          <input
            ref={inputRef}
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onBlur={saveEdit}
            onKeyDown={(e) => {
              if (e.key === "Enter") saveEdit();
              if (e.key === "Escape") {
                setDraft(todo.title);
                setEditing(false);
              }
            }}
            disabled={busy}
            className="w-full rounded border border-blue-400 bg-white px-2 py-1 text-sm focus:outline-none dark:bg-slate-900"
          />
        ) : (
          <button
            type="button"
            onClick={() => setEditing(true)}
            disabled={busy}
            className={clsx(
              "block w-full text-left text-sm",
              todo.done && "text-slate-400 line-through",
            )}
          >
            {todo.title}
          </button>
        )}
        {todo.description ? (
          <p
            className={clsx(
              "mt-1 text-xs text-slate-500 dark:text-slate-400",
              todo.done && "line-through",
            )}
          >
            {todo.description}
          </p>
        ) : null}
      </div>

      <button
        type="button"
        onClick={remove}
        disabled={busy}
        className="rounded-md border border-red-300 px-2 py-1 text-xs text-red-700 hover:bg-red-50 disabled:opacity-50 dark:border-red-700 dark:text-red-300 dark:hover:bg-red-950"
      >
        Delete
      </button>
    </li>
  );
}

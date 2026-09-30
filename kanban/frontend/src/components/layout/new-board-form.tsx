"use client";

import { useRouter } from "next/navigation";
import { useEffect, useRef, useState, type FormEvent } from "react";

import { Button } from "@/components/ui/button";
import { useBoards } from "@/providers/boards-provider";
import { cn } from "@/lib/utils";
import { FOCUS_RING } from "@/components/ui/styles";

/**
 * Inline "new board" form: expands in place, submits on Enter, cancels on
 * Escape, and navigates to the board it just created.
 */
export function NewBoardForm({
  onCreated,
  onCancel,
  autoFocus,
  className,
}: {
  onCreated?: () => void;
  onCancel?: () => void;
  autoFocus?: boolean;
  className?: string;
}) {
  const router = useRouter();
  const { createBoard } = useBoards();
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (autoFocus) inputRef.current?.focus();
  }, [autoFocus]);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    const trimmed = name.trim();
    // The contract requires a non-empty name; catch it before spending a call.
    if (!trimmed || busy) return;
    setBusy(true);
    try {
      const board = await createBoard(trimmed);
      setName("");
      onCreated?.();
      router.push(`/board/${board.id}`);
    } catch {
      // createBoard already surfaced a toast; keep the form open to retry.
    } finally {
      setBusy(false);
    }
  };

  return (
    <form onSubmit={submit} className={cn("space-y-2", className)}>
      <input
        ref={inputRef}
        value={name}
        onChange={(event) => setName(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Escape") {
            event.preventDefault();
            onCancel?.();
          }
        }}
        placeholder="Board name"
        aria-label="Board name"
        maxLength={80}
        className={cn(
          "w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground/70",
          FOCUS_RING,
        )}
      />
      <div className="flex items-center gap-2">
        <Button type="submit" variant="primary" size="sm" disabled={busy || !name.trim()}>
          {busy ? "Creating…" : "Create board"}
        </Button>
        {onCancel ? (
          <Button type="button" variant="ghost" size="sm" onClick={onCancel} disabled={busy}>
            Cancel
          </Button>
        ) : null}
      </div>
    </form>
  );
}

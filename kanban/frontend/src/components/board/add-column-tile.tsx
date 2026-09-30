"use client";

import { Plus, X } from "lucide-react";
import { useEffect, useRef, useState, type FormEvent } from "react";

import { Button, IconButton } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { FOCUS_RING, INPUT_CLASSES } from "@/components/ui/styles";

/** The tile at the end of the column row that creates a custom deal stage. */
export function AddColumnTile({
  onCreate,
}: {
  onCreate: (name: string) => Promise<boolean>;
}) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) inputRef.current?.focus();
  }, [open]);

  const close = () => {
    setOpen(false);
    setName("");
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    const trimmed = name.trim();
    if (!trimmed || saving) return;
    setSaving(true);
    const ok = await onCreate(trimmed);
    setSaving(false);
    if (ok) setName("");
  };

  if (open) {
    return (
      <div className="w-72 shrink-0 rounded-2xl border border-border bg-surface-muted p-2.5">
        <form onSubmit={submit} className="space-y-2">
          <input
            ref={inputRef}
            value={name}
            onChange={(event) => setName(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Escape") {
                event.preventDefault();
                close();
              }
            }}
            placeholder="Stage name"
            aria-label="Column name"
            maxLength={60}
            className={cn(INPUT_CLASSES, FOCUS_RING, "h-8 py-1 text-sm")}
          />
          <div className="flex items-center gap-1.5">
            <Button
              type="submit"
              variant="primary"
              size="sm"
              disabled={saving || !name.trim()}
            >
              {saving ? "Adding…" : "Add column"}
            </Button>
            <IconButton label="Cancel adding a column" size="icon-sm" onClick={close}>
              <X size={14} />
            </IconButton>
          </div>
        </form>
      </div>
    );
  }

  return (
    <button
      type="button"
      onClick={() => setOpen(true)}
      className={cn(
        "flex h-14 w-72 shrink-0 items-center gap-2 rounded-2xl border border-dashed border-border-strong px-4 text-sm font-medium text-muted-foreground transition-colors hover:border-accent hover:bg-accent-soft hover:text-accent",
        FOCUS_RING,
      )}
    >
      <Plus size={16} aria-hidden="true" />
      Add column
    </button>
  );
}

"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { Columns3, Trash, X } from "lucide-react";
import { useMemo, useState, type FormEvent } from "react";

import { Avatar } from "@/components/card/assignee-avatar";
import { PriorityBadge } from "@/components/card/priority-badge";
import { Button, IconButton } from "@/components/ui/button";
import { SelectField, TextAreaField, TextField } from "@/components/ui/inputs";
import { PRIORITIES } from "@/lib/constants";
import { formatTimestamp } from "@/lib/dates";
import { cn } from "@/lib/utils";
import type { Card, CreateCardRequest, Priority } from "@/lib/types";
import { FOCUS_RING } from "@/components/ui/styles";

const PRIORITY_LABELS: Record<Priority, string> = {
  low: "Low",
  medium: "Medium",
  high: "High",
  urgent: "Urgent",
};

interface FormState {
  title: string;
  description: string;
  priority: Priority;
  dueDate: string;
  assignee: string;
}

type Pending = "none" | "discard" | "delete";

export function CardDrawer({
  card,
  columnName,
  knownAssignees,
  onClose,
  onSave,
  onDelete,
}: {
  card: Card;
  columnName: string;
  knownAssignees: string[];
  onClose: () => void;
  onSave: (input: CreateCardRequest) => Promise<boolean>;
  onDelete: () => Promise<boolean>;
}) {
  const [form, setForm] = useState<FormState>({
    title: card.title,
    description: card.description,
    priority: card.priority,
    dueDate: card.due_date ?? "",
    assignee: card.assignee,
  });
  const [titleError, setTitleError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [pending, setPending] = useState<Pending>("none");

  // Assignee is free text; the datalist only offers names already on the board.
  const datalistId = `assignees-${card.id}`;
  const assignees = useMemo(
    () => [...knownAssignees].sort((a, b) => a.localeCompare(b)),
    [knownAssignees],
  );

  const isDirty =
    form.title !== card.title ||
    form.description !== card.description ||
    form.assignee !== card.assignee ||
    form.priority !== card.priority ||
    form.dueDate !== (card.due_date ?? "");

  const submit = async (event?: FormEvent) => {
    event?.preventDefault();
    const title = form.title.trim();
    if (!title) {
      setTitleError("A title is required.");
      return;
    }
    setTitleError(null);
    setSaving(true);
    // `dueDate` is passed through as the raw `YYYY-MM-DD` the input produced —
    // it is never round-tripped through a Date, which would shift the day for
    // negative-offset timezones.
    const ok = await onSave({
      title,
      description: form.description,
      assignee: form.assignee.trim(),
      priority: form.priority,
      due_date: form.dueDate || null,
    });
    setSaving(false);
    if (ok) onClose();
  };

  /**
   * Backdrop clicks and Escape both land here. Rather than silently dropping
   * edits, a dirty form is intercepted and an explicit choice is offered.
   */
  const requestClose = () => {
    if (isDirty && !saving) {
      setPending("discard");
      return;
    }
    onClose();
  };

  return (
    <Dialog.Root open onOpenChange={(open) => !open && requestClose()}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-black/45 backdrop-blur-[1px] data-[state=open]:animate-overlay-in data-[state=closed]:animate-overlay-out" />
        <Dialog.Content
          className={cn(
            "fixed inset-y-0 right-0 z-50 flex w-[min(30rem,100vw)] flex-col border-l border-border bg-surface shadow-2xl",
            "data-[state=open]:animate-drawer-in data-[state=closed]:animate-drawer-out",
          )}
        >
          <header className="flex shrink-0 items-center gap-3 border-b border-border px-5 py-3.5">
            <div className="min-w-0 flex-1">
              <Dialog.Title className="text-sm font-semibold text-foreground">
                Card details
              </Dialog.Title>
              <Dialog.Description className="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
                <Columns3 size={12} aria-hidden="true" />
                In “{columnName}”
              </Dialog.Description>
            </div>
            <PriorityBadge priority={card.priority} size="md" />
            <IconButton label="Close card details" onClick={requestClose}>
              <X size={16} />
            </IconButton>
          </header>

          <form
            onSubmit={submit}
            className="flex min-h-0 flex-1 flex-col"
            // Keep Enter in the title field from submitting a half-typed card
            // while the user is still tabbing through the form.
            onKeyDown={(event) => {
              if (event.key === "Enter" && event.target instanceof HTMLInputElement && event.target.type === "text") {
                event.preventDefault();
              }
            }}
          >
            <div className="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
              <TextField
                id="card-title"
                label="Title"
                value={form.title}
                error={titleError ?? undefined}
                onChange={(event) => {
                  setForm((current) => ({ ...current, title: event.target.value }));
                  if (titleError) setTitleError(null);
                }}
                placeholder="Short, action-oriented summary"
                maxLength={200}
                autoComplete="off"
              />

              <TextAreaField
                id="card-description"
                label="Description"
                value={form.description}
                onChange={(event) =>
                  setForm((current) => ({ ...current, description: event.target.value }))
                }
                placeholder="Context, acceptance criteria, links…"
                rows={5}
              />

              <div className="grid gap-4 sm:grid-cols-2">
                <SelectField
                  id="card-priority"
                  label="Priority"
                  value={form.priority}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      priority: event.target.value as Priority,
                    }))
                  }
                >
                  {PRIORITIES.map((priority) => (
                    <option key={priority} value={priority}>
                      {PRIORITY_LABELS[priority]}
                    </option>
                  ))}
                </SelectField>

                <TextField
                  id="card-due-date"
                  label="Due date"
                  type="date"
                  value={form.dueDate}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, dueDate: event.target.value }))
                  }
                />
              </div>

              <TextField
                id="card-assignee"
                label="Assignee"
                list={datalistId}
                value={form.assignee}
                onChange={(event) =>
                  setForm((current) => ({ ...current, assignee: event.target.value }))
                }
                placeholder="Who owns this?"
                autoComplete="off"
                hint="Free text — suggestions come from people already on this board."
              />
              <datalist id={datalistId}>
                {assignees.map((name) => (
                  <option key={name} value={name} />
                ))}
              </datalist>

              <div className="flex items-center gap-2 border-t border-border pt-4">
                <Avatar name={card.assignee} size={20} />
                <p className="text-xs text-muted-foreground">
                  {card.assignee ? `Assigned to ${card.assignee}` : "Unassigned"} · updated{" "}
                  {formatTimestamp(card.updated_at) || "just now"}
                </p>
              </div>
            </div>

            {pending !== "none" ? (
              <div className="shrink-0 border-t border-border bg-surface-strong px-5 py-3">
                <p className="text-sm font-medium text-foreground">
                  {pending === "delete" ? "Delete this card?" : "Discard unsaved changes?"}
                </p>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {pending === "delete"
                    ? "The card and everything on it will be removed permanently."
                    : "Your edits to this card will be lost."}
                </p>
                <div className="mt-2.5 flex justify-end gap-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => setPending("none")}
                    autoFocus
                  >
                    {pending === "delete" ? "Keep card" : "Keep editing"}
                  </Button>
                  <Button
                    variant={pending === "delete" ? "danger" : "secondary"}
                    size="sm"
                    onClick={async () => {
                      if (pending === "delete") {
                        const ok = await onDelete();
                        if (ok) onClose();
                        else setPending("none");
                      } else {
                        setPending("none");
                        onClose();
                      }
                    }}
                  >
                    {pending === "delete" ? "Delete" : "Discard"}
                  </Button>
                </div>
              </div>
            ) : null}

            <footer className="flex shrink-0 items-center gap-2 border-t border-border px-5 py-3">
              <Button
                variant="ghost"
                size="sm"
                className="mr-auto text-danger hover:bg-danger-soft"
                onClick={() => setPending("delete")}
                disabled={saving}
              >
                <Trash size={14} aria-hidden="true" />
                Delete
              </Button>
              <Button
                variant="secondary"
                onClick={requestClose}
                disabled={saving}
                className={cn(FOCUS_RING)}
              >
                Cancel
              </Button>
              <Button type="submit" variant="primary" disabled={saving || !isDirty}>
                {saving ? "Saving…" : "Save changes"}
              </Button>
            </footer>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

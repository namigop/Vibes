"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { TriangleAlert, X } from "lucide-react";
import { useState, type ReactNode } from "react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { FOCUS_RING } from "@/components/ui/styles";

export interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  /** Red confirm button + warning icon. */
  destructive?: boolean;
  onConfirm: () => void | Promise<void>;
}

/**
 * Destructive actions (deleting a board cascades to every column and card)
 * always go through here so the blast radius is spelled out before anything
 * is sent, and the request is not fired twice.
 */
export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  destructive,
  onConfirm,
}: ConfirmDialogProps) {
  const [busy, setBusy] = useState(false);

  const handleConfirm = async () => {
    setBusy(true);
    try {
      await onConfirm();
      onOpenChange(false);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog.Root open={open} onOpenChange={busy ? () => undefined : onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-[60] bg-black/45 backdrop-blur-[1px] data-[state=open]:animate-overlay-in" />
        <Dialog.Content
          className={cn(
            "fixed left-1/2 top-1/2 z-[60] w-[min(26rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 rounded-2xl border border-border bg-surface-strong p-5 shadow-2xl shadow-black/20",
            "data-[state=open]:animate-scale-in",
          )}
          onOpenAutoFocus={(event) => {
            // Land focus on Cancel so Enter cannot destroy anything.
            event.preventDefault();
          }}
        >
          <div className="flex items-start gap-3">
            {destructive ? (
              <span className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-danger-soft text-danger">
                <TriangleAlert size={16} />
              </span>
            ) : null}
            <div className="min-w-0 flex-1">
              <Dialog.Title className="text-sm font-semibold text-foreground">
                {title}
              </Dialog.Title>
              <Dialog.Description asChild>
                <div className="mt-1.5 space-y-2 text-sm leading-relaxed text-muted-foreground">
                  {description}
                </div>
              </Dialog.Description>
            </div>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close dialog"
                className={cn(
                  "-m-1 rounded-md p-1 text-muted-foreground transition-colors hover:bg-surface-muted hover:text-foreground",
                  FOCUS_RING,
                )}
              >
                <X size={16} />
              </button>
            </Dialog.Close>
          </div>

          <div className="mt-5 flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => onOpenChange(false)}
              disabled={busy}
              autoFocus
            >
              {cancelLabel}
            </Button>
            <Button
              variant={destructive ? "danger" : "primary"}
              onClick={handleConfirm}
              disabled={busy}
            >
              {busy ? "Working…" : confirmLabel}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

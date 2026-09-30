"use client";

import { TriangleAlert } from "lucide-react";

import { Button } from "@/components/ui/button";

export function BoardError({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="flex h-full items-center justify-center p-6">
      <div className="w-full max-w-md rounded-2xl border border-border bg-surface p-6 text-center">
        <span className="mx-auto flex size-10 items-center justify-center rounded-full bg-danger-soft text-danger">
          <TriangleAlert size={18} />
        </span>
        <h1 className="mt-3 text-base font-semibold text-foreground">
          This board could not be loaded
        </h1>
        <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">{message}</p>
        <div className="mt-4 flex justify-center gap-2">
          <Button variant="secondary" onClick={onRetry}>
            Retry
          </Button>
        </div>
      </div>
    </div>
  );
}

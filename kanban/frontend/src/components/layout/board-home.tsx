"use client";

import { Columns3 } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { BoardSkeleton } from "@/components/board/board-skeleton";
import { NewBoardForm } from "@/components/layout/new-board-form";
import { Button } from "@/components/ui/button";
import { useBoards } from "@/providers/boards-provider";

/**
 * `/` is a doorway, not a page: it sends you to the first board, or helps you
 * create one when the workspace is empty.
 */
export function BoardHome() {
  const router = useRouter();
  const { boards, status, error, reload } = useBoards();
  const first = boards[0];

  useEffect(() => {
    if (status === "ready" && first) router.replace(`/board/${first.id}`);
  }, [status, first, router]);

  if (status === "error") {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <div className="w-full max-w-md rounded-2xl border border-border bg-surface p-6 text-center">
          <h1 className="text-base font-semibold text-foreground">
            The board list could not be loaded
          </h1>
          <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">{error}</p>
          <Button variant="secondary" className="mt-4" onClick={reload}>
            Retry
          </Button>
        </div>
      </div>
    );
  }

  // Still loading, or about to be redirected to the first board.
  if (status === "loading" || first) return <BoardSkeleton />;

  return (
    <div className="flex h-full items-center justify-center p-6">
      <div className="w-full max-w-md rounded-2xl border border-dashed border-border-strong bg-surface-muted/60 p-8 text-center">
        <span className="mx-auto flex size-11 items-center justify-center rounded-xl bg-accent-soft text-accent">
          <Columns3 size={20} />
        </span>
        <h1 className="mt-4 text-lg font-semibold tracking-tight text-foreground">
          Create your first board
        </h1>
        <p className="mx-auto mt-1.5 max-w-sm text-sm leading-relaxed text-muted-foreground">
          A board holds the columns and cards for one workflow. Start with a pipeline,
          a sprint, or a set of deals.
        </p>
        <NewBoardForm className="mx-auto mt-5 max-w-xs text-left" autoFocus />
      </div>
    </div>
  );
}

"use client";

import { Columns3, Ellipsis, Kanban, Plus, Trash, TriangleAlert, X } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { NewBoardForm } from "@/components/layout/new-board-form";
import { Button, IconButton } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import {
  Menu,
  MenuContent,
  MenuItem,
  MenuSeparator,
  MenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/styles";
import { cn } from "@/lib/utils";
import type { Board } from "@/lib/types";
import { useActiveBoardId, useBoards } from "@/providers/boards-provider";

export function SideNav({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { boards, status, error, reload, deleteBoard } = useBoards();
  const activeId = useActiveBoardId();
  const router = useRouter();
  const [creating, setCreating] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<Board | null>(null);

  return (
    <>
      {/* Scrim, mobile only. */}
      <div
        onClick={onClose}
        className={cn(
          "fixed inset-0 z-40 bg-black/45 backdrop-blur-[1px] transition-opacity md:hidden",
          open ? "opacity-100" : "pointer-events-none opacity-0",
        )}
        aria-hidden="true"
      />
      <aside
        aria-label="Boards"
        className={cn(
          "fixed inset-y-0 left-0 z-50 flex w-72 shrink-0 flex-col border-r border-border bg-surface-muted transition-transform duration-200 md:static md:z-auto md:translate-x-0",
          open ? "translate-x-0 shadow-2xl shadow-black/20" : "-translate-x-full",
        )}
      >
        <div className="flex h-14 shrink-0 items-center gap-2.5 border-b border-border px-4">
          <span className="flex size-8 items-center justify-center rounded-lg bg-accent text-accent-foreground">
            <Kanban size={16} />
          </span>
          <span className="flex-1 truncate text-sm font-semibold tracking-tight">
            Kanban
          </span>
          <IconButton
            label="Close navigation"
            className="md:hidden"
            onClick={onClose}
          >
            <X size={16} />
          </IconButton>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-3">
          <div className="mb-2 flex items-center justify-between px-1">
            <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Boards
            </h2>
            {!creating && status === "ready" ? (
              <button
                type="button"
                onClick={() => setCreating(true)}
                className="inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-surface hover:text-foreground"
              >
                <Plus size={13} />
                New board
              </button>
            ) : null}
          </div>

          {creating ? (
            <NewBoardForm
              autoFocus
              className="mb-2 rounded-xl border border-border bg-surface p-2"
              onCreated={() => setCreating(false)}
              onCancel={() => setCreating(false)}
            />
          ) : null}

          {status === "loading" && !creating ? <BoardListSkeleton /> : null}

          {status === "error" ? (
            <div className="rounded-xl border border-dashed border-border bg-surface p-3 text-center">
              <TriangleAlert size={16} className="mx-auto text-warning" />
              <p className="mt-1.5 text-xs text-muted-foreground">{error}</p>
              <Button size="sm" variant="subtle" className="mt-2" onClick={reload}>
                Retry
              </Button>
            </div>
          ) : null}

          {status === "ready" && boards.length === 0 && !creating ? (
            <div className="rounded-xl border border-dashed border-border bg-surface px-3 py-6 text-center">
              <Columns3 size={18} className="mx-auto text-muted-foreground" />
              <p className="mt-2 text-sm font-medium">No boards yet</p>
              <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
                Boards hold the columns and cards for one workflow — a pipeline,
                a sprint, a set of deals.
              </p>
              <Button
                size="sm"
                variant="primary"
                className="mt-3"
                onClick={() => setCreating(true)}
              >
                <Plus size={14} />
                New board
              </Button>
            </div>
          ) : null}

          {boards.length > 0 ? (
            <ul className="space-y-0.5">
              {boards.map((board) => {
                const isActive = board.id === activeId;
                return (
                  <li key={board.id} className="group relative flex items-center">
                    <Link
                      href={`/board/${board.id}`}
                      onClick={onClose}
                      aria-current={isActive ? "page" : undefined}
                      className={cn(
                        "min-w-0 flex-1 truncate rounded-lg py-2 pl-3 pr-8 text-sm transition-colors",
                        isActive
                          ? "bg-accent-soft font-medium text-accent"
                          : "text-muted-foreground hover:bg-surface hover:text-foreground",
                      )}
                    >
                      {board.name}
                    </Link>
                    <div
                      className={cn(
                        "absolute right-1 transition-opacity",
                        isActive ? "opacity-100" : "opacity-0 group-hover:opacity-100 focus-within:opacity-100",
                      )}
                    >
                      <Menu>
                        <MenuTrigger label={`Actions for ${board.name}`} size="icon-sm">
                          <Ellipsis size={14} />
                        </MenuTrigger>
                        <MenuContent>
                          <MenuItem
                            onSelect={() => {
                              onClose();
                              router.push(`/board/${board.id}`);
                            }}
                          >
                            Open
                          </MenuItem>
                          <MenuSeparator />
                          <MenuItem
                            destructive
                            icon={<Trash size={14} />}
                            onSelect={() => setPendingDelete(board)}
                          >
                            Delete board
                          </MenuItem>
                        </MenuContent>
                      </Menu>
                    </div>
                  </li>
                );
              })}
            </ul>
          ) : null}
        </div>
      </aside>

      <ConfirmDialog
        open={pendingDelete !== null}
        onOpenChange={(next) => {
          if (!next) setPendingDelete(null);
        }}
        destructive
        title={`Delete “${pendingDelete?.name ?? ""}”?`}
        confirmLabel="Delete board"
        description={
          <>
            <p>
              This permanently deletes the board, every column in it and every card
              in those columns.
            </p>
            <p className="font-medium text-danger">This cannot be undone.</p>
          </>
        }
        onConfirm={async () => {
          if (!pendingDelete) return;
          const deletingActive = pendingDelete.id === activeId;
          try {
            await deleteBoard(pendingDelete.id);
            setPendingDelete(null);
            if (deletingActive) {
              onClose();
              router.push("/");
            }
          } catch {
            // deleteBoard rolled the list back and toasted the reason.
            setPendingDelete(null);
          }
        }}
      />
    </>
  );
}

function BoardListSkeleton() {
  return (
    <ul className="space-y-2" aria-hidden="true">
      {[0, 1, 2, 3].map((index) => (
        <li key={index}>
          <Skeleton className="h-8 w-full" />
        </li>
      ))}
    </ul>
  );
}

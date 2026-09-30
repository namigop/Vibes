"use client";

import { usePathname } from "next/navigation";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import * as api from "@/lib/api";
import { sortByPosition } from "@/lib/board-state";
import type { Board } from "@/lib/types";
import { useToast } from "@/providers/toast-provider";

type Status = "loading" | "ready" | "error";

interface BoardsContextValue {
  boards: Board[];
  status: Status;
  error: string | null;
  reload: () => void;
  /** Resolves with the created board so the caller can navigate to it. */
  createBoard: (name: string) => Promise<Board>;
  deleteBoard: (id: string) => Promise<void>;
  /** Keeps the nav in sync when a board is renamed from inside the board view. */
  applyBoard: (board: Board) => void;
}

const BoardsContext = createContext<BoardsContextValue | null>(null);

/** Shared by the side nav (which lists boards) and the home page. */
export function BoardsProvider({ children }: { children: ReactNode }) {
  const [boards, setBoards] = useState<Board[]>([]);
  const [status, setStatus] = useState<Status>("loading");
  const [error, setError] = useState<string | null>(null);
  const toast = useToast();

  // The effect only starts the request; setState lives in the callbacks.
  useEffect(() => {
    let ignore = false;
    api.listBoards().then(
      (result) => {
        if (ignore) return;
        setBoards(sortByPosition(result ?? []));
        setStatus("ready");
      },
      (cause: unknown) => {
        if (ignore) return;
        setError(api.describeError(cause));
        setStatus("error");
      },
    );
    return () => {
      ignore = true;
    };
  }, []);

  const reload = useCallback(() => {
    setStatus("loading");
    setError(null);
    api.listBoards().then(
      (result) => {
        setBoards(sortByPosition(result ?? []));
        setStatus("ready");
      },
      (cause: unknown) => {
        setError(api.describeError(cause));
        setStatus("error");
      },
    );
  }, []);

  const createBoard = useCallback(
    async (name: string) => {
      try {
        const board = await api.createBoard(name);
        setBoards((current) => [...current, board]);
        return board;
      } catch (cause) {
        toast.error("Could not create the board", api.describeError(cause));
        throw cause;
      }
    },
    [toast],
  );

  const deleteBoard = useCallback(
    async (id: string) => {
      const snapshot = boards;
      setBoards((current) => current.filter((board) => board.id !== id));
      try {
        await api.deleteBoard(id);
        toast.success("Board deleted");
      } catch (cause) {
        setBoards(snapshot);
        toast.error("Could not delete the board", api.describeError(cause));
        throw cause;
      }
    },
    [boards, toast],
  );

  const applyBoard = useCallback((board: Board) => {
    setBoards((current) => current.map((item) => (item.id === board.id ? board : item)));
  }, []);

  const value = useMemo<BoardsContextValue>(
    () => ({ boards, status, error, reload, createBoard, deleteBoard, applyBoard }),
    [boards, status, error, reload, createBoard, deleteBoard, applyBoard],
  );

  return <BoardsContext.Provider value={value}>{children}</BoardsContext.Provider>;
}

export function useBoards(): BoardsContextValue {
  const context = useContext(BoardsContext);
  if (!context) throw new Error("useBoards must be used inside <BoardsProvider>");
  return context;
}

/** The board id encoded in `/board/{id}`, or null anywhere else. */
export function useActiveBoardId(): string | null {
  const pathname = usePathname();
  const match = /^\/board\/([^/]+)$/.exec(pathname ?? "");
  return match?.[1] ?? null;
}

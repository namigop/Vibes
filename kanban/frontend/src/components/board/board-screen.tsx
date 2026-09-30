"use client";

import { BoardError } from "@/components/board/board-error";
import { BoardSkeleton } from "@/components/board/board-skeleton";
import { BoardView } from "@/components/board/board-view";
import { useBoard } from "@/hooks/use-board";

export function BoardScreen({ boardId }: { boardId: string }) {
  const board = useBoard(boardId);

  // Keep showing the last good board if a background refresh fails.
  if (board.status === "error" && !board.state) {
    return <BoardError message={board.error ?? "Unknown error"} onRetry={board.reload} />;
  }
  if (!board.state) return <BoardSkeleton />;

  return (
    <BoardView
      state={board.state}
      getState={board.getState}
      renameBoard={(name) => void board.renameBoard(name)}
      createColumn={board.createColumn}
      renameColumn={(columnId, name) => void board.renameColumn(columnId, name)}
      deleteColumn={(columnId) => void board.deleteColumn(columnId)}
      moveColumn={(columnId, toIndex, snapshot) =>
        void board.moveColumn(columnId, toIndex, snapshot)
      }
      createCard={board.createCard}
      updateCard={board.updateCard}
      deleteCard={board.deleteCard}
      moveCard={(cardId, toColumnId, toIndex, snapshot) =>
        void board.moveCard(cardId, toColumnId, toIndex, snapshot)
      }
      previewCardMove={board.previewCardMove}
      restoreSnapshot={board.restoreSnapshot}
    />
  );
}

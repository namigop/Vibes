import { BoardScreen } from "@/components/board/board-screen";

/**
 * The active board lives in the URL so a reload and the browser Back button
 * both land on the same board. The route stays a server component; only the
 * screen beneath it is a client component.
 */
export default async function Page({ params }: { params: Promise<{ boardId: string }> }) {
  const { boardId } = await params;
  // Keyed so switching boards remounts the screen with fresh state instead of
  // briefly rendering the previous board's columns.
  return <BoardScreen key={boardId} boardId={boardId} />;
}

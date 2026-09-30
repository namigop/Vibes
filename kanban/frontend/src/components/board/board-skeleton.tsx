import { Skeleton } from "@/components/ui/styles";

export function BoardSkeleton() {
  return (
    <div className="flex h-full min-h-0 flex-col" aria-busy="true">
      <span className="sr-only" role="status">
        Loading board…
      </span>
      <div className="flex shrink-0 items-center gap-3 border-b border-border px-4 py-3 sm:px-6">
        <Skeleton className="h-7 w-56" />
        <Skeleton className="h-5 w-32" />
      </div>
      <div className="flex min-h-0 flex-1 gap-4 overflow-hidden p-4 sm:px-6">
        {[0, 1, 2].map((column) => (
          <div
            key={column}
            className="flex h-full w-72 shrink-0 flex-col gap-2 rounded-2xl border border-border bg-surface-muted p-2.5"
          >
            <Skeleton className="h-6 w-32" />
            <Skeleton className="h-20 w-full rounded-xl" />
            <Skeleton className="h-20 w-full rounded-xl" />
            <Skeleton className="h-4 w-20" />
          </div>
        ))}
      </div>
    </div>
  );
}

"use client";

import { CircleCheck, Info, TriangleAlert, X } from "lucide-react";
import type { CSSProperties } from "react";

import { cn } from "@/lib/utils";
import type { ToastVariant } from "@/providers/toast-provider";

const ICONS: Record<ToastVariant, typeof Info> = {
  success: CircleCheck,
  error: TriangleAlert,
  info: Info,
};

const ACCENTS: Record<ToastVariant, string> = {
  success: "text-success",
  error: "text-danger",
  info: "text-accent",
};

export interface ToasterItem {
  id: string;
  title: string;
  description?: string;
  variant: ToastVariant;
  duration: number;
}

export function Toaster({
  items,
  onDismiss,
}: {
  items: ToasterItem[];
  onDismiss: (id: string) => void;
}) {
  return (
    <div
      className="pointer-events-none fixed inset-x-0 bottom-0 z-[70] flex flex-col items-center gap-2 p-4 sm:inset-x-auto sm:right-0 sm:items-end"
      // Errors interrupt, everything else is announced politely.
      aria-live="polite"
    >
      {items.map((item) => {
        const Icon = ICONS[item.variant];
        return (
          <div
            key={item.id}
            role={item.variant === "error" ? "alert" : "status"}
            className="animate-toast-in pointer-events-auto relative w-full max-w-sm overflow-hidden rounded-xl border border-border bg-surface-strong p-3 pr-9 shadow-lg shadow-black/10"
            style={
              item.duration > 0
                ? ({ "--toast-duration": `${item.duration}ms` } as CSSProperties)
                : undefined
            }
          >
            <div className="flex items-start gap-2.5">
              <Icon size={16} className={cn("mt-0.5 shrink-0", ACCENTS[item.variant])} />
              <div className="min-w-0">
                <p className="text-sm font-medium text-foreground">{item.title}</p>
                {item.description ? (
                  <p className="mt-0.5 text-xs leading-relaxed text-muted-foreground">
                    {item.description}
                  </p>
                ) : null}
              </div>
            </div>
            <button
              type="button"
              onClick={() => onDismiss(item.id)}
              aria-label="Dismiss notification"
              className="absolute right-2 top-2 rounded-md p-1 text-muted-foreground transition-colors hover:bg-surface-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <X size={14} />
            </button>
            {item.duration > 0 ? (
              <div className="animate-toast-bar absolute inset-x-0 bottom-0 h-0.5 origin-left bg-border-strong" />
            ) : null}
          </div>
        );
      })}
    </div>
  );
}

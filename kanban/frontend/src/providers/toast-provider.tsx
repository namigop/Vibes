"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";

import { Toaster } from "@/components/ui/toaster";

export type ToastVariant = "success" | "error" | "info";

export interface ToastInput {
  title: string;
  description?: string;
  variant?: ToastVariant;
  /** ms before auto-dismiss; 0 keeps it until dismissed. */
  duration?: number;
}

interface ToastItem extends Required<Omit<ToastInput, "description">> {
  id: string;
  description?: string;
}

interface ToastContextValue {
  toast: (input: ToastInput) => string;
  success: (title: string, description?: string) => string;
  error: (title: string, description?: string) => string;
  info: (title: string, description?: string) => string;
  dismiss: (id: string) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

const DEFAULT_DURATION: Record<ToastVariant, number> = {
  success: 3000,
  info: 4000,
  // Failures stay up long enough to actually be read.
  error: 7000,
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([]);
  const nextId = useRef(0);
  const timers = useRef(new Map<string, ReturnType<typeof setTimeout>>());

  const dismiss = useCallback((id: string) => {
    setItems((current) => current.filter((item) => item.id !== id));
    const timer = timers.current.get(id);
    if (timer) {
      clearTimeout(timer);
      timers.current.delete(id);
    }
  }, []);

  const toast = useCallback(
    (input: ToastInput) => {
      nextId.current += 1;
      const id = `toast-${nextId.current}`;
      const variant = input.variant ?? "info";
      const item: ToastItem = {
        id,
        title: input.title,
        description: input.description,
        variant,
        duration: input.duration ?? DEFAULT_DURATION[variant],
      };

      // Never stack more than four; drop the oldest.
      setItems((current) => [...current.slice(-3), item]);

      if (item.duration > 0) {
        timers.current.set(
          id,
          setTimeout(() => dismiss(id), item.duration),
        );
      }
      return id;
    },
    [dismiss],
  );

  useEffect(() => {
    const pending = timers.current;
    return () => {
      pending.forEach((timer) => clearTimeout(timer));
      pending.clear();
    };
  }, []);

  const value = useMemo<ToastContextValue>(
    () => ({
      toast,
      dismiss,
      success: (title, description) => toast({ title, description, variant: "success" }),
      error: (title, description) => toast({ title, description, variant: "error" }),
      info: (title, description) => toast({ title, description, variant: "info" }),
    }),
    [toast, dismiss],
  );

  return (
    <ToastContext.Provider value={value}>
      {children}
      <Toaster items={items} onDismiss={dismiss} />
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const context = useContext(ToastContext);
  if (!context) throw new Error("useToast must be used inside <ToastProvider>");
  return context;
}

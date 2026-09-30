"use client";

import * as RadixMenu from "@radix-ui/react-dropdown-menu";
import type { ComponentPropsWithoutRef, ReactNode } from "react";

import { cn } from "@/lib/utils";
import { FOCUS_RING } from "@/components/ui/styles";

export function Menu({ children }: { children: ReactNode }) {
  return <RadixMenu.Root>{children}</RadixMenu.Root>;
}

export interface MenuTriggerProps
  extends ComponentPropsWithoutRef<typeof RadixMenu.Trigger> {
  label: string;
  /** `icon` renders a bare 32px button; `icon-sm` a 28px one. */
  size?: "icon" | "icon-sm";
}

export function MenuTrigger({
  label,
  size = "icon",
  className,
  ...props
}: MenuTriggerProps) {
  return (
    <RadixMenu.Trigger asChild>
      <button
        type="button"
        aria-label={label}
        title={label}
        className={cn(
          "inline-flex shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-surface-muted hover:text-foreground disabled:pointer-events-none disabled:opacity-50 data-[state=open]:bg-surface-muted data-[state=open]:text-foreground",
          FOCUS_RING,
          size === "icon" ? "size-8" : "size-7",
          className,
        )}
        {...props}
      />
    </RadixMenu.Trigger>
  );
}

export function MenuContent({
  children,
  align = "end",
  className,
}: {
  children: ReactNode;
  align?: "start" | "center" | "end";
  className?: string;
}) {
  return (
    <RadixMenu.Portal>
      <RadixMenu.Content
        align={align}
        sideOffset={6}
        collisionPadding={8}
        className={cn(
          "z-[60] min-w-44 rounded-xl border border-border bg-surface-strong p-1 shadow-xl shadow-black/10",
          "data-[state=open]:animate-scale-in",
          className,
        )}
      >
        {children}
      </RadixMenu.Content>
    </RadixMenu.Portal>
  );
}

export interface MenuItemProps extends ComponentPropsWithoutRef<typeof RadixMenu.Item> {
  icon?: ReactNode;
  destructive?: boolean;
  shortcut?: string;
}

export function MenuItem({
  icon,
  destructive,
  shortcut,
  className,
  children,
  ...props
}: MenuItemProps) {
  return (
    <RadixMenu.Item
      className={cn(
        "flex cursor-pointer select-none items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm text-foreground outline-none transition-colors data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
        destructive
          ? "text-danger data-[highlighted]:bg-danger-soft"
          : "data-[highlighted]:bg-surface-muted",
        className,
      )}
      {...props}
    >
      {icon ? <span className="shrink-0 text-muted-foreground">{icon}</span> : null}
      <span className="flex-1">{children}</span>
      {shortcut ? (
        <span className="text-xs text-muted-foreground">{shortcut}</span>
      ) : null}
    </RadixMenu.Item>
  );
}

export function MenuSeparator() {
  return <RadixMenu.Separator className="my-1 h-px bg-border" />;
}

"use client";

import type { ButtonHTMLAttributes, ReactNode, Ref } from "react";

import { cn } from "@/lib/utils";
import { FOCUS_RING } from "@/components/ui/styles";

type Variant = "primary" | "secondary" | "ghost" | "danger" | "subtle";
type Size = "sm" | "md" | "icon" | "icon-sm";

const VARIANTS: Record<Variant, string> = {
  primary:
    "bg-accent text-accent-foreground hover:bg-accent-hover shadow-sm shadow-accent/20",
  secondary:
    "border border-border bg-surface text-foreground hover:bg-surface-muted hover:border-border-strong",
  ghost: "text-muted-foreground hover:bg-surface-muted hover:text-foreground",
  subtle: "bg-surface-muted text-foreground hover:bg-surface-strong",
  danger: "bg-danger text-white hover:opacity-90 shadow-sm",
};

const SIZES: Record<Size, string> = {
  sm: "h-8 gap-1.5 px-2.5 text-xs",
  md: "h-9 gap-2 px-3.5 text-sm",
  icon: "size-9",
  "icon-sm": "size-7",
};

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  children?: ReactNode;
  ref?: Ref<HTMLButtonElement>;
}

export function Button({
  variant = "secondary",
  size = "md",
  className,
  type = "button",
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        "inline-flex shrink-0 items-center justify-center rounded-lg font-medium transition-colors disabled:pointer-events-none disabled:opacity-50",
        FOCUS_RING,
        VARIANTS[variant],
        SIZES[size],
        className,
      )}
      {...props}
    />
  );
}

export interface IconButtonProps extends ButtonProps {
  /** Required: icon-only controls need an accessible name. */
  label: string;
}

export function IconButton({ label, className, size = "icon", ...props }: IconButtonProps) {
  return (
    <Button
      size={size}
      variant="ghost"
      aria-label={label}
      title={label}
      className={cn("rounded-lg", className)}
      {...props}
    />
  );
}

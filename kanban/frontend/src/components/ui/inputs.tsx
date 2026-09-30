"use client";

import type {
  InputHTMLAttributes,
  LabelHTMLAttributes,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
  ReactNode,
  Ref,
} from "react";

import { cn } from "@/lib/utils";
import { FOCUS_RING, INPUT_CLASSES } from "@/components/ui/styles";

export function FieldLabel({
  className,
  ...props
}: LabelHTMLAttributes<HTMLLabelElement>) {
  return (
    <label
      className={cn(
        "text-xs font-medium uppercase tracking-wide text-muted-foreground",
        className,
      )}
      {...props}
    />
  );
}

function FieldShell({
  label,
  htmlFor,
  hint,
  error,
  children,
}: {
  label: string;
  htmlFor: string;
  hint?: string;
  error?: string;
  children: ReactNode;
}) {
  const hintId = hint ? `${htmlFor}-hint` : undefined;
  const errorId = error ? `${htmlFor}-error` : undefined;
  return (
    <div className="space-y-1.5">
      <FieldLabel htmlFor={htmlFor}>{label}</FieldLabel>
      <div className="space-y-1">
        {children}
        {hint ? (
          <p id={hintId} className="text-xs text-muted-foreground">
            {hint}
          </p>
        ) : null}
        {error ? (
          <p id={errorId} role="alert" className="text-xs text-danger">
            {error}
          </p>
        ) : null}
      </div>
    </div>
  );
}

export interface TextFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  hint?: string;
  error?: string;
  inputRef?: Ref<HTMLInputElement>;
}

export function TextField({
  label,
  hint,
  error,
  id,
  className,
  inputRef,
  ...props
}: TextFieldProps) {
  const fieldId = id ?? `field-${label.toLowerCase().replace(/\s+/g, "-")}`;
  return (
    <FieldShell label={label} htmlFor={fieldId} hint={hint} error={error}>
      <input
        id={fieldId}
        ref={inputRef}
        aria-describedby={error ? `${fieldId}-error` : hint ? `${fieldId}-hint` : undefined}
        aria-invalid={error ? true : undefined}
        className={cn(INPUT_CLASSES, FOCUS_RING, error && "border-danger", className)}
        {...props}
      />
    </FieldShell>
  );
}

export interface TextAreaFieldProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string;
  hint?: string;
  error?: string;
  textareaRef?: Ref<HTMLTextAreaElement>;
}

export function TextAreaField({
  label,
  hint,
  error,
  id,
  className,
  textareaRef,
  rows = 4,
  ...props
}: TextAreaFieldProps) {
  const fieldId = id ?? `field-${label.toLowerCase().replace(/\s+/g, "-")}`;
  return (
    <FieldShell label={label} htmlFor={fieldId} hint={hint} error={error}>
      <textarea
        id={fieldId}
        rows={rows}
        ref={textareaRef}
        aria-describedby={error ? `${fieldId}-error` : hint ? `${fieldId}-hint` : undefined}
        aria-invalid={error ? true : undefined}
        className={cn(
          INPUT_CLASSES,
          FOCUS_RING,
          "resize-y leading-relaxed",
          error && "border-danger",
          className,
        )}
        {...props}
      />
    </FieldShell>
  );
}

export interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  hint?: string;
  error?: string;
  selectRef?: Ref<HTMLSelectElement>;
}

export function SelectField({
  label,
  hint,
  error,
  id,
  className,
  selectRef,
  children,
  ...props
}: SelectFieldProps) {
  const fieldId = id ?? `field-${label.toLowerCase().replace(/\s+/g, "-")}`;
  return (
    <FieldShell label={label} htmlFor={fieldId} hint={hint} error={error}>
      <select
        id={fieldId}
        ref={selectRef}
        aria-describedby={error ? `${fieldId}-error` : hint ? `${fieldId}-hint` : undefined}
        className={cn(INPUT_CLASSES, FOCUS_RING, "appearance-none pr-8", className)}
        {...props}
      >
        {children}
      </select>
    </FieldShell>
  );
}

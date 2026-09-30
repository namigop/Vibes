"use client";

import clsx from "clsx";
import { ApiClientError } from "@/lib/api";

interface Props {
  error: unknown;
  className?: string;
}

export function ErrorBanner({ error, className }: Props) {
  let title = "Something went wrong";
  let message = "Please try again.";
  let details: Array<{ field: string; message: string }> | undefined;

  if (error instanceof ApiClientError) {
    title = `Error ${error.status}`;
    message = error.message;
    details = error.details;
  } else if (error instanceof Error) {
    message = error.message;
  }

  return (
    <div
      role="alert"
      className={clsx(
        "rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-800",
        "dark:border-red-800 dark:bg-red-950 dark:text-red-200",
        className,
      )}
    >
      <p className="font-semibold">{title}</p>
      <p>{message}</p>
      {details && details.length > 0 ? (
        <ul className="mt-2 list-disc pl-5">
          {details.map((d) => (
            <li key={d.field}>
              <span className="font-mono">{d.field}</span>: {d.message}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}

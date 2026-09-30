"use client";

import { useSyncExternalStore } from "react";

/**
 * True only after hydration.
 *
 * Used instead of the `useState(false) + useEffect(setState(true))` mount flag
 * so nothing in the tree depends on a state update that happens inside an
 * effect. The first client render returns the *server* snapshot (`false`),
 * which is what keeps SSR markup and hydration in agreement.
 */
const noopSubscribe = () => () => {};

export function useIsClient(): boolean {
  return useSyncExternalStore(
    noopSubscribe,
    () => true,
    () => false,
  );
}

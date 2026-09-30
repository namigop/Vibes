"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import { THEME_STORAGE_KEY } from "@/lib/constants";

export type Theme = "light" | "dark";

interface ThemeContextValue {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

function isTheme(value: string | null): value is Theme {
  return value === "light" || value === "dark";
}

/**
 * The stored choice always wins; `prefers-color-scheme` is only the default for
 * a visitor who has never toggled. The identical resolution runs in the
 * blocking script in the document head — this function is the client-side
 * mirror of it, used once React takes over.
 */
function resolveInitialTheme(): Theme {
  if (typeof window === "undefined") return "light";
  try {
    const stored = window.localStorage.getItem(THEME_STORAGE_KEY);
    if (isTheme(stored)) return stored;
  } catch {
    // Private mode / disabled storage: fall through to the OS preference.
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function applyTheme(theme: Theme): void {
  const root = document.documentElement;
  root.classList.toggle("dark", theme === "dark");
  root.style.colorScheme = theme;
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(resolveInitialTheme);

  useEffect(() => {
    applyTheme(theme);
    try {
      window.localStorage.setItem(THEME_STORAGE_KEY, theme);
    } catch {
      // Non-fatal: the class is already applied for this session.
    }
  }, [theme]);

  const setTheme = useCallback((next: Theme) => setThemeState(next), []);
  const toggleTheme = useCallback(
    () => setThemeState((current) => (current === "dark" ? "light" : "dark")),
    [],
  );

  const value = useMemo<ThemeContextValue>(
    () => ({ theme, setTheme, toggleTheme }),
    [theme, setTheme, toggleTheme],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (!context) throw new Error("useTheme must be used inside <ThemeProvider>");
  return context;
}

/**
 * Inlined into <head> at build time. It runs before first paint so a dark-mode
 * reload never flashes a white screen. Kept tiny, dependency-free and wrapped
 * in try/catch — a storage exception must not break the page.
 */
export const themeInitScript = [
  "(function(){try{",
  `var k=${JSON.stringify(THEME_STORAGE_KEY)};`,
  "var s=null;",
  "try{s=window.localStorage.getItem(k);}catch(e){}",
  "var t=(s==='light'||s==='dark')?s:",
  "(window.matchMedia&&window.matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light');",
  "var r=document.documentElement;",
  "r.classList.toggle('dark',t==='dark');",
  "r.style.colorScheme=t;",
  "}catch(e){}})();",
].join("");

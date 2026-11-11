import React, { createContext, useContext, useState, useEffect, useCallback } from "react";

export type AppearanceMode = "light" | "dark" | "system";

interface ThemeContextType {
  appearance: AppearanceMode;
  setAppearance: (mode: AppearanceMode) => void;
  isDark: boolean;
}

const ThemeContext = createContext<ThemeContextType>({
  appearance: "system",
  setAppearance: () => {},
  isDark: false,
});

export function useTheme() {
  return useContext(ThemeContext);
}

function getSystemDark(): boolean {
  if (typeof window === "undefined") return false;
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [appearance, setAppearanceState] = useState<AppearanceMode>(() => {
    try {
      return (localStorage.getItem("echo-appearance") as AppearanceMode) || "system";
    } catch {
      return "system";
    }
  });

  const [systemDark, setSystemDark] = useState(getSystemDark);

  const isDark =
    appearance === "dark" ? true :
    appearance === "light" ? false :
    systemDark;

  // Listen for system color scheme changes
  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = (e: MediaQueryListEvent) => setSystemDark(e.matches);
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  // Apply appearance to DOM
  useEffect(() => {
    if (isDark) {
      document.documentElement.setAttribute("data-appearance", "dark");
    } else {
      document.documentElement.removeAttribute("data-appearance");
    }
  }, [isDark]);

  const setAppearance = useCallback((mode: AppearanceMode) => {
    setAppearanceState(mode);
    try {
      localStorage.setItem("echo-appearance", mode);
    } catch {}
  }, []);

  return (
    <ThemeContext.Provider value={{ appearance, setAppearance, isDark }}>
      {children}
    </ThemeContext.Provider>
  );
}

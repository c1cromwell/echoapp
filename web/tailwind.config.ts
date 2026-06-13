import type { Config } from "tailwindcss";

// Mirrors the Glacial theme palette used in the iOS app (GlacialTheme.swift) so the
// web portal reads as the same product surface.
const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        glacial: {
          bg: "#0B1220",
          surface: "#111B2E",
          border: "#1E2A40",
          accent: "#3BA9F4",
          muted: "#7C8BA1",
        },
      },
    },
  },
  plugins: [],
};

export default config;

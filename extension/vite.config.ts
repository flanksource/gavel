import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";
import { resolve } from "node:path";
import { developmentAliases } from "./scripts/dependencies.ts";

export default defineConfig(({ mode }) => ({
  plugins: [react()],
  resolve: { dedupe: ["react", "react-dom"], alias: mode === "development" ? developmentAliases(import.meta.dirname) : [] },
  server: mode === "development" ? { fs: { allow: [import.meta.dirname, resolve(import.meta.dirname, "../../clicky-ui")] } } : {},
  test: { environment: "jsdom", include: ["{src,host,comments,agents,scripts}/**/*.test.{ts,tsx}"] },
}));

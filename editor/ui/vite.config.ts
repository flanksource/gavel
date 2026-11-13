import { createRequire } from "node:module";
import { existsSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { localClicky } from "../../testrunner/ui/vite.clicky.ts";
import { runtimeAssets } from "./runtime-assets.ts";

const require = createRequire(import.meta.url);
const here = dirname(fileURLToPath(import.meta.url));

export default defineConfig(({ command }) => {
  // Resolves clicky-ui from the sibling checkout's source when it is present.
  const clicky = localClicky(resolve(here, "../.."), "editor/ui");
  const siblings = resolve(here, "../../..");
  const localRuntime = command === "serve" && existsSync(resolve(siblings, "facet/src/playground/browser-runtime.tsx"));
  return {
    plugins: [react(), tailwindcss(), runtimeAssets()],
    // Dev server for `gavel edit --dev`, which reverse-proxies the page here.
    // 5273 is pr/ui's, 5275 testrunner/ui's and 5274 the clicky-ui playground's;
    // strictPort makes a collision fail loudly rather than drift.
    server: {
      port: 5276,
      strictPort: true,
      ...clicky.server,
    },
    resolve: {
      // clicky-ui's source resolves its own React from outside this package;
      // two Reacts in one bundle break hooks, so every copy is pinned here.
      dedupe: ["react", "react-dom", "@tanstack/react-query"],
      alias: [
        ...(localRuntime ? [
          { find: "@flanksource/facet/browser-preview", replacement: resolve(siblings, "facet/src/playground/browser-runtime.tsx") },
          { find: "@flanksource/page-runtime/host", replacement: resolve(siblings, "clicky-ui/packages/page-runtime/src/host/index.ts") },
          { find: "@flanksource/page-runtime/styles", replacement: resolve(siblings, "clicky-ui/packages/page-runtime/src/styles.ts") },
        ] : []),
        ...clicky.aliases,
        { find: /^react$/, replacement: require.resolve("react") },
        { find: /^react-dom$/, replacement: require.resolve("react-dom") },
        { find: /^react-dom\/client$/, replacement: require.resolve("react-dom/client") },
        { find: /^react\/jsx-runtime$/, replacement: require.resolve("react/jsx-runtime") },
        { find: /^react\/jsx-dev-runtime$/, replacement: require.resolve("react/jsx-dev-runtime") },
        { find: /^@tanstack\/react-query$/, replacement: require.resolve("@tanstack/react-query") },
      ],
    },
    build: { outDir: "dist" },
  };
});

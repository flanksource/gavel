import { createRequire } from "node:module";
import { readdirSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import type { Plugin } from "vite";

export function runtimeAssets(): Plugin {
  const require = createRequire(import.meta.url);
  return {
    name: "page-runtime-assets",
    generateBundle() {
      const directory = dirname(require.resolve("@flanksource/page-runtime/static/sw.js"));
      for (const name of readdirSync(directory)) {
        if (name.endsWith(".map")) continue;
        this.emitFile({ type: "asset", fileName: `page-runtime/${name}`, source: readFileSync(join(directory, name)) });
      }
    },
  };
}

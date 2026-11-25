import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import { dirname, resolve } from "node:path";

/** bippy's classic-script installer for the React DevTools global hook (it patches a hook that already exists, so callers decide whether to run it). */
export function reactDevtoolsHookScript(root: string): string {
  const bippy = dirname(createRequire(resolve(root, "package.json")).resolve("bippy/package.json"));
  return readFileSync(resolve(bippy, "dist/install-hook-only.iife.js"), "utf8");
}


import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import type { Alias } from "vite";

export function developmentAliases(root: string): Alias[] {
  const ui = resolve(root, "../../clicky-ui/packages/ui");
  const manifest = JSON.parse(readFileSync(resolve(ui, "package.json"), "utf8")) as { exports: Record<string, string | { import?: string }> };
  const aliases: Alias[] = [];
  for (const [key, entry] of Object.entries(manifest.exports)) {
    const target = typeof entry === "string" ? entry : entry.import;
    const source = target?.match(/^\.\/dist\/(.*)\.js$/)?.[1];
    if (!source) continue;
    const name = key === "." ? "@flanksource/clicky-ui" : `@flanksource/clicky-ui/${key.slice(2)}`;
    aliases.push({ find: new RegExp(`^${name.replace(/[.*+?^${}()|[\]\\/]/g, "\\$&")}$`), replacement: resolve(ui, `src/${source}.ts`) });
  }
  return [
    ...aliases,
    { find: /^@flanksource\/page-runtime\/(.*)$/, replacement: resolve(root, "../../clicky-ui/packages/page-runtime/src/$1.ts") },
    { find: /^@flanksource\/react-grab\/core$/, replacement: resolve(root, "../../clicky-ui/packages/react-grab/dist/core/index.js") },
    { find: /^@flanksource\/react-grab$/, replacement: resolve(root, "../../clicky-ui/packages/react-grab/dist/index.js") },
  ];
}

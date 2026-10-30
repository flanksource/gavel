import { existsSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import type { Alias } from 'vite';

// localClicky links @flanksource/clicky-ui to the sibling checkout
// (../clicky-ui/packages/ui next to gavel) for both `vite serve` and
// `vite build`, so local bundles and `--dev` runs use clicky-ui's source,
// including subpaths not yet published. It is gated on the sibling being
// present, so CI keeps the published package.
//
// The aliases come from the sibling's own package.json exports: each
// `./dist/<x>.js` entry resolves to `src/<x>.ts(x)`; stylesheets stay on the
// sibling's built `dist/` (run its build once). New clicky-ui subpaths link
// without touching this file.
export interface LocalClicky {
  enabled: boolean;
  aliases: Alias[];
  server: { fs?: { allow: string[] }; watch?: { ignored: string[] } };
}

export function localClicky(gavelRoot: string, label: string): LocalClicky {
  const pkgDir = resolve(gavelRoot, '../clicky-ui/packages/ui');
  const src = resolve(pkgDir, 'src');
  if (!existsSync(src)) {
    return { enabled: false, aliases: [], server: {} };
  }
  console.log(`[${label}] resolving @flanksource/clicky-ui from local source (${src})`);

  const exports: Record<string, string | { import?: string }> = JSON.parse(
    readFileSync(resolve(pkgDir, 'package.json'), 'utf8'),
  ).exports;
  const aliases: Alias[] = [];
  for (const [subpath, target] of Object.entries(exports)) {
    const file = typeof target === 'string' ? target : target.import;
    if (!file) continue;
    const specifier = subpath === '.' ? '@flanksource/clicky-ui' : `@flanksource/clicky-ui/${subpath.slice(2)}`;
    const find = new RegExp(`^${specifier.replace(/[.*+?^${}()|[\]\\/]/g, '\\$&')}$`);
    if (file.endsWith('.css')) {
      aliases.push({ find, replacement: resolve(pkgDir, file) });
      continue;
    }
    const stem = file.match(/^\.\/dist\/(.+)\.js$/)?.[1];
    const source = stem && [`${stem}.ts`, `${stem}.tsx`].map(f => resolve(src, f)).find(existsSync);
    if (source) aliases.push({ find, replacement: source });
  }

  // Every module Vite resolves outside `root` is handed to chokidar
  // individually, and its FSEvents backend opens a stream per parent
  // directory — the clicky-ui source spans ~50 of them. GAVEL_UI_WATCH_CLICKY=0
  // keeps resolving sibling source while dropping those streams, at the cost of
  // no invalidation on clicky-ui edits.
  const watch = process.env.GAVEL_UI_WATCH_CLICKY !== '0';
  if (!watch) console.log(`[${label}] dev: GAVEL_UI_WATCH_CLICKY=0 — clicky-ui source is not watched`);
  return {
    enabled: true,
    aliases,
    server: {
      // The sibling lives outside the project root; let Vite serve it and its
      // hoisted deps.
      fs: { allow: [resolve(gavelRoot, '..')] },
      ...(watch ? {} : { watch: { ignored: [`${src}/**`] } }),
    },
  };
}

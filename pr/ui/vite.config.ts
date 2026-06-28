import { execSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

const require = createRequire(import.meta.url);
const here = dirname(fileURLToPath(import.meta.url));

// Frontend build version, baked in so the Settings panel can show which UI
// bundle is embedded — and reveal drift from the Go binary's own version.
function git(args: string, fallback: string): string {
  try {
    return (
      execSync(`git ${args}`, { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] }).trim() || fallback
    );
  } catch {
    return fallback;
  }
}
const uiVersion = git('describe --tags --always', 'dev');
const uiCommit = git('rev-parse --short HEAD', 'unknown');
const uiDate = new Date().toISOString();

// Sibling clicky-ui checkout (the package pnpm-workspace.yaml links locally). In
// dev (`pr list --ui --dev`) resolve clicky-ui's JS subpaths to its *source* so
// sibling edits are live with HMR — the package only publishes `dist`, so the
// linked symlink otherwise serves stale built output. Only the JS entry points
// are redirected; `styles.css` keeps resolving to the package's generated CSS.
const clickySrc = resolve(here, '../../../clicky-ui/packages/ui/src');
const clickySubpaths = ['components', 'data', 'icons', 'hooks', 'ai'];

export default defineConfig(({ command }) => {
  // Gated to `serve` + sibling-present so `vite build` and CI (no sibling) keep
  // resolving the installed/published package — nothing local-path is committed
  // into the dependency graph.
  const localClicky = command === 'serve' && existsSync(clickySrc);
  if (localClicky) {
    console.log(`[pr/ui] dev: resolving @flanksource/clicky-ui from local source (${clickySrc})`);
  }
  const clickyAliases = localClicky
    ? clickySubpaths.map(sub => ({
        find: `@flanksource/clicky-ui/${sub}`,
        replacement: resolve(clickySrc, `${sub}.ts`),
      }))
    : [];

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      dedupe: ['react', 'react-dom', '@tanstack/react-query'],
      alias: [
        ...clickyAliases,
        { find: /^react$/, replacement: require.resolve('react') },
        { find: /^react-dom$/, replacement: require.resolve('react-dom') },
        { find: /^react-dom\/client$/, replacement: require.resolve('react-dom/client') },
        { find: /^@tanstack\/react-query$/, replacement: require.resolve('@tanstack/react-query') },
      ],
    },
    define: {
      // Lib/IIFE builds don't auto-replace process.env.NODE_ENV, which React and
      // @tanstack/react-query reference at runtime — define it for `build` so the
      // bundle has no bare `process` reference. NOT for `serve`: forcing
      // 'production' there disables React Fast Refresh / HMR.
      ...(command === 'build' ? { 'process.env.NODE_ENV': JSON.stringify('production') } : {}),
      __GAVEL_UI_VERSION__: JSON.stringify(uiVersion),
      __GAVEL_UI_COMMIT__: JSON.stringify(uiCommit),
      __GAVEL_UI_DATE__: JSON.stringify(uiDate),
    },
    // Dev server for `pr list --ui --dev`. The Go server reverse-proxies the page
    // to this port; HMR runs on its own port so the websocket connects directly
    // and never traverses the Go proxy. Dedicated (non-default) ports avoid
    // colliding with the Vite default 5173 that sibling UIs (e.g. clicky-ui)
    // commonly hold; strictPort makes a collision fail loudly rather than drift.
    server: {
      port: 5273,
      strictPort: true,
      hmr: { port: 24778 },
      // The local clicky-ui source lives outside this project root; allow Vite to
      // serve it (and its hoisted deps) when the dev alias above is active.
      ...(localClicky ? { fs: { allow: [resolve(here, '../../..')] } } : {}),
    },
    build: {
      lib: {
        entry: 'src/index.tsx',
        name: 'PRUI',
        formats: ['iife'],
        fileName: () => 'prui.js',
      },
      outDir: 'dist',
      minify: true,
      rollupOptions: {
        output: {
          inlineDynamicImports: true,
          // Stable CSS filename so the Go server can go:embed it.
          assetFileNames: 'prui.[ext]',
        },
      },
    },
  };
});

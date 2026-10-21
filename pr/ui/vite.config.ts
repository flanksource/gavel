import { execSync } from 'node:child_process';
import { createRequire } from 'node:module';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { localClicky } from '../../testrunner/ui/vite.clicky';

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

const gavelTestRunnerHooks = resolve(here, '../../testrunner/ui/src/hooks.ts');

export default defineConfig(({ command }) => {
  // `pr list --ui --dev` resolves clicky-ui from the sibling checkout's source.
  const clicky = localClicky(resolve(here, '../..'), command, 'pr/ui');

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      dedupe: ['react', 'react-dom', '@tanstack/react-query', '@flanksource/clicky-ui'],
      alias: [
        ...clicky.aliases,
        { find: '@flanksource/gavel/testrunner/hooks', replacement: gavelTestRunnerHooks },
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
      // 'production' there changes React's development behavior.
      ...(command === 'build' ? { 'process.env.NODE_ENV': JSON.stringify('production') } : {}),
      __GAVEL_UI_VERSION__: JSON.stringify(uiVersion),
      __GAVEL_UI_COMMIT__: JSON.stringify(uiCommit),
      __GAVEL_UI_DATE__: JSON.stringify(uiDate),
    },
    // Dev server for `pr list --ui --dev`. The Go server reverse-proxies the page
    // to this port. Dedicated (non-default) ports avoid
    // colliding with the Vite default 5173 that sibling UIs (e.g. clicky-ui)
    // commonly hold; strictPort makes a collision fail loudly rather than drift.
    server: {
      port: 5273,
      strictPort: true,
      // Keep watching source while requiring a browser refresh to see changes.
      hmr: false,
      ...clicky.server,
    },
    build: {
      // ES-module library build: the stable `prui.js` entry is loaded as
      // `<script type="module">` and lazy imports (e.g. the heavy MDX editor)
      // are code-split into hashed `chunks/*.js` the Go server serves from
      // /_assets/. (IIFE + inlineDynamicImports would force a single bundle.)
      lib: {
        entry: 'src/index.tsx',
        formats: ['es'],
        fileName: () => 'prui.js',
      },
      outDir: 'dist',
      minify: true,
      // Keep all CSS in one prui.css so the Go server can inline it as today
      // (only the JS is split).
      cssCodeSplit: false,
      rollupOptions: {
        output: {
          chunkFileNames: 'chunks/[name]-[hash].js',
          // Stable CSS filename so the Go server can go:embed it.
          assetFileNames: 'prui.[ext]',
        },
      },
    },
  };
});

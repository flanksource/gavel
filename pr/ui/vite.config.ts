import { execSync } from 'node:child_process';
import { createRequire } from 'node:module';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

const require = createRequire(import.meta.url);

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

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    dedupe: ['react', 'react-dom', '@tanstack/react-query'],
    alias: [
      { find: /^react$/, replacement: require.resolve('react') },
      { find: /^react-dom$/, replacement: require.resolve('react-dom') },
      { find: /^react-dom\/client$/, replacement: require.resolve('react-dom/client') },
      { find: /^@tanstack\/react-query$/, replacement: require.resolve('@tanstack/react-query') },
    ],
  },
  // Lib/IIFE builds don't auto-replace process.env.NODE_ENV, which React and
  // @tanstack/react-query reference at runtime — define it so the bundle has no
  // bare `process` reference in the browser.
  define: {
    'process.env.NODE_ENV': JSON.stringify('production'),
    __GAVEL_UI_VERSION__: JSON.stringify(uiVersion),
    __GAVEL_UI_COMMIT__: JSON.stringify(uiCommit),
    __GAVEL_UI_DATE__: JSON.stringify(uiDate),
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
});

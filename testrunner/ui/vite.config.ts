import { rmSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig, type Plugin } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { localClicky } from './vite.clicky';

const require = createRequire(import.meta.url);
const here = dirname(fileURLToPath(import.meta.url));

export default defineConfig(({ command }) => {
  // `gavel test --ui --dev` resolves clicky-ui from the sibling checkout's source.
  const clicky = localClicky(resolve(here, '../..'), command, 'testrunner/ui');
  return {
    // Dev server for `gavel test|fixtures --ui --dev`, which reverse-proxies the
    // page here. 5273 is pr/ui's and 5274 the clicky-ui playground's; strictPort
    // makes a collision fail loudly rather than drift.
    server: {
      port: 5275,
      strictPort: true,
      // Keep watching source while requiring a browser refresh to see changes.
      hmr: false,
      ...clicky.server,
    },
    plugins: [react(), tailwindcss(), cleanChunks()],
    // Lib builds don't replace process.env.NODE_ENV, which React, TanStack
    // Query and clicky-ui's data views reference at runtime; without it the
    // bundle throws before rendering. Not for `serve`: forcing 'production' there
    // changes React's development behavior.
    define: command === 'build' ? { 'process.env.NODE_ENV': JSON.stringify('production') } : {},
    resolve: {
      // When @flanksource/clicky-ui resolves from the sibling checkout its
      // realpath lives outside this package's node_modules and resolves its own
      // copy of react; two Reacts in one bundle break hooks and elements. Pin
      // every copy to this package's.
      dedupe: ['react', 'react-dom', '@tanstack/react-query'],
      alias: [
        ...clicky.aliases,
        { find: /^react$/, replacement: require.resolve('react') },
        { find: /^react-dom$/, replacement: require.resolve('react-dom') },
        { find: /^react-dom\/client$/, replacement: require.resolve('react-dom/client') },
        { find: /^@tanstack\/react-query$/, replacement: require.resolve('@tanstack/react-query') },
      ],
    },
    build: {
      // ES-module build: the Go server (assets.go) serves the stable testui.js
      // entry and its hashed chunks/*.js under /_testui/, and the page loads it
      // as <script type="module">, so lazy imports stay split.
      lib: {
        entry: 'src/index.tsx',
        formats: ['es'],
        fileName: () => 'testui.js',
      },
      outDir: 'dist',
      // dist/ is shared with the library build (vite.lib.config.ts -> dist/lib),
      // which pr/ui imports as @flanksource/gavel/testrunner. Emptying it here
      // would delete that consumer's entrypoints on every embed-bundle build;
      // cleanChunks drops only the previous build's chunks.
      emptyOutDir: false,
      minify: true,
      rolldownOptions: {
        output: { chunkFileNames: 'chunks/[name]-[hash].js' },
      },
    },
  };
});

// cleanChunks removes the previous embed build's chunks, which emptyOutDir:false
// would otherwise leave to accumulate (and be embedded) under dist/chunks.
function cleanChunks(): Plugin {
  return {
    name: 'gavel-clean-chunks',
    apply: 'build',
    buildStart() {
      rmSync(resolve(here, 'dist/chunks'), { recursive: true, force: true });
    },
  };
}

import { createRequire } from 'node:module';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

const require = createRequire(import.meta.url);

export default defineConfig(({ command }) => ({
  server: {
    hmr: false,
  },
  plugins: [react(), tailwindcss()],
  // Lib/IIFE builds don't replace process.env.NODE_ENV, which React, TanStack
  // Query and clicky-ui's data views reference at runtime; without it the
  // bundle throws before rendering. Not for `serve`: forcing 'production' there
  // changes React's development behavior.
  define: command === 'build' ? { 'process.env.NODE_ENV': JSON.stringify('production') } : {},
  resolve: {
    // When @flanksource/clicky-ui is linked from a local sibling checkout its
    // realpath lives outside this package's node_modules and resolves its own
    // copy of react; two Reacts in one bundle break hooks and elements. Pin
    // every copy to this package's.
    dedupe: ['react', 'react-dom', '@tanstack/react-query'],
    alias: [
      { find: /^react$/, replacement: require.resolve('react') },
      { find: /^react-dom$/, replacement: require.resolve('react-dom') },
      { find: /^react-dom\/client$/, replacement: require.resolve('react-dom/client') },
      { find: /^@tanstack\/react-query$/, replacement: require.resolve('@tanstack/react-query') },
    ],
  },
  build: {
    lib: {
      entry: 'src/index.tsx',
      name: 'TestUI',
      formats: ['iife'],
      fileName: () => 'testui.js',
    },
    outDir: 'dist',
    // dist/ is shared with the library build (vite.lib.config.ts -> dist/lib),
    // which pr/ui imports as @flanksource/gavel/testrunner. Emptying it here
    // would delete that consumer's entrypoints on every embed-bundle build.
    emptyOutDir: false,
    minify: true,
    rollupOptions: {
      output: { inlineDynamicImports: true },
    },
  },
}));

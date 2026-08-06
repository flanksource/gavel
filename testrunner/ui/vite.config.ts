import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

export default defineConfig({
  plugins: [preact()],
  resolve: {
    // The preact preset aliases React (incl. react/jsx-runtime) to preact. When
    // @flanksource/clicky-ui is linked from a local sibling checkout, its realpath
    // lives outside this package's node_modules, so those aliased preact imports
    // would resolve from the sibling (which has no preact) and fail. Dedupe pins
    // preact to this package's copy regardless of the importer's location, while
    // leaving clicky-ui's own deps to resolve from its realpath.
    dedupe: ['preact'],
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
});

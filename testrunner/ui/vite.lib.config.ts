import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import dts from 'vite-plugin-dts';

// Library build for the published @flanksource/gavel package. Real React
// (externalized to the consumer), dual ESM + CJS, with .d.ts emitted on the
// ESM pass. Writes to dist/lib so it never touches the embed bundle
// (dist/testui.js, produced by vite.config.ts).
const __dirname = dirname(fileURLToPath(import.meta.url));

const entry = {
  index: resolve(__dirname, 'src/index.ts'),
  testrunner: resolve(__dirname, 'src/testrunner.ts'),
  hooks: resolve(__dirname, 'src/hooks.ts'),
  types: resolve(__dirname, 'src/types.ts'),
};

export default defineConfig(({ mode }) => {
  const isCjs = mode === 'cjs';
  const jsExt = isCjs ? 'cjs' : 'js';

  return {
    plugins: isCjs
      ? [react()]
      : [
          react(),
          dts({
            tsconfigPath: './tsconfig.lib.json',
            include: ['src'],
            exclude: ['src/**/*.test.*', 'src/index.tsx'],
            skipDiagnostics: true,
          }),
        ],
    build: {
      outDir: 'dist/lib',
      emptyOutDir: !isCjs,
      sourcemap: true,
      minify: false,
      lib: {
        entry,
        formats: [isCjs ? 'cjs' : 'es'],
        fileName: (_format, name) => `${name}.${jsExt}`,
      },
      rollupOptions: {
        external: (id) => !(id.startsWith('.') || id.startsWith('/')),
        output: {
          preserveModules: true,
          preserveModulesRoot: 'src',
          entryFileNames: `[name].${jsExt}`,
          chunkFileNames: `chunks/[name]-[hash].${jsExt}`,
        },
      },
    },
  };
});

import { createRequire } from 'node:module';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

const here = dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);

// Component tests run against real React in jsdom so the settings prompt row and
// its nested spec-editor dialog are exercised as the app renders them.
export default defineConfig({
  plugins: [react()],
  resolve: {
    dedupe: ['react', 'react-dom', '@tanstack/react-query'],
    alias: [
      { find: '@flanksource/gavel/testrunner/hooks', replacement: resolve(here, '../../testrunner/ui/src/hooks.ts') },
      // monaco-editor ships no `main`, only `module`, which vitest's node-side
      // resolver does not follow; setup mocks the package and needs the id to resolve.
      { find: /^monaco-editor$/, replacement: 'monaco-editor/esm/vs/editor/editor.api' },
      { find: /^react$/, replacement: require.resolve('react') },
      { find: /^react-dom$/, replacement: require.resolve('react-dom') },
      { find: /^react-dom\/client$/, replacement: require.resolve('react-dom/client') },
      { find: /^@tanstack\/react-query$/, replacement: require.resolve('@tanstack/react-query') },
    ],
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/vitest.setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
    server: {
      deps: {
        // Inlined so the react aliases above apply: a dependency vitest
        // externalizes loads whichever React sits beside its own realpath, and a
        // linked clicky-ui checkout's deps (radix, streamdown, ...) resolve its
        // React 18 next to this package's 19.
        inline: [
          /@flanksource\/clicky-ui/,
          /@floating-ui\/react/,
          /@tanstack\/react-virtual/,
          /@radix-ui\//,
          /streamdown/,
        ],
      },
    },
  },
});

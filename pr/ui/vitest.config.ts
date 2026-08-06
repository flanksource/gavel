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
      { find: /^react$/, replacement: require.resolve('react') },
      { find: /^react-dom$/, replacement: require.resolve('react-dom') },
      { find: /^react-dom\/client$/, replacement: require.resolve('react-dom/client') },
      { find: /^@tanstack\/react-query$/, replacement: require.resolve('@tanstack/react-query') },
    ],
  },
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.test.{ts,tsx}'],
    server: {
      deps: {
        inline: [/@flanksource\/clicky-ui/, /@floating-ui\/react/],
      },
    },
  },
});

import { createRequire } from 'node:module';
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

const require = createRequire(import.meta.url);

// Tests run against real React (jsdom) so the published library surface — the
// hooks, adapters and components — is exercised exactly as external consumers
// see it. Pure-logic tests (routes, utils) are React-agnostic and pass here too.
export default defineConfig({
  plugins: [react()],
  resolve: {
    // A linked clicky-ui checkout resolves its own React; pin every copy to this package's.
    dedupe: ['react', 'react-dom', '@tanstack/react-query'],
    alias: [
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
        inline: [/@flanksource\/clicky-ui/, /@floating-ui\/react/, /@tanstack\/react-virtual/, /@radix-ui\//],
      },
    },
  },
});
